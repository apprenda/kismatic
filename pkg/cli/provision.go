package cli

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

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
			path, err := os.Getwd()
			if err != nil {
				return err
			}
			//Get the user's name for cluster tagging
			user, err := user.Current()
			if err != nil {
				return err
			}
			tf := provision.Terraform{
				Output:          out,
				BinaryPath:      filepath.Join(path, "terraform/bin/terraform"),
				ClusterOwner:    user.Username,
				KismaticVersion: install.KismaticVersion,
			}
			switch plan.Provisioner.Provider {
			case "aws":
				access := os.Getenv("AWS_ACCESS_KEY_ID")
				secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
				aws := provision.AWS{
					Terraform:       tf,
					AccessKeyID:     access,
					SecretAccessKey: secret,
				}
				updatedPlan, err := aws.Provision(*plan)
				if err != nil {
					return err
				}
				if err := fp.Write(updatedPlan); err != nil {
					return fmt.Errorf("error writing updated plan file to %s: %v", opts.planFilename, err)
				}
				return nil

			default:
				return fmt.Errorf("provider %s not yet supported", plan.Provisioner.Provider)
			}
		},
	}
	return cmd
}

// NewCmdDestroy creates a new destroy command
func NewCmdDestroy(in io.Reader, out io.Writer, opts *installOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy your provisioned cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := &install.FilePlanner{File: opts.planFilename}
			plan, err := fp.Read()
			if err != nil {
				return fmt.Errorf("unable to read plan file: %v", err)
			}
			path, err := os.Getwd()
			if err != nil {
				return err
			}
			tf := provision.Terraform{
				Output:     out,
				BinaryPath: filepath.Join(path, "terraform/bin/terraform"),
			}
			switch plan.Provisioner.Provider {
			case "aws":
				access := os.Getenv("AWS_ACCESS_KEY_ID")
				secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
				aws := provision.AWS{
					Terraform:       tf,
					AccessKeyID:     access,
					SecretAccessKey: secret,
				}
				return aws.Destroy(plan.Cluster.Name)
			default:
				return fmt.Errorf("provider %s not yet supported", plan.Provisioner.Provider)
			}

		},
	}
	return cmd
}
