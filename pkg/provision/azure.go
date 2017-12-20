package provision

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/ssh"
)

func (azure Azure) getCommandEnvironment() []string {
	subID := fmt.Sprintf("ARM_SUBSCRIPTION_ID=%s", azure.SubscriptionID)
	cID := fmt.Sprintf("ARM_CLIENT_ID=%s", azure.ClientID)
	cSecret := fmt.Sprintf("ARM_CLIENT_SECRET=%s", azure.ClientSecret)
	tID := fmt.Sprintf("ARM_TENANT_ID=%s", azure.TenantID)
	return []string{subID, cID, cSecret, tID}
}

// Provision the necessary infrastructure as described in the plan
func (azure Azure) Provision(plan install.Plan) (*install.Plan, error) {
	// Create directory for keeping cluster state
	clusterStateDir, err := azure.getClusterStateDir(plan.Cluster.Name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory to keep cluster state: %v", err)
	}

	// Setup the environment for all Terraform commands.
	cmdEnv := append(os.Environ(), azure.getCommandEnvironment()...)
	cmdDir := clusterStateDir
	providerDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)

	// Generate SSH keypair
	absPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	pubKeyPath := filepath.Join(absPath, fmt.Sprintf("/terraform/clusters/%s/%s-ssh.pub", plan.Cluster.Name, plan.Cluster.Name))
	privKeyPath := filepath.Join(absPath, fmt.Sprintf("/terraform/clusters/%s/%s-ssh.pem", plan.Cluster.Name, plan.Cluster.Name))
	if err := ssh.NewKeyPair(pubKeyPath, privKeyPath); err != nil {
		return nil, fmt.Errorf("error generating SSH key pair: %v", err)
	}
	plan.Cluster.SSH.Key = privKeyPath
	plan.Cluster.SSH.User = "ubuntu"

	// Write out the terraform variables
	data := AzureTerraformData{
		KismaticVersion:   azure.Terraform.KismaticVersion.String(),
		Location:          plan.Provisioner.AzureOptions.Location,
		ClusterName:       plan.Cluster.Name,
		ClusterOwner:      azure.Terraform.ClusterOwner,
		MasterCount:       plan.Master.ExpectedCount,
		EtcdCount:         plan.Etcd.ExpectedCount,
		WorkerCount:       plan.Worker.ExpectedCount,
		IngressCount:      plan.Ingress.ExpectedCount,
		StorageCount:      plan.Storage.ExpectedCount,
		SSHUser:           plan.Cluster.SSH.User,
		PrivateSSHKeyPath: privKeyPath,
		PublicSSHKeyPath:  pubKeyPath,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filepath.Join(clusterStateDir, "terraform.tfvars.json"), b, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing terraform variables: %v", err)
	}

	// Terraform init
	initCmd := exec.Command(azure.BinaryPath, "init", providerDir)
	initCmd.Env = cmdEnv
	initCmd.Dir = cmdDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		fmt.Fprintln(azure.Output, string(out))
		return nil, fmt.Errorf("Error initializing terraform: %s", err)
	}

	// Terraform plan
	planCmd := exec.Command(azure.BinaryPath, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerDir)
	planCmd.Env = cmdEnv
	planCmd.Dir = cmdDir

	if out, err := planCmd.CombinedOutput(); err != nil {
		fmt.Fprintln(azure.Output, string(out))
		return nil, fmt.Errorf("Error running terraform plan: %s", out)
	}

	// Terraform apply
	applyCmd := exec.Command(azure.BinaryPath, "apply", plan.Cluster.Name)
	applyCmd.Stdout = azure.Terraform.Output
	applyCmd.Stderr = azure.Terraform.Output
	applyCmd.Env = cmdEnv
	applyCmd.Dir = cmdDir
	if err := applyCmd.Run(); err != nil {
		return nil, fmt.Errorf("Error running terraform apply: %s", err)
	}

	// Update plan
	provisionedPlan, err := azure.buildPopulatedPlan(plan)
	if err != nil {
		return nil, err
	}
	return provisionedPlan, nil
}

// Destroy destroys a provisioned cluster (using -force by default)
func (azure Azure) Destroy(clusterName string) error {
	cmd := exec.Command(azure.BinaryPath, "destroy", "-force")
	cmd.Stdout = azure.Terraform.Output
	cmd.Stderr = azure.Terraform.Output
	cmd.Env = append(os.Environ(), azure.getCommandEnvironment()...)
	dir, err := azure.getClusterStateDir(clusterName)
	cmd.Dir = dir
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return errors.New("Error destroying infrastructure with Terraform")
	}
	return nil
}
