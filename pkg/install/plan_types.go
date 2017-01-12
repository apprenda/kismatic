package install

import "fmt"

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
}

type SSHConnection struct {
	SSHConfig *SSHConfig
	Node      *Node
}

func (p *Plan) GetUniqueNodeIPs() []string {
	ipMap := make(map[string]bool)
	nodes := p.getAllNodes()
	for _, node := range nodes {
		ipMap[node.IP] = true
	}

	ips := make([]string, len(ipMap))

	i := 0
	for k := range ipMap {
		ips[i] = k
		i++
	}

	return ips
}

func (p *Plan) getAllNodes() []Node {
	nodes := []Node{}
	nodes = append(nodes, p.Etcd.Nodes...)
	nodes = append(nodes, p.Master.Nodes...)
	nodes = append(nodes, p.Worker.Nodes...)
	if p.Ingress.Nodes != nil {
		nodes = append(nodes, p.Ingress.Nodes...)
	}
	return nodes
}

// GetSSHConnection returns the SSHConnection struct containing the node and SSHConfig details
func (p *Plan) GetSSHConnection(host string) (*SSHConnection, error) {
	nodes := p.getAllNodes()

	// try to find the node with the provided hostname
	var foundNode *Node
	for _, node := range nodes {
		if node.Host == host {
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
		}
	}

	if foundNode == nil {
		return nil, fmt.Errorf("node %q not found in the plan", host)
	}

	return &SSHConnection{&p.Cluster.SSH, foundNode}, nil
}

func firstIfItExists(nodes []Node) *Node {
	if len(nodes) > 0 {
		return &nodes[0]
	}
	return nil
}
