package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type infoOpts struct {
	planFilename string
	outputFormat string
}

// NewCmdInfo returns the info command
func NewCmdInfo(out io.Writer) *cobra.Command {
	opts := &infoOpts{}
	cmd := &cobra.Command{
		Use:   "info CLUSTER_NAME",
		Short: "Display info about nodes in the cluster",
		Long: `will list the nodes that make up the cluster, along with their current versions & roles.

This will be retrieved by connecting to each node via ssh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			opts.planFilename, _, _ = generateDirsFromName(clusterName)
			return list(out, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options "simple"|"json")`)
	return cmd
}

func list(out io.Writer, opts *infoOpts) error {
	// Check if plan file exists
	planner := &install.FilePlanner{File: opts.planFilename}
	if !planner.PlanExists() {
		return fmt.Errorf("plan does not exist")
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}

	// Validate just the nodes
	if ok, errs := install.ValidateNodes(plan.GetUniqueNodes()); !ok {
		util.PrintValidationErrors(out, errs)
		return fmt.Errorf("error validating nodes")
	}

	// Validate SSH connections
	if ok, errs := install.ValidatePlanSSHConnections(plan); !ok {
		util.PrintValidationErrors(out, errs)
		return fmt.Errorf("error getting info from cluster nodes")
	}

	lv, err := install.ListVersions(plan)
	if err != nil {
		return fmt.Errorf("error getting version: %v", err)
	}

	if opts.outputFormat == "json" {
		b, err := json.MarshalIndent(lv, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling struct: %v", err)
		}
		fmt.Fprintln(out, string(b))
		return nil
	}

	fmt.Fprintf(out, "Cluster Version: ")
	if lv.IsTransitioning {
		fmt.Fprintf(out, "Transitioning from v%v to v%v\n", lv.EarliestVersion, lv.LatestVersion)
	} else {
		fmt.Fprintf(out, "v%v\n", lv.LatestVersion)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Nodes:\n")
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprint(w, "Name\tIP\tRoles\tKismatic Version\n")
	for _, listNode := range lv.Nodes {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", listNode.Node.Host, listNode.Node.IP, strings.Join(listNode.Roles, ","), listNode.Version)
	}
	return w.Flush()
}
