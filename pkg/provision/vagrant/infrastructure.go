package vagrant

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/util"
)

//NodeType
type NodeType uint32

const (
	etcd NodeType = 1 << iota
	master
	worker
	ingress
)

var nodeTypes = []NodeType{etcd, master, worker, ingress}

var nodeTypeStrings = map[NodeType]string{
	etcd:    "etcd",
	master:  "master",
	worker:  "worker",
	ingress: "ingress",
}

type InfrastructureOpts struct {
	Count             map[NodeType]uint16
	OverlapRoles      bool
	NodeCIDR          string
	Redhat            bool
	PrivateSSHKeyPath string
	Vagrantfile       string
	Storage           bool
}

type NodeDetails struct {
	Name  string
	IP    net.IP
	Types NodeType
}

type Infrastructure struct {
	Network           net.IPNet
	Broadcast         net.IP
	Nodes             []NodeDetails
	DNSReflector      string
	PrivateSSHKeyPath string
	PublicSSHKeyPath  string
}

func NewInfrastructure(opts *InfrastructureOpts) (*Infrastructure, error) {
	_, network, err := net.ParseCIDR(opts.NodeCIDR)

	if err != nil {
		return nil, err
	}

	broadcast, err := util.BroadcastIPv4(*network)
	if err != nil {
		return nil, err
	}

	i := &Infrastructure{
		Network:   *network,
		Broadcast: broadcast,
		Nodes:     []NodeDetails{},
	}

	var overlapTypes NodeType

	// keep creating nodes until counts are exhausted
	for j := uint16(1); ; j++ {
		overlapTypes = NodeType(0)
		finished := true

		for _, nodeType := range nodeTypes {
			if j <= opts.Count[nodeType] {
				if opts.OverlapRoles {
					overlapTypes |= nodeType
				} else {
					_, err := i.appendNode(j, nodeTypeStrings[nodeType], nodeType)
					if err != nil {
						return i, err
					}
				}
				finished = false
			}
		}

		if overlapTypes > 0 {
			_, err := i.appendNode(j, "node", overlapTypes)
			if err != nil {
				return i, err
			}
		}

		if finished {
			break
		}
	}

	return i, nil
}

func (i *Infrastructure) ensureSSHKeys(privateSSHKeyPath string) error {

	if privateSSHKeyPath == "" {
		i.PrivateSSHKeyPath = "kismatic-cluster.pem"
	} else {
		i.PrivateSSHKeyPath = privateSSHKeyPath
	}

	// ensure absolute path
	var absErr error
	i.PrivateSSHKeyPath, absErr = filepath.Abs(i.PrivateSSHKeyPath)
	if absErr != nil {
		return absErr
	}

	i.PublicSSHKeyPath = i.PrivateSSHKeyPath + ".pub"

	privateKey, privateKeyErr := util.LoadOrCreatePrivateSSHKey(i.PrivateSSHKeyPath)
	if privateKeyErr != nil {
		return privateKeyErr
	}

	publicKeyErr := util.CreatePublicKey(privateKey, i.PublicSSHKeyPath)
	if publicKeyErr != nil {
		return publicKeyErr
	}

	// ensure correct permissions
	os.Chmod(i.PrivateSSHKeyPath, 0600)
	os.Chmod(i.PublicSSHKeyPath, 0600)

	return nil
}

func (i *Infrastructure) appendNode(nodeIndex uint16, name string, types NodeType) (*NodeDetails, error) {
	ip, err := i.nextNodeIP()

	if err != nil {
		return nil, err
	}

	hostname := fmt.Sprintf("%v%03d", name, nodeIndex)

	node := NodeDetails{
		Name:  hostname,
		IP:    ip,
		Types: types,
	}

	i.Nodes = append(i.Nodes, node)

	return &node, nil
}

func (i *Infrastructure) nextNodeIP() (net.IP, error) {
	var ip net.IP
	var err error

	if len(i.Nodes) < 1 {
		ip = i.Network.IP

		// increment by 1 to account for gateway
		ip, err = util.IncrementIPv4(ip)
		if err != nil {
			return nil, err
		}
	} else {
		lastNode := i.Nodes[len(i.Nodes)-1:][0]
		ip = lastNode.IP
	}

	ip, err = util.IncrementIPv4(ip)
	if err != nil {
		return nil, err
	}

	// assumes broadcast address is last host in CIDR range
	if !i.Network.Contains(ip) || i.Broadcast.Equal(ip) {
		return ip, errors.New("infrastructure: ip address overflowed available cidr range")
	}

	return ip, nil
}

func (i *Infrastructure) nodesByType(nodeType NodeType) []NodeDetails {
	filtered := []NodeDetails{}
	for _, node := range i.Nodes {
		if (node.Types & nodeType) > NodeType(0) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}
