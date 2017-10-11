// +build packet_integration

package packet

import (
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// Create node
	client, err := newFromEnv()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	hostname := "testNode"
	osImage := CentOS7
	deviceID, err := client.CreateNode(hostname, osImage, USEast)
	if err != nil {
		t.Errorf("failed to create node: %v", err)
	}
	// Block until ssh is up
	timeout := 10 * time.Minute
	if _, err := client.GetSSHAccessibleNode(deviceID, timeout, os.Getenv("PACKET_SSH_KEY")); err != nil {
		t.Errorf("node did not become accessible")
	}
	// Delete node
	time.Sleep(5 * time.Second)
	if err := client.DeleteNode(deviceID); err != nil {
		t.Errorf("node %q was not deleted. MANUALLY CLEAN UP NODE IN PACKET.NET", deviceID)
	}
}
