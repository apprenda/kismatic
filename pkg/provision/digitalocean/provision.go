package digitalocean

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/apprenda/kismatic/pkg/provision/plan"
)

const (
	SSHKEY          = "apprenda-key"
	KET_INSTALL_DIR = "/ket"
)

type infrastructureProvisioner interface {
	ProvisionNodes(NodeCount, LinuxDistro) (ProvisionedNodes, error)

	TerminateNodes(ProvisionedNodes) error

	TerminateAllNodes() error

	SSHKey() string
}

type LinuxDistro string

type NodeCount struct {
	Etcd     uint16
	Master   uint16
	Worker   uint16
	Boostrap uint16
}

func (nc NodeCount) Total() uint16 {
	return nc.Etcd + nc.Master + nc.Worker
}

type ProvisionedNodes struct {
	Etcd     []plan.Node
	Master   []plan.Node
	Worker   []plan.Node
	Boostrap []plan.Node
}

func (p ProvisionedNodes) allNodes() []plan.Node {
	n := []plan.Node{}
	n = append(n, p.Etcd...)
	n = append(n, p.Master...)
	n = append(n, p.Worker...)
	n = append(n, p.Boostrap...)
	return n
}

type sshMachineProvisioner struct {
	sshKey string
}

func (p sshMachineProvisioner) SSHKey() string {
	return p.sshKey
}

type doProvisioner struct {
	sshMachineProvisioner
	client *Client
}

func GetProvisioner() (*doProvisioner, bool) {
	c := Client{}
	p := doProvisioner{client: &c}
	return &p, true
}

func dropletToNode(drop *Droplet, opts *DOOpts) plan.Node {
	node := plan.Node{}
	node.ID = string(drop.ID)
	node.Host = drop.Name
	node.PublicIPv4 = drop.PublicIP
	node.PrivateIPv4 = drop.PrivateIP
	node.SSHUser = opts.SSHUser
	return node
}

func optionsToConfig(opts *DOOpts, name string, sizeOverride string, userData string) NodeConfig {
	config := NodeConfig{}
	config.Image = opts.Image
	config.Name = name
	config.Region = opts.Region
	config.PrivateNetworking = true
	if sizeOverride != "" {
		config.Size = sizeOverride
	} else {
		config.Size = opts.InstanceType
	}
	if userData != "" {
		config.UserData = userData
	}

	if opts.ClusterTag != "" {
		config.Tags = append(config.Tags, opts.ClusterTag)
	} else {
		config.Tags = append(config.Tags, "apprenda")
	}
	return config
}

func (p doProvisioner) ProvisionNodes(opts DOOpts, nodeCount NodeCount) (ProvisionedNodes, error) {
	provisioned := ProvisionedNodes{}
	keyconf := KeyConfig{}
	keyconf.Name = SSHKEY
	keyconf.PublicKeyFile = opts.SSHPublicKey
	existing, _ := p.client.FindKeyByName(opts.Token, keyconf.Name)
	var key KeyConfig
	var errkey error
	if existing.Fingerprint != "" {
		key = existing
		fmt.Println("Using existing key", key)
	} else {
		fmt.Println("Creating new key")
		key, errkey = p.client.CreateKey(opts.Token, keyconf)
	}
	if errkey != nil {
		fmt.Println("Cannot create key", errkey)
		return provisioned, errkey
	}

	var dropletsETCD []Droplet
	var i uint16
	for i = 0; i < nodeCount.Etcd; i++ {
		config := optionsToConfig(&opts, fmt.Sprintf("etcd%d", i+1), "", "")
		drop, err := p.client.CreateNode(opts.Token, config, key)
		if err != nil {
			return provisioned, err
		}
		dropletsETCD = append(dropletsETCD, drop)
	}
	var dropletsMaster []Droplet
	for i = 0; i < nodeCount.Master; i++ {
		config := optionsToConfig(&opts, fmt.Sprintf("master%d", i+1), "", "")
		drop, err := p.client.CreateNode(opts.Token, config, key)
		if err != nil {
			return provisioned, err
		}
		dropletsMaster = append(dropletsMaster, drop)
	}
	var dropletsWorker []Droplet
	for i = 0; i < nodeCount.Worker; i++ {
		config := optionsToConfig(&opts, fmt.Sprintf("worker%d", i+1), opts.WorkerType, "")
		drop, err := p.client.CreateNode(opts.Token, config, key)
		if err != nil {
			return provisioned, err
		}
		dropletsWorker = append(dropletsWorker, drop)
	}

	var dropletsBoot []Droplet
	for i = 0; i < nodeCount.Boostrap; i++ {
		cmd := ""
		var cmderr error
		if opts.BootstrapFile != "" {
			cmd, cmderr = loadBootCmds(opts.BootstrapFile)
			if cmderr != nil {
				fmt.Println("Cannot load script file for boot init", cmderr)
			}
		}
		config := optionsToConfig(&opts, fmt.Sprintf("bootstrap%d", i+1), "", cmd)
		fmt.Println("Bootstrap node:", config)
		drop, err := p.client.CreateNode(opts.Token, config, key)
		if err != nil {
			return provisioned, err
		}
		dropletsBoot = append(dropletsBoot, drop)
	}

	//Wait for assigned IPs

	for i = 0; i < nodeCount.Etcd; i++ {
		drop := p.WaitForIPs(opts, dropletsETCD[i])
		if drop != nil {
			n := dropletToNode(drop, &opts)
			provisioned.Etcd = append(provisioned.Etcd, n)
		} else {
			return provisioned, fmt.Errorf("Unable to get IPs from %s", dropletsETCD[i].Name)
		}
	}

	for i = 0; i < nodeCount.Master; i++ {
		drop := p.WaitForIPs(opts, dropletsMaster[i])
		if drop != nil {
			n := dropletToNode(drop, &opts)
			provisioned.Master = append(provisioned.Master, n)
		} else {
			return provisioned, fmt.Errorf("Unable to get IPs from %s", dropletsMaster[i].Name)
		}
	}

	for i = 0; i < nodeCount.Worker; i++ {
		drop := p.WaitForIPs(opts, dropletsWorker[i])
		if drop != nil {
			n := dropletToNode(drop, &opts)
			provisioned.Worker = append(provisioned.Worker, n)
		} else {
			return provisioned, fmt.Errorf("Unable to get IPs from %s", dropletsWorker[i].Name)
		}
	}

	for i = 0; i < nodeCount.Boostrap; i++ {
		drop := p.WaitForIPs(opts, dropletsBoot[i])
		if drop != nil {
			n := dropletToNode(drop, &opts)
			provisioned.Boostrap = append(provisioned.Boostrap, n)
		} else {
			return provisioned, fmt.Errorf("Unable to get IPs from %s", dropletsBoot[i].Name)
		}
	}

	fmt.Println("Done provisioning")
	return provisioned, nil
}

func (p doProvisioner) WaitForIPs(opts DOOpts, drop Droplet) *Droplet {
	fmt.Printf("Waiting for IPs to be assigned for node %s\n", drop.Name)
	for {
		init, err := p.client.GetDroplet(opts.Token, drop.ID)

		if init.PublicIP != "" && err == nil {
			// command succeeded
			fmt.Printf("IP assinged to %s: Public = %s ; Private %s\n", init.Name, init.PublicIP, init.PrivateIP)
			return &init
		}
		fmt.Printf(".")
		time.Sleep(3 * time.Second)
	}
}

func (p doProvisioner) TerminateNodes(opts DOOpts) error {

	key := ""
	if opts.RemoveKey {
		key = SSHKEY
	}

	return p.client.DeleteDropletsByTag(opts.Token, opts.ClusterTag, key)
}

func WaitForSSH(ProvisionedNodes ProvisionedNodes, sshKey string) error {
	nodes := ProvisionedNodes.allNodes()
	for _, n := range nodes {
		BlockUntilSSHOpen(n.Host, n.PublicIPv4, n.SSHUser, sshKey)
	}
	fmt.Println("SSH established on all nodes")
	return nil
}

func loadBootCmds(path string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", fmt.Errorf("Cannot get path to exec %v\n", err)
	}

	cmdpath := filepath.Join(dir, path)
	cmd, errcmd := ioutil.ReadFile(cmdpath)
	if errcmd != nil {
		fmt.Println("Cannot read public boot init file", errcmd)
		return "", errcmd
	}
	s := string(cmd)

	root := os.Getenv("DO_KET_INSTALL_DIR")
	if root == "" {
		root = KET_INSTALL_DIR
	}
	initstatement := fmt.Sprintf("#!/bin/bash\nmkdir -p %s\ncd %s && ", root, root)
	s = strings.Replace(s, "#!/bin/bash", initstatement, -1)

	re := regexp.MustCompile(`\r?\n`)
	s = re.ReplaceAllString(s, "\n")
	return s, nil
}
