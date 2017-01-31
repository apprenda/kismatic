package ansible

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

type ClusterCatalog struct {
	ClusterName               string `yaml:"kubernetes_cluster_name"`
	AdminPassword             string `yaml:"kubernetes_admin_password"`
	TLSDirectory              string `yaml:"tls_directory"`
	CalicoNetworkType         string `yaml:"calico_network_type"`
	ServicesCIDR              string `yaml:"kubernetes_services_cidr"`
	PodCIDR                   string `yaml:"kubernetes_pods_cidr"`
	DNSServiceIP              string `yaml:"kubernetes_dns_service_ip"`
	EnableModifyHosts         bool   `yaml:"modify_hosts_file"`
	EnableCalicoPolicy        bool   `yaml:"enable_calico_policy"`
	EnablePackageInstallation bool   `yaml:"allow_package_installation"`
	DisconnectedInstallation         bool   `yaml:"disconnected_installation"`
	KuberangPath              string `yaml:"kuberang_path"`
	LoadBalancedFQDN          string `yaml:"kubernetes_load_balanced_fqdn"`

	EnablePrivateDockerRegistry  bool   `yaml:"use_private_docker_registry"`
	EnableInternalDockerRegistry bool   `yaml:"setup_internal_docker_registry"`
	DockerCAPath                 string `yaml:"docker_certificates_ca_path"`
	DockerRegistryAddress        string `yaml:"docker_registry_address"`
	DockerRegistryPort           string `yaml:"docker_registry_port"`

	ForceEtcdRestart              bool `yaml:"force_etcd_restart"`
	ForceAPIServerRestart         bool `yaml:"force_apiserver_restart"`
	ForceControllerManagerRestart bool `yaml:"force_controller_manager_restart"`
	ForceSchedulerRestart         bool `yaml:"force_scheduler_restart"`
	ForceProxyRestart             bool `yaml:"force_proxy_restart"`
	ForceKubeletRestart           bool `yaml:"force_kubelet_restart"`
	ForceCalicoNodeRestart        bool `yaml:"force_calico_node_restart"`
	ForceDockerRestart            bool `yaml:"force_docker_restart"`

	EnableConfigureIngress bool `yaml:"configure_ingress"`

	KismaticPreflightCheckerLinux string `yaml:"kismatic_preflight_checker"`
	KismaticPreflightCheckerLocal string `yaml:"kismatic_preflight_checker_local"`

	WorkerNode string `yaml:"worker_node"`

	NFSVolumes []NFSVolume `yaml:"nfs_volumes"`

	EnableGluster bool `yaml:"configure_storage"`

	// volume add vars
	VolumeName              string `yaml:"volume_name"`
	VolumeReplicaCount      int    `yaml:"volume_replica_count"`
	VolumeDistributionCount int    `yaml:"volume_distribution_count"`
	VolumeStorageClass      string `yaml:"volume_storage_class"`
	VolumeQuotaGB           int    `yaml:"volume_quota_gb"`
	VolumeQuotaBytes        int    `yaml:"volume_quota_bytes"`
	VolumeMount             string `yaml:"volume_mount"`
	VolumeAllowedIPs        string `yaml:"volume_allow_ips"`
}

type NFSVolume struct {
	Host string
	Path string
}

func (c *ClusterCatalog) EnableRestart() {
	c.ForceEtcdRestart = true
	c.ForceAPIServerRestart = true
	c.ForceControllerManagerRestart = true
	c.ForceSchedulerRestart = true
	c.ForceProxyRestart = true
	c.ForceKubeletRestart = true
	c.ForceCalicoNodeRestart = true
	c.ForceDockerRestart = true
}

func (c *ClusterCatalog) ToYAML() ([]byte, error) {
	bytez, marshalErr := yaml.Marshal(c)
	if marshalErr != nil {
		return []byte{}, fmt.Errorf("error marshalling plan to yaml: %v", marshalErr)
	}

	return bytez, nil
}
