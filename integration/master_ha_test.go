package integration

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("control plane high availability feature", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("running kuberang against an HA cluster", func() {
		Context("when one of the master nodes is unavailable", func() {
			ItOnAWS("should result in a successful smoke test [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructureAndDNS(NodeCount{1, 2, 1, 0, 0}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
					opts := installOptions{allowPackageInstallation: true}
					err := installKismatic(nodes, opts, sshKey)
					Expect(err).ToNot(HaveOccurred())

					By("Removing a Kubernetes master node")
					err = aws.TerminateNode(nodes.master[0])
					Expect(err).ToNot(HaveOccurred(), "could not remove node")

					By("Rerunning Kuberang")
					err = runViaSSH([]string{"sudo kuberang"}, []NodeDeets{nodes.master[1]}, sshKey, 5*time.Minute)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
