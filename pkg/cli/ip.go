package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/spf13/cobra"
)

type ipOpts struct {
	planFilename string
}

// NewCmdIP prints the cluster's IP
func NewCmdIP(out io.Writer) *cobra.Command {
	opts := &ipOpts{}
	cmd := &cobra.Command{
		Use:   "ip CLUSTER_NAME",
		Short: "retrieve the IP address of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			planPath, _, _ := generateDirsFromName(clusterName)
			opts.planFilename = planPath
			planner := &install.FilePlanner{File: planPath}
			return doIP(out, planner, opts)
		},
	}

	return cmd
}

func doIP(out io.Writer, planner install.Planner, opts *ipOpts) error {
	// Check if plan file exists
	if !planner.PlanExists() {
		return planFileNotFoundErr{filename: opts.planFilename}
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}
	address, err := getClusterAddress(*plan)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, address)
	return nil
}

func getClusterAddress(plan install.Plan) (string, error) {
	if plan.Master.LoadBalancedFQDN == "" {
		return "", errors.New("Master load balanced FQDN is not set in the plan file")
	}
	return plan.Master.LoadBalancedFQDN, nil
}
