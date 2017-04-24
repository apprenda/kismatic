package integration

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	useDirectLVM                bool
	serviceCIDR                 string
	enableNetworkPolicy         bool
}

func installKismaticMini(node NodeDeets, sshKey string) error {
	sshUser := node.SSHUser
	plan := PlanAWS{
		Etcd:                     []NodeDeets{node},
		Master:                   []NodeDeets{node},
		Worker:                   []NodeDeets{node},
		Ingress:                  []NodeDeets{node},
		Storage:                  []NodeDeets{node},
		MasterNodeFQDN:           node.PublicIP,
		MasterNodeShortName:      node.PublicIP,
		SSHKeyFile:               sshKey,
		SSHUser:                  sshUser,
		AllowPackageInstallation: true,
	}
	return installKismaticWithPlan(plan, sshKey)
}

func installKismatic(nodes provisionedNodes, installOpts installOptions, sshKey string) error {
	return installKismaticWithPlan(buildPlan(nodes, installOpts, sshKey), sshKey)
}

func buildPlan(nodes provisionedNodes, installOpts installOptions, sshKey string) PlanAWS {
	sshUser := nodes.master[0].SSHUser
	masterDNS := nodes.master[0].PublicIP
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
		UseDirectLVM:                 installOpts.useDirectLVM,
		ServiceCIDR:                  installOpts.serviceCIDR,
		EnableNetworkPolicy:          installOpts.enableNetworkPolicy,
	}
	return plan
}

func installKismaticWithPlan(plan PlanAWS, sshKey string) error {
	writePlanFile(plan)

	By("Punch it Chewie!")
	cmd := exec.Command("./kismatic", "install", "apply", "-f", "kismatic-testing.yaml", "--skip-preflight")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// run diagnostics on error
		diagsCmd := exec.Command("./kismatic", "diagnose", "-f", "kismatic-testing.yaml")
		diagsCmd.Stdout = os.Stdout
		diagsCmd.Stderr = os.Stderr
		if errDiags := diagsCmd.Run(); errDiags != nil {
			fmt.Printf("ERROR: error running diagnose command: %v", errDiags)
		}
		return err
	}
	return nil
}

func writePlanFile(plan PlanAWS) {
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

func canAccessDashboard(url string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   1000 * time.Millisecond,
		Transport: tr,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Could not create request for ingress via %s, %v", url, err)
	}
	// Access the dashboard a few times to hit all replicas
	for i := 0; i < 3; i++ {
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Could not reach ingress via %s, %v", url, err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("Ingress status code is not 200, got %d vi %s", resp.StatusCode, url)
		}
	}

	return nil
}

func FailIfError(err error, message ...interface{}) {
	Expect(err).ToNot(HaveOccurred(), message...)
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
