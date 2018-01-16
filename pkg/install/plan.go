package install

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/apprenda/kismatic/pkg/util"

	yaml "gopkg.in/yaml.v2"
)

const (
	ket133PackageManagerProvider = "helm"
	defaultCAExpiry              = "17520h"
)

// PlanTemplateOptions contains the options that are desired when generating
// a plan file template.
type PlanTemplateOptions struct {
	EtcdNodes     int
	MasterNodes   int
	WorkerNodes   int
	IngressNodes  int
	StorageNodes  int
	NFSVolumes    int
	AdminPassword string
}

// PlanReadWriter is capable of reading/writing a Plan
type PlanReadWriter interface {
	Read() (*Plan, error)
	Write(*Plan) error
}

// Planner is used to plan the installation
type Planner interface {
	PlanReadWriter
	PlanExists() bool
}

// FilePlanner is a file-based installation planner
type FilePlanner struct {
	File string
}

// Read the plan from the file system
func (fp *FilePlanner) Read() (*Plan, error) {
	d, err := ioutil.ReadFile(fp.File)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	p := &Plan{}
	if err = yaml.Unmarshal(d, p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %v", err)
	}

	// read deprecated fields and set it the new version of the cluster file
	readDeprecatedFields(p)

	// set nil values to defaults
	setDefaults(p)

	return p, nil
}

func readDeprecatedFields(p *Plan) {
	// only set if not already being set by the user
	// package_manager moved from features: to add_ons: after KET v1.3.3
	if p.Features != nil && p.Features.PackageManager != nil {
		p.AddOns.PackageManager.Disable = !p.Features.PackageManager.Enabled
		// KET v1.3.3 did not have a provider field
		p.AddOns.PackageManager.Provider = ket133PackageManagerProvider
	}
	// allow_package_installation renamed to disable_package_installation after KET v1.4.0
	if p.Cluster.AllowPackageInstallation != nil {
		p.Cluster.DisablePackageInstallation = !*p.Cluster.AllowPackageInstallation
	}

	// Only read the deprecated dashboard field if the new one is not set
	if p.AddOns.DashboardDeprecated != nil && p.AddOns.Dashboard == nil {
		p.AddOns.Dashboard = &Dashboard{
			Disable: p.AddOns.DashboardDeprecated.Disable,
		}
	}

	if p.DockerRegistry.Server == "" && p.DockerRegistry.Address != "" && p.DockerRegistry.Port != 0 {
		p.DockerRegistry.Server = fmt.Sprintf("%s:%d", p.DockerRegistry.Address, p.DockerRegistry.Port)
	}
}

func setDefaults(p *Plan) {
	if p.Docker.Logs.Driver == "" {
		p.Docker.Logs.Driver = "json-file"
		p.Docker.Logs.Opts = map[string]string{
			"max-size": "50m",
			"max-file": "1",
		}
	}

	if p.AddOns.CNI == nil {
		p.AddOns.CNI = &CNI{}
		p.AddOns.CNI.Provider = cniProviderCalico
		p.AddOns.CNI.Options.Calico.Mode = "overlay"
		p.AddOns.CNI.Options.Calico.LogLevel = "info"
		// read KET <v1.5.0 plan option
		if p.Cluster.Networking.Type != "" {
			p.AddOns.CNI.Options.Calico.Mode = p.Cluster.Networking.Type
		}
	}
	if p.AddOns.CNI.Options.Calico.LogLevel == "" {
		p.AddOns.CNI.Options.Calico.LogLevel = "info"
	}
	if p.AddOns.CNI.Options.Calico.FelixInputMTU == 0 {
		p.AddOns.CNI.Options.Calico.FelixInputMTU = 1440
	}

	if p.AddOns.CNI.Options.Calico.WorkloadMTU == 0 {
		p.AddOns.CNI.Options.Calico.WorkloadMTU = 1500
	}

	if p.AddOns.DNS.Provider == "" {
		p.AddOns.DNS.Provider = "kubedns"
	}

	if p.AddOns.HeapsterMonitoring == nil {
		p.AddOns.HeapsterMonitoring = &HeapsterMonitoring{}
	}
	if p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas == 0 {
		p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas = 2
	}
	// read field from KET < v1.5.0
	if p.AddOns.HeapsterMonitoring.Options.HeapsterReplicas != 0 {
		p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas = p.AddOns.HeapsterMonitoring.Options.HeapsterReplicas
	}
	if p.AddOns.HeapsterMonitoring.Options.Heapster.Sink == "" {
		p.AddOns.HeapsterMonitoring.Options.Heapster.Sink = "influxdb:http://heapster-influxdb.kube-system.svc:8086"
	}
	if p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType == "" {
		p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType = "ClusterIP"
	}
	if p.AddOns.HeapsterMonitoring.Options.InfluxDBPVCName != "" {
		p.AddOns.HeapsterMonitoring.Options.InfluxDB.PVCName = p.AddOns.HeapsterMonitoring.Options.InfluxDBPVCName
	}

	if p.Cluster.Certificates.CAExpiry == "" {
		p.Cluster.Certificates.CAExpiry = defaultCAExpiry
	}

	if p.AddOns.Dashboard == nil {
		p.AddOns.Dashboard = &Dashboard{}
	}
}

var yamlKeyRE = regexp.MustCompile(`[^a-zA-Z]*([a-z_\-A-Z]+)[ ]*:`)

// Write the plan to the file system
func (fp *FilePlanner) Write(p *Plan) error {
	// make a copy of the global comment map
	oneTimeComments := map[string][]string{}
	for k, v := range commentMap {
		oneTimeComments[k] = v
	}
	bytez, marshalErr := yaml.Marshal(p)
	if marshalErr != nil {
		return fmt.Errorf("error marshalling plan to yaml: %v", marshalErr)
	}

	f, err := os.Create(fp.File)
	if err != nil {
		return fmt.Errorf("error making plan file: %v", err)
	}
	defer f.Close()

	// the stack keeps track of the object we are in
	// for example, when we are inside cluster.networking, looking at the key 'foo'
	// the stack will have [cluster, networking, foo]
	s := newStack()
	scanner := bufio.NewScanner(bytes.NewReader(bytez))
	prevIndent := -1
	addNewLineBeforeComment := true
	var etcdBlock bool
	for scanner.Scan() {
		text := scanner.Text()
		matched := yamlKeyRE.FindStringSubmatch(text)
		if matched != nil && len(matched) > 1 {
			indent := strings.Count(matched[0], " ") / 2

			// Figure out if we are in the etcd block
			if indent == 0 {
				etcdBlock = (text == "etcd:")
			}
			// Don't print labels: {} for etcd group
			if etcdBlock && strings.Contains(text, "labels: {}") {
				continue
			}

			// Add a new line if we are leaving a major indentation block
			// (leaving a struct)..
			if indent < prevIndent {
				f.WriteString("\n")
				// suppress the new line that would be added if this
				// field has a comment
				addNewLineBeforeComment = false
			}
			if indent <= prevIndent {
				for i := 0; i <= (prevIndent - indent); i++ {
					// Pop from the stack when we have left an object
					// (we know because the indentation level has decreased)
					if _, err := s.Pop(); err != nil {
						return err
					}
				}
			}
			s.Push(matched[1])
			prevIndent = indent

			// Full key match (e.g. "cluster.networking.pod_cidr")
			if thiscomment, ok := oneTimeComments[strings.Join(s.s, ".")]; ok {
				if _, err := f.WriteString(getCommentedLine(text, thiscomment, addNewLineBeforeComment)); err != nil {
					return err
				}
				delete(oneTimeComments, matched[1])
				addNewLineBeforeComment = true
				continue
			}
		}
		// we don't want to comment this line... just print it out
		if _, err := f.WriteString(text + "\n"); err != nil {
			return err
		}
		addNewLineBeforeComment = true
	}

	return nil
}

func getCommentedLine(line string, commentLines []string, addNewLine bool) string {
	var b bytes.Buffer
	// Print out a new line before each comment block
	if addNewLine {
		b.WriteString("\n")
	}
	// Print out the comment lines
	for _, c := range commentLines {
		// Indent the comment to the same level as the field we are commenting
		b.WriteString(strings.Repeat(" ", countLeadingSpace(line)))
		b.WriteString(fmt.Sprintf("# %s\n", c))
	}
	// Print out the line
	b.WriteString(line + "\n")
	return b.String()
}

func countLeadingSpace(s string) int {
	var i int
	for _, r := range s {
		if r == ' ' {
			i++
			continue
		}
		break
	}
	return i
}

// PlanExists return true if the plan exists on the file system
func (fp *FilePlanner) PlanExists() bool {
	_, err := os.Stat(fp.File)
	return !os.IsNotExist(err)
}

// WritePlanTemplate writes an installation plan with pre-filled defaults.
func WritePlanTemplate(planTemplateOpts PlanTemplateOptions, w PlanReadWriter) error {
	p := buildPlanFromTemplateOptions(planTemplateOpts)
	if err := w.Write(&p); err != nil {
		return fmt.Errorf("error writing installation plan template: %v", err)
	}
	return nil
}

// fills out a plan with sensible defaults, according to the requested
// template options
func buildPlanFromTemplateOptions(templateOpts PlanTemplateOptions) Plan {
	p := Plan{}
	p.Cluster.Name = "kubernetes"
	p.Cluster.AdminPassword = templateOpts.AdminPassword
	p.Cluster.DisablePackageInstallation = false
	p.Cluster.DisconnectedInstallation = false

	// Set SSH defaults
	p.Cluster.SSH.User = "kismaticuser"
	p.Cluster.SSH.Key = "kismaticuser.key"
	p.Cluster.SSH.Port = 22

	// Set Networking defaults
	p.Cluster.Networking.PodCIDRBlock = "172.16.0.0/16"
	p.Cluster.Networking.ServiceCIDRBlock = "172.20.0.0/16"
	p.Cluster.Networking.UpdateHostsFiles = false

	// Set Certificate defaults
	p.Cluster.Certificates.Expiry = "17520h"
	p.Cluster.Certificates.CAExpiry = defaultCAExpiry

	// Docker
	p.Docker.Logs = DockerLogs{
		Driver: "json-file",
		Opts: map[string]string{
			"max-size": "50m",
			"max-file": "1",
		},
	}

	// Add-Ons
	// CNI
	p.AddOns.CNI = &CNI{}
	p.AddOns.CNI.Provider = cniProviderCalico
	p.AddOns.CNI.Options.Calico.Mode = "overlay"
	p.AddOns.CNI.Options.Calico.LogLevel = "info"
	p.AddOns.CNI.Options.Calico.WorkloadMTU = 1500
	p.AddOns.CNI.Options.Calico.FelixInputMTU = 1440
	// DNS
	p.AddOns.DNS.Provider = "kubedns"
	// Heapster
	p.AddOns.HeapsterMonitoring = &HeapsterMonitoring{}
	p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas = 2
	p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType = "ClusterIP"
	p.AddOns.HeapsterMonitoring.Options.Heapster.Sink = "influxdb:http://heapster-influxdb.kube-system.svc:8086"

	// Package Manager
	p.AddOns.PackageManager.Provider = "helm"

	p.AddOns.Dashboard = &Dashboard{}
	p.AddOns.Dashboard.Disable = false

	// Generate entries for all node types
	p.Etcd.ExpectedCount = templateOpts.EtcdNodes
	p.Master.ExpectedCount = templateOpts.MasterNodes
	p.Worker.ExpectedCount = templateOpts.WorkerNodes
	p.Ingress.ExpectedCount = templateOpts.IngressNodes
	p.Storage.ExpectedCount = templateOpts.StorageNodes

	for i := 0; i < templateOpts.NFSVolumes; i++ {
		v := NFSVolume{Host: "", Path: "/"}
		p.NFS.Volumes = append(p.NFS.Volumes, v)
	}

	n := Node{}
	for i := 0; i < p.Etcd.ExpectedCount; i++ {
		p.Etcd.Nodes = append(p.Etcd.Nodes, n)
	}

	for i := 0; i < p.Master.ExpectedCount; i++ {
		p.Master.Nodes = append(p.Master.Nodes, n)
	}

	for i := 0; i < p.Worker.ExpectedCount; i++ {
		p.Worker.Nodes = append(p.Worker.Nodes, n)
	}

	if p.Ingress.ExpectedCount > 0 {
		for i := 0; i < p.Ingress.ExpectedCount; i++ {
			p.Ingress.Nodes = append(p.Ingress.Nodes, n)
		}
	}

	if p.Storage.ExpectedCount > 0 {
		for i := 0; i < p.Storage.ExpectedCount; i++ {
			p.Storage.Nodes = append(p.Storage.Nodes, n)
		}
	}

	return p
}

func getKubernetesServiceIP(p *Plan) (string, error) {
	ip, err := util.GetIPFromCIDR(p.Cluster.Networking.ServiceCIDRBlock, 1)
	if err != nil {
		return "", fmt.Errorf("error getting kubernetes service IP: %v", err)
	}
	return ip.To4().String(), nil
}

func getDNSServiceIP(p *Plan) (string, error) {
	ip, err := util.GetIPFromCIDR(p.Cluster.Networking.ServiceCIDRBlock, 2)
	if err != nil {
		return "", fmt.Errorf("error getting DNS service IP: %v", err)
	}
	return ip.To4().String(), nil
}

// The comment map contains is keyed by the value that should be commented
// in the plan file. The value of the map contains the comment, split into
// separate lines.
var commentMap = map[string][]string{
	"cluster.admin_password":                             []string{"This password is used to login to the Kubernetes Dashboard and can also be", "used for administration without a security certificate."},
	"cluster.disable_package_installation":               []string{"Set to true if the nodes have the required packages installed."},
	"cluster.disconnected_installation":                  []string{"Set to true if you are performing a disconnected installation."},
	"cluster.networking":                                 []string{"Networking configuration of your cluster."},
	"cluster.networking.pod_cidr_block":                  []string{"Kubernetes will assign pods IPs in this range. Do not use a range that is", "already in use on your local network!"},
	"cluster.networking.service_cidr_block":              []string{"Kubernetes will assign services IPs in this range. Do not use a range", "that is already in use by your local network or pod network!"},
	"cluster.networking.update_hosts_files":              []string{"Set to true if your nodes cannot resolve each others' names using DNS."},
	"cluster.networking.http_proxy":                      []string{"Set the proxy server to use for HTTP connections."},
	"cluster.networking.https_proxy":                     []string{"Set the proxy server to use for HTTPs connections."},
	"cluster.networking.no_proxy":                        []string{"List of host names and/or IPs that shouldn't go through any proxy.", "All nodes' 'host' and 'IPs' are always set."},
	"cluster.certificates":                               []string{"Generated certs configuration."},
	"cluster.certificates.expiry":                        []string{"Self-signed certificate expiration period in hours; default is 2 years."},
	"cluster.certificates.ca_expiry":                     []string{"CA certificate expiration period in hours; default is 2 years."},
	"cluster.ssh":                                        []string{"SSH configuration for cluster nodes."},
	"cluster.ssh.user":                                   []string{"This user must be able to sudo without password."},
	"cluster.ssh.ssh_key":                                []string{"Absolute path to the ssh private key we should use to manage nodes."},
	"cluster.kube_apiserver":                             []string{"Override configuration of Kubernetes components."},
	"cluster.cloud_provider":                             []string{"Kubernetes cloud provider integration."},
	"cluster.cloud_provider.provider":                    []string{"Options: 'aws','azure','cloudstack','fake','gce','mesos','openstack',", "'ovirt','photon','rackspace','vsphere'.", "Leave empty for bare metal setups or other unsupported providers."},
	"cluster.cloud_provider.config":                      []string{"Path to the config file, leave empty if provider does not require it."},
	"docker":                                             []string{"Docker daemon configuration of all cluster nodes."},
	"docker.disable":                                     []string{"Set to true if docker is already installed and configured."},
	"etcd":                                               []string{"Etcd nodes are the ones that run the etcd distributed key-value database."},
	"etcd.nodes":                                         []string{"Provide the hostname and IP of each node. If the node has an IP for internal", "traffic, provide it in the internalip field. Otherwise, that field can be", "left blank."},
	"master":                                             []string{"Master nodes are the ones that run the Kubernetes control plane components."},
	"worker":                                             []string{"Worker nodes are the ones that will run your workloads on the cluster."},
	"ingress":                                            []string{"Ingress nodes will run the ingress controllers."},
	"storage":                                            []string{"Storage nodes will be used to create a distributed storage cluster that can", "be consumed by your workloads."},
	"master.load_balanced_fqdn":                          []string{"If you have set up load balancing for master nodes, enter the FQDN name here.", "Otherwise, use the IP address of a single master node."},
	"master.load_balanced_short_name":                    []string{"If you have set up load balancing for master nodes, enter the short name here.", "Otherwise, use the IP address of a single master node."},
	"docker.storage.direct_lvm":                          []string{"Configure devicemapper in direct-lvm mode (RHEL/CentOS only)."},
	"docker.storage.direct_lvm.block_device":             []string{"Path to the block device that will be used for direct-lvm mode. This", "device will be wiped and used exclusively by docker."},
	"docker.storage.direct_lvm.enable_deferred_deletion": []string{"Set to true if you want to enable deferred deletion when using", "direct-lvm mode."},
	"docker_registry":                                    []string{"If you want to use an internal registry for the installation or upgrade, you", "must provide its information here. You must seed this registry before the", "installation or upgrade of your cluster. This registry must be accessible from", "all nodes on the cluster."},
	"docker_registry.server":                             []string{"IP or hostname and port for your registry."},
	"docker_registry.CA":                                 []string{"Absolute path to the certificate authority that should be trusted when", "connecting to your registry."},
	"docker_registry.username":                           []string{"Leave blank for unauthenticated access."},
	"docker_registry.password":                           []string{"Leave blank for unauthenticated access."},
	"add_ons":                                            []string{"Add-ons are additional components that KET installs on the cluster."},
	"nfs":                                                []string{"A set of NFS volumes for use by on-cluster persistent workloads."},
	"nfs.nfs_host":                                       []string{"The host name or ip address of an NFS server."},
	"nfs.mount_path":                                     []string{"The mount path of an NFS share. Must start with '/'."},
	"add_ons.cni.provider":                               []string{"Selecting 'custom' will result in a CNI ready cluster, however it is up to", "you to configure a plugin after the install.", "Options: 'calico','weave','contiv','custom'."},
	"add_ons.cni.options.calico.mode":                    []string{"Options: 'overlay','routed'."},
	"add_ons.cni.options.calico.log_level":               []string{"Options: 'warning','info','debug'."},
	"add_ons.cni.options.calico.workload_mtu":            []string{"MTU for the workload interface, configures the CNI config."},
	"add_ons.cni.options.calico.felix_input_mtu":         []string{"MTU for the tunnel device used if IPIP is enabled."},
	"add_ons.dns.provider":                               []string{"Options: 'kubedns','coredns'."},
	"add_ons.heapster.options.influxdb.pvc_name":         []string{"Provide the name of the persistent volume claim that you will create", "after installation. If not specified, the data will be stored in", "ephemeral storage."},
	"add_ons.heapster.options.heapster.service_type":     []string{"Specify kubernetes ServiceType. Defaults to 'ClusterIP'.", "Options: 'ClusterIP','NodePort','LoadBalancer','ExternalName'."},
	"add_ons.heapster.options.heapster.sink":             []string{"Specify the sink to store heapster data. Defaults to an influxdb pod", "running on the cluster."},
	"add_ons.package_manager.provider":                   []string{"Options: 'helm'."},
	"add_ons.rescheduler":                                []string{"The rescheduler ensures that critical add-ons remain running on the cluster."},
}

type stack struct {
	lock sync.Mutex
	s    []string
}

func newStack() *stack {
	return &stack{sync.Mutex{}, make([]string, 0)}
}

func (s *stack) Push(v string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.s = append(s.s, v)
}

func (s *stack) Pop() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.s)
	if l == 0 {
		return "", errors.New("Empty Stack")
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res, nil
}
