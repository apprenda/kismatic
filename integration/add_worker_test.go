package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kismatic install add-worker tests", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("adding a worker to an existing cluster", func() {
		ItOnAWS("should result in a cluster with an additional worker [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{Worker: 2}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
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
