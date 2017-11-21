package integration

import (
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
		Context("when targeting non-existent infrastructure", func() {
			It("should fail in a reasonable amount of time", func() {
				if !completesInTime(installKismaticWithABadNode, 600*time.Second) {
					Fail("It shouldn't take 600 seconds for Kismatic to fail with bad nodes.")
				}
			})
		})

		Context("when deploying a cluster with all node roles", func() {
			installOpts := installOptions{}
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when deploying a cluster with all node roles and cloud-provider on CentOS", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 2, 1, 1}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
					testCloudProvider(nodes, sshKey)
				})
			})
		})

		Context("when deploying a cluster with all node roles and cloud-provider on RHEL", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 2, 1, 1}, RedHat7, aws, func(nodes provisionedNodes, sshKey string) {
					testCloudProvider(nodes, sshKey)
				})
			})
		})

		Context("when deploying a cluster with all node roles and cloud-provider on Ubuntu", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 2, 1, 1}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					testCloudProvider(nodes, sshKey)
				})
			})
		})

		Context("when deploying a cluster with all node roles and disabled CNI", func() {
			installOpts := installOptions{
				disableCNI: true,
			}
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targeting CentOS", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
					// Ensure preflight checks are idempotent on CentOS7
					err = runValidate("kismatic-testing.yaml")
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targeting RHEL", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
					// Ensure preflight checks are idempotent on RedHat7
					err = runValidate("kismatic-testing.yaml")
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when targeting Ubuntu", func() {
			ItOnAWS("should install successfully", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
					err := installKismaticMini(node, sshKey)
					Expect(err).ToNot(HaveOccurred())
					// Ensure preflight checks are idempotent on Ubuntu 1604
					err = runValidate("kismatic-testing.yaml")
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when using direct-lvm docker storage", func() {
			installOpts := installOptions{
				useDirectLVM: true,
			}
			Context("when targeting CentOS", func() {
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

			Context("when targeting RHEL", func() {
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

		Context("when deploying an HA cluster", func() {
			ItOnAWS("should still be a highly available cluster after removing a master node [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructureAndDNS(NodeCount{1, 2, 1, 0, 0}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
					// install cluster
					installOpts := installOptions{}
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())

					By("Removing a Kubernetes master node")
					if err = aws.TerminateNode(nodes.master[0]); err != nil {
						FailIfError(err, "could not remove node")
					}
					By("Re-running Kuberang")
					if err = runViaSSH([]string{"sudo kuberang"}, []NodeDeets{nodes.master[1]}, sshKey, 5*time.Minute); err != nil {
						FailIfError(err, "kuberang error")
					}
				})
			})
		})

		// This spec will be used for testing non-destructive kismatic features on
		// a new cluster.
		// This spec is open to modification when new assertions have to be made
		Context("when deploying a skunkworks cluster", func() {
			Context("with Calico as the CNI provider", func() {
				ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
					WithInfrastructure(NodeCount{3, 2, 3, 2, 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
						// reserve one of the workers for the add-worker test
						allWorkers := nodes.worker
						nodes.worker = allWorkers[0 : len(nodes.worker)-1]

						// install cluster
						installOpts := installOptions{
							heapsterReplicas:             3,
							heapsterInfluxdbPVC:          "influxdb",
							kubeAPIServerOptions:         map[string]string{"v": "3"},
							kubeControllerManagerOptions: map[string]string{"v": "3"},
							kubeSchedulerOptions:         map[string]string{"v": "3"},
							kubeProxyOptions:             map[string]string{"v": "3"},
							kubeletOptions:               map[string]string{"v": "3"},
						}
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())

						sub := SubDescribe("Using a running cluster")
						defer sub.Check()

						sub.It("should allow adding a worker node", func() error {
							newWorker := allWorkers[len(allWorkers)-1]
							return addWorkerToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"})
						})

						sub.It("should be able to deploy a workload with ingress", func() error {
							return verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						})

						// Use master[0] public IP
						// sub.It("should have an accessible dashboard", func() error {
						// 	return canAccessDashboard(fmt.Sprintf("https://admin:abbazabba@%s:6443/ui", nodes.master[0].PublicIP))
						// })

						sub.It("should respect network policies", func() error {
							return verifyNetworkPolicy(nodes.master[0], sshKey)
						})

						sub.It("should support heapster with persistent storage", func() error {
							return verifyHeapster(nodes.master[0], sshKey)
						})

						sub.It("should have tiller running", func() error {
							return verifyTiller(nodes.master[0], sshKey)
						})

						sub.It("nodes should contain expected labels", func() error {
							return containsLabels(nodes, sshKey)
						})

						sub.It("nodes should contain expected component overrides", func() error {
							return ContainsOverrides(nodes, sshKey)
						})

						sub.It("should allow for running preflight checks idempotently", func() error {
							return runValidate("kismatic-testing.yaml")
						})
					})
				})
			})
		})

		Context("when deploying a skunkworks cluster", func() {
			Context("with Weave as the CNI provider", func() {
				ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
					WithInfrastructure(NodeCount{3, 2, 3, 2, 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
						// reserve one of the workers for the add-worker test
						allWorkers := nodes.worker
						nodes.worker = allWorkers[0 : len(nodes.worker)-1]

						// install cluster
						installOpts := installOptions{
							heapsterReplicas:    3,
							heapsterInfluxdbPVC: "influxdb",
							cniProvider:         "weave",
						}
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())

						sub := SubDescribe("Using a running cluster")
						defer sub.Check()

						sub.It("should allow adding a worker node", func() error {
							newWorker := allWorkers[len(allWorkers)-1]
							return addWorkerToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"})
						})

						sub.It("should be able to deploy a workload with ingress", func() error {
							return verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						})

						// Use master[0] public IP
						// sub.It("should have an accessible dashboard", func() error {
						// 	return canAccessDashboard(fmt.Sprintf("https://admin:abbazabba@%s:6443/ui", nodes.master[0].PublicIP))
						// })

						sub.It("should respect network policies", func() error {
							return verifyNetworkPolicy(nodes.master[0], sshKey)
						})

						sub.It("should support heapster with persistent storage", func() error {
							return verifyHeapster(nodes.master[0], sshKey)
						})

						sub.It("should have tiller running", func() error {
							return verifyTiller(nodes.master[0], sshKey)
						})

						sub.It("nodes should contain expected labels", func() error {
							return containsLabels(nodes, sshKey)
						})

						sub.It("should allow for running preflight checks idempotently", func() error {
							return runValidate("kismatic-testing.yaml")
						})
					})
				})
			})
		})

		// Context("when deploying a skunkworks cluster", func() {
		// 	Context("with Contiv as the CNI provider", func() {
		// 		ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
		// 			WithInfrastructure(NodeCount{3, 2, 3, 2, 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
		// 				// reserve one of the workers for the add-worker test
		// 				allWorkers := nodes.worker
		// 				nodes.worker = allWorkers[0 : len(nodes.worker)-1]

		// 				// install cluster
		// 				installOpts := installOptions{
		// 					heapsterReplicas:    3,
		// 					heapsterInfluxdbPVC: "influxdb",
		// 					cniProvider:         "contiv",
		// 				}
		// 				err := installKismatic(nodes, installOpts, sshKey)
		// 				Expect(err).ToNot(HaveOccurred())

		// 				sub := SubDescribe("Using a running cluster")
		// 				defer sub.Check()

		// 				sub.It("should allow adding a worker node", func() error {
		// 					newWorker := allWorkers[len(allWorkers)-1]
		// 					return addWorkerToCluster(newWorker, sshKey, []string{})
		// 				})

		// 				// This test is flaky with contiv
		// 				// sub.It("should be able to deploy a workload with ingress", func() error {
		// 				// 	return verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
		// 				// })

		// 				// Use master[0] public IP
		// 				// There is an issue with contiv that prevents this test from passing consistently
		// 				// sub.It("should have an accessible dashboard", func() error {
		// 				// 	return canAccessDashboard(fmt.Sprintf("https://admin:abbazabba@%s:6443/ui", nodes.master[0].PublicIP))
		// 				// })

		// 				// Contiv does not support the Kubernetes network policy API
		// 				// sub.It("should respect network policies", func() error {
		// 				// 	return verifyNetworkPolicy(nodes.master[0], sshKey)
		// 				// })

		// 				sub.It("should support heapster with persistent storage", func() error {
		// 					return verifyHeapster(nodes.master[0], sshKey)
		// 				})

		// 				sub.It("should have tiller running", func() error {
		// 					return verifyTiller(nodes.master[0], sshKey)
		// 				})
		// 			})
		// 		})
		// 	})
		// })

		ItOnPacket("should install successfully [slow]", func(packet infrastructureProvisioner) {
			WithMiniInfrastructure(Ubuntu1604LTS, packet, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

func testCloudProvider(nodes provisionedNodes, sshKey string) {
	installOpts := installOptions{cloudProvider: "aws"}

	By("installing the cluster")
	err := installKismatic(nodes, installOpts, sshKey)
	Expect(err).ToNot(HaveOccurred())

	By("test the cloud provider integration")
	err = testAWSCloudProvider(nodes.master[0], sshKey)
	Expect(err).ToNot(HaveOccurred())
}
