package install

import (
	"fmt"
	"net"
	"strconv"

	"github.com/apprenda/kismatic/pkg/ssh"
)

const DefaultPackageManagerProvider = "helm"

func PackageManagerProviders() []string {
	return []string{"helm", ""}
}

// NetworkConfig describes the cluster's networking configuration
type NetworkConfig struct {
	Type             string
	PodCIDRBlock     string `yaml:"pod_cidr_block"`
	ServiceCIDRBlock string `yaml:"service_cidr_block"`
	UpdateHostsFiles bool   `yaml:"update_hosts_files"`
}

// CertsConfig describes the cluster's trust and certificate configuration
type CertsConfig struct {
	Expiry string
}

// SSHConfig describes the cluster's SSH configuration for accessing nodes
type SSHConfig struct {
	User string
	Key  string `yaml:"ssh_key"`
	Port int    `yaml:"ssh_port"`
}

// Cluster describes a Kubernetes cluster
type Cluster struct {
	Name                       string
	AdminPassword              string `yaml:"admin_password"`
	DisablePackageInstallation bool   `yaml:"disable_package_installation"`
	AllowPackageInstallation   *bool  `yaml:"allow_package_installation,omitempty"`
	PackageRepoURLs            string `yaml:"package_repository_urls"`
	DisconnectedInstallation   bool   `yaml:"disconnected_installation"`
	DisableRegistrySeeding     bool   `yaml:"disable_registry_seeding"`
	Networking                 NetworkConfig
	Certificates               CertsConfig
	SSH                        SSHConfig
	APIServerOptions           APIServerOptions `yaml:"kube_apiserver"`
}

// A Node is a compute unit, virtual or physical, that is part of the cluster
type Node struct {
	Host       string
	IP         string
	InternalIP string
}

// A NodeGroup is a collection of nodes
type NodeGroup struct {
	ExpectedCount int `yaml:"expected_count"`
	Nodes         []Node
}

// An OptionalNodeGroup is a collection of nodes that can be empty
type OptionalNodeGroup NodeGroup

type NFS struct {
	Volumes []NFSVolume `yaml:"nfs_volume"`
}

type NFSVolume struct {
	Host string `yaml:"nfs_host"`
	Path string `yaml:"mount_path"`
}

// MasterNodeGroup is the collection of master nodes
type MasterNodeGroup struct {
	ExpectedCount         int    `yaml:"expected_count"`
	LoadBalancedFQDN      string `yaml:"load_balanced_fqdn"`
	LoadBalancedShortName string `yaml:"load_balanced_short_name"`
	Nodes                 []Node
}

// DockerRegistry details for docker registry, either confgiured by the cli or customer provided
type DockerRegistry struct {
	SetupInternal bool `yaml:"setup_internal"`
	Address       string
	Port          int
	CAPath        string `yaml:"CA"`
}

// Docker includes the configuration for the docker installation owned by KET.
type Docker struct {
	// Storage includes the storage-specific configuration for docker
	Storage DockerStorage
}

// DockerStorage includes the storage-specific configuration for docker.
type DockerStorage struct {
	// DirectLVM is the configuration required for setting up device mapper in direct-lvm mode
	DirectLVM DockerStorageDirectLVM `yaml:"direct_lvm"`
}

// DockerStorageDirectLVM includes the configuration required for setting up
// device mapper in direct-lvm mode
type DockerStorageDirectLVM struct {
	// Determines whether direct-lvm mode is enabled
	Enabled bool
	// BlockDevice is the path to the block device that will be used. E.g. /dev/sdb
	BlockDevice string `yaml:"block_device"`
	// EnableDeferredDeletion determines whether deferred deletion should be enabled
	EnableDeferredDeletion bool `yaml:"enable_deferred_deletion"`
}

// Plan is the installation plan that the user intends to execute
type Plan struct {
	Cluster        Cluster
	Docker         Docker
	DockerRegistry DockerRegistry `yaml:"docker_registry"`
	AddOns         AddOns         `yaml:"add_ons"`
	Features       *Features      `yaml:"features,omitempty"`
	Etcd           NodeGroup
	Master         MasterNodeGroup
	Worker         NodeGroup
	Ingress        OptionalNodeGroup
	Storage        OptionalNodeGroup
	NFS            NFS
}

// StorageVolume managed by Kismatic
type StorageVolume struct {
	// Name of the storage volume
	Name string
	// SizeGB is the size of the volume, in gigabytes
	SizeGB int
	// ReplicateCount is the number of replicas
	ReplicateCount int
	// DistributionCount is the degree to which data will be distributed across the cluster
	DistributionCount int
	// StorageClass is the annotation that will be used when creating the persistent-volume in kubernetes
	StorageClass string
	// AllowAddresses is a list of address wildcards that have access to the volume
	AllowAddresses []string
}

type SSHConnection struct {
	SSHConfig *SSHConfig
	Node      *Node
}

type AddOns struct {
	HeapsterMonitoring HeapsterMonitoring `yaml:"heapster"`
	PackageManager     PackageManager     `yaml:"package_manager"`
}

// Features is deprecated, required to support KET v1.3.3
// When writing out a new plan file, this will be nil and will not appear
type Features struct {
	PackageManager *DeprecatedPackageManager `yaml:"package_manager,omitempty"`
}

type HeapsterMonitoring struct {
	Disable bool
	Options HeapsterOptions `yaml:"options"`
}

type HeapsterOptions struct {
	HeapsterReplicas int    `yaml:"heapster_replicas"`
	InfluxDBPVCName  string `yaml:"influxdb_pvc_name"`
}

type PackageManager struct {
	Disable  bool
	Provider string
}

type DeprecatedPackageManager struct {
	Enabled bool
}

// GetUniqueNodes returns a list of the unique nodes that are listed in the plan file.
// That is, if a node has multiple roles, it will only appear once in the list.
func (p *Plan) GetUniqueNodes() []Node {
	seenNodes := map[Node]bool{}
	nodes := []Node{}
	for _, node := range p.getAllNodes() {
		if seenNodes[node] {
			continue
		}
		nodes = append(nodes, node)
		seenNodes[node] = true
	}
	return nodes
}

func (p *Plan) getAllNodes() []Node {
	nodes := []Node{}
	nodes = append(nodes, p.Etcd.Nodes...)
	nodes = append(nodes, p.Master.Nodes...)
	nodes = append(nodes, p.Worker.Nodes...)
	if p.Ingress.Nodes != nil {
		nodes = append(nodes, p.Ingress.Nodes...)
	}
	if p.Storage.Nodes != nil {
		nodes = append(nodes, p.Storage.Nodes...)
	}
	return nodes
}

func (p *Plan) getNodeWithIP(ip string) (*Node, error) {
	for _, n := range p.getAllNodes() {
		if n.IP == ip {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("Node with IP %q was not found in plan", ip)
}

// GetSSHConnection returns the SSHConnection struct containing the node and SSHConfig details
func (p *Plan) GetSSHConnection(host string) (*SSHConnection, error) {
	nodes := p.getAllNodes()

	var isIP bool
	if ip := net.ParseIP(host); ip != nil {
		isIP = true
	}

	// try to find the node with the provided hostname
	var foundNode *Node
	for _, node := range nodes {
		nodeAddress := node.Host
		if isIP {
			nodeAddress = node.IP
		}
		if nodeAddress == host {
			foundNode = &node
			break
		}
	}

	if foundNode == nil {
		switch host {
		case "master":
			foundNode = firstIfItExists(p.Master.Nodes)
		case "etcd":
			foundNode = firstIfItExists(p.Etcd.Nodes)
		case "worker":
			foundNode = firstIfItExists(p.Worker.Nodes)
		case "ingress":
			foundNode = firstIfItExists(p.Ingress.Nodes)
		case "storage":
			foundNode = firstIfItExists(p.Storage.Nodes)
		}
	}

	if foundNode == nil {
		notFoundErr := fmt.Errorf("node %q not found in the plan", host)
		if isIP {
			notFoundErr = fmt.Errorf("node with IP %q not found in the plan", host)
		}
		return nil, notFoundErr
	}

	return &SSHConnection{&p.Cluster.SSH, foundNode}, nil
}

// GetSSHClient is a convience method that calls GetSSHConnection and returns an SSH client with the result
func (p *Plan) GetSSHClient(host string) (ssh.Client, error) {
	con, err := p.GetSSHConnection(host)
	if err != nil {
		return nil, err
	}
	client, err := ssh.NewClient(con.Node.IP, con.SSHConfig.Port, con.SSHConfig.User, con.SSHConfig.Key)
	if err != nil {
		return nil, fmt.Errorf("error creating SSH client for host %s: %v", host, err)
	}

	return client, nil
}

func firstIfItExists(nodes []Node) *Node {
	if len(nodes) > 0 {
		return &nodes[0]
	}
	return nil
}

func (p *Plan) GetRolesForIP(ip string) []string {
	allRoles := []string{}

	if hasIP(&p.Master.Nodes, ip) {
		allRoles = append(allRoles, "master")
	}

	if hasIP(&p.Etcd.Nodes, ip) {
		allRoles = append(allRoles, "etcd")
	}

	if hasIP(&p.Worker.Nodes, ip) {
		allRoles = append(allRoles, "worker")
	}

	if hasIP(&p.Ingress.Nodes, ip) {
		allRoles = append(allRoles, "ingress")
	}

	if hasIP(&p.Storage.Nodes, ip) {
		allRoles = append(allRoles, "storage")
	}

	return allRoles
}

func hasIP(nodes *[]Node, ip string) bool {
	for _, node := range *nodes {
		if node.IP == ip {
			return true
		}
	}
	return false
}

// ConfigureDockerWithPrivateRegistry returns true when confgiuring an external or on cluster registry is required
func (r DockerRegistry) ConfigureDockerWithPrivateRegistry() bool {
	return r.Address != "" || r.SetupInternal
}

func (p Plan) DockerRegistryAddress() string {
	address := p.DockerRegistry.Address
	// If external is not set use master[0]
	if address == "" {
		address = p.Master.Nodes[0].IP
		// Use internal address if available
		if p.Master.Nodes[0].InternalIP != "" {
			address = p.Master.Nodes[0].InternalIP
		}
	}
	return address
}

func (p Plan) DockerRegistryPort() string {
	port := 8443
	if p.DockerRegistry.Port != 0 {
		port = p.DockerRegistry.Port
	}
	return strconv.Itoa(port)
}
