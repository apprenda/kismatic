package tls

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
)

func TestGenerateNewCertificate(t *testing.T) {
	key, caCert, err := NewCACert("test/ca-csr.json", "someCN")
	if err != nil {
		t.Fatalf("error creating CA: %v", err)
	}
	parsedCACert, err := helpers.ParseCertificatePEM(caCert)
	if err != nil {
		t.Fatalf("error parsing CA Certificate: %v", err)
	}
	ca := &CA{
		Key:        key,
		Cert:       caCert,
		ConfigFile: "test/ca-config.json",
		Profile:    "kubernetes",
	}
	certHosts := []string{"testHostname", "otherName", "127.0.0.1", "10.5.6.217"}
	req := csr.CertificateRequest{
		CN: "testKube",
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
		Hosts: certHosts,
		Names: []csr.Name{
			{
				C:  "US",
				L:  "Troy",
				O:  "Kubernetes",
				OU: "Cluster",
				ST: "New York",
			},
		},
	}

	if err != nil {
		t.Fatalf("error decoding csr: %v", err)
	}

	_, cert, err := NewCert(ca, req)
	if err != nil {
		t.Errorf("error creating certificate: %v", err)
	}

	parsedCert, err := helpers.ParseCertificatePEM(cert)
	if err != nil {
		t.Fatalf("error parsing certificate: %v", err)
	}

	if parsedCert.IsCA {
		t.Errorf("Non-CA certificate is CA")
	}

	if parsedCert.Subject.CommonName != req.CN {
		t.Errorf("common name mismatch: expected %q, but got %q", req.CN, parsedCert.Subject.CommonName)
	}

	if parsedCert.Subject.Organization[0] != req.Names[0].O {
		t.Errorf("organization mismatch: expected %q, but got %q", req.Names[0].O, parsedCert.Subject.Organization)
	}

	if parsedCert.Subject.OrganizationalUnit[0] != req.Names[0].OU {
		t.Errorf("organizational unit mismatch: expected %q, but got %q", req.Names[0].OU, parsedCert.Subject.OrganizationalUnit)
	}

	if parsedCert.Subject.Country[0] != req.Names[0].C {
		t.Errorf("country mismatch: expected %q, but got %q", req.Names[0].C, parsedCert.Subject.Country[0])
	}

	if parsedCert.Subject.Locality[0] != req.Names[0].L {
		t.Errorf("locality mismatch: expected %q, but got %q", req.Names[0].L, parsedCert.Subject.Locality[0])
	}

	if parsedCert.Subject.Province[0] != req.Names[0].ST {
		t.Errorf("state mismatch: expected %q, but got %q", req.Names[0].ST, parsedCert.Subject.Province[0])
	}

	if !reflect.DeepEqual(parsedCert.Issuer, parsedCACert.Subject) {
		t.Errorf("cert issuer is not equal to the CA's subject")
	}

	if !bytes.Equal(parsedCert.AuthorityKeyId, parsedCACert.SubjectKeyId) {
		t.Errorf("certificate auth key ID is not the subject key ID of the CA")
	}

	expectedDNSNames := []string{"testHostname", "otherName"}
	if !reflect.DeepEqual(expectedDNSNames, parsedCert.DNSNames) {
		t.Errorf("DNS names of the generated certificate are invalid")
	}

	expectedIPAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("10.5.6.217")}
	if len(expectedIPAddresses) != len(parsedCert.IPAddresses) {
		t.Errorf("expected %d IP addresses, but got %d", len(expectedIPAddresses), len(parsedCert.IPAddresses))
	}
	for i := 0; i < len(expectedIPAddresses); i++ {
		if !expectedIPAddresses[i].Equal(parsedCert.IPAddresses[i]) {
			t.Errorf("expected IP %q, but got %q", expectedIPAddresses[i], parsedCert.IPAddresses[i])
		}
	}

	// Verify expiration
	now := time.Now().UTC()
	d, err := time.ParseDuration("8760h")
	if err != nil {
		t.Fatalf("error parsing duration: %v", err)
	}
	expectedExpiration := now.Add(d)
	if expectedExpiration.Year() != parsedCert.NotAfter.Year() || expectedExpiration.YearDay() != parsedCert.NotAfter.YearDay() {
		t.Errorf("expected expiration date %q, got %q", expectedExpiration, parsedCert.NotAfter)
	}

}

func TestCertValid(t *testing.T) {
	tests := []struct {
		expectedCN            string
		expectedSANs          []string
		expectedOrganizations []string
		certCN                string
		certSANs              []string
		certOrganizations     []string
		valid                 bool
	}{
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        true,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"node1", "10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"node1", "10.0.0.1", "192.168.99.101"},
			valid:        true,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{},
			certCN:       "node1",
			certSANs:     []string{},
			valid:        true,
		},
		{
			expectedCN: "node1",
			certCN:     "node1",
			valid:      true,
		},
		{
			expectedCN:   "kube-service-account",
			expectedSANs: []string{},
			certCN:       "kube-service-account",
			certSANs:     []string{},
			valid:        true,
		},
		{
			expectedCN:   "admin",
			expectedSANs: []string{"admin"},
			certCN:       "admin",
			certSANs:     []string{"admin"},
			valid:        true,
		},
		{
			expectedCN:   "other-node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        false,
		},
		{
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        true,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        true,
		},
		{
			expectedCN: "node1",
			certCN:     "node1",
			certSANs:   []string{"10.0.0.1", "192.168.99.101"},
			valid:      true,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1"},
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{},
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			valid:        false,
		},
		{
			certCN:   "node1",
			certSANs: []string{"10.0.0.1", "192.168.99.101"},
			valid:    false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"node1", "10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"10.0.0.1", "192.168.99.101"},
			valid:        false,
		},
		{
			expectedCN:   "node1",
			expectedSANs: []string{"10.0.0.1", "192.168.99.101"},
			certCN:       "node1",
			certSANs:     []string{"node1", "10.0.0.1", "192.168.99.101"},
			valid:        true,
		},
		{
			expectedOrganizations: []string{"one", "two"},
			certOrganizations:     []string{"one", "two"},
			valid:                 true,
		},
		{
			expectedOrganizations: []string{"one", "two"},
			certOrganizations:     []string{},
			valid:                 false,
		},
		{
			expectedOrganizations: []string{"one", "two"},
			certOrganizations:     []string{"one"},
			valid:                 false,
		},
	}

	tempDir, err := ioutil.TempDir("", "cert-tests")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer cleanup(tempDir, t)

	key, caCert, err := NewCACert("test/ca-csr.json", "someCN")
	if err != nil {
		t.Fatalf("error creating CA: %v", err)
	}
	ca := &CA{
		Key:        key,
		Cert:       caCert,
		ConfigFile: "test/ca-config.json",
		Profile:    "kubernetes",
	}

	// Assert that the method returns an error if the cert does not exist
	_, err = CertValid(tests[0].expectedCN, tests[0].expectedSANs, tests[0].expectedOrganizations, "doesnotexist", tempDir)
	if err == nil {
		t.Errorf("expected an error, as the certificate does not exist.")
	}

	for i, test := range tests {
		key, cert, err := NewCert(ca, *buildReq(test.certCN, test.certSANs, test.certOrganizations))
		if err != nil {
			t.Error(err)
		}
		name := "cert-test-" + strconv.Itoa(i)
		certPath := filepath.Join(tempDir, name+".pem")
		keyPath := filepath.Join(tempDir, name+"-key.pem")
		fCert, err := os.Create(certPath)
		if err != nil {
			t.Fatalf("failed to create certificate: %v", err)
		}
		if _, err := fCert.Write(cert); err != nil {
			t.Fatalf("failed to write certificate: %v", err)
		}
		fKey, err := os.Create(keyPath)
		if err != nil {
			t.Fatalf("failed to  create private key file: %v", err)
		}
		if _, err := fKey.Write(key); err != nil {
			t.Fatalf("failed to write certificate: %v", err)
		}

		warn, err := CertValid(test.expectedCN, test.expectedSANs, test.expectedOrganizations, name, tempDir)
		if err != nil {
			t.Errorf("Unexpected error for %d: %v", i, err)
		}
		if test.valid && len(warn) > 0 {
			t.Errorf("Test %d - Expected a valid certificate, but got validation warnings: %v: \n", i, warn)
		}
		if !test.valid && len(warn) == 0 {
			t.Errorf("Test %d - Expected an invalid certificate, but did not get validation warnings", i)
		}
	}
}

func buildReq(CN string, SANs []string, organizations []string) *csr.CertificateRequest {
	req := &csr.CertificateRequest{
		CN: CN,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
		Hosts: SANs,
	}
	for _, org := range organizations {
		req.Names = append(req.Names, csr.Name{O: org})
	}

	return req
}

func cleanup(dir string, t *testing.T) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("failed cleaning up temp directory: %v", err)
	}
}
