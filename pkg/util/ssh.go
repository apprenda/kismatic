package util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// GetUnencryptedPublicKeyAuth parses SSH private key and returns PublicKeys AuthMethod
func GetUnencryptedPublicKeyAuth(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	isEcnrypted, err := isEncrypted(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse SHH key error")
	}

	if isEcnrypted {
		return nil, fmt.Errorf("Encrypted SSH key is not permitted")
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse SHH key error: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}

func isEncrypted(buffer []byte) (bool, error) {
	block, err := pem.Decode(buffer)
	// File cannot be decoded, maybe it's some unecpected format
	if block == nil || err != nil {
		return false, fmt.Errorf("Parse SHH key error")
	}

	return x509.IsEncryptedPEMBlock(block), nil
}
