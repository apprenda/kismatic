package integration

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disconnected install feature", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("Installing on machines with no internet access", func() {
		Context("with kismatic packages installed", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
					By("Installing the RPMs on the node")
					theNode := []NodeDeets{node}
					nodes := provisionedNodes{
						etcd:    theNode,
						master:  theNode,
						worker:  theNode,
						ingress: theNode,
					}
					InstallKismaticPackages(nodes, CentOS7, sshKey, true)

					By("Verifying connectivity to google.com")
					err := runViaSSH([]string{"curl --head www.google.com"}, theNode, sshKey, 1*time.Minute)
					FailIfError(err, "Failed to curl google")

					err = disableInternetAccess(theNode, sshKey)
					FailIfError(err, "Failed to create iptable rule")

					By("Verifying that connections are blocked")
					err = runViaSSH([]string{"curl --max-time 5 www.google.com"}, theNode, sshKey, 1*time.Minute)
					if err == nil {
						Fail("was able to ping google with outgoing connections blocked")
					}

					By("Running kismatic install apply")
					installOpts := installOptions{
						allowPackageInstallation:    false,
						disconnectedInstallation:    true,
						modifyHostsFiles:            true,
						autoConfigureDockerRegistry: true,
					}
					err = installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})

func disableInternetAccess(nodes []NodeDeets, sshKey string) error {
	By("Blocking all outbound connections")
	allowPorts := "8888,2379,6666,2380,6660,6443,8443,80,443,4194,10249,10250,10251,10252,10254" // ports needed/checked by inspector
	cmd := []string{
		"sudo iptables -A OUTPUT -o lo -j ACCEPT",                                                         // allow loopback
		"sudo iptables -A OUTPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT",                // allow SSH
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %s -j ACCEPT", allowPorts), // allow inspector
		"sudo iptables -A OUTPUT -s 172.16.0.0/16 -j ACCEPT",
		"sudo iptables -A OUTPUT -d 172.16.0.0/16 -j ACCEPT", // Allow pod network
		"sudo iptables -P OUTPUT DROP",                       // drop everything else
	}
	return runViaSSH(cmd, nodes, sshKey, 1*time.Minute)
}
