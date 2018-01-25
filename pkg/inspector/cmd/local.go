package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/inspector/check"
	"github.com/apprenda/kismatic/pkg/inspector/rule"
	"github.com/spf13/cobra"
)

type localOpts struct {
	outputType                  string
	nodeRoles                   string
	rulesFile                   string
	packageInstallationDisabled bool
	dockerInstallationDisabled  bool
	useUpgradeDefaults          bool
	additionalVariables         map[string]string
}

var localExample = `# Run with a custom rules file
kismatic-inspector local --node-roles master -f inspector-rules.yaml
`

// NewCmdLocal returns the "local" command
func NewCmdLocal(out io.Writer) *cobra.Command {
	opts := localOpts{}
	var additionalVars []string
	cmd := &cobra.Command{
		Use:     "local",
		Short:   "Run the inspector checks against the local host",
		Example: localExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.additionalVariables = make(map[string]string)
			for _, v := range additionalVars {
				kv := strings.Split(v, "=")
				if len(kv) != 2 {
					return fmt.Errorf("invalid key-value %q", v)
				}
				opts.additionalVariables[kv[0]] = kv[1]
			}
			return runLocal(out, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.outputType, "output", "o", "table", "set the result output type. Options are 'json', 'table'")
	cmd.Flags().StringVar(&opts.nodeRoles, "node-roles", "", "comma-separated list of the node's roles. Valid roles are 'etcd', 'master', 'worker'")
	cmd.Flags().StringVarP(&opts.rulesFile, "file", "f", "", "the path to an inspector rules file. If blank, the inspector uses the default rules")
	cmd.Flags().BoolVar(&opts.packageInstallationDisabled, "pkg-installation-disabled", false, "when true, the inspector will ensure that the necessary packages are installed on the node")
	cmd.Flags().BoolVar(&opts.dockerInstallationDisabled, "docker-installation-disabled", false, "when true, the inspector will check for docker packages to be installed")
	cmd.Flags().BoolVarP(&opts.useUpgradeDefaults, "upgrade", "u", false, "use defaults for upgrade, rather than install")
	cmd.Flags().StringSliceVar(&additionalVars, "additional-vars", []string{}, "provide a key=value list to template ruleset")
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
	rules, err := getRulesFromFileOrDefault(out, opts.rulesFile, opts.useUpgradeDefaults, opts.additionalVariables)
	if err != nil {
		return err
	}
	// Set up engine dependencies
	distro, err := check.DetectDistro()
	if err != nil {
		return fmt.Errorf("error running checks locally: %v", err)
	}
	pkgMgr, err := check.NewPackageManager(distro)
	if err != nil {
		return err
	}

	// Create rule engine
	e := rule.Engine{
		RuleCheckMapper: rule.DefaultCheckMapper{
			PackageManager:              pkgMgr,
			PackageInstallationDisabled: opts.packageInstallationDisabled,
			DockerInstallationDisabled:  opts.dockerInstallationDisabled,
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
