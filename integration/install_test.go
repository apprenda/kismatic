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
		os.Chdir(kisPath)
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

	Describe("calling `install apply`", func() {
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
				WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when deploying a mini-kube style cluster", func() {
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

		Context("when deploying a skunkworks cluster", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithInfrastructure(NodeCount{3, 2, 3, 2, 2}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
					installOpts := installOptions{allowPackageInstallation: true}
					err := installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		ItOnPacket("should install successfully [slow]", func(packet infrastructureProvisioner) {
			WithMiniInfrastructure(CentOS7, packet, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
