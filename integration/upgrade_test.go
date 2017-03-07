package integration

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade", func() {
	Describe("Upgrading a cluster using offline mode", func() {
		Describe("From KET version v1.2.2", func() {
			BeforeEach(func() {
				dir := setupTestWorkingDirWithVersion("v1.2.2")
				os.Chdir(dir)
			})
			Context("Using a minikube layout", func() {
				Context("Using Ubuntu 16.04", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})

				Context("Using RedHat 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})
			})

			// This spec will be used for testing non-destructive kismatic features on
			// an upgraded cluster.
			// This spec is open to modification when new assertions have to be made.
			Context("Using a skunkworks cluster", func() {
				ItOnAWS("should result in an upgraded cluster [slow] [upgrade]", func(aws infrastructureProvisioner) {
					WithInfrastructureAndDNS(NodeCount{Etcd: 3, Master: 2, Worker: 3, Ingress: 2, Storage: 2}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
						// reserve one of the workers for the add-worker test
						allWorkers := nodes.worker
						nodes.worker = allWorkers[0 : len(nodes.worker)-1]

						// Standup cluster with previous version
						opts := installOptions{allowPackageInstallation: true}
						err := installKismatic(nodes, opts, sshKey)
						FailIfError(err)

						// Extract current version of kismatic
						extractCurrentKismaticInstaller()

						// Perform upgrade
						upgradeCluster()

						sub := SubDescribe("Using an upgraded cluster")
						defer sub.Check()

						sub.It("should allow adding a new storage volume", func() error {
							planFile, err := os.Open("kismatic-testing.yaml")
							if err != nil {
								return err
							}
							return createVolume(planFile, "test-vol", 1, 1, "")
						})

						sub.It("should allow adding a worker node", func() error {
							newWorker := allWorkers[len(allWorkers)-1]
							return addWorkerToCluster(newWorker)
						})

						sub.It("should have an accessible dashboard", func() error {
							return nil
						})

						sub.It("should be able to deploy a workload with ingress", func() error {
							return verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						})

						sub.It("should have an accessible dashboard", func() error {
							return canAccessDashboard()
						})
						sub.It("should not have kube-apiserver systemd service", func() error {
							return nil
						})
					})
				})
			})

			Context("Using a 1/2/1 cluster", func() {
				ItOnAWS("should still be a highly available cluster after upgrade", func(aws infrastructureProvisioner) {
					WithInfrastructureAndDNS(NodeCount{1, 2, 1, 0, 0}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
						opts := installOptions{allowPackageInstallation: true}
						err := installKismatic(nodes, opts, sshKey)
						Expect(err).ToNot(HaveOccurred())

						// Extract current version of kismatic
						extractCurrentKismaticInstaller()

						// Perform upgrade
						upgradeCluster()

						By("Removing a Kubernetes master node")
						err = aws.TerminateNode(nodes.master[0])
						Expect(err).ToNot(HaveOccurred(), "could not remove node")

						By("Re-running Kuberang")
						err = runViaSSH([]string{"sudo kuberang"}, []NodeDeets{nodes.master[1]}, sshKey, 5*time.Minute)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("Using a cluster that has no internet access", func() {
				ItOnAWS("should result in an upgraded cluster", func(aws infrastructureProvisioner) {
					distro := CentOS7
					WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
						// Standup cluster with previous version
						opts := installOptions{allowPackageInstallation: true}
						err := installKismatic(nodes, opts, sshKey)
						FailIfError(err)

						// Extract current version of kismatic
						extractCurrentKismaticInstaller()

						// Manually install the new packages
						InstallKismaticPackages(nodes, distro, sshKey, true)

						// Lock down internet access
						err = disableInternetAccess(nodes.allNodes(), sshKey)
						FailIfError(err)

						// Perform upgrade
						upgradeCluster()
					})
				})
			})
		})

		Describe("From KET version v1.1.1", func() {
			BeforeEach(func() {
				dir := setupTestWorkingDirWithVersion("v1.1.1")
				os.Chdir(dir)
			})
			Context("Using a minikube layout", func() {
				Context("Using CentOS 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})

				Context("Using RedHat 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})
			})
		})

		Describe("From KET version v1.0.3", func() {
			BeforeEach(func() {
				dir := setupTestWorkingDirWithVersion("v1.0.3")
				os.Chdir(dir)
			})
			Context("Using a minikube layout", func() {
				Context("Using CentOS 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})

				Context("Using Ubuntu 16.04", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey)
						})
					})
				})
			})
		})
	})
})

func installAndUpgradeMinikube(node NodeDeets, sshKey string) {
	// Install previous version cluster
	err := installKismaticMini(node, sshKey)
	FailIfError(err)
	extractCurrentKismaticInstaller()
	upgradeCluster()
}

func extractCurrentKismaticInstaller() {
	// Extract current version of kismatic
	pwd, err := os.Getwd()
	FailIfError(err)
	err = extractCurrentKismatic(pwd)
	FailIfError(err)
}
func upgradeCluster() {
	// Perform upgrade
	cmd := exec.Command("./kismatic", "upgrade", "offline", "-f", "kismatic-testing.yaml")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	FailIfError(err)

	assertClusterVersionIsCurrent()
}
