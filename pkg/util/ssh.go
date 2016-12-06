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

	isEncrypted, err := isEncrypted(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse SSH key error")
	}

	if isEncrypted {
		return nil, fmt.Errorf("Encrypted SSH key is not permitted")
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse SSH key error: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}

func isEncrypted(buffer []byte) (bool, error) {
	// There is no error, just a nil block
	block, _ := pem.Decode(buffer)
	// File cannot be decoded, maybe it's some unecpected format
	if block == nil {
		return false, fmt.Errorf("Parse SSH key error")
	}

	return x509.IsEncryptedPEMBlock(block), nil
}
