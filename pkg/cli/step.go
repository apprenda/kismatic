package cli

import (
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
		out:      out,
		planFile: opts.planFilename,
	}
	cmd := &cobra.Command{
		Use:   "step PLAY_NAME",
		Short: "run a specific task of the installation workflow (debug feature)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			execOpts := install.ExecutorOptions{
				GeneratedAssetsDirectory: stepCmd.generatedAssetsDir,
				RestartServices:          stepCmd.restartServices,
				OutputFormat:             stepCmd.outputFormat,
				Verbose:                  stepCmd.verbose,
			}
			executor, err := install.NewExecutor(out, os.Stderr, execOpts)
			if err != nil {
				return err
			}
			stepCmd.task = args[0]
			stepCmd.planFile = opts.planFilename
			stepCmd.planner = &install.FilePlanner{File: stepCmd.planFile}
			stepCmd.executor = executor
			return stepCmd.run()
		},
	}
	cmd.Flags().StringVar(&stepCmd.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
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
	if err = c.executor.RunPlay(c.task, plan); err != nil {
		return err
	}
	util.PrintColor(c.out, util.Green, "\nTask completed successfully\n\n")
	return nil
}
