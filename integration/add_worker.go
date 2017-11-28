package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
)

func addWorkerToCluster(newWorker NodeDeets, sshKey string, labels []string) error {
	By("Adding new worker")
	cmd := exec.Command("./kismatic", "install", "add-worker", "-f", "kismatic-testing.yaml", newWorker.Hostname, newWorker.PublicIP, newWorker.PrivateIP)
	if len(labels) > 0 {
		cmd.Args = append(cmd.Args, "--labels", strings.Join(labels, ","))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running add worker command: %v", err)
	}

	By("Verifying that the worker was added")
	sshCmd := fmt.Sprintf("sudo kubectl get nodes %s --kubeconfig /root/.kube/config", strings.ToLower(newWorker.Hostname)) // the api server is case-sensitive.
	out, err := executeCmd(sshCmd, newWorker.PublicIP, newWorker.SSHUser, sshKey)
	if err != nil {
		return fmt.Errorf("error getting nodes using kubectl: %v. Command output was: %s", err, out)
	}

	By("Verifying that the worker is in the ready state")
	if !strings.Contains(strings.ToLower(out), "ready") {
		return fmt.Errorf("the node was not in ready state. node details: %s", out)
	}
	return nil
}
