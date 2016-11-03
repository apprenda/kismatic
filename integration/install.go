package integration

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
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
func bailBeforeAnsible() bool {
	return os.Getenv("BAIL_BEFORE_ANSIBLE") != ""
}

func GetSSHKeyFile() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".ssh", "kismatic-integration-testing.pem"), nil
}

func ExtractKismaticToTemp() (string, error) {
	tmpDir, err := ioutil.TempDir("", "kisint")
	if err != nil {
		log.Fatal("Error making temp dir: ", err)
	}
	By(fmt.Sprintf("Extracting Kismatic to temp directory %q", tmpDir))
	cmd := exec.Command("tar", "-zxf", "../out/kismatic.tar.gz", "-C", tmpDir)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error extracting kismatic to temp dir: %v", err)
	}
	return tmpDir, nil
}

type InstallOptions struct {
	AllowPackageInstallation    bool
	AutoConfigureDockerRegistry bool
	DockerRegistryIP            string
	DockerRegistryPort          int
	DockerRegistryCAPath        string
}

func installKismaticMini(node NodeDeets, sshKey string) error {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	By("Building a plan to set up an overlay network cluster on this hardware")
	sshUser := node.SSHUser
	nodes := PlanAWS{
		Etcd:                     []NodeDeets{node},
		Master:                   []NodeDeets{node},
		Worker:                   []NodeDeets{node},
		MasterNodeFQDN:           node.Hostname,
		MasterNodeShortName:      node.Hostname,
		SSHKeyFile:               sshKey,
		SSHUser:                  sshUser,
		AllowPackageInstallation: true,
	}

	By("Writing plan file out to disk")
	f, err := os.Create("kismatic-testing.yaml")
	FailIfError(err, "Error waiting for nodes")
	defer f.Close()
	w := bufio.NewWriter(f)
	err = template.Execute(w, &nodes)
	FailIfError(err, "Error filling in plan template")
	w.Flush()

	By("Validing our plan")
	cmd := exec.Command("./kismatic", "install", "validate", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	FailIfError(err, "Error validating plan")

	By("Punch it Chewie!")
	cmd = exec.Command("./kismatic", "install", "apply", "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InstallKismatic(nodes provisionedNodes, installOpts InstallOptions, sshKey string) error {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	By("Building a plan to set up an overlay network cluster on this hardware")
	sshUser := nodes.master[0].SSHUser
	plan := PlanAWS{
		AllowPackageInstallation: installOpts.AllowPackageInstallation,
		Etcd:                 nodes.etcd,
		Master:               nodes.master,
		Worker:               nodes.worker,
		MasterNodeFQDN:       nodes.master[0].Hostname,
		MasterNodeShortName:  nodes.master[0].Hostname,
		SSHKeyFile:           sshKey,
		SSHUser:              sshUser,
		DockerRegistryCAPath: installOpts.DockerRegistryCAPath,
		DockerRegistryIP:     installOpts.DockerRegistryIP,
		DockerRegistryPort:   installOpts.DockerRegistryPort,
	}

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
	nodes := PlanAWS{
		Etcd:                []NodeDeets{fakeNode},
		Master:              []NodeDeets{fakeNode},
		Worker:              []NodeDeets{fakeNode},
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
	err = template.Execute(w, &nodes)
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

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
