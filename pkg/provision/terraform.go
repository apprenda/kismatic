package provision

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/ssh"
)

// The AnyTerraform provisioner uses Terraform to provision infrastructure using
// providers that adhere to the KET provisioner spec.
type AnyTerraform struct {
	KismaticVersion string
	ClusterOwner    string
	ProvidersDir    string
	StateDir        string
	BinaryPath      string
	Output          io.Writer
	SecretsGetter   SecretsGetter
}

type outputVariableReader struct {
	clusterName  string
	stateDir     string
	tfBinaryPath string
}

type tfCommand struct {
	binaryPath string
	output     io.Writer
	env        []string
	workDir    string
}

// An aggregate of different tfNodes (different fields, the same nodes)
// NOTE: these are organized a little differently than a traditional node group
// due to limitations of terraform. A tfNodeGroup organizes each field into
// parallel slices as opposed to a single slice with nodes containing the same
// data.
type tfNodeGroup struct {
	IPs         []string
	InternalIPs []string
	Hosts       []string
}

type tfOutputVar struct {
	Sensitive  bool     `json:"sensitive"`
	OutputType string   `json:"type"`
	Value      []string `json:"value"`
}

// Provision creates the infrastructure required to support the cluster defined
// in the plan
func (at AnyTerraform) Provision(plan install.Plan, opts ProvisionOpts) (*install.Plan, error) {
	providerName := plan.Provisioner.Provider
	providerDir := filepath.Join(at.ProvidersDir, providerName)
	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("provider %q is not supported", providerName)
	}

	p, err := readProviderDescriptor(providerDir)
	if err != nil {
		return nil, err
	}

	// Create directory for keeping cluster state
	clusterStateDir := filepath.Join(at.StateDir, plan.Cluster.Name)
	if err := os.MkdirAll(clusterStateDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory to keep cluster state: %v", err)
	}

	pubKeyPath := filepath.Join(clusterStateDir, fmt.Sprintf("%s-ssh.pub", plan.Cluster.Name))
	privKeyPath := filepath.Join(clusterStateDir, fmt.Sprintf("%s-ssh.pem", plan.Cluster.Name))

	var privKeyExists, pubKeyExists bool
	if _, err := os.Stat(pubKeyPath); err == nil {
		pubKeyExists = true
	}
	if _, err := os.Stat(privKeyPath); err == nil {
		privKeyExists = true
	}

	if pubKeyExists != privKeyExists {
		if !privKeyExists {
			return nil, fmt.Errorf("found an existing public key at %s, but did not find the corresponding private key at %s. The corresponding key must be recovered if possible. Otherwise, the existing key must be deleted", pubKeyPath, privKeyPath)
		}
		return nil, fmt.Errorf("found an existing private key at %s, but did not find the corresponding public key at %s. The corresponding key must be recovered if possible. Otherwise, the existing key must be deleted", privKeyPath, pubKeyPath)
	}

	if !privKeyExists && !pubKeyExists {
		if err := ssh.NewKeyPair(pubKeyPath, privKeyPath); err != nil {
			return nil, fmt.Errorf("error generating SSH key pair: %v", err)
		}
	}
	plan.Cluster.SSH.Key = privKeyPath

	// Write out the KET terraform variables
	data := ketVars{
		KismaticVersion:   at.KismaticVersion,
		ClusterOwner:      at.ClusterOwner,
		ClusterName:       plan.Cluster.Name,
		MasterCount:       plan.Master.ExpectedCount,
		EtcdCount:         plan.Etcd.ExpectedCount,
		WorkerCount:       plan.Worker.ExpectedCount,
		IngressCount:      plan.Ingress.ExpectedCount,
		StorageCount:      plan.Storage.ExpectedCount,
		PrivateSSHKeyPath: privKeyPath,
		PublicSSHKeyPath:  pubKeyPath,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filepath.Join(clusterStateDir, "terraform.tfvars"), b, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing terraform variables: %v", err)
	}

	// Write out the provisioner options as terraform variables
	b, err = json.MarshalIndent(plan.Provisioner.Options, "", "  ")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filepath.Join(clusterStateDir, "provider.auto.tfvars"), b, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing tfvars file for provider-specific options")
	}

	// Setup the environment for all Terraform commands.
	secretEnvVars, err := at.SecretsGetter.GetAsEnvironmentVariables(plan.Cluster.Name, p.EnvironmentVariables)
	if err != nil {
		return nil, fmt.Errorf("could not get secrets required for provisioning infrastructure: %v", err)
	}

	tf := tfCommand{
		binaryPath: at.BinaryPath,
		output:     at.Output,
		env:        secretEnvVars,
		workDir:    clusterStateDir,
	}

	// Terraform init
	if err := tf.run("init", providerDir); err != nil {
		return nil, fmt.Errorf("Error initializing terraform: %s", err)
	}

	// Terraform plan - checks for destruction
	var capture bytes.Buffer
	old := tf.output
	w := io.MultiWriter(&capture, tf.output)
	tf.output = w
	if err := tf.run("plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerDir); err != nil {
		return nil, fmt.Errorf("Error running terraform plan: %s", err)
	}
	if !strings.Contains(capture.String(), "0 to destroy") && !opts.AllowDestruction {
		return nil, fmt.Errorf("Destruction of resources detected when not issuing a destroy. If this is intended, please re-run the command with --allow-destruction.")
	}
	tf.output = old

	// Terraform apply
	if err := tf.run("apply", "-input=false", plan.Cluster.Name); err != nil {
		return nil, fmt.Errorf("Error running terraform apply: %s", err)
	}

	// Update plan with data from provider
	provisionedPlan, err := at.buildPopulatedPlan(plan)
	if err != nil {
		return nil, err
	}
	return provisionedPlan, nil
}

// Destroy tears down the cluster and infrastructure defined in the plan file
func (at AnyTerraform) Destroy(provider, clusterName string) error {
	providerDir := filepath.Join(at.ProvidersDir, provider)
	p, err := readProviderDescriptor(providerDir)
	if err != nil {
		return err
	}
	secretEnvVars, err := at.SecretsGetter.GetAsEnvironmentVariables(clusterName, p.EnvironmentVariables)
	if err != nil {
		return err
	}
	tf := tfCommand{
		binaryPath: at.BinaryPath,
		output:     at.Output,
		env:        secretEnvVars,
		workDir:    filepath.Join(at.StateDir, clusterName),
	}
	if err := tf.run("destroy", "-force"); err != nil {
		return errors.New("Error destroying infrastructure with Terraform")
	}
	return nil
}

func (at AnyTerraform) getLoadBalancer(clusterName, lbName string) (string, error) {
	ovr := outputVariableReader{
		clusterName:  clusterName,
		stateDir:     at.StateDir,
		tfBinaryPath: at.BinaryPath,
	}
	varName := fmt.Sprintf("%s_lb", lbName)
	values, err := ovr.readStringSlice(varName)
	if err != nil {
		return "", err
	}
	if len(values) != 1 {
		return "", fmt.Errorf("expected to get a single value for output variable %q, but got %d", varName, len(values))
	}
	return values[0], nil
}

func (at AnyTerraform) getTerraformNodes(clusterName, role string) (*tfNodeGroup, error) {
	ovr := outputVariableReader{
		clusterName:  clusterName,
		stateDir:     at.StateDir,
		tfBinaryPath: at.BinaryPath,
	}
	publicIPs, err := ovr.readStringSlice(fmt.Sprintf("%s_pub_ips", role))
	if err != nil {
		return nil, err
	}
	privateIPs, err := ovr.readStringSlice(fmt.Sprintf("%s_priv_ips", role))
	if err != nil {
		return nil, err
	}
	hostnames, err := ovr.readStringSlice(fmt.Sprintf("%s_hosts", role))
	if err != nil {
		return nil, err
	}

	if len(publicIPs) != len(hostnames) {
		return nil, fmt.Errorf("The number of public IPs (%d) does not match the number of hostnames (%d)", len(publicIPs), len(hostnames))
	}

	// Verify that we got the right number of internal IPs if we are using them
	if len(privateIPs) != 0 && len(publicIPs) != len(privateIPs) {
		return nil, fmt.Errorf("The number of public IPs (%d) does not match the number of private IPs (%d)", len(publicIPs), len(privateIPs))
	}

	return &tfNodeGroup{
		IPs:         publicIPs,
		InternalIPs: privateIPs,
		Hosts:       hostnames,
	}, nil
}

func (at AnyTerraform) buildPopulatedPlan(plan install.Plan) (*install.Plan, error) {
	// Masters
	tfNodes, err := at.getTerraformNodes(plan.Cluster.Name, "master")
	if err != nil {
		return nil, err
	}
	masterNodes := install.Merge(install.NodeGroup{Nodes: plan.Master.Nodes, ExpectedCount: plan.Master.ExpectedCount},
		nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))
	mng := install.MasterNodeGroup{
		ExpectedCount: masterNodes.ExpectedCount,
		Nodes:         masterNodes.Nodes,
	}
	mlb, err := at.getLoadBalancer(plan.Cluster.Name, "master")
	if err != nil {
		return nil, err
	}
	mng.LoadBalancedFQDN = mlb
	mng.LoadBalancedShortName = mlb
	plan.Master = mng

	// Etcds
	tfNodes, err = at.getTerraformNodes(plan.Cluster.Name, "etcd")
	if err != nil {
		return nil, err
	}
	plan.Etcd = install.Merge(plan.Etcd, nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))

	// Workers
	tfNodes, err = at.getTerraformNodes(plan.Cluster.Name, "worker")
	if err != nil {
		return nil, err
	}
	plan.Worker = install.Merge(plan.Worker, nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))

	// Ingress
	if plan.Ingress.ExpectedCount > 0 {
		tfNodes, err = at.getTerraformNodes(plan.Cluster.Name, "ingress")
		if err != nil {
			return nil, fmt.Errorf("error getting ingress node information: %v", err)
		}
		plan.Ingress = install.OptionalNodeGroup(install.Merge(install.NodeGroup{Nodes: plan.Ingress.Nodes, ExpectedCount: plan.Ingress.ExpectedCount},
			nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts)))
	}

	// Storage
	if plan.Storage.ExpectedCount > 0 {
		tfNodes, err = at.getTerraformNodes(plan.Cluster.Name, "storage")
		if err != nil {
			return nil, fmt.Errorf("error getting storage node information: %v", err)
		}
		plan.Storage = install.OptionalNodeGroup(nodeGroupFromSlices(tfNodes.IPs, tfNodes.InternalIPs, tfNodes.Hosts))
	}

	return &plan, nil
}

func (tfc tfCommand) run(args ...string) error {
	cmd := exec.Command(tfc.binaryPath, args...)
	cmd.Stdout = tfc.output
	cmd.Stderr = tfc.output
	cmd.Dir = tfc.workDir
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=True")
	cmd.Env = append(cmd.Env, tfc.env...)
	return cmd.Run()
}

func (ovr outputVariableReader) read(varName string) (*tfOutputVar, error) {
	cmd := exec.Command(ovr.tfBinaryPath, "output", "-json", varName)
	cmd.Dir = filepath.Join(ovr.stateDir, ovr.clusterName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting output variable %q: %s", varName, out)
	}
	var ov tfOutputVar
	if err := json.Unmarshal(out, &ov); err != nil {
		return nil, fmt.Errorf("error unmarshaling output variable %q: %v", out, err)
	}
	return &ov, nil
}

func (ovr outputVariableReader) readStringSlice(varName string) ([]string, error) {
	v, err := ovr.read(varName)
	if err != nil {
		return nil, err
	}
	return v.Value, nil
}
