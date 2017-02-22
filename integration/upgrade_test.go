package integration

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kismatic upgrade tests", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDirWithVersion("v1.2.2")
		os.Chdir(dir)
	})

	Describe("Doing an offline upgrade of a kubernetes cluster", func() {
		Context("Using a minikube layout", func() {
			Context("Using Ubuntu 16.04", func() {
				ItOnAWS("should result in an upgraded cluster [slow] [upgrade]", func(aws infrastructureProvisioner) {
					WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
						// Install previous version cluster
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())

						// Extract current version of kismatic
						pwd, err := os.Getwd()
						Expect(err).ToNot(HaveOccurred())
						err = extractCurrentKismatic(pwd)
						Expect(err).ToNot(HaveOccurred())

						// Perform upgrade
						cmd := exec.Command("./kismatic", "upgrade", "offline", "-f", "kismatic-testing.yaml", "--skip-preflight")
						cmd.Stderr = os.Stderr
						cmd.Stdout = os.Stdout
						err = cmd.Run()
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			PContext("Using CentOS 7", func() {
				ItOnAWS("should result in an upgraded cluster [slow] [upgrade]", func(aws infrastructureProvisioner) {
					WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
						// Install previous version cluster
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())

						// Extract new version of kismatic
						pwd, err := os.Getwd()
						Expect(err).ToNot(HaveOccurred())
						err = extractCurrentKismatic(pwd)
						Expect(err).ToNot(HaveOccurred())

						// Perform upgrade
						cmd := exec.Command("./kismatic", "upgrade", "offline", "-f", "kismatic-testing.yaml", "--skip-preflight")
						cmd.Stderr = os.Stderr
						cmd.Stdout = os.Stdout
						err = cmd.Run()
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

	})
})
