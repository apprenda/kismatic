package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Test a specific released version of Kismatic
var _ = Describe("Installing with previous version of Kismatic", func() {
	BeforeEach(func() {
		os.Chdir(kisReleasedPath)
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
