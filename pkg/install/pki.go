package install

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/cloudflare/cfssl/csr"
)

const (
	adminUser                           = "admin"
	adminGroup                          = "system:masters"
	adminCertFilename                   = "admin"
	dockerRegistryCertFilename          = "docker-registry"
	serviceAccountCertFilename          = "service-account"
	serviceAccountCertCommonName        = "kube-service-account"
	schedulerCertFilenamePrefix         = "kube-scheduler"
	schedulerUser                       = "system:kube-scheduler"
	controllerManagerCertFilenamePrefix = "kube-controller-manager"
	controllerManagerUser               = "system:kube-controller-manager"
	kubeProxyCertFilenamePrefix         = "kube-proxy"
	kubeProxyUser                       = "system:kube-proxy"
	kubeletUserPrefix                   = "system:node"
	kubeletGroup                        = "system:nodes"
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

type certificateSpec struct {
	description           string
	filename              string
	commonName            string
	subjectAlternateNames []string
	organizations         []string
}

func (s certificateSpec) equal(other certificateSpec) bool {
	prelimEqual := s.description == other.description &&
		s.filename == other.filename &&
		s.commonName == other.commonName &&
		len(s.subjectAlternateNames) == len(other.subjectAlternateNames) &&
		len(s.organizations) == len(other.organizations)
	if !prelimEqual {
		return false
	}
	// Compare subject alt. names
	thisSAN := make([]string, len(s.subjectAlternateNames))
	otherSAN := make([]string, len(other.subjectAlternateNames))
	// Clone and sort
	copy(thisSAN, s.subjectAlternateNames)
	copy(otherSAN, other.subjectAlternateNames)
	sort.Strings(thisSAN)
	sort.Strings(otherSAN)

	for _, x := range thisSAN {
		for _, y := range otherSAN {
			if x != y {
				return false
			}
		}
	}
	// Compare organizations
	thisOrgs := make([]string, len(s.organizations))
	otherOrgs := make([]string, len(other.organizations))
	// clone and sort
	copy(thisOrgs, s.organizations)
	copy(otherOrgs, other.organizations)
	sort.Strings(thisOrgs)
	sort.Strings(otherOrgs)

	for _, x := range thisOrgs {
		for _, y := range otherOrgs {
			if x != y {
				return false
			}
		}
	}
	return true
}

// returns a list of specs for all the certs that are required for the node
func certManifestForNode(plan Plan, node Node) ([]certificateSpec, error) {
	m := []certificateSpec{}
	roles := plan.GetRolesForIP(node.IP)

	// Certificates for etcd
	if contains("etcd", roles) {
		san := []string{node.Host, node.IP, "127.0.0.1"}
		if node.InternalIP != "" {
			san = append(san, node.InternalIP)
		}
		m = append(m, certificateSpec{
			description:           fmt.Sprintf("%s etcd server", node.Host),
			filename:              fmt.Sprintf("%s-etcd", node.Host),
			commonName:            node.Host,
			subjectAlternateNames: san,
		})
	}

	// Certificates for master
	if contains("master", roles) {
		// API Server certificate
		san, err := clusterCertsSubjectAlternateNames(plan)
		if err != nil {
			return nil, err
		}
		san = append(san, node.Host, node.IP, "127.0.0.1")
		if node.InternalIP != "" {
			san = append(san, node.InternalIP)
		}
		if !contains(plan.Master.LoadBalancedFQDN, san) {
			san = append(san, plan.Master.LoadBalancedFQDN)
		}
		if !contains(plan.Master.LoadBalancedShortName, san) {
			san = append(san, plan.Master.LoadBalancedShortName)
		}
		m = append(m, certificateSpec{
			description:           fmt.Sprintf("%s API server", node.Host),
			filename:              fmt.Sprintf("%s-apiserver", node.Host),
			commonName:            node.Host,
			subjectAlternateNames: san,
		})
		// Controller manager certificate
		m = append(m, certificateSpec{
			description: "kubernetes controller manager",
			filename:    controllerManagerCertFilenamePrefix,
			commonName:  controllerManagerUser,
		})
		// Scheduler client certificate
		m = append(m, certificateSpec{
			description: "kubernetes scheduler",
			filename:    schedulerCertFilenamePrefix,
			commonName:  schedulerUser,
		})
		// Certificate for signing service account tokens
		m = append(m, certificateSpec{
			description: "service account signing",
			filename:    serviceAccountCertFilename,
			commonName:  serviceAccountCertCommonName,
		})
	}

	// Kubelet and kube-proxy client certificate
	if containsAny([]string{"master", "worker", "ingress", "storage"}, roles) {
		m = append(m, certificateSpec{
			description:   fmt.Sprintf("%s kubelet", node.Host),
			filename:      fmt.Sprintf("%s-kubelet", node.Host),
			commonName:    fmt.Sprintf("%s:%s", kubeletUserPrefix, node.Host),
			organizations: []string{kubeletGroup},
		})

		m = append(m, certificateSpec{
			description: "kube-proxy",
			filename:    kubeProxyCertFilenamePrefix,
			commonName:  kubeProxyUser,
		})
		// etcd client certificate
		// all nodes need to be able to talk to etcd b/c of calico
		m = append(m, certificateSpec{
			description: "etcd client",
			filename:    "etcd-client",
			commonName:  "etcd-client",
		})
	}

	return m, nil
}

// returns a list of cert specs for the cluster described in the plan file
func certManifestForCluster(plan Plan) ([]certificateSpec, error) {
	m := []certificateSpec{}

	// Certificate for nodes
	nodes := plan.GetUniqueNodes()
	for _, n := range nodes {
		nodeManifest, err := certManifestForNode(plan, n)
		if err != nil {
			return nil, err
		}

		// Some nodes share common certificates between them. E.g. the kube-proxy client cert.
		// Before appending to the manifest, we ensure that this cert is not already in it.
		for _, s := range nodeManifest {
			if !certSpecInManifest(s, m) {
				m = append(m, s)
			}
		}
	}

	// Certificate for docker registry
	if plan.DockerRegistry.SetupInternal {
		dockerRegistryNode := plan.Master.Nodes[0]
		san := []string{dockerRegistryNode.Host, dockerRegistryNode.IP}
		if dockerRegistryNode.InternalIP != "" {
			san = append(san, dockerRegistryNode.InternalIP)
		}
		m = append(m, certificateSpec{
			description:           "internal private docker registry",
			filename:              dockerRegistryCertFilename,
			commonName:            dockerRegistryNode.Host,
			subjectAlternateNames: san,
		})
	}

	// Admin certificate
	m = append(m, certificateSpec{
		description:   "admin client",
		filename:      adminCertFilename,
		commonName:    adminUser,
		organizations: []string{adminGroup},
	})

	return m, nil
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

	manifest, err := certManifestForCluster(*p)
	if err != nil {
		return err
	}

	for _, s := range manifest {
		exists, err := tls.CertKeyPairExists(s.filename, lp.GeneratedCertsDirectory)
		if err != nil {
			return err
		}
		if exists {
			warn, err := tls.CertValid(s.commonName, s.subjectAlternateNames, s.organizations, s.filename, lp.GeneratedCertsDirectory)
			if err != nil {
				return err
			}
			if len(warn) > 0 {
				util.PrettyPrintErr(lp.Log, "Found certificate for %s, but it is not valid", s.description)
				util.PrintValidationErrors(lp.Log, warn)
				return fmt.Errorf("invalid certificate found for %q", s.description)
			}
			// This cert is valid, move on
			util.PrettyPrintOk(lp.Log, "Found valid certificate for %s", s.description)
			continue
		}

		// Cert doesn't exist. Generate it
		if err := generateCert(ca, lp.GeneratedCertsDirectory, s); err != nil {
			return err
		}
		util.PrettyPrintOk(lp.Log, "Generated certificate for %s", s.description)
	}
	return nil
}

// ValidateClusterCertificates validates any certificates that already exist
// in the expected directory.
func (lp *LocalPKI) ValidateClusterCertificates(p *Plan) (warns []error, errs []error) {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}
	manifest, err := certManifestForCluster(*p)
	if err != nil {
		return nil, []error{err}
	}
	for _, s := range manifest {
		exists, err := tls.CertKeyPairExists(s.filename, lp.GeneratedCertsDirectory)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !exists {
			continue // nothing to validate... move on
		}
		warn, err := tls.CertValid(s.commonName, s.subjectAlternateNames, s.organizations, s.filename, lp.GeneratedCertsDirectory)
		if err != nil {
			errs = append(errs, err)
		}
		if len(warn) > 0 {
			warns = append(warns, warn...)
		}
	}
	return warns, errs
}

// GenerateNodeCertificate creates a private key and certificate for the given node
func (lp *LocalPKI) GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error {
	m, err := certManifestForNode(*plan, node)
	if err != nil {
		return err
	}
	for _, s := range m {
		exists, err := tls.CertKeyPairExists(s.filename, lp.GeneratedCertsDirectory)
		if err != nil {
			return err
		}
		if exists {
			warn, err := tls.CertValid(s.commonName, s.subjectAlternateNames, s.organizations, s.filename, lp.GeneratedCertsDirectory)
			if err != nil {
				return err
			}
			if len(warn) > 0 {
				util.PrettyPrintErr(lp.Log, "Found certificate for %s, but it is not valid", s.description)
				util.PrintValidationErrors(lp.Log, warn)
				return fmt.Errorf("invalid certificate found for %q", s.description)
			}
			// This cert is valid, move on
			util.PrettyPrintOk(lp.Log, "Found valid certificate for %s", s.description)
			continue
		}
		// Cert doesn't exist. Generate it
		if err := generateCert(ca, lp.GeneratedCertsDirectory, s); err != nil {
			return err
		}
		util.PrettyPrintOk(lp.Log, "Generated certificate for %s", s.description)
	}
	return nil
}

func generateCert(ca *tls.CA, certDir string, spec certificateSpec) error {
	req := csr.CertificateRequest{
		CN: spec.commonName,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
	}

	if len(spec.subjectAlternateNames) > 0 {
		req.Hosts = spec.subjectAlternateNames
	}

	for _, org := range spec.organizations {
		name := csr.Name{O: org}
		req.Names = append(req.Names, name)
	}

	key, cert, err := tls.NewCert(ca, req)
	if err != nil {
		return fmt.Errorf("error generating certs for %q: %v", spec.description, err)
	}
	if err = tls.WriteCert(key, cert, spec.filename, certDir); err != nil {
		return fmt.Errorf("error writing cert for %q: %v", spec.description, err)
	}
	return nil
}

func clusterCertsSubjectAlternateNames(plan Plan) ([]string, error) {
	kubeServiceIP, err := getKubernetesServiceIP(&plan)
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

func contains(x string, xs []string) bool {
	for _, s := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func containsAny(x []string, xs []string) bool {
	for _, s := range x {
		if contains(s, xs) {
			return true
		}
	}
	return false
}

func certSpecInManifest(spec certificateSpec, manifest []certificateSpec) bool {
	for _, s := range manifest {
		if s.equal(spec) {
			return true
		}
	}
	return false
}
