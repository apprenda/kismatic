package digitalocean

import (
	"bufio"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"strings"

	"github.com/apprenda/kismatic/pkg/provision/plan"
	garbler "github.com/michaelbironneau/garbler/lib"
	"github.com/spf13/cobra"
)

type DOOpts struct {
	Token           string
	ClusterTag      string
	EtcdNodeCount   uint16
	MasterNodeCount uint16
	WorkerNodeCount uint16
	NoPlan          bool
	InstanceType    string
	WorkerType      string
	Image           string
	Region          string
	Storage         bool
	SSHUser         string
	SSHKeyName      string
	SSHPrivateKey   string
	SSHPublicKey    string
	BootstrapNode   bool
	RemoveKey       bool
	BootstrapFile   string
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "do",
		Short: "Provision infrastructure on Digital Ocean.",
		Long:  `Provision infrastructure on Digital Ocean.`,
	}

	cmd.AddCommand(DOCreateCmd())
	cmd.AddCommand(DODeleteCmd())

	return cmd
}

func DOCreateCmd() *cobra.Command {
	opts := DOOpts{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates infrastructure for a new cluster.",
		Long: `Creates infrastructure for a new cluster. Optionally creates a bootstrap node to run the orchestration of Kubernetes
cluster from. If the bootstrap node is requested, the provisioner will download kismatic executables and kubectl during the process
of VM initialization. By default, it will place the downloaded packages in the /ket/ folder. The default location can be overwritten
by setting an environmental variable 'DO_KET_INSTALL_DIR'. If the bootstrap node is not requested, the Kismatic and Kubectl packages
will have to be downloaded manually. See digitalocean/scripts/bootinit.sh for details.

In addition to the commands below, the provisioner relies on some environment variables and conventions:
Required:
  DO_API_TOKEN: [Required] Your Digital Ocean access token, required for all operations
  DO_SECRET_ACCESS_KEY: [Required] Your Digital Ocean ssh key, required for all operations. If the env varaible does
not exist, an attempt will be made to use ssh key file in the following relative location: ssh/cluster.pem file. If the file is
not found, the program will fail.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeInfra(opts)
		},
	}

	cmd.Flags().Uint16VarP(&opts.EtcdNodeCount, "etcdNodeCount", "e", 1, "Count of etcd nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.MasterNodeCount, "masterdNodeCount", "m", 1, "Count of master nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.WorkerNodeCount, "workerNodeCount", "w", 1, "Count of worker nodes to produce.")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().StringVarP(&opts.InstanceType, "instance-type", "i", "1gb", "Size of the instance. Current options: 1gb, 2gb, 4gb")
	cmd.Flags().StringVarP(&opts.WorkerType, "worker-type", "", "4gb", "Size of the worker node instance. Current options: 1gb, 2gb, 4gb")
	cmd.Flags().StringVarP(&opts.Image, "image", "", "ubuntu-16-04-x64", "Name of the image to use")
	cmd.Flags().StringVarP(&opts.Region, "region", "", "tor1", "Region to deploy to")
	cmd.Flags().StringVarP(&opts.ClusterTag, "tag", "", "apprenda", "TAG for all nodes in the cluster")
	cmd.Flags().StringVarP(&opts.SSHUser, "sshuser", "", "root", "SSH User name")
	cmd.Flags().BoolVarP(&opts.BootstrapNode, "bootstrap", "", true, "Create a bootstrap node from which users can work with the cluster.")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")
	cmd.Flags().StringVarP(&opts.BootstrapFile, "bootstrap-commands-file", "", "", "Relative path to the script file that will be run on the bootstrap node upon initialization. e.g.: digitalocean/scripts/bootinit.sh.")

	return cmd
}

func DODeleteCmd() *cobra.Command {
	opts := DOOpts{}
	cmd := &cobra.Command{
		Use:   "delete-all",
		Short: "Deletes all the nodes from the Digital Ocean account",
		Long:  `Deletes all the nodes based on the tag provided and also, if requested, removes the ssh key created during the provisioning`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteInfra(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ClusterTag, "tag", "", "apprenda", "All nodes with the provided tag will be removed")
	cmd.Flags().BoolVarP(&opts.RemoveKey, "remove-key", "", false, "Inidicator whether the ssh key used for the provisioning should be deleted")

	return cmd
}

func deleteInfra(opts DOOpts) error {
	opts.Token = os.Getenv("DO_API_TOKEN")
	reader := bufio.NewReader(os.Stdin)
	if opts.Token == "" {
		fmt.Print("Enter Digital Ocean API Token: ")
		url, _ := reader.ReadString('\n')
		opts.Token = strings.Trim(url, "\n")
		opts.Token = strings.Replace(opts.Token, "\r", "", -1) //for Windows
	}

	provisioner, _ := GetProvisioner()

	return provisioner.TerminateNodes(opts)
}

func validateKeyFile(opts DOOpts) (string, string, error) {
	var filePath string

	sshKeyPath := os.Getenv("DO_SECRET_ACCESS_KEY")
	if sshKeyPath == "" {
		//try ssh dir relative to the executable
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			fmt.Println("Cannot get path to exec", err)
		}
		sshKeyPath = filepath.Join(dir, "ssh/")
		fmt.Println("Trying to locate key in ssh/ folder", sshKeyPath)

		filePath = filepath.Join(sshKeyPath, "cluster.pem")
		_, staterr := os.Stat(filePath)
		if os.IsNotExist(staterr) {
			return "", "", fmt.Errorf("Private SSH file was not found in expected location. Create your own key pair and reference in options to the provision command. Change file permissions to allow w/r for the user (chmod 600) %v", err)
		}
	} else {
		filePath = sshKeyPath
	}

	return filePath, filePath + ".pub", nil
}

func makeInfra(opts DOOpts) error {
	opts.Token = os.Getenv("DO_API_TOKEN")
	reader := bufio.NewReader(os.Stdin)
	if opts.Token == "" {
		fmt.Print("Enter Digital Ocean API Token: \n")
		url, _ := reader.ReadString('\n')
		opts.Token = strings.Trim(url, "\n")
		opts.Token = strings.Replace(opts.Token, "\r", "", -1) //for Windows
	}
	if opts.Token == "" {
		return fmt.Errorf("The DigitalOcean API Token is required")
	}
	sshPrivate, sshPublic, errkey := validateKeyFile(opts)
	if errkey != nil {
		return errkey
	}
	s, err := os.Stat(sshPrivate)
	if os.IsNotExist(err) {
		return fmt.Errorf("Did not find SSH private key at %q", sshPrivate)
	}
	opts.SSHKeyName = s.Name()
	fmt.Println("SSH file name", opts.SSHKeyName)
	opts.SSHPrivateKey = sshPrivate
	opts.SSHPublicKey = sshPublic
	if errkey != nil {
		return errkey
	}

	fmt.Print("Provisioning\n")
	var bootCount uint16 = 0
	if opts.BootstrapNode {
		bootCount = 1
	}
	provisioner, _ := GetProvisioner()
	nodes, err := provisioner.ProvisionNodes(opts, NodeCount{
		Etcd:     opts.EtcdNodeCount,
		Worker:   opts.WorkerNodeCount,
		Master:   opts.MasterNodeCount,
		Boostrap: bootCount,
	})

	if err != nil {
		return err
	}

	fmt.Print("Waiting for SSH\n")
	if err = WaitForSSH(nodes, opts.SSHPrivateKey); err != nil {
		return err
	}

	if opts.NoPlan {
		fmt.Println("Your instances are ready.\n")
		printNodes(&nodes)
		return nil
	}

	storageNodes := []plan.Node{}
	if opts.Storage {
		storageNodes = nodes.Worker
	}
	root := os.Getenv("DO_KET_INSTALL_DIR")
	if root == "" {
		root = KET_INSTALL_DIR
	}

	sshKeyFile := opts.SSHPrivateKey
	// If the user asks for a bootstrap node, the generated plan file will contain
	// the path to the SSH key on the bootstrap node, and not on the node that is running
	// provision.
	if opts.BootstrapNode {
		sshKeyFile = fmt.Sprintf("%s/ssh/%s", root, opts.SSHKeyName)
	}

	return makePlan(&plan.Plan{
		AdminPassword:       generateAlphaNumericPassword(),
		Etcd:                nodes.Etcd,
		Master:              nodes.Master,
		Worker:              nodes.Worker,
		Ingress:             []plan.Node{nodes.Worker[0]},
		Storage:             storageNodes,
		MasterNodeFQDN:      nodes.Master[0].PublicIPv4,
		MasterNodeShortName: nodes.Master[0].PublicIPv4,
		SSHKeyFile:          sshKeyFile,
		SSHUser:             nodes.Master[0].SSHUser,
	}, opts, nodes)

}

func makePlan(pln *plan.Plan, opts DOOpts, nodes ProvisionedNodes) error {
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

	//scp plan file to bootstrap if requested
	if opts.BootstrapNode {
		boot := nodes.Boostrap[0]
		planPath, _ := filepath.Abs(f.Name())
		fmt.Println("Copying kismatic plan file to bootstrap node:", planPath)
		root := os.Getenv("DO_KET_INSTALL_DIR")
		if root == "" {
			root = KET_INSTALL_DIR
		}
		if opts.BootstrapFile == "" {
			root = ""
		}
		destPath := root + "/kismatic-cluster.yaml"
		out, scperr := scpFile(planPath, destPath, opts.SSHUser, boot.PublicIPv4, opts.SSHPrivateKey)
		if scperr != nil {
			return fmt.Errorf("Unable to push kismatic plan to boostrap node: %v", scperr)
		}
		fmt.Println("Output:", out)
	}
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
	}
	return makeUniqueFile(count + 1)
}

func printNodes(nodes *ProvisionedNodes) {
	printRole("Etcd", &nodes.Etcd)
	printRole("Master", &nodes.Master)
	printRole("Worker", &nodes.Worker)
	printRole("Bootstrap", &nodes.Boostrap)
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
