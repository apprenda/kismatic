package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ingress feature", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})

	Describe("accessing the ingress point of a cluster", func() {
		Context("when the cluster has an ingress node", func() {
			ItOnAWS("should return a successful response", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
					By("Installing a cluster with ingress")
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
					err = verifyIngressNodes(node, []NodeDeets{node}, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
