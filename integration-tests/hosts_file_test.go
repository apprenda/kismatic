package integration_tests

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
			WithInfrastructure(NodeCount{1, 1, 4, 0, 0}, Ubuntu1604LTS, aws, func(nodes provisionedNodes, sshKey string) {
				By("Setting the hostnames to be different than the actual ones")

				// test hostname feature and unusual hostname formats
				loadBalancedFQDN := nodes.master[0].PublicIP
				nodes.etcd[0].Hostname = "etcd01"
				nodes.master[0].Hostname = "MASTER01"
				nodes.worker[0].Hostname = "WORKER01"
				nodes.worker[1].Hostname = "worker02.test"
				nodes.worker[2].Hostname = "Worker-03"
				nodes.worker[3].Hostname = "WORKER04"
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
					Worker:              nodes.worker[0:3],
					Ingress:             nodes.worker[0:3],
					Storage:             nodes.worker[0:3],
					SSHKeyFile:          sshKey,
					SSHUser:             nodes.master[0].SSHUser,
					ModifyHostsFiles:    true,
				}

				By("Installing kismatic with bogus hostnames that are added to hosts files")
				err := installKismaticWithPlan(plan)
				FailIfError(err)

				By("Adding a worker with a bogus hostname that is added to hosts files")
				err = addWorkerToCluster(nodes.worker[3], sshKey, []string{})
				FailIfError(err)
			})
		})
	})
})
