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
	AllowPackageInstallation     bool
	DisconnectedInstallation     bool
	AutoConfiguredDockerRegistry bool
	DockerRegistryIP             string
	DockerRegistryPort           int
	DockerRegistryCAPath         string
	ModifyHostsFiles             bool
	UseDirectLVM                 bool
	ServiceCIDR                  string
}

const planAWSOverlay = `cluster:
  name: kubernetes
  admin_password: abbazabba
  allow_package_installation: {{.AllowPackageInstallation}}
  disconnected_installation: {{.DisconnectedInstallation}}
  networking:
    type: overlay
    pod_cidr_block: 172.16.0.0/16
    service_cidr_block: {{if .ServiceCIDR}}{{.ServiceCIDR}}{{else}}172.20.0.0/16{{end}}
    update_hosts_files: {{.ModifyHostsFiles}}
  certificates:
    expiry: 17520h
    location_city: Troy
    location_state: New York
    location_country: US
  ssh:
    user: {{.SSHUser}}
    ssh_key: {{.SSHKeyFile}}
    ssh_port: 22{{if .UseDirectLVM}}
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
features:
  package_manager:
    enabled: true
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
