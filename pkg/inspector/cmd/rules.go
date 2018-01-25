package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apprenda/kismatic/pkg/inspector/rule"
	"github.com/spf13/cobra"
)

// NewCmdRules returns the "rules" command
func NewCmdRules(out io.Writer) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manipulate the inspector's rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().StringVarP(&file, "file", "f", "inspector-rules.yaml", "file where inspector rules are to be written")
	cmd.AddCommand(NewCmdDumpRules(out, file))
	cmd.AddCommand(NewCmdValidateRules(out, file))
	return cmd
}

// NewCmdDumpRules returns the "dump" command
func NewCmdDumpRules(out io.Writer, file string) *cobra.Command {
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump the inspector rules to a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(file); err == nil && !overwrite {
				return fmt.Errorf("%q already exists. Use --overwrite to overwrite it", file)
			}
			f, err := os.Create(file)
			if err != nil {
				return fmt.Errorf("error creating %q: %v", file, err)
			}
			if err := rule.DumpDefaultRules(f); err != nil {
				return fmt.Errorf("error dumping rules: %v", err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite the destination file if it exists")
	return cmd
}

func NewCmdValidateRules(out io.Writer, file string) *cobra.Command {
	var additionalVars []string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the inspector rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("%q does not exist", file)
			}
			additionalVarsM := make(map[string]string)
			for _, v := range additionalVars {
				kv := strings.Split(v, "=")
				if len(kv) != 2 {
					return fmt.Errorf("invalid key-value %q", v)
				}
				additionalVarsM[kv[0]] = kv[1]
			}
			rules, err := rule.ReadFromFile(file, additionalVarsM)
			if err != nil {
				return err
			}
			if !validateRules(out, rules) {
				return fmt.Errorf("invalid rules found in %q", file)
			}
			fmt.Fprintf(out, "Rules are valid\n")
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&additionalVars, "additional-vars", []string{}, "provide a key=value list to template ruleset")
	return cmd
}

// validates rules, printing error messages to the console
func validateRules(out io.Writer, rules []rule.Rule) bool {
	allOK := true
	for i, r := range rules {
		errs := r.Validate()
		if len(errs) > 0 {
			allOK = false
			fmt.Fprintf(out, "%s (Rule #%d):\n", r.GetRuleMeta().Kind, i+1)
			for _, e := range errs {
				fmt.Fprintf(out, "- %v\n", e)
			}
			fmt.Fprintln(out, "")
		}
	}
	return allOK
}
