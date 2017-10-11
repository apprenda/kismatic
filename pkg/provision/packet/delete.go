package packet

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command {
	var deleteAll bool
	cmd := &cobra.Command{
		Use:   "delete [HOSTNAME]",
		Short: "Delete machines from the Packet.net project. This will destroy machines. Be ready.",
		Long: `Delete machines from the Packet.net project.

This command destroys machines on the project that is being managed with this tool.

It will destroy machines in the project, regardless of whether the machines were provisioned with this tool.

Be ready.
		`,
		Example: `# Delete a specific machine in the project
provision packet delete kismatic-master-0

# Delete all machines in the project
provision packet delete --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doDelete(cmd, args, deleteAll)
		},
	}
	cmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all machines in the project.")
	return cmd
}

func doDelete(cmd *cobra.Command, args []string, deleteAll bool) error {
	if !deleteAll && len(args) != 1 {
		return errors.New("You must provide the hostname of the machine to be deleted, or use the --all flag to destroy all machines in the project")
	}
	hostname := ""
	if !deleteAll {
		hostname = args[0]
	}
	client, err := newFromEnv()
	if err != nil {
		return err
	}
	nodes, err := client.ListNodes()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if hostname == n.Host || deleteAll {
			if err := client.DeleteNode(n.ID); err != nil {
				return err
			}
			fmt.Println("Deleted", n.Host)
		}
	}
	return nil
}
