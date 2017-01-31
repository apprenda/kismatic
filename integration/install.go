package integration

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
)

func leaveIt() bool {
	return os.Getenv("LEAVE_ARTIFACTS") != ""
}

func GetSSHKeyFile() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".ssh", "kismatic-integration-testing.pem"), nil
}

type installOptions struct {
	allowPackageInstallation    bool
	disconnectedInstallation    bool
	autoConfigureDockerRegistry bool
	dockerRegistryIP            string
	dockerRegistryPort          int
	dockerRegistryCAPath        string
	modifyHostsFiles            bool
}

func installKismaticMini(node NodeDeets, sshKey string) error {
	sshUser := node.SSHUser
	plan := PlanAWS{
		Etcd:                     []NodeDeets{node},
		Master:                   []NodeDeets{node},
		Worker:                   []NodeDeets{node},
		Ingress:                  []NodeDeets{node},
		Storage:                  []NodeDeets{node},
		MasterNodeFQDN:           node.Hostname,
		MasterNodeShortName:      node.Hostname,
		SSHKeyFile:               sshKey,
		SSHUser:                  sshUser,
		AllowPackageInstallation: true,
	}
	return installKismaticWithPlan(plan, sshKey)
}

func installKismatic(nodes provisionedNodes, installOpts installOptions, sshKey string) error {
	sshUser := nodes.master[0].SSHUser
	masterDNS := nodes.master[0].Hostname
	if nodes.dnsRecord != nil && nodes.dnsRecord.Name != "" {
		masterDNS = nodes.dnsRecord.Name
	}
	plan := PlanAWS{
		AllowPackageInstallation: installOpts.allowPackageInstallation,
		DisconnectedInstallation: installOpts.disconnectedInstallation,
		Etcd:                nodes.etcd,
		Master:              nodes.master,
		Worker:              nodes.worker,
		Ingress:             nodes.ingress,
		Storage:             nodes.storage,
		MasterNodeFQDN:      masterDNS,
		MasterNodeShortName: masterDNS,
		SSHKeyFile:          sshKey,
		SSHUser:             sshUser,
		AutoConfiguredDockerRegistry: installOpts.autoConfigureDockerRegistry,
		DockerRegistryCAPath:         installOpts.dockerRegistryCAPath,
		DockerRegistryIP:             installOpts.dockerRegistryIP,
		DockerRegistryPort:           installOpts.dockerRegistryPort,
		ModifyHostsFiles:             installOpts.modifyHostsFiles,
	}
	return installKismaticWithPlan(plan, sshKey)
}

func installKismaticWithPlan(plan PlanAWS, sshKey string) error {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	f, err := os.Create("kismatic-testing.yaml")
	FailIfError(err, "Error creating plan")
	defer f.Close()
	w := bufio.NewWriter(f)
	err = template.Execute(w, &plan)
	FailIfError(err, "Error filling in plan template")
	w.Flush()

	By("Punch it Chewie!")
	cmd := exec.Command("./kismatic", "install", "apply", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installKismaticWithABadNode() {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	By("Faking infrastructure")
	fakeNode := NodeDeets{
		id:       "FakeId",
		PublicIP: "10.0.0.0",
		Hostname: "FakeHostname",
	}

	By("Building a plan to set up an overlay network cluster on this hardware")
	sshKey, err := GetSSHKeyFile()
	FailIfError(err, "Error getting SSH Key file")
	plan := PlanAWS{
		Etcd:                []NodeDeets{fakeNode},
		Master:              []NodeDeets{fakeNode},
		Worker:              []NodeDeets{fakeNode},
		Ingress:             []NodeDeets{fakeNode},
		MasterNodeFQDN:      "yep.nope",
		MasterNodeShortName: "yep",
		SSHUser:             "Billy Rubin",
		SSHKeyFile:          sshKey,
	}
	By("Writing plan file out to disk")
	f, err := os.Create("kismatic-testing.yaml")
	FailIfError(err, "Error waiting for nodes")
	defer f.Close()
	w := bufio.NewWriter(f)
	err = template.Execute(w, &plan)
	FailIfError(err, "Error filling in plan template")
	w.Flush()
	f.Close()

	By("Validing our plan")
	cmd := exec.Command("./kismatic", "install", "validate", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		Fail("Validation succeeeded even though it shouldn't have")
	}

	By("Well, try it anyway")
	cmd = exec.Command("./kismatic", "install", "apply", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		Fail("Application succeeeded even though it shouldn't have")
	}
}

func completesInTime(dothis func(), howLong time.Duration) bool {
	c1 := make(chan string, 1)
	go func() {
		dothis()
		c1 <- "completed"
	}()

	select {
	case <-c1:
		return true
	case <-time.After(howLong):
		return false
	}
}

func FailIfError(err error, message ...string) {
	if err != nil {
		log.Printf(message[0]+": %v\n%v", err, message[1:])
		Fail(message[0])
	}
}

func FailIfSuccess(err error) {
	if err == nil {
		Fail("Expected failure")
	}
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
