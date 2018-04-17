package integration_tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func validateMiniPkgInstallationDisabled(provisioner infrastructureProvisioner, distro linuxDistro) {
	WithMiniInfrastructure(distro, provisioner, func(node NodeDeets, sshKey string) {
		sshUser := node.SSHUser
		if err := ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Adding docker repository")
		prep := getPrepForDistro(distro)
		theNode := []NodeDeets{node}
		err := runViaSSH(prep.CommandsToPrepDockerRepo, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to add docker repository")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Docker")
		err = runViaSSH(prep.CommandsToInstallDocker, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to install docker")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Adding kubernetes repository")
		err = runViaSSH(prep.CommandsToPrepKubernetesRepo, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to add kubernetes repository")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Kubelet")
		err = runViaSSH(prep.CommandsToInstallKubelet, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to install the kubelet package")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Kubectl")
		err = runViaSSH(prep.CommandsToInstallKubectl, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to install the kubectl package")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Glusterfs Server")
		err = runViaSSH(prep.CommandsToInstallGlusterfs, theNode, sshKey, 5*time.Minute)
		FailIfError(err, "failed to install the glusterfs package")

		err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey)
		Expect(err).To(BeNil())
	})
}

// ValidateKismaticMini runs validation against a mini Kubernetes cluster
func ValidateKismaticMini(node NodeDeets, user, sshKey string) PlanAWS {
	By("Building a template")
	plan := PlanAWS{
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		MasterNodeFQDN:      node.Hostname,
		MasterNodeShortName: node.Hostname,
		SSHUser:             user,
		SSHKeyFile:          sshKey,
	}
	writePlanFile(plan)
	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)

	// Run validation
	By("Validate our plan")
	err := runValidate(clusterPath)
	FailIfError(err, "Error validating plan")
	return plan
}

func ValidateKismaticMiniDenyPkgInstallation(node NodeDeets, sshUser, sshKey string) error {
	By("Building a template")
	plan := PlanAWS{
		DisablePackageInstallation: true,
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		Ingress:             []NodeDeets{node},
		Storage:             []NodeDeets{node},
		MasterNodeFQDN:      node.Hostname,
		MasterNodeShortName: node.Hostname,
		SSHUser:             sshUser,
		SSHKeyFile:          sshKey,
	}
	writePlanFile(plan)
	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)

	// Run validation
	By("Validate our plan")
	return runValidate(clusterPath)
}

func ValidateKismaticMiniWithBadSSH(node NodeDeets, user, sshKey string) PlanAWS {
	By("Building a template")
	plan := PlanAWS{
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		MasterNodeFQDN:      node.Hostname,
		MasterNodeShortName: node.Hostname,
		SSHUser:             user,
		SSHKeyFile:          sshKey,
	}
	writePlanFile(plan)
	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)

	// Run validation
	By("Validate our plan")
	err := runValidate(clusterPath)
	FailIfSuccess(err)
	return plan
}

func getBadSSHKeyFile() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	// create empty file
	_, err = os.Create(filepath.Join(dir, ".ssh", "bad.pem"))
	if err != nil {
		return "", fmt.Errorf("Unable to create tag file!")
	}

	return filepath.Join(dir, ".ssh", "bad.pem"), nil
}

func runValidate(planFile string) error {

	cmd := exec.Command("./kismatic", "validate", defaultClusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
