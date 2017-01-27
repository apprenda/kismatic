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

		By("Prepping nodes for the test")
		prep := getPrepForDistro(distro)
		prepNode := []NodeDeets{node}
		err := runViaSSH(prep.CommandsToPrepRepo, prepNode, sshKey, 5*time.Minute)
		FailIfError(err, "Failed to prep repo on the node")

		By("Installing etcd on the node")
		err = runViaSSH(prep.CommandsToInstallEtcd, prepNode, sshKey, 10*time.Minute)
		FailIfError(err, "Failed to install etcd on the node")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Docker")
		err = runViaSSH(prep.CommandsToInstallDocker, prepNode, sshKey, 10*time.Minute)
		FailIfError(err, "failed to install docker over SSH")

		if err = ValidateKismaticMiniDenyPkgInstallation(node, sshUser, sshKey); err == nil {
			Fail("Missing dependencies, but still passed")
		}

		By("Installing Master")
		err = runViaSSH(prep.CommandsToInstallK8sMaster, prepNode, sshKey, 15*time.Minute)
		FailIfError(err, "Failed to install master on node via SSH")

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
		Etcd:                     []NodeDeets{node},
		Master:                   []NodeDeets{node},
		Worker:                   []NodeDeets{node},
		MasterNodeFQDN:           node.Hostname,
		MasterNodeShortName:      node.Hostname,
		SSHUser:                  user,
		SSHKeyFile:               sshKey,
		AllowPackageInstallation: true,
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
		AllowPackageInstallation: false,
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
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
		Etcd:                     []NodeDeets{node},
		Master:                   []NodeDeets{node},
		Worker:                   []NodeDeets{node},
		MasterNodeFQDN:           node.Hostname,
		MasterNodeShortName:      node.Hostname,
		SSHUser:                  user,
		SSHKeyFile:               sshKey,
		AllowPackageInstallation: true,
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
