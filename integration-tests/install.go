package integration_tests

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
	adminPassword                string
	disablePackageInstallation   bool
	disableDockerInstallation    bool
	disconnectedInstallation     bool
	dockerRegistryServer         string
	dockerRegistryCAPath         string
	dockerRegistryUsername       string
	dockerRegistryPassword       string
	modifyHostsFiles             bool
	httpProxy                    string
	httpsProxy                   string
	noProxy                      string
	dockerStorageDriver          string
	serviceCIDR                  string
	disableCNI                   bool
	cniProvider                  string
	dnsProvider                  string
	heapsterReplicas             int
	heapsterInfluxdbPVC          string
	cloudProvider                string
	kubeAPIServerOptions         map[string]string
	kubeControllerManagerOptions map[string]string
	kubeSchedulerOptions         map[string]string
	kubeProxyOptions             map[string]string
	kubeletOptions               map[string]string
}

func installKismaticMini(node NodeDeets, sshKey string, adminPassword string) error {
	sshUser := node.SSHUser
	plan := PlanAWS{
		AdminPassword:       adminPassword,
		Etcd:                []NodeDeets{node},
		Master:              []NodeDeets{node},
		Worker:              []NodeDeets{node},
		Ingress:             []NodeDeets{node},
		Storage:             []NodeDeets{node},
		MasterNodeFQDN:      node.PublicIP,
		MasterNodeShortName: node.PublicIP,
		SSHKeyFile:          sshKey,
		SSHUser:             sshUser,
	}
	return installKismaticWithPlan(plan)
}

func installKismatic(nodes provisionedNodes, installOpts installOptions, sshKey string) error {
	return installKismaticWithPlan(buildPlan(nodes, installOpts, sshKey))
}

func validateKismatic(nodes provisionedNodes, installOpts installOptions, sshKey string) error {
	return validateKismaticWithPlan(buildPlan(nodes, installOpts, sshKey))
}

func buildPlan(nodes provisionedNodes, installOpts installOptions, sshKey string) PlanAWS {
	sshUser := nodes.master[0].SSHUser
	masterDNS := nodes.master[0].PublicIP
	disableHelm := false
	if nodes.dnsRecord != nil && nodes.dnsRecord.Name != "" {
		masterDNS = nodes.dnsRecord.Name
		// disable helm if using Route53
		disableHelm = true
	}
	plan := PlanAWS{
		AdminPassword:              installOpts.adminPassword,
		DisablePackageInstallation: installOpts.disablePackageInstallation,
		DisableDockerInstallation:  installOpts.disableDockerInstallation,
		DisconnectedInstallation:   installOpts.disconnectedInstallation,
		Etcd:                         nodes.etcd,
		Master:                       nodes.master,
		Worker:                       nodes.worker,
		Ingress:                      nodes.ingress,
		Storage:                      nodes.storage,
		MasterNodeFQDN:               masterDNS,
		MasterNodeShortName:          masterDNS,
		SSHKeyFile:                   sshKey,
		SSHUser:                      sshUser,
		DockerRegistryCAPath:         installOpts.dockerRegistryCAPath,
		DockerRegistryServer:         installOpts.dockerRegistryServer,
		DockerRegistryUsername:       installOpts.dockerRegistryUsername,
		DockerRegistryPassword:       installOpts.dockerRegistryPassword,
		ModifyHostsFiles:             installOpts.modifyHostsFiles,
		HTTPProxy:                    installOpts.httpProxy,
		HTTPSProxy:                   installOpts.httpsProxy,
		NoProxy:                      installOpts.noProxy,
		DockerStorageDriver:          installOpts.dockerStorageDriver,
		ServiceCIDR:                  installOpts.serviceCIDR,
		DisableCNI:                   installOpts.disableCNI,
		CNIProvider:                  installOpts.cniProvider,
		DNSProvider:                  installOpts.dnsProvider,
		DisableHelm:                  disableHelm,
		HeapsterReplicas:             installOpts.heapsterReplicas,
		HeapsterInfluxdbPVC:          installOpts.heapsterInfluxdbPVC,
		CloudProvider:                installOpts.cloudProvider,
		KubeAPIServerOptions:         installOpts.kubeAPIServerOptions,
		KubeControllerManagerOptions: installOpts.kubeControllerManagerOptions,
		KubeSchedulerOptions:         installOpts.kubeSchedulerOptions,
		KubeProxyOptions:             installOpts.kubeProxyOptions,
		KubeletOptions:               installOpts.kubeletOptions,
	}
	return plan
}

func installKismaticWithPlan(plan PlanAWS) error {
	writePlanFile(plan)

	By("Punch it Chewie!")
	cmd := exec.Command("./kismatic", "install", "apply", "-f", "kismatic-testing.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// run diagnostics on error
		fmt.Println("----- Running diagnose command -----")
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

func validateKismaticWithPlan(plan PlanAWS) error {
	writePlanFile(plan)

	By("Validate Plan")
	cmd := exec.Command("./kismatic", "install", "validate", "-f", "kismatic-testing.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writePlanFile(plan PlanAWS) {
	By("Building a template")
	template, err := template.New("planAWSOverlay").Parse(planAWSOverlay)
	FailIfError(err, "Couldn't parse template")

	path := "kismatic-testing.yaml"
	_, err = os.Stat(path)
	// create file if not exists
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		FailIfError(err, "Error creating plan")
		defer f.Close()
		w := bufio.NewWriter(f)
		err = template.Execute(w, &plan)
		FailIfError(err, "Error filling in plan template")
		w.Flush()
	}
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
