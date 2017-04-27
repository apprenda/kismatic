package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/integration/retry"
	"github.com/apprenda/kismatic/integration/tls"
	"github.com/cloudflare/cfssl/csr"
	. "github.com/onsi/ginkgo"
)

const (
	copyKismaticYumRepo       = `sudo curl https://kismatic-packages-rpm.s3-accelerate.amazonaws.com/kismatic.repo -o /etc/yum.repos.d/kismatic.repo`
	installCurlYum            = `sudo yum -y install curl`
	installEtcdYum            = `sudo yum -y install etcd-3.1.4-1`
	installTransitionEtcdYum  = `sudo yum -y install transition-etcd`
	installDockerYum          = `sudo yum -y install docker-engine-1.11.2-1.el7.centos`
	installKubeletYum         = `sudo yum -y install kubelet-1.6.2_1-1`
	installKubectlYum         = `sudo yum -y install kubectl-1.6.2_1-1`
	installKismaticOfflineYum = `sudo yum -y install kismatic-offline-1.6.2_1-1`

	copyKismaticKeyDeb        = `wget -qO - https://kismatic-packages-deb.s3-accelerate.amazonaws.com/public.key | sudo apt-key add - `
	copyKismaticRepoDeb       = `sudo add-apt-repository "deb https://kismatic-packages-deb.s3-accelerate.amazonaws.com kismatic-xenial main"`
	updateAptGet              = `sudo apt-get update`
	installCurlApt            = `sudo apt-get -y install curl`
	installEtcdApt            = `sudo apt-get -y install etcd=3.1.4`
	installTransitionEtcdApt  = `sudo apt-get -y install transition-etcd`
	installDockerApt          = `sudo apt-get -y install docker-engine=1.11.2-0~xenial`
	installKubeletApt         = `sudo apt-get -y install kubelet=1.6.2-1`
	installKubectlApt         = `sudo apt-get -y install kubectl=1.6.2-1`
	installKismaticOfflineApt = `sudo apt-get -y install kismatic-offline=1.6.2-1`
)

type nodePrep struct {
	CommandsToPrepRepo         []string
	CommandsToInstallEtcd      []string
	CommandsToInstallDocker    []string
	CommandsToInstallK8sMaster []string
	CommandsToInstallK8s       []string
	CommandsToInstallOffline   []string
}

var ubuntu1604Prep = nodePrep{
	CommandsToPrepRepo:         []string{copyKismaticKeyDeb, copyKismaticRepoDeb, updateAptGet},
	CommandsToInstallEtcd:      []string{installCurlApt, installEtcdApt, installTransitionEtcdApt},
	CommandsToInstallDocker:    []string{installDockerApt},
	CommandsToInstallK8sMaster: []string{installDockerApt, installKubeletApt, installKubectlApt},
	CommandsToInstallK8s:       []string{installDockerApt, installKubeletApt, installKubectlApt},
	CommandsToInstallOffline:   []string{installKismaticOfflineApt},
}

var rhel7FamilyPrep = nodePrep{
	CommandsToPrepRepo:         []string{copyKismaticYumRepo},
	CommandsToInstallEtcd:      []string{installCurlYum, installEtcdYum, installTransitionEtcdYum},
	CommandsToInstallDocker:    []string{installDockerYum},
	CommandsToInstallK8sMaster: []string{installDockerYum, installKubeletYum, installKubectlYum},
	CommandsToInstallK8s:       []string{installDockerYum, installKubeletYum, installKubectlYum},
	CommandsToInstallOffline:   []string{installKismaticOfflineYum},
}

func InstallKismaticPackages(nodes provisionedNodes, distro linuxDistro, sshKey string, disconnected bool) {
	prep := getPrepForDistro(distro)
	By("Configuring package repository")
	err := retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToPrepRepo, append(append(nodes.etcd, nodes.master...), nodes.worker...), sshKey, 5*time.Minute)
	}, 3)
	FailIfError(err, "failed to configure package repository over SSH")

	By("Installing Etcd")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallEtcd, nodes.etcd, sshKey, 10*time.Minute)
	}, 3)
	FailIfError(err, "failed to install Etcd over SSH")

	By("Installing Docker")
	dockerNodes := append(nodes.master, nodes.worker...)
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallDocker, dockerNodes, sshKey, 10*time.Minute)
	}, 3)
	FailIfError(err, "failed to install docker over SSH")

	By("Installing Master:")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallK8sMaster, nodes.master, sshKey, 15*time.Minute)
	}, 3)
	FailIfError(err, "failed to install the master over SSH")

	By("Installing Worker:")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallK8s, nodes.worker, sshKey, 10*time.Minute)
	}, 3)
	FailIfError(err, "failed to install the worker over SSH")

	if disconnected {
		By("Installing Offline:")
		err = retry.WithBackoff(func() error {
			return runViaSSH(prep.CommandsToInstallOffline, []NodeDeets{nodes.master[0]}, sshKey, 10*time.Minute)
		}, 3)
		FailIfError(err, "failed to install the worker over SSH")
	}
}

// RemoveKismaticPackages by running the _packages-cleanup.yaml play
func RemoveKismaticPackages() {
	// Reuse existing play to remove packages
	cmd := exec.Command("./kismatic", "install", "step", "-f", "kismatic-testing.yaml", "_packages-cleanup.yaml")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	FailIfError(err)
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
