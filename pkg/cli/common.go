package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultPlanName      = "kismatic-cluster.yaml"
	defaultClusterName   = "kubernetes"
	assetsFolder         = "clusters"
	defaultGeneratedName = "generated"
	defaultRunsName      = "runs"

	playbooksPath  = "playbooks"
	ansiblePath    = "ansible"
	clustersBucket = "kismatic"
)

type planFileNotFoundErr struct {
	filename string
}

func (e planFileNotFoundErr) Error() string {
	return fmt.Sprintf("Plan file not found at %q. If you don't have a plan file, you may generate one with 'kismatic install plan'", e.filename)
}

// Returns a path to a plan file, generated dir, and runs dir according to the clusterName
func generateDirsFromName(clusterName string) (string, string, string) {
	return filepath.Join(assetsFolder, clusterName, defaultPlanName), filepath.Join(assetsFolder, clusterName, defaultGeneratedName), filepath.Join(assetsFolder, clusterName, defaultRunsName)
}

// CheckClusterExists does a simple check to see if the cluster folder+plan file exists in clusters
// returns true even in the cases where an error exists if the scan hasn't completed.
func CheckClusterExists(name string) (bool, error) {
	if err := os.MkdirAll(assetsFolder, 0700); err != nil {
		return true, err
	}
	// TODO: also check db
	// MOVED TO "seamless" PR - requires store interface changes
	files, err := ioutil.ReadDir(assetsFolder)
	if err != nil {
		return true, err
	}
	for _, finfo := range files {
		if finfo.Name() == name {
			possiblePlans, err := ioutil.ReadDir(filepath.Join(assetsFolder, finfo.Name()))
			if err != nil {
				return true, err
			}
			for _, possiblePlan := range possiblePlans {
				if possiblePlan.Name() == defaultPlanName {
					return true, nil
				}
			}
		}
	}
	return false, fmt.Errorf("Cluster with name %s not found. If you have a plan file, but your cluster doesn't exist, please run kismatic import PLAN_FILE_PATH.", name)
}
