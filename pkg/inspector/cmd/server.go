package cmd

import (
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/inspector"
	"github.com/spf13/cobra"
)

var serverExample = `# Run the inspector in server mode
kismatic-inspector server --node-roles master,worker

# Run the inspector in server mode, in a specific port
kismatic-inspector server --port 9000 --node-roles master
`

type serverOpts struct {
	commandName                 string
	port                        int
	nodeRoles                   string
	packageInstallationDisabled bool
	dockerInstallationDisabled  bool
	disconnectedInstallation    bool
}

// NewCmdServer returns the "server" command
func NewCmdServer(out io.Writer) *cobra.Command {
	opts := serverOpts{}
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Stand up the inspector server for running checks remotely",
		Example: serverExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.commandName = cmd.Parent().Name()
			return runServer(out, opts)
		},
	}
	cmd.Flags().IntVar(&opts.port, "port", 9090, "the port number for standing up the Inspector server")
	cmd.Flags().StringVar(&opts.nodeRoles, "node-roles", "", "comma-separated list of the node's roles. Valid roles are 'etcd', 'master', 'worker', 'ingress', 'storage'")
	cmd.Flags().BoolVar(&opts.packageInstallationDisabled, "pkg-installation-disabled", false, "when true, the inspector will ensure that the necessary packages are installed on the node")
	cmd.Flags().BoolVar(&opts.dockerInstallationDisabled, "docker-installation-disabled", false, "when true, the inspector will check for docker packages to be installed")
	cmd.Flags().BoolVar(&opts.disconnectedInstallation, "disconnected-installation", false, "when true will check for the required packages needed during a disconnected install")
	return cmd
}

func runServer(out io.Writer, opts serverOpts) error {
	if opts.nodeRoles == "" {
		return fmt.Errorf("--node-roles is required")
	}
	nodeFacts, err := getNodeRoles(opts.nodeRoles)
	if err != nil {
		return err
	}
	if opts.disconnectedInstallation {
		nodeFacts = append(nodeFacts, "disconnected")
	}
	s, err := inspector.NewServer(nodeFacts, opts.port, opts.packageInstallationDisabled, opts.dockerInstallationDisabled)
	if err != nil {
		return fmt.Errorf("error starting up inspector server: %v", err)
	}
	fmt.Fprintf(out, "Inspector is listening on port %d\n", opts.port)
	fmt.Fprintf(out, "Node roles: %s\n", opts.nodeRoles)
	fmt.Fprintf(out, "Package installation disabled: %v\n", opts.packageInstallationDisabled)
	fmt.Fprintf(out, "Docker installation disabled: %v\n", opts.dockerInstallationDisabled)
	fmt.Fprintf(out, "Disconnected installation: %v\n", opts.disconnectedInstallation)
	fmt.Fprintf(out, "Run %s from another node to run checks remotely: %[1]s client [NODE_IP]:%d\n", opts.commandName, opts.port)
	return s.Start()
}
