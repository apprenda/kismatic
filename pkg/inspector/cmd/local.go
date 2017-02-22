package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/inspector/check"
	"github.com/apprenda/kismatic/pkg/inspector/rule"
	"github.com/spf13/cobra"
)

type localOpts struct {
	outputType         string
	nodeRoles          string
	rulesFile          string
	enforcePackages    bool
	useUpgradeDefaults bool
}

var localExample = `# Run with a custom rules file
kismatic-inspector local --node-roles master -f inspector-rules.yaml
`

// NewCmdLocal returns the "local" command
func NewCmdLocal(out io.Writer) *cobra.Command {
	opts := localOpts{}
	cmd := &cobra.Command{
		Use:     "local",
		Short:   "Run the inspector checks against the local host",
		Example: localExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLocal(out, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.outputType, "output", "o", "table", "set the result output type. Options are 'json', 'table'")
	cmd.Flags().StringVar(&opts.nodeRoles, "node-roles", "", "comma-separated list of the node's roles. Valid roles are 'etcd', 'master', 'worker'")
	cmd.Flags().StringVarP(&opts.rulesFile, "file", "f", "", "the path to an inspector rules file. If blank, the inspector uses the default rules")
	cmd.Flags().BoolVarP(&opts.enforcePackages, "enforcePackages", "e", false, "when provided the installer will test that all Kismatic packages have been installed")
	cmd.Flags().BoolVarP(&opts.useUpgradeDefaults, "upgrade", "u", false, "use defaults for upgrade, rather than install")
	return cmd
}

func runLocal(out io.Writer, opts localOpts) error {
	if opts.nodeRoles == "" {
		return fmt.Errorf("node role is required")
	}
	roles, err := getNodeRoles(opts.nodeRoles)
	if err != nil {
		return err
	}
	if err = validateOutputType(opts.outputType); err != nil {
		return err
	}
	// Gather rules
	rules, err := getRulesFromFileOrDefault(out, opts.rulesFile, opts.useUpgradeDefaults)
	if err != nil {
		return err
	}
	// Set up engine dependencies
	distro, err := check.DetectDistro()
	if err != nil {
		return fmt.Errorf("error running checks locally: %v", err)
	}
	pkgMgr, err := check.NewPackageManager(distro, opts.enforcePackages)
	if err != nil {
		return err
	}

	// Create rule engine
	e := rule.Engine{
		RuleCheckMapper: rule.DefaultCheckMapper{
			PackageManager: pkgMgr,
		},
	}
	labels := append(roles, string(distro))
	results, err := e.ExecuteRules(rules, labels)
	if err != nil {
		return fmt.Errorf("error running local rules: %v", err)
	}
	if err := printResults(out, results, opts.outputType); err != nil {
		return fmt.Errorf("error printing results: %v", err)
	}
	for _, r := range results {
		if !r.Success {
			return errors.New("inspector rules failed")
		}
	}
	return nil
}
