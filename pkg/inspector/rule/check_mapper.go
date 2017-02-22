package rule

import (
	"fmt"
	"time"

	"github.com/apprenda/kismatic/pkg/inspector/check"
)

// CheckMapper implements a mapping between a
// rule and a check.
type CheckMapper interface {
	GetCheckForRule(Rule) (check.Check, error)
}

// The DefaultCheckMapper contains the mappings for all
// supported rules and checks.
type DefaultCheckMapper struct {
	PackageManager check.PackageManager
	// IP of the remote node that is being inspected when in client mode
	TargetNodeIP string
}

// GetCheckForRule returns the check for the given rule. If the rule
// is unknown to the mapper, it returns an error.
func (m DefaultCheckMapper) GetCheckForRule(rule Rule) (check.Check, error) {
	var c check.Check
	switch r := rule.(type) {
	default:
		return nil, fmt.Errorf("Rule of type %T is not supported", r)
	case PackageAvailable:
		pkgQuery := check.PackageQuery{Name: r.PackageName, Version: r.PackageVersion}
		c = &check.PackageAvailableCheck{PackageQuery: pkgQuery, PackageManager: m.PackageManager}
	case ExecutableInPath:
		c = &check.ExecutableInPathCheck{Name: r.Executable}
	case FileContentMatches:
		c = check.FileContentCheck{File: r.File, SearchString: r.ContentRegex}
	case TCPPortAvailable:
		c = &check.TCPPortServerCheck{PortNumber: r.Port}
	case TCPPortAccessible:
		timeout, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q provided for the timeout field of the TCPPortAccessible rule: %v", r.Timeout, err)
		}
		c = &check.TCPPortClientCheck{PortNumber: r.Port, IPAddress: m.TargetNodeIP, Timeout: timeout}
	case Python2Version:
		c = &check.Python2Check{SupportedVersions: r.SupportedVersions}
	case FreeSpace:
		bytes, _ := r.MinimumBytesAsUint64()
		c = &check.FreeSpaceCheck{Path: r.Path, MinimumBytes: bytes}
	case PackageAvailableUpgrade:
		pkgQuery := check.PackageQuery{Name: r.PackageName, Version: r.PackageVersion}
		c = &check.PackageAvailableUpgradeCheck{PackageQuery: pkgQuery, PackageManager: m.PackageManager}
	}
	return c, nil
}
