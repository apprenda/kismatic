package aws

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"regexp"
	"strconv"

	"strings"

	"github.com/apprenda/kismatic/pkg/provision/plan"
	garbler "github.com/michaelbironneau/garbler/lib"
	"github.com/spf13/cobra"
)

type AWSOpts struct {
	EtcdNodeCount   uint16
	MasterNodeCount uint16
	WorkerNodeCount uint16
	LeaveArtifacts  bool
	RunKismatic     bool
	NoPlan          bool
	ForceProvision  bool
	KeyPairName     string
	InstanceType    string
	OS              string
	Storage         bool
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Provision infrastructure on AWS.",
		Long: `Provision infrastructure on AWS.
		
In addition to the commands below, AWS relies on some environment variables and conventions:
Required:
  AWS_ACCESS_KEY_ID: [Required] Your AWS access key, required for all operations
  AWS_SECRET_ACCESS_KEY: [Required] Your AWS secret key, required for all operations

Conditional: (These may be omitted if the -f flag is used)
  AWS_SUBNET_ID: The ID of a subnet to try to place machines into. If this environment variable exists, 
                 it must be a real subnet in the us-east-1 region or all commands will fail.
  AWS_SECURITY_GROUP_ID: The ID of a security group to place all new machines in. Must be a part of the 
                         above subnet or commands will fail.
  AWS_KEY_NAME: The name of a Keypair in AWS to be used to create machines. If empty, we will attempt 
                to use a key named 'kismatic-integration-testing' and fail if it does not exist.
  AWS_SSH_KEY_PATH: The absolute path to the private key associated with the Key Name above. If left blank,
                    we will attempt to use a key named 'kismaticuser.key' in the same directory as the 
					provision tool. This key is important as part of provisioning is ensuring that your
					instance is online and is able to be reached via SSH.
`,
	}

	cmd.AddCommand(AWSCreateCmd())
	cmd.AddCommand(AWSCreateMinikubeCmd())
	cmd.AddCommand(AWSDeleteCmd())

	return cmd
}

func AWSCreateCmd() *cobra.Command {
	opts := AWSOpts{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates infrastructure for a new cluster. For now, only the US East region is supported.",
		Long: `Creates infrastructure for a new cluster. 
		
For now, only the US East region is supported.

Smallish instances will be created with public IP addresses. The command will not return until the instances are all online and accessible via SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeInfra(opts)
		},
	}

	cmd.Flags().Uint16VarP(&opts.EtcdNodeCount, "etcdNodeCount", "e", 1, "Count of etcd nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.MasterNodeCount, "masterdNodeCount", "m", 1, "Count of master nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.WorkerNodeCount, "workerNodeCount", "w", 1, "Count of worker nodes to produce.")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().BoolVarP(&opts.ForceProvision, "force-provision", "f", false, "If present, generate anything needed to build a cluster including VPCs, keypairs, routes, subnets, & a very insecure security group.")
	cmd.Flags().StringVarP(&opts.InstanceType, "instance-type-blueprint", "i", "small", "A blueprint of instance type(s). Current options: micro (all t2 micros), small (t2 micros, workers are t2.medium), beefy (M4.large and xlarge)")
	cmd.Flags().StringVarP(&opts.OS, "operating-system", "o", "ubuntu", "Which flavor of Linux to provision. Try ubuntu, centos or rhel.")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")

	return cmd
}

func AWSCreateMinikubeCmd() *cobra.Command {
	opts := AWSOpts{}
	cmd := &cobra.Command{
		Use:   "create-mini",
		Short: "Creates infrastructure for a single-node instance. For now, only the US East region is supported.",
		Long: `Creates infrastructure for a single-node instance. 
		
For now, only the US East region is supported.

A smallish instance will be created with public IP addresses. The command will not return until the instance is online and accessible via SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeInfraMinikube(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OS, "operating-system", "o", "ubuntu", "Which flavor of Linux to provision. Try ubuntu, centos or rhel.")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().BoolVarP(&opts.ForceProvision, "force-provision", "f", false, "If present, generate anything needed to build a cluster including VPCs, keypairs, routes, subnets, & a very insecure security group.")
	cmd.Flags().StringVarP(&opts.InstanceType, "instance-type-blueprint", "i", "small", "A blueprint of instance type(s). Current options: micro (all t2 micros), small (t2 micros, workers are t2.medium), beefy (M4.large and xlarge)")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")

	return cmd
}

func AWSDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-all",
		Short: "Deletes all objects tagged as created by this machine with this tool. This will destroy clusters. Be ready.",
		Long: `Deletes all objects tagged as CreatedBy this machine and ProvisionedBy kismatic. 
		
This command destroys clusters.

It has no way of knowing that you had really important data on them. It is utterly remorseless.
		
Be ready.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteInfra()
		},
	}

	return cmd
}

func checkAWSCredentials() error {
	c := CompositeError{}
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		c.add(errors.New("Need AWS_ACCESS_KEY_ID env variable set to perform any AWS operations"))
	}
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		c.add(errors.New("Need AWS_SECRET_ACCESS_KEY env variable set to perform any AWS operations"))
	}
	if c.hasError() {
		return c
	}
	return nil
}

// An Error made up of many contributing errors that all have equal weight (e.g. do not form a stack)
type CompositeError struct {
	e []error
}

func (c CompositeError) Error() string {
	ret := ""
	for _, e := range c.e {
		ret = ret + fmt.Sprintf(" - %v\n", e)
	}
	return ret
}

func (c *CompositeError) add(woe error) {
	c.e = append(c.e, woe)
}

func (c *CompositeError) merge(c2 CompositeError) {
	for _, e := range c2.e {
		c.e = append(c.e, e)
	}
}

func (c *CompositeError) hasError() bool {
	return len(c.e) > 0
}

func checkAWSDeploymentEnvironment() error {
	c := CompositeError{}
	if os.Getenv("AWS_SUBNET_ID") == "" {
		c.add(errors.New("Need AWS_SUBNET_ID env variable set to perform this AWS operations"))
	}
	if os.Getenv("AWS_SECURITY_GROUP_ID") == "" {
		c.add(errors.New("Need AWS_SECURITY_GROUP_ID env variable set to perform this AWS operations"))
	}

	if c.hasError() {
		return c
	}
	return nil
}

func deleteInfra() error {
	if err := checkAWSCredentials(); err != nil {
		return err
	}

	awsClient, _ := AWSClientFromEnvironment()

	return awsClient.TerminateAllNodes()
}

func prepareToModifyAWS(forceProvision bool) error {
	if err := checkAWSCredentials(); err != nil {
		return err
	}

	awsClient, _ := AWSClientFromEnvironment()

	fmt.Printf("Using region %v\n", awsClient.client.Config.Region)

	if forceProvision {
		if err := awsClient.ForceProvision(); err != nil {
			return err
		}
	}

	if err := checkAWSDeploymentEnvironment(); err != nil {
		return err
	}

	s, err := os.Stat(awsClient.sshKey)
	if os.IsNotExist(err) {
		return err
	}

	if s.Mode().Perm()&0044 != 0000 {

		return fmt.Errorf("Set permissions of %v to 0600", awsClient.sshKey)
	}

	return nil
}

func assertOptions(opts AWSOpts) (NodeBlueprint, LinuxDistro, error) {
	blueprint, ok := NodeBlueprintMap[opts.InstanceType]
	if !ok {
		return NodeBlueprint{}, "", fmt.Errorf("%v is not valid option for instance type blueprint.", opts.InstanceType)
	}
	if err := prepareToModifyAWS(opts.ForceProvision); err != nil {
		return NodeBlueprint{}, "", err
	}

	distro := Ubuntu1604LTS
	switch strings.ToLower(opts.OS) {
	case "centos":
		distro = CentOS7
	case "ubuntu":
		distro = Ubuntu1604LTS
	case "rhel":
		distro = Redhat7
	default:
		return NodeBlueprint{}, "", fmt.Errorf("%s is not a known option for OS", opts.OS)
	}
	return blueprint, distro, nil
}

func makeInfraMinikube(opts AWSOpts) error {
	blueprint, distro, err := assertOptions(opts)
	if err != nil {
		return err
	}

	fmt.Print("Provisioning")
	awsClient, _ := AWSClientFromEnvironment()
	nodes, err := awsClient.ProvisionNodes(blueprint, NodeCount{
		Worker: 1,
	}, distro)

	if err != nil {
		return err
	}

	sshKey := awsClient.SSHKey()
	fmt.Print("Waiting for SSH")
	if err = WaitForSSH(nodes, sshKey); err != nil {
		return err
	}

	if opts.NoPlan {
		fmt.Println("Your instances are ready.\n")
		printRole("Minikube", &nodes.Worker)
	} else {
		storageNodes := []plan.Node{}
		if opts.Storage {
			storageNodes = []plan.Node{nodes.Worker[0]}
		}
		return makePlan(&plan.Plan{
			AdminPassword:       generateAlphaNumericPassword(),
			Etcd:                []plan.Node{nodes.Worker[0]},
			Master:              []plan.Node{nodes.Worker[0]},
			Worker:              []plan.Node{nodes.Worker[0]},
			Ingress:             []plan.Node{nodes.Worker[0]},
			Storage:             storageNodes,
			MasterNodeFQDN:      nodes.Worker[0].PublicIPv4,
			MasterNodeShortName: nodes.Worker[0].PrivateIPv4,
			SSHKeyFile:          sshKey,
			SSHUser:             nodes.Worker[0].SSHUser,
		})
	}
	return nil
}

func makeInfra(opts AWSOpts) error {
	blueprint, distro, err := assertOptions(opts)
	if err != nil {
		return err
	}

	fmt.Print("Provisioning")
	awsClient, _ := AWSClientFromEnvironment()
	nodes, err := awsClient.ProvisionNodes(blueprint, NodeCount{
		Etcd:   opts.EtcdNodeCount,
		Worker: opts.WorkerNodeCount,
		Master: opts.MasterNodeCount,
	}, distro)

	if err != nil {
		return err
	}

	sshKey := awsClient.SSHKey()
	fmt.Print("Waiting for SSH")
	if err = WaitForSSH(nodes, sshKey); err != nil {
		return err
	}

	if opts.NoPlan {
		fmt.Println("Your instances are ready.\n")
		printNodes(&nodes)
	} else {
		storageNodes := []plan.Node{}
		if opts.Storage {
			storageNodes = nodes.Worker
		}

		return makePlan(&plan.Plan{
			AdminPassword:       generateAlphaNumericPassword(),
			Etcd:                nodes.Etcd,
			Master:              nodes.Master,
			Worker:              nodes.Worker,
			Ingress:             []plan.Node{nodes.Worker[0]},
			Storage:             storageNodes,
			MasterNodeFQDN:      nodes.Master[0].PublicIPv4,
			MasterNodeShortName: nodes.Master[0].PrivateIPv4,
			SSHKeyFile:          sshKey,
			SSHUser:             nodes.Master[0].SSHUser,
		})
	}
	return nil
}

func makePlan(pln *plan.Plan) error {
	template, err := template.New("planAWSOverlay").Parse(plan.OverlayNetworkPlan)
	if err != nil {
		return err
	}

	f, err := makeUniqueFile(0)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	if err = template.Execute(w, &pln); err != nil {
		return err
	}

	w.Flush()
	fmt.Println("To install your cluster, run:")
	fmt.Println("./kismatic install apply -f " + f.Name())

	return nil
}

func makeUniqueFile(count int) (*os.File, error) {
	filename := "kismatic-cluster"
	if count > 0 {
		filename = filename + "-" + strconv.Itoa(count)
	}
	filename = filename + ".yaml"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	} else {
		return makeUniqueFile(count + 1)
	}
}

func printNodes(nodes *ProvisionedNodes) {
	printRole("Etcd", &nodes.Etcd)
	printRole("Master", &nodes.Master)
	printRole("Worker", &nodes.Worker)
}

func printRole(title string, nodes *[]plan.Node) {
	fmt.Printf("%v:\n", title)
	for _, node := range *nodes {
		fmt.Printf("  %v (%v, %v)\n", node.ID, node.PublicIPv4, node.PrivateIPv4)
	}
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
