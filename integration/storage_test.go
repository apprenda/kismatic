package integration

import (
	"os"

	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Storage feature", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})

	Describe("Specifying multiple storage nodes in the plan file", func() {
		Context("when targetting CentOS", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, CentOS7)
			})
		})
		Context("when targetting Ubuntu", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, Ubuntu1604LTS)
			})
		})
		Context("when targetting RHEL", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, RedHat7)
			})
		})
	})

	Describe("Deploying a stateful workload", func() {
		Context("on a cluster with storage nodes", func() {
			ItOnAWS("should be able to read/write to a persistent volume [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Storage: 2}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
					By("Installing a cluster with storage")
					opts := installOptions{
						allowPackageInstallation: true,
					}
					err := installKismatic(nodes, opts, sshKey)
					FailIfError(err, "Installation failed")

					// Helper for deploying on K8s
					kubeCreate := func(resource string) error {
						err := copyFileToRemote("test-resources/storage/"+resource, "/tmp/"+resource, nodes.master[0], sshKey, 30*time.Second)
						if err != nil {
							return err
						}
						return runViaSSH([]string{"sudo kubectl create -f /tmp/" + resource}, []NodeDeets{nodes.master[0]}, sshKey, 30*time.Second)
					}

					By("Creating a storage volume")
					plan, err := os.Open("kismatic-testing.yaml")
					FailIfError(err, "Failed to open plan file")
					createVolume(plan, "kis-int-test", 2, 1, "")

					By("Claiming the storage volume on the cluster")
					err = kubeCreate("pvc.yaml")
					FailIfError(err, "Failed to create pvc")

					By("Deploying a writer workload")
					err = kubeCreate("writer.yaml")
					FailIfError(err, "Failed to create writer workload")

					By("Verifying the completion of the write workload")
					time.Sleep(1 * time.Minute)
					jobStatusCmd := "sudo kubectl get jobs kismatic-writer -o jsonpath={.status.conditions[0].status}"
					err = runViaSSH([]string{jobStatusCmd, fmt.Sprintf("if [ \"`%s`\" = \"True\" ]; then exit 0; else exit 1; fi", jobStatusCmd)}, []NodeDeets{nodes.master[0]}, sshKey, 30*time.Second)
					FailIfError(err, "Writer workload failed")

					By("Deploying a reader workload")
					err = kubeCreate("reader.yaml")
					FailIfError(err, "Failed to create reader workload")

					By("Verifying the completion of the reader workload")
					time.Sleep(1 * time.Minute)
					jobStatusCmd = "sudo kubectl get jobs kismatic-reader -o jsonpath={.status.conditions[0].status}"
					runViaSSH([]string{jobStatusCmd, fmt.Sprintf("if [ \"`%s`\" = \"True\" ]; then exit 0; else exit 1; fi", jobStatusCmd)}, []NodeDeets{nodes.master[0]}, sshKey, 30*time.Second)
					FailIfError(err, "Reader workload failed")
				})
			})
		})
	})
})
