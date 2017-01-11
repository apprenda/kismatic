package tls

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
)

func TestGenerateNewCertificate(t *testing.T) {
	// Create CA Cert
	subject := Subject{
		Country:            "someCountry",
		State:              "someState",
		Locality:           "someLocality",
		Organization:       "someOrganization",
		OrganizationalUnit: "someOrgUnit",
	}
	key, caCert, err := NewCACert("test/ca-csr.json", "someCN", subject)
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
