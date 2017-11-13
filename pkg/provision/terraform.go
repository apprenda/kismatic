package provision

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
	yaml "gopkg.in/yaml.v2"
)

const terraformBinaryPath = "../../bin/terraform"

// Terraform provisioner
type Terraform struct {
	BinaryPath string
	Logger     *log.Logger
}

// AWS provisioner for creating and destroying infrastructure on AWS.
type AWS struct {
	KeyID  string
	Secret string
	Terraform
}

func (aws AWS) getCommandEnvironment() []string {
	key := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", aws.KeyID)
	secret := fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", aws.Secret)
	return []string{key, secret}
}

// Provision the necessary infrastructure as described in the plan
func (aws AWS) Provision(plan install.Plan) (*install.Plan, error) {
	// Create directory for keeping cluster state
	clusterStateDir := fmt.Sprintf("terraform/clusters/%s/", plan.Cluster.Name)
	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory to keep cluster state: %v", err)
	}
	if err := os.Chdir(clusterStateDir); err != nil {
		return nil, fmt.Errorf("error switching dir to %s: %v", clusterStateDir, err)
	}
	defer os.Chdir("../../../")

	// Setup the environment for all Terraform commands.
	cmdEnv := append(os.Environ(), aws.getCommandEnvironment()...)

	providerDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)

	// Terraform init
	initCmd := exec.Command(terraformBinaryPath, "init", providerDir)
	initCmd.Env = cmdEnv
	if out, err := initCmd.CombinedOutput(); err != nil {
		// TODO: We need to send this output somewhere else
		fmt.Println(string(out))
		return nil, fmt.Errorf("Error initializing terraform: %s", err)
	}

	// Terraform plan
	planCmd := exec.Command(terraformBinaryPath, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerDir)
	planCmd.Env = cmdEnv

	if out, err := planCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Error running terraform plan: %s", out)
	}

	// Terraform apply
	applyCmd := exec.Command(terraformBinaryPath, "apply", plan.Cluster.Name)
	applyCmd.Env = cmdEnv
	if out, err := applyCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Error running terraform apply: %s", out)
	}

	// Render template
	outputCmd := exec.Command(terraformBinaryPath, "output", "rendered_template")
	outputCmd.Env = cmdEnv
	out, err := outputCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error collecting terraform output: %s", out)
	}
	var provisionedPlan install.Plan
	if err := yaml.Unmarshal(out, &provisionedPlan); err != nil {
		return nil, fmt.Errorf("error unmarshaling plan: %v", err)
	}
	return &provisionedPlan, nil
}

// Destroy the infrastructure that was provisioned for the given cluster
func (aws AWS) Destroy(clusterName string) error {
	clusterStateDir := fmt.Sprintf("terraform/clusters/%s/", clusterName)
	if err := os.Chdir(clusterStateDir); err != nil {
		return err
	}
	defer os.Chdir("../../../")
	cmd := exec.Command(terraformBinaryPath, "destroy", "-force")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Error attempting to destroy: %s", out)
	}
	return nil
}
