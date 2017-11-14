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
type ExecutorCreator func(clusterName string) (install.Executor, error)

// ProvisionerCreator creates provisioners that can be used for standing up
// infrastructure for a specific cluster.
type ProvisionerCreator func(store.Cluster) provision.Provisioner

// New returns a cluster controller
func New(l *log.Logger, execCreator ExecutorCreator, provisionerCreator ProvisionerCreator, cs store.ClusterStore, reconFreq time.Duration) ClusterController {
	return &multiClusterController{
		log:                l,
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
//     - generated/
//     - runs/
func DefaultExecutorCreator(rootDir string) ExecutorCreator {
	return func(clusterName string) (install.Executor, error) {
		err := os.MkdirAll(filepath.Join(rootDir, clusterName), 0700)
		if err != nil {
			return nil, fmt.Errorf("error creating directories for executor: %v", err)
		}
		logFile, err := os.Create(filepath.Join(rootDir, clusterName, "kismatic.log"))
		if err != nil {
			return nil, fmt.Errorf("error creating log file for executor: %v", err)
		}
		executorOpts := install.ExecutorOptions{
			GeneratedAssetsDirectory: filepath.Join(rootDir, clusterName, "generated"),
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
		switch cluster.Plan.Provisioner.Provider {
		case "aws":
			p := provision.AWS{
				KeyID:     cluster.AwsID,
				Secret:    cluster.AwsKey,
				Terraform: terraform,
			}
			return p
		default:
			panic(fmt.Sprintf("provider not supported: %q", cluster.Plan.Provisioner.Provider))
		}
	}
}
