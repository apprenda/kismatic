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

func (aws AWS) getCommandEnvironment() []string {
	key := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", aws.AccessKeyID)
	secret := fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", aws.SecretAccessKey)
	return []string{key, secret}
}

// Provision the necessary infrastructure as described in the plan
func (aws AWS) Provision(plan install.Plan) (*install.Plan, error) {
	// Create directory for keeping cluster state
	clusterStateDir := aws.getClusterStateDir(plan.Cluster.Name)
	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory to keep cluster state: %v", err)
	}

	// Setup the environment for all Terraform commands.
	cmdEnv := append(os.Environ(), aws.getCommandEnvironment()...)
	cmdDir := clusterStateDir
	providerDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)

	// Generate SSH keypair using absolute paths
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	pubKeyPath := filepath.Join(wd, clusterStateDir, fmt.Sprintf("%s-ssh.pub", plan.Cluster.Name))
	privKeyPath := filepath.Join(wd, clusterStateDir, fmt.Sprintf("%s-ssh.pem", plan.Cluster.Name))
	if err := ssh.NewKeyPair(pubKeyPath, privKeyPath); err != nil {
		return nil, fmt.Errorf("error generating SSH key pair: %v", err)
	}
	plan.Cluster.SSH.Key = privKeyPath

	// Write out the terraform variables
	data := AWSTerraformData{
		Region:            plan.Provisioner.AWSOptions.Region,
		ClusterName:       plan.Cluster.Name,
		MasterCount:       len(plan.Master.Nodes),
		EtcdCount:         len(plan.Etcd.Nodes),
		WorkerCount:       len(plan.Worker.Nodes),
		IngressCount:      len(plan.Ingress.Nodes),
		StorageCount:      len(plan.Storage.Nodes),
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
	initCmd := exec.Command(aws.BinaryPath, "init", providerDir)
	initCmd.Env = cmdEnv
	initCmd.Dir = cmdDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		// TODO: We need to send this output somewhere else
		fmt.Println(string(out))
		return nil, fmt.Errorf("Error initializing terraform: %s", err)
	}

	// Terraform plan
	planCmd := exec.Command(aws.BinaryPath, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerDir)
	planCmd.Env = cmdEnv
	planCmd.Dir = cmdDir

	if out, err := planCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Error running terraform plan: %s", out)
	}

	// Terraform apply
	applyCmd := exec.Command(aws.BinaryPath, "apply", plan.Cluster.Name)
	applyCmd.Env = cmdEnv
	applyCmd.Dir = cmdDir
	if out, err := applyCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Error running terraform apply: %s", out)
	}

	provisionedPlan, err := aws.buildPopulatedPlan(plan)
	if err != nil {
		return nil, fmt.Errorf("error getting infrastructure details from Terraform: %v", err)
	}
	return provisionedPlan, nil
}

// updatePlan
func (aws *AWS) buildPopulatedPlan(plan install.Plan) (*install.Plan, error) {
	// Masters
	tfNodes, err := aws.getTerraformNodes(plan.Cluster.Name, "master")
	if err != nil {
		return nil, fmt.Errorf("error getting master node information: %v", err)
	}
	masterNodes := nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts)
	mng := install.MasterNodeGroup{
		ExpectedCount: masterNodes.ExpectedCount,
		Nodes:         masterNodes.Nodes,
	}
	mng.LoadBalancedFQDN = tfNodes.InternalIPs[0]
	mng.LoadBalancedShortName = tfNodes.IPs[0]
	plan.Master = mng

	// Etcds
	tfNodes, err = aws.getTerraformNodes(plan.Cluster.Name, "etcd")
	if err != nil {
		return nil, fmt.Errorf("error getting etcd node information: %v", err)
	}
	plan.Etcd = nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts)

	// Workers
	tfNodes, err = aws.getTerraformNodes(plan.Cluster.Name, "worker")
	if err != nil {
		return nil, fmt.Errorf("error getting worker node information: %v", err)
	}
	plan.Worker = nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts)

	// Ingress
	if plan.Ingress.ExpectedCount > 0 {
		tfNodes, err = aws.getTerraformNodes(plan.Cluster.Name, "ingress")
		if err != nil {
			return nil, fmt.Errorf("error getting ingress node information: %v", err)
		}
		plan.Ingress = install.OptionalNodeGroup(nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))
	}

	// Storage
	if plan.Storage.ExpectedCount > 0 {
		tfNodes, err = aws.getTerraformNodes(plan.Cluster.Name, "storage")
		if err != nil {
			return nil, fmt.Errorf("error getting storage node information: %v", err)
		}
		plan.Storage = install.OptionalNodeGroup(nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))
	}

	// SSH
	plan.Cluster.SSH.User = "ubuntu"
	return &plan, nil
}

// Destroy destroys a provisioned cluster (using -force by default)
func (aws AWS) Destroy(clusterName string) error {
	cmd := exec.Command(aws.BinaryPath, "destroy", "-force")
	cmd.Stdout = aws.Terraform.Output
	cmd.Stderr = aws.Terraform.Output
	cmd.Env = aws.getCommandEnvironment()
	cmd.Dir = aws.getClusterStateDir(clusterName)
	if err := cmd.Run(); err != nil {
		return errors.New("Error destroying infrastructure with Terraform")
	}
	return nil
}

func nodeGroupFromSlices(ips, internalIPs, hosts []string) install.NodeGroup {
	ng := install.NodeGroup{}
	ng.ExpectedCount = len(ips)
	ng.Nodes = []install.Node{}
	for i := range ips {
		n := install.Node{
			IP:   ips[i],
			Host: hosts[i],
		}
		if len(internalIPs) != 0 {
			n.InternalIP = internalIPs[i]
		}
		ng.Nodes = append(ng.Nodes, n)
	}
	return ng
}
