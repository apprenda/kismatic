package install

import (
	"errors"
	"fmt"

	"github.com/apprenda/kismatic/pkg/util"
)

var errMissingClusterCA = errors.New("The Certificate Authority's private key and certificate used to install " +
	"the cluster are required for adding worker nodes.")

// AddNode adds a worker node to the original cluster described in the plan.
// If successful, the updated plan is returned.
func (ae *ansibleExecutor) AddNode(originalPlan *Plan, newNode Node, roles []string) (*Plan, error) {
	if err := checkAddNodePrereqs(ae.pki, newNode); err != nil {
		return nil, err
	}
	updatedPlan := addNodeToPlan(*originalPlan, newNode, roles)

	// Generate node certificates
	util.PrintHeader(ae.stdout, "Generating Certificate For New Node", '=')
	ca, err := ae.pki.GetClusterCA()
	if err != nil {
		return nil, err
	}
	if err = ae.pki.GenerateNodeCertificate(&updatedPlan, newNode, ca); err != nil {
		return nil, fmt.Errorf("error generating certificate for new node: %v", err)
	}

	// Run the playbook to add the node
	inventory := buildInventoryFromPlan(&updatedPlan)
	cc, err := ae.buildClusterCatalog(&updatedPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ansible vars: %v", err)
	}

	// We need to run ansible against all hosts to update the hosts files
	if updatedPlan.Cluster.Networking.UpdateHostsFiles {
		util.PrintHeader(ae.stdout, "Updating Hosts Files On All Nodes", '=')
		t := task{
			name:           "add-node-update-hosts",
			playbook:       "hosts.yaml",
			plan:           updatedPlan,
			inventory:      inventory,
			clusterCatalog: *cc,
			explainer:      ae.defaultExplainer(),
		}
		if err = ae.execute(t); err != nil {
			return nil, fmt.Errorf("error updating hosts files on all nodes: %v", err)
		}
	}

	util.PrintHeader(ae.stdout, "Adding New Node to Cluster", '=')
	t := task{
		name:           "add-node",
		playbook:       "kubernetes-node.yaml",
		plan:           updatedPlan,
		inventory:      inventory,
		clusterCatalog: *cc,
		explainer:      ae.defaultExplainer(),
		limit:          []string{newNode.Host},
	}
	if err = ae.execute(t); err != nil {
		return nil, fmt.Errorf("error running playbook: %v", err)
	}

	// Verify that the node registered with API server
	util.PrintHeader(ae.stdout, "Running New Node Smoke Test", '=')
	cc.NewNode = newNode.Host
	t = task{
		name:           "add-node-smoke-test",
		playbook:       "_node-smoke-test.yaml",
		plan:           updatedPlan,
		inventory:      inventory,
		clusterCatalog: *cc,
		explainer:      ae.defaultExplainer(),
		limit:          []string{newNode.Host},
	}
	if err = ae.execute(t); err != nil {
		return nil, fmt.Errorf("error running node smoke test: %v", err)
	}

	// Allow access to new node to any storage volumes defined
	if len(originalPlan.Storage.Nodes) > 0 {
		util.PrintHeader(ae.stdout, "Updating Allowed IPs On Storage Volumes", '=')
		t = task{
			name:           "add-node-update-volumes",
			playbook:       "_volume-update-allowed.yaml",
			plan:           updatedPlan,
			inventory:      inventory,
			clusterCatalog: *cc,
			explainer:      ae.defaultExplainer(),
		}
		if err = ae.execute(t); err != nil {
			return nil, fmt.Errorf("error adding new node to volume allow list: %v", err)
		}
	}
	return &updatedPlan, nil
}

func addNodeToPlan(plan Plan, node Node, roles []string) Plan {
	if util.Contains("worker", roles) {
		plan.Worker.ExpectedCount++
		plan.Worker.Nodes = append(plan.Worker.Nodes, node)
	}
	if util.Contains("ingress", roles) {
		plan.Ingress.ExpectedCount++
		plan.Ingress.Nodes = append(plan.Ingress.Nodes, node)
	}
	if util.Contains("storage", roles) {
		plan.Storage.ExpectedCount++
		plan.Storage.Nodes = append(plan.Storage.Nodes, node)
	}

	return plan
}

// ensure the assumptions we are making are solid
func checkAddNodePrereqs(pki PKI, newNode Node) error {
	// 1. if the node certificate is not there, we need to ensure that
	// the CA is available for generating the new nodes's cert
	// don't check for a valid cert here since its already being done in GenerateNodeCertificate()
	certExists, err := pki.NodeCertificateExists(newNode)
	if err != nil {
		return fmt.Errorf("error while checking if node's certificate exists: %v", err)
	}
	if !certExists {
		caExists, err := pki.CertificateAuthorityExists()
		if err != nil {
			return fmt.Errorf("error while checking if cluster CA exists: %v", err)
		}
		if !caExists {
			return errMissingClusterCA
		}
	}
	return nil
}
