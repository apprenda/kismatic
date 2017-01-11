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

		By("Setting up a gluster volume on the nodes")
		// TODO replace with acutal CLI command
		ev := fmt.Sprintf("volume_mount=/,volume_replica_count=2,volume_name=gv1,volume_quota=1,volume_quota_raw=1073741824,allow_ips=%s", nodes.worker[0].PrivateIP)
		cmd = exec.Command("./kismatic", "install", "step", "_volume-add.yaml", "-f", f.Name(), "--extra-vars", ev)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume-add play")

		By("Mounting the volume on one of the nodes, and writing a file")
		mount := fmt.Sprintf("sudo mount -t nfs %s:/gv1 /mnt1", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt1", mount, "sudo touch /mnt1/test-file1"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfError(err, "Error mounting gluster volume")

		time.Sleep(3 * time.Second)
		By("Verifying file is on both nodes")
		err = runViaSSH([]string{"sudo cat /data/gv1/test-file1"}, nodes.worker[0:2], sshKey, 30*time.Second)
		FailIfError(err, "Error verifying that the test file is in the gluster volume")

		By("Setting up a gluster volume on one node")
		// TODO replace with acutal CLI command
		ev = fmt.Sprintf("volume_mount=/,volume_replica_count=1,volume_name=gv2,volume_quota=1,volume_quota_raw=1073741824,allow_ips=%s", nodes.worker[0].PrivateIP)
		cmd = exec.Command("./kismatic", "install", "step", "_volume-add.yaml", "-f", f.Name(), "--extra-vars", ev)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume-add play")

		By("Mounting the volume on one of the nodes, and writing a file")
		mount = fmt.Sprintf("sudo mount -t nfs %s:/gv2 /mnt2", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt2", mount, "sudo touch /mnt2/test-file2"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfError(err, "Error mounting gluster volume")

		time.Sleep(3 * time.Second)
		By("Verifying file is on the one node")
		err = runViaSSH([]string{"sudo cat /data/gv2/test-file2"}, nodes.worker[0:2], sshKey, 30*time.Second)
		// file shuld not exist and should error
		FailIfSuccess(err, "Error verifying that the test file is only in one volume")

		By("Trying to mount volume with bad permissions")
		// TODO replace with acutal CLI command
		ev = fmt.Sprintf("volume_mount=/,volume_replica_count=2,volume_name=gv3,volume_quota=1,volume_quota_raw=1073741824,allow_ips=%s", nodes.worker[1].PrivateIP)
		cmd = exec.Command("./kismatic", "install", "step", "_volume-add.yaml", "-f", f.Name(), "--extra-vars", ev)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error running volume-add play")

		By("Mounting the volume on one of the nodes, and writing a file")
		mount = fmt.Sprintf("sudo mount -t nfs %s:/gv3 /mnt3", nodes.worker[0].Hostname)
		err = runViaSSH([]string{"sudo mkdir /mnt3", mount, "sudo touch /mnt2/test-file2"}, nodes.worker[0:1], sshKey, 30*time.Second)
		FailIfSuccess(err, "Expected mount error")
	})
}
