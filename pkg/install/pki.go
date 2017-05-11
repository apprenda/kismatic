package install

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/cloudflare/cfssl/csr"
)

const (
	adminUser                    = "admin"
	adminGroup                   = "system:masters"
	dockerRegistryCertFilename   = "docker"
	serviceAccountCertFilename   = "service-account"
	serviceAccountCertCommonName = "kube-service-account"
)

// The PKI provides a way for generating certificates for the cluster described by the Plan
type PKI interface {
	CertificateAuthorityExists() (bool, error)
	NodeCertificateExists(node Node) (bool, error)
	GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error
	GetClusterCA() (*tls.CA, error)
	GenerateClusterCA(p *Plan) (*tls.CA, error)
	GenerateClusterCertificates(p *Plan, ca *tls.CA) error
}

// LocalPKI is a file-based PKI
type LocalPKI struct {
	CACsr                   string
	CAConfigFile            string
	CASigningProfile        string
	GeneratedCertsDirectory string
	Log                     io.Writer
}

// CertificateAuthorityExists returns true if the CA for the cluster exists
func (lp *LocalPKI) CertificateAuthorityExists() (bool, error) {
	return tls.CertKeyPairExists("ca", lp.GeneratedCertsDirectory)
}

// NodeCertificateExists returns true if the node's key and certificate exist
func (lp *LocalPKI) NodeCertificateExists(node Node) (bool, error) {
	return tls.CertKeyPairExists(node.Host, lp.GeneratedCertsDirectory)
}

// GetClusterCA returns the cluster CA
func (lp *LocalPKI) GetClusterCA() (*tls.CA, error) {
	ca := &tls.CA{
		ConfigFile: lp.CAConfigFile,
		Profile:    lp.CASigningProfile,
	}
	key, cert, err := tls.ReadCACert("ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate/key: %v", err)
	}
	ca.Cert = cert
	ca.Key = key
	return ca, nil
}

// GenerateClusterCA creates a Certificate Authority for the cluster
func (lp *LocalPKI) GenerateClusterCA(p *Plan) (*tls.CA, error) {
	ca := &tls.CA{
		ConfigFile: lp.CAConfigFile,
		Profile:    lp.CASigningProfile,
	}
	exists, err := tls.CertKeyPairExists("ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error verifying CA certificate/key: %v", err)
	}
	if exists {
		return lp.GetClusterCA()
	}

	// CA keypair doesn't exist, generate one
	util.PrettyPrintOk(lp.Log, "Generating cluster Certificate Authority")
	key, cert, err := tls.NewCACert(lp.CACsr, p.Cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA Cert: %v", err)
	}
	if err = tls.WriteCert(key, cert, "ca", lp.GeneratedCertsDirectory); err != nil {
		return nil, fmt.Errorf("error writing CA files: %v", err)
	}
	ca.Cert = cert
	ca.Key = key
	return ca, nil
}

// GenerateClusterCertificates creates all certificates required for the cluster
// described in the plan file.
func (lp *LocalPKI) GenerateClusterCertificates(p *Plan, ca *tls.CA) error {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}
	nodes := p.getAllNodes()
	seenNodes := map[string]bool{}
	for _, n := range nodes {
		// Only generate certs once for each node, nodes can be in more than one group
		if _, ok := seenNodes[n.Host]; ok {
			continue
		}
		seenNodes[n.Host] = true
		if err := lp.GenerateNodeCertificate(p, n, ca); err != nil {
			return err
		}
	}
	// Create certs for docker registry if it's missing
	if p.DockerRegistry.SetupInternal {
		if err := lp.generateDockerRegistryCert(p, ca); err != nil {
			return err
		}
	}
	// Create key for service account signing
	if err := lp.generateServiceAccountCert(p, ca); err != nil {
		return err
	}
	// Create the admin user's certificate
	if err := lp.generateUserCert(p, ca, adminUser, []string{adminGroup}); err != nil {
		return err
	}
	return nil
}

// ValidateClusterCertificates validates all certificates in the cluster
func (lp *LocalPKI) ValidateClusterCertificates(p *Plan) (warn []error, err []error) {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}
	// Validate node certificates
	nodes := p.getAllNodes()
	seenNodes := map[string]bool{}
	for _, n := range nodes {
		// Only generate certs once for each node, nodes can be in more than one group
		if _, ok := seenNodes[n.Host]; ok {
			continue
		}
		seenNodes[n.Host] = true
		_, nodeWarn, nodeErr := lp.validateNodeCertificate(p, n)
		warn = append(warn, nodeWarn...)
		if err != nil {
			err = append(err, nodeErr)
		}
	}
	// Validate docker registry cert
	if p.DockerRegistry.SetupInternal {
		_, dockerWarn, dockerErr := lp.validateDockerRegistryCert(p)
		warn = append(warn, dockerWarn...)
		if err != nil {
			err = append(err, dockerErr)
		}
	}
	// Validate service account certificate
	_, saWarn, saErr := lp.validateServiceAccountCert()
	warn = append(warn, saWarn...)
	if err != nil {
		err = append(err, saErr)
	}
	// Validate admin certificate
	_, userWarn, userErr := lp.validateUserCert(adminUser, []string{adminGroup})
	warn = append(warn, userWarn...)
	if err != nil {
		err = append(err, userErr)
	}
	return warn, err
}

// GenerateNodeCertificate creates a private key and certificate for the given node
func (lp *LocalPKI) GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error {
	commonName := node.Host
	// Build list of SANs
	clusterSANs, err := clusterCertsSubjectAlternateNames(plan)
	if err != nil {
		return err
	}
	nodeSANs := append(clusterSANs, node.Host, node.IP)
	if node.InternalIP != "" {
		nodeSANs = append(nodeSANs, node.InternalIP)
	}
	if isMasterNode(*plan, node) {
		if plan.Master.LoadBalancedFQDN != "" {
			nodeSANs = append(nodeSANs, plan.Master.LoadBalancedFQDN)
		}
		if plan.Master.LoadBalancedShortName != "" {
			nodeSANs = append(nodeSANs, plan.Master.LoadBalancedShortName)
		}
	}

	// Don't generate if the key pair exists and valid
	valid, warn, err := tls.CertExistsAndValid(commonName, nodeSANs, []string{}, node.Host, lp.GeneratedCertsDirectory)
	if err != nil {
		return err
	}
	if warn != nil && len(warn) > 0 {
		util.PrettyPrintErr(lp.Log, "Found key and certificate for node %q but it is not valid", node.Host)
		util.PrintValidationErrors(lp.Log, warn)
		return fmt.Errorf("error verifying certificates for node %q", node.Host)
	}
	if valid {
		util.PrettyPrintOk(lp.Log, "Found valid key and certificate for node %q", node.Host)
		return nil
	}

	util.PrettyPrintOk(lp.Log, "Generating certificates for host %q", node.Host)

	key, cert, err := generateCert(ca, commonName, nodeSANs)
	if err != nil {
		return fmt.Errorf("error during cluster cert generation: %v", err)
	}
	err = tls.WriteCert(key, cert, node.Host, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for host %q: %v", node.Host, err)
	}
	return nil
}

func (lp *LocalPKI) validateNodeCertificate(p *Plan, node Node) (valid bool, warn []error, err error) {
	CN := node.Host
	// Build list of SANs
	clusterSANs, err := clusterCertsSubjectAlternateNames(p)
	if err != nil {
		return false, nil, err
	}
	nodeSANs := append(clusterSANs, node.Host, node.IP)
	if node.InternalIP != "" {
		nodeSANs = append(nodeSANs, node.InternalIP)
	}
	if isMasterNode(*p, node) {
		if p.Master.LoadBalancedFQDN != "" {
			nodeSANs = append(nodeSANs, p.Master.LoadBalancedFQDN)
		}
		if p.Master.LoadBalancedShortName != "" {
			nodeSANs = append(nodeSANs, p.Master.LoadBalancedShortName)
		}
	}

	return tls.CertExistsAndValid(CN, nodeSANs, []string{}, node.Host, lp.GeneratedCertsDirectory)
}

func (lp *LocalPKI) generateDockerRegistryCert(p *Plan, ca *tls.CA) error {
	// Default registry will be deployed on the first master
	n := p.Master.Nodes[0]
	commonName := n.Host
	SANs := []string{n.Host, n.IP}
	if n.InternalIP != "" {
		SANs = append(SANs, n.InternalIP)
	}

	// Don't generate if the key pair exists and valid
	valid, warn, err := tls.CertExistsAndValid(commonName, SANs, []string{}, dockerRegistryCertFilename, lp.GeneratedCertsDirectory)
	if err != nil {
		return err
	}
	if warn != nil && len(warn) > 0 {
		util.PrettyPrintErr(lp.Log, "Found key and certificate for docker registry but it is not valid")
		util.PrintValidationErrors(lp.Log, warn)
		return fmt.Errorf("error verifying certificates for docker registry")
	}
	if valid {
		util.PrettyPrintOk(lp.Log, "Found certificate for docker registry")
		return nil
	}

	util.PrettyPrintOk(lp.Log, "Generating certificates for docker registry")

	dockerKey, dockerCert, err := generateCert(ca, commonName, SANs)
	if err != nil {
		return fmt.Errorf("error during user cert generation: %v", err)
	}
	err = tls.WriteCert(dockerKey, dockerCert, dockerRegistryCertFilename, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for docker registry")
	}
	return nil
}

func (lp *LocalPKI) validateDockerRegistryCert(p *Plan) (valid bool, warn []error, err error) {
	// Default registry will be deployed on the first master
	n := p.Master.Nodes[0]
	CN := n.Host
	SANs := []string{n.Host, n.IP}
	if n.InternalIP != "" {
		SANs = append(SANs, n.InternalIP)
	}

	return tls.CertExistsAndValid(CN, SANs, []string{}, dockerRegistryCertFilename, lp.GeneratedCertsDirectory)
}

func (lp *LocalPKI) generateServiceAccountCert(p *Plan, ca *tls.CA) error {
	SANs := []string{}
	// Don't generate if the key pair exists and valid
	valid, warn, err := tls.CertExistsAndValid(serviceAccountCertCommonName, SANs, []string{}, serviceAccountCertFilename, lp.GeneratedCertsDirectory)
	if err != nil {
		return err
	}
	if warn != nil && len(warn) > 0 {
		util.PrettyPrintErr(lp.Log, "Found key and certificate for service account but it is not valid")
		util.PrintValidationErrors(lp.Log, warn)
		return fmt.Errorf("error verifying certificates for service account")
	}
	if valid {
		util.PrettyPrintOk(lp.Log, "Found key and certificate for service accounts")
		return nil
	}
	util.PrettyPrintOk(lp.Log, "Generating certificates for service accounts")

	key, cert, err := generateCert(ca, serviceAccountCertCommonName, SANs)
	if err != nil {
		return fmt.Errorf("error generating service account certs: %v", err)
	}
	if err = tls.WriteCert(key, cert, serviceAccountCertFilename, lp.GeneratedCertsDirectory); err != nil {
		return fmt.Errorf("error writing generated service account cert: %v", err)
	}
	return nil
}

func (lp *LocalPKI) validateServiceAccountCert() (valid bool, warn []error, err error) {
	SANs := []string{}
	return tls.CertExistsAndValid(serviceAccountCertCommonName, SANs, []string{}, serviceAccountCertFilename, lp.GeneratedCertsDirectory)
}

func (lp *LocalPKI) generateUserCert(p *Plan, ca *tls.CA, user string, groups []string) error {
	SANs := []string{user}

	// Don't generate if the key pair exists and valid
	valid, warn, err := tls.CertExistsAndValid(user, SANs, groups, user, lp.GeneratedCertsDirectory)
	if err != nil {
		return err
	}
	if warn != nil && len(warn) > 0 {
		util.PrettyPrintErr(lp.Log, "Found key and certificate for user %q but it is not valid", user)
		util.PrintValidationErrors(lp.Log, warn)
		return fmt.Errorf("error verifying certificates for user %q", user)
	}
	if valid {
		util.PrettyPrintOk(lp.Log, "Found key and certificate for user %q", user)
		return nil
	}

	util.PrettyPrintOk(lp.Log, "Generating certificates for user %q", user)

	adminKey, adminCert, err := generateCert(ca, user, SANs, groups...)
	if err != nil {
		return fmt.Errorf("error during user cert generation: %v", err)
	}
	err = tls.WriteCert(adminKey, adminCert, user, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for user %q: %v", user, err)
	}
	return nil
}

func (lp *LocalPKI) validateUserCert(user string, groups []string) (valid bool, warn []error, err error) {
	SANs := []string{user}
	return tls.CertExistsAndValid(user, SANs, groups, user, lp.GeneratedCertsDirectory)
}

func generateCert(ca *tls.CA, commonName string, hostList []string, organizations ...string) (key, cert []byte, err error) {
	req := csr.CertificateRequest{
		CN: commonName,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
	}

	if len(hostList) > 0 {
		req.Hosts = hostList
	}

	for _, org := range organizations {
		name := csr.Name{O: org}
		req.Names = append(req.Names, name)
	}

	key, cert, err = tls.NewCert(ca, req)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating certs for %q: %v", commonName, err)
	}
	return key, cert, err
}

func clusterCertsSubjectAlternateNames(plan *Plan) ([]string, error) {
	kubeServiceIP, err := getKubernetesServiceIP(plan)
	if err != nil {
		return nil, fmt.Errorf("Error getting kubernetes service IP: %v", err)
	}
	defaultCertHosts := []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
		"127.0.0.1",
		kubeServiceIP,
	}
	return defaultCertHosts, nil
}

func isMasterNode(plan Plan, node Node) bool {
	for _, master := range plan.Master.Nodes {
		if node == master {
			return true
		}
	}
	return false
}
