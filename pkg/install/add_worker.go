package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install/explain"
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
	runDirectory, err := ae.createRunDirectory("add-worker")
	if err != nil {
		return nil, fmt.Errorf("error creating working directory for add-worker: %v", err)
	}
	updatedPlan := addWorkerToPlan(*originalPlan, newWorker)
	fp := FilePlanner{
		File: filepath.Join(runDirectory, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(&updatedPlan); err != nil {
		return nil, fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}
	// Generate node certificates
	util.PrintHeader(ae.stdout, "Generating Certificate For Worker Node", '=')
	ca, err := ae.pki.GetClusterCA()
	if err != nil {
		return nil, err
	}
	if err := ae.pki.GenerateNodeCertificate(originalPlan, newWorker, ca); err != nil {
		return nil, fmt.Errorf("error generating certificate for new worker: %v", err)
	}
	// Build the ansible inventory
	inventory := buildInventoryFromPlan(&updatedPlan)
	cc, err := ae.buildClusterCatalog(&updatedPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ansible vars: %v", err)
	}
	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return nil, fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}
	// Run the playbook for adding the node
	util.PrintHeader(ae.stdout, "Adding Worker Node to Cluster", '=')
	playbook := "kubernetes-worker.yaml"
	eventExplainer := &explain.DefaultEventExplainer{}
	runner, explainer, err := ae.ansibleRunnerWithExplainer(eventExplainer, ansibleLogFile, runDirectory)
	if err != nil {
		return nil, err
	}
	eventStream, err := runner.StartPlaybookOnNode(playbook, inventory, *cc, newWorker.Host)
	if err != nil {
		return nil, fmt.Errorf("error running ansible playbook: %v", err)
	}
	go explainer.Explain(eventStream)
	// Wait until ansible exits
	if err = runner.WaitPlaybook(); err != nil {
		return nil, fmt.Errorf("error running playbook: %v", err)
	}
	if updatedPlan.Cluster.Networking.UpdateHostsFiles {
		// We need to run ansible against all hosts to update the hosts files
		util.PrintHeader(ae.stdout, "Updating Hosts Files On All Nodes", '=')
		playbook := "_hosts.yaml"
		eventExplainer := &explain.DefaultEventExplainer{}
		runner, explainer, err := ae.ansibleRunnerWithExplainer(eventExplainer, ansibleLogFile, runDirectory)
		if err != nil {
			return nil, err
		}
		eventStream, err := runner.StartPlaybook(playbook, inventory, *cc)
		if err != nil {
			return nil, fmt.Errorf("error running playbook to update hosts files on all nodes: %v", err)
		}
		go explainer.Explain(eventStream)
		if err = runner.WaitPlaybook(); err != nil {
			return nil, fmt.Errorf("error updating hosts files on all nodes: %v", err)
		}
	}
	// Verify that the node registered with API server
	util.PrintHeader(ae.stdout, "Running New Worker Smoke Test", '=')
	playbook = "_worker-smoke-test.yaml"

	cc.WorkerNode = newWorker.Host

	eventExplainer = &explain.DefaultEventExplainer{}
	runner, explainer, err = ae.ansibleRunnerWithExplainer(eventExplainer, ansibleLogFile, runDirectory)
	if err != nil {
		return nil, err
	}
	eventStream, err = runner.StartPlaybook(playbook, inventory, *cc)
	if err != nil {
		return nil, fmt.Errorf("error running new worker smoke test: %v", err)
	}
	go explainer.Explain(eventStream)
	// Wait until ansible exits
	if err = runner.WaitPlaybook(); err != nil {
		return nil, fmt.Errorf("error running new worker smoke test: %v", err)
	}
	// Allow access to new worker to any storage volumes defined
	if len(originalPlan.Storage.Nodes) > 0 {
		util.PrintHeader(ae.stdout, "Updating Allowed IPs On Storage Volumes", '=')
		playbook = "_volume-update-allowed.yaml"
		eventExplainer = &explain.DefaultEventExplainer{}
		runner, explainer, err = ae.ansibleRunnerWithExplainer(eventExplainer, ansibleLogFile, runDirectory)
		if err != nil {
			return nil, err
		}
		eventStream, err = runner.StartPlaybook(playbook, inventory, *cc)
		if err != nil {
			return nil, fmt.Errorf("error adding new worker to volume allow list: %v", err)
		}
		go explainer.Explain(eventStream)
		if err = runner.WaitPlaybook(); err != nil {
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
