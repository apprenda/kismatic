package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kismatic", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("calling kismatic with no verb", func() {
		It("should output help text", func() {
			c := exec.Command("./kismatic")
			helpbytes, helperr := c.Output()
			Expect(helperr).To(BeNil())
			helpText := string(helpbytes)
			Expect(helpText).To(ContainSubstring("Usage"))
		})
	})

	Describe("Calling 'install plan'", func() {
		Context("and just hitting enter", func() {
			It("should result in the output of a well formed default plan file", func() {
				By("Outputing a file")
				c := exec.Command("./kismatic", "install", "plan")
				helpbytes, helperr := c.Output()
				Expect(helperr).To(BeNil())
				helpText := string(helpbytes)
				Expect(helpText).To(ContainSubstring("Generating installation plan file template"))
				Expect(helpText).To(ContainSubstring("3 etcd nodes"))
				Expect(helpText).To(ContainSubstring("2 master nodes"))
				Expect(helpText).To(ContainSubstring("3 worker nodes"))
				Expect(helpText).To(ContainSubstring("2 ingress nodes"))
				Expect(helpText).To(ContainSubstring("0 storage nodes"))
				Expect(helpText).To(ContainSubstring("0 nfs volumes"))

				Expect(FileExists("kismatic-cluster.yaml")).To(Equal(true))

				By("Reading generated plan file")
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

				By("Verifying generated plan file")
				Expect(planFromYaml.Etcd.ExpectedCount).To(Equal(3))
				Expect(planFromYaml.Master.ExpectedCount).To(Equal(2))
				Expect(planFromYaml.Worker.ExpectedCount).To(Equal(3))
				Expect(planFromYaml.Ingress.ExpectedCount).To(Equal(2))
				Expect(planFromYaml.Storage.ExpectedCount).To(Equal(0))
				Expect(len(planFromYaml.NFS.Volumes)).To(Equal(0))
			})
		})
	})

	Describe("calling install apply", func() {
		Context("when targetting non-existent infrastructure", func() {
			It("should fail in a reasonable amount of time", func() {
				if !completesInTime(installKismaticWithABadNode, 600*time.Second) {
					Fail("It shouldn't take 600 seconds for Kismatic to fail with bad nodes.")
				}
			})
		})

		Context("when deploying a cluster with all node roles", func() {
			installOpts := installOptions{
				allowPackageInstallation: true,
			}
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targetting CentOS", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targetting RHEL", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targetting Ubuntu", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when using direct-lvm docker storage", func() {
			installOpts := installOptions{
				allowPackageInstallation: true,
				useDirectLVM:             true,
			}
			Context("when targetting CentOS", func() {
				ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
					WithMiniInfrastructureAndBlockDevice(CentOS7, aws, func(node NodeDeets, sshKey string) {
						theNode := []NodeDeets{node}
						nodes := provisionedNodes{
							etcd:    theNode,
							master:  theNode,
							worker:  theNode,
							ingress: theNode,
						}
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("when targetting RHEL", func() {
				ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
					WithMiniInfrastructureAndBlockDevice(RedHat7, aws, func(node NodeDeets, sshKey string) {
						theNode := []NodeDeets{node}
						nodes := provisionedNodes{
							etcd:    theNode,
							master:  theNode,
							worker:  theNode,
							ingress: theNode,
						}
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		// This spec will be used for testing non-destructive kismatic features on
		// a new cluster.
		// This spec is open to modification when new assertions have to be made
		Context("when deploying a skunkworks cluster", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructureAndDNS(NodeCount{3, 2, 3, 2, 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					// reserve one of the workers for the add-worker test
					allWorkers := nodes.worker
					nodes.worker = allWorkers[0 : len(nodes.worker)-1]

					// install cluster
					installOpts := installOptions{allowPackageInstallation: true, enableNetworkPolicy: true}
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())

					sub := SubDescribe("Using a running cluster")
					defer sub.Check()

					sub.It("should allow adding a worker node", func() error {
						newWorker := allWorkers[len(allWorkers)-1]
						return addWorkerToCluster(newWorker)
					})

					sub.It("should be able to deploy a workload with ingress", func() error {
						return verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
					})

					// Use master[0] public IP
					sub.It("should have an accessible dashboard", func() error {
						return canAccessDashboard(fmt.Sprintf("https://admin:abbazabba@%s:6443/ui", nodes.master[0].PublicIP))
					})

					sub.It("should respect network policies", func() error {
						return verifyNetworkPolicy(nodes.master[0], sshKey)
					})

					// This test should always be last
					sub.It("should still be a highly available cluster after removing a master node", func() error {
						By("Removing a Kubernetes master node")
						if err = aws.TerminateNode(nodes.master[0]); err != nil {
							return fmt.Errorf("could not remove node: %v", err)
						}
						By("Re-running Kuberang")
						if err = runViaSSH([]string{"sudo kuberang"}, []NodeDeets{nodes.master[1]}, sshKey, 5*time.Minute); err != nil {
							return err
						}
						return nil
					})
				})
			})
		})

		ItOnPacket("should install successfully [slow]", func(packet infrastructureProvisioner) {
			WithMiniInfrastructure(Ubuntu1604LTS, packet, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
