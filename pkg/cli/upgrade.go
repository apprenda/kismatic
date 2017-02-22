package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

// NewCmdUpgrade returns the upgrade command
func NewCmdUpgrade(out io.Writer) *cobra.Command {
	var planFile string
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade your Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Subcommands
	cmd.AddCommand(NewCmdUpgradeOffline(out, &planFile))
	addPlanFileFlag(cmd.PersistentFlags(), &planFile)
	return cmd
}

type upgradeOpts struct {
	generatedAssetsDir string
	restartServices    bool
	verbose            bool
	outputFormat       string
	skipPreflight      bool
}

// NewCmdUpgradeOffline returns the command for running offline upgrades
func NewCmdUpgradeOffline(out io.Writer, planFile *string) *cobra.Command {
	opts := upgradeOpts{}
	cmd := cobra.Command{
		Use:   "offline",
		Short: "perform an offline upgrade of your Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doUpgradeOffline(out, *planFile, opts)
		},
	}
	cmd.Flags().StringVar(&opts.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
	cmd.Flags().BoolVar(&opts.restartServices, "restart-services", false, "force restart cluster services (Use with care)")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	cmd.Flags().BoolVar(&opts.skipPreflight, "skip-preflight", false, "skip upgrade pre-flight checks")
	return &cmd
}

func doUpgradeOffline(out io.Writer, planFile string, opts upgradeOpts) error {
	planner := install.FilePlanner{File: planFile}
	executorOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: opts.generatedAssetsDir,
		RestartServices:          opts.restartServices,
		OutputFormat:             opts.outputFormat,
		Verbose:                  opts.verbose,
	}
	executor, err := install.NewExecutor(out, os.Stderr, executorOpts)
	if err != nil {
		return err
	}
	util.PrintHeader(out, "Computing upgrade plan", '=')

	// Read plan file
	if !planner.PlanExists() {
		util.PrettyPrintErr(out, "Reading plan file")
		return fmt.Errorf("plan file %q does not exist", planFile)
	}
	util.PrettyPrintOk(out, "Reading plan file")
	plan, err := planner.Read()
	if err != nil {
		util.PrettyPrintErr(out, "Reading plan file")
		return fmt.Errorf("error reading plan file %q: %v", planFile, err)
	}

	// Validate SSH connectivity to nodes
	if ok, errs := install.ValidatePlanSSHConnections(plan); !ok {
		util.PrettyPrintErr(out, "Validate SSH connectivity to nodes")
		util.PrintValidationErrors(out, errs)
		return fmt.Errorf("SSH connectivity validation errors found")
	}
	util.PrettyPrintOk(out, "Validate SSH connectivity to nodes")

	// Figure out which nodes to upgrade
	cv, err := install.ListVersions(plan)
	if err != nil {
		return fmt.Errorf("error listing cluster versions: %v", err)
	}
	var toUpgrade []install.ListableNode
	var toSkip []install.ListableNode
	for _, n := range cv.Nodes {
		if install.IsOlderVersion(n.Version) {
			toUpgrade = append(toUpgrade, n)
		} else {
			toSkip = append(toSkip, n)
		}
	}

	// Print the nodes that will be skipped
	if len(toSkip) > 0 {
		util.PrintHeader(out, "Skipping nodes", '=')
		for _, n := range toSkip {
			util.PrettyPrintOk(out, "- %q is at the target version %q", n.Node.IP, n.Version)
		}
		fmt.Fprintln(out)
	}

	// Print message if there's no work to do
	if len(toUpgrade) == 0 {
		fmt.Fprintln(out, "All nodes are at the target version. Skipping node upgrades.")
	} else {
		// Run upgrade preflight on the nodes that are to be UpgradeNodes
		if !opts.skipPreflight {
			util.PrintHeader(out, "Validating nodes", '=')
			if err := executor.RunUpgradePreFlightCheck(plan); err != nil {
				return fmt.Errorf("Upgrade preflight check failed: %v", err)
			}
		}
		// Run the upgrade on the nodes that need it
		if err := executor.UpgradeNodes(*plan, toUpgrade); err != nil {
			return fmt.Errorf("Failed to upgrade nodes: %v", err)
		}
		// validate on the master
		util.PrintHeader(out, "Validate Kubernetes Control Plane", '=')
		if err := executor.ValidateControlPlane(*plan); err != nil {
			return fmt.Errorf("Failed to validate kuberntes control plane: %v", err)
		}
	}

	if plan.ConfigureDockerRegistry() && plan.Cluster.DisconnectedInstallation {
		util.PrintHeader(out, "Upgrade Docker Registry", '=')
		if err := executor.UpgradeDockerRegistry(*plan); err != nil {
			return fmt.Errorf("Failed to upgrade docker registry: %v", err)
		}
	}

	// Upgrade the cluster services
	util.PrintHeader(out, "Upgrade Cluster Services", '=')
	if err := executor.UpgradeClusterServices(*plan); err != nil {
		return fmt.Errorf("Failed to upgrade cluster services: %v", err)
	}

	util.PrintHeader(out, "Smoke Test Cluster", '=')
	if err := executor.RunSmokeTest(plan); err != nil {
		return fmt.Errorf("Smoke test failed: %v", err)
	}

	fmt.Fprintln(out)
	util.PrintColor(out, util.Green, "Upgrade complete\n")
	fmt.Fprintln(out)
	return nil
}
