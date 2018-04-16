package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type volumeDeleteOptions struct {
	verbose            bool
	outputFormat       string
	generatedAssetsDir string
	force              bool
}

// NewCmdVolumeDelete returns the command for deleting storage volumes
func NewCmdVolumeDelete(in io.Reader, out io.Writer) *cobra.Command {
	opts := volumeDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete CLUSTER_NAME VOLUME_NAME",
		Short: "delete storage volumes",
		Long: `Delete storage volumes created by the 'volume add' command.
		
WARNING all data in the volume will be lost.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			planPath, generatedPath, _ := generateDirsFromName(clusterName)
			opts.generatedAssetsDir = generatedPath
			if opts.force == false {
				ans, err := util.PromptForString(in, out, "Are you sure you want to delete this volume? All data will be lost", "N", []string{"N", "y"})
				if err != nil {
					return fmt.Errorf("error getting user response: %v", err)
				}
				if strings.ToLower(ans) != "y" {
					os.Exit(0)
				}
			}
			return doVolumeDelete(out, opts, planPath, args)
		},
	}
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options simple|raw)`)
	cmd.Flags().BoolVar(&opts.force, "force", false, `do not prompt`)
	return cmd
}

func doVolumeDelete(out io.Writer, opts volumeDeleteOptions, planFile string, args []string) error {
	// get volume name and size from arguments
	var volumeName string
	switch len(args) {
	case 1:
		volumeName = args[0]
	default:
		return fmt.Errorf("%d arguments were provided, but add does not support more than 1 arguments", len(args))
	}

	// setup ansible for execution
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return planFileNotFoundErr{filename: planFile}
	}
	execOpts := install.ExecutorOptions{
		OutputFormat: opts.outputFormat,
		Verbose:      opts.verbose,
		// Need to refactor executor code... this will do for now as we don't need the generated assets dir in this command
		GeneratedAssetsDirectory: opts.generatedAssetsDir,
	}
	exec, err := install.NewExecutor(out, out, execOpts)
	if err != nil {
		return err
	}
	plan, err := planner.Read()
	if err != nil {
		return err
	}

	// Run validation
	vopts := &validateOpts{
		outputFormat:       opts.outputFormat,
		verbose:            opts.verbose,
		planFile:           planFile,
		skipPreFlight:      true,
		generatedAssetsDir: opts.generatedAssetsDir,
	}
	if err := doValidate(out, planner, vopts); err != nil {
		return err
	}

	if err := exec.DeleteVolume(plan, volumeName); err != nil {
		return fmt.Errorf("error deleting volume: %v", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Successfully deleted the persistent volume in the kubernetes cluster.")
	return nil
}
