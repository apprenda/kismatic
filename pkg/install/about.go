package install

import (
	"fmt"

	"github.com/apprenda/kismatic/pkg/ssh"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/blang/semver"
)

// ClusterVersion contains version information about the cluster
type ClusterVersion struct {
	EarliestVersion semver.Version
	LatestVersion   semver.Version
	IsTransitioning bool
	Nodes           []ListableNode
}

// ListableNode contains version and role information about a given node
type ListableNode struct {
	Node    Node
	Roles   []string
	Version semver.Version
}

// KismaticVersion contains the version information of the currently running binary
var KismaticVersion semver.Version

// SetVersion parses the given version, and sets it as the global version of the binary
func SetVersion(v string) {
	ver, err := parseVersion(v)
	if err != nil {
		panic("failed to parse version " + v)
	}
	KismaticVersion = ver
}

// IsOlderVersion returns true if the provided version is older than the current Kismatic version
func IsOlderVersion(that semver.Version) bool {
	this := KismaticVersion
	return this.GT(that)
}

// IsLessThanVersion parses the version from a string and returns true if this version is less than that version
func IsLessThanVersion(this semver.Version, that string) bool {
	thatVersion, err := parseVersion(that)
	if err != nil {
		panic("failed to parse version " + that)
	}

	return this.LT(thatVersion)
}

func parseVersion(versionString string) (semver.Version, error) {
	// Support a 'v' prefix
	verString := versionString
	if versionString[0] == 'v' {
		verString = versionString[1:len(versionString)]
	}
	v, err := semver.Make(verString)
	if err != nil {
		return semver.Version{}, fmt.Errorf("Unable to parse version %q: %v", verString, err)
	}
	return v, nil
}

// ListVersions connects to the cluster described in the plan file and
// gathers version information about it.
func ListVersions(plan *Plan) (ClusterVersion, error) {
	nodes := plan.GetUniqueNodes()
	cv := ClusterVersion{
		Nodes: []ListableNode{},
	}

	sshDeets := plan.Cluster.SSH
	verFile := "/etc/kismatic-version"
	for i, node := range nodes {
		client, err := ssh.NewClient(node.IP, sshDeets.Port, sshDeets.User, sshDeets.Key)
		if err != nil {
			return cv, fmt.Errorf("error creating SSH client: %v", err)
		}

		output, err := client.Output(false, fmt.Sprintf("cat %s", verFile))
		if err != nil {
			// the output var contains the actual error message from the cat command, which has
			// more meaningful info
			return cv, fmt.Errorf("error getting version for node %q: %q", node.Host, output)
		}

		thisVersion, err := parseVersion(output)
		if err != nil {
			return cv, fmt.Errorf("invalid version %q found in version file %q of node %s", output, verFile, node.Host)
		}

		cv.Nodes = append(cv.Nodes, ListableNode{node, plan.GetRolesForIP(node.IP), thisVersion})

		// If looking at the first node, set the versions and move on
		if i == 0 {
			cv.EarliestVersion = thisVersion
			cv.LatestVersion = thisVersion
			continue
		}

		if thisVersion.GT(cv.LatestVersion) {
			cv.LatestVersion = thisVersion
		}
		if cv.EarliestVersion.GT(thisVersion) {
			cv.EarliestVersion = thisVersion
		}
	}

	cv.IsTransitioning = cv.EarliestVersion.NE(cv.LatestVersion)
	return cv, nil
}

// NodesWithRoles returns a filtered list of ListableNode slice based on the node's roles
func NodesWithRoles(nodes []ListableNode, roles ...string) []ListableNode {
	var subset []ListableNode
	for _, need := range roles {
		for _, n := range nodes {
			if util.Subset([]string{need}, n.Roles) {
				subset = append(subset, n)
			}
		}
	}
	return subset
}
