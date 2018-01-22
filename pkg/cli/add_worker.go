package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type addWorkerOpts struct {
	Roles                    []string
	NodeLabels               []string
	GeneratedAssetsDirectory string
	RestartServices          bool
	OutputFormat             string
	Verbose                  bool
	SkipPreFlight            bool
}

var validRoles = []string{"worker", "ingress", "storage"}

// NewCmdAddWorker returns the command for adding workers to the cluster
func NewCmdAddWorker(out io.Writer, installOpts *installOpts) *cobra.Command {
	opts := &addWorkerOpts{}
	cmd := &cobra.Command{
		Use:   "add-worker WORKER_NAME WORKER_IP [WORKER_INTERNAL_IP]",
		Short: "add a new node to an existing Kubernetes cluster",
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
			// default to 'worker'
			if len(opts.Roles) == 0 {
				opts.Roles = append(opts.Roles, "worker")
			}
			for _, r := range opts.Roles {
				if !util.Contains(r, validRoles) {
					return fmt.Errorf("invalid role %q, options %v", r, validRoles)
				}
			}
			if len(opts.NodeLabels) > 0 {
				newWorker.Labels = make(map[string]string)
				for _, l := range opts.NodeLabels {
					pair := strings.Split(l, "=")
					if len(pair) != 2 {
						return fmt.Errorf("invalid label %q provided, must be key=value pair", l)
					}
					newWorker.Labels[pair[0]] = pair[1]
				}
			}
			return doAddWorker(out, installOpts.planFilename, opts, newWorker)
		},
	}
	cmd.Flags().StringSliceVar(&opts.Roles, "roles", []string{}, "roles separated by ',' (options \"worker\"|\"ingress\"|\"storage\")")
	cmd.Flags().StringSliceVarP(&opts.NodeLabels, "labels", "l", []string{}, "key=value pairs separated by ','")
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
		return planFileNotFoundErr{filename: planFile}
	}
	execOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: opts.GeneratedAssetsDirectory,
		RestartServices:          opts.RestartServices,
		OutputFormat:             opts.OutputFormat,
		Verbose:                  opts.Verbose,
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
		util.PrintValidationErrors(out, errs)
		return errors.New("information provided about the new node is invalid")
	}
	if _, errs := install.ValidatePlan(plan); errs != nil {
		util.PrintValidationErrors(out, errs)
		return errors.New("the plan file failed validation")
	}
	workerSSHCon := &install.SSHConnection{
		SSHConfig: &plan.Cluster.SSH,
		Node:      &newWorker,
	}
	if _, errs := install.ValidateSSHConnection(workerSSHCon, "New node"); errs != nil {
		util.PrintValidationErrors(out, errs)
		return errors.New("could not establish SSH connection to the new node")
	}
	if err = ensureNodeIsNew(*plan, newWorker); err != nil {
		return err
	}
	if !opts.SkipPreFlight {
		util.PrintHeader(out, "Running Pre-Flight Checks On New Node", '=')
		if err = executor.RunNewNodePreFlightCheck(*plan, newWorker); err != nil {
			return err
		}
	}
	updatedPlan, err := executor.AddNode(plan, newWorker, opts.Roles)
	if err != nil {
		return err
	}
	if err := planner.Write(updatedPlan); err != nil {
		return fmt.Errorf("error updating plan file to include the new node: %v", err)
	}
	return nil
}

// returns an error if the plan contains a worker that is "equivalent"
// to the new node that is being added
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
	for _, n := range plan.Ingress.Nodes {
		if n.Host == newWorker.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another ingress node")
		}
		if n.IP == newWorker.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another ingress node")
		}
		if newWorker.InternalIP != "" && n.InternalIP == newWorker.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another ingress node")
		}
	}
	for _, n := range plan.Storage.Nodes {
		if n.Host == newWorker.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another storage node")
		}
		if n.IP == newWorker.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another storage node")
		}
		if newWorker.InternalIP != "" && n.InternalIP == newWorker.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another storage node")
		}
	}
	return nil
}
