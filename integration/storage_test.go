package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Storage feature", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("Specifying multiple storage nodes in the plan file", func() {
		Context("when targetting CentOS", func() {
			ItOnAWS("should result in a working storage cluster [slow]", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, CentOS7)
			})
		})
		Context("when targetting Ubuntu", func() {
			ItOnAWS("should result in a working storage cluster [slow]", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, Ubuntu1604LTS)
			})
		})
		Context("when targetting RHEL", func() {
			ItOnAWS("should result in a working storage cluster [slow]", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, RedHat7)
			})
		})
	})

	Describe("Deploying a stateful workload", func() {
		Context("on a cluster with storage nodes", func() {
			ItOnAWS("should be able to read/write to a persistent volume [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Storage: 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					By("Installing a cluster with storage")
					opts := installOptions{
						allowPackageInstallation: true,
					}
					err := installKismatic(nodes, opts, sshKey)
					FailIfError(err, "Installation failed")

					err = testStatefulWorkload(nodes, sshKey)
					FailIfError(err, "Stateful workload test failed")
				})
			})
		})
	})
})
