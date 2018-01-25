package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Upgrade", func() {

	Describe("Upgrading a cluster using online mode", func() {
		Context("From KET version v1.6.3", func() {
			BeforeEach(func() {
				dir := setupTestWorkingDirWithVersion("v1.6.3")
				os.Chdir(dir)
			})

			Context("Using a minikube layout", func() {
				Context("Using CentOS 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey, true)
						})
					})
				})

				Context("Using RedHat 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey, true)
						})
					})
				})
			})

			// This spec will be used for testing non-destructive kismatic features on
			// an upgraded cluster.
			// This spec is open to modification when new assertions have to be made.
			Context("Using a skunkworks cluster", func() {
				ItOnAWS("should result in an upgraded cluster [slow] [upgrade]", func(aws infrastructureProvisioner) {
					WithInfrastructureAndDNS(NodeCount{Etcd: 3, Master: 2, Worker: 5, Ingress: 2, Storage: 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
						// reserve 3 of the workers for the add-node test
						allWorkers := nodes.worker
						nodes.worker = allWorkers[0 : len(nodes.worker)-3]

						// Standup cluster with previous version
						opts := installOptions{adminPassword: "abbazabba"}
						err := installKismatic(nodes, opts, sshKey)
						FailIfError(err)

						// Extract current version of kismatic
						extractCurrentKismaticInstaller()

						// Perform upgrade
						upgradeCluster(true)

						sub := SubDescribe("Using an upgraded cluster")
						defer sub.Check()

						sub.It("should have working storage volumes", func() error {
							return testStatefulWorkload(nodes, sshKey)
						})

						sub.It("should allow adding a worker node", func() error {
							newWorker := allWorkers[len(allWorkers)-1]
							return addNodeToCluster(newWorker, sshKey, []string{}, []string{})
						})

						sub.It("should allow adding a ingress node", func() error {
							newWorker := allWorkers[len(allWorkers)-2]
							return addNodeToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"}, []string{"ingress"})
						})

						sub.It("should allow adding a storage node", func() error {
							newWorker := allWorkers[len(allWorkers)-3]
							return addNodeToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"}, []string{"storage"})
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
						sub.It("should still be a highly available cluster after upgrade", func() error {
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

			Context("Using a cluster that has no internet access [slow] [upgrade]", func() {
				Context("With nodes running CentOS 7", func() {
					ItOnAWS("should result in an upgraded cluster", func(aws infrastructureProvisioner) {
						distro := CentOS7
						WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Ingress: 1, Storage: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
							// One of the nodes will function as a repo mirror and image registry
							repoNode := nodes.worker[1]
							nodes.worker = nodes.worker[0:1]
							// Standup cluster with previous version
							opts := installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: false, // we want KET to install the packages, so let it use the package repo
								modifyHostsFiles:         true,
							}
							err := installKismatic(nodes, opts, sshKey)
							FailIfError(err)

							// Extract current version of kismatic
							extractCurrentKismaticInstaller()

							By("Creating a package repository")
							err = createPackageRepositoryMirror(repoNode, distro, sshKey)
							FailIfError(err, "Error creating local package repo")

							By("Deploying a docker registry")
							caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
							FailIfError(err, "Failed to deploy docker registry")

							By("Seeding the local registry")
							err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
							FailIfError(err, "Error seeding local registry")

							err = disableInternetAccess(nodes.allNodes(), sshKey)
							FailIfError(err)

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

							if err := verifyNoInternetAccess(nodes.allNodes(), sshKey); err == nil {
								Fail("was able to ping google with outgoing connections blocked")
							}

							// Cleanup old cluster file and create a new one
							By("Recreating kismatic-testing.yaml file")
							err = os.Remove("kismatic-testing.yaml")
							FailIfError(err)
							opts = installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: true,
								modifyHostsFiles:         true,
								dockerRegistryCAPath:     caFile,
								dockerRegistryServer:     fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
								dockerRegistryUsername:   "kismaticuser",
								dockerRegistryPassword:   "kismaticpassword",
							}
							writePlanFile(buildPlan(nodes, opts, sshKey))

							upgradeCluster(true)
						})
					})
				})

				Context("With nodes running Ubuntu 16.04", func() {
					ItOnAWS("should result in an upgraded cluster", func(aws infrastructureProvisioner) {
						distro := Ubuntu1604LTS
						WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Ingress: 1, Storage: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
							// One of the nodes will function as a repo mirror and image registry
							repoNode := nodes.worker[1]
							nodes.worker = nodes.worker[0:1]
							// Standup cluster with previous version
							opts := installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: false, // we want KET to install the packages, so let it use the package repo
								modifyHostsFiles:         true,
							}
							err := installKismatic(nodes, opts, sshKey)
							FailIfError(err)

							extractCurrentKismaticInstaller()

							By("Creating a package repository")
							err = createPackageRepositoryMirror(repoNode, distro, sshKey)
							FailIfError(err, "Error creating local package repo")

							By("Deploying a docker registry")
							caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
							FailIfError(err, "Failed to deploy docker registry")

							By("Seeding the local registry")
							err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
							FailIfError(err, "Error seeding local registry")

							err = disableInternetAccess(nodes.allNodes(), sshKey)
							FailIfError(err)

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

							if err := verifyNoInternetAccess(nodes.allNodes(), sshKey); err == nil {
								Fail("was able to ping google with outgoing connections blocked")
							}

							// Cleanup old cluster file and create a new one
							By("Recreating kismatic-testing.yaml file")
							err = os.Remove("kismatic-testing.yaml")
							FailIfError(err)
							opts = installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: true,
								modifyHostsFiles:         true,
								dockerRegistryCAPath:     caFile,
								dockerRegistryServer:     fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
								dockerRegistryUsername:   "kismaticuser",
								dockerRegistryPassword:   "kismaticpassword",
							}
							writePlanFile(buildPlan(nodes, opts, sshKey))

							upgradeCluster(true)
						})
					})
				})
			})
		})

		Context("From KET version v1.7.0", func() {
			BeforeEach(func() {
				dir := setupTestWorkingDirWithVersion("v1.7.0")
				os.Chdir(dir)
			})

			Context("Using a minikube layout", func() {
				Context("Using CentOS 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey, true)
						})
					})
				})

				Context("Using RedHat 7", func() {
					ItOnAWS("should be upgraded [slow] [upgrade]", func(aws infrastructureProvisioner) {
						WithMiniInfrastructure(RedHat7, aws, func(node NodeDeets, sshKey string) {
							installAndUpgradeMinikube(node, sshKey, true)
						})
					})
				})
			})

			// This spec will be used for testing non-destructive kismatic features on
			// an upgraded cluster.
			// This spec is open to modification when new assertions have to be made.
			Context("Using a skunkworks cluster", func() {
				ItOnAWS("should result in an upgraded cluster [slow] [upgrade]", func(aws infrastructureProvisioner) {
					WithInfrastructureAndDNS(NodeCount{Etcd: 3, Master: 2, Worker: 5, Ingress: 2, Storage: 2}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
						// reserve one of the workers for the add-worker test
						allWorkers := nodes.worker
						nodes.worker = allWorkers[0 : len(nodes.worker)-3]

						// Standup cluster with previous version
						opts := installOptions{adminPassword: "abbazabba"}
						err := installKismatic(nodes, opts, sshKey)
						FailIfError(err)

						// Extract current version of kismatic
						extractCurrentKismaticInstaller()

						// Perform upgrade
						upgradeCluster(true)

						sub := SubDescribe("Using an upgraded cluster")
						defer sub.Check()

						sub.It("should have working storage volumes", func() error {
							return testStatefulWorkload(nodes, sshKey)
						})

						sub.It("should allow adding a worker node", func() error {
							newWorker := allWorkers[len(allWorkers)-1]
							return addNodeToCluster(newWorker, sshKey, []string{}, []string{})
						})

						sub.It("should allow adding a ingress node", func() error {
							newWorker := allWorkers[len(allWorkers)-2]
							return addNodeToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"}, []string{"ingress"})
						})

						sub.It("should allow adding a storage node", func() error {
							newWorker := allWorkers[len(allWorkers)-3]
							return addNodeToCluster(newWorker, sshKey, []string{"com.integrationtest/worker=true"}, []string{"storage"})
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
						sub.It("should still be a highly available cluster after upgrade", func() error {
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

			Context("Using a cluster that has no internet access [slow] [upgrade]", func() {
				Context("With nodes running CentOS 7", func() {
					ItOnAWS("should result in an upgraded cluster", func(aws infrastructureProvisioner) {
						distro := CentOS7
						WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Ingress: 1, Storage: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
							// One of the nodes will function as a repo mirror and image registry
							repoNode := nodes.worker[1]
							nodes.worker = nodes.worker[0:1]
							// Standup cluster with previous version
							opts := installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: false, // we want KET to install the packages, so let it use the package repo
								modifyHostsFiles:         true,
							}
							err := installKismatic(nodes, opts, sshKey)
							FailIfError(err)

							// Extract current version of kismatic
							extractCurrentKismaticInstaller()

							By("Creating a package repository")
							err = createPackageRepositoryMirror(repoNode, distro, sshKey)
							FailIfError(err, "Error creating local package repo")

							By("Deploying a docker registry")
							caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
							FailIfError(err, "Failed to deploy docker registry")

							By("Seeding the local registry")
							err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
							FailIfError(err, "Error seeding local registry")

							err = disableInternetAccess(nodes.allNodes(), sshKey)
							FailIfError(err)

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

							if err := verifyNoInternetAccess(nodes.allNodes(), sshKey); err == nil {
								Fail("was able to ping google with outgoing connections blocked")
							}

							// Cleanup old cluster file and create a new one
							By("Recreating kismatic-testing.yaml file")
							err = os.Remove("kismatic-testing.yaml")
							FailIfError(err)
							opts = installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: true,
								modifyHostsFiles:         true,
								dockerRegistryCAPath:     caFile,
								dockerRegistryServer:     fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
								dockerRegistryUsername:   "kismaticuser",
								dockerRegistryPassword:   "kismaticpassword",
							}
							writePlanFile(buildPlan(nodes, opts, sshKey))

							upgradeCluster(true)
						})
					})
				})

				Context("With nodes running Ubuntu 16.04", func() {
					ItOnAWS("should result in an upgraded cluster", func(aws infrastructureProvisioner) {
						distro := Ubuntu1604LTS
						WithInfrastructure(NodeCount{Etcd: 1, Master: 1, Worker: 2, Ingress: 1, Storage: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
							// One of the nodes will function as a repo mirror and image registry
							repoNode := nodes.worker[1]
							nodes.worker = nodes.worker[0:1]
							// Standup cluster with previous version
							opts := installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: false, // we want KET to install the packages, so let it use the package repo
								modifyHostsFiles:         true,
							}
							err := installKismatic(nodes, opts, sshKey)
							FailIfError(err)

							extractCurrentKismaticInstaller()

							By("Creating a package repository")
							err = createPackageRepositoryMirror(repoNode, distro, sshKey)
							FailIfError(err, "Error creating local package repo")

							By("Deploying a docker registry")
							caFile, err := deployAuthenticatedDockerRegistry(repoNode, dockerRegistryPort, sshKey)
							FailIfError(err, "Failed to deploy docker registry")

							By("Seeding the local registry")
							err = seedRegistry(repoNode, caFile, dockerRegistryPort, sshKey)
							FailIfError(err, "Error seeding local registry")

							err = disableInternetAccess(nodes.allNodes(), sshKey)
							FailIfError(err)

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

							if err := verifyNoInternetAccess(nodes.allNodes(), sshKey); err == nil {
								Fail("was able to ping google with outgoing connections blocked")
							}

							// Cleanup old cluster file and create a new one
							By("Recreating kismatic-testing.yaml file")
							err = os.Remove("kismatic-testing.yaml")
							FailIfError(err)
							opts = installOptions{
								adminPassword:            "abbazabba",
								disconnectedInstallation: true,
								modifyHostsFiles:         true,
								dockerRegistryCAPath:     caFile,
								dockerRegistryServer:     fmt.Sprintf("%s:%d", repoNode.PrivateIP, dockerRegistryPort),
								dockerRegistryUsername:   "kismaticuser",
								dockerRegistryPassword:   "kismaticpassword",
							}
							writePlanFile(buildPlan(nodes, opts, sshKey))

							upgradeCluster(true)
						})
					})
				})
			})
		})
	})
})

func installAndUpgradeMinikube(node NodeDeets, sshKey string, online bool) {
	// Install previous version cluster
	err := installKismaticMini(node, sshKey, "abbazabba")
	FailIfError(err)
	extractCurrentKismaticInstaller()
	upgradeCluster(online)
}

func extractCurrentKismaticInstaller() {
	// Extract current version of kismatic
	pwd, err := os.Getwd()
	FailIfError(err)
	err = extractCurrentKismatic(pwd)
	FailIfError(err)
}
func upgradeCluster(online bool) {
	// Perform upgrade
	cmd := exec.Command("./kismatic", "upgrade", "offline", "-f", "kismatic-testing.yaml")
	if online {
		cmd = exec.Command("./kismatic", "upgrade", "online", "-f", "kismatic-testing.yaml", "--ignore-safety-checks")
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Running diagnostics command")
		// run diagnostics on error
		diagsCmd := exec.Command("./kismatic", "diagnose", "-f", "kismatic-testing.yaml")
		diagsCmd.Stdout = os.Stdout
		diagsCmd.Stderr = os.Stderr
		if errDiags := diagsCmd.Run(); errDiags != nil {
			fmt.Printf("ERROR: error running diagnose command: %v", errDiags)
		}
		FailIfError(err)
	}

	assertClusterVersionIsCurrent()
}
