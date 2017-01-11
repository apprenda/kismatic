package tls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/log"
)

func init() {
	log.Level = log.LevelError
}

// The Subject contains the fields of the X.509 Subject
type Subject struct {
	Country            string
	State              string
	Locality           string
	Organization       string
	OrganizationalUnit string
}

// NewCACert creates a new Certificate Authority and returns it's private key and public certificate.
func NewCACert(csrFile string, commonName string, subject Subject) (key, cert []byte, err error) {
	// Open CSR file
	f, err := os.Open(csrFile)
	if os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("%q does not exist", csrFile)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %q", csrFile)
	}
	// Create CSR struct
	caCSR := &csr.CertificateRequest{
		KeyRequest: csr.NewBasicKeyRequest(),
	}
	err = json.NewDecoder(f).Decode(caCSR)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding CSR: %v", err)
	}
	// Set the subject information
	name := csr.Name{
		C:  subject.Country,
		ST: subject.State,
		L:  subject.Locality,
		O:  subject.Organization,
		OU: subject.OrganizationalUnit,
	}
	caCSR.Names = []csr.Name{name}
	caCSR.CN = commonName
	// Generate CA Cert according to CSR
	cert, _, key, err = initca.New(caCSR)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating CA cert: %v", err)
	}
	return key, cert, nil
}

// ReadCACert read CA file
func ReadCACert(name, dir string) (key, cert []byte, err error) {
	dest := filepath.Join(dir, keyName(name))
	key, errKey := ioutil.ReadFile(dest)
	if errKey != nil {
		return nil, nil, fmt.Errorf("error reading private key: %v", errKey)
	}
	dest = filepath.Join(dir, certName(name))
	cert, errCert := ioutil.ReadFile(dest)
	if errCert != nil {
		return nil, nil, fmt.Errorf("error reading certificate: %v", errKey)
	}
	return key, cert, nil
}
