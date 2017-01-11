package integration

import (
	"html/template"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
)

func testNFSShare(aws infrastructureProvisioner, distro linuxDistro) {
	nfsServers, err := aws.CreateNFSServers()
	FailIfError(err, "Couldn't set up NFS shares")

	WithInfrastructure(NodeCount{Worker: 1}, distro, aws, func(nodes provisionedNodes, sshKey string) {
		By("Setting up a plan file with NFS Shares and no storage")
		plan := PlanAWS{
			Etcd:                     nodes.worker,
			Master:                   nodes.worker,
			Worker:                   nodes.worker,
			MasterNodeFQDN:           nodes.worker[0].Hostname,
			MasterNodeShortName:      nodes.worker[0].Hostname,
			AllowPackageInstallation: true,
			SSHKeyFile:               sshKey,
			SSHUser:                  nodes.worker[0].SSHUser,
			NFSVolume: []NFSVolume{
				{Host: nfsServers[0].IpAddress},
			},
		}

		By("Writing plan file out to disk")
		template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
		FailIfError(err, "Couldn't parse template")
		f, err := os.Create("kismatic-testing.yaml")
		FailIfError(err, "Error waiting for nodes")
		defer f.Close()
		err = template.Execute(f, &plan)
		FailIfError(err, "Error filling in plan template")

		By("Punch it Chewie!")
		cmd := exec.Command("./kismatic", "install", "apply", "-f", f.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		FailIfError(err, "Error installing NFS")
	})
}
