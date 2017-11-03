package controller

import (
	"context"
	"log"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/store"
)

// The ClusterController manages the lifecycle of clusters
type ClusterController interface {
	Run(ctx context.Context)
}

// New returns a cluster controller
func New(l *log.Logger, e install.Executor, s store.WatchedStore, genAssetsDir string, reconFreq time.Duration) ClusterController {
	return &multiClusterController{
		log:                l,
		executor:           e,
		clusterStore:       cs{bucket: "clusters", store: s},
		reconcileFreq:      reconFreq,
		generatedAssetsDir: genAssetsDir,
		clusterControllers: make(map[string]chan<- planWrapper),
	}
}
