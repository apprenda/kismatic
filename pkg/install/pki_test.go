package install

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/helpers"
)

func getPKI(t *testing.T) LocalPKI {
	tempDir, err := ioutil.TempDir("", "pki-tests")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	pki := LocalPKI{
		CACsr:                   "test/ca-csr.json",
		CAConfigFile:            "test/ca-config.json",
		CASigningProfile:        "kubernetes",
		GeneratedCertsDirectory: tempDir,
		Log: ioutil.Discard,
	}
	return pki
}

func cleanup(dir string, t *testing.T) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("failed cleaning up temp directory: %v", err)
	}
}

func mustReadCertFile(certFile string, t *testing.T) *x509.Certificate {
	certPEM, err := ioutil.ReadFile(certFile)
	if err != nil {
		t.Fatalf("failed to read certificate file: %v", err)
	}
	cert, err := helpers.ParseCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("error reading host certificate: %v", err)
	}
	return cert
}

func getPlan() *Plan {
	return &Plan{
		Cluster: Cluster{
			Name: "someName",
			Certificates: CertsConfig{
				Expiry: "1h",
			},
			Networking: NetworkConfig{
				ServiceCIDRBlock: "10.0.0.0/24", // required for DNS service
			},
		},
		Etcd: NodeGroup{
			Nodes: []Node{
				Node{
					Host:       "etcd",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
		Master: MasterNodeGroup{
			Nodes: []Node{
				Node{
					Host:       "master",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
			LoadBalancedFQDN:      "someFQDN",
			LoadBalancedShortName: "someShortName",
		},
		Worker: NodeGroup{
			Nodes: []Node{
				Node{
					Host:       "worker",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
	}
}

// Validate the certificate subject information
func validateCertSubject(t *testing.T, cert *x509.Certificate, p *Plan) {
	if cert.Subject.Country[0] != certCountry {
		t.Errorf("country mismatch: expected %q, but got %q", certCountry, cert.Subject.Country[0])
	}
	if cert.Subject.Locality[0] != certLocality {
		t.Errorf("locality mismatch: expected %q, but got %q", certLocality, cert.Subject.Locality[0])
	}
	if cert.Subject.Province[0] != certState {
		t.Errorf("province mismatch: expected %q, but got %q", certState, cert.Subject.Province[0])
	}
	if cert.Subject.Organization[0] != certOrganization {
		t.Errorf("invalid organization in generated cert. Expected %q but got %q", certOrganization, cert.Subject.Organization[0])
	}
	if cert.Subject.OrganizationalUnit[0] != certOrgUnit {
		t.Errorf("invalid organizational unit in generated cert. Expected %q but got %q", certOrgUnit, cert.Subject.OrganizationalUnit[0])
	}
}

func TestGeneratedClusterCASubject(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Errorf("failed to generate cluster CA: %v", err)
	}
	caCert, err := helpers.ParseCertificatePEM(ca.Cert)
	if err != nil {
		t.Errorf("failed to parse generated cert: %v", err)
	}
	if !caCert.IsCA {
		t.Errorf("generated cert is not CA")
	}
	validateCertSubject(t, caCert, p)

	if caCert.Subject.CommonName != p.Cluster.Name {
		t.Errorf("common name mismatch: expected %q, got %q", p.Cluster.Name, caCert.Subject.CommonName)
	}
}

func TestGeneratedClusterCAWrittenToDestinationDir(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	_, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Errorf("error generating cluster CA: %v", err)
	}
	destDir := pki.GeneratedCertsDirectory
	certFile := filepath.Join(destDir, "ca.pem")
	_, err = os.Stat(certFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Error("generated certificate was not created in dest directory")
		} else {
			t.Errorf("error validating file existence: %v", err)
		}
	}
	keyFile := filepath.Join(destDir, "ca-key.pem")
	_, err = os.Stat(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Error("key not found for generated CA")
		} else {
			t.Errorf("error checking if CA private key exists: %v", err)
		}
	}
}

func TestClusterCAExistsGenerationSkipped(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	caFile := filepath.Join(pki.GeneratedCertsDirectory, "ca.pem")
	if _, err := os.Create(caFile); err != nil {
		t.Fatalf("error creating ca.pem file: %v", err)
	}

	keyFile := filepath.Join(pki.GeneratedCertsDirectory, "ca-key.pem")
	if _, err := os.Create(keyFile); err != nil {
		t.Fatalf("error creating ca-key.pem: %v", err)
	}

	if _, err := pki.GenerateClusterCA(&Plan{}); err != nil {
		t.Fatalf("generate CA method returned error: %v", err)
	}

	// Verify cert file wasn't touched
	caContents, err := ioutil.ReadFile(caFile)
	if err != nil {
		t.Errorf("error getting contents for %q: %v", caFile, err)
	}
	if len(caContents) != 0 {
		t.Error("CA File was modified")
	}

	// Verify key file wasn't touched
	keyContents, err := ioutil.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("error getting stat for %q: %v", keyFile, err)
	}
	if len(keyContents) != 0 {
		t.Error("Key file was modified")
	}
}

func TestGenerateClusterCertificatesNodeCert(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	node := p.Worker.Nodes[0]

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, node.Host+".pem")
	cert := mustReadCertFile(certFile, t)

	if cert.Subject.CommonName != node.Host {
		t.Errorf("common name mismatch: got %q, expected %q", cert.Subject.CommonName, node.Host)
	}

	validateCertSubject(t, cert, p)
	// Validate DNS names
	nameFound := false
	for _, name := range cert.DNSNames {
		if name == node.Host {
			nameFound = true
			break
		}
	}
	if !nameFound {
		t.Error("Expected node's DNS name in certificate, but was not there")
	}
	// Validate IP Addresses
	ipFound := false
	for _, ip := range cert.IPAddresses {
		if net.ParseIP(node.InternalIP).Equal(ip) {
			ipFound = true
			break
		}
	}
	if !ipFound {
		t.Error("Expected node's Internal IP in cert, but was not there")
	}
	ipFound = false
	for _, ip := range cert.IPAddresses {
		if net.ParseIP(node.IP).Equal(ip) {
			ipFound = true
			break
		}
	}
	if !ipFound {
		t.Error("Expected node's IP in cert, but was not there")
	}
}

func TestNodeCertExistsSkipGeneration(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	node := p.Worker.Nodes[0]
	p.Master = MasterNodeGroup{}
	p.Etcd = NodeGroup{}

	// Create the node cert and key file
	certFile := filepath.Join(pki.GeneratedCertsDirectory, node.Host+".pem")
	keyFile := filepath.Join(pki.GeneratedCertsDirectory, node.Host+"-key.pem")
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// try to generate again
	time := time.Now()
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Assert cert and key were not regenerated
	cert, err := os.Stat(certFile)
	if err != nil {
		t.Fatalf("error reading cert file:%v", err)
	}
	if cert.ModTime().After(time) {
		t.Error("cert file was modified")
	}

	key, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("error reading key file:%v", err)
	}
	if key.ModTime().After(time) {
		t.Error("key file was modified")
	}

	// Assert no other files were created
	expectedFiles := []string{"ca.pem", "ca-key.pem", node.Host + ".pem", node.Host + "-key.pem", "admin.pem", "admin-key.pem", "service-account.pem", "service-account-key.pem"}
	files, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error getting generated dir contents: %v", err)
	}
	if len(files) != len(expectedFiles) {
		t.Errorf("expected 6 files in certs directory, but found %d", len(files))
	}
	for _, f := range files {
		found := false
		for _, fn := range expectedFiles {
			if f.Name() == fn {
				found = true
			}
		}
		if !found {
			t.Errorf("found an unexpected file %q", f.Name())
		}
	}
}

func TestGenerateClusterCertificatesMultipleNodes(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	if err := pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("error generating cluster certs: %v", err)
	}

	// Validate all certs were created
	expectedFiles := []string{
		"ca.pem", "ca-key.pem",
		"master.pem", "master-key.pem",
		"etcd.pem", "etcd-key.pem",
		"worker.pem", "worker-key.pem",
		"admin.pem", "admin-key.pem",
		"service-account.pem", "service-account-key.pem",
	}
	files, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error reading generated certs directory: %v", err)
	}
	if len(expectedFiles) != len(files) {
		t.Errorf("expected %d files, but found %d", len(expectedFiles), len(files))
	}
	for _, expected := range expectedFiles {
		found := false
		for _, file := range files {
			if file.Name() == expected {
				found = true
			}
		}
		if !found {
			t.Errorf("did not find expected file %q in list of generated certs", expected)
		}
	}
}

func TestLoadBalancedNamesInMasterCert(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.Etcd = NodeGroup{}
	p.Worker = NodeGroup{}

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	if err := pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("error generating cluster certs: %v", err)
	}

	// Verify master node has load balanced name and FQDN
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "master.pem")
	cert := mustReadCertFile(certFile, t)

	found := false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedFQDN {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("load balanced FQDN was not found in master certificate")
	}

	found = false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedShortName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("load balanced name was not found in master certificate")
	}
}

func TestLoadBalancedNamesNotInWorkerCert(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.Etcd = NodeGroup{}

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	if err := pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("error generating cluster certs: %v", err)
	}

	// Verify master node has load balanced name and FQDN
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "worker.pem")
	cert := mustReadCertFile(certFile, t)

	found := false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedFQDN {
			found = true
			break
		}
	}
	if found {
		t.Errorf("load balanced FQDN was found in worker certificate")
	}

	found = false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedShortName {
			found = true
			break
		}
	}
	if found {
		t.Errorf("load balanced name was found in worker certificate")
	}
}

func TestLoadBalancedNamesNotInEtcdCert(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.Worker = NodeGroup{}

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Verify master node has load balanced name and FQDN
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "etcd.pem")
	cert := mustReadCertFile(certFile, t)

	found := false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedFQDN {
			found = true
			break
		}
	}
	if found {
		t.Errorf("load balanced FQDN was found in etcd certificate")
	}

	found = false
	for _, name := range cert.DNSNames {
		if name == p.Master.LoadBalancedShortName {
			found = true
			break
		}
	}
	if found {
		t.Errorf("load balanced name was found in etcd certificate")
	}
}

func TestGenerateClusterCertificatesUserSubject(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.Worker = NodeGroup{}

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "admin.pem")
	userCert := mustReadCertFile(certFile, t)

	if userCert.Subject.CommonName != users[0] {
		t.Errorf("common name mismatch: got %q, expected %q", userCert.Subject.CommonName, users[0])
	}
}

func TestGenerateClusterCertificatesDockerCert(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.Worker = NodeGroup{}
	p.DockerRegistry.SetupInternal = true

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "docker.pem")
	mustReadCertFile(certFile, t)
}

func TestBadNodeCertificate(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	p.Master.Nodes[0] = Node{
		Host:       "master",
		IP:         "11.12.13.14",
		InternalIP: "22.33.44.55",
	}

	fmt.Println("Regenerating Certs")
	err = pki.GenerateClusterCertificates(p, ca, users)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestBadUserCertificate(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	// Rename file to simulate bad cert
	os.Rename(path.Join(pki.GeneratedCertsDirectory, "admin.pem"), path.Join(pki.GeneratedCertsDirectory, "user.pem"))
	os.Rename(path.Join(pki.GeneratedCertsDirectory, "admin-key.pem"), path.Join(pki.GeneratedCertsDirectory, "user-key.pem"))

	fmt.Println("Regenerating Certs")
	err = pki.GenerateClusterCertificates(p, ca, []string{"user"})
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestBadServiceAccountCertificate(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	file := path.Join(pki.GeneratedCertsDirectory, "service-account.pem")
	os.Remove(file)
	os.Create(file)

	fmt.Println("Regenerating Certs")
	err = pki.GenerateClusterCertificates(p, ca, users)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestBadDockerCertificate(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.DockerRegistry.SetupInternal = true

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	users := []string{"admin"}
	if err = pki.GenerateClusterCertificates(p, ca, users); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	p.Master.Nodes[0] = Node{
		Host:       "master",
		IP:         "11.12.13.14",
		InternalIP: "22.33.44.55",
	}
	err = pki.generateDockerRegistryCert(p, ca)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestGenerateNodeCertNoEmptyDNSNames(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.DockerRegistry.SetupInternal = true

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	node := Node{
		Host: "master",
		IP:   "11.12.13.14",
	}
	if err := pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certificate for node: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "master.pem")
	cert := mustReadCertFile(certFile, t)
	for _, name := range cert.DNSNames {
		if name == "" {
			t.Errorf("found an empty DNS name")
		}
	}
}

func TestGenerateNodeCertWithInternalIP(t *testing.T) {
	pki := getPKI(t)
	pki.Log = os.Stdout
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.DockerRegistry.SetupInternal = true

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	node := Node{
		Host:       "master",
		IP:         "11.12.13.14",
		InternalIP: "22.33.44.55",
	}
	if err := pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certificate for node: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "master.pem")
	cert := mustReadCertFile(certFile, t)
	found := false
	internalIP := net.ParseIP(node.InternalIP)
	for _, ip := range cert.IPAddresses {
		if ip.Equal(internalIP) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Node certificate does not have the internal IP as a DNS name")
	}
}
