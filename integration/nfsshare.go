package integration

import . "github.com/onsi/ginkgo"

func testNFSShare(aws infrastructureProvisioner, distro linuxDistro) {
	nfsServers, err := aws.CreateNFSServers()
	FailIfError(err, "Couldn't set up NFS shares")

	WithMiniInfrastructure(distro, aws, func(node NodeDeets, sshKey string) {
		By("Setting up a plan file with NFS Shares and no storage")
		plan := PlanAWS{
			Etcd:                     []NodeDeets{node},
			Master:                   []NodeDeets{node},
			Worker:                   []NodeDeets{node},
			MasterNodeFQDN:           node.Hostname,
			MasterNodeShortName:      node.Hostname,
			AllowPackageInstallation: true,
			SSHKeyFile:               sshKey,
			SSHUser:                  node.SSHUser,
			NFSVolume: []NFSVolume{
				{Host: nfsServers[0].IpAddress},
			},
		}

		err := installKismaticWithPlan(plan, sshKey)
		FailIfError(err, "Error installing cluster with NFS shares")
	})
}
