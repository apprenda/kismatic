package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
)

func addNodeToCluster(newNode NodeDeets, sshKey string, labels []string, roles []string) error {
	By("Adding new worker")
	cmd := exec.Command("./kismatic", "install", "add-node", "-f", "kismatic-testing.yaml", "--roles", strings.Join(roles, ","), newNode.Hostname, newNode.PublicIP, newNode.PrivateIP)
	if len(labels) > 0 {
		cmd.Args = append(cmd.Args, "--labels", strings.Join(labels, ","))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running add node command: %v", err)
	}

	By("Verifying that the node was added")
	sshCmd := fmt.Sprintf("sudo kubectl get nodes %s --kubeconfig /root/.kube/config", strings.ToLower(newNode.Hostname)) // the api server is case-sensitive.
	out, err := executeCmd(sshCmd, newNode.PublicIP, newNode.SSHUser, sshKey)
	if err != nil {
		return fmt.Errorf("error getting nodes using kubectl: %v. Command output was: %s", err, out)
	}

	By("Verifying that the node is in the ready state")
	if !strings.Contains(strings.ToLower(out), "ready") {
		return fmt.Errorf("the node was not in ready state. node details: %s", out)
	}
	return nil
}
