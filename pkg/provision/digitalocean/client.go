package digitalocean

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// Droplet
type Droplet struct {
	ID        int
	Name      string
	PrivateIP string
	PublicIP  string
	SSHUser   string
}

type NodeConfig struct {
	Image             string
	Name              string
	Region            string
	Size              string
	UserData          string
	Keys              []string
	Tags              []string
	PrivateNetworking bool
}

type KeyConfig struct {
	ID            int
	Name          string
	PublicKeyFile string
	Fingerprint   string
}

// Client for provisioning machines on AWS
type Client struct {
	doClient *godo.Client
}

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func (c *Client) getAPIClient(token string) (*godo.Client, error) {
	if c.doClient == nil {
		tokenSource := &TokenSource{
			AccessToken: token,
		}
		oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
		c.doClient = godo.NewClient(oauthClient)
	}
	return c.doClient, nil
}

func (c Client) GetDroplet(token string, dropletID int) (Droplet, error) {
	drop := Droplet{}
	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return drop, err
	}

	ctx := context.TODO()

	newDroplet, _, errhost := client.Droplets.Get(ctx, dropletID)

	if errhost != nil {
		fmt.Println("Cannot create host", errhost)
		return drop, errhost
	}
	drop.ID = newDroplet.ID
	drop.Name = newDroplet.Name
	if newDroplet.Networks.V4 != nil {
		for i := 0; i < len(newDroplet.Networks.V4); i++ {
			if newDroplet.Networks.V4[i].Type == "public" {
				drop.PublicIP = newDroplet.Networks.V4[i].IPAddress
			}
			if newDroplet.Networks.V4[i].Type == "private" {
				drop.PrivateIP = newDroplet.Networks.V4[i].IPAddress
			}
		}
	}

	return drop, nil

}

func (c Client) CreateNode(token string, config NodeConfig, keyconfig KeyConfig) (Droplet, error) {
	drop := Droplet{}
	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return drop, err
	}

	sshKey := godo.DropletCreateSSHKey{
		ID:          keyconfig.ID,
		Fingerprint: keyconfig.Fingerprint,
	}
	var keys []godo.DropletCreateSSHKey
	keys = append(keys, sshKey)
	createRequest := &godo.DropletCreateRequest{
		Name:   config.Name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		UserData:          config.UserData,
		Tags:              config.Tags,
		SSHKeys:           keys,
		PrivateNetworking: config.PrivateNetworking,
	}

	ctx := context.TODO()

	newDroplet, _, errhost := client.Droplets.Create(ctx, createRequest)

	if errhost != nil {
		fmt.Println("Cannot create host", errhost)
		return drop, errhost
	}

	drop.ID = newDroplet.ID
	drop.Name = newDroplet.Name

	return drop, nil
}

func (c Client) CreateKey(token string, config KeyConfig) (KeyConfig, error) {
	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return config, err
	}

	key, keyerr := ioutil.ReadFile(config.PublicKeyFile)
	if keyerr != nil {
		fmt.Println("Cannot read public key file", keyerr)
		return config, keyerr
	}

	ctx := context.TODO()

	keyRequest := &godo.KeyCreateRequest{
		Name:      config.Name,
		PublicKey: string(key),
	}

	keyObj, _, errreq := client.Keys.Create(ctx, keyRequest)

	if errreq != nil {
		fmt.Println("Cannot create public key", errreq)
		return config, errreq
	}

	config.ID = keyObj.ID
	config.Fingerprint = keyObj.Fingerprint
	return config, nil
}

func (c Client) FindKeyByName(token string, keyName string) (KeyConfig, error) {
	config := KeyConfig{}
	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return config, err
	}
	ctx := context.TODO()
	opts := &godo.ListOptions{}
	keys, _, err := client.Keys.List(ctx, opts)
	if err != nil {
		fmt.Println("Cannot load keys", err)
		return config, err
	}
	for i := 0; i < len(keys); i++ {

		if keys[i].Name == keyName {
			config.ID = keys[i].ID
			config.Fingerprint = keys[i].Fingerprint
			break
		}
	}

	return config, nil
}

func (c Client) DeleteKeyByName(token string, keyName string) error {
	config := KeyConfig{}
	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return err
	}
	ctx := context.TODO()
	opts := &godo.ListOptions{}
	keys, _, err := client.Keys.List(ctx, opts)
	if err != nil {
		fmt.Println("Cannot load keys", err)
		return err
	}
	for i := 0; i < len(keys); i++ {

		if keys[i].Name == keyName {
			fmt.Println("Key found")
			config.ID = keys[i].ID
			config.Fingerprint = keys[i].Fingerprint
			break
		}
	}

	fmt.Println("Deleting ssh key ", keyName)
	if config.Fingerprint != "" {
		_, delerr := client.Keys.DeleteByFingerprint(ctx, config.Fingerprint)
		if delerr != nil {
			return delerr
		}
	}

	return nil
}

func (c Client) DeleteDropletsByTag(token string, tag string, keyname string) error {

	client, err := c.getAPIClient(token)
	if err != nil {
		fmt.Println("Cannot get api object", err)
		return err
	}
	ctx := context.TODO()

	fmt.Println("Deleting droplets with tag ", tag)
	_, errdel := client.Droplets.DeleteByTag(ctx, tag)

	if keyname != "" {
		c.DeleteKeyByName(token, keyname)
	}
	return errdel
}
