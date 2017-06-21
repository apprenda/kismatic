package integration

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("hosts file modification feature", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("enabling the hosts file modification feature", func() {
		ItOnAWS("should result in a functional cluster [slow]", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{1, 1, 2, 0, 0}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
				By("Setting the hostnames to be different than the actual ones")

				loadBalancedFQDN := nodes.master[0].PublicIP
				nodes.etcd[0].Hostname = "etcd01"
				nodes.master[0].Hostname = "master01"
				nodes.worker[0].Hostname = "worker01"
				nodes.worker[1].Hostname = "worker02"
				// change the hostnames on the machines
				for _, n := range nodes.allNodes() {
					err := runViaSSH([]string{fmt.Sprintf("sudo hostnamectl set-hostname %s", n.Hostname)}, []NodeDeets{n}, sshKey, 1*time.Minute)
					FailIfError(err, "Could not set change firewall")
				}

				plan := PlanAWS{
					Etcd:                nodes.etcd,
					Master:              nodes.master,
					MasterNodeFQDN:      loadBalancedFQDN,
					MasterNodeShortName: loadBalancedFQDN,
					Worker:              nodes.worker[0:1],
					SSHKeyFile:          sshKey,
					SSHUser:             nodes.master[0].SSHUser,
					ModifyHostsFiles:    true,
				}

				By("Installing kismatic with bogus hostnames that are added to hosts files")
				err := installKismaticWithPlan(plan, sshKey)
				FailIfError(err)

				By("Adding a worker with a bogus hostname that is added to hosts files")
				err = addWorkerToCluster(nodes.worker[1])
				FailIfError(err)
			})
		})
	})
})
