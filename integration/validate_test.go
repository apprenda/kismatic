package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kismatic install validate tests", func() {
	Context("Targeting AWS infrastructure", func() {
		Describe("Running validation with package installation enabled", func() {
			Context("using a Minikube Ubuntu layout", func() {
				ItOnAWS("should succeed if and only if all packages are installed on the node", func(aws infrastructureProvisioner) {
					validateMiniPkgInstallEnabled(aws, Ubuntu1604LTS)
				})
			})
			Context("using a Minikube CentOS 7 layout", func() {
				ItOnAWS("should succeed if and only if all packages are installed on the node", func(aws infrastructureProvisioner) {
					validateMiniPkgInstallEnabled(aws, CentOS7)
				})
			})
		})

		Describe("Running validation with package installation disabled", func() {
			Context("using a Minikube Ubuntu layout", func() {
				ItOnAWS("should succeed if and only if all packages are installed on the node", func(aws infrastructureProvisioner) {
					validateMiniPkgInstallationDisabled(aws, Ubuntu1604LTS)
				})
			})
			Context("using a Minikube CentOS 7 layout", func() {
				ItOnAWS("should succeed if and only if all packages are installed on the node", func(aws infrastructureProvisioner) {
					validateMiniPkgInstallationDisabled(aws, CentOS7)
				})
			})
		})
	})
})

func validateMiniPkgInstallEnabled(provisioner infrastructureProvisioner, distro linuxDistro) {
	By("Provisioning nodes on AWS")
	nodes, err := provisioner.ProvisionNodes(NodeCount{Worker: 1}, distro)
	defer provisioner.TerminateNodes(nodes)
	FailIfError(err, "Failed to provision nodes for test")

	By("Waiting until nodes are SSH-accessible")
	sshUser := nodes.worker[0].SSHUser
	sshKey := provisioner.SSHKey()
	err = WaitForSSH(nodes, sshKey)
	FailIfError(err, "Error waiting for nodes to become SSH-accessible")
	ValidateKismaticMini(nodes.worker[0], sshUser, sshKey)
}

func validateMiniPkgInstallationDisabled(provisioner infrastructureProvisioner, distro linuxDistro) {
	By("Provisioning nodes on AWS")
	nodes, err := provisioner.ProvisionNodes(NodeCount{Worker: 1}, distro)
	defer provisioner.TerminateNodes(nodes)
	FailIfError(err, "Failed to provision nodes for test")

	By("Waiting until nodes are SSH-accessible")
	sshUser := nodes.worker[0].SSHUser
	sshKey := provisioner.SSHKey()
	FailIfError(err, "Failed to get SSH Key")
	err = WaitForSSH(nodes, sshKey)
	FailIfError(err, "Error waiting for nodes to become SSH-accessible")
	theNode := nodes.worker[0]

	if err = ValidateKismaticMiniDenyPkgInstallation(theNode, sshUser, sshKey); err == nil {
		Fail("Missing dependencies, but still passed")
	}

	By("Prepping nodes for the test")
	prep := getPrepForDistro(distro)
	prepNode := []NodeDeets{theNode}
	RunViaSSH(prep.CommandsToPrepRepo, sshUser, prepNode, 5*time.Minute)
	By("Installing etcd on the node")
	RunViaSSH(prep.CommandsToInstallEtcd, sshUser, prepNode, 5*time.Minute)
	if err = ValidateKismaticMiniDenyPkgInstallation(theNode, sshUser, sshKey); err == nil {
		Fail("Missing dependencies, but still passed")
	}

	By("Installing Docker")
	RunViaSSH(prep.CommandsToInstallDocker, sshUser, prepNode, 5*time.Minute)
	if err = ValidateKismaticMiniDenyPkgInstallation(theNode, sshUser, sshKey); err == nil {
		Fail("Missing dependencies, but still passed")
	}

	By("Installing Master")
	RunViaSSH(prep.CommandsToInstallK8sMaster, sshUser, prepNode, 5*time.Minute)
	err = ValidateKismaticMiniDenyPkgInstallation(theNode, sshUser, sshKey)
	Expect(err).To(BeNil())
}
