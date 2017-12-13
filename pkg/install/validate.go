package install

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apprenda/kismatic/pkg/validation"

	"github.com/apprenda/kismatic/pkg/ssh"
	"github.com/apprenda/kismatic/pkg/util"
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

// ValidateNodes runs validation against the given node.
// Validates if the details of the nodes are unique.
func ValidateNodes(nodes []Node) (bool, []error) {
	v := newValidator()
	v.validate(nodeList{Nodes: nodes})
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

	warn, err := pki.ValidateClusterCertificates(p)
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
	if p.Cluster.DisconnectedInstallation && !p.PrivateRegistryProvided() {
		v.addError(fmt.Errorf("A container image registry is required when disconnected_installation is true"))
	}

	v.validateWithErrPrefix("Docker", p.Docker)
	v.validate(&p.AddOns)
	v.validate(nodeList{Nodes: p.getAllNodes()})
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
	v.validate(&c.Networking)
	v.validate(&c.Certificates)
	v.validate(&c.SSH)
	v.validate(&c.APIServerOptions)
	v.validate(&c.KubeControllerManagerOptions)
	v.validate(&c.KubeProxyOptions)
	v.validate(&c.KubeSchedulerOptions)
	v.validate(&c.KubeletOptions)
	v.validate(&c.CloudProvider)

	return v.valid()
}

func (n *NetworkConfig) validate() (bool, []error) {
	v := newValidator()
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
	if _, err := time.ParseDuration(c.CAExpiry); c.CAExpiry != "" && err != nil { // don't error when empty for backwards compat
		v.addError(fmt.Errorf("Invalid CA certificate expiry %q provider: %v", c.CAExpiry, err))
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

func (c *CloudProvider) validate() (bool, []error) {
	v := newValidator()
	if c.Provider != "" {
		if !util.Contains(c.Provider, cloudProviders()) {
			v.addError(fmt.Errorf("%q is not a valid cloud provider. Options are %v", c.Provider, cloudProviders()))
		}
		if c.Config != "" {
			if _, err := os.Stat(c.Config); os.IsNotExist(err) {
				v.addError(fmt.Errorf("cloud config file was not found at %q", c.Config))
			}
		}
	}
	return v.valid()
}

func (f *AddOns) validate() (bool, []error) {
	v := newValidator()
	v.validate(f.CNI)
	v.validate(f.HeapsterMonitoring)
	v.validate(&f.PackageManager)
	return v.valid()
}

func (n *CNI) validate() (bool, []error) {
	v := newValidator()
	if n != nil && !n.Disable {
		if !util.Contains(n.Provider, cniProviders()) {
			v.addError(fmt.Errorf("%q is not a valid CNI provider. Options are %v", n.Provider, cniProviders()))
		}
		if n.Provider == "calico" {
			if !util.Contains(n.Options.Calico.Mode, calicoMode()) {
				v.addError(fmt.Errorf("%q is not a valid Calico mode. Options are %v", n.Options.Calico.Mode, calicoMode()))
			}
			if !util.Contains(n.Options.Calico.LogLevel, calicoLogLevel()) {
				v.addError(fmt.Errorf("%q is not a valid Calico log level. Options are %v", n.Options.Calico.LogLevel, calicoLogLevel()))
			}
		}
	}
	return v.valid()
}

func (h *HeapsterMonitoring) validate() (bool, []error) {
	v := newValidator()
	if h != nil && !h.Disable {
		if h.Options.Heapster.Replicas <= 0 {
			v.addError(fmt.Errorf("Heapster replicas %d is not valid, must be greater than 0", h.Options.HeapsterReplicas))
		}
		if !util.Contains(h.Options.Heapster.ServiceType, serviceTypes()) {
			v.addError(fmt.Errorf("Heapster Service Type %q is not a valid option %v", h.Options.Heapster.ServiceType, serviceTypes()))
		}
	}
	return v.valid()
}

func (p *PackageManager) validate() (bool, []error) {
	v := newValidator()
	if !p.Disable {
		if !util.Contains(p.Provider, packageManagerProviders()) {
			v.addError(fmt.Errorf("Package Manager %q is not a valid option %v", p.Provider, packageManagerProviders()))
		}
	}
	return v.valid()
}

// validate SSH access to the nodes
func (s sshConnectionSet) validate() (bool, []error) {
	v := newValidator()

	err := ssh.ValidUnencryptedPrivateKey(s.SSHConfig.Key)
	if err != nil {
		v.addError(fmt.Errorf("SSH key validation error: %v", err))
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

type nodeList struct {
	Nodes []Node
}

func (nl nodeList) validate() (bool, []error) {
	v := newValidator()
	v.addError(validateNoDuplicateNodeInfo(nl.Nodes)...)
	v.addError(validateKubeletOptionsDefinedOnce(nl.Nodes)...)
	return v.valid()
}

func validateNoDuplicateNodeInfo(nodes []Node) []error {
	errs := []error{}
	hostnames := map[string]string{}
	ips := map[string]string{}
	internalIPs := map[string]string{}
	for _, n := range nodes {
		// Validate all hostnames are unique
		if val, ok := hostnames[n.Host]; n.Host != "" && ok && val != n.HashCode() {
			errs = append(errs, fmt.Errorf("Two different nodes cannot have the same hostname %q", n.Host))
		} else if n.Host != "" {
			hostnames[n.Host] = n.HashCode()
		}
		// Validate all IPs are unique
		if val, ok := ips[n.IP]; n.IP != "" && ok && val != n.HashCode() {
			errs = append(errs, fmt.Errorf("Two different nodes cannot have the same IP %q", n.IP))
		} else if n.IP != "" {
			ips[n.IP] = n.HashCode()
		}
		// Validate all internal IPs are unique
		if val, ok := internalIPs[n.InternalIP]; n.InternalIP != "" && ok && val != n.HashCode() {
			errs = append(errs, fmt.Errorf("Two different nodes cannot have the same internal IP %q", n.InternalIP))
		} else if n.InternalIP != "" {
			internalIPs[n.InternalIP] = n.HashCode()
		}
	}
	return errs
}

func validateKubeletOptionsDefinedOnce(nodes []Node) []error {
	errs := []error{}
	seenNodes := map[string]map[string]string{}
	for _, n := range nodes {
		if val, ok := seenNodes[n.HashCode()]; ok && !reflect.DeepEqual(val, n.KubeletOptions.Overrides) {
			errs = append(errs, fmt.Errorf("Cannot redefine kubelet options for node %q", n.Host))
		} else {
			seenNodes[n.HashCode()] = n.KubeletOptions.Overrides
		}
	}
	return errs
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
	for i, n := range ng.Nodes {
		v.validateWithErrPrefix(fmt.Sprintf("Node #%d", i+1), &n)
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
	// validate node labels don't start with 'kismatic/' as that is reserved
	for key, val := range n.Labels {
		if strings.HasPrefix(key, "kismatic/") {
			v.addError(fmt.Errorf("Node label %q cannot start with 'kismatic/'", key))
		}
		errs := validation.IsQualifiedName(key)
		for _, err := range errs {
			v.addError(fmt.Errorf("Node label name %q is not valid %s", key, err))
		}
		errs = validation.IsValidLabelValue(val)
		for _, err := range errs {
			v.addError(fmt.Errorf("Node label %q is not valid %s", val, err))
		}
	}
	return v.valid()
}

func (dr *DockerRegistry) validate() (bool, []error) {
	v := newValidator()
	if (dr.Server == "" && dr.Address == "") && (dr.CAPath != "") {
		v.addError(fmt.Errorf("Docker Registry server cannot be empty when CA is provided"))
	}
	if (dr.Server == "" && dr.Address == "") && (dr.Username != "") {
		v.addError(fmt.Errorf("Docker Registry server cannot be empty when a username is provided"))
	}
	if _, err := os.Stat(dr.CAPath); dr.CAPath != "" && os.IsNotExist(err) {
		v.addError(fmt.Errorf("Docker Registry CA file was not found at %q", dr.CAPath))
	}
	if dr.Username != "" && dr.Password == "" {
		v.addError(fmt.Errorf("Docker Registry password cannot be blank for username %q", dr.Username))
	}
	if dr.Password != "" && dr.Username == "" {
		v.addError(fmt.Errorf("Docker Registry username cannot be blank when a password is provided"))
	}
	return v.valid()
}

func (d Docker) validate() (bool, []error) {
	v := newValidator()
	v.validateWithErrPrefix("Storage", d.Storage)
	return v.valid()
}

func (ds DockerStorage) validate() (bool, []error) {
	v := newValidator()
	v.validateWithErrPrefix("Direct LVM", ds.DirectLVM)
	return v.valid()
}

func (dlvm DockerStorageDirectLVM) validate() (bool, []error) {
	v := newValidator()
	if dlvm.Enabled {
		if dlvm.BlockDevice == "" {
			v.addError(errors.New("DirectLVM is enabled, but no block device was specified"))
		}
		if !filepath.IsAbs(dlvm.BlockDevice) {
			v.addError(errors.New("Path to the block device must be absolute"))
		}
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
	for _, a := range sv.AllowAddresses {
		if ok := validateAllowedAddress(a); !ok {
			v.addError(fmt.Errorf("Invalid address %q in the list of allowed addresses", a))
		}
	}
	reclaimPolicies := []string{"Retain", "Recycle", "Delete"} // API is case-sensitive
	if !util.Contains(sv.ReclaimPolicy, reclaimPolicies) {
		v.addError(fmt.Errorf("%q is not a valid reclaim policy. Valid reclaim policies are: %v", sv.ReclaimPolicy, reclaimPolicies))
	}

	if len(sv.AccessModes) < 1 {
		v.addError(errors.New("Access mode was not provided"))
	}

	accessModes := []string{"ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany"} // API is case-sensitive
	for _, m := range sv.AccessModes {
		if !util.Contains(m, accessModes) {
			v.addError(fmt.Errorf("%q is not a valid access mode. Valid access modes are: %v", m, accessModes))
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
