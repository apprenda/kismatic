package util

import (
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

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("Parse PK error: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}
