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

// NewCmdServer returns the "server" command
func NewCmdServer(out io.Writer) *cobra.Command {
	var port int
	var nodeRoles string
	var enforcePackages bool
	var disconnectedInstallation bool
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Stand up the inspector server for running checks remotely",
		Example: serverExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(out, cmd.Parent().Name(), port, nodeRoles, enforcePackages, disconnectedInstallation)
		},
	}
	cmd.Flags().IntVar(&port, "port", 9090, "the port number for standing up the Inspector server")
	cmd.Flags().StringVar(&nodeRoles, "node-roles", "", "comma-separated list of the node's roles. Valid roles are 'etcd', 'master', 'worker'")
	cmd.Flags().BoolVarP(&enforcePackages, "enforcePackages", "e", false, "when provided the installer will test that all Kismatic packages have been installed")
	cmd.Flags().BoolVar(&disconnectedInstallation, "disconnected-installation", false, "when true will check for the required packages needed during a disconnected install")
	return cmd
}

func runServer(out io.Writer, commandName string, port int, nodeRoles string, enforcePackages bool, disconnectedInstallation bool) error {
	if nodeRoles == "" {
		return fmt.Errorf("--node-roles is required")
	}
	nodeFacts, err := getNodeRoles(nodeRoles)
	if err != nil {
		return err
	}
	if disconnectedInstallation {
		nodeFacts = append(nodeFacts, "disconnected")
	}
	s, err := inspector.NewServer(nodeFacts, port, enforcePackages)
	if err != nil {
		return fmt.Errorf("error starting up inspector server: %v", err)
	}
	fmt.Fprintf(out, "Inspector is listening on port %d\n", port)
	fmt.Fprintf(out, "Node roles: %s\n", nodeRoles)
	fmt.Fprintf(out, "Enforce packages: %v\n", enforcePackages)
	fmt.Fprintf(out, "Disconnected installation: %v\n", disconnectedInstallation)
	fmt.Fprintf(out, "Run %s from another node to run checks remotely: %[1]s client [NODE_IP]:%d\n", commandName, port)
	if err := s.Start(); err != nil {
		return err
	}
	return nil
}
