package check

import "fmt"

type PackageNotInstalledCheck struct {
	PackageQuery             PackageQuery
	AcceptablePackageVersion string
	PackageManager           PackageManager
	InstallationDisabled     bool
}

// Check returns true if the specified package is not installed.
// This will also return true if the version installed matches AcceptablePackageVersion.
// When InstallationDisabled is true this check will always return true.
func (c PackageNotInstalledCheck) Check() (bool, error) {
	// don't check when installation is disabled
	if c.InstallationDisabled {
		return true, nil
	}
	// check for the package with optional version is installed
	installed, err := c.PackageManager.IsInstalled(c.PackageQuery)
	if err != nil {
		return false, fmt.Errorf("failed to determine if package is installed: %v", err)
	}
	// return true if nothing is installed
	if !installed {
		return true, nil
	}
	if c.AcceptablePackageVersion == "" {
		return false, fmt.Errorf("package should not be installed")
	}
	// check if the version installed is the acceptable version
	acceptableVersionInstalled, err := c.PackageManager.IsInstalled(PackageQuery{Name: c.PackageQuery.Name, Version: c.AcceptablePackageVersion})
	if err != nil {
		return false, fmt.Errorf("failed to determine if package is installed: %v", err)
	}
	if acceptableVersionInstalled {
		return true, nil
	}
	return false, nil
}
