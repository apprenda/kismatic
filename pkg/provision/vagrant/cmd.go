package vagrant

import (
	"fmt"

	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type vagrantCmdOpts struct {
	PlanOpts
	NoPlan                  bool
	OnlyGenerateVagrantfile bool
}

//Cmd provides a Cobra Command Interface for the vagrant subset of commands
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vagrant",
		Short: "Provision a virtualized infrastructure using Vagrant.",
		Long:  `Provision a virtualized infrastructure using Vagrant. Vagrant will provide Ubuntu instances by default, but can be configured to use CentOS instead.`,
	}

	cmd.AddCommand(vagrantCreateCmd())
	cmd.AddCommand(vagrantCreateMinikubeCmd())

	return cmd
}

func addSharedFlags(cmd *cobra.Command, opts *vagrantCmdOpts) {
	opts.NodeCIDR = "192.168.42.2/24"
	cmd.Flags().BoolVarP(&opts.Redhat, "useCentOS", "r", false, "If present, will install CentOS 7.3 rather than Ubuntu 16.04")
	opts.Vagrantfile = "Vagrantfile"

	//PlanOpts
	opts.DisablePackageInstallation = false

	opts.AutoConfiguredDockerRegistry = false
	opts.DockerRegistryPort = 8443

	cmd.Flags().StringVar(&opts.AdminPassword, "adminPassword", util.GenerateAlphaNumericPassword(), "This password is used to login to the Kubernetes Dashboard and can also be used for administration without a security certificate")
	opts.AdminPassword = util.GenerateAlphaNumericPassword()
	opts.PodCIDR = "172.16.0.0/16"
	opts.ServiceCIDR = "172.20.0.0/16"
	// VagrantCmdOpts
	cmd.Flags().BoolVar(&opts.NoPlan, "noplan", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")
	cmd.Flags().BoolVar(&opts.FailSwapOn, "fail-swap-on", true, "If present, allows for the usage of swap on a local vagrant cluster.")
}

func vagrantCreateCmd() *cobra.Command {
	var etcdCount, masterCount, workerCount, ingressCount uint16

	opts := vagrantCmdOpts{
		PlanOpts: PlanOpts{
			InfrastructureOpts: InfrastructureOpts{
				Count: map[NodeType]uint16{},
			},
		},
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates infrastructure for a new cluster.",
		Long: `Creates infrastructure for a new cluster.

Smallish instances will be created with public IP addresses. Unless option onlyGenerateVagrantfile is true, the command will not return 
until the instances are all online and accessible via SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Count[etcd] = etcdCount
			opts.Count[master] = masterCount
			opts.Count[worker] = workerCount
			opts.Count[ingress] = ingressCount
			return makeInfrastructure(&opts)
		},
	}

	cmd.Flags().Uint16VarP(&etcdCount, "etcdNodeCount", "e", 1, "Count of etcd nodes to produce.")
	cmd.Flags().Uint16VarP(&masterCount, "masterdNodeCount", "m", 1, "Count of master nodes to produce.")
	cmd.Flags().Uint16VarP(&workerCount, "workerNodeCount", "w", 1, "Count of worker nodes to produce.")
	cmd.Flags().Uint16VarP(&ingressCount, "ingressNodeCount", "i", 1, "Count of ingress nodes to produce")
	// cmd.Flags().BoolVar(&opts.OverlapRoles, "overlapRoles", false, "Overlap roles to create as few nodes as possible")

	addSharedFlags(cmd, &opts)

	return cmd
}

func vagrantCreateMinikubeCmd() *cobra.Command {
	opts := vagrantCmdOpts{
		PlanOpts: PlanOpts{
			InfrastructureOpts: InfrastructureOpts{
				Count: map[NodeType]uint16{
					etcd:    1,
					master:  1,
					worker:  1,
					ingress: 1,
				},
				OverlapRoles: true,
			},
		},
	}

	cmd := &cobra.Command{
		Use:   "create-mini",
		Short: "Creates infrastructure for a single-node instance.",
		Long: `Creates infrastructure for a single-node instance. 

A smallish instance will be created with public IP addresses. The command will not return until the instance is online and accessible via SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeInfrastructure(&opts)
		},
	}

	addSharedFlags(cmd, &opts)

	return cmd
}

func makeInfrastructure(opts *vagrantCmdOpts) error {
	infrastructure, infraErr := NewInfrastructure(&opts.InfrastructureOpts)
	if infraErr != nil {
		return infraErr
	}

	_, vagrantErr := createVagrantfile(opts, infrastructure)
	if vagrantErr != nil {
		return vagrantErr
	}

	if opts.OnlyGenerateVagrantfile {
		fmt.Println("To create your local VMs, run:")
		fmt.Println("vagrant up")
	} else {
		if vagrantUpErr := vagrantUp(); vagrantUpErr != nil {
			return vagrantUpErr
		}
	}

	infrastructure.PrivateSSHKeyPath = grabSSHConfig()

	if !opts.NoPlan {
		planFile, planErr := createPlan(opts, infrastructure)
		if planErr != nil {
			return planErr
		}

		fmt.Println("To install your cluster, run:")
		fmt.Println("./kismatic install apply -f " + planFile)
	}

	return nil
}

func createVagrantfile(opts *vagrantCmdOpts, infrastructure *Infrastructure) (string, error) {
	vagrantfile, err := util.MakeFileAskOnOverwrite("Vagrantfile")
	if err != nil {
		return "", err
	}

	defer vagrantfile.Close()

	vagrant := &Vagrant{
		Opts:           &opts.InfrastructureOpts,
		Infrastructure: infrastructure,
	}

	err = vagrant.Write(vagrantfile)
	if err != nil {
		return "", err
	}

	return vagrantfile.Name(), nil
}

func createPlan(opts *vagrantCmdOpts, infrastructure *Infrastructure) (string, error) {
	planFile, err := util.MakeUniqueFile("kismatic-cluster", ".yaml", 0)
	if err != nil {
		return "", err
	}

	defer planFile.Close()

	plan := &Plan{
		Opts:           &opts.PlanOpts,
		Infrastructure: infrastructure,
	}

	err = plan.Write(planFile)
	if err != nil {
		return "", err
	}

	return planFile.Name(), nil
}
