package http

import (
	"bytes"
	"time"

	"github.com/apprenda/kismatic/pkg/tls"
	"github.com/cloudflare/cfssl/csr"
)

const selfSignedCSR = `
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Troy",
      "O": "Kubernetes",
      "OU": "CA",
      "ST": "New York"
    }
  ]
}
`

const expiry = 8760 * time.Hour

func selfSignedCert() (key, cert []byte, err error) {
	caKey, caCert, err := tls.NewCACertFromReader(bytes.NewBufferString(selfSignedCSR), "kismatic", expiry.String())
	if err != nil {
		return nil, nil, err
	}
	certHosts := []string{"127.0.0.1", "localhost", "0.0.0.0"}
	req := csr.CertificateRequest{
		CN:    "localhost",
		Hosts: certHosts,
		KeyRequest: &csr.BasicKeyRequest{
			A: "rsa",
			S: 2048,
		},
	}
	ca := &tls.CA{
		Key:  caKey,
		Cert: caCert,
	}
	if err != nil {
		return nil, nil, err
	}
	return tls.NewCert(ca, req, 8760*time.Hour)
}
