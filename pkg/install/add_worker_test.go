package install

import (
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/install/explain"
	"github.com/apprenda/kismatic/pkg/tls"
)

func mustGetTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "add-node-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
}

func TestAddWorkerCertMissingCAMissing(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki:                 &fakePKI{},
		certsDir:            mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Worker: NodeGroup{
			Nodes: []Node{},
		},
	}
	newWorker := Node{}
	newPlan, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if newPlan != nil {
		t.Errorf("add worker returned an updated plan")
	}
	if err != errMissingClusterCA {
		t.Errorf("AddWorker did not return the expected error. Instead returned: %v", err)
	}
}

// Verify that cert gets generated
func TestAddWorkerCertMissingCAExists(t *testing.T) {
	pki := &fakePKI{
		caExists: true,
	}
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki:                 pki,
		runnerExplainerFactory: fakeRunnerExplainer(nil),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			Nodes: []Node{},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{}
	_, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if err != nil {
		t.Errorf("unexpected error while adding worker: %v", err)
	}
	if pki.generateCACalled {
		t.Errorf("CA was generated, even though it already existed")
	}
	if !pki.generateNodeCertCalled {
		t.Error("node certificate was not generated")
	}
}

func TestAddWorkerPlanIsUpdated(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: fakeRunnerExplainer(nil),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	updatedPlan, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if err != nil {
		t.Errorf("unexpected error while adding worker: %v", err)
	}
	if updatedPlan.Worker.ExpectedCount != 2 {
		t.Errorf("expected count was not incremented")
	}
	if updatedPlan.Ingress.ExpectedCount != 0 {
		t.Errorf("expected ingress count was not 0")
	}
	if updatedPlan.Storage.ExpectedCount != 0 {
		t.Errorf("expected storage count was not 0")
	}
	found := false
	for _, w := range updatedPlan.Worker.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new node")
	}
}

func TestAddIngressPlanIsUpdated(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: fakeRunnerExplainer(nil),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Ingress: OptionalNodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	updatedPlan, err := e.AddNode(originalPlan, newWorker, []string{"ingress"})
	if err != nil {
		t.Errorf("unexpected error while adding worker: %v", err)
	}
	if updatedPlan.Ingress.ExpectedCount != 2 {
		t.Errorf("expected count was not incremented")
	}
	if updatedPlan.Worker.ExpectedCount != 0 {
		t.Errorf("expected worker count was not 0")
	}
	if updatedPlan.Storage.ExpectedCount != 0 {
		t.Errorf("expected storage count was not 0")
	}
	found := false
	for _, w := range updatedPlan.Ingress.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new node")
	}
}

func TestAddStoragePlanIsUpdated(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: fakeRunnerExplainer(nil),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Storage: OptionalNodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	updatedPlan, err := e.AddNode(originalPlan, newWorker, []string{"storage"})
	if err != nil {
		t.Errorf("unexpected error while adding worker: %v", err)
	}
	if updatedPlan.Storage.ExpectedCount != 2 {
		t.Errorf("expected count was not incremented")
	}
	if updatedPlan.Worker.ExpectedCount != 0 {
		t.Errorf("expected worker count was not 0")
	}
	if updatedPlan.Ingress.ExpectedCount != 0 {
		t.Errorf("expected ingress count was not 0")
	}
	found := false
	for _, w := range updatedPlan.Storage.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new node")
	}
}

func TestAddAllRolesPlanIsUpdated(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: fakeRunnerExplainer(nil),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Ingress: OptionalNodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Storage: OptionalNodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	updatedPlan, err := e.AddNode(originalPlan, newWorker, []string{"worker", "ingress", "storage"})
	if err != nil {
		t.Errorf("unexpected error while adding worker: %v", err)
	}
	if updatedPlan.Worker.ExpectedCount != 2 {
		t.Errorf("expected worker count was not incremented")
	}
	if updatedPlan.Ingress.ExpectedCount != 2 {
		t.Errorf("expected ingress count was not incremented")
	}
	if updatedPlan.Storage.ExpectedCount != 2 {
		t.Errorf("expected storage count was not incremented")
	}
	found := false
	for _, w := range updatedPlan.Worker.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new worker node")
	}
	found = false
	for _, w := range updatedPlan.Ingress.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new ingress node")
	}
	found = false
	for _, w := range updatedPlan.Storage.Nodes {
		if w.Equal(newWorker) {
			found = true
		}
	}
	if !found {
		t.Errorf("the updated plan does not include the new storage node")
	}
}

func TestAddWorkerPlanNotUpdatedAfterFailure(t *testing.T) {
	e := ansibleExecutor{
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: fakeRunnerExplainer(errors.New("exec error")),
		certsDir:               mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	updatedPlan, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if err == nil {
		t.Errorf("expected an error, but didn't get one")
	}
	if updatedPlan != nil {
		t.Error("plan was updated, even though adding worker failed")
	}
}

func TestAddWorkerRestartServicesEnabled(t *testing.T) {
	fakeRunner := fakeRunner{}
	e := ansibleExecutor{
		certsDir:            mustGetTempDir(t),
		options:             ExecutorOptions{RestartServices: true, RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: func(explain.AnsibleEventExplainer, io.Writer) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
			return &fakeRunner, &explain.AnsibleEventStreamExplainer{}, nil
		},
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	_, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if err != nil {
		t.Errorf("unexpected error")
	}

	if !fakeRunner.incomingCatalog.ForceProxyRestart {
		t.Errorf("missing restart flag for service kube-proxy")
	}

	if !fakeRunner.incomingCatalog.ForceKubeletRestart {
		t.Errorf("missing restart flag for service kubelet")
	}

	if !fakeRunner.incomingCatalog.ForceCalicoNodeRestart {
		t.Errorf("missing restart flag for service calico-node")
	}

	if !fakeRunner.incomingCatalog.ForceDockerRestart {
		t.Errorf("missing restart flag for service docker")
	}
}

func TestAddWorkerHostsFilesDNSEnabled(t *testing.T) {
	fakeRunner := fakeRunner{}
	e := ansibleExecutor{
		options:             ExecutorOptions{RunsDirectory: mustGetTempDir(t)},
		stdout:              ioutil.Discard,
		consoleOutputFormat: ansible.RawFormat,
		pki: &fakePKI{
			caExists: true,
		},
		runnerExplainerFactory: func(explain.AnsibleEventExplainer, io.Writer) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
			return &fakeRunner, &explain.AnsibleEventStreamExplainer{}, nil
		},
		certsDir: mustGetTempDir(t),
	}
	originalPlan := &Plan{
		Master: MasterNodeGroup{
			Nodes: []Node{{InternalIP: "10.10.2.20"}},
		},
		Worker: NodeGroup{
			ExpectedCount: 1,
			Nodes: []Node{
				{
					Host: "existingWorker",
				},
			},
		},
		Cluster: Cluster{
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/16",
				UpdateHostsFiles: true,
			},
		},
	}
	newWorker := Node{
		Host: "test",
	}
	_, err := e.AddNode(originalPlan, newWorker, []string{"worker"})
	if err != nil {
		t.Errorf("unexpected error")
	}
	expectedPlaybook := "hosts.yaml"
	found := false
	for _, p := range fakeRunner.allNodesPlaybooks {
		if p == expectedPlaybook {
			found = true
		}
	}
	if !found {
		t.Errorf("expected playbook %s was not run during add-worker. The following plays ran: %v", expectedPlaybook, fakeRunner.allNodesPlaybooks)
	}
}

//// Fakes for testing
type fakePKI struct {
	caExists               bool
	nodeCertExists         bool
	err                    error
	generateCACalled       bool
	generateNodeCertCalled bool
}

func (f *fakePKI) CertificateAuthorityExists() (bool, error)     { return f.caExists, f.err }
func (f *fakePKI) NodeCertificateExists(node Node) (bool, error) { return f.nodeCertExists, f.err }
func (f *fakePKI) GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error {
	f.generateNodeCertCalled = true
	return f.err
}
func (f *fakePKI) GetClusterCA() (*tls.CA, error) { return nil, f.err }
func (f *fakePKI) GenerateClusterCA(p *Plan) (*tls.CA, error) {
	f.generateCACalled = true
	return nil, f.err
}
func (f *fakePKI) GenerateClusterCertificates(p *Plan, ca *tls.CA) error { return f.err }
func (f *fakePKI) GenerateCertificate(name string, validityPeriod string, commonName string, subjectAlternateNames []string, organizations []string, ca *tls.CA, overwrite bool) (bool, error) {
	return false, f.err
}

type fakeRunner struct {
	eventChan         chan ansible.Event
	err               error
	incomingCatalog   ansible.ClusterCatalog
	allNodesPlaybooks []string
}

func (f *fakeRunner) StartPlaybook(playbookFile string, inventory ansible.Inventory, cc ansible.ClusterCatalog) (<-chan ansible.Event, error) {
	f.allNodesPlaybooks = append(f.allNodesPlaybooks, playbookFile)
	return f.eventChan, f.err
}
func (f *fakeRunner) WaitPlaybook() error { return f.err }
func (f *fakeRunner) StartPlaybookOnNode(playbookFile string, inventory ansible.Inventory, cc ansible.ClusterCatalog, node ...string) (<-chan ansible.Event, error) {
	f.incomingCatalog = cc
	return f.eventChan, f.err
}

func fakeRunnerExplainer(execError error) func(explain.AnsibleEventExplainer, io.Writer) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
	return func(explain.AnsibleEventExplainer, io.Writer) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
		return &fakeRunner{err: execError}, &explain.AnsibleEventStreamExplainer{}, nil
	}
}
