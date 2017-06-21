package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const previousKismaticVersion = "v1.3.2"

// Test a specific released version of Kismatic
var _ = Describe("Installing with previous version of Kismatic", func() {
	BeforeEach(func() {
		// setup previous version of Kismatic
		tmp := setupTestWorkingDirWithVersion(previousKismaticVersion)
		os.Chdir(tmp)
	})

	installOpts := installOptions{
		serviceCIDR: "172.17.0.0/16",
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
