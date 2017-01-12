package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	libmachine "github.com/docker/machine/libmachine/ssh"
	"golang.org/x/crypto/ssh"
)

type Client interface {
	libmachine.Client
}

// Connects to ip:port as user with key and immediately exits.
func TestConnection(ip string, port int, user, key string) error {
	client, error := libmachine.NewClient(user, ip, port,
		&libmachine.Auth{
			Keys: []string{key},
		})
	if error != nil {
		return error
	}

	return client.Shell("exit")
}

func OpenConnection(ip string, port int, user, key string) (Client, error) {
	client, error := libmachine.NewClient(user, ip, port,
		&libmachine.Auth{
			Keys: []string{key},
		})
	return client, error
}

// ValidUnecryptedPrivateKey parses SSH private key
func ValidUnencryptedPrivateKey(file string) error {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	isEncrypted, err := isEncrypted(buffer)
	if err != nil {
		return fmt.Errorf("Parse SSH key error")
	}

	if isEncrypted {
		return fmt.Errorf("Encrypted SSH key is not permitted")
	}

	_, err = ssh.ParsePrivateKey(buffer)
	if err != nil {
		return fmt.Errorf("Parse SSH key error: %v", err)
	}

	return nil
}

func isEncrypted(buffer []byte) (bool, error) {
	// There is no error, just a nil block
	block, _ := pem.Decode(buffer)
	// File cannot be decoded, maybe it's some unexpected format
	if block == nil {
		return false, fmt.Errorf("Parse SSH key error")
	}

	return x509.IsEncryptedPEMBlock(block), nil
}
