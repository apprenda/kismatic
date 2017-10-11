package packet

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/apprenda/kismatic/pkg/provision/plan"

	garbler "github.com/michaelbironneau/garbler/lib"
	"github.com/spf13/cobra"
)

func createCmd() *cobra.Command {
	opts := &packetOpts{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates infrastructure for a new cluster.",
		Example: `# Create 3 etcd nodes, 2 master nodes and 3 worker nodes
provision packet create -e 3 -m 2 -w 3

# Create 1 etcd node, 1 master node and 1 worker node using CentOS 7
provision packet create --useCentos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts)
		},
	}
	cmd.Flags().Uint16VarP(&opts.EtcdNodeCount, "etcdNodeCount", "e", 1, "Count of etcd nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.MasterNodeCount, "masterdNodeCount", "m", 1, "Count of master nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.WorkerNodeCount, "workerNodeCount", "w", 1, "Count of worker nodes to produce.")
	cmd.Flags().BoolVar(&opts.CentOS, "useCentos", false, "If present, will install CentOS 7 rather than Ubuntu 16.04")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().StringVar(&opts.Region, "region", "us-east", "The region to be used for provisioning machines. One of us-east|us-west|eu-west")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")

	return cmd
}

func regionFromString(region string) (Region, error) {
	switch region {
	case "us-east":
		return USEast, nil
	case "us-west":
		return USWest, nil
	case "eu-west":
		return EUWest, nil
	default:
		return "", fmt.Errorf("unknown region: %s", region)
	}
}

func runCreate(opts *packetOpts) error {
	startTime := time.Now()
	c, err := newFromEnv()
	if err != nil {
		return err
	}

	distro := Ubuntu1604LTS
	if opts.CentOS {
		distro = CentOS7
	}
	provTime := strconv.FormatInt(time.Now().Unix(), 10)
	generateHostname := hostnameGenerator("kismatic", provTime)
	nodeIDs := struct {
		etcd   []string
		master []string
		worker []string
	}{}
	region, err := regionFromString(opts.Region)
	if err != nil {
		return err
	}

	fmt.Println("Provisioning nodes")
	var i uint16
	for i = 0; i < opts.EtcdNodeCount; i++ {
		hostname := generateHostname("etcd", int(i))
		nodeID, err := c.CreateNode(hostname, distro, region)
		if err != nil {
			return err
		}
		nodeIDs.etcd = append(nodeIDs.etcd, nodeID)
	}
	for i = 0; i < opts.MasterNodeCount; i++ {
		hostname := generateHostname("master", int(i))
		nodeID, err := c.CreateNode(hostname, distro, region)
		if err != nil {
			return err
		}
		nodeIDs.master = append(nodeIDs.master, nodeID)
	}
	for i = 0; i < opts.WorkerNodeCount; i++ {
		hostname := generateHostname("worker", int(i))
		nodeID, err := c.CreateNode(hostname, distro, region)
		if err != nil {
			return err
		}
		nodeIDs.worker = append(nodeIDs.worker, nodeID)
	}

	fmt.Println("Waiting for nodes to be accessible via SSH. This takes a while...")
	nodes := struct {
		etcd   []plan.Node
		master []plan.Node
		worker []plan.Node
	}{}
	for _, id := range nodeIDs.etcd {
		node, err := c.GetSSHAccessibleNode(id, 15*time.Minute, c.SSHKey)
		if err != nil {
			return fmt.Errorf("error waiting for node to be ready")
		}
		nodes.etcd = append(nodes.etcd, *node)
	}
	for _, id := range nodeIDs.master {
		node, err := c.GetSSHAccessibleNode(id, 15*time.Minute, c.SSHKey)
		if err != nil {
			return fmt.Errorf("error waiting for node to be ready")
		}
		nodes.master = append(nodes.master, *node)
	}
	for _, id := range nodeIDs.worker {
		node, err := c.GetSSHAccessibleNode(id, 15*time.Minute, c.SSHKey)
		if err != nil {
			return fmt.Errorf("error waiting for node to be ready")
		}
		nodes.worker = append(nodes.worker, *node)
	}
	fmt.Println()
	fmt.Printf("Finished provisioning nodes on Packet.net in %s\n", time.Now().Sub(startTime))

	if opts.NoPlan {
		fmt.Println("Etcd:")
		for _, n := range nodes.etcd {
			printNode(n)
		}
		fmt.Println("Master:")
		for _, n := range nodes.master {
			printNode(n)
		}
		fmt.Println("Worker:")
		for _, n := range nodes.worker {
			printNode(n)
		}
		return nil
	}

	storageNodes := []plan.Node{}
	if opts.Storage {
		storageNodes = nodes.worker
	}

	// Write the plan file out
	planit := plan.Plan{
		Etcd:                nodes.etcd,
		Master:              nodes.master,
		Worker:              nodes.worker,
		Ingress:             nodes.worker[0:1],
		Storage:             storageNodes,
		MasterNodeFQDN:      nodes.master[0].PublicIPv4,
		MasterNodeShortName: nodes.master[0].PublicIPv4,
		SSHUser:             nodes.master[0].SSHUser,
		SSHKeyFile:          c.SSHKey,
		AdminPassword:       generateAlphaNumericPassword(),
	}

	template, err := template.New("plan").Parse(plan.OverlayNetworkPlan)
	if err != nil {
		return err
	}
	f, err := makeUniqueFile(0)
	if err := template.Execute(f, planit); err != nil {
		return err
	}
	fmt.Println("To install your cluster, run:")
	fmt.Println("./kismatic install apply -f " + f.Name())
	return nil
}

func generateAlphaNumericPassword() string {
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

func makeUniqueFile(count int) (*os.File, error) {
	filename := "kismatic-cluster"
	if count > 0 {
		filename = filename + "-" + strconv.Itoa(count)
	}
	filename = filename + ".yaml"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	}
	return makeUniqueFile(count + 1)
}

func printNode(n plan.Node) {
	fmt.Printf("  %v (Public: %v, Private: %v)\n", n.Host, n.PublicIPv4, n.PrivateIPv4)
}

func hostnameGenerator(nodePrefix, timestamp string) func(string, int) string {
	return func(nodeType string, nodeIndex int) string {
		return fmt.Sprintf("%s-%s-%d-%s", nodePrefix, nodeType, nodeIndex, timestamp)
	}
}
