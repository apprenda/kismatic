package packet

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	var quiet bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists infrastructure running on Packet.net",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(quiet)
		},
	}
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only display hostnames")
	return cmd
}

func runList(quiet bool) error {
	client, err := newFromEnv()
	if err != nil {
		return err
	}
	nodes, err := client.ListNodes()
	if err != nil {
		return err
	}
	if quiet {
		for _, n := range nodes {
			fmt.Println(n.Host)
		}
		return nil
	}
	printNodes(os.Stdout, nodes)
	return nil
}
