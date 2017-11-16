package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("NFS Shares", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Context("Specifying valid NFS shares in the plan file", func() {
		Context("targeting CentOS", func() {
			ItOnAWS("should result in a working deployment [slow]", func(aws infrastructureProvisioner) {
				testNFSShare(aws, CentOS7)
			})
		})
		Context("targeting Ubuntu", func() {
			ItOnAWS("should result in a working deployment [slow]", func(aws infrastructureProvisioner) {
				testNFSShare(aws, Ubuntu1604LTS)
			})
		})
		Context("targeting RHEL", func() {
			ItOnAWS("should result in a working deployment [slow]", func(aws infrastructureProvisioner) {
				testNFSShare(aws, RedHat7)
			})
		})
	})
})
