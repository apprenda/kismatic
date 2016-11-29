package integration

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("install step commands", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})
	Context("Running the api server play against an existing cluster", func() {
		ItOnAWS("should return successfully", func(provisioner infrastructureProvisioner) {
			WithMiniInfrastructure(CentOS7, provisioner, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())

				c := exec.Command("./kismatic", "install", "step", "_apiserver.yaml", "-f", "kismatic-testing.yaml")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
