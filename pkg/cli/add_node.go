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

type addNodeOpts struct {
	Roles                    []string
	NodeLabels               []string
	GeneratedAssetsDirectory string
	RestartServices          bool
	OutputFormat             string
	Verbose                  bool
	SkipPreFlight            bool
}

var validRoles = []string{"worker", "ingress", "storage"}

// NewCmdAddNode returns the command for adding node to the cluster
func NewCmdAddNode(out io.Writer, installOpts *installOpts) *cobra.Command {
	opts := &addNodeOpts{}
	cmd := &cobra.Command{
		Use:     "add-node CLUSTER_NAME NODE_NAME NODE_IP [NODE_INTERNAL_IP]",
		Short:   "add a new node to an existing Kubernetes cluster",
		Aliases: []string{"add-worker"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 3 || len(args) > 4 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			var planPath string
			planPath, opts.GeneratedAssetsDirectory, _ = generateDirsFromName(clusterName)
			newNode := install.Node{
				Host: args[1],
				IP:   args[2],
			}
			if len(args) == 3 {
				newNode.InternalIP = args[3]
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
				newNode.Labels = make(map[string]string)
				for _, l := range opts.NodeLabels {
					pair := strings.Split(l, "=")
					if len(pair) != 2 {
						return fmt.Errorf("invalid label %q provided, must be key=value pair", l)
					}
					newNode.Labels[pair[0]] = pair[1]
				}
			}
			return doAddNode(out, planPath, opts, newNode)
		},
	}
	cmd.Flags().StringSliceVar(&opts.Roles, "roles", []string{}, "roles separated by ',' (options \"worker\"|\"ingress\"|\"storage\")")
	cmd.Flags().StringSliceVarP(&opts.NodeLabels, "labels", "l", []string{}, "key=value pairs separated by ','")
	cmd.Flags().BoolVar(&opts.RestartServices, "restart-services", false, "force restart clusters services (Use with care)")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	cmd.Flags().BoolVar(&opts.SkipPreFlight, "skip-preflight", false, "skip pre-flight checks, useful when rerunning kismatic")
	return cmd
}

func doAddNode(out io.Writer, planFile string, opts *addNodeOpts, newNode install.Node) error {
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return planFileNotFoundErr{filename: planFile}
	}
	execOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: opts.GeneratedAssetsDirectory,
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
	if _, errs := install.ValidateNode(&newNode); errs != nil {
		util.PrintValidationErrors(out, errs)
		return errors.New("information provided about the new node is invalid")
	}
	if _, errs := install.ValidatePlan(plan); errs != nil {
		util.PrintValidationErrors(out, errs)
		return errors.New("the plan file failed validation")
	}
	nodeSSHCon := &install.SSHConnection{
		SSHConfig: &plan.Cluster.SSH,
		Node:      &newNode,
	}
	if _, errs := install.ValidateSSHConnection(nodeSSHCon, "New node"); errs != nil {
		util.PrintValidationErrors(out, errs)
		return errors.New("could not establish SSH connection to the new node")
	}
	if err = ensureNodeIsNew(*plan, newNode); err != nil {
		return err
	}
	if !opts.SkipPreFlight {
		util.PrintHeader(out, "Running Pre-Flight Checks On New Node", '=')
		if err = executor.RunNewNodePreFlightCheck(*plan, newNode); err != nil {
			return err
		}
	}
	updatedPlan, err := executor.AddNode(plan, newNode, opts.Roles, opts.RestartServices)
	if err != nil {
		return err
	}
	if err := planner.Write(updatedPlan); err != nil {
		return fmt.Errorf("error updating plan file to include the new node: %v", err)
	}
	return nil
}

// returns an error if the plan contains a node that is "equivalent"
// to the new node that is being added
func ensureNodeIsNew(plan install.Plan, newNode install.Node) error {
	for _, n := range plan.Worker.Nodes {
		if n.Host == newNode.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another worker node")
		}
		if n.IP == newNode.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another worker node")
		}
		if newNode.InternalIP != "" && n.InternalIP == newNode.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another worker node")
		}
	}
	for _, n := range plan.Ingress.Nodes {
		if n.Host == newNode.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another ingress node")
		}
		if n.IP == newNode.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another ingress node")
		}
		if newNode.InternalIP != "" && n.InternalIP == newNode.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another ingress node")
		}
	}
	for _, n := range plan.Storage.Nodes {
		if n.Host == newNode.Host {
			return fmt.Errorf("according to the plan file, the host name of the new node is already being used by another storage node")
		}
		if n.IP == newNode.IP {
			return fmt.Errorf("according to the plan file, the IP of the new node is already being used by another storage node")
		}
		if newNode.InternalIP != "" && n.InternalIP == newNode.InternalIP {
			return fmt.Errorf("according to the plan file, the internal IP of the new node is already being used by another storage node")
		}
	}
	return nil
}
