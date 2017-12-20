package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/provision"
	"github.com/apprenda/kismatic/pkg/store"
)

// The ClusterController manages the lifecycle of clusters
type ClusterController interface {
	Run(ctx context.Context)
}

// ExecutorCreator creates executors that can be used for executing actions
// against a specific cluster.
type ExecutorCreator func(clusterName, assetsRootDir string) (install.Executor, error)

// ProvisionerCreator creates provisioners that can be used for standing up
// infrastructure for a specific cluster.
type ProvisionerCreator func(store.Cluster) provision.Provisioner

// New returns a cluster controller
func New(
	logger *log.Logger,
	execCreator ExecutorCreator,
	provisionerCreator ProvisionerCreator,
	cs store.ClusterStore,
	reconFreq time.Duration,
	assetsRootDir string) ClusterController {
	return &multiClusterController{
		assetsRootDir:      assetsRootDir,
		log:                logger,
		newExecutor:        execCreator,
		clusterStore:       cs,
		reconcileFreq:      reconFreq,
		clusterControllers: make(map[string]chan<- struct{}),
		provisionerCreator: provisionerCreator,
	}
}

// DefaultExecutorCreator creates an executor that can be used to run operations
// against a single cluster. The given rootDir is used as the root directory
// under which new directories are created for each executor.
//
// The following directory structure is created under the rootDir for each
// cluster executor:
// - clusterName/
//     - kismatic.log
//     - assets/
//     - runs/
func DefaultExecutorCreator() ExecutorCreator {
	return func(clusterName string, rootDir string) (install.Executor, error) {
		err := os.MkdirAll(filepath.Join(rootDir, clusterName), 0700)
		if err != nil {
			return nil, fmt.Errorf("error creating directories for executor: %v", err)
		}
		logFile, err := os.Create(filepath.Join(rootDir, clusterName, "kismatic.log"))
		if err != nil {
			return nil, fmt.Errorf("error creating log file for executor: %v", err)
		}
		executorOpts := install.ExecutorOptions{
			GeneratedAssetsDirectory: filepath.Join(rootDir, clusterName, "assets"),
			RunsDirectory:            filepath.Join(rootDir, clusterName, "runs"),
			OutputFormat:             "simple",
			Verbose:                  true,
		}
		executor, err := install.NewExecutor(logFile, logFile, executorOpts)
		if err != nil {
			return nil, err
		}
		return executor, nil
	}
}

// DefaultProvisionerCreator uses terraform for provisioning infrastructure
// on the clouds we support.
func DefaultProvisionerCreator(terraform provision.Terraform) ProvisionerCreator {
	return func(cluster store.Cluster) provision.Provisioner {
		switch cluster.Spec.Provisioner.Provider {
		case "aws":
			p := provision.AWS{
				AccessKeyID:     cluster.Spec.Provisioner.Credentials.AWS.AccessKeyId,
				SecretAccessKey: cluster.Spec.Provisioner.Credentials.AWS.SecretAccessKey,
				Terraform:       terraform,
			}
			return p
		default:
			panic(fmt.Sprintf("provider not supported: %q", cluster.Spec.Provisioner.Provider))
		}
	}
}
