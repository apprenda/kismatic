package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Test a specific released version of Kismatic
var _ = Describe("Installing with previous version of Kismatic", func() {
	var path string
	BeforeEach(func() {
		// setup previous version of Kismatic
		kisReleasedPath, err := DownloadKismaticRelease(previousKismaticVersion)
		path = kisReleasedPath
		if err != nil {
			Fail("Failed to download kismatic released")
		}
		os.Chdir(kisReleasedPath)
	})
	AfterEach(func() {
		if !leaveIt() {
			os.RemoveAll(path)
		}
	})
	installOpts := installOptions{
		allowPackageInstallation: true,
	}
	Context("Targeting AWS infrastructure", func() {
		Context("using a 1/1/1 layout with Ubuntu 16.04 LTS", func() {
			ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 1}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("using a 1/1/1 layout with CentOS 7", func() {
			ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 1}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
