package ansible

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

type ClusterCatalog struct {
	Versions struct {
		Kubernetes    string `yaml:"kubernetes"`
		KubernetesYum string `yaml:"kubernetes_yum"`
		KubernetesDeb string `yaml:"kubernetes_deb"`
	}

	ClusterName               string `yaml:"kubernetes_cluster_name"`
	AdminPassword             string `yaml:"kubernetes_admin_password"`
	TLSDirectory              string `yaml:"tls_directory"`
	ServicesCIDR              string `yaml:"kubernetes_services_cidr"`
	PodCIDR                   string `yaml:"kubernetes_pods_cidr"`
	DNSServiceIP              string `yaml:"kubernetes_dns_service_ip"`
	EnableModifyHosts         bool   `yaml:"modify_hosts_file"`
	EnablePackageInstallation bool   `yaml:"allow_package_installation"`
	DisconnectedInstallation  bool   `yaml:"disconnected_installation"`
	KuberangPath              string `yaml:"kuberang_path"`
	LoadBalancedFQDN          string `yaml:"kubernetes_load_balanced_fqdn"`

	APIServerOptions             map[string]string `yaml:"kubernetes_api_server_option_overrides"`
	KubeControllerManagerOptions map[string]string `yaml:"kube_controller_manager_option_overrides"`
	KubeSchedulerOptions         map[string]string `yaml:"kube_scheduler_option_overrides"`
	KubeProxyOptions             map[string]string `yaml:"kube_proxy_option_overrides"`
	KubeletOptions               map[string]string `yaml:"kubelet_overrides"`

	ConfigureDockerWithPrivateRegistry bool   `yaml:"configure_docker_with_private_registry"`
	DockerRegistryCAPath               string `yaml:"docker_certificates_ca_path"`
	DockerRegistryServer               string `yaml:"docker_registry_full_url"`
	DockerRegistryUsername             string `yaml:"docker_registry_username"`
	DockerRegistryPassword             string `yaml:"docker_registry_password"`

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

	NewNode string `yaml:"new_node"`

	NFSVolumes []NFSVolume `yaml:"nfs_volumes"`

	EnableGluster bool `yaml:"configure_storage"`

	// volume add vars
	VolumeName              string   `yaml:"volume_name"`
	VolumeReplicaCount      int      `yaml:"volume_replica_count"`
	VolumeDistributionCount int      `yaml:"volume_distribution_count"`
	VolumeStorageClass      string   `yaml:"volume_storage_class"`
	VolumeQuotaGB           int      `yaml:"volume_quota_gb"`
	VolumeQuotaBytes        int      `yaml:"volume_quota_bytes"`
	VolumeMount             string   `yaml:"volume_mount"`
	VolumeAllowedIPs        string   `yaml:"volume_allow_ips"`
	VolumeReclaimPolicy     string   `yaml:"volume_reclaim_policy"`
	VolumeAccessModes       []string `yaml:"volume_access_modes"`

	TargetVersion string `yaml:"kismatic_short_version"`

	OnlineUpgrade bool `yaml:"online_upgrade"`

	DiagnosticsDirectory string `yaml:"diagnostics_dir"`
	DiagnosticsDateTime  string `yaml:"diagnostics_date_time"`

	Docker struct {
		Enabled bool
		Logs    struct {
			Driver string            `yaml:"driver"`
			Opts   map[string]string `yaml:"opts"`
		}
		Storage struct {
			Driver               string               `yaml:"driver"`
			Opts                 map[string]string    `yaml:"opts"`
			OptsList             []string             `yaml:"opts_list"`
			DirectLVMBlockDevice DirectLVMBlockDevice `yaml:"direct_lvm_block_device"`
		}
	}

	LocalKubeconfigDirectory string `yaml:"local_kubeconfig_directory"`

	CloudProvider string `yaml:"cloud_provider"`
	CloudConfig   string `yaml:"cloud_config_local"`

	DNS struct {
		Enabled  bool
		Provider string
	}

	RunPodValidation bool `yaml:"run_pod_validation"`

	CNI struct {
		Enabled  bool
		Provider string
		Options  struct {
			Calico struct {
				Mode          string
				LogLevel      string `yaml:"log_level"`
				WorkloadMTU   int    `yaml:"workload_mtu"`
				FelixInputMTU int    `yaml:"felix_input_mtu"`
			}
		}
	}

	Heapster struct {
		Enabled bool
		Options struct {
			Heapster struct {
				Replicas    int    `yaml:"replicas"`
				Sink        string `yaml:"sink"`
				ServiceType string `yaml:"service_type"`
			}
			InfluxDB struct {
				PVCName string `yaml:"pvc_name"`
			}
		}
	}

	Dashboard struct {
		Enabled bool
	}

	Helm struct {
		Enabled   bool
		Namespace string
	}

	Rescheduler struct {
		Enabled bool
	}

	InsecureNetworkingEtcd bool `yaml:"insecure_networking_etcd"`

	HTTPProxy  string `yaml:"http_proxy"`
	HTTPSProxy string `yaml:"https_proxy"`
	NoProxy    string `yaml:"no_proxy"`

	NodeLabels         map[string][]string          `yaml:"node_labels"`
	KubeletNodeOptions map[string]map[string]string `yaml:"kubelet_node_overrides"`
}

type DirectLVMBlockDevice struct {
	Path                        string
	ThinpoolPercent             string `yaml:"thinpool_percent"`
	ThinpoolMetaPercent         string `yaml:"thinpool_metapercent"`
	ThinpoolAutoextendThreshold string `yaml:"thinpool_autoextend_threshold"`
	ThinpoolAutoextendPercent   string `yaml:"thinpool_autoextend_percent"`
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
