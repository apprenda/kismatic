package packet

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/pkg/provision/plan"
	"github.com/packethost/packngo"
)

// OS is an operating system supported on Packet
type OS string

// Region where nodes are deployed
type Region string

const (
	// Ubuntu1604LTS OS image
	Ubuntu1604LTS = OS("ubuntu_16_04_image")
	// CentOS7 OS image
	CentOS7 = OS("centos_7_image")
	// USEast region
	USEast = Region("ewr1")
	// USWest region
	USWest = Region("sjc1")
	// EUWest region
	EUWest = Region("ams1")
)

// Client for managing infrastructure on Packet
type Client struct {
	APIKey    string
	ProjectID string
	SSHKey    string

	apiClient *packngo.Client
}

func newFromEnv() (*Client, error) {
	apiKey := os.Getenv("PACKET_API_KEY")
	projectID := os.Getenv("PACKET_PROJECT_ID")
	if apiKey == "" || projectID == "" {
		return nil, errors.New("PACKET_API_KEY and PACKET_PROJECT_ID are required environment variables")
	}
	sshKey := os.Getenv("PACKET_SSH_KEY_PATH")
	if sshKey == "" {
		cwd, _ := os.Getwd()
		sshKey = filepath.Join(cwd, "kismatic-packet.pem")
	}
	return &Client{
		APIKey:    apiKey,
		ProjectID: projectID,
		SSHKey:    sshKey,
	}, nil
}

// CreateNode creates a node in packet with the given hostname and OS
func (c Client) CreateNode(hostname string, os OS, region Region) (string, error) {
	device := &packngo.DeviceCreateRequest{
		HostName:     hostname,
		OS:           string(os),
		Tags:         []string{"integration-test"},
		ProjectID:    c.ProjectID,
		Plan:         "baremetal_0",
		BillingCycle: "hourly",
		Facility:     string(region),
	}
	client := c.getAPIClient()
	dev, _, err := client.Devices.Create(device)
	if err != nil {
		return "", err
	}
	return dev.ID, nil
}

func (c *Client) getAPIClient() *packngo.Client {
	if c.apiClient != nil {
		return c.apiClient
	}
	c.apiClient = packngo.NewClient("", c.APIKey, http.DefaultClient)
	return c.apiClient
}

// DeleteNode deletes the node that matches the given ID
func (c Client) DeleteNode(deviceID string) error {
	client := c.getAPIClient()
	resp, err := client.Devices.Delete(deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete node with ID %q", deviceID)
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete node with ID %q", deviceID)
	}
	return nil
}

// GetNode returns the node that matches the given ID

func (c Client) GetNode(deviceID string) (*plan.Node, error) {
	client := c.getAPIClient()
	dev, _, err := client.Devices.Get(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %q: %v", deviceID, err)
	}
	if dev == nil {
		return nil, fmt.Errorf("did not get a device from server")
	}
	node := &plan.Node{
		ID:          deviceID,
		Host:        dev.Hostname,
		PublicIPv4:  getPublicIPv4(dev),
		PrivateIPv4: getPrivateIPv4(dev),
		SSHUser:     "root",
	}
	return node, nil
}

// GetSSHAccessibleNode blocks until the node is accessible via SSH and returns the node's information.
func (c Client) GetSSHAccessibleNode(deviceID string, timeout time.Duration, sshKey string) (*plan.Node, error) {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- true
	}()
	// Loop until we get the node's public IP
	var node *plan.Node
	var err error
	for {
		select {
		case <-timeoutChan:
			return nil, fmt.Errorf("timed out waiting for node to be accessible")
		default:
			node, err = c.GetNode(deviceID)
			if err != nil {
				continue
			}
		}
		if node.PublicIPv4 != "" {
			break
		}
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}
	// Loop until ssh is accessible
	for {
		select {
		case <-timeoutChan:
			return nil, fmt.Errorf("timed out waiting for node to be accessible")
		default:
			if sshAccessible(node.PublicIPv4, sshKey, node.SSHUser) {
				return node, nil
			}
		}
		fmt.Print(".")
		time.Sleep(10 * time.Second)
	}
}

func (c Client) ListNodes() ([]plan.Node, error) {
	client := c.getAPIClient()
	devices, _, err := client.Devices.List(c.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %v", err)
	}
	nodes := []plan.Node{}
	for _, d := range devices {
		n := plan.Node{
			ID:          d.ID,
			Host:        d.Hostname,
			PublicIPv4:  getPublicIPv4(&d),
			PrivateIPv4: getPrivateIPv4(&d),
			SSHUser:     "root",
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func getPublicIPv4(device *packngo.Device) string {
	for _, net := range device.Network {
		if net.Public != true || net.AddressFamily != 4 {
			continue
		}
		if net.Address != "" {
			return net.Address
		}
	}
	return ""
}

func getPrivateIPv4(device *packngo.Device) string {
	for _, net := range device.Network {
		if net.Public == true || net.AddressFamily != 4 {
			continue
		}
		if net.Address != "" {
			return net.Address
		}
	}
	return ""
}

func sshAccessible(ip string, sshKey, sshUser string) bool {
	cmd := exec.Command("ssh")
	cmd.Args = append(cmd.Args, "-i", sshKey)
	cmd.Args = append(cmd.Args, "-o", "ConnectTimeout=5")
	cmd.Args = append(cmd.Args, "-o", "BatchMode=yes")
	cmd.Args = append(cmd.Args, "-o", "StrictHostKeyChecking=no")
	cmd.Args = append(cmd.Args, fmt.Sprintf("%s@%s", sshUser, ip), "exit") // just call exit if we are able to connect
	err := cmd.Run()
	return err == nil
}
