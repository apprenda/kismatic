package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic-platform/integration/aws"
	"github.com/apprenda/kismatic-platform/integration/packet"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	Ubuntu1604LTS = linuxDistro("ubuntu1604LTS")
	CentOS7       = linuxDistro("centos7")

	AWSTargetRegion     = "us-east-1"
	AWSSubnetID         = "subnet-25e13d08"
	AWSKeyName          = "kismatic-integration-testing"
	AWSSecurityGroupID  = "sg-d1dc4dab"
	AMIUbuntu1604USEAST = "ami-29f96d3e"
	AMICentos7UsEast    = "ami-6d1c2007"
)

type infrastructureProvisioner interface {
	ProvisionNodes(NodeCount, linuxDistro) (provisionedNodes, error)
	TerminateNodes(provisionedNodes) error
	SSHKey() string
}

type linuxDistro string

type NodeCount struct {
	Etcd   uint16
	Master uint16
	Worker uint16
}

func (nc NodeCount) Total() uint16 {
	return nc.Etcd + nc.Master + nc.Worker
}

type provisionedNodes struct {
	etcd   []NodeDeets
	master []NodeDeets
	worker []NodeDeets
}

func (p provisionedNodes) allNodes() []NodeDeets {
	n := []NodeDeets{}
	n = append(n, p.etcd...)
	n = append(n, p.master...)
	n = append(n, p.worker...)
	return n
}

type NodeDeets struct {
	id        string
	Hostname  string
	PublicIP  string
	PrivateIP string
	SSHUser   string
}

type sshMachineProvisioner struct {
	sshKey string
}

func (p sshMachineProvisioner) SSHKey() string {
	return p.sshKey
}

type awsProvisioner struct {
	sshMachineProvisioner
	client aws.Client
}

func AWSClientFromEnvironment() (infrastructureProvisioner, bool) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, false
	}
	c := aws.Client{
		Config: aws.ClientConfig{
			Region:          AWSTargetRegion,
			SubnetID:        AWSSubnetID,
			Keyname:         AWSKeyName,
			SecurityGroupID: AWSSecurityGroupID,
		},
		Credentials: aws.Credentials{
			ID:     accessKeyID,
			Secret: secretAccessKey,
		},
	}
	overrideRegion := os.Getenv("AWS_TARGET_REGION")
	if overrideRegion != "" {
		c.Config.Region = overrideRegion
	}
	overrideSubnet := os.Getenv("AWS_SUBNET_ID")
	if overrideSubnet != "" {
		c.Config.SubnetID = overrideSubnet
	}
	overrideKeyName := os.Getenv("AWS_KEY_NAME")
	if overrideKeyName != "" {
		c.Config.Keyname = overrideKeyName
	}
	overrideSecGroup := os.Getenv("AWS_SECURITY_GROUP_ID")
	if overrideSecGroup != "" {
		c.Config.SecurityGroupID = overrideSecGroup
	}
	p := awsProvisioner{client: c}
	p.sshKey = os.Getenv("AWS_SSH_KEY_PATH")
	if p.sshKey == "" {
		dir, _ := homedir.Dir()
		p.sshKey = filepath.Join(dir, ".ssh", "kismatic-integration-testing.pem")
	}
	return p, true
}

func (p awsProvisioner) ProvisionNodes(nodeCount NodeCount, distro linuxDistro) (provisionedNodes, error) {
	var ami aws.AMI
	switch distro {
	case Ubuntu1604LTS:
		ami = aws.Ubuntu1604LTSEast
	case CentOS7:
		ami = aws.CentOS7East
	default:
		panic(fmt.Sprintf("Used an unsupported distribution: %s", distro))
	}
	provisioned := provisionedNodes{}
	var i uint16
	for i = 0; i < nodeCount.Etcd; i++ {
		nodeID, err := p.client.CreateNode(ami, aws.T2Micro)
		if err != nil {
			return provisioned, err
		}
		provisioned.etcd = append(provisioned.etcd, NodeDeets{id: nodeID})
	}
	for i = 0; i < nodeCount.Master; i++ {
		nodeID, err := p.client.CreateNode(ami, aws.T2Micro)
		if err != nil {
			return provisioned, err
		}
		provisioned.master = append(provisioned.master, NodeDeets{id: nodeID})
	}
	for i = 0; i < nodeCount.Worker; i++ {
		nodeID, err := p.client.CreateNode(ami, aws.T2Medium)
		if err != nil {
			return provisioned, err
		}
		provisioned.worker = append(provisioned.worker, NodeDeets{id: nodeID})
	}
	// Wait until all instances have their public IPs assigned
	for i := range provisioned.etcd {
		etcd := &provisioned.etcd[i]
		node, err := p.waitForPublicIP(etcd.id)
		if err != nil {
			return provisioned, err
		}
		etcd.Hostname = node.Hostname
		etcd.PrivateIP = node.PrivateIP
		etcd.PublicIP = node.PublicIP
		etcd.SSHUser = node.SSHUser
	}
	for i := range provisioned.master {
		master := &provisioned.master[i]
		node, err := p.waitForPublicIP(master.id)
		if err != nil {
			return provisioned, err
		}
		master.Hostname = node.Hostname
		master.PrivateIP = node.PrivateIP
		master.PublicIP = node.PublicIP
		master.SSHUser = node.SSHUser
	}
	for i := range provisioned.worker {
		worker := &provisioned.worker[i]
		node, err := p.waitForPublicIP(worker.id)
		if err != nil {
			return provisioned, err
		}
		worker.Hostname = node.Hostname
		worker.PrivateIP = node.PrivateIP
		worker.PublicIP = node.PublicIP
		worker.SSHUser = node.SSHUser
	}
	return provisioned, nil
}

func (p awsProvisioner) waitForPublicIP(nodeID string) (*aws.Node, error) {
	for {
		fmt.Print(".")
		node, err := p.client.GetNode(nodeID)
		if err != nil {
			return nil, err
		}
		if node.PublicIP != "" {
			fmt.Println()
			return node, nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (p awsProvisioner) TerminateNodes(runningNodes provisionedNodes) error {
	nodes := runningNodes.allNodes()
	nodeIDs := []string{}
	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.id)
	}
	return p.client.DestroyNodes(nodeIDs)
}

type packetProvisioner struct {
	sshMachineProvisioner
	client packet.Client
}

func packetClientFromEnv() (infrastructureProvisioner, bool) {
	token := os.Getenv("PACKET_AUTH_TOKEN")
	projectID := os.Getenv("PACKET_PROJECT_ID")
	if token == "" || projectID == "" {
		return nil, false
	}
	p := packetProvisioner{
		client: packet.Client{
			Token:     token,
			ProjectID: projectID,
		},
	}
	p.sshKey = os.Getenv("PACKET_SSH_KEY_PATH")
	if p.sshKey == "" {
		dir, _ := homedir.Dir()
		p.sshKey = filepath.Join(dir, ".ssh", "packet-kismatic-integration-testing.pem")
	}
	return p, true
}

func (p packetProvisioner) ProvisionNodes(nodeCount NodeCount, distro linuxDistro) (provisionedNodes, error) {
	var packetDistro packet.OS
	switch distro {
	case Ubuntu1604LTS:
		packetDistro = packet.Ubuntu1604LTS
	case CentOS7:
		packetDistro = packet.CentOS7
	default:
		panic(fmt.Sprintf("Used an unsupported distribution: %s", distro))
	}
	// Create all the nodes
	nodes := provisionedNodes{}
	for i := uint16(0); i < nodeCount.Etcd; i++ {
		id, err := p.createNode(packetDistro, i)
		if err != nil {
			return nodes, err
		}
		nodes.etcd = append(nodes.etcd, NodeDeets{id: id})
	}
	for i := uint16(0); i < nodeCount.Master; i++ {
		id, err := p.createNode(packetDistro, i)
		if err != nil {
			return nodes, err
		}
		nodes.master = append(nodes.master, NodeDeets{id: id})
	}
	for i := uint16(0); i < nodeCount.Worker; i++ {
		id, err := p.createNode(packetDistro, i)
		if err != nil {
			return nodes, err
		}
		nodes.worker = append(nodes.worker, NodeDeets{id: id})
	}
	// Wait until all nodes are ready
	err := p.updateNodeUntilPublicIPAvailable(nodes.etcd)
	if err != nil {
		return nodes, err
	}
	err = p.updateNodeUntilPublicIPAvailable(nodes.master)
	if err != nil {
		return nodes, err
	}
	err = p.updateNodeUntilPublicIPAvailable(nodes.worker)
	if err != nil {
		return nodes, err
	}
	return nodes, nil
}

func (p packetProvisioner) TerminateNodes(nodes provisionedNodes) error {
	allNodes := append(nodes.etcd, nodes.master...)
	allNodes = append(allNodes, nodes.worker...)
	failedDeletes := []string{}
	for _, n := range allNodes {
		if err := p.client.DeleteNode(n.id); err != nil {
			failedDeletes = append(failedDeletes, n.Hostname)
		}
	}
	if len(failedDeletes) > 0 {
		return fmt.Errorf("FAILED TO DELETE THE FOLLOWING NODES ON PACKET: %v", failedDeletes)
	}
	return nil
}

func (p packetProvisioner) createNode(distro packet.OS, count uint16) (string, error) {
	hostname := fmt.Sprintf("kismatic-integration-%d-%d", time.Now().UnixNano(), count)
	node, err := p.client.CreateNode(hostname, distro)
	if err != nil {
		return "", err
	}
	return node.ID, nil
}

func (p packetProvisioner) updateNodeUntilPublicIPAvailable(nodes []NodeDeets) error {
	for i := range nodes {
		node := &nodes[i]
		nodeDeets, err := p.waitForPublicIP(node.id)
		if err != nil {
			return err
		}
		node.Hostname = nodeDeets.Host
		node.PrivateIP = nodeDeets.PrivateIPv4
		node.PublicIP = nodeDeets.PublicIPv4
		node.SSHUser = nodeDeets.SSHUser
	}
	return nil
}

func (p packetProvisioner) waitForPublicIP(nodeID string) (*packet.Node, error) {
	for {
		fmt.Printf(".")
		node, err := p.client.GetNode(nodeID)
		if err != nil {
			return nil, err
		}
		if node.PublicIPv4 != "" {
			return node, nil
		}
		time.Sleep(1 * time.Minute)
	}
}

func WaitForSSH(provisionedNodes provisionedNodes, sshKey string) error {
	nodes := provisionedNodes.allNodes()
	for _, n := range nodes {
		BlockUntilSSHOpen(n.PublicIP, n.SSHUser, sshKey)
	}
	return nil
}
