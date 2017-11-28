package integration

type NFSVolume struct {
	Host string
}

type PlanAWS struct {
	Etcd                         []NodeDeets
	Master                       []NodeDeets
	Worker                       []NodeDeets
	Ingress                      []NodeDeets
	Storage                      []NodeDeets
	NFSVolume                    []NFSVolume
	MasterNodeFQDN               string
	MasterNodeShortName          string
	SSHUser                      string
	SSHKeyFile                   string
	HomeDirectory                string
	DisablePackageInstallation   bool
	DisconnectedInstallation     bool
	DockerRegistryServer         string
	DockerRegistryCAPath         string
	DockerRegistryUsername       string
	DockerRegistryPassword       string
	ModifyHostsFiles             bool
	HTTPProxy                    string
	HTTPSProxy                   string
	NoProxy                      string
	UseDirectLVM                 bool
	ServiceCIDR                  string
	DisableCNI                   bool
	CNIProvider                  string
	DisableHelm                  bool
	HeapsterReplicas             int
	HeapsterInfluxdbPVC          string
	CloudProvider                string
	KubeAPIServerOptions         map[string]string
	KubeControllerManagerOptions map[string]string
	KubeSchedulerOptions         map[string]string
	KubeProxyOptions             map[string]string
	KubeletOptions               map[string]string
}

// Certain fields are still present for backwards compatabilty when testing upgrades
const planAWSOverlay = `cluster:
  name: kubernetes
  admin_password: abbazabba
  disable_package_installation: {{.DisablePackageInstallation}}
  disconnected_installation: {{.DisconnectedInstallation}}
  networking:
    type: overlay                                                 # Required for KET <= v1.4.1
    pod_cidr_block: 172.16.0.0/16
    service_cidr_block: {{if .ServiceCIDR}}{{.ServiceCIDR}}{{else}}172.20.0.0/16{{end}}
    update_hosts_files: {{.ModifyHostsFiles}}
    http_proxy: {{.HTTPProxy}}
    https_proxy: {{.HTTPSProxy}}
    no_proxy: {{.NoProxy}}
  certificates:
    expiry: 17520h
    ca_expiry: 17520h
  ssh:
    user: {{.SSHUser}}
    ssh_key: {{.SSHKeyFile}}
    ssh_port: 22
  kube_apiserver:
    option_overrides: { {{ if .KubeAPIServerOptions }}{{ range $k, $v := .KubeAPIServerOptions }}"{{ $k }}": "{{ $v }}"{{end}}{{end}} }
  kube_controller_manager:
    option_overrides: { {{if .KubeControllerManagerOptions}}{{ range $k, $v := .KubeControllerManagerOptions }}"{{ $k }}": "{{ $v }}"{{end}}{{end}} }
  kube_scheduler: 
    option_overrides: { {{if .KubeSchedulerOptions}}{{ range $k, $v := .KubeSchedulerOptions }}"{{ $k }}": "{{ $v }}"{{end}}{{end}} }
  kube_proxy: 
    option_overrides: { {{if .KubeProxyOptions}}{{ range $k, $v := .KubeProxyOptions }}"{{ $k }}": "{{ $v }}"{{end}}{{end}} }
  kubelet: 
    option_overrides: { {{if .KubeletOptions}}{{ range $k, $v := .KubeletOptions }}"{{ $k }}": "{{ $v }}"{{end}}{{end}} }
  cloud_provider:
    provider: {{.CloudProvider}}{{if .UseDirectLVM}}
docker:
  storage:
    direct_lvm:
      enabled: true
      block_device: "/dev/xvdb"
      enable_deferred_deletion: false{{end}}
docker_registry:
  server: {{.DockerRegistryServer}}
  CA: {{.DockerRegistryCAPath}}
  username: {{.DockerRegistryUsername}}
  password: {{.DockerRegistryPassword}}
add_ons:
  cni:
    disable: {{.DisableCNI}}
    provider: {{if .CNIProvider}}{{.CNIProvider}}{{else}}calico{{end}}
    options:
      calico:
        mode: overlay
        log_level: info
  heapster:
    disable: false
    options:
      heapster_replicas: {{if eq .HeapsterReplicas 0}}2{{else}}{{.HeapsterReplicas}}{{end}}
      heapster:
        replicas: {{if eq .HeapsterReplicas 0}}2{{else}}{{.HeapsterReplicas}}{{end}}
        service_type: ClusterIP
        sink: influxdb:http://heapster-influxdb.kube-system.svc:8086
      influxdb:
        pvc_name: {{.HeapsterInfluxdbPVC}}
  package_manager:
    disable: {{.DisableHelm}}
    provider: helm
  rescheduler:
    disable: false
etcd:
  expected_count: {{len .Etcd}}
  nodes:{{range .Etcd}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}{{end}}
master:
  expected_count: {{len .Master}}
  nodes:{{range .Master}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}
    labels:
      com.integrationtest/master: true{{end}}
  load_balanced_fqdn: {{.MasterNodeFQDN}}
  load_balanced_short_name: {{.MasterNodeShortName}}
worker:
  expected_count: {{len .Worker}}
  nodes:{{range .Worker}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}
    labels:
      com.integrationtest/worker: true{{end}}
ingress:
  expected_count: {{len .Ingress}}
  nodes:{{range .Ingress}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}
    labels:
      com.integrationtest/ingress: true{{end}}
storage:
  expected_count: {{len .Storage}}
  nodes:{{range .Storage}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}
    labels:
      com.integrationtest/storage: true{{end}}
nfs:
  nfs_volume:{{range .NFSVolume}}
  - nfs_host: {{.Host}}
    mount_path: /{{end}}
`
