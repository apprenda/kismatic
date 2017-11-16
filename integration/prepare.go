package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/integration/retry"
	"github.com/apprenda/kismatic/integration/tls"
	"github.com/cloudflare/cfssl/csr"
	. "github.com/onsi/ginkgo"
)

const (
	createKubernetesRepoFileYum = `cat <<EOF > /tmp/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
	https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
`

	createDockerRepoFileYum = `cat <<EOF > /tmp/docker.repo
[docker]
name=Docker
baseurl=https://yum.dockerproject.org/repo/main/centos/7/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://yum.dockerproject.org/gpg
EOF
`

	createGlusterRepoFileYum = `cat <<EOF > /tmp/gluster.repo
[gluster]
name=Gluster
baseurl=http://buildlogs.centos.org/centos/7/storage/x86_64/gluster-3.8/
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://download.gluster.org/pub/gluster/glusterfs/3.8/3.8.7/rsa.pub
EOF`

	moveKubernetesRepoFileYum = `sudo mv /tmp/kubernetes.repo /etc/yum.repos.d`
	moveDockerRepoFileYum     = `sudo mv /tmp/docker.repo /etc/yum.repos.d`
	moveGlusterRepoFileYum    = `sudo mv /tmp/gluster.repo /etc/yum.repos.d`

	installDockerYum          = `sudo yum -y install docker-engine-1.12.6-1.el7.centos`
	installKubeletYum         = `sudo yum -y install kubelet-1.8.3-0`
	installNFSUtilsYum        = `sudo yum -y install nfs-utils` // required for the kubelet
	installKubectlYum         = `sudo yum -y install kubectl-1.8.3-0`
	installGlusterfsServerYum = `sudo yum -y install --nogpgcheck glusterfs-server-3.8.15-2.el7`

	updateAptGet        = `sudo apt-get update`
	addDockerRepoKeyApt = `wget -qO - https://apt.dockerproject.org/gpg | sudo apt-key add -`
	addDockerRepoApt    = `sudo add-apt-repository "deb https://apt.dockerproject.org/repo/ ubuntu-xenial main"`
	installDockerApt    = `sudo apt-get -y install docker-engine=1.12.6-0~ubuntu-xenial`

	addKubernetesRepoKeyApt = `wget -qO - https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -`
	addKubernetesRepoApt    = `sudo add-apt-repository "deb https://packages.cloud.google.com/apt/ kubernetes-xenial main"`
	installKubeletApt       = `sudo apt-get -y install kubelet=1.8.3-00`
	stopKubeletService      = `sudo systemctl stop kubelet`
	installNFSCommonApt     = `sudo apt-get -y install nfs-common`
	installKubectlApt       = `sudo apt-get -y install kubectl=1.8.3-00`

	addGlusterRepoApt         = `sudo add-apt-repository -y ppa:gluster/glusterfs-3.8`
	installGlusterfsServerApt = `sudo apt-get -y install glusterfs-server=3.8.15-ubuntu1~xenial1`
)

type nodePrep struct {
	CommandsToPrepDockerRepo     []string
	CommandsToInstallDocker      []string
	CommandsToPrepKubernetesRepo []string
	CommandsToInstallKubelet     []string
	CommandsToInstallKubectl     []string
	CommandsToInstallGlusterfs   []string
}

var ubuntu1604Prep = nodePrep{
	CommandsToPrepDockerRepo:     []string{addDockerRepoKeyApt, addDockerRepoApt, updateAptGet},
	CommandsToInstallDocker:      []string{installDockerApt},
	CommandsToPrepKubernetesRepo: []string{addKubernetesRepoKeyApt, addKubernetesRepoApt, updateAptGet},
	CommandsToInstallKubelet:     []string{installKubeletApt, installNFSCommonApt, stopKubeletService},
	CommandsToInstallKubectl:     []string{installKubectlApt},
	CommandsToInstallGlusterfs:   []string{addGlusterRepoApt, updateAptGet, installGlusterfsServerApt},
}

var rhel7FamilyPrep = nodePrep{
	CommandsToPrepDockerRepo:     []string{createDockerRepoFileYum, moveDockerRepoFileYum},
	CommandsToInstallDocker:      []string{installDockerYum},
	CommandsToPrepKubernetesRepo: []string{createKubernetesRepoFileYum, moveKubernetesRepoFileYum},
	CommandsToInstallKubelet:     []string{installKubeletYum, installNFSUtilsYum},
	CommandsToInstallKubectl:     []string{installKubectlYum},
	CommandsToInstallGlusterfs:   []string{createGlusterRepoFileYum, moveGlusterRepoFileYum, installGlusterfsServerYum},
}

func InstallKismaticPackages(nodes provisionedNodes, distro linuxDistro, sshKey string, disconnected bool) {
	prep := getPrepForDistro(distro)
	dockerNodes := append(nodes.etcd, nodes.master...)
	dockerNodes = append(dockerNodes, nodes.worker...)
	dockerNodes = append(dockerNodes, nodes.ingress...)
	dockerNodes = append(dockerNodes, nodes.storage...)
	By("Configuring docker repository")
	err := retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToPrepDockerRepo, dockerNodes, sshKey, 5*time.Minute)
	}, 3)
	FailIfError(err, "failed to configure package repository over SSH")

	By("Installing Docker")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallDocker, dockerNodes, sshKey, 10*time.Minute)
	}, 3)
	FailIfError(err, "failed to install docker")

	kubeNodes := append(nodes.master, nodes.worker...)
	kubeNodes = append(kubeNodes, nodes.ingress...)
	kubeNodes = append(kubeNodes, nodes.storage...)

	By("Configuring kubernetes repository")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToPrepKubernetesRepo, kubeNodes, sshKey, 5*time.Minute)
	}, 3)
	FailIfError(err, "failed to configure package repository")

	By("Installing Kubelet package")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallKubelet, kubeNodes, sshKey, 15*time.Minute)
	}, 3)
	FailIfError(err, "failed to install the kubelet package")

	By("Installing Kubectl")
	err = retry.WithBackoff(func() error {
		return runViaSSH(prep.CommandsToInstallKubectl, kubeNodes, sshKey, 10*time.Minute)
	}, 3)
	FailIfError(err, "failed to install the kubectl package")

	if len(nodes.storage) > 0 {
		By("Installing Glusterfs:")
		err = retry.WithBackoff(func() error {
			return runViaSSH(prep.CommandsToInstallGlusterfs, nodes.storage, sshKey, 10*time.Minute)
		}, 3)
		FailIfError(err, "failed to install glustefs")
	}
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

func deployAuthenticatedDockerRegistry(node NodeDeets, listeningPort int, sshKey string) (string, error) {
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

	htpasswdCmd := []string{"mkdir auth", "sudo docker run --entrypoint htpasswd registry -Bbn kismaticuser kismaticpassword > auth/htpasswd"}
	err = runViaSSH(htpasswdCmd, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Failed to create htpasswd file for Docker registry")

	startDockerRegistryCmd := []string{fmt.Sprintf("sudo docker run -d -p %d:5000 --restart=always ", listeningPort) +
		" --name registry" +
		" -v ~/certs:/certs" +
		" -v `pwd`/auth:/auth" +
		" -e \"REGISTRY_AUTH=htpasswd\"" +
		" -e \"REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm\"" +
		" -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" +
		" -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/docker.pem" +
		" -e REGISTRY_HTTP_TLS_KEY=/certs/docker-key.pem registry"}
	err = runViaSSH(startDockerRegistryCmd, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Failed to start docker registry over SSH")

	// Need the full path, otherwise ansible looks for it in the wrong place
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "docker-ca.pem"), nil
}
