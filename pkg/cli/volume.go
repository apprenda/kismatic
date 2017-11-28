package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// NewCmdVolume returns the storage command
func NewCmdVolume(in io.Reader, out io.Writer) *cobra.Command {
	var planFile string
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "manage storage volumes on your Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	addPlanFileFlag(cmd.PersistentFlags(), &planFile)
	cmd.AddCommand(NewCmdVolumeAdd(out, &planFile))
	cmd.AddCommand(NewCmdVolumeList(out, &planFile))
	cmd.AddCommand(NewCmdVolumeDelete(in, out, &planFile))
	return cmd
}
