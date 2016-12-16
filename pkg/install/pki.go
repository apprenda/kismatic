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
	certOrganization = "Apprenda"
	certOrgUnit      = "Kismatic"
	certCountry      = "US"
	certState        = "NY"
	certLocality     = "Troy"
)

// The PKI provides a way for generating certificates for the cluster described by the Plan
type PKI interface {
	CertificateAuthorityExists() (bool, error)
	NodeCertificateExists(node Node) (bool, error)
	GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error
	GetClusterCA() (*tls.CA, error)
	GenerateClusterCA(p *Plan) (*tls.CA, error)
	GenerateClusterCertificates(p *Plan, ca *tls.CA, users []string) error
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

	util.PrettyPrintOk(lp.Log, "Generating cluster Certificate Authority")
	// It doesn't exist, generate one
	caSubject := tls.Subject{
		Organization:       certOrganization,
		OrganizationalUnit: certOrgUnit,
		Country:            certCountry,
		State:              certState,
		Locality:           certLocality,
	}
	key, cert, err := tls.NewCACert(lp.CACsr, p.Cluster.Name, caSubject)
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

// GenerateClusterCertificates creates a Certificates for all nodes on the cluster
func (lp *LocalPKI) GenerateClusterCertificates(p *Plan, ca *tls.CA, users []string) error {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}
	nodes := []Node{}
	nodes = append(nodes, p.Etcd.Nodes...)
	nodes = append(nodes, p.Master.Nodes...)
	nodes = append(nodes, p.Worker.Nodes...)
	if p.Ingress.Nodes != nil {
		nodes = append(nodes, p.Ingress.Nodes...)
	}

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
	// Finally, create certs for user if they are missing
	for _, user := range users {
		if err := lp.generateUserCert(p, user, ca); err != nil {
			return err
		}
	}
	return nil
}

// GenerateNodeCertificate creates a private key and certificate for the given node
func (lp *LocalPKI) GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error {
	// Don't generate if the key pair is already there
	exist, err := tls.CertKeyPairExists(node.Host, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error verifying certificates for node %q: %v", node.Host, err)
	}
	if exist {
		util.PrettyPrintOk(lp.Log, "Found key and certificate for node %q", node.Host)
		return nil
	}
	// Build list of SANs
	clusterSANs, err := clusterCertsSubjectAlternateNames(plan)
	if err != nil {
		return err
	}
	util.PrettyPrintOk(lp.Log, "Generating certificates for host %q", node.Host)
	nodeSANs := append(clusterSANs, node.Host, node.IP, node.InternalIP)
	if isMasterNode(*plan, node) {
		if plan.Master.LoadBalancedFQDN != "" {
			nodeSANs = append(nodeSANs, plan.Master.LoadBalancedFQDN)
		}
		if plan.Master.LoadBalancedShortName != "" {
			nodeSANs = append(nodeSANs, plan.Master.LoadBalancedShortName)
		}
	}
	key, cert, err := generateCert(node.Host, plan, nodeSANs, ca)
	if err != nil {
		return fmt.Errorf("error during cluster cert generation: %v", err)
	}
	err = tls.WriteCert(key, cert, node.Host, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for host %q: %v", node.Host, err)
	}
	return nil
}

func (lp *LocalPKI) generateDockerRegistryCert(p *Plan, ca *tls.CA) error {
	// Skip generation if already exist
	exist, err := tls.CertKeyPairExists("docker", lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error verifying certificates for docker registry: %v", err)
	}
	if exist {
		util.PrettyPrintOk(lp.Log, "Found certificate for docker registry")
	}
	util.PrettyPrintOk(lp.Log, "Generating certificates for docker registry")
	// Default registry will be deployed on the first master
	n := p.Master.Nodes[0]
	dockerKey, dockerCert, err := generateCert(n.Host, p, []string{n.Host, n.IP, n.InternalIP}, ca)
	if err != nil {
		return fmt.Errorf("error during user cert generation: %v", err)
	}
	err = tls.WriteCert(dockerKey, dockerCert, "docker", lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for docker registry")
	}
	return nil
}

func (lp *LocalPKI) generateServiceAccountCert(p *Plan, ca *tls.CA) error {
	certName := "service-account"
	exists, err := tls.CertKeyPairExists(certName, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error verifying service account certs: %v", err)
	}
	if exists {
		util.PrettyPrintOk(lp.Log, "Found key and certificate for service accounts")
		return nil
	}
	key, cert, err := generateCert("kube-service-account", p, []string{}, ca)
	if err != nil {
		return fmt.Errorf("error generating service account certs: %v", err)
	}
	if err = tls.WriteCert(key, cert, certName, lp.GeneratedCertsDirectory); err != nil {
		return fmt.Errorf("error writing generated service account cert: %v", err)
	}
	util.PrettyPrintOk(lp.Log, "Generating certificates for service accounts")
	return nil
}

func (lp *LocalPKI) generateUserCert(p *Plan, user string, ca *tls.CA) error {
	exists, err := tls.CertKeyPairExists(user, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error verifying user certificates: %v", err)
	}
	if exists {
		util.PrettyPrintOk(lp.Log, "Found key and certificate for user %q", user)
		return nil
	}
	util.PrettyPrintOk(lp.Log, "Generating certificates for user %q", user)
	adminKey, adminCert, err := generateCert(user, p, []string{user}, ca)
	if err != nil {
		return fmt.Errorf("error during user cert generation: %v", err)
	}
	err = tls.WriteCert(adminKey, adminCert, user, lp.GeneratedCertsDirectory)
	if err != nil {
		return fmt.Errorf("error writing cert files for user %q: %v", user, err)
	}
	return nil
}

func generateCert(cnName string, p *Plan, hostList []string, ca *tls.CA) (key, cert []byte, err error) {
	req := csr.CertificateRequest{
		CN: cnName,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
		Hosts: hostList,
		Names: []csr.Name{
			{
				O:  certOrganization,
				OU: certOrgUnit,
				C:  certCountry,
				ST: certState,
				L:  certLocality,
			},
		},
	}
	key, cert, err = tls.NewCert(ca, req)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating certs for %q: %v", cnName, err)
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
