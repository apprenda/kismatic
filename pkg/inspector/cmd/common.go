package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/inspector/rule"
)

func getNodeRoles(commaSepRoles string) ([]string, error) {
	roles := strings.Split(commaSepRoles, ",")
	for _, r := range roles {
		if r != "etcd" && r != "master" && r != "worker" && r != "ingress" && r != "storage" {
			return nil, fmt.Errorf("%s is not a valid node role", r)
		}
	}
	return roles, nil
}

func getRulesFromFileOrDefault(out io.Writer, file string, useUpgradeRules bool, vars map[string]string) ([]rule.Rule, error) {
	if file != "" {
		rules, err := rule.ReadFromFile(file, vars)
		if err != nil {
			return nil, err
		}
		if ok := validateRules(out, rules); !ok {
			return nil, fmt.Errorf("rules read from %q did not pass validation", file)
		}
		return rules, nil
	}
	if useUpgradeRules {
		return rule.UpgradeRules(vars), nil
	}
	return rule.DefaultRules(vars), nil
}

func validateOutputType(outputType string) error {
	if outputType != "json" && outputType != "table" {
		return fmt.Errorf("output type %q not supported", outputType)
	}
	return nil
}
