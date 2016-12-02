package util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// GetPublicKeyAuth parses SSH private key and returns PublicKeys AuthMethod
func GetPublicKeyAuth(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if isEncrypted(buffer) {
		return nil, fmt.Errorf("Encrypted SSH key is not permitted")
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse SHH key error: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}

func isEncrypted(buffer []byte) bool {
	fmt.Println(string(buffer))
	block, _ := pem.Decode(buffer)
	if x509.IsEncryptedPEMBlock(block) {
		return true
	}
	return false
}
