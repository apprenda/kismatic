package install

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/cloudflare/cfssl/helpers"
)

func getPKI(t *testing.T) LocalPKI {
	tempDir, err := ioutil.TempDir("", "pki-tests")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	pki := LocalPKI{
		CACsr: "test/ca-csr.json",
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
		AddOns: AddOns{
			CNI: &CNI{},
		},
		Etcd: NodeGroup{
			Nodes: []Node{
				Node{
					Host:       "etcd01",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
				Node{
					Host:       "etcd02",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
		Master: MasterNodeGroup{
			Nodes: []Node{
				Node{
					Host:       "master01",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
				Node{
					Host:       "master02",
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
					Host:       "worker01",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
				Node{
					Host:       "worker02",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
		Ingress: OptionalNodeGroup{
			Nodes: []Node{
				Node{
					Host:       "ingress01",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
				Node{
					Host:       "ingress02",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
		Storage: OptionalNodeGroup{
			Nodes: []Node{
				Node{
					Host:       "storage01",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
				Node{
					Host:       "storage02",
					IP:         "99.99.99.99",
					InternalIP: "88.88.88.88",
				},
			},
		},
	}
}

func TestGeneratedClusterCACommonNameMatchesClusterName(t *testing.T) {
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

func TestGenerateClusterCAPlanFileExpirationIsRespected(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	validity := 5 * 365 * 24 * time.Hour // 5 years
	p.Cluster.Certificates.Expiry = validity.String()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	caCert, err := helpers.ParseCertificatePEM(ca.Cert)
	if err != nil {
		t.Errorf("failed to parse generated cert: %v", err)
	}

	expirationDate := time.Now().Add(validity)
	if caCert.NotAfter.Year() != expirationDate.Year() || caCert.NotAfter.YearDay() != expirationDate.YearDay() {
		t.Errorf("bad expiration date on generated cert. expected %v, got %v", expirationDate, caCert.NotAfter)
	}
}

func TestGenerateClusterCertificatesExistingCertsAreNotRegen(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateClusterCertificates(p, ca); err != nil {
		t.Fatalf("error generating cluster certificates: %v", err)
	}

	// Get the mod time of all the generated files
	files, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error listing files in generated certs dir: %v", err)
	}
	modTime := map[string]time.Time{}
	for _, f := range files {
		modTime[f.Name()] = f.ModTime()
	}

	// Run generation again. Nothing should be touched.
	if err = pki.GenerateClusterCertificates(p, ca); err != nil {
		t.Fatalf("error generating cluster certificates: %v", err)
	}

	files2, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error listing files in generated certs dir: %v", err)
	}
	modTime2 := map[string]time.Time{}
	for _, f := range files2 {
		modTime2[f.Name()] = f.ModTime()
	}

	for k := range modTime {
		if modTime[k] != modTime2[k] {
			t.Errorf("file %s was modified. modification time changed from %v to %v", k, modTime[k], modTime2[k])
		}
	}
}

func TestNodeCertExistsSkipGeneration(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	node := p.Master.Nodes[0]

	// Create the node cert and key file
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Get the mod time of all the generated files
	files, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error listing files in generated certs dir: %v", err)
	}
	modTime := map[string]time.Time{}
	for _, f := range files {
		modTime[f.Name()] = f.ModTime()
	}

	// Run generation again
	if err = pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Assert files did not change
	files2, err := ioutil.ReadDir(pki.GeneratedCertsDirectory)
	if err != nil {
		t.Fatalf("error listing files in generated certs dir: %v", err)
	}
	modTime2 := map[string]time.Time{}
	for _, f := range files2 {
		modTime2[f.Name()] = f.ModTime()
	}

	for k := range modTime {
		if modTime[k] != modTime2[k] {
			t.Errorf("file %s was modified. modification time changed from %v to %v", k, modTime[k], modTime2[k])
		}
	}
}

func TestGenerateClusterCertificatesValidateCertificateInformation(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("failed to generate cluster CA")
	}
	etcdNode := p.Etcd.Nodes[0]
	masterNode := p.Master.Nodes[0]
	workerNode := p.Worker.Nodes[0]
	ingressNode := p.Ingress.Nodes[0]
	storageNode := p.Storage.Nodes[0]

	// Generate the cluster certificates
	err = pki.GenerateClusterCertificates(p, ca)
	if err != nil {
		t.Fatalf("failed to generate cluster certificates")
	}

	t.Run("etcd node: etcd server certificate", func(t *testing.T) {
		certFilename := fmt.Sprintf("%s-etcd.pem", etcdNode.Host)
		cert := mustReadCertFile(filepath.Join(pki.GeneratedCertsDirectory, certFilename), t)

		if cert.Subject.CommonName != etcdNode.Host {
			t.Errorf("expected common name %q but got %q", etcdNode.Host, cert.Subject.CommonName)
		}

		var found bool
		for _, n := range cert.DNSNames {
			if n == etcdNode.Host {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("did not find node's hostname %q in certificate DNS names %v", etcdNode.Host, cert.DNSNames)
		}

		found = false
		internalFound := false
		nodeIP := net.ParseIP(etcdNode.IP)
		internalIP := net.ParseIP(etcdNode.IP)
		for _, ip := range cert.IPAddresses {
			if ip.Equal(nodeIP) {
				found = true
			}
			if ip.Equal(internalIP) {
				internalFound = true
			}
		}
		if !found {
			t.Errorf("did not find node's IP address %q in cert's IP addresses %v", etcdNode.IP, cert.IPAddresses)
		}
		if !internalFound {
			t.Errorf("did not find node's internal IP address %q in cert's IP address %v", etcdNode.InternalIP, cert.IPAddresses)
		}
	})

	// Verify the API server certificate
	t.Run("master node: api server certificate", func(t *testing.T) {
		certFilename := fmt.Sprintf("%s-apiserver.pem", masterNode.Host)
		cert := mustReadCertFile(filepath.Join(pki.GeneratedCertsDirectory, certFilename), t)

		// The common name should match the hostname
		if cert.Subject.CommonName != masterNode.Host {
			t.Errorf("expected common name: %q, but got %q", masterNode.Host, cert.Subject.CommonName)
		}

		// DNS names should contain node's hostname and load balanced names
		for _, expected := range []string{masterNode.Host, p.Master.LoadBalancedFQDN, p.Master.LoadBalancedShortName} {
			var found bool
			for _, n := range cert.DNSNames {
				if n == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("did not find name %q in the certificates DNS names: %v", masterNode.Host, cert.DNSNames)
			}
		}

		// Node's ip  and private ip should be in the cert
		found := false
		internalFound := false
		nodeIP := net.ParseIP(masterNode.IP)
		internalIP := net.ParseIP(masterNode.InternalIP)
		for _, ip := range cert.IPAddresses {
			if ip.Equal(nodeIP) {
				found = true
			}
			if ip.Equal(internalIP) {
				internalFound = true
			}
		}
		if !found {
			t.Errorf("did not find node's IP %q in the certificate's IPs: %v", masterNode.IP, cert.IPAddresses)
		}
		if !internalFound {
			t.Errorf("did not find node's internal IP %q in the certificate's IPs: %v", masterNode.InternalIP, cert.IPAddresses)
		}
	})

	// Validate API server client certificates
	tests := []struct {
		name                  string
		certFilename          string
		expectedCommonName    string
		expectedOrganizations []string
	}{
		{
			name:               "kube scheduler certificate",
			certFilename:       "kube-scheduler.pem",
			expectedCommonName: "system:kube-scheduler",
		},
		{
			name:               "kube controller mgr certificate",
			certFilename:       "kube-controller-manager.pem",
			expectedCommonName: "system:kube-controller-manager",
		},
		{
			name:               "kube-proxy certificate",
			certFilename:       "kube-proxy.pem",
			expectedCommonName: "system:kube-proxy",
		},
		{
			name:                  "master node/kubelet certificate",
			certFilename:          fmt.Sprintf("%s-kubelet.pem", masterNode.Host),
			expectedCommonName:    fmt.Sprintf("system:node:%s", masterNode.Host),
			expectedOrganizations: []string{"system:nodes"},
		},
		{
			name:                  "worker node/kubelet certificate",
			certFilename:          fmt.Sprintf("%s-kubelet.pem", workerNode.Host),
			expectedCommonName:    fmt.Sprintf("system:node:%s", workerNode.Host),
			expectedOrganizations: []string{"system:nodes"},
		},
		{
			name:                  "ingress node/kubelet certificate",
			certFilename:          fmt.Sprintf("%s-kubelet.pem", ingressNode.Host),
			expectedCommonName:    fmt.Sprintf("system:node:%s", ingressNode.Host),
			expectedOrganizations: []string{"system:nodes"},
		},
		{
			name:                  "storage node/kubelet certificate",
			certFilename:          fmt.Sprintf("%s-kubelet.pem", storageNode.Host),
			expectedCommonName:    fmt.Sprintf("system:node:%s", storageNode.Host),
			expectedOrganizations: []string{"system:nodes"},
		},
		{
			name:                  "admin user certificate",
			certFilename:          "admin.pem",
			expectedCommonName:    "admin",
			expectedOrganizations: []string{"system:masters"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, validateClientCertificateAndKey(pki.GeneratedCertsDirectory,
			test.certFilename, test.expectedCommonName, test.expectedOrganizations...))
	}
}

func TestGenerateClusterCertificatesPlanFileExpirationIsRespected(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	validity := 5 * 365 * 24 * time.Hour // 5 years
	p.Cluster.Certificates.Expiry = validity.String()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	node := p.Master.Nodes[0]
	if err := pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certificate for node: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, fmt.Sprintf("%s-apiserver.pem", node.Host))
	cert := mustReadCertFile(certFile, t)

	expirationDate := time.Now().Add(validity)
	if cert.NotAfter.Year() != expirationDate.Year() || cert.NotAfter.YearDay() != expirationDate.YearDay() {
		t.Errorf("bad expiration date on generated cert. expected %v, got %v", expirationDate, cert.NotAfter)
	}
}

func validateClientCertificateAndKey(certsDir, filename, expectedCommonName string, expectedOrganizations ...string) func(t *testing.T) {
	return func(t *testing.T) {
		cert := mustReadCertFile(filepath.Join(certsDir, filename), t)
		if expectedCommonName != cert.Subject.CommonName {
			t.Errorf("Expected common name %q but got %q", expectedCommonName, cert.Subject.CommonName)
		}

		if !reflect.DeepEqual(cert.Subject.Organization, expectedOrganizations) {
			t.Errorf("Expected organizations: %v, but got %v", expectedOrganizations, cert.Subject.Organization)
		}
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
	if err = pki.GenerateNodeCertificate(p, p.Etcd.Nodes[0], ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Verify master node has load balanced name and FQDN
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "etcd01-etcd.pem")
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

func TestContivProxyServerCertGenerated(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	p.AddOns.CNI = &CNI{Provider: cniProviderContiv}

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateClusterCertificates(p, ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, "contiv-proxy-server.pem")
	mustReadCertFile(certFile, t)
}

func TestInvalidNodeCertificateShouldFailValidation(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateNodeCertificate(p, p.Master.Nodes[0], ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	// Change the IP so that the IPs don't match the valid ones
	p.Master.Nodes[0] = Node{
		Host:       "master01",
		IP:         "11.12.13.14",
		InternalIP: "22.33.44.55",
	}

	err = pki.GenerateClusterCertificates(p, ca)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestAPIServerCertNoEmptyDNSNames(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	node := p.Master.Nodes[0]
	if err := pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certificate for node: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, fmt.Sprintf("%s-apiserver.pem", node.Host))
	cert := mustReadCertFile(certFile, t)
	for _, name := range cert.DNSNames {
		if name == "" {
			t.Errorf("found an empty DNS name")
		}
	}
}

func TestAPIServerCertContainsInternalIP(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()
	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	node := p.Master.Nodes[0]
	if err := pki.GenerateNodeCertificate(p, node, ca); err != nil {
		t.Fatalf("failed to generate certificate for node: %v", err)
	}
	certFile := filepath.Join(pki.GeneratedCertsDirectory, fmt.Sprintf("%s-apiserver.pem", node.Host))
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

func TestValidateClusterCertificatesNoExistingCerts(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	warn, err := pki.ValidateClusterCertificates(getPlan())
	if len(err) != 0 {
		t.Errorf("expected no errors when validating directory with no certificates, but got: %v", err)
	}
	if len(warn) != 0 {
		t.Errorf("expected no warnings when validating directory with no certificates, but got: %v", warn)
	}
}

func TestValidateClusterCertificatesWithValidExistingCerts(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateClusterCertificates(p, ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	warn, errs := pki.ValidateClusterCertificates(p)
	if len(errs) != 0 {
		t.Errorf("expected no errors when validating certs that are valid, but got: %v", err)
	}
	if len(warn) != 0 {
		t.Errorf("expected no warnings when validating certs that are valid, but got: %v", warn)
	}
}

func TestValidateClusterCertificatesInvalidCerts(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	p := getPlan()

	ca, err := pki.GenerateClusterCA(p)
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}
	if err = pki.GenerateClusterCertificates(p, ca); err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}

	tests := []struct {
		description      string
		plan             func(Plan) Plan
		expectedWarnings int
	}{
		{
			description: "bad etcd certificate",
			plan: func(p Plan) Plan {
				etcd := p.Etcd.Nodes[0]
				etcd.IP = "20.0.0.1"
				p.Etcd.Nodes = []Node{etcd}
				return p
			},
			expectedWarnings: 1,
		},
		{
			description: "bad master certificates",
			plan: func(p Plan) Plan {
				master := p.Master.Nodes[0]
				master.IP = "20.0.0.1"
				p.Master.Nodes = []Node{master}
				return p
			},
			expectedWarnings: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			plan := test.plan(*p)
			warn, errs := pki.ValidateClusterCertificates(&plan)
			if len(errs) != 0 {
				t.Fatalf("unexpected error occurred validating certs: %v", err)
			}
			if len(warn) != test.expectedWarnings {
				t.Errorf("expected %d warnings, but got %d. warnings were: %v", test.expectedWarnings, len(warn), warn)
			}
		})
	}
}

func TestCertSpecEqual(t *testing.T) {
	tests := []struct {
		x     certificateSpec
		y     certificateSpec
		equal bool
	}{
		{
			x:     certificateSpec{},
			y:     certificateSpec{},
			equal: true,
		},
		{
			x: certificateSpec{
				description: "foo",
			},
			y: certificateSpec{
				description: "foo",
			},
			equal: true,
		},
		{
			x: certificateSpec{
				description: "foo",
			},
			y: certificateSpec{
				description: "bar",
			},
			equal: false,
		},
		{
			x: certificateSpec{
				filename: "foo",
			},
			y: certificateSpec{
				filename: "bar",
			},
			equal: false,
		},
		{
			x: certificateSpec{
				filename: "foo",
			},
			y: certificateSpec{
				filename: "foo",
			},
			equal: true,
		},
		{
			x: certificateSpec{
				commonName: "foo",
			},
			y: certificateSpec{
				commonName: "bar",
			},
			equal: false,
		},
		{
			x: certificateSpec{
				subjectAlternateNames: []string{"foo"},
			},
			y: certificateSpec{
				subjectAlternateNames: []string{"foo"},
			},
			equal: true,
		},
		{
			x: certificateSpec{
				subjectAlternateNames: []string{"foo"},
			},
			y: certificateSpec{
				subjectAlternateNames: []string{"bar"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				organizations: []string{"foo"},
			},
			y: certificateSpec{
				organizations: []string{"foo"},
			},
			equal: true,
		},
		{
			x: certificateSpec{
				organizations: []string{"foo"},
			},
			y: certificateSpec{
				organizations: []string{"bar"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			equal: true,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "bar",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "bar",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"bar"},
				organizations:         []string{"foo"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"bar"},
			},
			equal: false,
		},
		{
			x: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo"},
				organizations:         []string{"foo"},
			},
			y: certificateSpec{
				description:           "foo",
				filename:              "foo",
				commonName:            "foo",
				subjectAlternateNames: []string{"foo", "bar"},
				organizations:         []string{"foo"},
			},
			equal: false,
		},
	}

	for _, test := range tests {
		if test.x.equal(test.y) != test.equal {
			t.Errorf("expected equal = %v, but got %v. x = %+v, y = %+v", test.equal, !test.equal, test.x, test.y)
		}
		if test.y.equal(test.x) != test.equal {
			t.Errorf("expected equal = %v, but got %v. x = %+v, y = %+v", test.equal, !test.equal, test.x, test.y)
		}
	}
}

func TestGenerateCertificate(t *testing.T) {
	pki := getPKI(t)
	defer cleanup(pki.GeneratedCertsDirectory, t)

	ca, err := pki.GenerateClusterCA(getPlan())
	if err != nil {
		t.Fatalf("error generating CA for test: %v", err)
	}

	tests := []struct {
		name                  string
		validityPeriod        string
		commonName            string
		subjectAlternateNames []string
		organizations         []string
		ca                    *tls.CA
		overwrite             bool
		exists                bool
		valid                 bool
	}{
		{
			name:           "foo",
			commonName:     "foo",
			validityPeriod: "8650h",
			ca:             ca,
			valid:          true,
		},
		{
			name:                  "bar",
			validityPeriod:        "17300h",
			commonName:            "bar-cn",
			subjectAlternateNames: []string{"bar-alt"},
			organizations:         []string{"admin"},
			ca:                    ca,
			valid:                 true,
		},
		{
			name:           "foo",
			commonName:     "foo",
			validityPeriod: "8650h",
			ca:             ca,
			exists:         true,
			valid:          true,
		},
		{
			name:                  "foo",
			validityPeriod:        "8650h",
			commonName:            "foo-cn",
			subjectAlternateNames: []string{"foo-alt"},
			organizations:         []string{"admin"},
			ca:                    ca,
			overwrite:             true,
			exists:                true,
			valid:                 true,
		},
		{
			name:  "alice",
			ca:    ca,
			valid: false,
		},
		{
			validityPeriod: "8650h",
			ca:             ca,
			valid:          false,
		},
		{
			name:           "alice",
			validityPeriod: "8650h",
			valid:          false,
		},
	}
	for i, test := range tests {
		exists, err := pki.GenerateCertificate(test.name, test.validityPeriod, test.commonName, test.subjectAlternateNames, test.organizations, test.ca, test.overwrite)

		if (err != nil) == test.valid {
			t.Errorf("test %d: expect valid to be %t, but got %v", i, test.valid, err)
		}
		if exists != test.exists {
			t.Errorf("test %d: expect exists to be %t, but got %t", i, test.exists, exists)
		}

		if test.valid {
			cert := mustReadCertFile(filepath.Join(pki.GeneratedCertsDirectory, fmt.Sprintf("%s.pem", test.name)), t)
			if cert.Subject.CommonName != test.commonName {
				t.Errorf("test %d: expect commonName to be %s, but got %s", i, test.commonName, cert.Subject.CommonName)
			}
			if !util.Subset(cert.DNSNames, test.subjectAlternateNames) {
				t.Errorf("test %d: expect subjectAlternateNames to be %v, but got %v", i, test.subjectAlternateNames, cert.DNSNames)
			}
			if !util.Subset(cert.Subject.Organization, test.organizations) {
				t.Errorf("test %d: expect organizations to be %v, but got %v", i, test.organizations, cert.Subject.Organization)
			}
			if test.validityPeriod != "" {

			}
			validity, err := strconv.Atoi(strings.TrimRight(test.validityPeriod, "h"))
			if err != nil {
				t.Errorf("test %d: could not parse validityPeriod %s", i, test.validityPeriod)
			} else {
				expirationDate := time.Now().Add(time.Duration(validity) * time.Hour)
				if cert.NotAfter.Year() != expirationDate.Year() || cert.NotAfter.YearDay() != expirationDate.YearDay() {
					t.Errorf("test %d: bad expiration date on generated cert. expected %v, got %v", i, expirationDate, cert.NotAfter)
				}
			}
		}
	}
}
