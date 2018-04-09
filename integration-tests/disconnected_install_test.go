package integration_tests

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	proxyPort          = 3128
	dockerRegistryPort = 8443
)

var _ = Describe("disconnected installation", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Context("with an existing package mirror and image registry", func() {
		Context("on CentOS", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				distro := CentOS7
				WithInfrastructure(NodeCount{1, 1, 2, 1, 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
					// One of the workers is used as the repository and registry
					repoNode := nodes.worker[1]
					nodes.worker = nodes.worker[0:1]

					By("Creating a package repository")
					err := createPackageRepositoryMirror(repoNode, distro, sshKey)
					FailIfError(err, "Error creating local package repo")

					By("Deploying a docker registry")
					caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
					FailIfError(err, "Failed to deploy docker registry")

					By("Seeding the local registry")
					err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
					FailIfError(err, "Error seeding local registry")

					By("Disabling internet access")
					err = disableInternetAccess(nodes.allNodes(), sshKey)
					FailIfError(err, "Failed to create iptable rule")

					// Configure the repos on the nodes
					By("Configuring repository on nodes")
					for _, n := range nodes.allNodes() {
						err = copyFileToRemote("test-resources/disconnected-installation/configure-rpm-mirrors.sh", "/tmp/configure-rpm-mirrors.sh", n, sshKey, 15*time.Second)
						FailIfError(err, "Failed to copy script to nodes")
					}
					cmds := []string{
						"chmod +x /tmp/configure-rpm-mirrors.sh",
						fmt.Sprintf("sudo /tmp/configure-rpm-mirrors.sh http://%s", repoNode.PrivateIP),
					}
					err = runViaSSH(cmds, nodes.allNodes(), sshKey, 5*time.Minute)
					FailIfError(err, "Failed to run mirror configuration script")

					installOpts := installOptions{
						disablePackageInstallation: false,
						disconnectedInstallation:   true,
						modifyHostsFiles:           true,
						dockerRegistryCAPath:       caFile,
						dockerRegistryServer:       fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
						dockerRegistryUsername:     "kismaticuser",
						dockerRegistryPassword:     "kismaticpassword",
					}

					// installOpts.disableRegistrySeeding = true
					By("Running kismatic apply")
					err = installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("on Ubuntu", func() {
			ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
				distro := Ubuntu1604LTS
				WithInfrastructure(NodeCount{1, 1, 2, 1, 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
					// One of the workers is used as the repository and registry
					repoNode := nodes.worker[1]
					nodes.worker = nodes.worker[0:1]

					By("Creating a package repository")
					err := createPackageRepositoryMirror(repoNode, distro, sshKey)
					FailIfError(err, "Error creating local package repo")

					// Deploy a docker registry on the node
					By("Deploying a docker registry")
					caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
					FailIfError(err, "Failed to deploy docker registry")

					By("Seeding the local registry")
					err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
					FailIfError(err, "Error seeding local registry")

					By("Disabling internet access")
					err = disableInternetAccess(nodes.allNodes(), sshKey)
					FailIfError(err, "Failed to create iptable rule")

					// Configure the repos on the nodes
					By("Configuring repository on nodes")
					for _, n := range nodes.allNodes() {
						err = copyFileToRemote("test-resources/disconnected-installation/configure-deb-mirrors.sh", "/tmp/configure-deb-mirrors.sh", n, sshKey, 15*time.Second)
						FailIfError(err, "Failed to copy script to nodes")
					}
					cmds := []string{
						"chmod +x /tmp/configure-deb-mirrors.sh",
						fmt.Sprintf("sudo /tmp/configure-deb-mirrors.sh http://%s", repoNode.PrivateIP),
					}
					err = runViaSSH(cmds, nodes.allNodes(), sshKey, 5*time.Minute)
					FailIfError(err, "Failed to run mirror configuration script")

					installOpts := installOptions{
						disablePackageInstallation: false,
						disconnectedInstallation:   true,
						modifyHostsFiles:           true,
						dockerRegistryCAPath:       caFile,
						dockerRegistryServer:       fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
						dockerRegistryUsername:     "kismaticuser",
						dockerRegistryPassword:     "kismaticpassword",
					}

					// installOpts.disableRegistrySeeding = true
					By("Running kismatic apply")
					err = installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Context("when there is a proxy between the nodes and the internet", func() {
		ItOnAWS("should install successfully [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{2, 1, 1, 1, 1}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
				// setup cluster nodes, and a proxy node
				clusterNodes := provisionedNodes{}
				clusterNodes.etcd = []NodeDeets{nodes.etcd[1]}
				clusterNodes.master = nodes.master
				clusterNodes.worker = nodes.worker
				clusterNodes.ingress = nodes.ingress
				clusterNodes.storage = nodes.storage
				proxyNode := nodes.etcd[0]

				By("Installing the proxy")
				err := runViaSSH([]string{"sudo yum install -y squid"}, []NodeDeets{proxyNode}, sshKey, 5*time.Minute)
				FailIfError(err, "Failed install proxy")
				err = runViaSSH([]string{"sudo sed -i -e 's/http_access deny all/http_access allow all/g' /etc/squid/squid.conf"}, []NodeDeets{proxyNode}, sshKey, 5*time.Minute)
				FailIfError(err, "Failed modify squif.conf")
				err = runViaSSH([]string{"sudo systemctl restart squid"}, []NodeDeets{proxyNode}, sshKey, 5*time.Minute)
				FailIfError(err, "Failed install proxy")

				By("Disabling internet access")
				err = disableInternetAccess(clusterNodes.allNodes(), sshKey)
				FailIfError(err, "Failed to create iptable rules")

				if err = verifyNoInternetAccess(clusterNodes.allNodes(), sshKey); err == nil {
					Fail("was able to ping google with outgoing connections blocked")
				}

				By("Enabling proxy access")
				err = enableProxyConnection(clusterNodes.allNodes(), sshKey)
				FailIfError(err, "Failed to create iptable rules")

				By("Verifying connectivity to google.com")
				err = runViaSSH([]string{fmt.Sprintf("export http_proxy=%s:%d && curl --head www.google.com", proxyNode.PrivateIP, proxyPort)}, clusterNodes.allNodes(), sshKey, 1*time.Minute)
				FailIfError(err, "Failed to curl google with proxy")

				By("Running kismatic apply")
				// don't use the proxy for cluster communication

				installOpts := installOptions{
					modifyHostsFiles: true,
					httpProxy:        fmt.Sprintf("%s:%d", proxyNode.PrivateIP, proxyPort),
					httpsProxy:       fmt.Sprintf("%s:%d", proxyNode.PrivateIP, proxyPort),
					noProxy:          fmt.Sprintf("%s,%s", "kubernetes-charts.storage.googleapis.com", "apprenda.github.io"),
				}
				err = installKismatic(clusterNodes, installOpts, sshKey)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

func disableInternetAccess(nodes []NodeDeets, sshKey string) error {
	By("Blocking all outbound connections")
	allowSourcePorts := "8888,2379,6666,2380,6660,6443,8443,80,443,10256,10250,10251,10252,10254" // ports needed/checked by inspector
	allowDestPorts := "8888,2379,6666,2380,6660,6443,8443,80,443,10256,10250,10251,10252,10254"
	allowedStorgePorts := "8081,111,2049,38465,38466,38467"
	cmd := []string{
		"sudo iptables -A OUTPUT -o lo -j ACCEPT",                                                                 // allow loopback
		"sudo iptables -A OUTPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT",                        // allow SSH
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %s -j ACCEPT", allowSourcePorts),   // allow inspector
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --dports %s -j ACCEPT", allowDestPorts),     // allow internal traffic for: inspector, etcd, docker registry
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %s -j ACCEPT", allowedStorgePorts), // storage
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --dports %s -j ACCEPT", allowedStorgePorts), // storage
		"sudo iptables -A OUTPUT -s 172.16.0.0/16 -j ACCEPT",
		"sudo iptables -A OUTPUT -d 172.16.0.0/16 -j ACCEPT", // Allow pod network
		"sudo iptables -A OUTPUT -s 172.20.0.0/16 -j ACCEPT",
		"sudo iptables -A OUTPUT -d 172.20.0.0/16 -j ACCEPT",                 // Allow pod service network
		"sudo iptables -A INPUT -p icmp --icmp-type echo-request -j ACCEPT",  // ping outside to inside
		"sudo iptables -A OUTPUT -p icmp --icmp-type echo-reply -j ACCEPT",   // ping outside to inside
		"sudo iptables -A OUTPUT -p icmp --icmp-type echo-request -j ACCEPT", // ping inside to outside
		"sudo iptables -A INPUT -p icmp --icmp-type echo-reply -j ACCEPT",    // ping inside to outside
		"sudo iptables -P OUTPUT DROP",                                       // drop everything else
	}
	return runViaSSH(cmd, nodes, sshKey, 1*time.Minute)
}

func enableProxyConnection(nodes []NodeDeets, sshKey string) error {
	By("Allowing connection to proxy")
	// proxy server running on 3128
	cmd := []string{
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %d -j ACCEPT", proxyPort),
		fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --dports %d -j ACCEPT", proxyPort),
	}
	return runViaSSH(cmd, nodes, sshKey, 1*time.Minute)
}

func verifyNoInternetAccess(nodes []NodeDeets, sshKey string) error {
	By("Verifying that connections are blocked")
	return runViaSSH([]string{"curl --head www.google.com"}, nodes, sshKey, 1*time.Minute)
}
