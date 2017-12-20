package controller

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/pkg/provision"

	"github.com/apprenda/kismatic/pkg/store"
)

// The size of the buffer assigned to each cluster controller created by the
// multiClusterController.
const clusterControllerNotificationBuffer = 10

// The multiClusterController (mcc) manages a set of cluster controllers
// (workers). Whenever a new cluster is defined in the store, the mcc creates a
// new worker that will be responsible for that cluster's lifecycle.
//
// In the event that the state of a given cluster changes in the store, the mcc
// is notified. The mcc, in turn, notifies the worker that is responsible for
// that cluster.
//
// Given that there is only one communication channel between the store and the
// mcc, the mcc creates buffered channels for each worker so that notifications
// can be dispatched immediately. In the case that the buffer is full, the
// notification is dropped.
//
// When a cluster is deleted from the store, the corresponding worker is
// terminated.
type multiClusterController struct {
	assetsRootDir      string
	log                *log.Logger
	newExecutor        ExecutorCreator
	provisionerCreator func(store.Cluster) provision.Provisioner
	clusterStore       store.ClusterStore
	reconcileFreq      time.Duration
	clusterControllers map[string]chan<- struct{}
}

// Run starts the multiClusterController. The controller will run until the
// passed context is canceled.
func (mcc *multiClusterController) Run(ctx context.Context) {
	mcc.log.Println("started multi-cluster controller")
	watch := mcc.clusterStore.Watch(context.Background(), 0)
	ticker := time.Tick(mcc.reconcileFreq)
	for {
		select {
		case resp := <-watch:
			clusterName := resp.Key
			ch, found := mcc.clusterControllers[clusterName]

			// Stop the cluster controller if the cluster has been deleted
			if found && resp.Value == nil {
				close(ch)
				delete(mcc.clusterControllers, clusterName)
				continue
			}

			// Create a new controller if this is the first time we hear about
			// this cluster
			if !found {
				var cluster store.Cluster
				err := json.Unmarshal(resp.Value, &cluster)
				if err != nil {
					mcc.log.Printf("error unmarshaling watch event value for cluster %q: %v", clusterName, err)
					continue
				}

				newChan := make(chan struct{}, clusterControllerNotificationBuffer)
				ch = newChan
				mcc.clusterControllers[clusterName] = newChan
				executor, err := mcc.newExecutor(clusterName, mcc.assetsRootDir)
				if err != nil {
					mcc.log.Printf("error creating executor for new cluster: %v", err)
					continue
				}
				cc := clusterController{
					clusterName:      clusterName,
					clusterSpec:      cluster.Spec,
					clusterAssetsDir: filepath.Join(mcc.assetsRootDir, clusterName),
					log:              mcc.log,
					executor:         executor,
					clusterStore:     mcc.clusterStore,
					newProvisioner:   mcc.provisionerCreator,
				}
				go cc.run(newChan)
			}

			// Don't block if the cluster controller's buffer is full.
			select {
			case ch <- struct{}{}:
			default:
				mcc.log.Printf("buffer of cluster %s is full. dropping notification.", clusterName)
			}

		case <-ticker:
			mcc.log.Println("tick")
			definedClusters, err := mcc.clusterStore.GetAll()
			if err != nil {
				mcc.log.Printf("failed to get all the clusters defined in the store: %v", err)
				continue
			}
			// Make sure we have workers for all the clusters that are defined in the store
			for clusterName, cluster := range definedClusters {
				_, found := mcc.clusterControllers[clusterName]
				if !found {
					newChan := make(chan struct{}, clusterControllerNotificationBuffer)
					mcc.clusterControllers[clusterName] = newChan
					executor, err := mcc.newExecutor(clusterName, mcc.assetsRootDir)
					if err != nil {
						mcc.log.Printf("error creating executor for new cluster: %v", err)
						continue
					}
					cc := clusterController{
						clusterName:      clusterName,
						clusterSpec:      cluster.Spec,
						clusterAssetsDir: filepath.Join(mcc.assetsRootDir, clusterName),
						log:              mcc.log,
						executor:         executor,
						clusterStore:     mcc.clusterStore,
						newProvisioner:   mcc.provisionerCreator,
					}
					go cc.run(newChan)
				}
			}

			// Remove lingering cluster controllers, if any
			for clusterName, ch := range mcc.clusterControllers {
				_, found := definedClusters[clusterName]
				if !found {
					close(ch)
					delete(mcc.clusterControllers, clusterName)
				}
			}

			// Poke each cluster controller with the latest cluster definition
			for clusterName, ch := range mcc.clusterControllers {
				// Don't block if the cluster controller's buffer is full.
				select {
				case ch <- struct{}{}:
				default:
					mcc.log.Printf("buffer of cluster %s is full. dropping notification.", clusterName)
				}
			}

		case <-ctx.Done():
			mcc.log.Println("stopping the multi-cluster controller")
			for _, v := range mcc.clusterControllers {
				close(v)
			}
			return
		}
	}
}
