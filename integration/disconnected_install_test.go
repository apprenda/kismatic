package integration

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disconnected install feature", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("installing on machines with no internet access", func() {
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

					if err := verifyNoInternetAccess(theNode, sshKey); err == nil {
						Fail("was able to ping google with outgoing connections blocked")
					}

					By("Running kismatic install apply")
					installOpts := installOptions{
						disablePackageInstallation:  true,
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

	Describe("using an existing private docker registry with images pre-seeded", func() {
		ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{1, 1, 1, 0, 0}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
				dockerRegistryPort := 8443
				By("Configuring an insecure registry on the master")
				cmds := []string{
					"sudo mkdir /etc/docker/",
					"sudo touch /etc/docker/daemon.json",
					fmt.Sprintf("printf '{\n  \"insecure-registries\" : [\"%s:%d\"]\n}\n' | sudo tee --append /etc/docker/daemon.json", nodes.etcd[0].PrivateIP, dockerRegistryPort),
				}
				err := runViaSSH(cmds, []NodeDeets{nodes.master[0]}, sshKey, 10*time.Minute)
				FailIfError(err, "Failed to allow insecure registries")

				By("Installing the RPMs on the node")
				InstallKismaticPackages(nodes, CentOS7, sshKey, true)

				By("Installing an external Docker registry on one of the nodes")
				caFile, err := deployDockerRegistry(nodes.etcd[0], dockerRegistryPort, sshKey)
				Expect(err).ToNot(HaveOccurred())

				By("Disabling internet access")
				err = disableInternetAccess(nodes.allNodes(), sshKey)
				FailIfError(err, "Failed to create iptable rule")

				// disableRegistrySeeding = false, run step to seed
				installOpts := installOptions{
					disablePackageInstallation: true,
					disconnectedInstallation:   true,
					modifyHostsFiles:           true,
					dockerRegistryCAPath:       caFile,
					dockerRegistryIP:           nodes.etcd[0].PrivateIP,
					dockerRegistryPort:         dockerRegistryPort,
				}
				By("Seeding images")
				writePlanFile(buildPlan(nodes, installOpts, sshKey))
				c := exec.Command("./kismatic", "install", "step", "_docker-registry.yaml", "-f", "kismatic-testing.yaml")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				Expect(err).ToNot(HaveOccurred())

				installOpts.disableRegistrySeeding = true
				By("Running kismatic install apply")
				err = installKismatic(nodes, installOpts, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

func disableInternetAccess(nodes []NodeDeets, sshKey string) error {
	By("Blocking all outbound connections")
	allowSourcePorts := "8888,2379,6666,2380,6660,6443,8443,80,443,4194,10249,10250,10251,10252,10254" // ports needed/checked by inspector
	allowDestPorts := "8888,2379,6666,2380,6660,6443,8443,10250"
	cmd := []string{
		"sudo iptables -A OUTPUT -o lo -j ACCEPT",                                                               // allow loopback
		"sudo iptables -A OUTPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT",                      // allow SSH
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %s -j ACCEPT", allowSourcePorts), // allow inspector
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --dports %s -j ACCEPT", allowDestPorts),   // allow internal traffic for: inspector, etcd, docker registry
		"sudo iptables -A OUTPUT -s 172.16.0.0/16 -j ACCEPT",
		"sudo iptables -A OUTPUT -d 172.16.0.0/16 -j ACCEPT", // Allow pod network
		"sudo iptables -A OUTPUT -s 172.20.0.0/16 -j ACCEPT",
		"sudo iptables -A OUTPUT -d 172.20.0.0/16 -j ACCEPT", // Allow pod service network
		"sudo iptables -P OUTPUT DROP",                       // drop everything else
	}
	return runViaSSH(cmd, nodes, sshKey, 1*time.Minute)
}

func verifyNoInternetAccess(nodes []NodeDeets, sshKey string) error {
	By("Verifying that connections are blocked")
	return runViaSSH([]string{"curl --head www.google.com"}, nodes, sshKey, 1*time.Minute)
}
