package packet

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/apprenda/kismatic/pkg/provision/plan"
	"github.com/spf13/cobra"
)

func createMinikubeCmd() *cobra.Command {
	opts := &packetOpts{}
	cmd := &cobra.Command{
		Use:   "create-mini",
		Short: "Creates infrastructure for a single node cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateMinikube(opts)
		},
	}
	cmd.Flags().BoolVar(&opts.CentOS, "useCentos", false, "If present, will install CentOS 7 rather than Ubuntu 16.04")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().StringVar(&opts.Region, "region", "us-east", "The region to be used for provisioning machines. One of us-east|us-west|eu-west")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")

	return cmd
}

func runCreateMinikube(opts *packetOpts) error {
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
	region, err := regionFromString(opts.Region)
	if err != nil {
		return err
	}

	fmt.Println("Provisioning node")
	hostname := fmt.Sprintf("kismatic-node-%s", provTime)
	nodeID, err := c.CreateNode(hostname, distro, region)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for node to be accessible via SSH. This takes a while...")
	node, err := c.GetSSHAccessibleNode(nodeID, 15*time.Minute, c.SSHKey)
	if err != nil {
		return fmt.Errorf("error waiting for node to be ready")
	}

	fmt.Println()
	fmt.Printf("Finished provisioning nodes on Packet.net in %s\n", time.Now().Sub(startTime))

	if opts.NoPlan {
		fmt.Println("")
		fmt.Printf("%+v", node)
		return nil
	}

	template, err := template.New("plan").Parse(plan.OverlayNetworkPlan)
	if err != nil {
		return err
	}

	storageNodes := []plan.Node{}
	if opts.Storage {
		storageNodes = []plan.Node{*node}
	}

	plan := plan.Plan{
		Etcd:                []plan.Node{*node},
		Master:              []plan.Node{*node},
		Worker:              []plan.Node{*node},
		Ingress:             []plan.Node{*node},
		Storage:             storageNodes,
		MasterNodeFQDN:      node.PublicIPv4,
		MasterNodeShortName: node.PublicIPv4,
		SSHUser:             node.SSHUser,
		SSHKeyFile:          c.SSHKey,
		AdminPassword:       generateAlphaNumericPassword(),
	}
	f, err := makeUniqueFile(0)
	if err := template.Execute(f, plan); err != nil {
		return err
	}
	fmt.Println("To install your cluster, run:")
	fmt.Println("./kismatic install apply -f " + f.Name())

	return nil
}
