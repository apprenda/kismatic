package install

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/apprenda/kismatic/pkg/util"
	garbler "github.com/michaelbironneau/garbler/lib"

	yaml "gopkg.in/yaml.v2"
)

const ket133PackageManagerProvider = "helm"

type stack struct {
	lock sync.Mutex
	s    []string
}

func newStack() *stack {
	return &stack{sync.Mutex{}, make([]string, 0)}
}

func (s *stack) Push(v string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.s = append(s.s, v)
}

func (s *stack) Pop() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.s)
	if l == 0 {
		return "", errors.New("Empty Stack")
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res, nil
}

// PlanReadWriter is capable of reading/writing a Plan
type PlanReadWriter interface {
	Read() (*Plan, error)
	Write(*Plan) error
}

// Planner is used to plan the installation
type Planner interface {
	PlanReadWriter
	PlanExists() bool
}

// FilePlanner is a file-based installation planner
type FilePlanner struct {
	File string
}

// Read the plan from the file system
func (fp *FilePlanner) Read() (*Plan, error) {
	d, err := ioutil.ReadFile(fp.File)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	p := &Plan{}
	if err = yaml.Unmarshal(d, p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %v", err)
	}

	// read deprecated fields and set it the new version of the cluster file
	readDeprecatedFields(p)

	return p, nil
}

func readDeprecatedFields(p *Plan) {
	// only set if not already being set by the user
	// package_manager moved from features: to add_ons: after KET v1.3.3
	if p.Features != nil && p.Features.PackageManager != nil {
		p.AddOns.PackageManager.Disable = !p.Features.PackageManager.Enabled
		// KET v1.3.3 did not have a provider field
		p.AddOns.PackageManager.Provider = ket133PackageManagerProvider
	}
	// allow_package_installation renamed to disable_package_installation after KET v1.4.0
	if p.Cluster.AllowPackageInstallation != nil {
		p.Cluster.DisablePackageInstallation = !*p.Cluster.AllowPackageInstallation
	}
}

var yamlKeyRE = regexp.MustCompile(`[^a-zA-Z]*([a-z_\-A-Z]+)[ ]*:`)

// Write the plan to the file system
func (fp *FilePlanner) Write(p *Plan) error {
	oneTimeComments := commentMap
	bytez, marshalErr := yaml.Marshal(p)
	if marshalErr != nil {
		return fmt.Errorf("error marshalling plan to yaml: %v", marshalErr)
	}

	f, err := os.Create(fp.File)
	if err != nil {
		return fmt.Errorf("error making plan file: %v", err)
	}
	defer f.Close()

	s := newStack()
	scanner := bufio.NewScanner(bytes.NewReader(bytez))
	prevIndent := -1
	for scanner.Scan() {
		text := scanner.Text()
		matched := yamlKeyRE.FindStringSubmatch(text)
		if matched != nil && len(matched) > 1 {
			indent := strings.Count(matched[0], " ") / 2
			if indent <= prevIndent {
				for i := 0; i <= (prevIndent - indent); i++ {
					s.Pop()
				}
			}
			s.Push(matched[1])
			prevIndent = indent
			if thiscomment, ok := oneTimeComments[strings.Join(s.s, ".")]; ok {
				f.WriteString(fmt.Sprintf("%-40s # %s\n", text, thiscomment))
				delete(oneTimeComments, matched[1])
				continue
			}
			if thiscomment, ok := oneTimeComments[matched[1]]; ok {
				f.WriteString(fmt.Sprintf("%-40s # %s\n", text, thiscomment))
				delete(oneTimeComments, matched[1])
				continue
			}
		}
		f.WriteString(text + "\n")
	}

	return nil
}

// PlanExists return true if the plan exists on the file system
func (fp *FilePlanner) PlanExists() bool {
	_, err := os.Stat(fp.File)
	return !os.IsNotExist(err)
}

// WritePlanTemplate writes an installation plan with pre-filled defaults.
func WritePlanTemplate(p *Plan, w PlanReadWriter) error {
	// Set sensible defaults
	p.Cluster.Name = "kubernetes"
	generatedAdminPass, err := generateAlphaNumericPassword()
	if err != nil {
		return fmt.Errorf("error generating random password: %v", err)
	}
	p.Cluster.AdminPassword = generatedAdminPass
	p.Cluster.DisablePackageInstallation = false
	p.Cluster.DisconnectedInstallation = false

	// Set SSH defaults
	p.Cluster.SSH.User = "kismaticuser"
	p.Cluster.SSH.Key = "kismaticuser.key"
	p.Cluster.SSH.Port = 22

	// Set Networking defaults
	p.Cluster.Networking.Type = "overlay"
	p.Cluster.Networking.PodCIDRBlock = "172.16.0.0/16"
	p.Cluster.Networking.ServiceCIDRBlock = "172.20.0.0/16"
	p.Cluster.Networking.UpdateHostsFiles = false

	// Set Certificate defaults
	p.Cluster.Certificates.Expiry = "17520h"

	// Set DockerRegistry defaults
	p.DockerRegistry.Port = 8443

	// Add-Ons
	// Heapster
	p.AddOns.HeapsterMonitoring.Options.HeapsterReplicas = 2
	// Package Manager
	p.AddOns.PackageManager.Provider = "helm"

	// Generate entries for all node types
	n := Node{}
	for i := 0; i < p.Etcd.ExpectedCount; i++ {
		p.Etcd.Nodes = append(p.Etcd.Nodes, n)
	}

	for i := 0; i < p.Master.ExpectedCount; i++ {
		p.Master.Nodes = append(p.Master.Nodes, n)
	}

	for i := 0; i < p.Worker.ExpectedCount; i++ {
		p.Worker.Nodes = append(p.Worker.Nodes, n)
	}

	if p.Ingress.ExpectedCount > 0 {
		for i := 0; i < p.Ingress.ExpectedCount; i++ {
			p.Ingress.Nodes = append(p.Ingress.Nodes, n)
		}
	}

	if p.Storage.ExpectedCount > 0 {
		for i := 0; i < p.Storage.ExpectedCount; i++ {
			p.Storage.Nodes = append(p.Storage.Nodes, n)
		}
	}

	if err := w.Write(p); err != nil {
		return fmt.Errorf("error writing installation plan template: %v", err)
	}
	return nil
}

func getKubernetesServiceIP(p *Plan) (string, error) {
	ip, err := util.GetIPFromCIDR(p.Cluster.Networking.ServiceCIDRBlock, 1)
	if err != nil {
		return "", fmt.Errorf("error getting kubernetes service IP: %v", err)
	}
	return ip.To4().String(), nil
}

func getDNSServiceIP(p *Plan) (string, error) {
	ip, err := util.GetIPFromCIDR(p.Cluster.Networking.ServiceCIDRBlock, 2)
	if err != nil {
		return "", fmt.Errorf("error getting DNS service IP: %v", err)
	}
	return ip.To4().String(), nil
}

func generateAlphaNumericPassword() (string, error) {
	attempts := 0
	for {
		reqs := &garbler.PasswordStrengthRequirements{
			MinimumTotalLength: 16,
			Uppercase:          rand.Intn(6),
			Digits:             rand.Intn(6),
			Punctuation:        -1, // disable punctuation
		}
		pass, err := garbler.NewPassword(reqs)
		if err != nil {
			return "", err
		}
		// validate that the library actually returned an alphanumeric password
		re := regexp.MustCompile("^[a-zA-Z1-9]+$")
		if re.MatchString(pass) {
			return pass, nil
		}
		if attempts == 5 {
			return "", errors.New("failed to generate alphanumeric password")
		}
		attempts++
	}
}

var commentMap = map[string]string{
	"cluster.admin_password":                             "This password is used to login to the Kubernetes Dashboard and can also be used for administration without a security certificate.",
	"cluster.disable_package_installation":               "When true, installation will not occur if any node is missing the correct deb/rpm packages. When false, the installer will attempt to install missing packages for you.",
	"cluster.package_repository_urls":                    "Comma-separated list of URLs of the repositories that should be used during installation. These repositories must contain the kismatic packages and all their transitive dependencies.",
	"cluster.disconnected_installation":                  "Set to true if you have already installed the required packages on the nodes or provided a local URL in package_repository_urls containing those packages.",
	"cluster.disable_registry_seeding":                   "Set to true if you have seeded your registry with the required images for the installation.",
	"cluster.networking.type":                            "overlay or routed. Routed pods can be addressed from outside the Kubernetes cluster; Overlay pods can only address each other.",
	"cluster.networking.pod_cidr_block":                  "Kubernetes will assign pods IPs in this range. Do not use a range that is already in use on your local network!",
	"cluster.networking.service_cidr_block":              "Kubernetes will assign services IPs in this range. Do not use a range that is already in use by your local network or pod network!",
	"cluster.networking.update_hosts_files":              "When true, the installer will add entries for all nodes to other nodes' hosts files. Use when you don't have access to DNS.",
	"cluster.certificates.expiry":                        "Self-signed certificate expiration period in hours; default is 2 years.",
	"cluster.ssh.ssh_key":                                "Absolute path to the ssh private key we should use to manage nodes.",
	"etcd":                                               "Here you will identify all of the nodes that should play the etcd role on your cluster.",
	"master":                                             "Here you will identify all of the nodes that should play the master role.",
	"worker":                                             "Here you will identify all of the nodes that will be workers.",
	"host":                                               "The (short) hostname of a node, e.g. etcd01.",
	"ip":                                                 "The ip address the installer should use to manage this node, e.g. 8.8.8.8.",
	"internalip":                                         "If the node has an IP for internal traffic, enter it here; otherwise leave blank.",
	"master.load_balanced_fqdn":                          "If you have set up load balancing for master nodes, enter the FQDN name here. Otherwise, use the IP address of a single master node.",
	"master.load_balanced_short_name":                    "If you have set up load balancing for master nodes, enter the short name here. Otherwise, use the IP address of a single master node.",
	"docker.storage.direct_lvm":                          "Configure devicemapper in direct-lvm mode (RHEL/CentOS only).",
	"docker.storage.direct_lvm.block_device":             "Path to the block device that will be used for direct-lvm mode. This device will be wiped and used exclusively by docker.",
	"docker.storage.direct_lvm.enable_deferred_deletion": "Set to true if you want to enable deferred deletion when using direct-lvm mode.",
	"docker_registry":                                    "Here you will provide the details of your Docker registry or setup an internal one to run in the cluster. This is optional and the cluster will always have access to the Docker Hub.",
	"docker_registry.setup_internal":                     "When true, a Docker Registry will be installed on top of your cluster and used to host Docker images needed for its installation.",
	"docker_registry.address":                            "IP or hostname for your Docker registry. An internal registry will NOT be setup when this field is provided. Must be accessible from all the nodes in the cluster.",
	"docker_registry.port":                               "Port for your Docker registry.",
	"docker_registry.CA":                                 "Absolute path to the CA that was used when starting your Docker registry. The docker daemons on all nodes in the cluster will be configured with this CA.",
	"nfs":                                                "A set of NFS volumes for use by on-cluster persistent workloads, managed by Kismatic.",
	"nfs.nfs_host":                                       "The host name or ip address of an NFS server.",
	"nfs.mount_path":                                     "The mount path of an NFS share. Must start with /",
	"add_ons.heapster.options.influxdb_pvc_name":         "Provide the name of the persistent volume claim that you will create after installation. If not specified, the data will be stored in ephemeral storage.",
	"add_ons.package_manager.provider":                   "Options: 'helm'",
}
