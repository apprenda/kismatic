package aws

import (
	"bufio"
	"fmt"
	"os"

	"github.com/apprenda/kismatic/pkg/provision/retry"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// Ubuntu1604LTSEast is the AMI for Ubuntu 16.04 LTS
	Ubuntu1604LTSEast = AMI("ami-40d28157")
	// CentOS7East is the AMI for CentOS 7
	CentOS7East = AMI("ami-6d1c2007")
	// Redhat7East is the AMI for RedHat 7
	RedHat7East = AMI("ami-b63769a1")
)

// A Node on AWS
type Node struct {
	PrivateDNSName string
	PrivateIP      string
	PublicIP       string
	SSHUser        string
	ImageID        string
}

// AMI is the Amazon Machine Image
type AMI string

// InstanceType is the type of the Amazon machine
type InstanceType string

// ClientConfig of the AWS client
type ClientConfig struct {
	Region          string
	SubnetID        string
	Keyname         string
	SecurityGroupID string
}

// Credentials to be used for accessing the AI
type Credentials struct {
	ID     string
	Secret string
}

// Client for provisioning machines on AWS
type Client struct {
	Config      *ClientConfig
	Credentials Credentials
	ec2Client   *ec2.EC2
}

func (c *Client) getAPIClient() (*ec2.EC2, error) {
	if c.ec2Client == nil {
		creds := credentials.NewStaticCredentials(c.Credentials.ID, c.Credentials.Secret, "")
		_, err := creds.Get()
		if err != nil {
			return nil, fmt.Errorf("Error with credentials provided: %v", err)
		}
		config := aws.NewConfig().WithRegion(c.Config.Region).WithCredentials(creds).WithMaxRetries(10)
		c.ec2Client = ec2.New(session.New(config))
	}
	return c.ec2Client, nil
}

// CreateNode is for creating a machine on AWS using the given AMI and InstanceType.
// Returns the ID of the newly created machine.
func (c Client) CreateNode(ami AMI, instanceType InstanceType, size int64) (string, error) {
	api, err := c.getAPIClient()
	if err != nil {
		return "", err
	}
	req := &ec2.RunInstancesInput{
		ImageId: aws.String(string(ami)),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					VolumeSize:          aws.Int64(size),
				},
			},
		},
		InstanceType: aws.String(string(instanceType)),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(c.Config.Keyname),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			&ec2.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress: aws.Bool(true),
				DeviceIndex:              aws.Int64(0),
				SubnetId:                 aws.String(c.Config.SubnetID),
				Groups:                   []*string{aws.String(c.Config.SecurityGroupID)},
			},
		},
	}
	res, err := api.RunInstances(req)
	if err != nil {
		return "", err
	}
	instanceID := res.Instances[0].InstanceId
	// Modify the node
	modifyReq := &ec2.ModifyInstanceAttributeInput{
		InstanceId: instanceID,
		SourceDestCheck: &ec2.AttributeBooleanValue{
			Value: aws.Bool(false),
		},
	}
	_, err = api.ModifyInstanceAttribute(modifyReq)
	if err != nil {
		if err = c.DestroyNodes([]string{*instanceID}); err != nil {
			fmt.Printf("AWS NODE %q MUST BE CLEANED UP MANUALLY\n", instanceID)
		}
		return "", err
	}
	if err := c.tagResourceProvisionedBy(instanceID); err != nil {
		if err = c.DestroyNodes([]string{*instanceID}); err != nil {
			fmt.Printf("AWS NODE %q MUST BE CLEANED UP MANUALLY\n", *instanceID)
		}
		return "", err
	}

	return *res.Instances[0].InstanceId, nil
}

func (c Client) tagResourceProvisionedBy(resourceId *string) error {
	api, err := c.getAPIClient()
	if err != nil {
		return err
	}

	thisHost, _ := os.Hostname()
	tagReq := &ec2.CreateTagsInput{
		Resources: []*string{resourceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("ProvisionedBy"),
				Value: aws.String("Kismatic"),
			},
			{
				Key:   aws.String("CreatedBy"),
				Value: aws.String(thisHost),
			},
		},
	}
	return retry.WithBackoff(3, func() error {
		_, err = api.CreateTags(tagReq)
		return err
	})
}

func (c Client) TagResourceName(resourceId *string, name string) error {
	api, err := c.getAPIClient()
	if err != nil {
		return err
	}

	tagReq := &ec2.CreateTagsInput{
		Resources: []*string{resourceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}
	if _, err = api.CreateTags(tagReq); err != nil {
		return err
	}
	return nil
}

// GetNode returns information about a specific node. The consumer of this method
// is responsible for checking that the information it needs has been returned
// in the Node. (i.e. it's possible for the hostname, public IP to be empty)
func (c Client) GetNode(id string) (*Node, error) {
	api, err := c.getAPIClient()
	if err != nil {
		return nil, err
	}
	req := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	}
	resp, err := api.DescribeInstances(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Reservations) != 1 {
		return nil, fmt.Errorf("Attempted to get a single node, but API returned %d reservations", len(resp.Reservations))
	}
	if len(resp.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("Attempted to get a single node, but API returned %d instances", len(resp.Reservations[0].Instances))
	}
	instance := resp.Reservations[0].Instances[0]

	var publicIP string
	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}
	return &Node{
		PrivateDNSName: *instance.PrivateDnsName,
		PrivateIP:      *instance.PrivateIpAddress,
		PublicIP:       publicIP,
		SSHUser:        defaultSSHUserForAMI(AMI(*instance.ImageId)),
		ImageID:        *instance.ImageId,
	}, nil
}

// DestroyNodes destroys the nodes identified by the ID.
func (c Client) DestroyNodes(nodeIDs []string) error {
	api, err := c.getAPIClient()
	if err != nil {
		return err
	}
	req := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(nodeIDs),
	}

	fmt.Printf("Issuing termination requests for instances %v\n", nodeIDs)
	_, err = api.TerminateInstances(req)
	if err != nil {
		return err
	}
	return nil
}

func defaultSSHUserForAMI(ami AMI) string {
	switch ami {
	case Ubuntu1604LTSEast:
		return "ubuntu"
	case CentOS7East:
		return "centos"
	case RedHat7East:
		return "ec2-user"
	default:
		panic(fmt.Sprintf("unsupported AMI: %q", ami))
	}
}

func (c Client) GetNodes() ([]string, error) {
	thisHost, _ := os.Hostname()
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("pending")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:ProvisionedBy"),
			Values: []*string{aws.String("Kismatic")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:CreatedBy"),
			Values: []*string{aws.String(thisHost)},
		},
	}
	allids := []string{}

	request := ec2.DescribeInstancesInput{Filters: filters}
	client, err := c.getAPIClient()
	if err != nil {
		return allids, err
	}
	result, err := client.DescribeInstances(&request)
	if err != nil {
		return allids, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			allids = append(allids, *instance.InstanceId)
		}
	}
	return allids, nil
}

func (c *Client) MaybeProvisionKeypair(keyloc string) error {
	client, err := c.getAPIClient()
	if err != nil {
		return err
	}

	//look for an existing keypair
	q := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{aws.String(c.Config.Keyname)},
	}
	a, err := client.DescribeKeyPairs(q)

	switch err := err.(type) {
	case nil:
		if len(a.KeyPairs) > 0 {
			fmt.Printf("Found keypair %v\n", a.KeyPairs[0].KeyFingerprint)
		}
		return nil
	case awserr.Error:
		if err.Code() != "InvalidKeyPair.NotFound" {
			return err
		}
	default:
		return err
	}

	//if it isn't there, try to make it
	fmt.Printf("Creating new keypair %v\n", c.Config.Keyname)
	q2 := &ec2.CreateKeyPairInput{KeyName: aws.String(c.Config.Keyname)}
	a2, err := client.CreateKeyPair(q2)
	if err != nil {
		return err
	}

	//write newly created key to key dir
	fmt.Printf("Writing private key to %v\n", keyloc)
	f, err := os.Create(keyloc)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(*a2.KeyMaterial)

	w.Flush()
	os.Chmod(keyloc, 0600)

	return err
}

func (c *Client) MaybeProvisionVPC() (string, error) {
	client, err := c.getAPIClient()
	if err != nil {
		return "", err
	}
	//Look for tagged VPC
	q := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("tag:ProvisionedBy"),
				Values: []*string{aws.String("Kismatic")},
			},
		},
	}

	a, err := client.DescribeVpcs(q)
	if err != nil {
		return "", err
	}
	if len(a.Vpcs) > 0 {
		fmt.Println("Found tagged VPC")
		return *a.Vpcs[0].VpcId, nil
	}

	//make a new VPC
	q2 := &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	}

	fmt.Println("Creating new VPC")
	a2, err := client.CreateVpc(q2)
	if err != nil {
		return "", err
	}

	if err := c.tagResourceProvisionedBy(a2.Vpc.VpcId); err != nil {
		fmt.Println("Error tagging new VPC")
	}

	c.TagResourceName(a2.Vpc.VpcId, "Kismatic VPC")

	return *a2.Vpc.VpcId, nil
}

func (c *Client) MaybeProvisionRoute(vpc, igw, subnet string) (string, error) {
	client, err := c.getAPIClient()
	if err != nil {
		return "", err
	}

	q := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}

	a, err := client.DescribeRouteTables(q)
	if err != nil {
		return "", err
	}

	fmt.Printf("Found Route Table %v\n", *a.RouteTables[0].RouteTableId)

	for _, r := range a.RouteTables[0].Routes {
		if *r.GatewayId == igw {
			return *a.RouteTables[0].RouteTableId, nil //exit if already provisioned
		}
	}

	fmt.Printf("Creating route from Internet Gateway %v to Route %v\n", igw, *a.RouteTables[0].RouteTableId)
	q3 := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(igw),
		RouteTableId:         a.RouteTables[0].RouteTableId,
	}

	if _, err := client.CreateRoute(q3); err != nil {
		return "", err
	}

	q4 := &ec2.AssociateRouteTableInput{
		RouteTableId: a.RouteTables[0].RouteTableId,
		SubnetId:     aws.String(subnet),
	}

	fmt.Printf("Associating Subnet %v with Route %v\n", *q4.SubnetId, *q4.RouteTableId)
	if _, err := client.AssociateRouteTable(q4); err != nil {
		return "", err
	}

	if err := c.tagResourceProvisionedBy(a.RouteTables[0].RouteTableId); err != nil {
		fmt.Println("Error tagging new Route Table")
	}

	c.TagResourceName(a.RouteTables[0].RouteTableId, "Kismatic Route Table")

	return *a.RouteTables[0].RouteTableId, nil
}

func (c *Client) MaybeProvisionSubnet(vpc string) (string, error) {
	client, err := c.getAPIClient()
	if err != nil {
		return "", err
	}
	q := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}
	a, err := client.DescribeSubnets(q)
	if err != nil {
		return "", err
	}
	if len(a.Subnets) > 0 {
		return *a.Subnets[0].SubnetId, nil
	}

	q2 := &ec2.CreateSubnetInput{
		CidrBlock: aws.String("10.0.0.0/24"),
		VpcId:     aws.String(vpc),
	}
	fmt.Println("Creating new Subnet")
	a2, err := client.CreateSubnet(q2)
	if err != nil {
		return "", err
	}

	if err := c.tagResourceProvisionedBy(a2.Subnet.SubnetId); err != nil {
		fmt.Println("Error tagging new Subnet")
	}
	c.TagResourceName(a2.Subnet.SubnetId, "Kismatic Subnet")

	return *a2.Subnet.SubnetId, nil
}

func (c *Client) MaybeProvisionIG(vpc string) (string, error) {
	client, err := c.getAPIClient()
	if err != nil {
		return "", err
	}
	q := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("attachment.vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}
	a, err := client.DescribeInternetGateways(q)
	if err != nil {
		return "", err
	}
	if len(a.InternetGateways) > 0 {
		return *a.InternetGateways[0].InternetGatewayId, nil
	}

	q2 := &ec2.CreateInternetGatewayInput{}
	fmt.Println("Creating new Internet Gateway")
	a2, err := client.CreateInternetGateway(q2)
	if err != nil {
		return "", err
	}

	q3 := &ec2.AttachInternetGatewayInput{
		VpcId:             aws.String(vpc),
		InternetGatewayId: a2.InternetGateway.InternetGatewayId,
	}
	fmt.Printf("Attaching Internet Gateway %v to VPC %v\n", *q3.InternetGatewayId, *q3.VpcId)

	if _, err := client.AttachInternetGateway(q3); err != nil {
		return "", err
	}

	if err := c.tagResourceProvisionedBy(a2.InternetGateway.InternetGatewayId); err != nil {
		fmt.Println("Error tagging new Internet Gateway")
	}
	c.TagResourceName(a2.InternetGateway.InternetGatewayId, "Kismatic Internet Gateway")

	return *a2.InternetGateway.InternetGatewayId, nil
}

func (c *Client) MaybeProvisionSGs(vpc string) (string, error) {
	client, err := c.getAPIClient()
	if err != nil {
		return "", err
	}

	q := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}

	a, err := client.DescribeSecurityGroups(q)
	if err != nil {
		return "", err
	}

	fmt.Printf("Found Security Group %v\n", *a.SecurityGroups[0].GroupId)

	for _, t := range a.SecurityGroups[0].Tags {
		if *t.Key == "ProvisionedBy" && *t.Value == "Kismatic" {
			return *a.SecurityGroups[0].GroupId, nil //return already provisioned SG
		}
	}

	// q2 := &ec2.CreateSecurityGroupInput{
	// 	Description: aws.String("Kismatic Wide Open SG"),
	// 	GroupName:   aws.String("Kismatic Wide Open SG"),
	// 	VpcId:       aws.String(vpc),
	// }
	// fmt.Println("Creating new Security Group")
	// a2, err := client.CreateSecurityGroup(q2)
	// if err != nil {
	// 	return "", err
	// }

	q3 := &ec2.AuthorizeSecurityGroupIngressInput{
		IpProtocol: aws.String("-1"),
		GroupId:    a.SecurityGroups[0].GroupId,
		CidrIp:     aws.String("0.0.0.0/0"),
	}
	fmt.Println("Opening new SG to all incoming traffic")
	if _, err := client.AuthorizeSecurityGroupIngress(q3); err != nil {
		return "", err
	}
	// q4 := &ec2.AuthorizeSecurityGroupEgressInput{
	// 	IpProtocol: aws.String("-1"),
	// 	GroupId:    a2.GroupId,
	// 	CidrIp:     aws.String("0.0.0.0/0"),
	// }
	// fmt.Println("Opening new SG for all outgoing traffic")
	// if _, err := client.AuthorizeSecurityGroupEgress(q4); err != nil {
	// 	return "", err
	// }

	if err := c.tagResourceProvisionedBy(a.SecurityGroups[0].GroupId); err != nil {
		fmt.Println("Error tagging new Internet Gateway")
	}
	c.TagResourceName(a.SecurityGroups[0].GroupId, "Kismatic Wide Open SG")

	return *a.SecurityGroups[0].GroupId, err
}
