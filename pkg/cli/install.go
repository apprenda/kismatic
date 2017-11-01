package cli

import (
	"io"

	"github.com/spf13/cobra"
)

type installOpts struct {
	planFilename string
}

// NewCmdInstall creates a new install command
func NewCmdInstall(in io.Reader, out io.Writer) *cobra.Command {
	opts := &installOpts{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "install your Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Subcommands
	cmd.AddCommand(NewCmdPlan(in, out, opts))
	cmd.AddCommand(NewCmdValidate(out, opts))
	cmd.AddCommand(NewCmdApply(out, opts))
	cmd.AddCommand(NewCmdAddWorker(out, opts))
	cmd.AddCommand(NewCmdStep(out, opts))
	cmd.AddCommand(NewCmdProvision(in, out, opts))
	cmd.AddCommand(NewCmdDestroy(in, out, opts))

	// These probably need to be up a level?
	// PersistentFlags
	addPlanFileFlag(cmd.PersistentFlags(), &opts.planFilename)

	return cmd
}
