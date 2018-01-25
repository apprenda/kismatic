package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/inspector"
	"github.com/spf13/cobra"
)

type clientOpts struct {
	outputType          string
	nodeRoles           string
	rulesFile           string
	targetNode          string
	useUpgradeDefaults  bool
	additionalVariables map[string]string
}

var clientExample = `# Run the inspector against an etcd node
kismatic-inspector client 10.0.1.24:9090 --node-roles etcd

# Run the inspector against a remote node, and ask for JSON output
kismatic-inspector client 10.0.1.24:9090 --node-roles etcd -o json

# Run the inspector against a remote node using a custom rules file
kismatic-inspector client 10.0.1.24:9090 -f inspector-rules.yaml --node-roles etcd`

// NewCmdClient returns the "client" command
func NewCmdClient(out io.Writer) *cobra.Command {
	opts := clientOpts{}
	var additionalVars []string
	cmd := &cobra.Command{
		Use:     "client HOST:PORT",
		Short:   "Run the inspector against a remote inspector server.",
		Example: clientExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			// Set the target node as the first argument
			opts.targetNode = args[0]
			opts.additionalVariables = make(map[string]string)
			for _, v := range additionalVars {
				kv := strings.Split(v, "=")
				if len(kv) != 2 {
					return fmt.Errorf("invalid key=value %q", v)
				}
				opts.additionalVariables[kv[0]] = kv[1]
			}
			return runClient(out, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.outputType, "output", "o", "table", "set the result output type. Options are 'json', 'table'")
	cmd.Flags().StringVar(&opts.nodeRoles, "node-roles", "", "comma-separated list of the node's roles. Valid roles are 'etcd', 'master', 'worker'")
	cmd.Flags().StringVarP(&opts.rulesFile, "file", "f", "", "the path to an inspector rules file. If blank, the inspector uses the default rules")
	cmd.Flags().BoolVarP(&opts.useUpgradeDefaults, "upgrade", "u", false, "use defaults for upgrade, rather than install")
	cmd.Flags().StringSliceVar(&additionalVars, "additional-vars", []string{}, "key=value pairs separated by ',' to template ruleset")
	return cmd
}

func runClient(out io.Writer, opts clientOpts) error {
	if err := validateOutputType(opts.outputType); err != nil {
		return err
	}
	if opts.nodeRoles == "" {
		return fmt.Errorf("--node-roles is required")
	}
	roles, err := getNodeRoles(opts.nodeRoles)
	if err != nil {
		return err
	}
	c, err := inspector.NewClient(opts.targetNode, roles)
	if err != nil {
		return fmt.Errorf("error creating inspector client: %v", err)
	}
	rules, err := getRulesFromFileOrDefault(out, opts.rulesFile, opts.useUpgradeDefaults, opts.additionalVariables)
	if err != nil {
		return err
	}

	results, err := c.ExecuteRules(rules)
	if err != nil {
		return fmt.Errorf("error running inspector against remote node: %v", err)
	}
	if err := printResults(out, results, opts.outputType); err != nil {
		return err
	}
	for _, r := range results {
		if !r.Success {
			return errors.New("inspector rules failed")
		}
	}
	return nil
}
