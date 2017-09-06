package integration

import (
	"bufio"
	"fmt"
	"html/template"
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
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)
	By("Building a plan to set up an overlay network cluster on this hardware")
	plan := PlanAWS{
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		MasterNodeFQDN:      node.Hostname,
		MasterNodeShortName: node.Hostname,
		SSHUser:             user,
		SSHKeyFile:          sshKey,
	}

	// Create template file
	f, fileErr := os.Create("kismatic-testing.yaml")
	FailIfError(fileErr, "Error waiting for nodes")
	defer f.Close()
	w := bufio.NewWriter(f)
	execErr := template.Execute(w, &plan)
	FailIfError(execErr, "Error filling in plan template")
	w.Flush()

	// Run validation
	By("Validate our plan")
	ver := exec.Command("./kismatic", "install", "validate", "-f", f.Name())
	ver.Stdout = os.Stdout
	ver.Stderr = os.Stderr
	err = ver.Run()
	FailIfError(err, "Error validating plan")
	return plan
}

func ValidateKismaticMiniDenyPkgInstallation(node NodeDeets, sshUser, sshKey string) error {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)
	By("Building a plan to set up an overlay network cluster on this hardware")
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

	// Create template file
	f, fileErr := os.Create("kismatic-testing.yaml")
	FailIfError(fileErr, "Error waiting for nodes")
	defer f.Close()
	w := bufio.NewWriter(f)
	execErr := template.Execute(w, &plan)
	FailIfError(execErr, "Error filling in plan template")
	w.Flush()

	// Run validation
	By("Validate our plan")
	cmd := exec.Command("./kismatic", "install", "validate", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ValidateKismaticMiniWithBadSSH(node NodeDeets, user, sshKey string) PlanAWS {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	log.Printf("Created single node for Kismatic Mini: %s (%s)", node.id, node.PublicIP)
	By("Building a plan to set up an overlay network cluster on this hardware")
	plan := PlanAWS{
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		MasterNodeFQDN:      node.Hostname,
		MasterNodeShortName: node.Hostname,
		SSHUser:             user,
		SSHKeyFile:          sshKey,
	}

	// Create template file
	f, fileErr := os.Create("kismatic-testing.yaml")
	FailIfError(fileErr, "Error waiting for nodes")
	defer f.Close()
	w := bufio.NewWriter(f)
	execErr := template.Execute(w, &plan)
	FailIfError(execErr, "Error filling in plan template")
	w.Flush()

	// Run validation
	By("Validate our plan")
	ver := exec.Command("./kismatic", "install", "validate", "-f", f.Name())
	ver.Stdout = os.Stdout
	ver.Stderr = os.Stderr
	err = ver.Run()
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
