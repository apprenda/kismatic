package rule

import (
	"errors"
	"fmt"
)

// PackageAvailable is a rule that ensures that the given package
// is either installed on the host operating system
// or available for installation
type PackageAvailable struct {
	Meta
	PackageName    string
	PackageVersion string
}

// Name returns the name of the rule
func (p PackageAvailable) Name() string {
	return fmt.Sprintf("Package Available: %s %s", p.PackageName, p.PackageVersion)
}

// IsRemoteRule returns true if the rule is to be run from outside of the node
func (p PackageAvailable) IsRemoteRule() bool { return false }

// Validate the rule
func (p PackageAvailable) Validate() []error {
	err := []error{}
	if p.PackageName == "" {
		err = append(err, errors.New("PackageName cannot be empty"))
	}
	if p.PackageVersion == "" {
		err = append(err, errors.New("PackageVersion cannot be empty"))
	}
	if len(err) > 0 {
		return err
	}
	return nil
}

type PackageAvailableUpgrade struct {
	Meta
	PackageName    string
	PackageVersion string
}

// Name returns the name of the rule
func (p PackageAvailableUpgrade) Name() string {
	return fmt.Sprintf("Package Available: %s %s", p.PackageName, p.PackageVersion)
}

// IsRemoteRule returns true if the rule is to be run from outside of the node
func (p PackageAvailableUpgrade) IsRemoteRule() bool { return false }

// Validate the rule
func (p PackageAvailableUpgrade) Validate() []error {
	err := []error{}
	if p.PackageName == "" {
		err = append(err, errors.New("PackageName cannot be empty"))
	}
	if p.PackageVersion == "" {
		err = append(err, errors.New("PackageVersion cannot be empty"))
	}
	if len(err) > 0 {
		return err
	}
	return nil
}
