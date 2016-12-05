package cli

import (
	"fmt"
	"io"

	"os"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type validateOpts struct {
	planFile      string
	verbose       bool
	outputFormat  string
	skipPreFlight bool
}

// NewCmdValidate creates a new install validate command
func NewCmdValidate(out io.Writer, installOpts *installOpts) *cobra.Command {
	opts := &validateOpts{}
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "validate your plan file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("Unexpected args: %v", args)
			}
			planner := &install.FilePlanner{File: installOpts.planFilename}
			opts.planFile = installOpts.planFilename
			return doValidate(out, planner, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", "installation output format (options simple|raw)")
	cmd.Flags().BoolVar(&opts.skipPreFlight, "skip-preflight", false, "skip pre-flight checks")
	return cmd
}

func doValidate(out io.Writer, planner install.Planner, opts *validateOpts) error {
	util.PrintHeader(out, "Validating", '=')
	// Check if plan file exists
	if !planner.PlanExists() {
		util.PrettyPrintErr(out, "Reading installation plan file [ERROR]")
		fmt.Fprintln(out, "Run \"kismatic install plan\" to generate it")
		return fmt.Errorf("plan does not exist")
	}
	plan, err := planner.Read()
	if err != nil {
		util.PrettyPrintErr(out, "Reading installation plan file %q", opts.planFile)
		return fmt.Errorf("error reading plan file: %v", err)
	}
	util.PrettyPrintOk(out, "Reading installation plan file %q", opts.planFile)

	// Validate plan file
	ok, errs := install.ValidatePlan(plan)
	if !ok {
		util.PrettyPrintErr(out, "Validating installation plan file")
		printValidationErrors(out, errs)
		return fmt.Errorf("validation error prevents installation from proceeding")
	}
	util.PrettyPrintOk(out, "Validating installation plan file")

	// Validate SSH connections
	ok, errs = install.ValidatePlanSSHConnection(plan)
	if !ok {
		util.PrettyPrintErr(out, "Validating SSH connections to nodes")
		printValidationErrors(out, errs)
		return fmt.Errorf("node SSH error prevents installation from proceeding")
	}

	if opts.skipPreFlight {
		return nil
	}
	// Run pre-flight
	options := install.ExecutorOptions{
		OutputFormat: opts.outputFormat,
		Verbose:      opts.verbose,
	}
	e, err := install.NewPreFlightExecutor(out, os.Stderr, options)
	if err != nil {
		return err
	}
	if err = e.RunPreFlightCheck(plan); err != nil {
		return err
	}
	return nil
}

func printValidationErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		util.PrintColor(out, util.Red, "- %v\n", err)
	}
}
