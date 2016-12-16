package integration

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
)

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
