package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kismatic install add-worker tests", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})
	Context("Targeting AWS infrastructure", func() {
		Context("Using Ubuntu 16.04 LTS", func() {
			ItOnAWS("should successfully add a new worker", func(provisioner infrastructureProvisioner) {
				WithInfrastructure(NodeCount{Worker: 2}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
					theNode := nodes.worker[0]
					err := installKismaticMini(theNode, sshKey)
					Expect(err).ToNot(HaveOccurred())

					newWorker := nodes.worker[1]
					err = addWorkerToKismaticMini(newWorker)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
