package provision

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
)

const terraformBinaryPath = "../../bin/terraform"

// Terraform provisioner
type Terraform struct {
	Output     io.Writer
	BinaryPath string
	Logger     *log.Logger
}

//An aggregate of different tfNodes (different fields, the same nodes)
//NOTE: these are organized a little differently than a traditional node group
//due to limitations of terraform. A tfNodeGroup organizes each field into
//parallel slices as opposed to a single slice with nodes containing the same data.
type tfNodeGroup struct {
	IPs         []string
	InternalIPs []string
	Hosts       []string
}

// For deserializing terraform output
type tfOutputVar struct {
	Sensitive  bool     `json:"sensitive"`
	OutputType string   `json:"type"`
	Value      []string `json:"value"`
}

// Provisioner is responsible for creating and destroying infrastructure for
// a given cluster.
type Provisioner interface {
	Provision(install.Plan) (*install.Plan, error)
	Destroy(string) error
}

// Creates a new terraform struct with specified logger.
func NewTerraform(logger *log.Logger) *Terraform {
	tf := &Terraform{}
	tf.BinaryPath = terraformBinaryPath
	tf.Logger = logger
	return tf
}

func (tf Terraform) getTerraformNodes(role string) (*tfNodeGroup, error) {
	tfOutPubIPs := fmt.Sprintf("%s_pub_ips", role)
	tfOutPrivIPs := fmt.Sprintf("%s_priv_ips", role)
	tfOutHosts := fmt.Sprintf("%s_hosts", role)

	nodes := &tfNodeGroup{}

	//Public IPs
	tfCmdOutputPub := exec.Command(tf.BinaryPath, "output", "-json", tfOutPubIPs)
	stdoutStderrPub, err := tfCmdOutputPub.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error collecting terraform output: %s", stdoutStderrPub)
	}
	pubIPData := tfOutputVar{}
	json.Unmarshal(stdoutStderrPub, &pubIPData)
	nodes.IPs = pubIPData.Value

	//Private IPs
	tfCmdOutputPriv := exec.Command(tf.BinaryPath, "output", "-json", tfOutPrivIPs)
	stdoutStderrPriv, err := tfCmdOutputPriv.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error collecting terraform output: %s", stdoutStderrPriv)
	}
	privIPData := tfOutputVar{}
	json.Unmarshal(stdoutStderrPriv, &privIPData)
	nodes.InternalIPs = privIPData.Value

	//Hosts
	tfCmdOutputHost := exec.Command(tf.BinaryPath, "output", "-json", tfOutHosts)
	stdoutStderrHost, err := tfCmdOutputHost.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error collecting terraform output: %s", stdoutStderrHost)
	}
	hostData := tfOutputVar{}
	json.Unmarshal(stdoutStderrHost, &hostData)
	nodes.Hosts = hostData.Value

	if len(nodes.IPs) != len(nodes.Hosts) {
		return nil, fmt.Errorf("Expected to get %d host names, but got %d", len(nodes.IPs), len(nodes.Hosts))
	}

	// Verify that we got the right number of internal IPs if we are using them
	if len(nodes.InternalIPs) != 0 && len(nodes.IPs) != len(nodes.InternalIPs) {
		return nil, fmt.Errorf("Expected to get %d internal IPs, but got %d", len(nodes.IPs), len(nodes.InternalIPs))
	}

	return nodes, nil
}

func (t Terraform) getClusterStateDir(clusterName string) string {
	return fmt.Sprintf("terraform/clusters/%s/", clusterName)
}
