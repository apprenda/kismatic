package integration

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
)

func testGlusterCluster(aws infrastructureProvisioner, distro linuxDistro) {
	WithInfrastructure(NodeCount{Worker: 2}, distro, aws, func(nodes provisionedNodes, sshKey string) {
		By("Setting up a plan file with storage nodes")
		plan := PlanAWS{
			Etcd:                     nodes.worker,
			Master:                   nodes.worker,
			Worker:                   nodes.worker,
			Storage:                  nodes.worker,
			MasterNodeFQDN:           nodes.worker[0].Hostname,
			MasterNodeShortName:      nodes.worker[0].Hostname,
			AllowPackageInstallation: true,
			SSHKeyFile:               sshKey,
			SSHUser:                  nodes.worker[0].SSHUser,
		}

		By("Writing plan file out to disk")
		template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
		FailIfError(err, "Couldn't parse template")
		f, err := os.Create("kismatic-testing.yaml")
		FailIfError(err, "Error waiting for nodes")
		defer f.Close()
		err = template.Execute(f, &plan)
		FailIfError(err, "Error filling in plan template")

		if distro == Ubuntu1604LTS { // Ubuntu doesn't have python installed
			By("Running the all play with the plan")
			cmd := exec.Command("./kismatic", "install", "step", "_all.yaml", "-f", f.Name())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			FailIfError(err, "Error running all play")
		}

		// The gluster play will attempt to create the endpoint using kubectl
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
		err = runViaSSH([]string{"sudo mv ~/kubectl /usr/bin/kubectl", "sudo chmod +x /usr/bin/kubectl"}, nodes.worker[0:1], sshKey, 1*time.Minute)

		By("Running the storage play with the plan")
		cmd := exec.Command("./kismatic", "install", "step", "_storage.yaml", "-f", f.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running storage play")

		By("Setting up a gluster volume with replication = 2")
		cmd = exec.Command("./kismatic", "volume", "add", "-f", f.Name(), "--replica-count", "2", "allow-address", nodes.worker[0].PrivateIP, "1", "gv1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume add command")

		By("Mounting the volume on one of the nodes, and writing a file")
		mount := fmt.Sprintf("sudo mount -t nfs %s:/gv1 /mnt1", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt1", mount, "sudo touch /mnt1/test-file1"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfError(err, "Error mounting gluster volume")

		time.Sleep(3 * time.Second)
		By("Verifying file is on both nodes")
		err = runViaSSH([]string{"sudo cat /data/gv1/test-file1"}, nodes.worker[0:2], sshKey, 30*time.Second)
		FailIfError(err, "Error verifying that the test file is in the gluster volume")

		By("Setting up a gluster volume with no replication")
		cmd = exec.Command("./kismatic", "volume", "add", "-f", f.Name(), "1", "--replica-count", "1", "gv2", "--allow-address", nodes.worker[0].PrivateIP)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume-add play")

		By("Mounting the volume on one of the nodes, and writing a file")
		mount = fmt.Sprintf("sudo mount -t nfs %s:/gv2 /mnt2", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt2", mount, "sudo touch /mnt2/test-file2"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfError(err, "Error mounting gluster volume")

		time.Sleep(3 * time.Second)
		By("Verifying file was not replicated")
		err = runViaSSH([]string{"sudo cat /data/gv2/test-file2"}, nodes.worker[0:2], sshKey, 30*time.Second) // expect the command to fail on one of the nodes
		FailIfSuccess(err, "Error verifying that the test file is only in one node")

		By("Creating a volume that allows access to worker[1]")
		cmd = exec.Command("./kismatic", "volume", "add", "-f", f.Name(), "--allow-address", nodes.worker[1].PrivateIP, "1", "gv3")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume-add play")

		By("Attempting to mount the volume on worker[0], which should not have access to the NFS share")
		mount = fmt.Sprintf("sudo mount -t nfs %s:/gv3 /mnt3", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt3", mount, "sudo touch /mnt3/test-file3"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfSuccess(err, "Expected mount error")
	})
}

func testVolumeAdd(masterNode NodeDeets, sshKey string) {
	By("Adding a volume using kismatic")
	volName := "kismatic-test-volume"
	cmd := exec.Command("./kismatic", "volume", "add", "-f", "kismatic-testing.yaml", "1", volName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	FailIfError(err, "Error creating a new volume")

	By("Verifying Kuberntes PV was created")
	err = runViaSSH([]string{"sudo kubectl get pv " + volName}, []NodeDeets{masterNode}, sshKey, 1*time.Minute)
	FailIfError(err, "Error verifying if PV gv0 was created")
}
