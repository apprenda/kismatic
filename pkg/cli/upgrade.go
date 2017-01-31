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
	for _, n := range cv.Nodes {
		if install.IsOlderVersion(n.Version.String()) {
			toUpgrade = append(toUpgrade, n)
		} else {
			fmt.Fprintf(out, "Node %s is at target version %s. Skipping", n.IP, n.Version)
		}
	}
	if len(toUpgrade) == 0 {
		fmt.Fprintln(out, "All nodes are at the target version. Nothing to do.")
		return nil
	}
	// Run the upgrade on the nodes that need it
	if err := executor.UpgradeNodes(*plan, toUpgrade); err != nil {
		return fmt.Errorf("Upgrade failed: %v", err)
	}
	fmt.Fprintln(out)
	util.PrintColor(out, util.Green, "Upgrade complete\n")
	fmt.Fprintln(out)
	return nil
}
