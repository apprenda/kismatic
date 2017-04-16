package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type dashboardOpts struct {
	planFilename     string
	dashboardURLMode bool
}

// NewCmdDashboard opens or displays the dashboard URL
func NewCmdDashboard(out io.Writer) *cobra.Command {
	opts := &dashboardOpts{}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Opens/displays the kubernetes dashboard URL of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("Unexpected args: %v", args)
			}
			planner := &install.FilePlanner{File: opts.planFilename}
			return doDashboard(out, planner, opts)
		},
	}

	// PersistentFlags
	addPlanFileFlag(cmd.Flags(), &opts.planFilename)
	cmd.Flags().BoolVar(&opts.dashboardURLMode, "url", false, "Display the kubernetes dashboard URL instead of opening it in the default browser")
	return cmd
}

func doDashboard(out io.Writer, planner install.Planner, opts *dashboardOpts) error {
	if !planner.PlanExists() {
		return planFileNotFoundErr{filename: opts.planFilename}
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("Error reading plan file %q: %v", opts.planFilename, err)
	}
	authenticatedURL, err := getAuthenticatedDashboardURL(*plan)
	if err != nil {
		return err
	}
	unauthURL, err := getDashboardURL(*plan)
	if err != nil {
		return err
	}
	// Validate dashboard is accessible
	if err = verifyDashboardConnectivity(authenticatedURL); err != nil {
		return fmt.Errorf("Error verifying connectivity to cluster dashboard: %v", err)
	}
	// Dashboard is accessible.. take action
	// In URL mode or
	// Running in a docker container, can't open a browser from here
	if opts.dashboardURLMode || util.RunningInDocker() {
		fmt.Fprintln(out, unauthURL)
		return nil
	}
	fmt.Fprintln(os.Stdout, "Opening kubernetes dashboard in default browser...")
	if err := browser.OpenURL(authenticatedURL); err != nil {
		// Don't error. Just print a message if something goes wrong
		fmt.Fprintf(os.Stdout, "Unexpected error opening the kubernetes dashboard: %v. You may access it at %q", err, unauthURL)
	}
	return nil
}

func verifyDashboardConnectivity(url string) error {
	status, err := util.HTTPGet(url, 2*time.Second, true)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("Got %d HTTP status code when trying to reach the dashboard at %q", status, url)
	}
	return nil
}

func getAuthenticatedDashboardURL(plan install.Plan) (string, error) {
	if plan.Master.LoadBalancedFQDN == "" {
		return "", errors.New("Master load balanced FQDN is not set in the plan file")
	}
	return fmt.Sprintf("https://admin:%s@%s:6443/ui", plan.Cluster.AdminPassword, plan.Master.LoadBalancedFQDN), nil
}

func getDashboardURL(plan install.Plan) (string, error) {
	if plan.Master.LoadBalancedFQDN == "" {
		return "", errors.New("Master load balanced FQDN is not set in the plan file")
	}
	return fmt.Sprintf("https://%s:6443/ui", plan.Master.LoadBalancedFQDN), nil
}
