package rule

import (
	"errors"
	"fmt"
)

// The PackageDependency rule declares a dependency on a software package
// that can be installed via an operating system's package manager.
type PackageDependency struct {
	Meta
	PackageName          string
	PackageVersion       string
	AnyVersion           bool
	ShouldNotBeInstalled bool
}

// Name returns the name of the rule
func (p PackageDependency) Name() string {
	name := fmt.Sprintf(`Package "%s %s"`, p.PackageName, p.PackageVersion)
	if p.AnyVersion {
		name = fmt.Sprintf(`Package "%s"`, p.PackageName)
	}
	return name
}

// IsRemoteRule returns true if the rule is to be run from outside of the node
func (p PackageDependency) IsRemoteRule() bool { return false }

// Validate the rule
func (p PackageDependency) Validate() []error {
	err := []error{}
	if p.PackageName == "" {
		err = append(err, errors.New("PackageName cannot be empty"))
	}
	if !p.AnyVersion && p.PackageVersion == "" {
		err = append(err, errors.New("PackageVersion cannot be empty"))
	}
	if len(err) > 0 {
		return err
	}
	return nil
}
