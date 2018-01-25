package install

import (
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/blang/semver"

	yaml "gopkg.in/yaml.v2"
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
	Node              Node
	Roles             []string
	Version           semver.Version
	ComponentVersions ComponentVersions
}

type ComponentVersions struct {
	Kubernetes string
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

// ListVersions connects to the cluster described in the plan file and
// gathers version information about it.
func ListVersions(plan *Plan) (ClusterVersion, error) {
	nodes := plan.GetUniqueNodes()
	cv := ClusterVersion{
		Nodes: []ListableNode{},
	}

	sshDeets := plan.Cluster.SSH
	ketVerFile := "/etc/kismatic-version"
	componentVerFile := "/etc/component-versions"
	for i, node := range nodes {
		client, err := ssh.NewClient(node.IP, sshDeets.Port, sshDeets.User, sshDeets.Key)
		if err != nil {
			return cv, fmt.Errorf("error creating SSH client: %v", err)
		}

		// get KET version
		ketOutput, err := client.Output(false, fmt.Sprintf("cat %s", ketVerFile))
		if err != nil {
			// the output var contains the actual error message from the cat command, which has
			// more meaningful info
			return cv, fmt.Errorf("error getting KET version for node %q: %q", node.Host, ketOutput)
		}

		thisVersion, err := parseVersion(ketOutput)
		if err != nil {
			return cv, fmt.Errorf("invalid version %q found in version file %q of node %s", ketOutput, ketVerFile, node.Host)
		}

		// get component versions
		versionsOutput, err := client.Output(false, fmt.Sprintf("cat %s", componentVerFile))
		// don't fail if the file is not found, will default to empty
		// TODO remove
		if err != nil && !strings.Contains(versionsOutput, "No such file or directory") {
			// the output var contains the actual error message from the cat command, which has
			// more meaningful info
			return cv, fmt.Errorf("error getting component versions for node %q: %q", node.Host, versionsOutput)
		}
		versions := ComponentVersions{}
		if !strings.Contains(versionsOutput, "No such file or directory") {
			err = yaml.Unmarshal([]byte(versionsOutput), &versions)
			if err != nil {
				return cv, fmt.Errorf("error unmarshalling component versions file: %q", componentVerFile)
			}
		}

		cv.Nodes = append(cv.Nodes, ListableNode{node, plan.GetRolesForIP(node.IP), thisVersion, versions})

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
