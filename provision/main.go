package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/apprenda/kismatic-platform/integration"
)

func main() {
	if err := os.Setenv("LEAVE_ARTIFACTS", "true"); err != nil {
		log.Fatal("Error setting environment variable", err)
	}
	os.Setenv("BAIL_BEFORE_ANSIBLE", "true")

	tmpDir, err := ioutil.TempDir("", "kisint")
	if err != nil {
		log.Fatal("error getting temp dir", err)
	}

	c := exec.Command("tar", "-zxf", "out/kismatic.tar.gz", "-C", tmpDir)
	out, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatal("Error unpacking installer", err)
	}
	os.Chdir(tmpDir)

	aws, ok := integration.AWSClientFromEnvironment()
	if !ok {
		log.Fatal("Required AWS environment variables not defined")
	}

	nodeCount := integration.NodeCount{Etcd: 1, Master: 1, Worker: 1}
	distro := integration.Ubuntu1604LTS
	nodes, err := aws.ProvisionNodes(nodeCount, distro)
	if err != nil {
		log.Fatal("Error provisioning nodes", err)
	}

	err = integration.WaitForSSH(nodes, aws.SSHKey())
	if err != nil {
		log.Fatal("Error waiting for SSH", err)
	}

	installOpts := integration.InstallOptions{
		AllowPackageInstallation: true,
	}

	err = integration.InstallKismatic(nodes, installOpts, aws.SSHKey())
	if err != nil {
		log.Fatalf("Error installing kismatic: %v", err)
	}
	fmt.Println("Your cluster is ready.")
	fmt.Println(nodes)

}
