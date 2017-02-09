package cli

import (
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/spf13/cobra"
)

type infoOpts struct {
	planFilename string
}

// NewCmdSSH returns an ssh shell
func NewCmdInfo(out io.Writer) *cobra.Command {
	opts := &infoOpts{}

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display info about nodes in the cluster",
		Long: `will list the external IP addresses of nodes that make up the cluster, along with their current versions & roles.

This will retreived by connecting to each node via ssh`,
		RunE: func(cmd *cobra.Command, args []string) error {

			planner := &install.FilePlanner{File: opts.planFilename}

			return list(planner, out)
		},
	}

	// PersistentFlags
	cmd.PersistentFlags().StringVarP(&opts.planFilename, "plan-file", "f", "kismatic-cluster.yaml", "path to the installation plan file")

	return cmd
}

func list(planner *install.FilePlanner, out io.Writer) error {
	// Check if plan file exists
	if !planner.PlanExists() {
		return fmt.Errorf("plan does not exist")
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}

	lv, err := install.ListVersions(plan)

	fmt.Fprintf(out, "Cluster: ")
	if lv.IsTransitioning {
		fmt.Fprintf(out, "Transitioning from v%v to v%v\n", lv.EarliestVersion, lv.LatestVersion)
	} else {
		fmt.Fprintf(out, "v%v\n", lv.LatestVersion)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Nodes:\n")
	for _, node := range lv.Nodes {
		fmt.Fprintf(out, "  - %v: v%v %v\n", node.IP, node.Version, node.Roles)
	}

	return err
}
