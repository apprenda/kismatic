package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type certificatesGenerateOpts struct {
	commonName         string
	validityPeriod     int
	subjAltNames       []string
	organizations      []string
	overwrite          bool
	generatedAssetsDir string
}

// NewCmdGenerate creates a new certificates generate command
func NewCmdGenerate(out io.Writer) *cobra.Command {
	opts := &certificatesGenerateOpts{}

	cmd := &cobra.Command{
		Use:   "generate <name> [options]",
		Short: "Generate a cluster certificate, expects 'ca.pem' and 'ca-key.pem' to be in the --generated-assets-dir",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" {
				cmd.Help()
				return fmt.Errorf("no valid <name> argument provided")
			}
			if len(args) != 1 {
				cmd.Help()
				return fmt.Errorf("invalid arguments provided: %v", args)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.validityPeriod <= 0 {
				cmd.Help()
				return fmt.Errorf("--validity-period must be greater than 0")
			}
			return doCertificatesGenerate(args[0], opts, out)
		},
	}

	cmd.Flags().StringVar(&opts.commonName, "common-name", "", "override the common name. If left blank, will use <name>")
	cmd.Flags().IntVar(&opts.validityPeriod, "validity-period", 365, "specify the number of days this certificate should be valid for. Expiration date will be calculated relative to the machine's clock.")
	cmd.Flags().StringSliceVar(&opts.subjAltNames, "subj-alt-names", []string{}, "comma-separated list of names that should be included in the certificate's subject alternative names field.")
	cmd.Flags().StringSliceVar(&opts.organizations, "organizations", []string{}, "comma-separated list of names that should be included in the certificate's organization field.")
	cmd.Flags().BoolVar(&opts.overwrite, "overwrite", false, "overwrite existing certificate if it already exists in the target directory.")
	cmd.Flags().StringVar(&opts.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")

	return cmd
}

func doCertificatesGenerate(name string, opts *certificatesGenerateOpts, out io.Writer) error {
	ansibleDir := "ansible"
	certsDir := filepath.Join(opts.generatedAssetsDir, "keys")
	pki := &install.LocalPKI{
		CACsr: filepath.Join(ansibleDir, "playbooks", "tls", "ca-csr.json"),
		GeneratedCertsDirectory: certsDir,
		Log: out,
	}
	ca, err := pki.GetClusterCA()
	if err != nil {
		return err
	}
	commonName := opts.commonName
	if commonName == "" {
		commonName = name
	}
	validityPeriod := fmt.Sprintf("%dh", opts.validityPeriod*24)
	exists, err := pki.GenerateCertificate(name, validityPeriod, commonName, opts.subjAltNames, opts.organizations, ca, opts.overwrite)
	if err != nil {
		return err
	}
	if exists && !opts.overwrite {
		util.PrettyPrintWarn(out, "Certficate '%s.pem' already exists in '%s' directory, use --overwrite option", name, opts.generatedAssetsDir)
	} else {
		util.PrettyPrintOk(out, "Certficate '%s.pem' created successfully in '%s' directory", name, opts.generatedAssetsDir)
	}

	return nil
}
