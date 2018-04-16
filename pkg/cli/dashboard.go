package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type dashboardOpts struct {
	dashboardURLMode   bool
	generatedAssetsDir string
	planFilename       string
}

const url = "http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login"

// NewCmdDashboard opens or displays the dashboard URL
func NewCmdDashboard(in io.Reader, out io.Writer) *cobra.Command {
	opts := &dashboardOpts{}

	cmd := &cobra.Command{
		Use:   "dashboard CLUSTER_NAME",
		Short: "Opens/displays the kubernetes dashboard URL of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			opts.planFilename, opts.generatedAssetsDir, _ = generateDirsFromName(clusterName)
			return doDashboard(in, out, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dashboardURLMode, "url", false, "Display the kubernetes dashboard URL instead of opening it in the default browser")
	return cmd
}

func doDashboard(in io.Reader, out io.Writer, opts *dashboardOpts) error {
	if opts.dashboardURLMode {
		fmt.Fprintln(out, url)
		return nil
	}

	kubeconfig := filepath.Join(opts.generatedAssetsDir, "kubeconfig")
	if stat, err := os.Stat(kubeconfig); os.IsNotExist(err) || stat.IsDir() {
		return fmt.Errorf("Did not find required kubeconfig file %q", kubeconfig)
	}

	var generateErr error
	adminKubeconfig := filepath.Join(opts.generatedAssetsDir, "dashboard-admin-kubeconfig")
	// Generate dashboard admin certificate if it does not exist
	if _, err := os.Stat(adminKubeconfig); os.IsNotExist(err) {
		planner := &install.FilePlanner{File: opts.planFilename}
		plan, err := planner.Read()
		if err != nil {
			return fmt.Errorf("error reading plan file: %v", err)
		}

		if generateErr = generateDashboardAdminKubeconfig(out, opts.generatedAssetsDir, *plan); generateErr != nil {
			fmt.Fprintf(out, "Error generating a kubeconfig file, you may still use the dashboard with your own ServiceAccount token\n\n")
		}
	}

	fmt.Fprintf(out, "Opening kubernetes dashboard in default browser...\n")
	if generateErr == nil {
		fmt.Fprintf(out, "Use the kubeconfig in %q\n", adminKubeconfig)
	}
	if err := browser.OpenURL(url); err != nil {
		fmt.Fprintf(out, "Unexpected error opening the kubernetes dashboard: %v. You may access it at %q", err, url)
	}

	cmd := exec.Command("./kubectl", "proxy", "--kubeconfig", kubeconfig)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running kubectl proxy: %v", err)
	}

	return nil
}

func generateDashboardAdminKubeconfig(out io.Writer, generatedAssetsDir string, plan install.Plan) error {
	// All of this is required because cannot set a label on the secret so no selectors
	cmd := exec.Command("./kubectl", "-n", "kube-system", "get", "sa", "kubernetes-dashboard-admin", "-o", "jsonpath={.secrets[0].name}", "--kubeconfig", filepath.Join(generatedAssetsDir, "kubeconfig"))
	secret, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting token secret: %v", err)
	}
	if len(secret) == 0 || !strings.Contains(string(secret), "kubernetes-dashboard-admin-token") {
		return fmt.Errorf("kubernetes-dashboard-admin-token secret not found")
	}

	cmd = exec.Command("./kubectl", "-n", "kube-system", "get", "secrets", string(secret), "-o", "jsonpath={.data.token}", "--kubeconfig", filepath.Join(generatedAssetsDir, "kubeconfig"))
	token, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting the token: %v", err)
	}
	if len(token) == 0 {
		return fmt.Errorf("got an empty token")
	}
	err = install.GenerateDashboardAdminKubeconfig(strings.Trim(string(token), "'"), &plan, generatedAssetsDir)
	if err != nil {
		return fmt.Errorf("error generating dashboard-admin kubeconfig file: %v", err)
	}
	return nil
}
