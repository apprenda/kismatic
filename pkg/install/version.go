package install

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver"
)

var (
	httpTimeout                  = 5 * time.Second
	kubeReleaseRegex             = regexp.MustCompile(`^v1\.9\.(0|[1-9][0-9]*)$`)
	kubernetesReleaseURL         = "https://storage.googleapis.com/kubernetes-release/release/stable-1.9.txt"
	kubernetesVersionString      = "v1.9.3"
	kubernetesMinorVersionString = "v1.9.x"
	kubernetesVersion            = semver.Version{Major: 1, Minor: 9, Patch: 3} // build the struct directly to not get an error
)

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

// kubernetesStableVersion fetches the latest stable version
// if an error occurs it will return the tested version
func kubernetesLatestStableVersion() (semver.Version, string, error) {
	timeout := time.Duration(httpTimeout)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(kubernetesReleaseURL)
	if err != nil {
		return kubernetesVersion, kubernetesVersionString, fmt.Errorf("Error getting latest version from %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return kubernetesVersion, kubernetesVersionString, fmt.Errorf("Bad status code getting latest version")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return kubernetesVersion, kubernetesVersionString, fmt.Errorf("Error reading response %v", err)
	}
	latest := strings.Trim(string(body), " \t\n")
	if !kubeReleaseRegex.MatchString(latest) {
		return kubernetesVersion, kubernetesVersionString, fmt.Errorf("Invalid version format %q", latest)
	}
	parsedLatest, err := parseVersion(latest)
	if err != nil {
		return kubernetesVersion, kubernetesVersionString, fmt.Errorf("Could not parse version string %q", latest)
	}
	return parsedLatest, latest, nil
}

// validates that the version is the expected Minor version
func kubernetesVersionValid(version string) bool {
	return kubeReleaseRegex.MatchString(version)
}

// VersionOverrides returns a map of all image names and their versions that can be modified by the user
func VersionOverrides() map[string]string {
	versions := make(map[string]string, 0)
	versions["kube_proxy"] = kubernetesVersionString
	versions["kube_controller_manager"] = kubernetesVersionString
	versions["kube_scheduler"] = kubernetesVersionString
	versions["kube_apiserver"] = kubernetesVersionString

	return versions
}
