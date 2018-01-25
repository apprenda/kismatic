package rule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

// ReadFromFile returns the list of rules contained in the specified file
func ReadFromFile(file string, vars map[string]string) ([]Rule, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("%q does not exist", file)
	}
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing file %q: %v", file, err)
	}
	var rawRules bytes.Buffer
	err = tmpl.Execute(&rawRules, vars)
	if err != nil {
		return nil, fmt.Errorf("error reading rules from %q: %v", file, err)
	}
	rules, err := UnmarshalRulesYAML(rawRules.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling rules from %q: %v", file, err)
	}
	// TODO: Validate the rules we just read
	return rules, nil
}

// This catch all rule is used for unmarshaling
// The reason for having this is that we don't know the Kind
// of the rule we are reading before unmarshaling, so we
// unmarshal into this catch all, where all fields are captured.
// There might be a better way of doing this, but taking this
// approach for now...
type catchAllRule struct {
	Meta                     `yaml:",inline"`
	PackageName              string   `yaml:"packageName"`
	PackageVersion           string   `yaml:"packageVersion"`
	AcceptablePackageVersion string   `yaml:"acceptablePackageVersion"`
	Executable               string   `yaml:"executable"`
	Port                     int      `yaml:"port"`
	ProcName                 string   `yaml:"procName"`
	File                     string   `yaml:"file"`
	ContentRegex             string   `yaml:"contentRegex"`
	Timeout                  string   `yaml:"timeout"`
	SupportedVersions        []string `yaml:"supportedVersions"`
	Path                     string   `yaml:"path"`
	MinimumBytes             string   `yaml:"minimumBytes"`
}

// UnmarshalRulesYAML unmarshals the data into a list of rules
func UnmarshalRulesYAML(data []byte) ([]Rule, error) {
	catchAllRules := []catchAllRule{}
	if err := yaml.Unmarshal(data, &catchAllRules); err != nil {
		return nil, err
	}
	return rulesFromCatchAllRules(catchAllRules)
}

// UnmarshalRulesJSON unmarshals the JSON rules into a list of rules
func UnmarshalRulesJSON(data []byte) ([]Rule, error) {
	catchAllRules := []catchAllRule{}
	if err := json.Unmarshal(data, &catchAllRules); err != nil {
		return nil, err
	}
	return rulesFromCatchAllRules(catchAllRules)
}

func rulesFromCatchAllRules(catchAllRules []catchAllRule) ([]Rule, error) {
	rules := []Rule{}
	for _, catchAllRule := range catchAllRules {
		r, err := buildRule(catchAllRule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func buildRule(catchAll catchAllRule) (Rule, error) {
	kind := strings.ToLower(strings.TrimSpace(catchAll.Kind))
	meta := Meta{
		Kind: kind,
		When: catchAll.When,
	}
	switch kind {
	default:
		return nil, fmt.Errorf("rule with kind %q is not supported", catchAll.Kind)
	case "packagedependency":
		r := PackageDependency{
			PackageName:    catchAll.PackageName,
			PackageVersion: catchAll.PackageVersion,
		}
		r.Meta = meta
		return r, nil
	case "packagenotinstalled":
		r := PackageNotInstalled{
			PackageName:              catchAll.PackageName,
			PackageVersion:           catchAll.PackageVersion,
			AcceptablePackageVersion: catchAll.AcceptablePackageVersion,
		}
		r.Meta = meta
		return r, nil
	case "executableinpath":
		r := ExecutableInPath{
			Executable: catchAll.Executable,
		}
		r.Meta = meta
		return r, nil
	case "dockerinpath":
		r := DockerInPath{}
		r.Meta = meta
		return r, nil
	case "tcpportavailable":
		r := TCPPortAvailable{
			Port:     catchAll.Port,
			ProcName: catchAll.ProcName,
		}
		r.Meta = meta
		return r, nil
	case "tcpportaccessible":
		r := TCPPortAccessible{
			Port:    catchAll.Port,
			Timeout: catchAll.Timeout,
		}
		r.Meta = meta
		return r, nil
	case "filecontentmatches":
		r := FileContentMatches{
			File:         catchAll.File,
			ContentRegex: catchAll.ContentRegex,
		}
		r.Meta = meta
		return r, nil
	case "python2version":
		r := Python2Version{
			SupportedVersions: catchAll.SupportedVersions,
		}
		r.Meta = meta
		return r, nil
	case "freespace":
		r := FreeSpace{
			Path:         catchAll.Path,
			MinimumBytes: catchAll.MinimumBytes,
		}
		r.Meta = meta
		return r, nil

	}
}
