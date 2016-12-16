package integration

type PlanAWS struct {
	Etcd                         []NodeDeets
	Master                       []NodeDeets
	Worker                       []NodeDeets
	Ingress                      []NodeDeets
	MasterNodeFQDN               string
	MasterNodeShortName          string
	SSHUser                      string
	SSHKeyFile                   string
	HomeDirectory                string
	AllowPackageInstallation     bool
	AutoConfiguredDockerRegistry bool
	DockerRegistryIP             string
	DockerRegistryPort           int
	DockerRegistryCAPath         string
}

const planAWSOverlay = `cluster:
  name: kubernetes
  admin_password: abbazabba
  allow_package_installation: {{.AllowPackageInstallation}}
  networking:
    type: overlay
    pod_cidr_block: 172.16.0.0/16
    service_cidr_block: 172.17.0.0/16
    policy_enabled: false
    update_hosts_files: false
  certificates:
    expiry: 17520h
    location_city: Troy
    location_state: New York
    location_country: US
  ssh:
    user: {{.SSHUser}}
    ssh_key: {{.SSHKeyFile}}
    ssh_port: 22
docker_registry:
  setup_internal: {{.AutoConfiguredDockerRegistry}}
  address: {{.DockerRegistryIP}}
  port: {{.DockerRegistryPort}}
  CA: {{.DockerRegistryCAPath}}
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
`
