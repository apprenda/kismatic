package ansible

import (
	"bytes"
	"fmt"
)

// Inventory is a collection of Nodes, keyed by role.
type Inventory struct {
	Roles []Role
}

// Role is an Ansible role, containing nodes that belong to the role.
type Role struct {
	// Name of the role
	Name string
	// The nodes that belong to this role
	Nodes []Node
}

// Node is an Ansible target node
type Node struct {
	// Host is the hostname of the target node
	Host string
	// PublicIP is the publicly accessible IP
	PublicIP string
	// InternalIP is the internal IP, if different from PublicIP.
	InternalIP string
	// SSHPrivateKey is the private key to be used for SSH authentication
	SSHPrivateKey string
	// SSHPort is the SSH port number for connecting to the node
	SSHPort int
	// SSHUser is the SSH user for logging into the node
	SSHUser string
}

// ToINI converts the inventory into INI format
func (i Inventory) ToINI() []byte {
	w := &bytes.Buffer{}
	for _, role := range i.Roles {
		fmt.Fprintf(w, "[%s]\n", role.Name)
		for _, n := range role.Nodes {
			internalIP := n.PublicIP
			if n.InternalIP != "" {
				internalIP = n.InternalIP
			}
			fmt.Fprintf(w, "%q ansible_host=%q internal_ipv4=%q ansible_ssh_private_key_file=%q ansible_port=%d ansible_user=%q\n", n.Host, n.PublicIP, internalIP, n.SSHPrivateKey, n.SSHPort, n.SSHUser)
		}
	}

	return w.Bytes()
}
