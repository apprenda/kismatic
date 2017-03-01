package install

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apprenda/kismatic/pkg/ssh"
)

// TODO: There is need to run validation against anything that is validatable.
// Expose the validatable interface so that it can be consumed when
// validating objects other than a Plan or a Node

// ValidatePlan runs validation against the installation plan to ensure
// that the plan contains valid user input. Returns true, nil if the validation
// is successful. Otherwise, returns false and a collection of validation errors.
func ValidatePlan(p *Plan) (bool, []error) {
	v := newValidator()
	v.validate(p)
	return v.valid()
}

// ValidateNode runs validation against the given node.
func ValidateNode(node *Node) (bool, []error) {
	v := newValidator()
	v.validate(node)
	return v.valid()
}

// ValidatePlanSSHConnections tries to establish SSH connections to all nodes in the cluster
func ValidatePlanSSHConnections(p *Plan) (bool, []error) {
	v := newValidator()

	s := sshConnectionSet{p.Cluster.SSH, p.GetUniqueNodes()}

	v.validateWithErrPrefix("Node Connnection", s)

	return v.valid()
}

type sshConnectionSet struct {
	SSHConfig SSHConfig
	Nodes     []Node
}

// ValidateSSHConnection tries to establish SSH connection with the details provieded for a single node
func ValidateSSHConnection(con *SSHConnection, prefix string) (bool, []error) {
	v := newValidator()
	s := sshConnectionSet{*con.SSHConfig, []Node{*con.Node}}
	v.validateWithErrPrefix(prefix, s)
	return v.valid()
}

// ValidateCertificates checks if certificates exist and are valid
func ValidateCertificates(p *Plan, pki *LocalPKI) (bool, []error) {
	v := newValidator()

	warn, err := pki.ValidateClusterCertificates(p, []string{"admin"})
	if err != nil && len(err) > 0 {
		v.addError(err...)
	}
	if warn != nil && len(warn) > 0 {
		v.addError(warn...)
	}

	return v.valid()
}

// ValidateStorageVolume validates the storage volume attributes
func ValidateStorageVolume(sv StorageVolume) (bool, []error) {
	return sv.validate()
}

type validatable interface {
	validate() (bool, []error)
}

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{
		errs: []error{},
	}
}

func (v *validator) addError(err ...error) {
	v.errs = append(v.errs, err...)
}

func (v *validator) validate(obj validatable) {
	if ok, err := obj.validate(); !ok {
		v.addError(err...)
	}
}

func (v *validator) validateWithErrPrefix(prefix string, objs ...validatable) {
	for _, obj := range objs {
		if ok, err := obj.validate(); !ok {
			newErrs := make([]error, len(err), len(err))
			for i, err := range err {
				newErrs[i] = fmt.Errorf("%s: %v", prefix, err)
			}
			v.addError(newErrs...)
		}
	}
}

func (v *validator) valid() (bool, []error) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}

func (p *Plan) validate() (bool, []error) {
	v := newValidator()

	v.validate(&p.Cluster)
	v.validate(&p.DockerRegistry)
	// on a disconnected_installation a registry must be provided
	v.validate(disconnectedInstallation{cluster: p.Cluster, registryProvided: p.DockerRegistryProvided()})
	v.validateWithErrPrefix("Etcd nodes", &p.Etcd)
	v.validateWithErrPrefix("Master nodes", &p.Master)
	v.validateWithErrPrefix("Worker nodes", &p.Worker)
	v.validateWithErrPrefix("Ingress nodes", &p.Ingress)
	v.validate(&p.NFS)
	v.validateWithErrPrefix("Storage nodes", &p.Storage)

	return v.valid()
}

func (c *Cluster) validate() (bool, []error) {
	v := newValidator()
	if c.Name == "" {
		v.addError(errors.New("Cluster name cannot be empty"))
	}
	if c.AdminPassword == "" {
		v.addError(errors.New("Admin password cannot be empty"))
	}
	v.validate(&c.Networking)
	v.validate(&c.Certificates)
	v.validate(&c.SSH)

	return v.valid()
}

func (n *NetworkConfig) validate() (bool, []error) {
	v := newValidator()
	if n.Type == "" {
		v.addError(errors.New("Networking type cannot be empty"))
	}
	if n.Type != "routed" && n.Type != "overlay" {
		v.addError(fmt.Errorf("Invalid networking type %q was provided", n.Type))
	}
	if n.PodCIDRBlock == "" {
		v.addError(errors.New("Pod CIDR block cannot be empty"))
	}
	if _, _, err := net.ParseCIDR(n.PodCIDRBlock); n.PodCIDRBlock != "" && err != nil {
		v.addError(fmt.Errorf("Invalid Pod CIDR block provided: %v", err))
	}

	if n.ServiceCIDRBlock == "" {
		v.addError(errors.New("Service CIDR block cannot be empty"))
	}
	if _, _, err := net.ParseCIDR(n.ServiceCIDRBlock); n.ServiceCIDRBlock != "" && err != nil {
		v.addError(fmt.Errorf("Invalid Service CIDR block provided: %v", err))
	}
	return v.valid()
}

func (c *CertsConfig) validate() (bool, []error) {
	v := newValidator()
	if _, err := time.ParseDuration(c.Expiry); err != nil {
		v.addError(fmt.Errorf("Invalid certificate expiry %q provided: %v", c.Expiry, err))
	}
	return v.valid()
}

func (s *SSHConfig) validate() (bool, []error) {
	v := newValidator()
	if s.User == "" {
		v.addError(errors.New("SSH user field is required"))
	}
	if s.Key == "" {
		v.addError(errors.New("SSH key field is required"))
	}
	if _, err := os.Stat(s.Key); os.IsNotExist(err) {
		v.addError(fmt.Errorf("SSH Key file was not found at %q", s.Key))
	}
	if !filepath.IsAbs(s.Key) {
		v.addError(errors.New("SSH Key field must be an absolute path"))
	}
	if s.Port < 1 || s.Port > 65535 {
		v.addError(fmt.Errorf("SSH port %d is invalid. Port must be in the range 1-65535", s.Port))
	}
	return v.valid()
}

// validate SSH access to the nodes
func (s sshConnectionSet) validate() (bool, []error) {
	v := newValidator()

	err := ssh.ValidUnencryptedPrivateKey(s.SSHConfig.Key)
	if err != nil {
		v.addError(fmt.Errorf("error parsing SSH key: %v", err))
	} else {
		var wg sync.WaitGroup
		errQueue := make(chan error, len(s.Nodes))
		// number of nodes
		wg.Add(len(s.Nodes))
		for _, node := range s.Nodes {
			go func(ip string) {
				defer wg.Done()
				sshErr := ssh.TestConnection(ip, s.SSHConfig.Port, s.SSHConfig.User, s.SSHConfig.Key)
				// Need to send something the buffered channel
				if sshErr != nil {
					errQueue <- fmt.Errorf("SSH connectivity validation failed for %q: %v", ip, sshErr)
				} else {
					errQueue <- nil
				}
			}(node.IP)
		}

		// Wait for all nodes to complete, then close channel
		go func() {
			wg.Wait()
			close(errQueue)
		}()

		// Read any error
		for err := range errQueue {
			if err != nil {
				v.addError(err)
			}
		}
	}

	return v.valid()
}

func (ng *NodeGroup) validate() (bool, []error) {
	v := newValidator()
	if ng == nil || len(ng.Nodes) <= 0 {
		v.addError(fmt.Errorf("At least one node is required"))
	}
	if ng.ExpectedCount <= 0 {
		v.addError(fmt.Errorf("Node count must be greater than 0"))
	}
	if len(ng.Nodes) != ng.ExpectedCount && (len(ng.Nodes) > 0 && ng.ExpectedCount > 0) {
		v.addError(fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", ng.ExpectedCount, len(ng.Nodes)))
	}
	hostnames := map[string]int{}
	ips := map[string]int{}
	internalIPs := map[string]int{}
	for i, n := range ng.Nodes {
		v.validateWithErrPrefix(fmt.Sprintf("Node #%d", i+1), &n)
		// Validate all hostnames are unique
		if nodeID, ok := hostnames[n.Host]; ok && n.Host != "" {
			v.addError(fmt.Errorf("Node #%d has the same hostname %q as node #%d", i+1, n.Host, nodeID))
		} else if n.Host != "" {
			hostnames[n.Host] = i + 1
		}
		// Validate all IPs are unique
		if nodeID, ok := ips[n.IP]; ok && n.IP != "" {
			v.addError(fmt.Errorf("Node #%d has the same IP %q as node #%d", i+1, n.IP, nodeID))
		} else if n.IP != "" {
			ips[n.IP] = i + 1
		}
		// Validate all internal IPs are unique
		if nodeID, found := internalIPs[n.InternalIP]; found && n.InternalIP != "" {
			v.addError(fmt.Errorf("Node #%d has the same internal IP %q as node #%d", i+1, n.InternalIP, nodeID))
		} else if n.InternalIP != "" {
			internalIPs[n.InternalIP] = i + 1
		}
	}

	return v.valid()
}

// In order to make this node group optional, we consider it to be valid if:
// - it's nil
// - the number of nodes is zero, and the expected count is zero
// We eagerly test the mismatch between given and expected node counts
// because otherwise the regular NodeGroup validation returns confusing errors.
func (ong *OptionalNodeGroup) validate() (bool, []error) {
	if ong == nil {
		return true, nil
	}
	if len(ong.Nodes) == 0 && ong.ExpectedCount == 0 {
		return true, nil
	}
	if len(ong.Nodes) != ong.ExpectedCount {
		return false, []error{fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", ong.ExpectedCount, len(ong.Nodes))}
	}
	ng := NodeGroup(*ong)
	return ng.validate()
}

func (mng *MasterNodeGroup) validate() (bool, []error) {
	v := newValidator()

	if len(mng.Nodes) <= 0 {
		v.addError(fmt.Errorf("At least one node is required"))
	}
	if mng.ExpectedCount <= 0 {
		v.addError(fmt.Errorf("Node count must be greater than 0"))
	}
	if len(mng.Nodes) != mng.ExpectedCount && (len(mng.Nodes) > 0 && mng.ExpectedCount > 0) {
		v.addError(fmt.Errorf("Expected node count (%d) does not match the number of nodes provided (%d)", mng.ExpectedCount, len(mng.Nodes)))
	}
	for i, n := range mng.Nodes {
		v.validateWithErrPrefix(fmt.Sprintf("Node #%d", i+1), &n)
	}

	if mng.LoadBalancedFQDN == "" {
		v.addError(fmt.Errorf("Load balanced FQDN is required"))
	}

	if mng.LoadBalancedShortName == "" {
		v.addError(fmt.Errorf("Load balanced shortname is required"))
	}

	return v.valid()
}

func (n *Node) validate() (bool, []error) {
	v := newValidator()
	if n.Host == "" {
		v.addError(fmt.Errorf("Node host field is required"))
	}
	if n.IP == "" {
		v.addError(fmt.Errorf("Node IP field is required"))
	}
	if ip := net.ParseIP(n.IP); ip == nil && n.IP != "" {
		v.addError(fmt.Errorf("Invalid IP provided"))
	}
	if ip := net.ParseIP(n.InternalIP); n.InternalIP != "" && ip == nil {
		v.addError(fmt.Errorf("Invalid InternalIP provided"))
	}
	return v.valid()
}

func (dr *DockerRegistry) validate() (bool, []error) {
	v := newValidator()
	if dr.SetupInternal == true && (dr.Address != "" || dr.CAPath != "") {
		v.addError(fmt.Errorf("Cannot setup internal registry when DockerRegistry address or CA is provided"))
	}
	if dr.Address == "" && (dr.CAPath != "") {
		v.addError(fmt.Errorf("Docker Registry address cannot be empty when CA is provided"))
	}
	if dr.Address != "" && (dr.Port < 1 || dr.Port > 65535) {
		v.addError(fmt.Errorf("Docker Registry port %d is invalid. Port must be in the range 1-65535", dr.Port))
	}
	if dr.Address != "" && (dr.CAPath == "") {
		v.addError(fmt.Errorf("Docker Registry CA cannot be empty when registry address is provided"))
	}
	if _, err := os.Stat(dr.CAPath); dr.CAPath != "" && os.IsNotExist(err) {
		v.addError(fmt.Errorf("Docker Registry CA file was not found at %q", dr.CAPath))
	}
	return v.valid()
}

func (nfs *NFS) validate() (bool, []error) {
	v := newValidator()
	uniqueVolumes := make(map[NFSVolume]bool)
	for _, vol := range nfs.Volumes {
		v.validate(vol)
		if _, ok := uniqueVolumes[vol]; ok {
			v.addError(fmt.Errorf("Duplicate NFS volume %v", vol))
		} else {
			uniqueVolumes[vol] = true
		}
	}
	return v.valid()
}

func (nfsVol NFSVolume) validate() (bool, []error) {
	v := newValidator()
	if nfsVol.Host == "" {
		v.addError(errors.New("NFS volume host cannot be empty"))
	}
	if nfsVol.Path == "" {
		v.addError(errors.New("NFS volume path cannot be empty"))
	}
	if len(nfsVol.Path) > 0 && nfsVol.Path[0] != '/' {
		v.addError(errors.New("NFS volume path must be absolute"))
	}
	return v.valid()
}

func (sv StorageVolume) validate() (bool, []error) {
	v := newValidator()
	notAllowed := ": / \\ & < > |"
	if strings.ContainsAny(sv.Name, notAllowed) {
		v.addError(fmt.Errorf("Volume name may not contain spaces or any of the following characters: %q", notAllowed))
	}
	if sv.SizeGB < 1 {
		v.addError(errors.New("Volume size must be 1GB or larger"))
	}
	if sv.DistributionCount < 1 {
		v.addError(errors.New("Distribution count must be greater than zero"))
	}
	if sv.ReplicateCount < 1 {
		v.addError(errors.New("Replication count must be greater than zero"))
	}
	if sv.ReplicateCount == 2 && sv.DistributionCount == 2 {
		bs, _ := base64.StdEncoding.DecodeString("ICAgICAgIC4tIiItLg0KICAgICAgL1tdIF8gX1wNCiAgICAgX3xfb19MSUl8Xw0KICAgIC8gfCA9PT09IHwgXA0KICAgIHxffCA9PT09IHxffA0KICAgICB8fCIgfHwgIHx8DQogICAgIHx8TEkgIG8gfHwNCiAgICAgfHwnLS0tLSd8fA0KICAgIC9fX3wgICAgfF9fXA0K")
		os.Stdout.Write(bs)
	}
	for _, a := range sv.AllowAddresses {
		if ok := validateAllowedAddress(a); !ok {
			v.addError(fmt.Errorf("Invalid address %q in the list of allowed addresses", a))
		}
	}
	return v.valid()
}

func validateAllowedAddress(address string) bool {
	// First, validate that there are four octets with 1, 2 or 3 chars, separated by dots
	r := regexp.MustCompile(`^[0-9*]{1,3}\.[0-9*]{1,3}\.[0-9*]{1,3}\.[0-9*]{1,3}$`)
	if !r.MatchString(address) {
		return false
	}
	// Validate each octet on its own
	oct := strings.Split(address, ".")
	for _, o := range oct {
		// Valid if the octet is a wildcard, or if it's a number between 0-255 (inclusive)
		n, err := strconv.Atoi(o)
		valid := o == "*" || (err == nil && 0 <= n && n <= 255)
		if !valid {
			return false
		}
	}
	return true
}

type disconnectedInstallation struct {
	cluster          Cluster
	registryProvided bool
}

func (l disconnectedInstallation) validate() (bool, []error) {
	v := newValidator()
	if l.cluster.DisconnectedInstallation && !l.registryProvided {
		v.addError(fmt.Errorf("A docker registry is required when disconnected_installation is true"))
	}
	return v.valid()
}
