package tls

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/apprenda/kismatic/pkg/util"
	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
)

// CA contains information about the Certificate Authority
type CA struct {
	// Key is the CA's private key.
	Key []byte
	// Password is the CA's private key password. Can be empty if not password is set.
	Password string
	// Cert is the CA's public certificate.
	Cert []byte
	// ConfigFile contains a cfssl configuration file for the Certificate Authority
	ConfigFile string
	// Profile to be used when signing with this Certificate Authority
	Profile string
}

// NewCert creates a new certificate/key pair using the CertificateAuthority provided
func NewCert(ca *CA, req csr.CertificateRequest) (key, cert []byte, err error) {
	g := &csr.Generator{Validator: genkey.Validator}
	csrBytes, key, err := g.ProcessRequest(&req)
	if err != nil {
		return nil, nil, fmt.Errorf("error processing CSR: %v", err)
	}
	// Get CA private key
	caPriv, err := helpers.ParsePrivateKeyPEMWithPassword(ca.Key, []byte(ca.Password))
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing privte key: %v", err)
	}
	// Parse CA Cert
	caCert, err := helpers.ParseCertificatePEM(ca.Cert)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing CA cert: %v", err)
	}
	sigAlgo := signer.DefaultSigAlgo(caPriv)
	// Get CA config from file
	caConfig, err := config.LoadFile(ca.ConfigFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading CA Config: %v", err)
	}
	// Create signer using CA
	s, err := local.NewSigner(caPriv, caCert, sigAlgo, caConfig.Signing)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating signer: %v", err)
	}
	// Generate cert using CA signer
	signReq := signer.SignRequest{
		Request: string(csrBytes),
		Profile: ca.Profile,
	}
	cert, err = s.Sign(signReq)
	if err != nil {
		return nil, nil, fmt.Errorf("error signing certificate: %v", err)
	}
	return key, cert, nil
}

// WriteCert writes cert and key files
func WriteCert(key, cert []byte, name, dir string) error {
	// Create destination dir if it doesn't exist
	err := util.CreateDir(dir, 0744)
	if err != nil {
		return err
	}
	// Write private key with read-only for user
	err = ioutil.WriteFile(filepath.Join(dir, keyName(name)), key, 0600)
	if err != nil {
		return fmt.Errorf("error writing private key: %v", err)
	}
	// Write cert
	err = ioutil.WriteFile(filepath.Join(dir, certName(name)), cert, 0644)
	if err != nil {
		return fmt.Errorf("error writing certificate: %v", err)
	}
	return nil
}

// CertKeyPairExists returns true if a key and matching certificate exist.
// Matching is defined as having the expected file names. No validation
// is performed on the actual bytes of the cert/key
func CertKeyPairExists(name, dir string) (bool, error) {
	kn := keyName(name)
	var err error
	if _, err = os.Stat(filepath.Join(dir, kn)); os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	cn := certName(name)
	if _, err = os.Stat(filepath.Join(dir, cn)); os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CertValid returns a list of validation warnings if the certificate values do not match
// the expected values.
// Validation rules:
// - common name: must match exactly
// - subject alternate names: the expected SANs must be a subset of the cert's SANs
// - organizations: the expected organizations must be a subset of the cert's organizations
// Subset validation is performed to allow operator to supply their own SANs and organizations
// Returns an error if trying to validate a cert that does not exist, or there
// is an issue reading or parsing the certificate
func CertValid(commonName string, SANs []string, organizations []string, name, dir string) (warn []error, err error) {
	// check if cert exists
	cn := certName(name)
	if _, err = os.Stat(filepath.Join(dir, cn)); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate %s does not exist", cn)
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error looking for certificate %s", cn)
	}

	// read the certificate file
	certBytes, err := ioutil.ReadFile(filepath.Join(dir, cn))
	if err != nil {
		return nil, fmt.Errorf("error reding cert %s: %v", name, err)
	}

	// verify certificate
	cert, err := helpers.ParseCertificatePEM(certBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing cert %s: %v", name, err)
	}

	if cert.Subject.CommonName != commonName {
		warn = append(warn, fmt.Errorf("Certificate %q: CN validation failed\n    expected %q, instead got %q", cn, commonName, cert.Subject.CommonName))
	}

	var certSANs []string
	for _, ip := range cert.IPAddresses {
		certSANs = append(certSANs, ip.String())
	}
	// DNS can be any string value
	certSANs = append(certSANs, cert.DNSNames...)

	// check if the SANs in the certificate contain the requested SANs
	// allows for operators to add their own custom SANs in the cert
	subset := util.Subset(SANs, certSANs)
	if !subset {
		// sort for readability
		sort.Strings(SANs)
		sort.Strings(certSANs)
		warn = append(warn, fmt.Errorf("Certificate %q: SANs validation failed\n    expected: \n\t%v \n    instead got: \n\t%v", cn, SANs, certSANs))
	}

	// Validate organizations
	subset = util.Subset(organizations, cert.Subject.Organization)
	if !subset {
		sort.Strings(organizations)
		sort.Strings(cert.Subject.Organization)
		warn = append(warn,
			fmt.Errorf("Certificate %q: Organizations validation failed\n    expected: \n\t%v \n    instead got: \n\t%v",
				cn, organizations, cert.Subject.Organization),
		)
	}

	return warn, nil
}

func keyName(s string) string { return fmt.Sprintf("%s-key.pem", s) }

func certName(s string) string { return fmt.Sprintf("%s.pem", s) }
