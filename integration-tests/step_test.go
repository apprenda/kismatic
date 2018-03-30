package integration_tests

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("install step commands", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("Running the api server play against an existing cluster", func() {
		ItOnAWS("should return successfully [slow]", func(aws infrastructureProvisioner) {
			WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())
				c := exec.Command("./kismatic", "step", defaultClusterName, "kube-apiserver")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
