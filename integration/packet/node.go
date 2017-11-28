package packet

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/packethost/packngo"
)

// OS is an operating system supported on Packet
type OS string

const (
	// Ubuntu1604LTS OS image
	Ubuntu1604LTS = OS("ubuntu_16_04_image")
	// CentOS7 OS image
	CentOS7 = OS("centos_7_image")
)

// Client for managing infrastructure on Packet
type Client struct {
	Token     string
	ProjectID string

	apiClient *packngo.Client
}

// Node is a Packet.net node
type Node struct {
	ID          string
	Host        string
	PublicIPv4  string
	PrivateIPv4 string
	SSHUser     string
}

// CreateNode creates a node in packet with the given hostname and OS
func (c Client) CreateNode(hostname string, os OS, _ map[string]string) (*Node, error) {
	device := &packngo.DeviceCreateRequest{
		HostName:     hostname,
		OS:           string(os),
		Tags:         []string{"integration-test"},
		ProjectID:    c.ProjectID,
		Plan:         "baremetal_0",
		BillingCycle: "hourly",
		Facility:     "ewr1",
	}
	client := c.getAPIClient()
	dev, _, err := client.Devices.Create(device)
	if err != nil {
		return nil, err
	}
	node := &Node{
		ID: dev.ID,
	}
	return node, nil
}

func (c *Client) getAPIClient() *packngo.Client {
	if c.apiClient != nil {
		return c.apiClient
	}
	c.apiClient = packngo.NewClient("", c.Token, http.DefaultClient)
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
func (c Client) GetNode(deviceID string) (*Node, error) {
	client := c.getAPIClient()
	dev, _, err := client.Devices.Get(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %q: %v", deviceID, err)
	}
	if dev == nil {
		return nil, fmt.Errorf("did not get a device from server")
	}
	node := &Node{
		ID:          deviceID,
		Host:        dev.Hostname,
		PublicIPv4:  getPublicIPv4(dev),
		PrivateIPv4: getPrivateIPv4(dev),
		SSHUser:     "root",
	}
	return node, nil
}

// BlockUntilNodeAccessible blocks until the given node is accessible,
// or the timeout is reached.
func (c Client) BlockUntilNodeAccessible(deviceID string, timeout time.Duration, sshKey, sshUser string) error {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- true
	}()
	fmt.Printf("Waiting for node %s to be accessible", deviceID)
	// Loop until we get the node IP
	var nodeIP string
	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timed out waiting for node to be accessible")
		default:
			dev, _, err := c.getAPIClient().Devices.Get(deviceID)
			if err != nil {
				continue
			}
			nodeIP = getPublicIPv4(dev)
			fmt.Printf("\nGot node's IP: %s\n", nodeIP)
		}
		if nodeIP != "" {
			break
		}
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}
	// Loop until state is active
	fmt.Print("Waiting for node state to be 'active'")
	active := false
	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timedout waiting for node to be active")
		default:
			dev, _, err := c.getAPIClient().Devices.Get(deviceID)
			if err != nil {
				continue
			}
			if dev.State == "active" {
				active = true
			}
		}
		if active {
			break
		}
		fmt.Print(".")
		time.Sleep(30 * time.Second)
	}
	// Loop until timeout or ssh is accessible
	fmt.Printf("Waiting for SSH access")
	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timed out waiting for node to be accessible")
		default:
			if sshAccessible(nodeIP, sshKey, sshUser) {
				fmt.Printf("\nSSH is GO!\n")
				return nil
			}
		}
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}
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
