package integration

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
)

func testAddVolumeVerifyGluster(aws infrastructureProvisioner, distro linuxDistro) {
	WithInfrastructure(NodeCount{Worker: 5}, distro, aws, func(nodes provisionedNodes, sshKey string) {
		planFile, err := os.Create("kismatic-testing.yaml")
		FailIfError(err, "error creating file for kismatic plan")
		defer planFile.Close()

		clusterNodes := nodes.worker[0:4]
		standupGlusterCluster(planFile, clusterNodes, sshKey, distro)
		storageNode := nodes.worker[0]

		tests := []struct {
			replicaCount      int
			distributionCount int
		}{
			{
				replicaCount:      1,
				distributionCount: 1,
			},
			{
				replicaCount:      2,
				distributionCount: 1,
			},
			{
				replicaCount:      1,
				distributionCount: 2,
			},
			{
				replicaCount:      2,
				distributionCount: 2,
			},
		}

		for _, test := range tests {
			By(fmt.Sprintf("Setting up a volume with Replica = %d, Distributed = %d", test.replicaCount, test.distributionCount))
			volumeName := fmt.Sprintf("gv-r%d-d%d", test.replicaCount, test.distributionCount)
			createVolume(planFile, volumeName, test.replicaCount, test.distributionCount, "")

			By("Verifying gluster volume properties")
			verifyGlusterVolume(storageNode, sshKey, volumeName, test.replicaCount, test.distributionCount, "")
		}

		By("Creating a volume which allows access to nodes in the cluster")
		createVolume(planFile, "foo", 1, 1, "")

		By("Installing NFS library on out-of-cluster node")
		unauthNode := nodes.worker[4:5]
		var cmd string
		switch distro {
		case Ubuntu1604LTS:
			cmd = "sudo apt-get update -y && sudo apt-get install -y nfs-common"
		case CentOS7, RedHat7:
			cmd = "sudo yum install -y nfs-utils"
		}
		err = runViaSSH([]string{cmd}, unauthNode, sshKey, 2*time.Minute)
		FailIfError(err, "Failed to install nfs-common on Ubuntu")
		By("Attempting to mount the volume a node that is not part of the cluster, which should not have access to the NFS share")
		mount := fmt.Sprintf("sudo mount -t nfs %s:/foo /mnt3", clusterNodes[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt3", mount, "sudo touch /mnt3/test-file3"}, unauthNode, sshKey, 30*time.Second)
		FailIfSuccess(err, "Expected mount error")
	})
}
func verifyGlusterVolume(storageNode NodeDeets, sshKey string, name string, replicationCount int, distributionCount int, allowedIpList string) {
	// verify allowed IP List
	commands := []string{}
	if allowedIpList != "" {
		commands = append(commands, fmt.Sprintf(`sudo gluster volume info %s | grep "nfs.rpc-auth-allow: %s"`, name, allowedIpList))
	}
	// verify replication and distribution
	if replicationCount > 1 {
		cmd := fmt.Sprintf(`sudo gluster volume info %s | grep "Number of Bricks: %d x %d"`, name, distributionCount, replicationCount)
		commands = append(commands, cmd)
	} else {
		cmd := fmt.Sprintf(`sudo gluster volume info %s | grep "Number of Bricks: %d"`, name, distributionCount)
		commands = append(commands, cmd)
	}
	err := runViaSSH(commands, []NodeDeets{storageNode}, sshKey, 1*time.Minute)
	if err != nil {
		// get volume details to print in the console
		runViaSSH([]string{"sudo gluster volume info " + name}, []NodeDeets{storageNode}, sshKey, 1*time.Minute)
	}
	FailIfError(err, "Gluster volume verification failed")
}

func createVolume(planFile *os.File, name string, replicationCount int, distributionCount int, allowAddress string) {
	cmd := exec.Command("./kismatic", "volume", "add",
		"-f", planFile.Name(),
		"--replica-count", strconv.Itoa(replicationCount),
		"--distribution-count", strconv.Itoa(distributionCount),
		"1", name)
	if allowAddress != "" {
		cmd.Args = append(cmd.Args, "--allow-address", allowAddress)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	FailIfError(err, "Error running volume add command")
}

func standupGlusterCluster(planFile *os.File, nodes []NodeDeets, sshKey string, distro linuxDistro) {
	By("Setting up a plan file with storage nodes")
	plan := PlanAWS{
		Etcd:                     nodes,
		Master:                   nodes,
		Worker:                   nodes,
		Storage:                  nodes,
		MasterNodeFQDN:           nodes[0].Hostname,
		MasterNodeShortName:      nodes[0].Hostname,
		AllowPackageInstallation: true,
		SSHKeyFile:               sshKey,
		SSHUser:                  nodes[0].SSHUser,
	}
	By("Writing plan file out to disk")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	err = template.Execute(planFile, &plan)
	FailIfError(err, "Error filling in plan template")
	if distro == Ubuntu1604LTS { // Ubuntu doesn't have python installed
		By("Running the all play with the plan")
		cmd := exec.Command("./kismatic", "install", "step", "_all.yaml", "-f", planFile.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running all play")
	}
	By("Mocking kubectl on the first master node")
	kubectlDummy := `#!/bin/bash
		# This is a dummy generated for a Kismatic integration test
		exit 0
		`
	kubectlDummyFile, err := ioutil.TempFile("", "kubectl-dummy")
	FailIfError(err, "Error creating temp file")
	err = ioutil.WriteFile(kubectlDummyFile.Name(), []byte(kubectlDummy), 0644)
	FailIfError(err, "Error writing kubectl dummy file")
	err = copyFileToRemote(kubectlDummyFile.Name(), "~/kubectl", plan.Master[0], sshKey, 1*time.Minute)
	FailIfError(err, "Error copying kubectl dummy")
	err = runViaSSH([]string{"sudo mv ~/kubectl /usr/bin/kubectl", "sudo chmod +x /usr/bin/kubectl"}, nodes[0:1], sshKey, 1*time.Minute)

	By("Running the storage play with the plan")
	cmd := exec.Command("./kismatic", "install", "step", "_storage.yaml", "-f", planFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	FailIfError(err, "Error running storage play")
}

func testVolumeAdd(masterNode NodeDeets, sshKey string) {
	By("Adding a volume using kismatic")
	volName := "kismatic-test-volume"
	cmd := exec.Command("./kismatic", "volume", "add", "-f", "kismatic-testing.yaml", "--replica-count", "1", "1", volName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	FailIfError(err, "Error creating a new volume")

	By("Verifying Kuberntes PV was created")
	err = runViaSSH([]string{"sudo kubectl get pv " + volName}, []NodeDeets{masterNode}, sshKey, 1*time.Minute)
	FailIfError(err, "Error verifying if PV gv0 was created")
}
