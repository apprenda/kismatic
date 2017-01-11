package tls

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	err := CreateDir(dir, 0744)
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

func keyName(s string) string { return fmt.Sprintf("%s-key.pem", s) }

func certName(s string) string { return fmt.Sprintf("%s.pem", s) }

// CreateDir check if directory exists and create it
func CreateDir(dir string, perm os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, perm)
		if err != nil {
			return fmt.Errorf("error creating destination dir: %v", err)
		}
	}

	return nil
}
