package packet

import "github.com/spf13/cobra"

type packetOpts struct {
	EtcdNodeCount   uint16
	MasterNodeCount uint16
	WorkerNodeCount uint16
	CentOS          bool
	NoPlan          bool
	Region          string
	Storage         bool
}

// Cmd returns the command for managing Packet infrastructure
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packet",
		Short: "Provision infrastructure on Packet.net",
		Long: `Provision infrastructure on Packet.net.

The following environment variables are used when provisioning on Packet:
Required:
  PACKET_API_KEY: Your Packet.net API key, required for all operations
  PACKET_PROJECT_ID: The ID of the project where machines will be provisioned

Optional:
  PACKET_SSH_KEY_PATH: The path to the SSH key to be used for accessing the machines.
    If empty, a file called "kismatic-packet.pem" in the current working directory is
    used as the SSH key.
`,
	}
	cmd.AddCommand(createCmd())
	cmd.AddCommand(createMinikubeCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(listCmd())
	return cmd
}
