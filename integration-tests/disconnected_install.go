package integration_tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
)

// creates a package repository mirror on the given node.
func createPackageRepositoryMirror(repoNode NodeDeets, distro linuxDistro, sshKey string) error {
	var mirrorScript string
	switch distro {
	case CentOS7:
		mirrorScript = "mirror-rpms.sh"
	case Ubuntu1604LTS:
		mirrorScript = "mirror-debs.sh"
	default:
		return fmt.Errorf("unable to create repo mirror for distro %q", distro)
	}
	start := time.Now()
	err := copyFileToRemote("test-resources/disconnected-installation/"+mirrorScript, "/tmp/"+mirrorScript, repoNode, sshKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to copy script to remote node: %v", err)
	}
	cmds := []string{"chmod +x /tmp/" + mirrorScript, "sudo /tmp/" + mirrorScript}
	err = runViaSSH(cmds, []NodeDeets{repoNode}, sshKey, 120*time.Minute)
	if err != nil {
		return fmt.Errorf("error running mirroring script: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Println("Creating a package repository took", elapsed)
	return nil
}

// seeds a container image registry using the kismatic seed-registry command
func seedRegistry(repoNode NodeDeets, registryCAFile string, registryPort int, sshKey string) error {
	By("Adding the docker registry self-signed cert to the registry node")
	registry := fmt.Sprintf("%s:%d", repoNode.PublicIP, registryPort)
	err := copyFileToRemote(registryCAFile, "/tmp/docker-registry-ca.crt", repoNode, sshKey, 30*time.Second)
	if err != nil {
		return fmt.Errorf("Failed to copy registry cert to registry node: %v", err)
	}
	cmds := []string{
		fmt.Sprintf("sudo mkdir -p /etc/docker/certs.d/%s", registry),
		fmt.Sprintf("sudo mv /tmp/docker-registry-ca.crt /etc/docker/certs.d/%s/ca.crt", registry),
	}
	err = runViaSSH(cmds, []NodeDeets{repoNode}, sshKey, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("Error adding self-signed cert to registry node: %v", err)
	}

	By("Copying KET to the registry node for seeding")
	err = copyFileToRemote(filepath.Join(currentKismaticDir, "kismatic-"+runtime.GOOS+".tar.gz"), "/tmp/kismatic-"+runtime.GOOS+".tar.gz", repoNode, sshKey, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("Error copying KET to the registry node: %v", err)
	}

	By("Seeding the registry")
	start := time.Now()
	cmds = []string{
		fmt.Sprintf("sudo docker login -u kismaticuser -p kismaticpassword %s", registry),
		"sudo mkdir kismatic",
		"sudo tar -xf /tmp/kismatic-" + runtime.GOOS + ".tar.gz -C kismatic",
		fmt.Sprintf("sudo ./kismatic/kismatic seed-registry --server %s", registry),
	}
	err = runViaSSH(cmds, []NodeDeets{repoNode}, sshKey, 60*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to seed the registry: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Println("Seeding the registry took", elapsed)
	return nil
}
