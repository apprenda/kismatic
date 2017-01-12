package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Storage", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})

	Context("Specifying multiple storage nodes in the plan file", func() {
		Context("targetting CentOS", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, CentOS7)
			})
		})
		Context("targetting Ubuntu", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, Ubuntu1604LTS)
			})
		})
		Context("targetting RHEL", func() {
			ItOnAWS("should result in a working storage cluster", func(aws infrastructureProvisioner) {
				testAddVolumeVerifyGluster(aws, RedHat7)
			})
		})
	})
})
