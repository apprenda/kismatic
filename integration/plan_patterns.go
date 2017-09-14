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
	DisableRegistrySeeding       bool
	AutoConfiguredDockerRegistry bool
	DockerRegistryIP             string
	DockerRegistryPort           int
	DockerRegistryCAPath         string
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
}

// Certain fields are still present for backwards compatabilty when testing upgrades
const planAWSOverlay = `cluster:
  name: kubernetes
  admin_password: abbazabba
  disable_package_installation: {{.DisablePackageInstallation}}
  disconnected_installation: {{.DisconnectedInstallation}}
  disable_registry_seeding: {{.DisableRegistrySeeding}}
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
    option_overrides: {}
  cloud_provider:
    provider: {{.CloudProvider}}{{if .UseDirectLVM}}
docker:
  storage:
    direct_lvm:
      enabled: true
      block_device: "/dev/xvdb"
      enable_deferred_deletion: false{{end}}
docker_registry:
  setup_internal: {{.AutoConfiguredDockerRegistry}}
  address: {{.DockerRegistryIP}}
  port: {{.DockerRegistryPort}}
  CA: {{.DockerRegistryCAPath}}
add_ons:
  cni:
    disable: {{.DisableCNI}}
    provider: {{if .CNIProvider}}{{.CNIProvider}}{{else}}calico{{end}}
    options:
      calico:
        mode: overlay
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
    internalip: {{.PrivateIP}}{{end}}
  load_balanced_fqdn: {{.MasterNodeFQDN}}
  load_balanced_short_name: {{.MasterNodeShortName}}
worker:
  expected_count: {{len .Worker}}
  nodes:{{range .Worker}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}{{end}}
ingress:
  expected_count: {{len .Ingress}}
  nodes:{{range .Ingress}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}{{end}}
storage:
  expected_count: {{len .Storage}}
  nodes:{{range .Storage}}
  - host: {{.Hostname}}
    ip: {{.PublicIP}}
    internalip: {{.PrivateIP}}{{end}}
nfs:
  nfs_volume:{{range .NFSVolume}}
  - nfs_host: {{.Host}}
    mount_path: /{{end}}
`
