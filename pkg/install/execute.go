package install

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"strings"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/install/explain"
	"github.com/apprenda/kismatic/pkg/util"
)

// The PreFlightExecutor will run pre-flight checks against the
// environment defined in the plan file
type PreFlightExecutor interface {
	RunPreFlightCheck(*Plan) error
	RunUpgradePreFlightCheck(*Plan) error
}

// The Executor will carry out the installation plan
type Executor interface {
	PreFlightExecutor
	Install(p *Plan) error
	RunSmokeTest(*Plan) error
	AddWorker(*Plan, Node) (*Plan, error)
	RunPlay(string, *Plan) error
	AddVolume(*Plan, StorageVolume) error
	UpgradeNodes(plan Plan, nodesToUpgrade []ListableNode) error
	ValidateControlPlane(plan Plan) error
	UpgradeDockerRegistry(plan Plan) error
	UpgradeClusterServices(plan Plan) error
}

// ExecutorOptions are used to configure the executor
type ExecutorOptions struct {
	// SkipCAGeneration determines whether the Certificate Authority should
	// be generated. If false, an existing CA file must exist.
	SkipCAGeneration bool
	// GeneratedAssetsDirectory is the location where generated assets
	// are to be stored
	GeneratedAssetsDirectory string
	// RestartServices determines whether the cluster services should be
	// restarted during the installation.
	RestartServices bool
	// OutputFormat sets the format of the executor
	OutputFormat string
	// Verbose output from the executor
	Verbose bool
	// RunsDirectory is where information about installation runs is kept
	RunsDirectory string
}

// NewExecutor returns an executor for performing installations according to the installation plan.
func NewExecutor(stdout io.Writer, errOut io.Writer, options ExecutorOptions) (Executor, error) {
	ansibleDir := "ansible"
	if options.GeneratedAssetsDirectory == "" {
		return nil, fmt.Errorf("GeneratedAssetsDirectory option cannot be empty")
	}
	if options.RunsDirectory == "" {
		options.RunsDirectory = "./runs"
	}

	// Setup the console output format
	var outFormat ansible.OutputFormat
	switch options.OutputFormat {
	case "raw":
		outFormat = ansible.RawFormat
	case "simple":
		outFormat = ansible.JSONLinesFormat
	default:
		return nil, fmt.Errorf("Output format %q is not supported", options.OutputFormat)
	}
	certsDir := filepath.Join(options.GeneratedAssetsDirectory, "keys")
	pki := &LocalPKI{
		CACsr:                   filepath.Join(ansibleDir, "playbooks", "tls", "ca-csr.json"),
		CAConfigFile:            filepath.Join(ansibleDir, "playbooks", "tls", "ca-config.json"),
		CASigningProfile:        "kubernetes",
		GeneratedCertsDirectory: certsDir,
		Log: stdout,
	}
	return &ansibleExecutor{
		options:             options,
		stdout:              stdout,
		consoleOutputFormat: outFormat,
		ansibleDir:          ansibleDir,
		certsDir:            certsDir,
		pki:                 pki,
	}, nil
}

// NewPreFlightExecutor returns an executor for running preflight
func NewPreFlightExecutor(stdout io.Writer, errOut io.Writer, options ExecutorOptions) (PreFlightExecutor, error) {
	ansibleDir := "ansible"
	if options.RunsDirectory == "" {
		options.RunsDirectory = "./runs"
	}
	// Setup the console output format
	var outFormat ansible.OutputFormat
	switch options.OutputFormat {
	case "raw":
		outFormat = ansible.RawFormat
	case "simple":
		outFormat = ansible.JSONLinesFormat
	default:
		return nil, fmt.Errorf("Output format %q is not supported", options.OutputFormat)
	}

	return &ansibleExecutor{
		options:             options,
		stdout:              stdout,
		consoleOutputFormat: outFormat,
		ansibleDir:          ansibleDir,
	}, nil
}

type ansibleExecutor struct {
	options             ExecutorOptions
	stdout              io.Writer
	consoleOutputFormat ansible.OutputFormat
	ansibleDir          string
	certsDir            string
	pki                 PKI

	// Hook for testing purposes.. default implementation is used at runtime
	runnerExplainerFactory func(explain.AnsibleEventExplainer, io.Writer) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error)
}

type task struct {
	// name of the task used for the runs dir
	name string
	// the inventory of nodes to use
	inventory ansible.Inventory
	// the cluster catalog to use
	clusterCatalog ansible.ClusterCatalog
	// the playbook filename
	playbook string
	// the explainer to use
	explainer explain.AnsibleEventExplainer
	// the plan
	plan Plan
	// run the task on specific nodes
	limit []string
}

// execute will run the given task, and setup all what's needed for us to run ansible.
func (ae *ansibleExecutor) execute(t task) error {
	runDirectory, err := ae.createRunDirectory(t.name)
	if err != nil {
		return fmt.Errorf("error creating working directory for %q: %v", t.name, err)
	}
	// Save the plan file that was used for this execution
	fp := FilePlanner{
		File: filepath.Join(runDirectory, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(&t.plan); err != nil {
		return fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}
	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}
	runner, explainer, err := ae.ansibleRunnerWithExplainer(t.explainer, ansibleLogFile, runDirectory)
	if err != nil {
		return err
	}

	// Start running ansible with the given playbook
	var eventStream <-chan ansible.Event
	if t.limit != nil && len(t.limit) != 0 {
		eventStream, err = runner.StartPlaybookOnNode(t.playbook, t.inventory, t.clusterCatalog, t.limit...)
	} else {
		eventStream, err = runner.StartPlaybook(t.playbook, t.inventory, t.clusterCatalog)
	}
	if err != nil {
		return fmt.Errorf("error running ansible playbook: %v", err)
	}
	// Ansible blocks until explainer starts reading from stream. Start
	// explainer in a separate go routine
	go explainer.Explain(eventStream)

	// Wait until ansible exits
	if err = runner.WaitPlaybook(); err != nil {
		return fmt.Errorf("error running playbook: %v", err)
	}
	return nil
}

// Install the cluster according to the installation plan
func (ae *ansibleExecutor) Install(p *Plan) error {
	// Generate private keys and certificates for the cluster
	if err := ae.generateTLSAssets(p); err != nil {
		return err
	}
	// Build the ansible inventory
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	t := task{
		name:           "apply",
		playbook:       "kubernetes.yaml",
		plan:           *p,
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
		explainer:      &explain.DefaultEventExplainer{},
	}
	util.PrintHeader(ae.stdout, "Installing Cluster", '=')
	return ae.execute(t)
}

func (ae *ansibleExecutor) RunSmokeTest(p *Plan) error {
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	t := task{
		name:           "smoketest",
		playbook:       "smoketest.yaml",
		explainer:      &explain.DefaultEventExplainer{},
		plan:           *p,
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
	}
	util.PrintHeader(ae.stdout, "Running Smoke Test", '=')
	return ae.execute(t)
}

// RunPreflightCheck against the nodes defined in the plan
func (ae *ansibleExecutor) RunPreFlightCheck(p *Plan) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	cc.KismaticPreflightCheckerLinux = filepath.Join("inspector", "linux", "amd64", "kismatic-inspector")
	cc.KismaticPreflightCheckerLocal = filepath.Join(pwd, "ansible", "playbooks", "inspector", runtime.GOOS, runtime.GOARCH, "kismatic-inspector")
	cc.EnablePackageInstallation = p.Cluster.AllowPackageInstallation

	t := task{
		name:           "preflight",
		playbook:       "preflight.yaml",
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
		explainer: &explain.PreflightEventExplainer{
			DefaultExplainer: &explain.DefaultEventExplainer{},
		},
		plan: *p,
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) RunUpgradePreFlightCheck(p *Plan) error {
	inventory := buildInventoryFromPlan(p)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	cc.KismaticPreflightCheckerLinux = filepath.Join("inspector", "linux", "amd64", "kismatic-inspector")
	cc.KismaticPreflightCheckerLocal = filepath.Join(pwd, "ansible", "playbooks", "inspector", runtime.GOOS, runtime.GOARCH, "kismatic-inspector")
	cc.EnablePackageInstallation = p.Cluster.AllowPackageInstallation
	t := task{
		name:     "upgrade-preflight",
		playbook: "upgrade-preflight.yaml",
		explainer: &explain.PreflightEventExplainer{
			DefaultExplainer: &explain.DefaultEventExplainer{},
		},
		plan:           *p,
		inventory:      inventory,
		clusterCatalog: *cc,
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) RunPlay(playName string, p *Plan) error {
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	t := task{
		name:           "step",
		playbook:       playName,
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
		explainer:      &explain.DefaultEventExplainer{},
		plan:           *p,
	}
	util.PrintHeader(ae.stdout, "Running Task", '=')
	return ae.execute(t)
}

func (ae *ansibleExecutor) AddVolume(plan *Plan, volume StorageVolume) error {
	// Validate that there are enough storage nodes to satisfy the request
	nodesRequired := volume.ReplicateCount * volume.DistributionCount
	if nodesRequired > len(plan.Storage.Nodes) {
		return fmt.Errorf("the requested volume configuration requires %d storage nodes, but the cluster only has %d.", nodesRequired, len(plan.Storage.Nodes))
	}

	cc, err := ae.buildClusterCatalog(plan)
	if err != nil {
		return err
	}
	// Add storage related vars
	cc.VolumeName = volume.Name
	cc.VolumeReplicaCount = volume.ReplicateCount
	cc.VolumeDistributionCount = volume.DistributionCount
	cc.VolumeStorageClass = volume.StorageClass
	cc.VolumeQuotaGB = volume.SizeGB
	cc.VolumeQuotaBytes = volume.SizeGB * (1 << (10 * 3))
	cc.VolumeMount = "/"

	// Allow nodes and pods to access volumes
	allowedNodes := plan.Master.Nodes
	allowedNodes = append(allowedNodes, plan.Worker.Nodes...)
	allowedNodes = append(allowedNodes, plan.Ingress.Nodes...)
	allowedNodes = append(allowedNodes, plan.Storage.Nodes...)

	allowed := volume.AllowAddresses
	allowed = append(allowed, plan.Cluster.Networking.PodCIDRBlock)
	for _, n := range allowedNodes {
		ip := n.IP
		if n.InternalIP != "" {
			ip = n.InternalIP
		}
		allowed = append(allowed, ip)
	}
	cc.VolumeAllowedIPs = strings.Join(allowed, ",")

	t := task{
		name:           "add-volume",
		playbook:       "volume-add.yaml",
		plan:           *plan,
		inventory:      buildInventoryFromPlan(plan),
		clusterCatalog: *cc,
		explainer:      &explain.DefaultEventExplainer{},
	}
	util.PrintHeader(ae.stdout, "Add Persistent Storage Volume", '=')
	return ae.execute(t)
}

// UpgradeNodes upgrades the nodes of the cluster in the following phases:
//   1. Etcd nodes
//   2. Master nodes
//   3. Worker nodes (regardless of specialization)
//
// When a node is being upgraded, all the components of the node are upgraded, regardless of
// which phase of the upgrade we are in. For example, when upgrading a node that is both an etcd and master,
// the etcd components and the master components will be upgraded when we are in the upgrade etcd nodes
// phase.
func (ae *ansibleExecutor) UpgradeNodes(plan Plan, nodesToUpgrade []ListableNode) error {
	// Nodes can have multiple roles. For this reason, we need to keep track of which nodes
	// have been upgraded to avoid re-upgrading them.
	upgradedNodes := map[string]bool{}
	// Upgrade etcd nodes
	for _, nodeToUpgrade := range nodesToUpgrade {
		for _, role := range nodeToUpgrade.Roles {
			if role == "etcd" {
				node := nodeToUpgrade.Node
				if err := ae.upgradeNode(plan, node); err != nil {
					return fmt.Errorf("error upgrading node %q: %v", node.Host, err)
				}
				upgradedNodes[node.IP] = true
				break
			}
		}
	}

	// Upgrade master nodes
	for _, nodeToUpgrade := range nodesToUpgrade {
		if upgradedNodes[nodeToUpgrade.Node.IP] == true {
			continue
		}
		for _, role := range nodeToUpgrade.Roles {
			if role == "master" {
				node := nodeToUpgrade.Node
				if err := ae.upgradeNode(plan, node); err != nil {
					return fmt.Errorf("error upgrading node %q: %v", node.Host, err)
				}
				upgradedNodes[node.IP] = true
				break
			}
		}
	}

	// Upgrade the rest of the nodes
	for _, nodeToUpgrade := range nodesToUpgrade {
		if upgradedNodes[nodeToUpgrade.Node.IP] == true {
			continue
		}
		for _, role := range nodeToUpgrade.Roles {
			if role != "etcd" && role != "master" {
				node := nodeToUpgrade.Node
				if err := ae.upgradeNode(plan, node); err != nil {
					return fmt.Errorf("error upgrading node %q: %v", node.Host, err)
				}
				upgradedNodes[node.IP] = true
				break
			}
		}
	}
	return nil
}

func (ae *ansibleExecutor) upgradeNode(plan Plan, node Node) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	t := task{
		name:           "upgrade-nodes",
		playbook:       "upgrade-nodes.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      &explain.DefaultEventExplainer{},
		limit:          []string{node.Host},
	}
	util.PrintHeader(ae.stdout, fmt.Sprintf("Upgrade Node %q", node.Host), '=')
	return ae.execute(t)
}

func (ae *ansibleExecutor) ValidateControlPlane(plan Plan) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	t := task{
		name:           "validate-control-plane",
		playbook:       "validate-control-plane.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      &explain.DefaultEventExplainer{},
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) UpgradeDockerRegistry(plan Plan) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	t := task{
		name:           "upgrade-docker-registry",
		playbook:       "upgrade-docker-registry.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      &explain.DefaultEventExplainer{},
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) UpgradeClusterServices(plan Plan) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	t := task{
		name:           "upgrade-cluster-services",
		playbook:       "upgrade-cluster-services.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      &explain.DefaultEventExplainer{},
	}
	return ae.execute(t)
}

// creates the extra vars that are required for the installation playbook.
func (ae *ansibleExecutor) buildClusterCatalog(p *Plan) (*ansible.ClusterCatalog, error) {
	tlsDir, err := filepath.Abs(ae.certsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to determine absolute path to %s: %v", ae.certsDir, err)
	}

	dnsIP, err := getDNSServiceIP(p)
	if err != nil {
		return nil, fmt.Errorf("error getting DNS service IP: %v", err)
	}

	cc := ansible.ClusterCatalog{
		ClusterName:               p.Cluster.Name,
		AdminPassword:             p.Cluster.AdminPassword,
		TLSDirectory:              tlsDir,
		CalicoNetworkType:         p.Cluster.Networking.Type,
		ServicesCIDR:              p.Cluster.Networking.ServiceCIDRBlock,
		PodCIDR:                   p.Cluster.Networking.PodCIDRBlock,
		DNSServiceIP:              dnsIP,
		EnableModifyHosts:         p.Cluster.Networking.UpdateHostsFiles,
		EnableCalicoPolicy:        p.Cluster.Networking.PolicyEnabled,
		EnablePackageInstallation: p.Cluster.AllowPackageInstallation,
		KuberangPath:              filepath.Join("kuberang", "linux", "amd64", "kuberang"),
		DisconnectedInstallation:  p.Cluster.DisconnectedInstallation,
		TargetVersion:             AboutKismatic.ShortVersion.String(),
	}

	// Setup FQDN or default to first master
	if p.Master.LoadBalancedFQDN != "" {
		cc.LoadBalancedFQDN = p.Master.LoadBalancedFQDN
	} else {
		cc.LoadBalancedFQDN = p.Master.Nodes[0].InternalIP
	}

	if p.DockerRegistry.Address != "" {
		cc.EnableInternalDockerRegistry = false
		cc.EnablePrivateDockerRegistry = true
		cc.DockerCAPath = p.DockerRegistry.CAPath
		cc.DockerRegistryAddress = p.DockerRegistry.Address
		cc.DockerRegistryPort = strconv.Itoa(p.DockerRegistry.Port)
	} else if p.DockerRegistry.SetupInternal {
		cc.EnableInternalDockerRegistry = true
		cc.EnablePrivateDockerRegistry = true
		cc.DockerRegistryAddress = p.Master.Nodes[0].IP
		if p.Master.Nodes[0].InternalIP != "" {
			cc.DockerRegistryAddress = p.Master.Nodes[0].InternalIP
		}
		cc.DockerCAPath = tlsDir + "/ca.pem"
		cc.DockerRegistryPort = "8443"
	} // Else just use DockerHub

	if ae.options.RestartServices {
		cc.EnableRestart()
	}

	if p.Ingress.Nodes != nil && len(p.Ingress.Nodes) > 0 {
		cc.EnableConfigureIngress = true
	} else {
		cc.EnableConfigureIngress = false
	}

	for _, n := range p.NFS.Volumes {
		cc.NFSVolumes = append(cc.NFSVolumes, ansible.NFSVolume{
			Path: n.Path,
			Host: n.Host,
		})
	}
	cc.EnableGluster = p.Storage.Nodes != nil && len(p.Storage.Nodes) > 0

	return &cc, nil
}

func (ae *ansibleExecutor) createRunDirectory(runName string) (string, error) {
	start := time.Now()
	runDirectory := filepath.Join(ae.options.RunsDirectory, runName, start.Format("2006-01-02-15-04-05"))
	if err := os.MkdirAll(runDirectory, 0777); err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}
	return runDirectory, nil
}

func (ae *ansibleExecutor) generateTLSAssets(p *Plan) error {
	if err := os.MkdirAll(ae.certsDir, 0777); err != nil {
		return fmt.Errorf("error creating directory %s for storing TLS assets: %v", ae.certsDir, err)
	}

	// Generate cluster Certificate Authority
	util.PrintHeader(ae.stdout, "Configuring Certificates", '=')
	ca, err := ae.pki.GenerateClusterCA(p)
	if err != nil {
		return fmt.Errorf("error generating CA for the cluster: %v", err)
	}

	// Generate node and user certificates
	err = ae.pki.GenerateClusterCertificates(p, ca, []string{"admin"})
	if err != nil {
		return fmt.Errorf("error generating certificates for the cluster: %v", err)
	}
	util.PrettyPrintOk(ae.stdout, "Cluster certificates can be found in the %q directory", ae.options.GeneratedAssetsDirectory)
	return nil
}

func (ae *ansibleExecutor) ansibleRunnerWithExplainer(explainer explain.AnsibleEventExplainer, ansibleLog io.Writer, runDirectory string) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
	if ae.runnerExplainerFactory != nil {
		return ae.runnerExplainerFactory(explainer, ansibleLog)
	}

	// Setup sinks for explainer and ansible stdout
	var explainerOut, ansibleOut io.Writer
	switch ae.consoleOutputFormat {
	case ansible.JSONLinesFormat:
		explainerOut = ae.stdout
		ansibleOut = timestampWriter(ansibleLog)
	case ansible.RawFormat:
		explainerOut = ioutil.Discard
		ansibleOut = io.MultiWriter(ae.stdout, timestampWriter(ansibleLog))
	}

	// Send stdout and stderr to ansibleOut
	runner, err := ansible.NewRunner(ansibleOut, ansibleOut, ae.ansibleDir, runDirectory)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating ansible runner: %v", err)
	}

	streamExplainer := &explain.AnsibleEventStreamExplainer{
		Out:            explainerOut,
		Verbose:        ae.options.Verbose,
		EventExplainer: explainer,
	}

	return runner, streamExplainer, nil
}

func buildInventoryFromPlan(p *Plan) ansible.Inventory {
	etcdNodes := []ansible.Node{}
	for _, n := range p.Etcd.Nodes {
		etcdNodes = append(etcdNodes, installNodeToAnsibleNode(&n, &p.Cluster.SSH))
	}
	masterNodes := []ansible.Node{}
	for _, n := range p.Master.Nodes {
		masterNodes = append(masterNodes, installNodeToAnsibleNode(&n, &p.Cluster.SSH))
	}
	workerNodes := []ansible.Node{}
	for _, n := range p.Worker.Nodes {
		workerNodes = append(workerNodes, installNodeToAnsibleNode(&n, &p.Cluster.SSH))
	}
	ingressNodes := []ansible.Node{}
	if p.Ingress.Nodes != nil {
		for _, n := range p.Ingress.Nodes {
			ingressNodes = append(ingressNodes, installNodeToAnsibleNode(&n, &p.Cluster.SSH))
		}
	}
	storageNodes := []ansible.Node{}
	if p.Storage.Nodes != nil {
		for _, n := range p.Storage.Nodes {
			storageNodes = append(storageNodes, installNodeToAnsibleNode(&n, &p.Cluster.SSH))
		}
	}

	inventory := ansible.Inventory{
		Roles: []ansible.Role{
			{
				Name:  "etcd",
				Nodes: etcdNodes,
			},
			{
				Name:  "master",
				Nodes: masterNodes,
			},
			{
				Name:  "worker",
				Nodes: workerNodes,
			},
			{
				Name:  "ingress",
				Nodes: ingressNodes,
			},
			{
				Name:  "storage",
				Nodes: storageNodes,
			},
		},
	}

	return inventory
}

// Converts plan node to ansible node
func installNodeToAnsibleNode(n *Node, s *SSHConfig) ansible.Node {
	return ansible.Node{
		Host:          n.Host,
		PublicIP:      n.IP,
		InternalIP:    n.InternalIP,
		SSHPrivateKey: s.Key,
		SSHUser:       s.User,
		SSHPort:       s.Port,
	}
}

// Prepend each line of the incoming stream with a timestamp
func timestampWriter(out io.Writer) io.Writer {
	pr, pw := io.Pipe()
	go func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			fmt.Fprintf(out, "%s - %s\n", time.Now().UTC().Format("2006-01-02 15:04:05.000-0700"), s.Text())
		}
	}(pr)
	return pw
}

func findNodeWithIP(nodes []Node, ip string) (*Node, error) {
	for _, n := range nodes {
		if n.IP == ip {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("Node with IP %q not found", ip)
}
