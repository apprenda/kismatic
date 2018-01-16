package check

import (
	"fmt"
	"strings"
)

// PackageQuery is a query for finding a package
type PackageQuery struct {
	Name    string
	Version string
}

func (p PackageQuery) String() string {
	return fmt.Sprintf("%s %s", p.Name, p.Version)
}

// The PackageCheck uses the operating system to determine whether a
// package is installed.
type PackageCheck struct {
	PackageQuery               PackageQuery
	PackageManager             PackageManager
	InstallationDisabled       bool
	DockerInstallationDisabled bool
}

// Check returns true if the package is installed. If pkg installation is disabled,
// we would like to check if the package is available for install. However,
// there is no guarantee that the node will have the kismatic package repo configured.
// For this reason, this check is a no-op when package installation is disabled.
func (c PackageCheck) Check() (bool, error) {
	if !c.InstallationDisabled {
		return true, nil
	}
	// When docker installation is disabled do not check for any packages that contain "docker" in the name
	// The package name could be different, we will only validate the docker executable is present
	if c.DockerInstallationDisabled && strings.Contains(c.PackageQuery.Name, "docker") {
		return true, nil
	}
	installed, err := c.PackageManager.IsInstalled(c.PackageQuery)
	if err != nil {
		return false, fmt.Errorf("failed to determine if package is installed: %v", err)
	}
	if installed {
		return true, nil
	}
	// We check to see if it's available to give useful feedback to the user
	available, err := c.PackageManager.IsAvailable(c.PackageQuery)
	if err != nil {
		return false, fmt.Errorf("failed to determine if package is available for install: %v", err)
	}
	if !available {
		return false, fmt.Errorf("package is not installed, and is not available in known package repositories")
	}
	return false, fmt.Errorf("package is not installed, but is available in a package repository")
}
