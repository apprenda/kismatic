package cli

import (
	"io"

	"github.com/apprenda/kismatic/pkg/provision/aws"
	"github.com/apprenda/kismatic/pkg/provision/digitalocean"
	"github.com/apprenda/kismatic/pkg/provision/packet"
	"github.com/apprenda/kismatic/pkg/provision/vagrant"
	"github.com/spf13/cobra"
)

type provisionOpts struct {
	planFilename string
	failSwapOn   bool
}

// NewCmdProvision provisions a cluster on one of the providers
func NewCmdProvision(in io.Reader, out io.Writer) *cobra.Command {
	opts := &provisionOpts{}
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provides tooling for making Kubernetes capable infrastructure",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Subcommands
	cmd.AddCommand(aws.Cmd())
	cmd.AddCommand(digitalocean.Cmd())
	cmd.AddCommand(packet.Cmd())
	cmd.AddCommand(vagrant.Cmd())

	// PersistentFlags
	addPlanFileFlag(cmd.PersistentFlags(), &opts.planFilename)
	return cmd
}
