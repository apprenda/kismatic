package cli

import (
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/provision"

	"github.com/spf13/cobra"
)

// NewCmdProvision creates a new provision command
func NewCmdProvision(in io.Reader, out io.Writer, opts *installOpts) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "provision your Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := &install.FilePlanner{File: opts.planFilename}
			plan, err := fp.Read()
			if err != nil {
				return fmt.Errorf("unable to read plan file: %v", err)
			}
			tf := provision.NewTerraform(nil)
			switch plan.Provisioner.Provider {
			case "aws":
				aws := provision.AWS{Terraform: *tf}
				return aws.Provision(out, plan)
			default:
				return fmt.Errorf("provider %s not yet supported", plan.Provisioner.Provider)
			}
		},
	}
	cmd.Flags().StringVarP(&opts.planFilename, "plan-file", "f", "kismatic-cluster", "name of the kismatic cluster to create.)")

	return cmd
}

// NewCmdDestroy creates a new destroy command
func NewCmdDestroy(in io.Reader, out io.Writer, opts *installOpts) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy your provisioned cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := &install.FilePlanner{File: fmt.Sprintf("terraform/clusters/%s/%s.yaml", opts.planFilename, opts.planFilename)}
			plan, err := fp.Read()
			if err != nil {
				return fmt.Errorf("unable to read plan file: %v", err)
			}
			tf := provision.NewTerraform(nil)
			switch plan.Provisioner.Provider {
			case "aws:":
				aws := provision.AWS{Terraform: *tf}
				return aws.Destroy(out, opts.planFilename)
			default:
				return fmt.Errorf("provider %s not yet supported", plan.Provisioner.Provider)
			}

		},
	}
	cmd.Flags().StringVarP(&opts.planFilename, "plan-file", "f", "kismatic-cluster", "name of the kismatic cluster to destroy.)")
	return cmd
}
