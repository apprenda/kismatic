package integration

import (
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Happy Path Installation Tests", func() {
	Describe("Calling installer with no input", func() {
		It("should output help text", func() {
			c := exec.Command("./kismatic")
			helpbytes, helperr := c.Output()
			Expect(helperr).To(BeNil())
			helpText := string(helpbytes)
			Expect(helpText).To(ContainSubstring("Usage"))
		})
	})

	Describe("Calling installer with 'install plan'", func() {
		Context("and just hitting enter", func() {
			It("should result in the output of a well formed default plan file", func() {
				By("Outputing a file")
				c := exec.Command("./kismatic", "install", "plan")
				helpbytes, helperr := c.Output()
				Expect(helperr).To(BeNil())
				helpText := string(helpbytes)
				Expect(helpText).To(ContainSubstring("Generating installation plan file with 3 etcd nodes, 2 master nodes and 3 worker nodes"))
				Expect(FileExists("kismatic-cluster.yaml")).To(Equal(true))

				By("Outputing a file with valid YAML")
				yamlBytes, err := ioutil.ReadFile("kismatic-cluster.yaml")
				if err != nil {
					Fail("Could not read cluster file")
				}
				yamlBlob := string(yamlBytes)

				planFromYaml := ClusterPlan{}

				unmarshallErr := yaml.Unmarshal([]byte(yamlBlob), &planFromYaml)
				if unmarshallErr != nil {
					Fail("Could not unmarshall cluster yaml: %v")
				}
			})
		})
	})
	Describe("Calling installer with a plan targeting bad infrastructure", func() {
		Context("Using a 1/1/1 Ubuntu 16.04 layout pointing to bad ip addresses", func() {
			It("should bomb validate and apply", func() {
				if !completesInTime(installKismaticWithABadNode, 30*time.Second) {
					Fail("It shouldn't take 30 seconds for Kismatic to fail with bad nodes.")
				}
			})
		})
	})

	Describe("Installing with package installation enabled", func() {
		installOpts := InstallOptions{
			AllowPackageInstallation: true,
		}
		Context("Targeting AWS infrastructure", func() {
			Context("using a 1/1/1 layout with Ubuntu 16.04 LTS", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("using a 1/1/1 layout with CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("using a 3/2/3 layout with CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{3, 2, 3}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing against a minikube layout", func() {
		Context("Targeting AWS infrastructure", func() {
			Context("Using CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(CentOS7, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("Using Ubuntu 16.04 LTS", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(Ubuntu1604LTS, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("Targeting Packet Infrastructure", func() {
			Context("Using CentOS 7", func() {
				ItOnPacket("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(CentOS7, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing with package installation disabled", func() {
		installOpts := InstallOptions{
			AllowPackageInstallation: false,
		}
		Context("Targeting AWS infrastructure", func() {
			Context("Using a 1/1/1 layout with Ubuntu 16.04 LTS", func() {
				ItOnAWS("Should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
						By("Installing the Kismatic RPMs")
						InstallKismaticRPMs(nodes, Ubuntu1604LTS, sshKey)
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("Using a 1/1/1 CentOS 7 layout", func() {
				ItOnAWS("Should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						By("Installing the Kismatic RPMs")
						InstallKismaticRPMs(nodes, CentOS7, sshKey)
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing with private Docker registry", func() {
		Context("Using a 1/1/1 CentOS 7 layout", func() {
			nodeCount := NodeCount{Etcd: 1, Master: 1, Worker: 1}
			distro := CentOS7

			Context("Using the auto-configured docker registry", func() {
				ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
					WithInfrastructure(nodeCount, distro, aws, func(nodes provisionedNodes, sshKey string) {
						installOpts := InstallOptions{
							AllowPackageInstallation:    true,
							AutoConfigureDockerRegistry: true,
						}
						err := InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("Using a custom registry provided by the user", func() {
				ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
					WithInfrastructure(nodeCount, distro, aws, func(nodes provisionedNodes, sshKey string) {
						By("Installing an external Docker registry on one of the etcd nodes")
						dockerRegistryPort := 443
						caFile, err := deployDockerRegistry(nodes.etcd[0], dockerRegistryPort, sshKey)
						Expect(err).ToNot(HaveOccurred())
						installOpts := InstallOptions{
							AllowPackageInstallation: true,
							DockerRegistryCAPath:     caFile,
							DockerRegistryIP:         nodes.etcd[0].PrivateIP,
							DockerRegistryPort:       dockerRegistryPort,
						}
						err = InstallKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
})
