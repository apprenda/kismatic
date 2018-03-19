package install

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/cloudflare/cfssl/csr"
)

const (
	adminUser                                  = "admin"
	adminGroup                                 = "system:masters"
	adminCertFilename                          = "admin"
	adminCertFilenameKETPre133                 = "admin"
	serviceAccountCertFilename                 = "service-account"
	serviceAccountCertCommonName               = "kube-service-account"
	schedulerCertFilenamePrefix                = "kube-scheduler"
	schedulerUser                              = "system:kube-scheduler"
	controllerManagerCertFilenamePrefix        = "kube-controller-manager"
	controllerManagerUser                      = "system:kube-controller-manager"
	kubeletUserPrefix                          = "system:node"
	kubeletGroup                               = "system:nodes"
	kubeAPIServerKubeletClientClientFilename   = "apiserver-kubelet-client"
	kubeAPIServerKubeletClientClientCommonName = "kube-apiserver-kubelet-client"
	contivProxyServerCertFilename              = "contiv-proxy-server"
	proxyClientCACommonName                    = "proxyClientCA"
	proxyClientCertFilename                    = "proxy-client"
	proxyClientCertCommonName                  = "aggregator"
)

// The PKI provides a way for generating certificates for the cluster described by the Plan
type PKI interface {
	CertificateAuthorityExists() (bool, error)
	GenerateClusterCA(p *Plan) (*tls.CA, error)
	GetClusterCA() (*tls.CA, error)
	GenerateProxyClientCA(p *Plan) (*tls.CA, error)
	GetProxyClientCA() (*tls.CA, error)
	GenerateClusterCertificates(p *Plan, clusterCA *tls.CA, proxyClientCA *tls.CA) error
	NodeCertificateExists(node Node) (bool, error)
	GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error
	GenerateCertificate(name string, validityPeriod string, commonName string, subjectAlternateNames []string, organizations []string, ca *tls.CA, overwrite bool) (bool, error)
}

// LocalPKI is a file-based PKI
type LocalPKI struct {
	CACsr                   string
	GeneratedCertsDirectory string
	Log                     io.Writer
}

type certificateSpec struct {
	description           string
	filename              string
	commonName            string
	subjectAlternateNames []string
	organizations         []string
	ca                    *tls.CA
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

// CertificateAuthorityExists returns true if the CA for the cluster exists
func (lp *LocalPKI) CertificateAuthorityExists() (bool, error) {
	return tls.CertKeyPairExists("ca", lp.GeneratedCertsDirectory)
}

// GenerateClusterCA creates a Certificate Authority for the cluster
func (lp *LocalPKI) GenerateClusterCA(p *Plan) (*tls.CA, error) {
	exists, err := tls.CertKeyPairExists("ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error verifying CA certificate/key: %v", err)
	}
	if exists {
		return lp.GetClusterCA()
	}

	// CA keypair doesn't exist, generate one
	util.PrettyPrintOk(lp.Log, "Generating cluster Certificate Authority")
	key, cert, err := tls.NewCACert(lp.CACsr, p.Cluster.Name, p.Cluster.Certificates.CAExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA Cert: %v", err)
	}
	if err = tls.WriteCert(key, cert, "ca", lp.GeneratedCertsDirectory); err != nil {
		return nil, fmt.Errorf("error writing CA files: %v", err)
	}
	return &tls.CA{
		Cert: cert,
		Key:  key,
	}, nil
}

// GetClusterCA returns the cluster CA
func (lp *LocalPKI) GetClusterCA() (*tls.CA, error) {
	key, cert, err := tls.ReadCACert("ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate/key: %v", err)
	}
	return &tls.CA{
		Cert: cert,
		Key:  key,
	}, nil
}

// GenerateProxyClientCA creates a Certificate Authority for the cluster
func (lp *LocalPKI) GenerateProxyClientCA(p *Plan) (*tls.CA, error) {
	exists, err := tls.CertKeyPairExists("proxy-client-ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error verifying proxy-client CA certificate/key: %v", err)
	}
	if exists {
		return lp.GetProxyClientCA()
	}

	// CA keypair doesn't exist, generate one
	util.PrettyPrintOk(lp.Log, "Generating proxy-client Certificate Authority")
	key, cert, err := tls.NewCACert(lp.CACsr, proxyClientCACommonName, p.Cluster.Certificates.CAExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy-client CA Cert: %v", err)
	}
	if err = tls.WriteCert(key, cert, "proxy-client-ca", lp.GeneratedCertsDirectory); err != nil {
		return nil, fmt.Errorf("error writing proxy-client CA files: %v", err)
	}
	return &tls.CA{
		Cert: cert,
		Key:  key,
	}, nil
}

// GetProxyClientCA returns the cluster CA
func (lp *LocalPKI) GetProxyClientCA() (*tls.CA, error) {
	key, cert, err := tls.ReadCACert("proxy-client-ca", lp.GeneratedCertsDirectory)
	if err != nil {
		return nil, fmt.Errorf("error reading proxy-client CA certificate/key: %v", err)
	}
	return &tls.CA{
		Cert: cert,
		Key:  key,
	}, nil
}

// GenerateClusterCertificates creates all certificates required for the cluster
// described in the plan file.
func (lp *LocalPKI) GenerateClusterCertificates(p *Plan, clusterCA *tls.CA, proxyClientCA *tls.CA) error {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}

	manifest, err := p.certSpecs(clusterCA, proxyClientCA)
	if err != nil {
		return err
	}

	for _, s := range manifest {
		exists, err := tls.CertKeyPairExists(s.filename, lp.GeneratedCertsDirectory)
		if err != nil {
			return err
		}

		// Pre-existing admin certificates from KET < 1.3.3 are not valid
		// due to changes required for RBAC. Rename it if necessary.
		if exists && s.filename == adminCertFilenameKETPre133 {
			ok, err := renamePre133AdminCert(s.filename, lp.GeneratedCertsDirectory)
			if err != nil {
				return err
			}
			// We renamed it, so it doesn't exist anymore
			if ok {
				util.PrettyPrintWarn(lp.Log, "Existing admin certificate is invalid. Backing up and regenerating.")
				exists = false
			}
		}

		if exists {
			warnings, err := tls.CertValid(s.commonName, s.subjectAlternateNames, s.organizations, s.filename, lp.GeneratedCertsDirectory)
			if err != nil {
				return err
			}
			if len(warnings) > 0 {
				util.PrettyPrintErr(lp.Log, "Found certificate for %s, but it is not valid", s.description)
				util.PrintValidationErrors(lp.Log, warnings)
				return fmt.Errorf("invalid certificate found for %q", s.description)
			}
			// This cert is valid, move onto the next certificate
			util.PrettyPrintOk(lp.Log, "Found valid certificate for %s", s.description)
			continue
		}

		// Cert doesn't exist. Generate it
		if err := generateCert(lp.GeneratedCertsDirectory, s, p.Cluster.Certificates.Expiry); err != nil {
			return err
		}
		util.PrettyPrintOk(lp.Log, "Generated certificate for %s", s.description)
	}
	return nil
}

// Validates that the certificate was generated by us. If so, renames it
// to make a backup and returns true. Otherwise returns false.
func renamePre133AdminCert(filename, dir string) (bool, error) {
	cert, err := tls.ReadCert(filename, dir)

	if err != nil {
		return false, fmt.Errorf("error reading admin certificate: %v", err)
	}
	// Ensure it was generated by us
	if len(cert.Subject.Organization) == 1 && cert.Subject.Organization[0] == "Apprenda" &&
		len(cert.Subject.OrganizationalUnit) == 1 && cert.Subject.OrganizationalUnit[0] == "Kismatic" &&
		len(cert.Subject.Country) == 1 && cert.Subject.Country[0] == "US" &&
		len(cert.Subject.Province) == 1 && cert.Subject.Province[0] == "NY" &&
		len(cert.Subject.Locality) == 1 && cert.Subject.Locality[0] == "Troy" {

		certFile := filepath.Join(dir, filename+".pem")
		if err = os.Rename(certFile, certFile+".bak"); err != nil {
			return false, fmt.Errorf("error backing up existing admin certificate: %v", err)
		}
		return true, nil
	}
	return false, nil
}

// ValidateClusterCertificates validates any certificates that already exist
// in the expected directory.
func (lp *LocalPKI) ValidateClusterCertificates(p *Plan) (warns []error, errs []error) {
	if lp.Log == nil {
		lp.Log = ioutil.Discard
	}
	manifest, err := p.certSpecs(nil, nil)
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

// NodeCertificateExists returns true if the node's key and certificate exist
func (lp *LocalPKI) NodeCertificateExists(node Node) (bool, error) {
	return tls.CertKeyPairExists(node.Host, lp.GeneratedCertsDirectory)
}

// GenerateNodeCertificate creates a private key and certificate for the given node
func (lp *LocalPKI) GenerateNodeCertificate(plan *Plan, node Node, ca *tls.CA) error {
	m, err := node.certSpecs(*plan, ca)
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
		if err := generateCert(lp.GeneratedCertsDirectory, s, plan.Cluster.Certificates.Expiry); err != nil {
			return err
		}
		util.PrettyPrintOk(lp.Log, "Generated certificate for %s", s.description)
	}
	return nil
}

// GenerateCertificate creates a private key and certificate for the given name, CN, subjectAlternateNames and organizations
// If cert exists, will not fail
// Pass overwrite to replace an existing cert
func (lp *LocalPKI) GenerateCertificate(name string, validityPeriod string, commonName string, subjectAlternateNames []string, organizations []string, ca *tls.CA, overwrite bool) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name cannot be empty")
	}
	if validityPeriod == "" {
		return false, fmt.Errorf("validityPeriod cannot be empty")
	}
	if ca == nil {
		return false, fmt.Errorf("ca cannot be nil")
	}
	exists, err := tls.CertKeyPairExists(name, lp.GeneratedCertsDirectory)
	if err != nil {
		return false, fmt.Errorf("could not determine if certificate for %s exists: %v", name, err)
	}
	// if exists and overwrite == false return
	// otherwise cert will be created and replaced if exists
	if exists && !overwrite {
		return true, nil
	}

	spec := certificateSpec{
		description:           name,
		filename:              name,
		commonName:            commonName,
		subjectAlternateNames: subjectAlternateNames,
		organizations:         organizations,
		ca:                    ca,
	}

	if err := generateCert(lp.GeneratedCertsDirectory, spec, validityPeriod); err != nil {
		return exists, fmt.Errorf("could not generate certificate %s: %v", name, err)
	}

	return exists, nil
}

func generateCert(certDir string, spec certificateSpec, expiryStr string) error {
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return fmt.Errorf("%q is not a valid duration for certificate expiry", expiryStr)
	}
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

	key, cert, err := tls.NewCert(spec.ca, req, expiry)
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
