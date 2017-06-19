package integration

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
)

func addWorkerToCluster(newWorker NodeDeets) error {
	By("Adding new worker")
	cmd := exec.Command("./kismatic", "install", "add-worker", "-f", "kismatic-testing.yaml", newWorker.Hostname, newWorker.PublicIP, newWorker.PrivateIP)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
