package integration

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ItOnAWS runs a spec if the AWS details have been provided
func ItOnAWS(description string, f func(infrastructureProvisioner)) {
	Context("when using AWS infrastructure [aws]", func() {
		It(description, func() {
			awsClient, ok := AWSClientFromEnvironment()
			if !ok {
				Skip("AWS environment variables were not defined")
			}
			f(awsClient)
		})
	})
}

// ItOnPacket runs a spec if the Packet.Net details have been provided
func ItOnPacket(description string, f func(infrastructureProvisioner)) {
	Context("when using Packet.net infrastructure [packet]", func() {
		It(description, func() {
			packetClient, ok := packetClientFromEnv()
			if !ok {
				Skip("Packet environment variables were not defined")
			}
			f(packetClient)
		})
	})
}

type infraDependentTest func(nodes provisionedNodes, sshKey string)

// WithInfrastructure runs the spec with the requested infrastructure
func WithInfrastructure(nodeCount NodeCount, distro linuxDistro, provisioner infrastructureProvisioner, f infraDependentTest) {
	By(fmt.Sprintf("Provisioning nodes: %+v", nodeCount))
	start := time.Now()
	nodes, err := provisioner.ProvisionNodes(nodeCount, distro)
	if !leaveIt() {
		defer provisioner.TerminateNodes(nodes)
	}
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Provisioning infrastructure took", time.Since(start))

	By("Waiting until nodes are SSH-accessible")
	start = time.Now()
	sshKey := provisioner.SSHKey()
	err = waitForSSH(nodes, sshKey)
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Waiting for SSH took", time.Since(start))

	f(nodes, sshKey)
}

// WithInfrastructureAndDNS runs the spec with the requested infrastructure and DNS
func WithInfrastructureAndDNS(nodeCount NodeCount, distro linuxDistro, provisioner infrastructureProvisioner, f infraDependentTest) {
	By(fmt.Sprintf("Provisioning nodes and DNS: %+v", nodeCount))
	start := time.Now()
	nodes, err := provisioner.ProvisionNodes(nodeCount, distro)
	if !leaveIt() {
		defer provisioner.TerminateNodes(nodes)
	}
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Provisioning infrastructure took", time.Since(start))

	By("Waiting until nodes are SSH-accessible")
	start = time.Now()
	sshKey := provisioner.SSHKey()
	err = waitForSSH(nodes, sshKey)
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Waiting for SSH took", time.Since(start))

	By("Configuring DNS entries")
	start = time.Now()
	var masterIPs []string
	for _, node := range nodes.master {
		masterIPs = append(masterIPs, node.PrivateIP)
	}
	dnsRecord, err := provisioner.ConfigureDNS(masterIPs)
	nodes.dnsRecord = dnsRecord
	Expect(err).ToNot(HaveOccurred())
	if !leaveIt() {
		By("Removing DNS entries")
		defer provisioner.RemoveDNS(dnsRecord)
	}
	fmt.Println("Configuring DNS entries took", time.Since(start))

	f(nodes, sshKey)
}

type miniInfraDependentTest func(node NodeDeets, sshKey string)

// WithMiniInfrastructure runs the spec with a Minikube-like infrastructure setup.
func WithMiniInfrastructure(distro linuxDistro, provisioner infrastructureProvisioner, f miniInfraDependentTest) {
	By("Provisioning minikube node")
	start := time.Now()
	nodes, err := provisioner.ProvisionNodes(NodeCount{Worker: 1}, distro)
	if !leaveIt() {
		defer provisioner.TerminateNodes(nodes)
	}
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Provisioning node took", time.Since(start))

	By("Waiting until nodes are SSH-accessible")
	start = time.Now()
	sshKey := provisioner.SSHKey()
	err = waitForSSH(nodes, sshKey)
	Expect(err).ToNot(HaveOccurred())
	fmt.Println("Waiting for SSH took", time.Since(start))

	f(nodes.worker[0], sshKey)
}
