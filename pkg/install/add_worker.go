package install

import (
	"errors"
	"fmt"

	"github.com/apprenda/kismatic/pkg/util"
)

var errMissingClusterCA = errors.New("The Certificate Authority's private key and certificate used to install " +
	"the cluster are required for adding worker nodes.")

// AddWorker adds a worker node to the original cluster described in the plan.
// If successful, the updated plan is returned.
func (ae *ansibleExecutor) AddWorker(originalPlan *Plan, newWorker Node) (*Plan, error) {
	if err := checkAddWorkerPrereqs(ae.pki, newWorker); err != nil {
		return nil, err
	}
	updatedPlan := addWorkerToPlan(*originalPlan, newWorker)

	// Generate node certificates
	util.PrintHeader(ae.stdout, "Generating Certificate For Worker Node", '=')
	ca, err := ae.pki.GetClusterCA()
	if err != nil {
		return nil, err
	}
	if err = ae.pki.GenerateNodeCertificate(&updatedPlan, newWorker, ca); err != nil {
		return nil, fmt.Errorf("error generating certificate for new worker: %v", err)
	}

	// Run the playbook to add the worker
	inventory := buildInventoryFromPlan(&updatedPlan)
	cc, err := ae.buildClusterCatalog(&updatedPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ansible vars: %v", err)
	}
	util.PrintHeader(ae.stdout, "Adding Worker Node to Cluster", '=')
	t := task{
		name:           "add-worker",
		playbook:       "kubernetes-worker.yaml",
		plan:           updatedPlan,
		inventory:      inventory,
		clusterCatalog: *cc,
		explainer:      ae.defaultExplainer(),
		limit:          []string{newWorker.Host},
	}
	if err = ae.execute(t); err != nil {
		return nil, fmt.Errorf("error running playbook: %v", err)
	}

	// We need to run ansible against all hosts to update the hosts files
	if updatedPlan.Cluster.Networking.UpdateHostsFiles {
		util.PrintHeader(ae.stdout, "Updating Hosts Files On All Nodes", '=')
		t = task{
			name:           "add-worker-update-hosts",
			playbook:       "_hosts.yaml",
			plan:           updatedPlan,
			inventory:      inventory,
			clusterCatalog: *cc,
			explainer:      ae.defaultExplainer(),
		}
		if err = ae.execute(t); err != nil {
			return nil, fmt.Errorf("error updating hosts files on all nodes: %v", err)
		}
	}

	// Verify that the node registered with API server
	util.PrintHeader(ae.stdout, "Running New Worker Smoke Test", '=')
	cc.WorkerNode = newWorker.Host
	t = task{
		name:           "add-worker-smoke-test",
		playbook:       "_worker-smoke-test.yaml",
		plan:           updatedPlan,
		inventory:      inventory,
		clusterCatalog: *cc,
		explainer:      ae.defaultExplainer(),
		limit:          []string{newWorker.Host},
	}
	if err = ae.execute(t); err != nil {
		return nil, fmt.Errorf("error running worker smoke test: %v", err)
	}

	// Allow access to new worker to any storage volumes defined
	if len(originalPlan.Storage.Nodes) > 0 {
		util.PrintHeader(ae.stdout, "Updating Allowed IPs On Storage Volumes", '=')
		t = task{
			name:           "add-worker-update-volumes",
			playbook:       "_volume-update-allowed.yaml",
			plan:           updatedPlan,
			inventory:      inventory,
			clusterCatalog: *cc,
			explainer:      ae.defaultExplainer(),
		}
		if err = ae.execute(t); err != nil {
			return nil, fmt.Errorf("error adding new worker to volume allow list: %v", err)
		}
	}
	return &updatedPlan, nil
}

func addWorkerToPlan(plan Plan, worker Node) Plan {
	plan.Worker.ExpectedCount++
	plan.Worker.Nodes = append(plan.Worker.Nodes, worker)
	return plan
}

// ensure the assumptions we are making are solid
func checkAddWorkerPrereqs(pki PKI, newWorker Node) error {
	// 1. if the node certificate is not there, we need to ensure that
	// the CA is available for generating the new worker's cert
	// don't check for a valid cert here since its already being done in GenerateNodeCertificate()
	certExists, err := pki.NodeCertificateExists(newWorker)
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
