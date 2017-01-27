package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const previousKismaticVersion = "v1.2.0"

// Test a specific released version of Kismatic
var _ = Describe("Installing with previous version of Kismatic", func() {
	var kisReleasedPath string
	BeforeEach(func() {
		// setup previous version of Kismatic
		var err error
		kisReleasedPath, err = DownloadKismaticRelease(previousKismaticVersion)
		Expect(err).ToNot(HaveOccurred(), "Failed to download kismatic release")
		os.Chdir(kisReleasedPath)
	})

	AfterEach(func() {
		if !leaveIt() {
			os.RemoveAll(kisReleasedPath)
		}
	})

	installOpts := installOptions{
		allowPackageInstallation: true,
	}

	Context("using Ubuntu 16.04 LTS", func() {
		ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{1, 1, 1, 0, 0}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
				err := installKismatic(nodes, installOpts, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("using CentOS", func() {
		ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{1, 1, 1, 0, 0}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
				err := installKismatic(nodes, installOpts, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
