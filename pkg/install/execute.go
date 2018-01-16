package install

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/install/explain"
	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/apprenda/kismatic/pkg/util"
)

// The PreFlightExecutor will run pre-flight checks against the
// environment defined in the plan file
type PreFlightExecutor interface {
	RunPreFlightCheck(*Plan) error
	RunNewWorkerPreFlightCheck(Plan, Node) error
	RunUpgradePreFlightCheck(*Plan, ListableNode) error
}

// The Executor will carry out the installation plan
type Executor interface {
	PreFlightExecutor
	Install(p *Plan) error
	GenerateCertificates(p *Plan, useExistingCA bool) error
	RunSmokeTest(*Plan) error
	AddWorker(*Plan, Node) (*Plan, error)
	RunPlay(string, *Plan) error
	AddVolume(*Plan, StorageVolume) error
	DeleteVolume(*Plan, string) error
	UpgradeNodes(plan Plan, nodesToUpgrade []ListableNode, onlineUpgrade bool, maxParallelWorkers int) error
	ValidateControlPlane(plan Plan) error
	UpgradeClusterServices(plan Plan) error
}

// DiagnosticsExecutor will run diagnostics on the nodes after an install
type DiagnosticsExecutor interface {
	DiagnoseNodes(plan Plan) error
}

// ExecutorOptions are used to configure the executor
type ExecutorOptions struct {
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
	// DiagnosticsDirecty is where the doDiagnostics information about the cluster will be dumped
	DiagnosticsDirecty string
	// DryRun determines if the executor should actually run the task
	DryRun bool
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
		CACsr: filepath.Join(ansibleDir, "playbooks", "tls", "ca-csr.json"),
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

// NewDiagnosticsExecutor returns an executor for running preflight
func NewDiagnosticsExecutor(stdout io.Writer, errOut io.Writer, options ExecutorOptions) (DiagnosticsExecutor, error) {
	ansibleDir := "ansible"
	if options.RunsDirectory == "" {
		options.RunsDirectory = "./runs"
	}
	if options.DiagnosticsDirecty == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Could not get working directory: %v", err)
		}
		options.DiagnosticsDirecty = filepath.Join(wd, "diagnostics")
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
	if ae.options.DryRun {
		return nil
	}
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

// GenerateCertificatesprivate generates keys and certificates for the cluster, if needed
func (ae *ansibleExecutor) GenerateCertificates(p *Plan, useExistingCA bool) error {
	if err := os.MkdirAll(ae.certsDir, 0777); err != nil {
		return fmt.Errorf("error creating directory %s for storing TLS assets: %v", ae.certsDir, err)
	}

	// Generate cluster Certificate Authority
	util.PrintHeader(ae.stdout, "Configuring Certificates", '=')

	var caCert *tls.CA
	var err error
	if useExistingCA {
		exists, err := ae.pki.CertificateAuthorityExists()
		if err != nil {
			return fmt.Errorf("error checking if CA exists: %v", err)
		}
		if !exists {
			return errors.New("The Certificate Authority is required, but it was not found.")
		}
		caCert, err = ae.pki.GetClusterCA()
		if err != nil {
			return fmt.Errorf("error reading CA certificate: %v", err)
		}

	} else {
		caCert, err = ae.pki.GenerateClusterCA(p)
		if err != nil {
			return fmt.Errorf("error generating CA for the cluster: %v", err)
		}
	}

	// Generate node and user certificates
	err = ae.pki.GenerateClusterCertificates(p, caCert)
	if err != nil {
		return fmt.Errorf("error generating certificates for the cluster: %v", err)
	}

	util.PrettyPrintOk(ae.stdout, "Cluster certificates can be found in the %q directory", ae.options.GeneratedAssetsDirectory)
	return nil
}

// Install the cluster according to the installation plan
func (ae *ansibleExecutor) Install(p *Plan) error {
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
		explainer:      ae.defaultExplainer(),
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
		explainer:      ae.defaultExplainer(),
		plan:           *p,
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
	}
	util.PrintHeader(ae.stdout, "Running Smoke Test", '=')
	return ae.execute(t)
}

// RunPreflightCheck against the nodes defined in the plan
func (ae *ansibleExecutor) RunPreFlightCheck(p *Plan) error {
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	cc, err = setPreflightOptions(*p, *cc)
	if err != nil {
		return err
	}
	t := task{
		name:           "preflight",
		playbook:       "preflight.yaml",
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
		explainer:      ae.preflightExplainer(),
		plan:           *p,
	}
	return ae.execute(t)
}

// RunNewWorkerPreFlightCheck runs the preflight checks against a new worker node
func (ae *ansibleExecutor) RunNewWorkerPreFlightCheck(p Plan, node Node) error {
	cc, err := ae.buildClusterCatalog(&p)
	if err != nil {
		return err
	}
	cc, err = setPreflightOptions(p, *cc)
	if err != nil {
		return err
	}
	t := task{
		name:           "copy-inspector",
		playbook:       "copy-inspector.yaml",
		inventory:      buildInventoryFromPlan(&p),
		clusterCatalog: *cc,
		explainer:      ae.preflightExplainer(),
		plan:           p,
	}
	if err := ae.execute(t); err != nil {
		return err
	}
	p.Worker.ExpectedCount++
	p.Worker.Nodes = append(p.Worker.Nodes, node)
	t = task{
		name:           "add-worker-preflight",
		playbook:       "preflight.yaml",
		inventory:      buildInventoryFromPlan(&p),
		clusterCatalog: *cc,
		explainer:      ae.preflightExplainer(),
		plan:           p,
		limit:          []string{node.Host},
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) RunUpgradePreFlightCheck(p *Plan, node ListableNode) error {
	inventory := buildInventoryFromPlan(p)
	cc, err := ae.buildClusterCatalog(p)
	if err != nil {
		return err
	}
	cc, err = setPreflightOptions(*p, *cc)
	if err != nil {
		return err
	}
	t := task{
		name:           "copy-inspector",
		playbook:       "copy-inspector.yaml",
		inventory:      buildInventoryFromPlan(p),
		clusterCatalog: *cc,
		explainer:      ae.preflightExplainer(),
		plan:           *p,
	}
	if err := ae.execute(t); err != nil {
		return err
	}
	t = task{
		name:           "upgrade-preflight",
		playbook:       "upgrade-preflight.yaml",
		explainer:      ae.preflightExplainer(),
		plan:           *p,
		inventory:      inventory,
		clusterCatalog: *cc,
		limit:          []string{node.Node.Host},
	}
	return ae.execute(t)
}

func setPreflightOptions(p Plan, cc ansible.ClusterCatalog) (*ansible.ClusterCatalog, error) {
	cc.KismaticPreflightCheckerLinux = filepath.Join("inspector", "linux", "amd64", "kismatic-inspector")
	cc.EnablePackageInstallation = !p.Cluster.DisablePackageInstallation
	return &cc, nil
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
		explainer:      ae.defaultExplainer(),
		plan:           *p,
	}
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
	cc.VolumeReclaimPolicy = volume.ReclaimPolicy
	cc.VolumeAccessModes = volume.AccessModes

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
		explainer:      ae.defaultExplainer(),
	}
	util.PrintHeader(ae.stdout, "Add Persistent Storage Volume", '=')
	return ae.execute(t)
}

func (ae *ansibleExecutor) DeleteVolume(plan *Plan, name string) error {
	cc, err := ae.buildClusterCatalog(plan)
	if err != nil {
		return err
	}
	// Add storage related vars
	cc.VolumeName = name
	cc.VolumeMount = "/"

	t := task{
		name:           "delete-volume",
		playbook:       "volume-delete.yaml",
		plan:           *plan,
		inventory:      buildInventoryFromPlan(plan),
		clusterCatalog: *cc,
		explainer:      ae.defaultExplainer(),
	}
	util.PrintHeader(ae.stdout, "Delete Persistent Storage Volume", '=')
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
func (ae *ansibleExecutor) UpgradeNodes(plan Plan, nodesToUpgrade []ListableNode, onlineUpgrade bool, maxParallelWorkers int) error {
	// Nodes can have multiple roles. For this reason, we need to keep track of which nodes
	// have been upgraded to avoid re-upgrading them.
	upgradedNodes := map[string]bool{}
	// Upgrade etcd nodes
	for _, nodeToUpgrade := range nodesToUpgrade {
		for _, role := range nodeToUpgrade.Roles {
			if role == "etcd" {
				node := nodeToUpgrade
				if err := ae.upgradeNodes(plan, onlineUpgrade, node); err != nil {
					return fmt.Errorf("error upgrading node %q: %v", node.Node.Host, err)
				}
				upgradedNodes[node.Node.IP] = true
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
				node := nodeToUpgrade
				if err := ae.upgradeNodes(plan, onlineUpgrade, node); err != nil {
					return fmt.Errorf("error upgrading node %q: %v", node.Node.Host, err)
				}
				upgradedNodes[node.Node.IP] = true
				break
			}
		}
	}

	var limitNodes []ListableNode
	// Upgrade the rest of the nodes
	for n, nodeToUpgrade := range nodesToUpgrade {
		if upgradedNodes[nodeToUpgrade.Node.IP] == true {
			continue
		}
		for _, role := range nodeToUpgrade.Roles {
			if role != "etcd" && role != "master" {
				node := nodeToUpgrade
				limitNodes = append(limitNodes, node)
				// don't forget to run the remaining nodes if its < maxParallelWorkers
				if len(limitNodes) == maxParallelWorkers || n == len(nodesToUpgrade)-1 {
					if err := ae.upgradeNodes(plan, onlineUpgrade, limitNodes...); err != nil {
						return fmt.Errorf("error upgrading node %q: %v", node.Node.Host, err)
					}
					// empty the slice
					limitNodes = limitNodes[:0]
				}
				upgradedNodes[node.Node.IP] = true
				break
			}
		}
	}
	return nil
}

func (ae *ansibleExecutor) upgradeNodes(plan Plan, onlineUpgrade bool, nodes ...ListableNode) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	cc.OnlineUpgrade = onlineUpgrade
	var limit []string
	nodeRoles := make(map[string][]string)
	for _, node := range nodes {
		limit = append(limit, node.Node.Host)
		nodeRoles[node.Node.Host] = node.Roles
	}
	t := task{
		name:           "upgrade-nodes",
		playbook:       "upgrade-nodes.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      ae.defaultExplainer(),
		limit:          limit,
	}
	if len(limit) == 1 {
		util.PrintHeader(ae.stdout, fmt.Sprintf("Upgrade Node: %s %s", limit, nodes[0].Roles), '=')
	} else { // print the roles for multiple nodes
		util.PrintHeader(ae.stdout, "Upgrade Nodes:", '=')
		util.PrintTable(ae.stdout, nodeRoles)
	}
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
		explainer:      ae.defaultExplainer(),
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
		explainer:      ae.defaultExplainer(),
	}
	return ae.execute(t)
}

func (ae *ansibleExecutor) DiagnoseNodes(plan Plan) error {
	inventory := buildInventoryFromPlan(&plan)
	cc, err := ae.buildClusterCatalog(&plan)
	if err != nil {
		return err
	}
	// dateTime will be appended to the diagnostics directory
	now := time.Now().Format("2006-01-02-15-04-05")
	cc.DiagnosticsDirectory = filepath.Join(ae.options.DiagnosticsDirecty, now)
	cc.DiagnosticsDateTime = now
	t := task{
		name:           "diagnose",
		playbook:       "diagnose-nodes.yaml",
		inventory:      inventory,
		clusterCatalog: *cc,
		plan:           plan,
		explainer:      ae.defaultExplainer(),
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
		ClusterName:                  p.Cluster.Name,
		AdminPassword:                p.Cluster.AdminPassword,
		TLSDirectory:                 tlsDir,
		ServicesCIDR:                 p.Cluster.Networking.ServiceCIDRBlock,
		PodCIDR:                      p.Cluster.Networking.PodCIDRBlock,
		DNSServiceIP:                 dnsIP,
		EnableModifyHosts:            p.Cluster.Networking.UpdateHostsFiles,
		EnablePackageInstallation:    !p.Cluster.DisablePackageInstallation,
		KuberangPath:                 filepath.Join("kuberang", "linux", "amd64", "kuberang"),
		DisconnectedInstallation:     p.Cluster.DisconnectedInstallation,
		HTTPProxy:                    p.Cluster.Networking.HTTPProxy,
		HTTPSProxy:                   p.Cluster.Networking.HTTPSProxy,
		TargetVersion:                KismaticVersion.String(),
		APIServerOptions:             p.Cluster.APIServerOptions.Overrides,
		KubeControllerManagerOptions: p.Cluster.KubeControllerManagerOptions.Overrides,
		KubeSchedulerOptions:         p.Cluster.KubeSchedulerOptions.Overrides,
		KubeProxyOptions:             p.Cluster.KubeProxyOptions.Overrides,
		KubeletOptions:               p.Cluster.KubeletOptions.Overrides,
	}

	cc.NoProxy = p.AllAddresses()
	if p.Cluster.Networking.NoProxy != "" {
		cc.NoProxy = cc.NoProxy + "," + p.Cluster.Networking.NoProxy
	}

	cc.LocalKubeconfigDirectory = filepath.Join(ae.options.GeneratedAssetsDirectory, "kubeconfig")
	// absolute path required for ansible
	generatedDir, err := filepath.Abs(filepath.Join(ae.options.GeneratedAssetsDirectory, "kubeconfig"))
	if err != nil {
		return nil, fmt.Errorf("failed to determine absolute path to %s: %v", filepath.Join(ae.options.GeneratedAssetsDirectory, "kubeconfig"), err)
	}
	cc.LocalKubeconfigDirectory = generatedDir

	// Setup FQDN or default to first master
	if p.Master.LoadBalancedFQDN != "" {
		cc.LoadBalancedFQDN = p.Master.LoadBalancedFQDN
	} else {
		cc.LoadBalancedFQDN = p.Master.Nodes[0].InternalIP
	}

	if p.PrivateRegistryProvided() {
		cc.ConfigureDockerWithPrivateRegistry = true
		cc.DockerRegistryServer = p.DockerRegistry.Server
		cc.DockerRegistryCAPath = p.DockerRegistry.CAPath
		cc.DockerRegistryUsername = p.DockerRegistry.Username
		cc.DockerRegistryPassword = p.DockerRegistry.Password
	}

	// Setup docker options
	cc.Docker.Enabled = !p.Docker.Disable
	cc.Docker.Logs.Driver = p.Docker.Logs.Driver
	cc.Docker.Logs.Opts = p.Docker.Logs.Opts

	cc.Docker.Storage.DirectLVM.Enabled = p.Docker.Storage.DirectLVM.Enabled
	if cc.Docker.Storage.DirectLVM.Enabled {
		cc.Docker.Storage.DirectLVM.BlockDevice = p.Docker.Storage.DirectLVM.BlockDevice
		cc.Docker.Storage.DirectLVM.EnableDeferredDeletion = p.Docker.Storage.DirectLVM.EnableDeferredDeletion
	}
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

	cc.CloudProvider = p.Cluster.CloudProvider.Provider
	cc.CloudConfig = p.Cluster.CloudProvider.Config

	// add_ons
	cc.RunPodValidation = p.NetworkConfigured()
	// CNI
	if p.AddOns.CNI != nil && !p.AddOns.CNI.Disable {
		cc.CNI.Enabled = true
		cc.CNI.Provider = p.AddOns.CNI.Provider
		cc.CNI.Options.Calico.Mode = p.AddOns.CNI.Options.Calico.Mode
		cc.CNI.Options.Calico.LogLevel = p.AddOns.CNI.Options.Calico.LogLevel
		cc.CNI.Options.Calico.WorkloadMTU = p.AddOns.CNI.Options.Calico.WorkloadMTU
		cc.CNI.Options.Calico.FelixInputMTU = p.AddOns.CNI.Options.Calico.FelixInputMTU

		if cc.CNI.Provider == cniProviderContiv {
			cc.InsecureNetworkingEtcd = true
		}
	}

	// DNS
	cc.DNS.Enabled = !p.AddOns.DNS.Disable
	cc.DNS.Provider = p.AddOns.DNS.Provider

	// heapster
	if p.AddOns.HeapsterMonitoring != nil && !p.AddOns.HeapsterMonitoring.Disable {
		cc.Heapster.Enabled = true
		cc.Heapster.Options.Heapster.Replicas = p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas
		cc.Heapster.Options.Heapster.ServiceType = p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType
		cc.Heapster.Options.Heapster.Sink = p.AddOns.HeapsterMonitoring.Options.Heapster.Sink
		cc.Heapster.Options.InfluxDB.PVCName = p.AddOns.HeapsterMonitoring.Options.InfluxDB.PVCName
	}

	// dashboard
	cc.Dashboard.Enabled = true
	if p.AddOns.Dashboard != nil && p.AddOns.Dashboard.Disable {
		cc.Dashboard.Enabled = false
	}

	// package_manager
	if !p.AddOns.PackageManager.Disable {
		// Currently only helm is supported
		switch p.AddOns.PackageManager.Provider {
		case "helm":
			cc.Helm.Enabled = true
		default:
			cc.Helm.Enabled = true
		}
	}

	cc.Rescheduler.Enabled = !p.AddOns.Rescheduler.Disable

	// merge node labels
	// cannot use inventory file because nodes share roles
	// set it to a map[host][]key=value
	cc.NodeLabels = make(map[string][]string)
	for _, n := range p.getAllNodes() {
		if val, ok := cc.NodeLabels[n.Host]; ok {
			cc.NodeLabels[n.Host] = append(val, keyValueList(n.Labels)...)
		} else {
			cc.NodeLabels[n.Host] = keyValueList(n.Labels)
		}
	}

	// setup kubelet node overrides
	cc.KubeletNodeOptions = make(map[string]map[string]string)
	for _, n := range p.GetUniqueNodes() {
		cc.KubeletNodeOptions[n.Host] = n.KubeletOptions.Overrides
	}

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

func (ae *ansibleExecutor) ansibleRunnerWithExplainer(explainer explain.AnsibleEventExplainer, ansibleLog io.Writer, runDirectory string) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
	if ae.runnerExplainerFactory != nil {
		return ae.runnerExplainerFactory(explainer, ansibleLog)
	}

	// Setup sink for ansible stdout
	var ansibleOut io.Writer
	switch ae.consoleOutputFormat {
	case ansible.JSONLinesFormat:
		ansibleOut = timestampWriter(ansibleLog)
	case ansible.RawFormat:
		ansibleOut = io.MultiWriter(ae.stdout, timestampWriter(ansibleLog))
	}

	// Send stdout and stderr to ansibleOut
	runner, err := ansible.NewRunner(ansibleOut, ansibleOut, ae.ansibleDir, runDirectory)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating ansible runner: %v", err)
	}

	streamExplainer := &explain.AnsibleEventStreamExplainer{
		EventExplainer: explainer,
	}

	return runner, streamExplainer, nil
}

func (ae *ansibleExecutor) defaultExplainer() explain.AnsibleEventExplainer {
	var out io.Writer
	switch ae.consoleOutputFormat {
	case ansible.JSONLinesFormat:
		out = ae.stdout
	case ansible.RawFormat:
		out = ioutil.Discard
	}
	return explain.DefaultExplainer(ae.options.Verbose, out)
}

func (ae *ansibleExecutor) preflightExplainer() explain.AnsibleEventExplainer {
	var out io.Writer
	switch ae.consoleOutputFormat {
	case ansible.JSONLinesFormat:
		out = ae.stdout
	case ansible.RawFormat:
		out = ioutil.Discard
	}
	return explain.PreflightExplainer(ae.options.Verbose, out)
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
		lr := util.NewLineReader(r, 64*1024)
		var (
			err  error
			line []byte
		)
		for err == nil {
			line, err = lr.Read()
			fmt.Fprintf(out, "%s - %s\n", time.Now().UTC().Format("2006-01-02 15:04:05.000-0700"), string(line))
		}
		if err != io.EOF {
			fmt.Printf("Error timestamping ansible logs: %v", err)
		}
	}(pr)
	return pw
}

// key=value slice
func keyValueList(in map[string]string) []string {
	pairs := make([]string, 0, len(in))
	for k, v := range in {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return pairs
}
