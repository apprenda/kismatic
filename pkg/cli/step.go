package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type stepCmd struct {
	out      io.Writer
	planFile string
	task     string
	planner  install.Planner
	executor install.Executor

	// Flags
	generatedAssetsDir string
	restartServices    bool
	verbose            bool
	outputFormat       string
}

// NewCmdStep returns the step command
func NewCmdStep(out io.Writer, opts *installOpts) *cobra.Command {
	stepCmd := &stepCmd{
		out: out,
	}
	cmd := &cobra.Command{
		Use:   "step CLUSTER_NAME PLAY_NAME",
		Short: "run a specific task of the installation workflow (debug feature)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			planPath, generatedPath, runsPath := generateDirsFromName(clusterName)
			execOpts := install.ExecutorOptions{
				GeneratedAssetsDirectory: generatedPath,
				OutputFormat:             stepCmd.outputFormat,
				Verbose:                  stepCmd.verbose,
				RunsDirectory:            runsPath,
			}
			executor, err := install.NewExecutor(out, os.Stderr, execOpts)
			if err != nil {
				return err
			}
			stepCmd.task = fmt.Sprintf("_%s.yaml", args[1])
			stepCmd.planFile = planPath
			stepCmd.planner = &install.FilePlanner{File: planPath}
			stepCmd.executor = executor
			return stepCmd.run()
		},
	}
	cmd.Flags().BoolVar(&stepCmd.restartServices, "restart-services", false, "force restart cluster services (Use with care)")
	cmd.Flags().BoolVar(&stepCmd.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&stepCmd.outputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")
	return cmd
}

func (c stepCmd) run() error {
	valOpts := &validateOpts{
		planFile:           c.planFile,
		verbose:            c.verbose,
		outputFormat:       c.outputFormat,
		skipPreFlight:      true,
		generatedAssetsDir: c.generatedAssetsDir,
	}
	if err := doValidate(c.out, c.planner, valOpts); err != nil {
		return err
	}
	plan, err := c.planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}
	util.PrintHeader(c.out, "Running Task", '=')
	if err := c.executor.RunPlay(c.task, plan, c.restartServices); err != nil {
		return err
	}
	util.PrintColor(c.out, util.Green, "\nTask completed successfully\n\n")
	return nil
}
