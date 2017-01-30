package check

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// PackageManager runs queries against the underlying operating system's
// package manager
type PackageManager interface {
	IsAvailable(PackageQuery) (bool, error)
	IsInstalled(PackageQuery) (bool, error)
	Enforced() bool
}

func IsPackageReadyToContinue(m PackageManager, q PackageQuery) (bool, error) {
	if m.Enforced() {
		installed, _ := m.IsInstalled(q)

		if installed {
			return true, nil
		}

		available, _ := m.IsAvailable(q)

		if available {
			return false, fmt.Errorf("%v is not installed but is available in a package repository", q)
		}

		return false, fmt.Errorf("%v is not installed and isn't available in known package repositories", q)
	}
	return true, nil
}

func IsPackageAvailable(m PackageManager, q PackageQuery) (bool, error) {
	installed, _ := m.IsInstalled(q)

	if installed {
		return true, nil
	}

	available, _ := m.IsAvailable(q)

	if available {
		return true, nil
	}

	return false, fmt.Errorf("%v is not installed and isn't available in known package repositories", q)
}

// NewPackageManager returns a package manager for the given distribution
func NewPackageManager(distro Distro, enforcePackages bool) (PackageManager, error) {
	run := func(name string, arg ...string) ([]byte, error) {
		r, err := exec.Command(name, arg...).CombinedOutput()
		return r, err
	}
	switch distro {
	case RHEL, CentOS:
		return &rpmManager{
			run:             run,
			enforcePackages: enforcePackages,
		}, nil
	case Ubuntu:
		return &debManager{
			run:             run,
			enforcePackages: enforcePackages,
		}, nil
	case Darwin:
		return noopManager{}, nil
	default:
		return nil, fmt.Errorf("%s is not supported", distro)
	}
}

type noopManager struct{}

func (noopManager) IsAvailable(PackageQuery) (bool, error) {
	return false, fmt.Errorf("unable to determine if package is available using noop pkg manager")
}
func (noopManager) IsInstalled(PackageQuery) (bool, error) {
	return false, fmt.Errorf("unable to determine if package is installed using noop pkg manager")
}
func (noopManager) Enforced() bool {
	return false
}

// package manager for EL-based distributions
type rpmManager struct {
	run             func(string, ...string) ([]byte, error)
	enforcePackages bool
}

func (m rpmManager) Enforced() bool {
	return m.enforcePackages
}

func (m rpmManager) IsAvailable(p PackageQuery) (bool, error) {
	out, err := m.run("yum", "list", "available", "-q", p.Name)
	if err != nil && strings.Contains(string(out), "No matching Packages to list") {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("unable to determine if %s %s is available: %v", p.Name, p.Version, err)
	}
	return m.isPackageListed(p, out), nil
}

func (m rpmManager) IsInstalled(p PackageQuery) (bool, error) {
	out, err := m.run("yum", "list", "installed", "-q", p.Name)
	if err != nil && strings.Contains(string(out), "No matching Packages to list") {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("unable to determine if %s %s is installed: %v", p.Name, p.Version, err)
	}
	return m.isPackageListed(p, out), nil
}

func (m rpmManager) isPackageListed(p PackageQuery, list []byte) bool {
	s := bufio.NewScanner(bytes.NewReader(list))

	for s.Scan() {
		line := s.Text()
		f := strings.Fields(line)
		if len(f) != 3 {
			// Ignore lines that don't match the expected format
			continue
		}
		maybeName := strings.Split(f[0], ".")[0]
		maybeVersion := f[1]
		if p.Name == maybeName && p.Version == maybeVersion {
			return true
		}
	}
	return false
}

// package manager for debian-based distributions
type debManager struct {
	run             func(string, ...string) ([]byte, error)
	enforcePackages bool
}

func (m debManager) Enforced() bool {
	return m.enforcePackages
}

func (m debManager) IsInstalled(p PackageQuery) (bool, error) {
	// First check if the package is installed
	installed, err := m.isPackageListed(p)
	if err != nil {
		return false, err
	}
	return installed, nil
}

func (m debManager) IsAvailable(p PackageQuery) (bool, error) {
	// If it's not installed, ensure that it is available via the
	// package manager. We attempt to install using --dry-run. If exit status is zero, we
	// know the package is available for download
	out, err := m.run("apt-get", "install", "-q", "--dry-run", fmt.Sprintf("%s=%s", p.Name, p.Version))
	if err != nil && strings.Contains(string(out), "Unable to locate package") {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m debManager) isPackageListed(p PackageQuery) (bool, error) {
	out, err := m.run("dpkg", "-l", p.Name)
	if err != nil && strings.Contains(string(out), "no packages found matching") {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("unable to determine if %s %s is installed: %v", p.Name, p.Version, err)
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		line := s.Text()
		f := strings.Fields(line)
		if len(f) < 5 {
			// Ignore lines with unexpected format
			continue
		}
		maybeName := strings.Split(f[1], ".")[0]
		maybeVersion := f[2]
		if p.Name == maybeName && p.Version == maybeVersion {
			return true, nil
		}
	}
	return false, nil
}
