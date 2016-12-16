package aws

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/apprenda/kismatic/integration/retry"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

const (
	// StateAvailable is the AWS string returned when machine is available
	StateAvailable = ec2.StateAvailable
	// Ubuntu1604LTSEast is the AMI for Ubuntu 16.04 LTS
	Ubuntu1604LTSEast = AMI("ami-29f96d3e")
	// CentOS7East is the AMI for CentOS 7
	CentOS7East = AMI("ami-6d1c2007")
	// RedHat7East is the AMI for RedHat 7
	RedHat7East = AMI("ami-b63769a1")
	// T2Micro is the T2 Micro instance type
	T2Micro = InstanceType(ec2.InstanceTypeT2Micro)
	// T2Medium is the T2 Medium instance type
	T2Medium = InstanceType(ec2.InstanceTypeT2Medium)
	// exponentialBackoffMaxAttempts is the number of times will try before failing
	// Exponential backoff for AWS eventual consistency
	exponentialBackoffMaxEC2Attempts     = 5
	exponentialBackoffMaxRoute53Attempts = 5
)

// A Node on AWS
type Node struct {
	PrivateDNSName string
	PrivateIP      string
	PublicIP       string
	SSHUser        string
	State          string
}

// DNSRecord in Router53 on AWS
type DNSRecord struct {
	Name   string
	Values []string
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
	HostedZoneID    string
}

// Credentials to be used for accessing the AI
type Credentials struct {
	ID     string
	Secret string
}

// Client for provisioning machines on AWS
type Client struct {
	Config      ClientConfig
	Credentials Credentials
	session     *session.Session
}

func (c *Client) getEC2APIClient() (*ec2.EC2, error) {
	if err := c.prepareSession(); err != nil {
		return nil, err
	}
	return ec2.New(c.session), nil
}

func (c *Client) getRoute53APIClient() (*route53.Route53, error) {
	if err := c.prepareSession(); err != nil {
		return nil, err
	}
	return route53.New(c.session), nil
}

func (c *Client) prepareSession() error {
	if c.session == nil {
		creds := credentials.NewStaticCredentials(c.Credentials.ID, c.Credentials.Secret, "")
		_, err := creds.Get()
		if err != nil {
			return fmt.Errorf("Error with credentials provided: %v", err)
		}
		config := aws.NewConfig().WithRegion(c.Config.Region).WithCredentials(creds).WithMaxRetries(10)
		c.session = session.New(config)
	}
	return nil
}

// CreateNode is for creating a machine on AWS using the given AMI and InstanceType.
// Returns the ID of the newly created machine.
func (c Client) CreateNode(ami AMI, instanceType InstanceType) (string, error) {
	api, err := c.getEC2APIClient()
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
					VolumeSize:          aws.Int64(10),
				},
			},
		},
		InstanceType:     aws.String(string(instanceType)),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SubnetId:         aws.String(c.Config.SubnetID),
		KeyName:          aws.String(c.Config.Keyname),
		SecurityGroupIds: []*string{aws.String(c.Config.SecurityGroupID)},
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
	err = retry.WithBackoff(func() error {
		var err2 error
		_, err2 = api.ModifyInstanceAttribute(modifyReq)
		return err2
	}, exponentialBackoffMaxEC2Attempts)
	if err != nil {
		fmt.Println("Failed to modify instance attributes")
		if err = c.DestroyNodes([]string{*instanceID}); err != nil {
			fmt.Printf("AWS NODE %q MUST BE CLEANED UP MANUALLY\n", *instanceID)
		}
		return "", err
	}
	// Tag the nodes
	thisHost, _ := os.Hostname()
	tagReq := &ec2.CreateTagsInput{
		Resources: []*string{instanceID},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("ApprendaTeam"),
				Value: aws.String("Kismatic"),
			},
			{
				Key:   aws.String("CreatedBy"),
				Value: aws.String(thisHost),
			},
		},
	}
	err = retry.WithBackoff(func() error {
		var err2 error
		_, err2 = api.CreateTags(tagReq)
		return err2
	}, exponentialBackoffMaxEC2Attempts)
	if err != nil {
		fmt.Println("Failed to tag instance")
		if err = c.DestroyNodes([]string{*instanceID}); err != nil {
			fmt.Printf("AWS NODE %q MUST BE CLEANED UP MANUALLY\n", *instanceID)
		}
		return "", err
	}
	return *res.Instances[0].InstanceId, nil
}

// GetNode returns information about a specific node. The consumer of this method
// is responsible for checking that the information it needs has been returned
// in the Node. (i.e. it's possible for the hostname, public IP to be empty)
func (c Client) GetNode(id string) (*Node, error) {
	api, err := c.getEC2APIClient()
	if err != nil {
		return nil, err
	}
	req := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	}
	var resp *ec2.DescribeInstancesOutput
	err = retry.WithBackoff(func() error {
		var err2 error
		resp, err2 = api.DescribeInstances(req)
		return err2
	}, exponentialBackoffMaxEC2Attempts)
	if err != nil {
		fmt.Println("Failed to get node information")
		return nil, err
	}
	if len(resp.Reservations) != 1 {
		return nil, fmt.Errorf("Attempted to get a single node, but API returned %d reservations", len(resp.Reservations))
	}
	if len(resp.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("Attempted to get a single node, but API returned %d instances", len(resp.Reservations[0].Instances))
	}
	instance := resp.Reservations[0].Instances[0]

	var privateDNSName string
	if instance.PrivateDnsName != nil {
		privateDNSName = *instance.PrivateDnsName
	}
	var privateIP string
	if instance.PrivateIpAddress != nil {
		privateIP = *instance.PrivateIpAddress
	}
	var publicIP string
	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}

	return &Node{
		PrivateDNSName: privateDNSName,
		PrivateIP:      privateIP,
		PublicIP:       publicIP,
		SSHUser:        defaultSSHUserForAMI(AMI(*instance.ImageId)),
		State:          instance.State.GoString(),
	}, nil
}

// DestroyNodes destroys the nodes identified by the ID.
func (c Client) DestroyNodes(nodeIDs []string) error {
	api, err := c.getEC2APIClient()
	if err != nil {
		return err
	}
	req := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(nodeIDs),
	}
	_, err = api.TerminateInstances(req)
	if err != nil {
		return err
	}
	return nil
}

// CreateDNSRecords generates an Route53 configured with the master nodes
func (c Client) CreateDNSRecords(nodeIPs []string) (*DNSRecord, error) {
	api, err := c.getRoute53APIClient()
	if err != nil {
		return nil, err
	}
	// Setup variables to modify hosted zone
	name := strconv.FormatInt(time.Now().Unix(), 10) + ".kismatic.integration."
	dnsRecord := &DNSRecord{Name: name, Values: nodeIPs}
	err = modifyHostedZone(dnsRecord, route53.ChangeActionUpsert, c.Config.HostedZoneID, api)
	if err != nil {
		return nil, err
	}

	return dnsRecord, nil
}

// GetDNSRecords returns all Record Sets for the Hosted Zone
func (c Client) GetDNSRecords() ([]*route53.ResourceRecordSet, error) {
	api, err := c.getRoute53APIClient()
	if err != nil {
		return nil, err
	}
	req := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(c.Config.HostedZoneID), // Required
	}

	resp, err := api.ListResourceRecordSets(req)
	if err != nil {
		return nil, err
	}
	return resp.ResourceRecordSets, nil
}

// DeleteDNSRecords deletes the specified Record Set
func (c Client) DeleteDNSRecords(dnsRecord *DNSRecord) error {
	api, err := c.getRoute53APIClient()
	if err != nil {
		return err
	}
	err = modifyHostedZone(dnsRecord, route53.ChangeActionDelete, c.Config.HostedZoneID, api)
	if err != nil {
		return err
	}

	return nil
}

func modifyHostedZone(record *DNSRecord, action string, hostedZoneID string, api *route53.Route53) error {
	if len(record.Values) == 0 && action != route53.ChangeActionDelete {
		return fmt.Errorf("Values cannot be empty when setting up DNS records")
	}
	var records []*route53.ResourceRecord
	for _, value := range record.Values {
		records = append(records, &route53.ResourceRecord{Value: aws.String(value)})
	}

	req := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(record.Name),
						Type:            aws.String(route53.RRTypeA),
						ResourceRecords: records,
						TTL:             aws.Int64(300),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneID),
	}

	resp, err := api.ChangeResourceRecordSets(req)
	if err != nil {
		return err
	}
	changeID := resp.ChangeInfo.Id
	if changeID == nil || *changeID == "" {
		return fmt.Errorf("Something went wrong, DNS change ID is nil")
	}

	changeReq := &route53.GetChangeInput{
		Id: aws.String(*changeID),
	}

	err = retry.WithBackoff(func() error {
		changeResp, err2 := api.GetChange(changeReq)
		if err2 != nil {
			return err2
		}
		if *changeResp.ChangeInfo.Status != route53.ChangeStatusInsync {
			return fmt.Errorf("DNS change status is still %s, took too long", *changeResp.ChangeInfo.Status)
		}
		return nil
	}, exponentialBackoffMaxRoute53Attempts)
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
