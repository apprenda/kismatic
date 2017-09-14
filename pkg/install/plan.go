package install

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/apprenda/kismatic/pkg/util"
	garbler "github.com/michaelbironneau/garbler/lib"

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
}

func setDefaults(p *Plan) {
	if p.AddOns.CNI == nil {
		p.AddOns.CNI = &CNI{}
		p.AddOns.CNI.Provider = cniProviderCalico
		p.AddOns.CNI.Options.Calico.Mode = "overlay"
		// read KET <v1.5.0 plan option
		if p.Cluster.Networking.Type != "" {
			p.AddOns.CNI.Options.Calico.Mode = p.Cluster.Networking.Type
		}
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
	//  the stack will have [cluster, networking, foo]
	s := newStack()
	scanner := bufio.NewScanner(bytes.NewReader(bytez))
	prevIndent := -1
	for scanner.Scan() {
		text := scanner.Text()
		matched := yamlKeyRE.FindStringSubmatch(text)
		if matched != nil && len(matched) > 1 {
			indent := strings.Count(matched[0], " ") / 2
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
				if _, err := f.WriteString(getCommentedLine(text, thiscomment)); err != nil {
					return err
				}
				delete(oneTimeComments, matched[1])
				continue
			}
			// Partial key match (e.g. "pod_cidr")
			if thiscomment, ok := oneTimeComments[matched[1]]; ok {
				if _, err := f.WriteString(getCommentedLine(text, thiscomment)); err != nil {
					return err
				}
				delete(oneTimeComments, matched[1])
				continue
			}
		}
		// we don't want to comment this line... just print it out
		if _, err := f.WriteString(text + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func getCommentedLine(line string, commentLines []string) string {
	var b bytes.Buffer
	// Print out a new line before each comment block
	b.WriteString("\n")
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
	if planTemplateOpts.AdminPassword == "" {
		pw, err := generateAlphaNumericPassword()
		if err != nil {
			return fmt.Errorf("error generating random password: %v", err)
		}
		planTemplateOpts.AdminPassword = pw
	}
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

	// Set DockerRegistry defaults
	p.DockerRegistry.Port = 8443

	// Add-Ons
	// CNI
	p.AddOns.CNI = &CNI{}
	p.AddOns.CNI.Provider = cniProviderCalico
	p.AddOns.CNI.Options.Calico.Mode = "overlay"
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

func generateAlphaNumericPassword() (string, error) {
	attempts := 0
	for {
		reqs := &garbler.PasswordStrengthRequirements{
			MinimumTotalLength: 16,
			Uppercase:          rand.Intn(6),
			Digits:             rand.Intn(6),
			Punctuation:        -1, // disable punctuation
		}
		pass, err := garbler.NewPassword(reqs)
		if err != nil {
			return "", err
		}
		// validate that the library actually returned an alphanumeric password
		re := regexp.MustCompile("^[a-zA-Z1-9]+$")
		if re.MatchString(pass) {
			return pass, nil
		}
		if attempts == 5 {
			return "", errors.New("failed to generate alphanumeric password")
		}
		attempts++
	}
}

// The comment map contains is keyed by the value that should be commented
// in the plan file. The value of the map contains the comment, split into
// separate lines.
var commentMap = map[string][]string{
	"cluster.admin_password":                []string{"This password is used to login to the Kubernetes Dashboard and can also be", "used for administration without a security certificate."},
	"cluster.disable_package_installation":  []string{"When true, installation will not occur if any node is missing the correct", "deb/rpm packages. When false, the installer will attempt to install missing", "packages for you."},
	"cluster.package_repository_urls":       []string{"Comma-separated list of URLs of the repositories that should be used during", "installation. These repositories must contain the kismatic packages and all", "their transitive dependencies."},
	"cluster.disconnected_installation":     []string{"Set to true if you have already installed the required packages on the nodes", "or provided a local URL in package_repository_urls containing those packages."},
	"cluster.disable_registry_seeding":      []string{"Set to true if you have seeded your registry with the required images for", "the installation."},
	"cluster.networking.pod_cidr_block":     []string{"Kubernetes will assign pods IPs in this range. Do not use a range that is", "already in use on your local network!"},
	"cluster.networking.service_cidr_block": []string{"Kubernetes will assign services IPs in this range. Do not use a range", "that is already in use by your local network or pod network!"},
	"cluster.networking.update_hosts_files": []string{"When true, the installer will add entries for all nodes to other nodes'", "hosts files. Use when you don't have access to DNS."},
	"cluster.networking.http_proxy":         []string{"Set the proxy server to use for HTTP connections."},
	"cluster.networking.https_proxy":        []string{"Set the proxy server to use for HTTPs connections."},
	"cluster.networking.no_proxy":           []string{"List of host names and/or IPs that shouldn't go through any proxy.", "If set to an asterisk '*' only, it matches all hosts.", "All nodes' 'host' and 'IPs' are always set."},
	"cluster.certificates.expiry":           []string{"Self-signed certificate expiration period in hours; default is 2 years."},
	"cluster.certificates.ca_expiry":        []string{"CA certificate expiration period in hours; default is 2 years."},
	"cluster.ssh.ssh_key":                   []string{"Absolute path to the ssh private key we should use to manage nodes."},
	"cluster.cloud_provider.provider":       []string{"Options: 'aws','azure','cloudstack','fake','gce','mesos','openstack',", "'ovirt','photon','rackspace','vsphere'.", "Leave empty for bare metal setups or other unsupported providers."},
	"cluster.cloud_provider.config":         []string{"Path to the config file, leave empty if provider does not require it."},
	"etcd":                                               []string{"Here you will identify all of the nodes that should play the etcd role", "on your cluster."},
	"master":                                             []string{"Here you will identify all of the nodes that should play the master role."},
	"worker":                                             []string{"Here you will identify all of the nodes that will be workers."},
	"host":                                               []string{"The (short) hostname of a node, e.g. etcd01."},
	"ip":                                                 []string{"The ip address the installer should use to manage this node, e.g. 8.8.8.8."},
	"internalip":                                         []string{"If the node has an IP for internal traffic, enter it here.", "Otherwise leave blank."},
	"master.load_balanced_fqdn":                          []string{"If you have set up load balancing for master nodes, enter the FQDN name here.", "Otherwise, use the IP address of a single master node."},
	"master.load_balanced_short_name":                    []string{"If you have set up load balancing for master nodes, enter the short name here.", "Otherwise, use the IP address of a single master node."},
	"docker.storage.direct_lvm":                          []string{"Configure devicemapper in direct-lvm mode (RHEL/CentOS only)."},
	"docker.storage.direct_lvm.block_device":             []string{"Path to the block device that will be used for direct-lvm mode. This", "device will be wiped and used exclusively by docker."},
	"docker.storage.direct_lvm.enable_deferred_deletion": []string{"Set to true if you want to enable deferred deletion when using", "direct-lvm mode."},
	"docker_registry":                                    []string{"Here you will provide the details of your Docker registry or setup an internal", "one to run in the cluster. This is optional and the cluster will always have", "access to the Docker Hub."},
	"docker_registry.setup_internal":                     []string{"When true, a Docker Registry will be installed on top of your cluster and", "used to host Docker images needed for its installation."},
	"docker_registry.address":                            []string{"IP or hostname for your Docker registry. An internal registry will NOT be", "setup when this field is provided. Must be accessible from all the nodes", "in the cluster."},
	"docker_registry.port":                               []string{"Port for your Docker registry."},
	"docker_registry.CA":                                 []string{"Absolute path to the CA that was used when starting your Docker registry.", "The docker daemons on all nodes in the cluster will be configured with this CA."},
	"nfs":                                                []string{"A set of NFS volumes for use by on-cluster persistent workloads"},
	"nfs.nfs_host":                                       []string{"The host name or ip address of an NFS server."},
	"nfs.mount_path":                                     []string{"The mount path of an NFS share. Must start with /"},
	"add_ons.cni.provider":                               []string{"Selecting 'custom' will result in a CNI ready cluster, however it is up to", "you to configure a plugin after the install.", "Options: 'calico','weave','contiv','custom'."},
	"add_ons.cni.options.calico.mode":                    []string{"Routed pods can be addressed from outside the Kubernetes cluster", "Overlay pods can only address each other.", "Options: 'overlay','routed'."},
	"add_ons.heapster.options.influxdb.pvc_name":         []string{"Provide the name of the persistent volume claim that you will create", "after installation. If not specified, the data will be stored in", "ephemeral storage."},
	"add_ons.heapster.options.heapster.service_type":     []string{"Specify kubernetes ServiceType; default 'ClusterIP'", "Options: 'ClusterIP','NodePort','LoadBalancer','ExternalName'."},
	"add_ons.heapster.options.heapster.sink":             []string{"Specify the sink to store heapster data. Defaults to a pod running", "on the cluster."},
	"add_ons.package_manager.provider":                   []string{"Options: 'helm'"},
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
