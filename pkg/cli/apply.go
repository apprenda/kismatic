package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type applyCmd struct {
	out                io.Writer
	planner            install.Planner
	executor           install.Executor
	planFile           string
	generatedAssetsDir string
	verbose            bool
	outputFormat       string
	skipPreFlight      bool
}

type applyOpts struct {
	generatedAssetsDir string
	restartServices    bool
	verbose            bool
	outputFormat       string
	skipPreFlight      bool
}

// NewCmdApply creates a cluter using the plan file
func NewCmdApply(out io.Writer, installOpts *installOpts) *cobra.Command {
	applyOpts := applyOpts{}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply your plan file to create a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("Unexpected args: %v", args)
			}
			planner := &install.FilePlanner{File: installOpts.planFilename}
			executorOpts := install.ExecutorOptions{
				GeneratedAssetsDirectory: applyOpts.generatedAssetsDir,
				RestartServices:          applyOpts.restartServices,
				OutputFormat:             applyOpts.outputFormat,
				Verbose:                  applyOpts.verbose,
			}
			executor, err := install.NewExecutor(out, os.Stderr, executorOpts)
			if err != nil {
				return err
			}

			applyCmd := &applyCmd{
				out:                out,
				planner:            planner,
				executor:           executor,
				planFile:           installOpts.planFilename,
				generatedAssetsDir: applyOpts.generatedAssetsDir,
				verbose:            applyOpts.verbose,
				outputFormat:       applyOpts.outputFormat,
				skipPreFlight:      applyOpts.skipPreFlight,
			}
			return applyCmd.run()
		},
	}

	// Flags
	cmd.Flags().StringVar(&applyOpts.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
	cmd.Flags().BoolVar(&applyOpts.restartServices, "restart-services", false, "force restart cluster services (Use with care)")
	cmd.Flags().BoolVar(&applyOpts.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&applyOpts.outputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	cmd.Flags().BoolVar(&applyOpts.skipPreFlight, "skip-preflight", false, "skip pre-flight checks, useful when rerunning kismatic")

	return cmd
}

func (c *applyCmd) run() error {
	// Validate and run pre-flight
	opts := &validateOpts{
		planFile:           c.planFile,
		verbose:            c.verbose,
		outputFormat:       c.outputFormat,
		skipPreFlight:      c.skipPreFlight,
		generatedAssetsDir: c.generatedAssetsDir,
	}
	err := doValidate(c.out, c.planner, opts)
	if err != nil {
		return fmt.Errorf("error validating plan: %v", err)
	}
	plan, _ := c.planner.Read()

	// Perform the installation
	err = c.executor.Install(plan)
	if err != nil {
		return fmt.Errorf("error installing: %v", err)
	}

	// Install Helm
	if plan.Features.PackageManager.Enabled {
		util.PrintHeader(c.out, "Installing Helm on the Cluster", '=')
		helm, err := install.DefaultHelmClient()
		if err != nil {
			return fmt.Errorf("error getting a valid Helm client: %v", err)
		}
		helm.Kubeconfig = filepath.Join(c.generatedAssetsDir, "kubeconfig")
		// On a disconnected install set tiller image with the correct tag
		// Prepend the custom registry address and port
		if plan.ConfgiureDockerWithPrivateRegistry() && plan.Cluster.DisconnectedInstallation {
			helm.TillerImage = fmt.Sprintf("%s:%d/%s", plan.DockerRegistryAddress(), plan.DockerRegistry.Port, helm.TillerImage)
		}

		// Backup helm directory if exists
		if backedup, err := helm.BackupClient(); err != nil {
			return fmt.Errorf("error preparing Helm client: %v", err)
		} else if backedup {
			util.PrettyPrintOk(c.out, "Backed up %q directory", helm.ClientDirectory)
		}
		// Run 'helm init'
		if err := helm.Init(); err != nil {
			util.PrettyPrintErr(c.out, "Installed Helm on the cluster")
			return fmt.Errorf("error installing Helm on the cluster: %v", err)
		}
		util.PrettyPrintOk(c.out, "Installed Tiller (the helm server side component)")

		// HelmRBAC will create a new role RBAC manually
		// TODO remove when https://github.com/kubernetes/helm/issues/2224 gets fully fixed
		if err := c.executor.RunPlay("_helm-rbac.yaml", plan); err != nil {
			return fmt.Errorf("error configuring Helm RBAC: %v", err)
		}
	}

	// Run smoketest
	if err := c.executor.RunSmokeTest(plan); err != nil {
		return fmt.Errorf("error running smoke test: %v", err)
	}

	util.PrintColor(c.out, util.Green, "\nThe cluster was installed successfully!\n\n")

	msg := "- To use the generated kubeconfig file with kubectl:" +
		"\n    * use \"kubectl --kubeconfig %s/kubeconfig\"" +
		"\n    * or copy the config file \"cp %[1]s/kubeconfig ~/.kube/config\"\n"
	util.PrintColor(c.out, util.Blue, msg, c.generatedAssetsDir)
	util.PrintColor(c.out, util.Blue, "- To view the Kubernetes dashboard: \"./kismatic dashboard\"\n")
	util.PrintColor(c.out, util.Blue, "- To SSH into a cluster node: \"./kismatic ssh etcd|master|worker|storage|$node.host\"\n")

	return nil
}
