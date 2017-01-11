package cli

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	// PersistentFlags
	addPlanFileFlag(cmd.PersistentFlags(), &opts.planFilename)

	return cmd
}

func addPlanFileFlag(flagSet *pflag.FlagSet, p *string) {
	flagSet.StringVarP(p, "plan-file", "f", "kismatic-cluster.yaml", "path to the installation plan file")
}