package cli

import (
	"io"
	"os"

	"fmt"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
	"strings"
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
	extraVars          []string
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
			ev, err := stepCmd.getExtraVarsMap()
			if err != nil {
				return err
			}
			execOpts := install.ExecutorOptions{
				GeneratedAssetsDirectory: stepCmd.generatedAssetsDir,
				RestartServices:          stepCmd.restartServices,
				OutputFormat:             stepCmd.outputFormat,
				Verbose:                  stepCmd.verbose,
				ExtraVars:                ev,
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
	cmd.Flags().StringSliceVar(&stepCmd.extraVars, "extra-vars", []string{}, "ansible varaibles (comma separated, key=value)")
	cmd.Flags().MarkHidden("extra-vars")
	return cmd
}

func (c stepCmd) run() error {
	valOpts := &validateOpts{
		planFile:      c.planFile,
		verbose:       c.verbose,
		outputFormat:  c.outputFormat,
		skipPreFlight: true,
	}
	if err := doValidate(c.out, c.planner, valOpts); err != nil {
		return err
	}
	plan, err := c.planner.Read()
	if err = c.executor.RunTask(c.task, plan); err != nil {
		return err
	}
	util.PrintColor(c.out, util.Green, "\nTask completed successfully\n\n")
	return nil
}

func (c stepCmd) getExtraVarsMap() (map[string]string, error) {
	ev := map[string]string{}
	for _, kv := range c.extraVars {
		kvSplit := strings.Split(kv, "=")
		if len(kvSplit) != 2 {
			return nil, fmt.Errorf("Bad request in extra-vars with %q", c.extraVars)
		}
		ev[kvSplit[0]] = kvSplit[1]
	}

	return ev, nil
}
