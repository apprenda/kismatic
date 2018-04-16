package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type diagsOpts struct {
	planFilename string
	verbose      bool
	outputFormat string
}

// NewCmdDiagnostic collects diagnostic data on remote nodes
func NewCmdDiagnostic(out io.Writer) *cobra.Command {
	opts := &diagsOpts{}

	cmd := &cobra.Command{
		Use:   "diagnose CLUSTER_NAME",
		Short: "Collects diagnostics about the nodes in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			opts.planFilename, _, _ = generateDirsFromName(clusterName)
			return doDiagnostics(out, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging from the installation")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", "installation output format (options \"simple\"|\"raw\")")

	return cmd
}

func doDiagnostics(out io.Writer, opts *diagsOpts) error {
	util.PrintHeader(out, "Gathering Diagnostic Data", '=')

	planFile := opts.planFilename
	planner := install.FilePlanner{File: planFile}

	// Read plan file
	if !planner.PlanExists() {
		util.PrettyPrintErr(out, "Reading plan file")
		return planFileNotFoundErr{filename: planFile}
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

	// Get diagnostics from nodes
	options := install.ExecutorOptions{
		OutputFormat: opts.outputFormat,
		Verbose:      opts.verbose,
	}
	executor, err := install.NewDiagnosticsExecutor(out, os.Stderr, options)
	if err != nil {
		return err
	}

	if err := executor.DiagnoseNodes(*plan); err != nil {
		return err
	}

	util.PrintColor(out, util.Green, "\nFinished gathering diagnostic information from all nodes in the cluster.\n")
	util.PrintColor(out, util.Green, "You may find the diagnostic data in the ./diagnostics directory.\n\n")

	return nil
}
