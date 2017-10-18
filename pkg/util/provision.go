package util

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	garbler "github.com/michaelbironneau/garbler/lib"
	"golang.org/x/crypto/ssh"
)

func MakeFileAskOnOverwrite(name string) (*os.File, error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return os.Create(name)
	} else {
		prompt := fmt.Sprintf("Existing file with name %v, Overwrite?", name)
		if AskForConfirmation(prompt) {
			truncateErr := os.Truncate(name, 0)
			if truncateErr != nil {
				return nil, truncateErr
			}

			return os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0666)
		} else {
			return nil, errors.New(fmt.Sprintf("existing file %v cannot be overwritten", name))
		}
	}
}

func AskForConfirmation(prompt string) bool {
	fmt.Printf("%v  (y/N):", prompt)
	var response string
	i, err := fmt.Scanln(&response)
	if err != nil {
		if i == 0 {
			response = ""
		} else {
			log.Fatal(err)
		}
	}
	okayResponseSet := MakeStringSet([]string{"y", "yes"})
	nokayResponseSet := MakeStringSet([]string{"n", "no"})
	response = strings.ToLower(response)
	if len(response) == 0 || StringSetContains(nokayResponseSet, response) {
		return false
	} else if StringSetContains(okayResponseSet, response) {
		return true
	} else {
		return AskForConfirmation(prompt)
	}
}

func StringSetContains(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

func MakeStringSet(slice []string) map[string]struct{} {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	return set
}

func MakeUniqueFile(name string, suffix string, count int) (*os.File, error) {
	var filename string

	if count > 0 {
		filename = name + "-" + strconv.Itoa(count)
	} else {
		filename = name
	}

	filename = filename + suffix

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	} else {
		return MakeUniqueFile(name, suffix, count+1)
	}
}

func GenerateAlphaNumericPassword() string {
	attempts := 0
	for {
		reqs := &garbler.PasswordStrengthRequirements{
			MinimumTotalLength: 16,
			Uppercase:          rand.Intn(6),
			Digits:             rand.Intn(6),
			Punctuation:        -1, // disable punctuation
		}
		pass, err := garbler.NewPassword(reqs)
		if err != nil {
			return "weakpassword"
		}
		// validate that the library actually returned an alphanumeric password
		re := regexp.MustCompile("^[a-zA-Z1-9]+$")
		if re.MatchString(pass) {
			return pass
		}
		if attempts == 50 {
			return "weakpassword"
		}
		attempts++
	}
}

func IncrementIPv4(ip net.IP) (net.IP, error) {

	if len(ip) != net.IPv4len {
		return nil, errors.New("IncrementIPv4: only IPv4 addresses supported")
	}

	if ip.Equal(net.IPv4bcast) {
		return nil, errors.New("IncrementIPv4: incrementing IPv4 broadcast (255.255.255.255) results in overflow")
	}

	nextIp := make(net.IP, len(ip))
	copy(nextIp, ip)

	// increment IP accounting for rollovers at each byte
	for j := len(nextIp) - 1; j >= 0; j-- {
		nextIp[j]++
		if nextIp[j] > 0 {
			break
		}
	}

	return nextIp, nil
}

func BroadcastIPv4(network net.IPNet) (net.IP, error) {
	if len(network.IP) != net.IPv4len {
		return nil, errors.New("BroadcastIPv4: only IPv4 addresses supported")
	}

	broadcast := net.IP(make([]byte, 4))
	for i := range network.IP {
		broadcast[i] = network.IP[i] | ^network.Mask[i]
	}

	return broadcast, nil
}

func LoadOrCreatePrivateSSHKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	blockType := "RSA PRIVATE KEY"

	if _, statErr := os.Stat(privateKeyPath); os.IsNotExist(statErr) {
		privateKey, generateErr := rsa.GenerateKey(cryptorand.Reader, 1024)
		if generateErr != nil {
			return nil, generateErr
		}

		// generate and write private key as PEM
		privateKeyFile, createErr := os.Create(privateKeyPath)
		defer privateKeyFile.Close()
		if createErr != nil {
			return nil, createErr
		}

		privateKeyPEM := &pem.Block{Type: blockType, Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
		if encodeErr := pem.Encode(privateKeyFile, privateKeyPEM); encodeErr != nil {
			return nil, encodeErr
		}

		return privateKey, nil
	} else {
		buffer, readErr := ioutil.ReadFile(privateKeyPath)
		if readErr != nil {
			return nil, readErr
		}

		block, rest := pem.Decode(buffer)
		if len(rest) > 0 {
			return nil, errors.New("LoadOrCreatePrivateSSHKey: extra data in private key PEM block")
		}

		if block.Type != blockType {
			return nil, errors.New(fmt.Sprintf("LoadOrCreatePrivateSSHKey: expecting a block type of %v but got %v", blockType, block.Type))
		}

		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
}

func CreatePublicKey(privateKey *rsa.PrivateKey, publicKeyPath string) error {
	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(publicKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
}
