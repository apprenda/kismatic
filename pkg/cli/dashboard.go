package cli

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
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

	req, err := getDashboardRequest(*plan)
	if err != nil {
		return err
	}
	// Validate dashboard is accessible
	if err = verifyDashboardConnectivity(req); err != nil {
		return fmt.Errorf("Error verifying connectivity to cluster dashboard: %v", err)
	}
	// Dashboard is accessible.. take action
	if opts.dashboardURLMode {
		fmt.Fprintln(out, req.URL)
		return nil
	}
	fmt.Fprintln(out, "Opening kubernetes dashboard in default browser...")
	//Not obvious, but this is for escaping userinfo
	urlFmted := fmt.Sprintf("https://%s@%s:6443/ui", url.UserPassword("admin", plan.Cluster.AdminPassword), plan.Master.LoadBalancedFQDN)
	if err := browser.OpenURL(urlFmted); err != nil {
		// Don't error. Just print a message if something goes wrong
		fmt.Fprintf(out, "Unexpected error opening the kubernetes dashboard: %v. You may access it at %q", err, req.URL)
	}
	return nil
}

func verifyDashboardConnectivity(req *http.Request) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//This was always set to true within the http util, but probably worth adding as a flag?
	}
	client := http.Client{
		Timeout:   2 * time.Second,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed with error: %q", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got %d HTTP status code when trying to reach the dashboard at %q", resp.StatusCode, req.URL)
	}
	return nil
}

func getDashboardRequest(plan install.Plan) (*http.Request, error) {
	if plan.Master.LoadBalancedFQDN == "" {
		return nil, errors.New("master load balanced FQDN is not set in the plan file")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:6443/ui", plan.Master.LoadBalancedFQDN), nil)
	if err != nil {
		return nil, fmt.Errorf("request failed with error: %q", err)
	}
	req.SetBasicAuth("admin", plan.Cluster.AdminPassword)
	return req, nil
}
