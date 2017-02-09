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
}

// The Executor will carry out the installation plan
type Executor interface {
	PreFlightExecutor
	Install(p *Plan) error
	RunSmokeTest(*Plan) error
	AddWorker(*Plan, Node) (*Plan, error)
	RunTask(string, *Plan) error
	AddVolume(*Plan, StorageVolume) error
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
	// TODO: Is there a better way to handle this path to the ansible install dir?
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

// Install the cluster according to the installation plan
func (ae *ansibleExecutor) Install(p *Plan) error {
	runDirectory, err := ae.createRunDirectory("install")
	if err != nil {
		return fmt.Errorf("error creating working directory for installation: %v", err)
	}
	// Save the plan file that was used for this execution
	fp := FilePlanner{
		File: filepath.Join(runDirectory, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(p); err != nil {
		return fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}
	// Generate private keys and certificates for the cluster
	if err = ae.generateTLSAssets(p); err != nil {
		return err
	}
	// Build the ansible inventory
	inventory := buildInventoryFromPlan(p)

	cc, err := ae.buildInstallExtraVars(p)
	if err != nil {
		return err
	}
	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}
	// Run the installation playbook
	util.PrintHeader(ae.stdout, "Installing Cluster", '=')
	playbook := "kubernetes.yaml"
	eventExplainer := &explain.DefaultEventExplainer{}
	if err = ae.runPlaybookWithExplainer(playbook, eventExplainer, inventory, *cc, ansibleLogFile, runDirectory); err != nil {
		return err
	}
	return nil
}

// creates the extra vars that are required for the installation playbook.
func (ae *ansibleExecutor) buildInstallExtraVars(p *Plan) (*ansible.ClusterCatalog, error) {
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
		TargetVersion:             AboutKismatic.ShortVersion,
		DisconnectedInstallation:  p.Cluster.DisconnectedInstallation,
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

func (ae *ansibleExecutor) RunSmokeTest(p *Plan) error {
	runDirectory, err := ae.createRunDirectory("smoketest")
	if err != nil {
		return fmt.Errorf("error creating working directory for smoke test: %v", err)
	}

	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}

	cc, err := ae.buildInstallExtraVars(p)
	if err != nil {
		return err
	}
	inventory := buildInventoryFromPlan(p)

	// run the preflight playbook with preflight explainer
	util.PrintHeader(ae.stdout, "Running Smoke Test", '=')
	playbook := "smoketest.yaml"
	explainer := &explain.PreflightEventExplainer{
		DefaultExplainer: &explain.DefaultEventExplainer{},
	}
	if err = ae.runPlaybookWithExplainer(playbook, explainer, inventory, *cc, ansibleLogFile, runDirectory); err != nil {
		return fmt.Errorf("error running smoketest: %v", err)
	}
	return nil
}

// RunPreflightCheck against the nodes defined in the plan
func (ae *ansibleExecutor) RunPreFlightCheck(p *Plan) error {
	runDirectory, err := ae.createRunDirectory("preflight")
	if err != nil {
		return fmt.Errorf("error creating working directory for preflight: %v", err)
	}
	// Save the plan file that was used for this execution
	fp := FilePlanner{
		File: filepath.Join(runDirectory, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(p); err != nil {
		return fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}

	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}

	// Build inventory and save it in runs directory
	inventory := buildInventoryFromPlan(p)

	pwd, _ := os.Getwd()

	cc, err := ae.buildInstallExtraVars(p)
	if err != nil {
		return err
	}

	cc.KismaticPreflightCheckerLinux = filepath.Join("inspector", "linux", "amd64", "kismatic-inspector")
	cc.KismaticPreflightCheckerLocal = filepath.Join(pwd, "ansible", "playbooks", "inspector", runtime.GOOS, runtime.GOARCH, "kismatic-inspector")
	cc.EnablePackageInstallation = p.Cluster.AllowPackageInstallation

	// run the pre-flight playbook with pre-flight explainer
	playbook := "preflight.yaml"
	explainer := &explain.PreflightEventExplainer{
		DefaultExplainer: &explain.DefaultEventExplainer{},
	}
	if err = ae.runPlaybookWithExplainer(playbook, explainer, inventory, *cc, ansibleLogFile, runDirectory); err != nil {
		return fmt.Errorf("error running preflight: %v", err)
	}
	return nil
}

func (ae *ansibleExecutor) RunTask(taskName string, p *Plan) error {
	runDir, err := ae.createRunDirectory("step")
	if err != nil {
		return err
	}
	// Save the plan file that was used for this execution
	fp := FilePlanner{
		File: filepath.Join(runDir, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(p); err != nil {
		return fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}
	ansibleLogFilename := filepath.Join(runDir, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}
	explainer := &explain.DefaultEventExplainer{}
	inventory := buildInventoryFromPlan(p)
	ev, err := ae.buildInstallExtraVars(p)
	if err != nil {
		return err
	}
	util.PrintHeader(ae.stdout, "Running Task", '=')
	if err := ae.runPlaybookWithExplainer(taskName, explainer, inventory, *ev, ansibleLogFile, runDir); err != nil {
		return fmt.Errorf("error running task: %v", err)
	}
	return nil
}

func (ae *ansibleExecutor) AddVolume(plan *Plan, volume StorageVolume) error {
	runDirectory, err := ae.createRunDirectory("add-volume")
	if err != nil {
		return fmt.Errorf("error creating working directory for add-volume: %v", err)
	}
	fp := FilePlanner{
		File: filepath.Join(runDirectory, "kismatic-cluster.yaml"),
	}
	if err = fp.Write(plan); err != nil {
		return fmt.Errorf("error recording plan file to %s: %v", fp.File, err)
	}
	inventory := buildInventoryFromPlan(plan)
	cc, err := ae.buildInstallExtraVars(plan)
	if err != nil {
		return err
	}

	// Validate that there are enough storage nodes to satisfy the request
	nodesRequired := volume.ReplicateCount * volume.DistributionCount
	if nodesRequired > len(plan.Storage.Nodes) {
		return fmt.Errorf("the requested volume configuration requires %d storage nodes, but the cluster only has %d.", nodesRequired, len(plan.Storage.Nodes))
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

	ansibleLogFilename := filepath.Join(runDirectory, "ansible.log")
	ansibleLogFile, err := os.Create(ansibleLogFilename)
	if err != nil {
		return fmt.Errorf("error creating ansible log file %q: %v", ansibleLogFilename, err)
	}
	util.PrintHeader(ae.stdout, "Add Persistent Storage Volume", '=')
	playbook := "volume-add.yaml"
	eventExplainer := &explain.DefaultEventExplainer{}
	if err = ae.runPlaybookWithExplainer(playbook, eventExplainer, inventory, *cc, ansibleLogFile, runDirectory); err != nil {
		return err
	}
	return nil
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

func (ae *ansibleExecutor) runPlaybookWithExplainer(playbook string, eventExplainer explain.AnsibleEventExplainer, inv ansible.Inventory, cc ansible.ClusterCatalog, ansibleLog io.Writer, runDirectory string) error {
	// Setup sinks for explainer and ansible stdout
	runner, explainer, err := ae.getAnsibleRunnerAndExplainer(eventExplainer, ansibleLog, runDirectory)
	if err != nil {
		return err
	}

	// Start running ansible with the given playbook
	eventStream, err := runner.StartPlaybook(playbook, inv, cc)
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

func (ae *ansibleExecutor) getAnsibleRunnerAndExplainer(explainer explain.AnsibleEventExplainer, ansibleLog io.Writer, runDirectory string) (ansible.Runner, *explain.AnsibleEventStreamExplainer, error) {
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
