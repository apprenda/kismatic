package cli

import (
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

			url, err := doDashboard(out, planner, opts)
			if err != nil {
				return err
			}

			if opts.dashboardURLMode {
				fmt.Fprintln(out, url)
			} else {
				fmt.Fprintln(os.Stdout, "Opening kubernetes dashboard in default browser...")
				err := browser.OpenURL(url)
				// Don't exit just print a message
				if err != nil {
					fmt.Fprintf(os.Stdout, "Unexpected error opening the kubernetes dashboard URL at %q", url)
				}
			}

			return nil
		},
	}

	// PersistentFlags
	cmd.PersistentFlags().StringVarP(&opts.planFilename, "plan-file", "f", "kismatic-cluster.yaml", "path to the installation plan file")

	cmd.Flags().BoolVar(&opts.dashboardURLMode, "url", false, "Display the kubernetes dashboard URL in the CLI instead of opening it in the default browser")

	return cmd
}

func doDashboard(out io.Writer, planner install.Planner, opts *dashboardOpts) (string, error) {
	// Check if plan file exists
	if !planner.PlanExists() {
		return "", fmt.Errorf("plan does not exist")
	}
	plan, err := planner.Read()
	if err != nil {
		return "", fmt.Errorf("error reading plan file: %v", err)
	}

	url, err := planner.GetDashboardURL(plan)
	if err != nil {
		return "", fmt.Errorf("Error getting dashboard URL: %v", err)
	}

	status, err := util.HTTPGet(url, 2*time.Second, true)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Check that the cluster and the dashboard is running and accessible via %q\n", url)
		return "", fmt.Errorf("Error trying to reach cluster dashboard: %v", err)
	}
	if status != http.StatusOK {
		fmt.Fprintf(os.Stdout, "Got %d HTTP status code when trying to reach the dashboard URL at %q\n", status, url)
		return "", fmt.Errorf("Error trying to reach cluster dashboard")
	}

	return url, nil
}
