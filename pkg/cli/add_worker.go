package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type addWorkerOpts struct {
	GeneratedAssetsDirectory string
	RestartServices          bool
	OutputFormat             string
	Verbose                  bool
	SkipPreFlight            bool
}

// NewCmdAddWorker returns the command for adding workers to the cluster
func NewCmdAddWorker(out io.Writer, installOpts *installOpts) *cobra.Command {
	opts := &addWorkerOpts{}
	cmd := &cobra.Command{
		Use:   "add-worker WORKER_NAME WORKER_IP [WORKER_INTERNAL_IP]",
		Short: "add a Worker node to an existing Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 || len(args) > 3 {
				return cmd.Usage()
			}
			newWorker := install.Node{
				Host: args[0],
				IP:   args[1],
			}
			if len(args) == 3 {
				newWorker.InternalIP = args[2]
			}
			return doAddWorker(out, installOpts.planFilename, opts, newWorker)
		},
	}
	cmd.Flags().StringVar(&opts.GeneratedAssetsDirectory, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
	cmd.Flags().BoolVar(&opts.RestartServices, "restart-services", false, "force restart clusters services (Use with care)")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	cmd.Flags().BoolVar(&opts.SkipPreFlight, "skip-preflight", false, "skip pre-flight checks, useful when rerunning kismatic")
	return cmd
}

func doAddWorker(out io.Writer, planFile string, opts *addWorkerOpts, newWorker install.Node) error {
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return errors.New("add-worker can only be used with an existin plan file")
	}
	execOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: opts.GeneratedAssetsDirectory,
		RestartServices:          opts.RestartServices,
		OutputFormat:             opts.OutputFormat,
		Verbose:                  opts.Verbose,
		SkipCAGeneration:         true,
	}
	executor, err := install.NewExecutor(out, os.Stderr, execOpts)
	if err != nil {
		return err
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("failed to read plan file: %v", err)
	}
	if _, errs := install.ValidateNode(&newWorker); errs != nil {
		printValidationErrors(out, errs)
		return errors.New("information provided about the new worker node is invalid")
	}
	if _, errs := install.ValidatePlan(plan); errs != nil {
		printValidationErrors(out, errs)
		return errors.New("the plan file failed validation")
	}
	workerSSHCon := &install.SSHConnection{
		SSHConfig: &plan.Cluster.SSH,
		Nodes:     []install.Node{newWorker},
	}
	if _, errs := install.ValidateSSHConnection(workerSSHCon, "New worker node"); errs != nil {
		printValidationErrors(out, errs)
		return errors.New("could not establish SSH connection to the new node")
	}
	if err := ensureNodeIsNew(*plan, newWorker); err != nil {
		return err
	}
	if !opts.SkipPreFlight {
		util.PrintHeader(out, "Running Pre-Flight Checks On New Worker", '=')
		if err := runPreFlightOnWorker(executor, *plan, newWorker); err != nil {
			return err
		}
	}
	updatedPlan, err := executor.AddWorker(plan, newWorker)
	if err != nil {
		return err
	}
	if err := planner.Write(updatedPlan); err != nil {
		return fmt.Errorf("error updating plan file to inlcude new worker node: %v", err)
	}
	return nil
}

// returns an error if the plan contains a worker that is "equivalent"
// to the new worker that is being added
func ensureNodeIsNew(plan install.Plan, newWorker install.Node) error {
	for _, n := range plan.Worker.Nodes {
		if n.Host == newWorker.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another worker node")
		}
		if n.IP == newWorker.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another worker node")
		}
		if newWorker.InternalIP != "" && n.InternalIP == newWorker.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another worker node")
		}
	}
	return nil
}

func runPreFlightOnWorker(executor install.Executor, plan install.Plan, workerNode install.Node) error {
	// use the original plan, but only run against the new worker
	preFlightPlan := plan
	preFlightPlan.Master.Nodes = []install.Node{}
	preFlightPlan.Master.ExpectedCount = 0
	preFlightPlan.Etcd.Nodes = []install.Node{}
	preFlightPlan.Etcd.ExpectedCount = 0
	preFlightPlan.Worker.Nodes = []install.Node{workerNode}
	preFlightPlan.Worker.ExpectedCount = 1
	preFlightPlan.Ingress.Nodes = []install.Node{}
	preFlightPlan.Ingress.ExpectedCount = 0
	return executor.RunPreFlightCheck(&preFlightPlan)
}
