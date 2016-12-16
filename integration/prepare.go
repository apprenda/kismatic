package integration

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/cloudflare/cfssl/csr"
	. "github.com/onsi/ginkgo"
)

const (
	copyKismaticYumRepo        = `sudo curl https://kismatic-packages-rpm-test.s3-accelerate.amazonaws.com/kismatic.repo -o /etc/yum.repos.d/kismatic.repo`
	installEtcdYum             = `sudo yum -y install kismatic-etcd`
	installDockerEngineYum     = `sudo yum -y install kismatic-docker-engine`
	installKubernetesMasterYum = `sudo yum -y install kismatic-kubernetes-master`
	installKubernetesYum       = `sudo yum -y install kismatic-kubernetes-node`

	copyKismaticKeyDeb         = `wget -qO - https://kismatic-packages-deb-test.s3-accelerate.amazonaws.com/public.key | sudo apt-key add - `
	copyKismaticRepoDeb        = `sudo add-apt-repository "deb https://kismatic-packages-deb-test.s3-accelerate.amazonaws.com xenial main"`
	updateAptGet               = `sudo apt-get update`
	installEtcdApt             = `sudo apt-get -y install kismatic-etcd`
	installDockerApt           = `sudo apt-get -y install kismatic-docker-engine`
	installKubernetesMasterApt = `sudo apt-get -y install kismatic-kubernetes-master`
	installKubernetesApt       = `sudo apt-get -y install kismatic-kubernetes-node`
)

type nodePrep struct {
	CommandsToPrepRepo         []string
	CommandsToInstallEtcd      []string
	CommandsToInstallDocker    []string
	CommandsToInstallK8sMaster []string
	CommandsToInstallK8s       []string
}

var ubuntu1604Prep = nodePrep{
	CommandsToPrepRepo:         []string{copyKismaticKeyDeb, copyKismaticRepoDeb, updateAptGet},
	CommandsToInstallEtcd:      []string{installEtcdApt},
	CommandsToInstallDocker:    []string{installDockerApt},
	CommandsToInstallK8sMaster: []string{installKubernetesMasterApt},
	CommandsToInstallK8s:       []string{installKubernetesApt},
}

var rhel7FamilyPrep = nodePrep{
	CommandsToPrepRepo:         []string{copyKismaticYumRepo},
	CommandsToInstallEtcd:      []string{installEtcdYum},
	CommandsToInstallDocker:    []string{installDockerEngineYum},
	CommandsToInstallK8sMaster: []string{installKubernetesMasterYum},
	CommandsToInstallK8s:       []string{installKubernetesYum},
}

func ExtractKismaticToTemp() (string, error) {
	tmpDir, err := ioutil.TempDir("", "kisint-dev-")
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

func DownloadKismaticRelease(version string) (string, error) {
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("kisint-%s-", strings.Replace(version, ".", "_", -1)))
	if err != nil {
		log.Fatal("Error making temp dir: ", err)
	}
	By(fmt.Sprintf("Downloading Kismatic %s to temp directory %q", version, tmpDir))
	var url string
	if runtime.GOOS == "darwin" {
		url = fmt.Sprintf("https://github.com/apprenda/kismatic/releases/download/%[1]s/kismatic-%[1]s-darwin-amd64.tar.gz", version)
	} else if runtime.GOOS == "linux" {
		url = fmt.Sprintf("https://github.com/apprenda/kismatic/releases/download/%[1]s/kismatic-%[1]s-linux-amd64.tar.gz", version)
	} else {
		return "", fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
	}
	if err := exec.Command("wget", url, "-O", path.Join(tmpDir, "kismatic-release.tar.gz")).Run(); err != nil {
		return "", err
	}
	if err := exec.Command("tar", "-zxf", path.Join(tmpDir, "kismatic-release.tar.gz"), "-C", tmpDir).Run(); err != nil {
		return "", err
	}
	return tmpDir, nil
}

func InstallKismaticRPMs(nodes provisionedNodes, distro linuxDistro, sshKey string) {
	prep := getPrepForDistro(distro)
	By("Configuring package repository")
	err := runViaSSH(prep.CommandsToPrepRepo, append(append(nodes.etcd, nodes.master...), nodes.worker...), sshKey, 5*time.Minute)
	FailIfError(err, "failed to configure package repository over SSH")

	By("Installing Etcd")
	err = runViaSSH(prep.CommandsToInstallEtcd, nodes.etcd, sshKey, 10*time.Minute)
	FailIfError(err, "failed to install Etcd over SSH")

	By("Installing Docker")
	dockerNodes := append(nodes.master, nodes.worker...)
	err = runViaSSH(prep.CommandsToInstallDocker, dockerNodes, sshKey, 10*time.Minute)
	FailIfError(err, "failed to install docker over SSH")

	By("Installing Master:")
	err = runViaSSH(prep.CommandsToInstallK8sMaster, nodes.master, sshKey, 15*time.Minute)
	FailIfError(err, "failed to install the master over SSH")

	By("Installing Worker:")
	err = runViaSSH(prep.CommandsToInstallK8s, nodes.worker, sshKey, 10*time.Minute)
	FailIfError(err, "failed to install the worker over SSH")
}

func getPrepForDistro(distro linuxDistro) nodePrep {
	switch distro {
	case Ubuntu1604LTS:
		return ubuntu1604Prep
	case CentOS7, RedHat7:
		return rhel7FamilyPrep
	default:
		panic(fmt.Sprintf("Unsupported distro %s", distro))
	}
}

func deployDockerRegistry(node NodeDeets, listeningPort int, sshKey string) (string, error) {
	// Install Docker on the node
	installDockerCmds := []string{
		"sudo curl -sSL https://get.docker.com/ | sh",
		"sudo systemctl start docker",
		"mkdir ~/certs",
	}
	err := runViaSSH(installDockerCmds, []NodeDeets{node}, sshKey, 10*time.Minute)
	FailIfError(err, "Failed to install docker over SSH")
	// Generate CA
	subject := tls.Subject{
		Organization:       "someOrg",
		OrganizationalUnit: "someOrgUnit",
	}
	key, caCert, err := tls.NewCACert("test-resources/ca-csr.json", "someCommonName", subject)
	if err != nil {
		return "", fmt.Errorf("error generating CA cert for Docker: %v", err)
	}
	err = ioutil.WriteFile("docker-ca.pem", caCert, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing CA cert to file")
	}
	// Generate Certificate
	ca := &tls.CA{
		Key:        key,
		Cert:       caCert,
		ConfigFile: "test-resources/ca-config.json",
		Profile:    "kubernetes",
	}
	certHosts := []string{node.Hostname, node.PrivateIP, node.PublicIP}
	req := csr.CertificateRequest{
		CN: node.Hostname,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
		Hosts: certHosts,
		Names: []csr.Name{
			{
				C:  "US",
				L:  "Troy",
				O:  "Kubernetes",
				OU: "Cluster",
				ST: "New York",
			},
		},
	}
	key, cert, err := tls.NewCert(ca, req)
	if err != nil {
		return "", fmt.Errorf("error generating certificate for Docker registry: %v", err)
	}
	if err = ioutil.WriteFile("docker.pem", cert, 0644); err != nil {
		return "", fmt.Errorf("error writing certificate to file: %v", err)
	}
	if err = ioutil.WriteFile("docker-key.pem", key, 0644); err != nil {
		return "", fmt.Errorf("error writing private key to file: %v", err)
	}

	err = copyFileToRemote("docker.pem", "~/certs/docker.pem", node, sshKey, 1*time.Minute)
	FailIfError(err, "failed to copy docker.pem file")
	err = copyFileToRemote("docker-key.pem", "~/certs/docker-key.pem", node, sshKey, 1*time.Minute)
	FailIfError(err, "failed to copy docker-key.pem")

	startDockerRegistryCmd := []string{fmt.Sprintf("sudo docker run -d -p %d:5000 --restart=always ", listeningPort) +
		"--name registry -v ~/certs:/certs -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/docker.pem " +
		"-e REGISTRY_HTTP_TLS_KEY=/certs/docker-key.pem registry"}
	err = runViaSSH(startDockerRegistryCmd, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Failed to start docker registry over SSH")

	// Need the full path, otherwise ansible looks for it in the wrong place
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "docker-ca.pem"), nil
}
