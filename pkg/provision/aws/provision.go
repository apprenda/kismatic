package aws

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/pkg/provision/plan"
)

const (
	Ubuntu1604LTS = LinuxDistro("ubuntu1604LTS")
	CentOS7       = LinuxDistro("centos7")
	Redhat7       = LinuxDistro("redhat7")

	AWSTargetRegion = "us-east-1"
	AWSKeyName      = "kismatic-integration-testing"
)

type infrastructureProvisioner interface {
	ProvisionNodes(NodeCount, LinuxDistro) (ProvisionedNodes, error)

	TerminateNodes(ProvisionedNodes) error

	TerminateAllNodes() error

	ForceProvision() error

	SSHKey() string
}

type LinuxDistro string

type NodeCount struct {
	Etcd   uint16
	Master uint16
	Worker uint16
}

func (nc NodeCount) Total() uint16 {
	return nc.Etcd + nc.Master + nc.Worker
}

type ProvisionedNodes struct {
	Etcd   []plan.Node
	Master []plan.Node
	Worker []plan.Node
}

func (p ProvisionedNodes) allNodes() []plan.Node {
	n := []plan.Node{}
	n = append(n, p.Etcd...)
	n = append(n, p.Master...)
	n = append(n, p.Worker...)
	return n
}

type sshMachineProvisioner struct {
	sshKey string
}

func (p sshMachineProvisioner) SSHKey() string {
	return p.sshKey
}

type awsProvisioner struct {
	sshMachineProvisioner
	client *Client
}

func AWSClientFromEnvironment() (*awsProvisioner, bool) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, false
	}
	c := Client{
		Config: &ClientConfig{
			Region:  AWSTargetRegion,
			Keyname: AWSKeyName,
		},
		Credentials: Credentials{
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
	overrideSecGroup := os.Getenv("AWS_SECURITY_GROUP_ID")
	if overrideSecGroup != "" {
		c.Config.SecurityGroupID = overrideSecGroup
	}
	overrideKeyName := os.Getenv("AWS_KEY_NAME")
	if overrideKeyName != "" {
		c.Config.Keyname = overrideKeyName
	}
	p := awsProvisioner{client: &c}
	p.sshKey = os.Getenv("AWS_SSH_KEY_PATH")
	if p.sshKey == "" {
		dir, _ := os.Getwd()
		p.sshKey = filepath.Join(dir, "kismatic.pem")
	}
	return &p, true
}

func (p awsProvisioner) TerminateAllNodes() error {
	nodes, err := p.client.GetNodes()
	if err != nil {
		return err
	}

	if len(nodes) > 0 {
		return p.client.DestroyNodes(nodes)
	}
	return nil
}

func (p *awsProvisioner) ForceProvision() error {
	if _, err := os.Stat(p.sshKey); os.IsNotExist(err) {
		if err := p.client.MaybeProvisionKeypair(p.sshKey); err != nil {
			return err
		}
	}

	if p.client.Config.SubnetID == "" || p.client.Config.SecurityGroupID == "" {
		vpc, err := p.client.MaybeProvisionVPC()
		if err != nil {
			return err
		}

		//maybe provision subnet
		sn, err := p.client.MaybeProvisionSubnet(vpc)
		if err != nil {
			return err
		}

		//maybe provision internet gateway
		ig, err := p.client.MaybeProvisionIG(vpc)
		if err != nil {
			return err
		}

		//maybe provision Routes
		_, err = p.client.MaybeProvisionRoute(vpc, ig, sn)
		if err != nil {
			return err
		}

		//maybe provision SGs
		sg, err := p.client.MaybeProvisionSGs(vpc)
		if err != nil {
			return err
		}

		os.Setenv("AWS_SUBNET_ID", sn)
		os.Setenv("AWS_SECURITY_GROUP_ID", sg)
	}

	return nil
}

func (p awsProvisioner) ProvisionNodes(blueprint NodeBlueprint, nodeCount NodeCount, distro LinuxDistro) (ProvisionedNodes, error) {
	var ami AMI
	switch distro {
	case Ubuntu1604LTS:
		ami = Ubuntu1604LTSEast
	case CentOS7:
		ami = CentOS7East
	case Redhat7:
		ami = RedHat7East
	default:
		panic(fmt.Sprintf("Used an unsupported distribution: %s", distro))
	}
	provisioned := ProvisionedNodes{}
	var i uint16
	for i = 0; i < nodeCount.Etcd; i++ {
		nodeID, err := p.client.CreateNode(ami, blueprint.EtcdInstanceType, blueprint.EtcdDisk)
		if err != nil {
			return provisioned, err
		}
		provisioned.Etcd = append(provisioned.Etcd, plan.Node{ID: nodeID})
	}
	for i = 0; i < nodeCount.Master; i++ {
		nodeID, err := p.client.CreateNode(ami, blueprint.MasterInstanceType, blueprint.MasterDisk)
		if err != nil {
			return provisioned, err
		}
		provisioned.Master = append(provisioned.Master, plan.Node{ID: nodeID})
	}
	for i = 0; i < nodeCount.Worker; i++ {
		nodeID, err := p.client.CreateNode(ami, blueprint.WorkerInstanceType, blueprint.WorkerDisk)
		if err != nil {
			return provisioned, err
		}
		provisioned.Worker = append(provisioned.Worker, plan.Node{ID: nodeID})
	}
	// Wait until all instances have their public IPs assigned
	for i := range provisioned.Etcd {
		etcd := &provisioned.Etcd[i]
		if err := p.updateNodeWithDeets(etcd.ID, etcd); err != nil {
			return provisioned, err
		}
	}
	for i := range provisioned.Master {
		master := &provisioned.Master[i]
		if err := p.updateNodeWithDeets(master.ID, master); err != nil {
			return provisioned, err
		}
	}
	for i := range provisioned.Worker {
		worker := &provisioned.Worker[i]
		if err := p.updateNodeWithDeets(worker.ID, worker); err != nil {
			return provisioned, err
		}
	}
	fmt.Println()
	return provisioned, nil
}

func (p awsProvisioner) updateNodeWithDeets(nodeID string, node *plan.Node) error {
	for {
		fmt.Print(".")
		awsNode, err := p.client.GetNode(nodeID)
		if err != nil {
			return err
		}
		node.PublicIPv4 = awsNode.PublicIP
		node.PrivateIPv4 = awsNode.PrivateIP
		node.SSHUser = awsNode.SSHUser

		node.Host = awsNode.PrivateDNSName
		if node.PublicIPv4 != "" && node.Host != "" && node.PrivateIPv4 != "" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (p awsProvisioner) TerminateNodes(runningNodes ProvisionedNodes) error {
	nodes := runningNodes.allNodes()
	nodeIDs := []string{}
	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.ID)
	}
	return p.client.DestroyNodes(nodeIDs)
}

func WaitForSSH(ProvisionedNodes ProvisionedNodes, sshKey string) error {
	nodes := ProvisionedNodes.allNodes()
	for _, n := range nodes {
		BlockUntilSSHOpen(n.PublicIPv4, n.SSHUser, sshKey)
	}
	fmt.Println()
	return nil
}
