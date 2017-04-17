package integration

import (
	"fmt"
	"strings"
	"time"
)

func verifyRBAC(master NodeDeets, sshKey string) error {
	// copy rbac policy over to master node
	err := copyFileToRemote("test-resources/rbac/pod-reader.yaml", "/tmp/pod-reader.yaml", master, sshKey, 10*time.Second)
	if err != nil {
		return err
	}
	// we need the CA key to generate a new user cert
	err = copyFileToRemote("generated/keys/ca-key.pem", "/tmp/ca-key.pem", master, sshKey, 10*time.Second)
	if err != nil {
		return err
	}
	// create using kubectl
	commands := []string{
		// create the RBAC policy
		"sudo kubectl create -f /tmp/pod-reader.yaml",
		// generate a private key for jane
		"sudo openssl genrsa -out /tmp/jane-key.pem 2048",
		// generate a CSR for jane
		"sudo openssl req -new -key /tmp/jane-key.pem -out /tmp/jane-csr.pem -subj \"/CN=jane/O=some-group\"",
		// generate certificate for jane
		"sudo openssl x509 -req -in /tmp/jane-csr.pem -CA /etc/kubernetes/ca.pem -CAkey /tmp/ca-key.pem -CAcreateserial -out /tmp/jane.pem -days 10",
		// configure new user in kubeconfig
		"sudo kubectl config set-credentials jane --client-certificate=/tmp/jane.pem --client-key=/tmp/jane-key.pem",
	}
	err = runViaSSH(commands, []NodeDeets{master}, sshKey, 30*time.Second)
	if err != nil {
		return err
	}
	// Using kubectl to get pods should succeed
	err = runViaSSH([]string{"sudo kubectl get pods --user=jane"}, []NodeDeets{master}, sshKey, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get pods as jane: %v", err)
	}
	// This command is expected to fail, so we ignore the error.
	// We expect the output to contain the string "Forbidden"
	out, _ := executeCmd("sudo kubectl get nodes --user=jane", master.PublicIP, master.SSHUser, sshKey)
	if !strings.Contains(out, "Forbidden") {
		return fmt.Errorf("expected a forbidden response from the server, but output did not indicate this. Output was: %s", out)
	}
	return nil
}
