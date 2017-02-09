package install

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type KismaticInfo struct {
	ShortVersion string
	BuildNumber  string
}

type Version struct {
	Major int
	Minor int
	Patch int
}

type ClusterVersion struct {
	EarliestVersion Version
	LatestVersion   Version
	IsTransitioning bool
	Nodes           []ListableNode
}

type ListableNode struct {
	IP      string
	Roles   []string
	Version Version
}

var AboutKismatic KismaticInfo

// Takes a version in form "v{short version}-{changeset number}-{changeset ID}[-{dirtyflag}]"
func SetVersion(polyVersion string) {
	re := regexp.MustCompile("v([^-]+)-([^-]+)-([^-]+)-?([^-]+)")
	matches := re.FindStringSubmatch(polyVersion)
	if len(matches) > 2 {
		AboutKismatic = KismaticInfo{matches[1], matches[2]}
	} else {
		fmt.Printf("Could not parse %v", polyVersion)
	}

}

// Takes a version in form "major.minor.patch[-build]"
// If version fails to convert or is otherwise older than the current version, returns true
func IsOlderVersion(comparedVersion string) bool {
	re := regexp.MustCompile("([^-]+)-([^-]+)")
	matches := re.FindStringSubmatch(comparedVersion)

	if len(matches) > 1 {
		this := parseVersion(AboutKismatic.ShortVersion)
		that := parseVersion(matches[1])
		return this.isNewerThan(that)
	}

	return true
}

// Returns true if that is older than this; this > that
func (this Version) isNewerThan(that Version) bool {
	if this.Major > that.Major ||
		this.Minor > that.Minor ||
		this.Patch > that.Patch {
		return true
	}
	return false
}

func parseVersion(versionString string) Version {
	newVer := Version{}
	versions := strings.Split(versionString, ".")
	versionLevel := len(versions)
	if versionLevel > 0 {
		newVer.Major, _ = strconv.Atoi(versions[0])
		if versionLevel > 1 {
			newVer.Minor, _ = strconv.Atoi(versions[1])
			if versionLevel > 2 {
				newVer.Patch, _ = strconv.Atoi(versions[2])
			}
		}
	}

	return newVer
}

func ListVersions(plan *Plan) (ClusterVersion, error) {
	ips := plan.GetUniqueNodeIPs()

	cv := ClusterVersion{
		Nodes: []ListableNode{},
	}

	for i, ip := range ips {
		sshDeets := plan.Cluster.SSH
		client, err := ssh.NewClient(ip, sshDeets.Port, sshDeets.User, sshDeets.Key)
		if err != nil {
			return cv, fmt.Errorf("error creating SSH client: %v", err)
		}

		o, err := client.Output(false, "cat /etc/kismatic-version")
		var thisVersion Version
		if err != nil {
			thisVersion = Version{Major: 1}
		} else {
			thisVersion = parseVersion(o)
		}

		if i == 0 {
			cv.EarliestVersion = thisVersion
			cv.LatestVersion = thisVersion
		} else {
			if thisVersion.isNewerThan(cv.LatestVersion) {
				cv.LatestVersion = thisVersion
			}
			if cv.EarliestVersion.isNewerThan(thisVersion) {
				cv.EarliestVersion = thisVersion
			}
		}

		cv.Nodes = append(cv.Nodes, ListableNode{ip, plan.GetRolesForIP(ip), thisVersion})
	}

	cv.IsTransitioning = cv.EarliestVersion != cv.LatestVersion

	return cv, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}
