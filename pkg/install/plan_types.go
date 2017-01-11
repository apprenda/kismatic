package install

// NetworkConfig describes the cluster's networking configuration
type NetworkConfig struct {
	Type             string
	PodCIDRBlock     string `yaml:"pod_cidr_block"`
	ServiceCIDRBlock string `yaml:"service_cidr_block"`
	PolicyEnabled    bool   `yaml:"policy_enabled"`
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
	Name                     string
	AdminPassword            string `yaml:"admin_password"`
	AllowPackageInstallation bool   `yaml:"allow_package_installation"`
	Networking               NetworkConfig
	Certificates             CertsConfig
	SSH                      SSHConfig
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

// Plan is the installation plan that the user intends to execute
type Plan struct {
	Cluster        Cluster
	DockerRegistry DockerRegistry `yaml:"docker_registry"`
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
	// AllowAddresses is a list of address wildcards that have access to the volume
	AllowAddresses []string
}
