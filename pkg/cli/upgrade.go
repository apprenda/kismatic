package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/apprenda/kismatic/pkg/data"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type upgradeOpts struct {
	generatedAssetsDir string
	verbose            bool
	outputFormat       string
	skipPreflight      bool
	online             bool
	planFile           string
	restartServices    bool
	partialAllowed     bool
	maxParallelWorkers int
	dryRun             bool
}

// NewCmdUpgrade returns the upgrade command
func NewCmdUpgrade(out io.Writer) *cobra.Command {
	var opts upgradeOpts
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade your Kubernetes cluster",
		Long: `Upgrade your Kubernetes cluster.

The upgrade process is applied to each node, one node at a time. If a private docker registry
is being used, the new container images will be pushed by Kismatic before starting to upgrade
nodes.

Nodes in the cluster are upgraded in the following order:

1. Etcd nodes
2. Master nodes
3. Worker nodes (regardless of specialization)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&opts.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
	cmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.PersistentFlags().StringVarP(&opts.outputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	cmd.PersistentFlags().BoolVar(&opts.skipPreflight, "skip-preflight", false, "skip upgrade pre-flight checks")
	cmd.PersistentFlags().BoolVar(&opts.restartServices, "restart-services", false, "force restart cluster services (Use with care)")
	cmd.PersistentFlags().BoolVar(&opts.partialAllowed, "partial-ok", false, "allow the upgrade of ready nodes, and skip nodes that have been deemed unready for upgrade")
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", false, "simulate the upgrade, but don't actually upgrade the cluster")
	addPlanFileFlag(cmd.PersistentFlags(), &opts.planFile)

	// Subcommands
	cmd.AddCommand(NewCmdUpgradeOffline(out, &opts))
	cmd.AddCommand(NewCmdUpgradeOnline(out, &opts))
	return cmd
}

// NewCmdUpgradeOffline returns the command for running offline upgrades
func NewCmdUpgradeOffline(out io.Writer, opts *upgradeOpts) *cobra.Command {
	cmd := cobra.Command{
		Use:   "offline",
		Short: "Perform an offline upgrade of your Kubernetes cluster",
		Long: `Perform an offline upgrade of your Kubernetes cluster.

The offline upgrade is available for those clusters in which safety and availabilty are not a concern.
In this mode, the safety and availability checks will not be performed, nor will the nodes in the cluster
be drained of workloads.

Performing an offline upgrade could result in loss of critical data and reduced service
availability. For this reason, this method should not be used for clusters that are housing
production workloads.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doUpgrade(out, opts)
		},
	}
	cmd.Flags().IntVar(&opts.maxParallelWorkers, "max-parallel-workers", 1, "the maximum number of worker nodes to be upgraded in parallel")
	return &cmd
}

// NewCmdUpgradeOnline returns the command for running online upgrades
func NewCmdUpgradeOnline(out io.Writer, opts *upgradeOpts) *cobra.Command {
	cmd := cobra.Command{
		Use:   "online",
		Short: "Perform an online upgrade of your Kubernetes cluster",
		Long: `Perform an online upgrade of your Kubernetes cluster.

During an online upgrade, Kismatic will run safety and availability checks (see table below) against the
existing cluster before performing the upgrade. If any unsafe condition is detected, a report will
be printed, and the upgrade will not proceed.

If the node under upgrade is a Kubernetes node, it is cordoned and drained of workloads
before any changes are applied.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.online = true
			return doUpgrade(out, opts)
		},
	}
	return &cmd
}

func doUpgrade(out io.Writer, opts *upgradeOpts) error {
	if opts.maxParallelWorkers < 1 {
		return fmt.Errorf("max-parallel-workers must be greater or equal to 1, got: %d", opts.maxParallelWorkers)
	}

	planFile := opts.planFile
	planner := install.FilePlanner{File: planFile}
	executorOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: opts.generatedAssetsDir,
		RestartServices:          opts.restartServices,
		OutputFormat:             opts.outputFormat,
		Verbose:                  opts.verbose,
		DryRun:                   opts.dryRun,
	}
	executor, err := install.NewExecutor(out, os.Stderr, executorOpts)
	if err != nil {
		return err
	}
	preflightExecOpts := executorOpts
	preflightExecOpts.DryRun = false // We always want to run preflight, even if doing a dry-run
	preflightExec, err := install.NewPreFlightExecutor(out, os.Stderr, preflightExecOpts)
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
			util.PrettyPrintOk(out, "- %q is at the target version %q", n.Node.Host, n.Version)
		}
		fmt.Fprintln(out)
	}

	if plan.ConfigureDockerRegistry() && plan.Cluster.DisconnectedInstallation {
		util.PrintHeader(out, "Upgrade: Docker Registry", '=')
		if err = executor.UpgradeDockerRegistry(*plan); err != nil {
			return fmt.Errorf("Failed to upgrade docker registry: %v", err)
		}
	}

	// Print message if there's no work to do
	if len(toUpgrade) == 0 {
		fmt.Fprintln(out, "All nodes are at the target version. Skipping node upgrades.")
	} else {
		if err = upgradeNodes(out, *plan, *opts, toUpgrade, executor, preflightExec); err != nil {
			return err
		}
	}

	if opts.partialAllowed {
		util.PrintColor(out, util.Green, `

Partial upgrade complete.

Cluster level services are still left to upgrade. These can only be upgraded
when performing a full upgrade. When you are ready, you may use "kismatic upgrade"
without the "--partial-ok" flag to perform a full upgrade.

`)
		return nil
	}

	// Upgrade the cluster services
	util.PrintHeader(out, "Upgrade: Cluster Services", '=')
	if err := executor.UpgradeClusterServices(*plan); err != nil {
		return fmt.Errorf("Failed to upgrade cluster services: %v", err)
	}

	if err := executor.RunSmokeTest(plan); err != nil {
		return fmt.Errorf("Smoke test failed: %v", err)
	}

	if !opts.dryRun {
		fmt.Fprintln(out)
		util.PrintColor(out, util.Green, "Upgrade complete\n")
		fmt.Fprintln(out)
	}
	return nil
}

func upgradeNodes(out io.Writer, plan install.Plan, opts upgradeOpts, nodesNeedUpgrade []install.ListableNode, executor install.Executor, preflightExec install.PreFlightExecutor) error {
	// Run safety checks if doing an online upgrade
	unsafeNodes := []install.ListableNode{}
	if opts.online {
		util.PrintHeader(out, "Validate Online Upgrade", '=')
		// Use the first master node for running kubectl
		client, err := plan.GetSSHClient(plan.Master.Nodes[0].Host)
		if err != nil {
			return fmt.Errorf("error getting SSH client: %v", err)
		}
		kubeClient := data.RemoteKubectl{SSHClient: client}
		for _, node := range nodesNeedUpgrade {
			util.PrettyPrint(out, "%s %v", node.Node.Host, node.Roles)
			errs := install.DetectNodeUpgradeSafety(plan, node.Node, kubeClient)
			if len(errs) != 0 {
				util.PrintError(out)
				fmt.Fprintln(out)
				for _, err := range errs {
					fmt.Println("-", err.Error())
				}
				unsafeNodes = append(unsafeNodes, node)
			} else {
				util.PrintOkln(out)
			}
		}
		// If we found any unsafe nodes, and we are not doing a partial upgrade, exit.
		if len(unsafeNodes) > 0 && !opts.partialAllowed {
			return errors.New("Unable to perform an online upgrade due to the unsafe conditions detected.")
		}
		// Block the upgrade if partial is allowed but there is an etcd or master node
		// that cannot be upgraded
		if opts.partialAllowed {
			for _, n := range unsafeNodes {
				for _, r := range n.Roles {
					if r == "master" || r == "etcd" {
						return errors.New("Unable to perform an online upgrade due to the unsafe conditions detected.")
					}
				}
			}
		}
	}

	// Run upgrade preflight on the nodes that are to be upgraded
	unreadyNodes := []install.ListableNode{}
	if !opts.skipPreflight {
		for _, node := range nodesNeedUpgrade {
			util.PrintHeader(out, fmt.Sprintf("Preflight Checks: %s %s", node.Node.Host, node.Roles), '=')
			if err := preflightExec.RunUpgradePreFlightCheck(&plan, node); err != nil {
				// return fmt.Errorf("Upgrade preflight check failed: %v", err)
				unreadyNodes = append(unreadyNodes, node)
			}
		}
	}

	// Block upgrade if we found unready nodes, and we are not doing a partial upgrade
	if len(unreadyNodes) > 0 && !opts.partialAllowed {
		return errors.New("Errors found during preflight checks")
	}

	// Block the upgrade if partial is allowed but there is an etcd or master node
	// that cannot be upgraded
	if opts.partialAllowed {
		for _, n := range unreadyNodes {
			for _, r := range n.Roles {
				if r == "master" || r == "etcd" {
					return errors.New("Errors found during preflight checks")
				}
			}
		}
	}

	// Filter out the nodes that are unsafe/unready
	toUpgrade := []install.ListableNode{}
	for _, n := range nodesNeedUpgrade {
		upgrade := true
		for _, unsafe := range unsafeNodes {
			if unsafe.Node == n.Node {
				upgrade = false
			}
		}
		for _, unready := range unreadyNodes {
			if unready.Node == n.Node {
				upgrade = false
			}
		}
		if upgrade {
			toUpgrade = append(toUpgrade, n)
		}
	}

	// get all etcd nodes
	etcdToUpgrade := install.NodesWithRoles(toUpgrade, "etcd")
	// it's safe to upgrade one node etcd cluster from 2.3 to 3.1
	// it will always be required for this version because all prior ket versions had a etcd2
	if len(etcdToUpgrade) > 1 {
		// Run the upgrade on the nodes to Etcd v3.0.x
		if err := executor.UpgradeEtcd2Nodes(plan, etcdToUpgrade); err != nil {
			return fmt.Errorf("Failed to upgrade etcd2 nodes: %v", err)
		}
	}

	// Run the upgrade on the nodes that need it
	if err := executor.UpgradeNodes(plan, toUpgrade, opts.online, opts.maxParallelWorkers); err != nil {
		return fmt.Errorf("Failed to upgrade nodes: %v", err)
	}
	return nil
}
