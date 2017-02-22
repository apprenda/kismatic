package install

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type KismaticInfo struct {
	ShortVersion Version
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
	Node    Node
	Roles   []string
	Version Version
}

var AboutKismatic KismaticInfo

// Takes a version in form "v{short version}-{changeset number}-{changeset ID}[-{dirtyflag}]"
func SetVersion(polyVersion string) {
	re := regexp.MustCompile("v([^-]+)-([^-]+)-([^-]+)-?([^-]+)")
	matches := re.FindStringSubmatch(polyVersion)
	if len(matches) > 2 {
		ver := parseVersion(matches[1])
		AboutKismatic = KismaticInfo{ver, matches[2]}
	} else {
		fmt.Printf("Could not parse %v", polyVersion)
	}

}

// Returns true if the provided version is older than the current Kismatic version
func IsOlderVersion(that Version) bool {
	this := AboutKismatic.ShortVersion
	return this.isNewerThan(that)
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
	nodes := plan.GetUniqueNodes()

	cv := ClusterVersion{
		Nodes: []ListableNode{},
	}

	for i, node := range nodes {
		sshDeets := plan.Cluster.SSH
		client, err := ssh.NewClient(node.IP, sshDeets.Port, sshDeets.User, sshDeets.Key)
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

		cv.Nodes = append(cv.Nodes, ListableNode{node, plan.GetRolesForIP(node.IP), thisVersion})
	}

	cv.IsTransitioning = cv.EarliestVersion != cv.LatestVersion

	return cv, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}
