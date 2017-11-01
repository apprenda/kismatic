package provision

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/ssh"
)

func (aws *AWS) getCommandEnvironment() []string {
	key := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", os.Getenv("AWS_ACCESS_KEY_ID"))
	secret := fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	region := fmt.Sprintf("AWS_DEFAULT_REGION=%s", os.Getenv("AWS_DEFAULT_REGION"))
	return []string{key, secret, region}
}

//This might be a better fit with Terraform as the receiver,
//but I think other providers will expect keys not in this format.
func (aws *AWS) getAWSKeyData(pubFile, privFile string) (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("Could not find current working dir: %v", err)
	}
	privPath := fmt.Sprintf("%s/%s", wd, privFile)
	pubBytes, err := ioutil.ReadFile(pubFile)
	if err != nil {
		return "", "", fmt.Errorf("Error reading public ssh key file: %v", err)
	}
	pubData := strings.TrimSpace(string(pubBytes))
	return pubData, privPath, nil
}

//Provision provides a wrapper for terraform init, terraform plan, and terraform apply.
func (aws *AWS) Provision(out io.Writer, plan *install.Plan) error {

	// Use the provider found in the plan
	providerPathFromClusterDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)

	// Create directory for keeping cluster state
	clusterStateDir := fmt.Sprintf("terraform/clusters/%s/", plan.Cluster.Name)

	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
		return fmt.Errorf("error creating dir to keep cluster state: %v", err)
	}
	if err := os.Chdir(clusterStateDir); err != nil {
		return fmt.Errorf("error switching dir to %s: %v", clusterStateDir, err)
	}
	defer os.Chdir("../../../")

	//Read tfvars from plan
	//and fill in the blanks
	aws.Populate(plan)

	//Write tfvars out
	file := "terraform.tfvars.json"
	aws.Write(file)

	tfInit := exec.Command(aws.BinaryPath, "init", providerPathFromClusterDir)
	if stdoutStderr, err := tfInit.CombinedOutput(); err != nil {
		return fmt.Errorf("Error initializing terraform: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner initialization successful.\n")

	tfPlan := exec.Command(aws.BinaryPath, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerPathFromClusterDir)

	if stdoutStderr, err := tfPlan.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform plan: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner planning successful.\n")

	fmt.Fprintf(out, "Provisioning...\n")
	tfApply := exec.Command(aws.Terraform.BinaryPath, "apply", plan.Cluster.Name)
	if stdoutStderr, err := tfApply.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform apply: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioning successful!\n")

	fmt.Fprintf(out, "Rendering plan file...\n")
	return aws.updatePlan(plan)
}

func (aws *AWS) updatePlan(plan *install.Plan) error {

	//For mutations eventually
	masterDiff := plan.Master.ExpectedCount - aws.MasterCount
	if masterDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	etcdDiff := plan.Etcd.ExpectedCount - aws.EtcdCount
	if etcdDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	workerDiff := plan.Worker.ExpectedCount - aws.WorkerCount
	if workerDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	storageDiff := plan.Storage.ExpectedCount - aws.StorageCount
	if storageDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}
	ingressDiff := plan.Ingress.ExpectedCount - aws.IngressCount
	if ingressDiff != 0 {
		return fmt.Errorf("We do not yet support cluster mutations")
	}

	//This is probably not as DRY as it could be, but I think it's the cleanest way to convert
	//the representations without doing any interface type assertion.
	//Node roles need to be refactored.

	//Masters
	tfNodes, err := aws.getTerraformNodes("master")
	if err != nil {
		return err
	}
	mng := &install.MasterNodeGroup{}

	if err := mng.Convert(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts); err != nil {
		return err
	}
	if len(mng.Nodes) == 1 {
		//Write these in in the case of a single master with no external LB
		mng.LoadBalancedFQDN = tfNodes.InternalIPs[0]
		mng.LoadBalancedShortName = tfNodes.IPs[0]
	} else {
		//Just grab the one already in the plan otherwise,
		//so Update doesn't overwrite it to a blank.
		mng.LoadBalancedFQDN = plan.Master.LoadBalancedFQDN
		mng.LoadBalancedShortName = plan.Master.LoadBalancedShortName

	}
	plan.Master.Update(mng)

	//Etcds
	tfNodes, err = aws.getTerraformNodes("etcd")
	if err != nil {
		return err
	}
	eng := &install.NodeGroup{}
	if err := eng.Convert(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts); err != nil {
		return err
	}
	plan.Etcd.Update(eng)

	//Workers
	tfNodes, err = aws.getTerraformNodes("worker")
	if err != nil {
		return err
	}
	wng := &install.NodeGroup{}
	if err = wng.Convert(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts); err != nil {
		return err
	}
	plan.Worker.Update(wng)

	//Ingress
	tfNodes, err = aws.getTerraformNodes("ingress")
	if err != nil {
		return err
	}
	ing := &install.NodeGroup{}
	if err := ing.Convert(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts); err != nil {
		return err
	}
	plan.Ingress.Update(ing)

	//Storage
	tfNodes, err = aws.getTerraformNodes("storage")
	if err != nil {
		return err
	}
	sng := &install.NodeGroup{}
	if err = sng.Convert(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts); err != nil {
		return err
	}
	plan.Storage.Update(sng)

	//SSH
	plan.Cluster.SSH.Key = aws.PrivateSSHKeyPath
	if aws.AMI == "" {
		plan.Cluster.SSH.User = "ubuntu"
	} else {
		plan.Cluster.SSH.User = "root"
	}

	//Write the updated plan
	clusterYaml := fmt.Sprintf("%s.yaml", aws.ClusterName)
	fp := install.FilePlanner{File: clusterYaml}
	if err := fp.Write(plan); err != nil {
		return fmt.Errorf("Error writing rendered file to file system")
	}
	return nil
}

//Populate fills in the fields for an AWS struct from the plan file and environment variables.
//This is also where the SSH keys are generated.
func (aws *AWS) Populate(plan *install.Plan) error {

	env := aws.getCommandEnvironment()

	//Terraform only cares about the data to the right
	aws.AccessKey = strings.Split(env[0], "=")[1]
	aws.SecretKey = strings.Split(env[1], "=")[1]

	if plan.Provisioner.AWSOptions.Region != "" {
		//Prefer the one defined in the plan
		aws.Region = plan.Provisioner.AWSOptions.Region
	} else if regionEnv := strings.Split(env[2], "=")[1]; regionEnv != "" {
		//Otherwise use the one defined in the environment variable
		aws.Region = regionEnv
	}
	//Allows for the case where the user already has a key they want to use
	if plan.Cluster.SSH.Key == "" {
		err := ssh.NewKeyPair("sshkey.pub", "sshkey.pem")
		if err != nil {
			return err
		}
		aws.PublicSSHKey, aws.PrivateSSHKeyPath, err = aws.getAWSKeyData("sshkey.pub", "sshkey.pem")
		if err != nil {
			return err
		}
	}
	aws.ClusterName = plan.Cluster.Name
	aws.AMI = plan.Provisioner.AWSOptions.AMI
	aws.EC2InstanceType = plan.Provisioner.AWSOptions.EC2InstanceType
	aws.MasterCount = len(plan.Master.Nodes)
	aws.EtcdCount = len(plan.Etcd.Nodes)
	aws.WorkerCount = len(plan.Worker.Nodes)
	aws.IngressCount = len(plan.Ingress.Nodes)
	aws.StorageCount = len(plan.Storage.Nodes)
	return nil
}

// Write the TFVars to the file system
func (aws *AWS) Write(file string) error {
	bytez, err := json.Marshal(aws.AWSTerraformData)
	if err != nil {
		return fmt.Errorf("error marshalling tfvars to json: %v", err)
	}
	err = ioutil.WriteFile(file, bytez, 0700)
	if err != nil {
		return fmt.Errorf("error making tfvars file: %v", err)
	}
	return nil
}

//Destroy destroys a provisioned cluster (using -force by default)
func (aws *AWS) Destroy(out io.Writer, clusterName string) error {
	clusterPathFromWd := clusterName
	os.Chdir(clusterPathFromWd)
	defer os.Chdir("../../../")
	defer os.RemoveAll(clusterPathFromWd)

	tfDestroy := exec.Command(aws.BinaryPath, "destroy", "-force")
	if stdoutStderr, err := tfDestroy.CombinedOutput(); err != nil {
		return fmt.Errorf("Error attempting to destroy: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Cluster destruction successful.\n")

	return nil
}
