package install

import (
	"fmt"
	"net"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
	"github.com/apprenda/kismatic/pkg/tls"
)

const (
	cniProviderContiv = "contiv"
	cniProviderCalico = "calico"
	cniProviderWeave  = "weave"
	cniProviderCustom = "custom"
)

const (
	dnsProviderKubedns = "kubedns"
	dnsProviderCoredns = "coredns"
)

func packageManagerProviders() []string {
	return []string{"helm", ""}
}

func cniProviders() []string {
	return []string{cniProviderCalico, cniProviderContiv, cniProviderWeave, cniProviderCustom}
}

func dnsProviders() []string {
	return []string{dnsProviderKubedns, dnsProviderCoredns}
}

func calicoMode() []string {
	return []string{"overlay", "routed"}
}

func calicoLogLevel() []string {
	return []string{"warning", "info", "debug", ""}
}

func serviceTypes() []string {
	return []string{"ClusterIP", "NodePort", "LoadBalancer", "ExternalName"}
}

func cloudProviders() []string {
	return []string{"aws", "azure", "cloudstack", "fake", "gce", "mesos", "openstack", "ovirt", "photon", "rackspace", "vsphere"}
}

// Plan is the installation plan that the user intends to execute
type Plan struct {
	// Kubernetes cluster configuration
	// +required
	Cluster Cluster
	// Configuration for the docker engine installed by KET
	Docker Docker
	// Docker registry configuration
	DockerRegistry DockerRegistry `yaml:"docker_registry"`
	// Add on configuration
	AddOns AddOns `yaml:"add_ons"`
	// Feature configuration
	// +deprecated
	Features *Features `yaml:"features,omitempty"`
	// Etcd nodes of the cluster
	// +required
	Etcd NodeGroup
	// Master nodes of the cluster
	// +required
	Master MasterNodeGroup
	// Worker nodes of the cluster
	// +required
	Worker NodeGroup
	// Ingress nodes of the cluster
	Ingress OptionalNodeGroup
	// Storage nodes of the cluster.
	Storage OptionalNodeGroup
	// NFS volumes of the cluster.
	NFS NFS
}

// Cluster describes a Kubernetes cluster
type Cluster struct {
	// Name of the cluster to be used when generating assets that require a
	// cluster name, such as kubeconfig files and certificates.
	// +required
	Name string
	// The Kubernetes version to install.
	// If left blank will be set to the latest tested version.
	// Only a single Minor version is supported with.
	// +default=v1.9.3
	Version string
	// The password for the admin user.
	// If provided, ABAC will be enabled in the cluster.
	// This field will be removed completely in a future release.
	// +deprecated
	AdminPassword string `yaml:"admin_password,omitempty"`
	// Whether KET should install the packages on the cluster nodes.
	// When true, KET will not install the required packages.
	// Instead, it will verify that the packages have been installed by the operator.
	DisablePackageInstallation bool `yaml:"disable_package_installation"`
	// Whether KET should install the packages on the cluster nodes.
	// Use DisablePackageInstallation instead.
	// +deprecated
	AllowPackageInstallation *bool `yaml:"allow_package_installation,omitempty"`
	// Whether the cluster nodes are disconnected from the internet.
	// When set to `true`, internal package repositories and a container image
	// registry are required for installation.
	// +default=false
	DisconnectedInstallation bool `yaml:"disconnected_installation"`
	// The Networking configuration for the cluster.
	Networking NetworkConfig
	// The Certificates configuration for the cluster.
	Certificates CertsConfig
	// The SSH configuration for the cluster nodes.
	SSH SSHConfig
	// Kubernetes API Server configuration.
	APIServerOptions APIServerOptions `yaml:"kube_apiserver"`
	// Kubernetes Controller Manager configuration.
	KubeControllerManagerOptions KubeControllerManagerOptions `yaml:"kube_controller_manager"`
	// Kubernetes Scheduler configuration.
	KubeSchedulerOptions KubeSchedulerOptions `yaml:"kube_scheduler"`
	// Kubernetes Proxy configuration.
	KubeProxyOptions KubeProxyOptions `yaml:"kube_proxy"`
	// Kubelet configuration applied to all nodes.
	KubeletOptions KubeletOptions `yaml:"kubelet"`
	// The CloudProvider configuration for the cluster.
	CloudProvider CloudProvider `yaml:"cloud_provider"`
}

type APIServerOptions struct {
	// Listing of option overrides that are to be applied to the Kubernetes
	// API server configuration. This is an advanced feature that can prevent
	// the API server from starting up if invalid configuration is provided.
	Overrides map[string]string `yaml:"option_overrides"`
}

type KubeControllerManagerOptions struct {
	// Listing of option overrides that are to be applied to the Kubernetes
	// Controller Manager configuration. This is an advanced feature that can prevent
	// the Controller Manager from starting up if invalid configuration is provided.
	Overrides map[string]string `yaml:"option_overrides"`
}

type KubeProxyOptions struct {
	// Listing of option overrides that are to be applied to the Kubernetes
	// Proxy configuration. This is an advanced feature that can prevent
	// the Proxy from starting up if invalid configuration is provided.
	Overrides map[string]string `yaml:"option_overrides"`
}

type KubeSchedulerOptions struct {
	// Listing of option overrides that are to be applied to the Kubernetes
	// Scheduler configuration. This is an advanced feature that can prevent
	// the Scheduler from starting up if invalid configuration is provided.
	Overrides map[string]string `yaml:"option_overrides"`
}

type KubeletOptions struct {
	// Listing of option overrides that are to be applied to the Kubelet configurations.
	// This is an advanced feature that can prevent the Kubelet from starting up if invalid configuration is provided.
	Overrides map[string]string `yaml:"option_overrides"`
}

// NetworkConfig describes the cluster's networking configuration
type NetworkConfig struct {
	// The datapath technique that should be configured in Calico.
	// +default=overlay
	// +options=overlay,routed
	// +deprecated
	Type string `yaml:"type,omitempty"`
	// The pod network's CIDR block. For example: `172.16.0.0/16`
	// +required
	PodCIDRBlock string `yaml:"pod_cidr_block"`
	// The Kubernetes service network's CIDR block. For example: `172.20.0.0/16`
	// +required
	ServiceCIDRBlock string `yaml:"service_cidr_block"`
	// Whether the /etc/hosts file should be updated on the cluster nodes.
	// When set to true, KET will update the hosts file on all nodes to include
	// entries for all other nodes in the cluster.
	// +default=false
	UpdateHostsFiles bool `yaml:"update_hosts_files"`
	// The URL of the proxy that should be used for HTTP connections.
	HTTPProxy string `yaml:"http_proxy"`
	// The URL of the proxy that should be used for HTTPS connections.
	HTTPSProxy string `yaml:"https_proxy"`
	// Comma-separated list of host names and/or IPs for which connections
	// should not go through a proxy.
	// All nodes' 'host' and 'IPs' are always set.
	NoProxy string `yaml:"no_proxy"`
}

// CertsConfig describes the cluster's trust and certificate configuration
type CertsConfig struct {
	// The length of time that the generated certificates should be valid for.
	// For example: "17520h" for 2 years.
	// +required
	Expiry string
	// The length of time that the generated Certificate Authority should be valid for.
	// For example: "17520h" for 2 years.
	// +required.
	CAExpiry string `yaml:"ca_expiry"`
}

// SSHConfig describes the cluster's SSH configuration for accessing nodes
type SSHConfig struct {
	// The user for accessing the cluster nodes via SSH.
	// This user requires sudo elevation privileges on the cluster nodes.
	// +required
	User string
	// The absolute path of the SSH key that should be used for accessing the
	// cluster nodes via SSH.
	// +required
	Key string `yaml:"ssh_key"`
	// The port number on which cluster nodes are listening for SSH connections.
	// +required
	Port int `yaml:"ssh_port"`
}

// CloudProvider controls the Kubernetes cloud providers feature
type CloudProvider struct {
	// The cloud provider that should be set in the Kubernetes components
	// +options=aws,azure,cloudstack,fake,gce,mesos,openstack,ovirt,photon,rackspace,vsphere
	Provider string
	// Path to the cloud provider config file. This will be copied to all the machines in the cluster
	Config string
}

// Docker includes the configuration for the docker installation owned by KET.
type Docker struct {
	// Set to true to disable the installation of docker container runtime on the nodes.
	// The installer will validate that docker is installed and running prior to proceeding.
	// Use this option if a different version of docker from the included one is required.
	Disable bool
	// Log configuration for the docker engine.
	Logs DockerLogs
	// Storage configuration for the docker engine.
	Storage DockerStorage
}

// DockerLogs includes the log-specific configuration for docker.
type DockerLogs struct {
	// Docker logging driver, more details https://docs.docker.com/engine/admin/logging/overview/.
	// +default=json-file
	Driver string
	// Driver specific options.
	Opts map[string]string
}

// DockerStorage includes the storage-specific configuration for docker.
type DockerStorage struct {
	// Docker storage driver, more details https://docs.docker.com/engine/userguide/storagedriver/.
	// Leave empty to have docker automatically select the driver.
	// +default='empty'
	Driver string
	// Driver specific options
	Opts map[string]string
	// DirectLVMBlockDevice is the configuration required for setting up Device Mapper storage driver in direct-lvm mode.
	// Refer to https://docs.docker.com/v17.03/engine/userguide/storagedriver/device-mapper-driver/#manage-devicemapper docs.
	DirectLVMBlockDevice DirectLVMBlockDevice `yaml:"direct_lvm_block_device"`
	// DirectLVM is the configuration required for setting up device mapper in direct-lvm mode.
	// +deprecated
	DirectLVM *DockerStorageDirectLVMDeprecated `yaml:"direct_lvm,omitempty"`
}

type DirectLVMBlockDevice struct {
	// The path to the block device.
	Path string
	// The percentage of space to use for storage from the passed in block device.
	// +default=95
	ThinpoolPercent string `yaml:"thinpool_percent"`
	// The percentage of space to for metadata storage from the passed in block device.
	// +default=1
	ThinpoolMetaPercent string `yaml:"thinpool_metapercent"`
	// The threshold for when lvm should automatically extend the thin pool as a percentage of the total storage space.
	// +default=80
	ThinpoolAutoextendThreshold string `yaml:"thinpool_autoextend_threshold"`
	// The percentage to increase the thin pool by when an autoextend is triggered.
	// +default=20
	ThinpoolAutoextendPercent string `yaml:"thinpool_autoextend_percent"`
}

// DockerStorageDirectLVMDeprecated includes the configuration required for setting up
// device mapper in direct-lvm mode.
type DockerStorageDirectLVMDeprecated struct {
	// Whether the direct_lvm mode of the devicemapper storage driver should be enabled.
	// When set to true, a dedicated block storage device must be available on each cluster node.
	// +default=false
	Enabled bool
	// The path to the block storage device that will be used by the devicemapper storage driver.
	BlockDevice string `yaml:"block_device"`
	// Whether deferred deletion should be enabled when using devicemapper in direct_lvm mode.
	// +default=false
	EnableDeferredDeletion bool `yaml:"enable_deferred_deletion"`
}

// DockerRegistry details for docker registry, either confgiured by the cli or customer provided
type DockerRegistry struct {
	// The hostname or IP address and port of a private container image registry.
	// Do not include http or https.
	// When performing a disconnected installation, this registry will be used
	// to fetch all the required container images.
	Server string
	// The hostname or IP address of a private container image registry.
	// When performing a disconnected installation, this registry will be used
	// to fetch all the required container images.
	// +deprecated
	Address string `yaml:"address,omitempty"`
	// The port on which the private container image registry is listening on.
	// +deprecated
	Port int `yaml:"port,omitempty"`
	// The absolute path of the Certificate Authority that should be installed on
	// all cluster nodes that have a docker daemon.
	// This is required to establish trust between the daemons and the private
	// registry when the registry is using a self-signed certificate.
	CAPath string `yaml:"CA"`
	// The username that should be used when connecting to a registry that has authentication enabled.
	// Otherwise leave blank for unauthenticated access.
	Username string
	// The password that should be used when connecting to a registry that has authentication enabled.
	// Otherwise leave blank for unauthenticated access.
	Password string
}

// AddOns are components that are deployed on the cluster that KET considers
// necessary for producing a production cluster.
type AddOns struct {
	// The Container Networking Interface (CNI) add-on configuration.
	CNI *CNI `yaml:"cni"`
	// The DNS add-on configuration.
	DNS DNS `yaml:"dns"`
	// The Heapster Monitoring add-on configuration.
	HeapsterMonitoring *HeapsterMonitoring `yaml:"heapster"`
	// Metrics Server add-on configuration.
	// A cluster-wide aggregator of resource usage data.
	// Required for Horizontal Pod Autoscaler to function properly.
	MetricsServer MetricsServer `yaml:"metrics_server"`
	// The Dashboard add-on configuration.
	Dashboard *Dashboard `yaml:"dashboard"`
	// The Dashboard add-on configuration.
	// +deprecated
	DashboardDeprecated *Dashboard `yaml:"dashbard,omitempty"`
	// The PackageManager add-on configuration.
	PackageManager PackageManager `yaml:"package_manager"`
	// The Rescheduler add-on configuration.
	// Because the Rescheduler does not have leader election and therefore can only run as a single instance in a cluster, it will be deployed as a static pod on the first master.
	// More information about the Rescheduler can be found here: https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
	Rescheduler Rescheduler `yaml:"rescheduler"`
}

// Features configuration
// +deprecated
type Features struct {
	// The PackageManager feature configuration.
	// +deprecated
	PackageManager *DeprecatedPackageManager `yaml:"package_manager,omitempty"`
}

// CNI add-on configuration
type CNI struct {
	// Whether the CNI add-on is disabled. When set to true,
	// CNI will not be installed on the cluster. Furthermore, the smoke test and
	// any validation that depends on a functional pod network will be skipped.
	// +default=false
	Disable bool
	// The CNI provider that should be installed on the cluster.
	// +default=calico
	// +options=calico,weave,contiv,custom
	Provider string
	// The CNI options that can be configured for each CNI provider.
	Options CNIOptions `yaml:"options"`
}

// CNIOptions that can be configured for each CNI provider.
type CNIOptions struct {
	// The options that can be configured for the Calico CNI provider.
	Calico CalicoOptions
}

// The CalicoOptions that can be configured for the Calico CNI provider.
type CalicoOptions struct {
	// The datapath technique that should be configured in Calico.
	// +default=overlay
	// +options=overlay,routed
	Mode string
	// The logging level for the CNI plugin
	// +default=info
	// +options=warning,info,debug
	LogLevel string `yaml:"log_level"`
	// MTU for the workload interface, configures the CNI config.
	// +default=1500
	WorkloadMTU int `yaml:"workload_mtu"`
	// MTU for the tunnel device used if IPIP is enabled.
	// +default=1440
	FelixInputMTU int `yaml:"felix_input_mtu"`
	// IPAutodetectionMethod is used to detect the IPv4 address of the host.
	// The value gets set in IP_AUTODETECTION_METHOD variable in the pod.
	// +default=first-found
	IPAutodetectionMethod string `yaml:"ip_autodetection_method"`
}

// The DNS add-on configuration
type DNS struct {
	// Whether the DNS add-on should be disabled.
	// When set to true, no DNS solution will be deployed on the cluster.
	Disable bool
	// This property indicates the in-cluster DNS provider.
	// +required
	// +options=kubedns,coredns
	// +default=kubedns
	Provider string
}

// The HeapsterMonitoring add-on configuration
type HeapsterMonitoring struct {
	// Whether the Heapster add-on should be disabled.
	// When set to true, Heapster and InfluxDB will not be deployed on the cluster.
	// +default=false
	Disable bool
	// The options that can be configured for the Heapster add-on
	Options HeapsterOptions `yaml:"options"`
}

// The HeapsterOptions for the HeapsterMonitoring add-on
type HeapsterOptions struct {
	// The Heapster configuration options.
	Heapster Heapster `yaml:"heapster"`
	// The InfluxDB configuration options.
	InfluxDB InfluxDB `yaml:"influxdb"`
	// Number of Heapster replicas that should be scheduled on the cluster.
	// +deprecated
	HeapsterReplicas int `yaml:"heapster_replicas,omitempty"`
	// Name of the Persistent Volume Claim that will be used by InfluxDB.
	// When set, this PVC must be created after the installation.
	// If not set, InfluxDB will be configured with ephemeral storage.
	// +deprecated
	InfluxDBPVCName string `yaml:"influxdb_pvc_name,omitempty"`
}

// Heapster configuration options for the Heapster add-on
type Heapster struct {
	// Number of Heapster replicas that should be scheduled on the cluster.
	// +default=2
	Replicas int `yaml:"replicas"`
	// Kubernetes service type of the Heapster service.
	// +default=ClusterIP
	// +options=ClusterIP,NodePort,LoadBalancer,ExternalName
	ServiceType string `yaml:"service_type"`
	// URL of the backend store that will be used as the Heapster sink.
	// +default=influxdb:http://heapster-influxdb.kube-system.svc:8086
	Sink string `yaml:"sink"`
}

// The MetricsServer add-on configuration.
type MetricsServer struct {
	// Whether the metrics-server add-on should be disabled.
	// When set to true, metrics-server will not be deployed on the cluster.
	// +default=false
	Disable bool
}

// InfluxDB configuration options for the Heapster add-on
type InfluxDB struct {
	// Name of the Persistent Volume Claim that will be used by InfluxDB.
	// This PVC must be created after the installation.
	// If not set, InfluxDB will be configured with ephemeral storage.
	PVCName string `yaml:"pvc_name"`
}

// Dashboard add-on configuration
type Dashboard struct {
	// Whether the dashboard add-on should be disabled.
	// When set to true, the Kubernetes Dashboard will not be installed on the cluster.
	// +default=false
	Disable bool
}

// PackageManager add-on configuration
type PackageManager struct {
	// Whether the package manager add-on should be disabled.
	// When set to true, the package manager will not be installed on the cluster.
	// +default=false
	Disable bool
	// This property indicates the package manager provider.
	// +required
	// +options=helm
	Provider string
	// The PackageManager options.
	Options PackageManagerOptions `yaml:"options"`
}

// The PackageManagerOptions for the PackageManager add-on
type PackageManagerOptions struct {
	// Helm PackageManager options
	Helm HelmOptions
}

// HelmOptions for the helm PackageManager add-on
type HelmOptions struct {
	// Namespace to deploy tiller
	// +default=kube-system
	Namespace string
}

// Rescheduler add-on configuration
type Rescheduler struct {
	// Whether the pod rescheduler add-on should be disabled.
	// When set to true, the rescheduler will not be installed on the cluster.
	// +default=false
	Disable bool
}

type DeprecatedPackageManager struct {
	// Whether the package manager add-on should be enabled.
	// +deprecated
	Enabled bool
}

// MasterNodeGroup is the collection of master nodes
type MasterNodeGroup struct {
	// Number of master nodes that are part of the cluster.
	// +required
	ExpectedCount int `yaml:"expected_count"`
	// The FQDN of the load balancer that is fronting multiple master nodes.
	// In the case where there is only one master node, this can be set to the IP address of the master node.
	// +required
	LoadBalancedFQDN string `yaml:"load_balanced_fqdn"`
	// The short name of the load balancer that is fronting multiple master nodes.
	// In the case where there is only one master node, this can be set to the IP address of the master nodes.
	// +required
	LoadBalancedShortName string `yaml:"load_balanced_short_name"`
	// List of master nodes that are part of the cluster.
	// +required
	Nodes []Node
}

// A NodeGroup is a collection of nodes
type NodeGroup struct {
	// Number of nodes.
	// +required
	ExpectedCount int `yaml:"expected_count"`
	// List of nodes.
	// +required
	Nodes []Node
}

// An OptionalNodeGroup is a collection of nodes that can be empty
type OptionalNodeGroup NodeGroup

// A Node is a compute unit, virtual or physical, that is part of the cluster
type Node struct {
	// The hostname of the node. The hostname is verified
	// in the validation phase of the installation.
	// +required
	Host string
	// The IP address of the node. This is the IP address that will be used to
	// connect to the node over SSH.
	// +required
	IP string
	// The internal (or private) IP address of the node.
	// If set, this IP will be used when configuring cluster components.
	InternalIP string
	// Labels to add when installing the node in the cluster.
	// If a node is defined under multiple roles, the labels for that node will be merged.
	// If a label is repeated for the same node,
	// only one will be used in this order: etcd,master,worker,ingress,storage roles where 'storage' has the highest precedence.
	// It is recommended to use reverse-DNS notation to avoid collision with other labels.
	Labels map[string]string
	// Kubelet configuration applied to this node.
	// If a node is repeated for multiple roles, the overrides cannot be different.
	KubeletOptions KubeletOptions `yaml:"kubelet,omitempty"`
}

// Equal returns true of 2 nodes have the same host, IP and InternalIP
func (node Node) Equal(other Node) bool {
	return node.Host == other.Host && node.IP == other.IP && node.InternalIP == other.InternalIP
}

// HashCode is crude implementation for the Node struct
func (node Node) HashCode() string {
	return fmt.Sprint(node.Host, node.IP, node.InternalIP)
}

type NFS struct {
	// List of NFS volumes that should be attached to the cluster during
	// the installation.
	Volumes []NFSVolume `yaml:"nfs_volume"`
}

type NFSVolume struct {
	// The hostname or IP of the NFS volume.
	// +required
	Host string `yaml:"nfs_host"`
	// The path where the NFS volume should be mounted.
	// +required
	Path string `yaml:"mount_path"`
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
	// ReclaimPolicy is the persistent volume's reclaim policy
	// ref: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaim-policy
	ReclaimPolicy string
	// AccessModes supported by the persistent volume
	// ref: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes
	AccessModes []string
}

type SSHConnection struct {
	SSHConfig *SSHConfig
	Node      *Node
}

// GetUniqueNodes returns a list of the unique nodes that are listed in the plan file.
// That is, if a node has multiple roles, it will only appear once in the list.
// Nodes are considered unique if the combination of 'host', 'IP' or 'internalIP' is unique to all other nodes.
func (p *Plan) GetUniqueNodes() []Node {
	seenNodes := map[string]bool{}
	nodes := []Node{}
	for _, node := range p.getAllNodes() {
		// Cannot use the Node struct directly as it contains a map
		key := node.HashCode()
		if seenNodes[key] {
			continue
		}
		nodes = append(nodes, node)
		seenNodes[key] = true
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

// AllAddresses will return the hostnames, IPs and internal IPs for all nodes
func (p *Plan) AllAddresses() string {
	nodes := p.GetUniqueNodes()
	var addr []string
	for _, n := range nodes {
		addr = append(addr, n.Host)
		addr = append(addr, n.IP)
		if n.InternalIP != "" {
			addr = append(addr, n.InternalIP)
		}
	}
	return strings.Join(addr, ",")
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

// PrivateRegistryProvided returns true when the details about a private
// registry have been provided
func (p Plan) PrivateRegistryProvided() bool {
	return p.DockerRegistry.Server != ""
}

// NetworkConfigured returns true if pod validation/smoketest should run
func (p Plan) NetworkConfigured() bool {
	// CNI disabled or "custom" return false
	return p.AddOns.CNI == nil || (!p.AddOns.CNI.Disable && p.AddOns.CNI.Provider != "custom")
}

func (p Plan) Versions() map[string]string {
	kubernetesVersion := kubernetesVersionString
	if p.Cluster.Version != "" {
		kubernetesVersion = p.Cluster.Version
	}
	versions := make(map[string]string)
	versions["kube_proxy"] = kubernetesVersion
	versions["kube_controller_manager"] = kubernetesVersion
	versions["kube_scheduler"] = kubernetesVersion
	versions["kube_apiserver"] = kubernetesVersion

	return versions
}

// returns a list of specs for all the certs that are required for the node
func (node Node) certSpecs(plan Plan, ca *tls.CA) ([]certificateSpec, error) {
	m := []certificateSpec{}
	roles := plan.GetRolesForIP(node.IP)

	// Certificates for etcd
	if contains("etcd", roles) {
		san := []string{node.Host, node.IP, "127.0.0.1"}
		if node.InternalIP != "" {
			san = append(san, node.InternalIP)
		}
		m = append(m, certificateSpec{
			description:           fmt.Sprintf("%s etcd server", node.Host),
			filename:              fmt.Sprintf("%s-etcd", node.Host),
			commonName:            node.Host,
			subjectAlternateNames: san,
			ca: ca,
		})
	}

	// Certificates for master
	if contains("master", roles) {
		// API Server certificate
		san, err := clusterCertsSubjectAlternateNames(plan)
		if err != nil {
			return nil, err
		}
		san = append(san, node.Host, node.IP, "127.0.0.1")
		if node.InternalIP != "" {
			san = append(san, node.InternalIP)
		}
		if !contains(plan.Master.LoadBalancedFQDN, san) {
			san = append(san, plan.Master.LoadBalancedFQDN)
		}
		if !contains(plan.Master.LoadBalancedShortName, san) {
			san = append(san, plan.Master.LoadBalancedShortName)
		}
		m = append(m, certificateSpec{
			description:           fmt.Sprintf("%s API server", node.Host),
			filename:              fmt.Sprintf("%s-apiserver", node.Host),
			commonName:            node.Host,
			subjectAlternateNames: san,
			ca: ca,
		})
		// Controller manager certificate
		m = append(m, certificateSpec{
			description: "kubernetes controller manager",
			filename:    controllerManagerCertFilenamePrefix,
			commonName:  controllerManagerUser,
			ca:          ca,
		})
		// Scheduler client certificate
		m = append(m, certificateSpec{
			description: "kubernetes scheduler",
			filename:    schedulerCertFilenamePrefix,
			commonName:  schedulerUser,
			ca:          ca,
		})
		// Certificate for signing service account tokens
		m = append(m, certificateSpec{
			description: "service account signing",
			filename:    serviceAccountCertFilename,
			commonName:  serviceAccountCertCommonName,
			ca:          ca,
		})
	}

	// Kubelet and etcd client certificate
	if containsAny([]string{"master", "worker", "ingress", "storage"}, roles) {
		m = append(m, certificateSpec{
			description:   fmt.Sprintf("%s kubelet", node.Host),
			filename:      fmt.Sprintf("%s-kubelet", node.Host),
			commonName:    fmt.Sprintf("%s:%s", kubeletUserPrefix, strings.ToLower(node.Host)),
			organizations: []string{kubeletGroup},
			ca:            ca,
		})

		// etcd client certificate
		// all nodes need to be able to talk to etcd b/c of calico
		m = append(m, certificateSpec{
			description: "etcd client",
			filename:    "etcd-client",
			commonName:  "etcd-client",
			ca:          ca,
		})
	}

	return m, nil
}

// returns a list of cert specs for the cluster described in the plan file
func (plan Plan) certSpecs(clusterCA *tls.CA, proxyClientCA *tls.CA) ([]certificateSpec, error) {
	m := []certificateSpec{}

	// Certificate for nodes
	nodes := plan.GetUniqueNodes()
	for _, n := range nodes {
		nodeManifest, err := n.certSpecs(plan, clusterCA)
		if err != nil {
			return nil, err
		}

		// Some nodes share common certificates between them. E.g. the kube-proxy client cert.
		// Before appending to the manifest, we ensure that this cert is not already in it.
		for _, s := range nodeManifest {
			if !certSpecInManifest(s, m) {
				m = append(m, s)
			}
		}
	}

	// Kube APIServer Kubelet Client certificate
	m = append(m, certificateSpec{
		description:   "kube-apiserver kubelet client",
		filename:      kubeAPIServerKubeletClientClientFilename,
		commonName:    kubeAPIServerKubeletClientClientCommonName,
		organizations: []string{adminGroup},
		ca:            clusterCA,
	})

	// Proxy Client certificate
	m = append(m, certificateSpec{
		description:   "proxy client",
		filename:      proxyClientCertFilename,
		commonName:    proxyClientCertCommonName,
		organizations: []string{adminGroup},
		ca:            proxyClientCA,
	})

	// Contiv certificates
	if plan.AddOns.CNI.Provider == cniProviderContiv {
		m = append(m, certificateSpec{
			description: "contiv proxy server",
			filename:    contivProxyServerCertFilename,
			commonName:  "auth-local.cisco.com", // using the same as contiv install script
			ca:          clusterCA,
		})
	}

	// Admin certificate
	m = append(m, certificateSpec{
		description:   "admin client",
		filename:      adminCertFilename,
		commonName:    adminUser,
		organizations: []string{adminGroup},
		ca:            clusterCA,
	})

	return m, nil
}
