package cli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/spf13/cobra"
)

// NewCmdVersion returns the version command
func NewCmdVersion(buildDate string, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "display the Kismatic CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(out, "Kismatic:")
			fmt.Fprintf(out, "  Version: %s\n", install.AboutKismatic.ShortVersion)
			fmt.Fprintf(out, "  Built: %s\n", buildDate)
			fmt.Fprintf(out, "  Go Version: %s\n", runtime.Version())
		},
	}
}
