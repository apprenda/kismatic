package provision

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
)

const terraform string = "./../../bin/terraform"

//Provision provides a wrapper for terraform init, terraform plan, and terraform apply.
func Provision(out io.Writer, plan *install.Plan) error {

	clusterPathFromWd := fmt.Sprintf("terraform/clusters/%s/", plan.Cluster.Name)
	providerPathFromClusterDir := fmt.Sprintf("../../providers/%s", plan.Provisioner.Provider)
	os.Chdir(clusterPathFromWd)
	tfInit := exec.Command(terraform, "init", providerPathFromClusterDir)
	if stdoutStderr, err := tfInit.CombinedOutput(); err != nil {
		return fmt.Errorf("Error initializing terraform: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner initialization successful.\n")

	tfPlan := exec.Command(terraform, "plan", fmt.Sprintf("-out=%s", plan.Cluster.Name), providerPathFromClusterDir)
	// TODO: make -out=plan.Name
	// TODO: make target=cluster and symlink it to the provider
	if stdoutStderr, err := tfPlan.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform plan: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioner planning successful.\n")

	fmt.Fprintf(out, "Provisioning...\n")

	tfApply := exec.Command(terraform, "apply", plan.Cluster.Name)
	if stdoutStderr, err := tfApply.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running terraform apply: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Provisioning successful!\n")
	fmt.Fprintf(out, "Rendering plan file...\n")

	// Render with KET
	// tfOutput := exec.Command(terraform, "output", "rendered_template")
	// stdoutStderr, err := tfOutput.CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("Error collecting terraform output: %s", stdoutStderr)
	// }

	// if err := ioutil.WriteFile("kismatic-cluster.yaml", stdoutStderr, 0644); err != nil {
	// 	return fmt.Errorf("Error writing rendered file to file system")
	// }
	// fmt.Fprintf(out, "Plan file %s rendered.\n", "kismatic-cluster.yaml")
	os.Chdir("../../../")
	return nil
}
