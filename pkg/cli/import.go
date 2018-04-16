package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"

	"github.com/spf13/cobra"
)

type importOpts struct {
	srcGeneratedAssetsDir string
	dstGeneratedAssetsDir string
	srcRunsDir            string
	dstRunsDir            string
	// srcKeyFile            string
	// dstKeyFile            string
	srcPlanFilePath string
	dstPlanFilePath string
}

// NewCmdImport imports a cluster plan, and potentially a generated or runs dir
func NewCmdImport(out io.Writer) *cobra.Command {
	opts := &importOpts{}
	cmd := &cobra.Command{
		Use:   "import PLAN_FILE_PATH GENERATED_DIR_PATH",
		Short: "imports a cluster plan file, and generated assets",
		Long: `imports a cluster plan file, and generated assets. 
		The GENERATED_DIR_PATH is the path to the directory where assets generated during the installation process were stored from a previous KET installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Usage()
			}
			opts.srcPlanFilePath = args[0]
			opts.srcGeneratedAssetsDir = args[1]
			fp := install.FilePlanner{File: opts.srcPlanFilePath}
			if !fp.PlanExists() {
				return planFileNotFoundErr{filename: opts.srcPlanFilePath}
			}
			plan, err := fp.Read()
			if err != nil {
				return fmt.Errorf("error reading plan: %v", err)
			}
			clusterName := plan.Cluster.Name
			// Pull destinations from the name

			opts.dstPlanFilePath, opts.dstGeneratedAssetsDir, opts.dstRunsDir = generateDirsFromName(clusterName)
			parent, _ := filepath.Split(opts.dstPlanFilePath)
			if err := os.MkdirAll(parent, 0700); err != nil {
				return fmt.Errorf("error creating destination %s: %v", parent, err)
			}
			exists, err := CheckClusterExists(clusterName)
			if exists {
				if err != nil {
					return fmt.Errorf("cluster with name %s already exists, cannot import: %v", clusterName, err)
				}
				return fmt.Errorf("cluster with name %s already exists, cannot import", clusterName)
			}
			return doImport(out, clusterName, opts)
		},
	}
	cmd.Flags().StringVar(&opts.srcRunsDir, "runs-dir", "", "path to the directory where artifacts created during the installation process were stored")
	return cmd
}

func doImport(out io.Writer, name string, opts *importOpts) error {
	if err := util.CopyDir(opts.srcGeneratedAssetsDir, opts.dstGeneratedAssetsDir); err != nil {
		return fmt.Errorf("error copying from %s to %s: %v", opts.srcGeneratedAssetsDir, opts.dstGeneratedAssetsDir, err)
	}
	fmt.Fprintf(out, "Successfully copied generated dir from %s to %s.\n", opts.srcGeneratedAssetsDir, opts.dstGeneratedAssetsDir)

	if opts.srcRunsDir != "" {
		if err := util.CopyDir(opts.srcRunsDir, opts.dstRunsDir); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcRunsDir, opts.dstRunsDir, err)
		}
		fmt.Fprintf(out, "Successfully copied runs dir from %s to %s.\n", opts.srcRunsDir, opts.dstRunsDir)
	}
	if err := util.CopyDir(opts.srcPlanFilePath, opts.dstPlanFilePath); err != nil {
		return fmt.Errorf("error copying from %s to %s: %v", opts.srcPlanFilePath, opts.dstPlanFilePath, err)
	}
	fmt.Fprintf(out, "Successfully copied plan file from %s to %s.\n", opts.srcPlanFilePath, opts.dstPlanFilePath)

	return nil
}
