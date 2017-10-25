package rule

import (
	"errors"
	"fmt"
)

// The PackageNotInstalled validates that a specified package in not installed.
type PackageNotInstalled struct {
	Meta
	PackageName              string
	PackageVersion           string
	AcceptablePackageVersion string
}

// Name returns the name of the rule
func (p PackageNotInstalled) Name() string {
	name := fmt.Sprintf(`Package "%s" should not be installed`, p.PackageName)
	if p.AcceptablePackageVersion != "" {
		name = fmt.Sprintf(`%s, only acceptable version is "%s"`, name, p.AcceptablePackageVersion)
	}
	return name
}

// IsRemoteRule returns true if the rule is to be run from outside of the node
func (p PackageNotInstalled) IsRemoteRule() bool { return false }

// Validate the rule
func (p PackageNotInstalled) Validate() []error {
	err := []error{}
	if p.PackageName == "" {
		err = append(err, errors.New("PackageName cannot be empty"))
	}
	if len(err) > 0 {
		return err
	}
	return nil
}
