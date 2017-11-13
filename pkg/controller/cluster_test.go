package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/apprenda/kismatic/pkg/provision"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/store"
)

type dummyExec struct {
	installSleep time.Duration
}

func (e dummyExec) Install(p *install.Plan, restartServices bool) error {
	time.Sleep(e.installSleep)
	return nil
}

func (e dummyExec) RunPreFlightCheck(*install.Plan) error {
	panic("not implemented")
}

func (e dummyExec) RunNewWorkerPreFlightCheck(install.Plan, install.Node) error {
	panic("not implemented")
}

func (e dummyExec) RunUpgradePreFlightCheck(*install.Plan, install.ListableNode) error {
	panic("not implemented")
}

func (e dummyExec) GenerateCertificates(p *install.Plan, useExistingCA bool) error {
	return nil
}

func (e dummyExec) GenerateKubeconfig(plan install.Plan) error {
	return nil
}

func (e dummyExec) RunSmokeTest(*install.Plan) error {
	return nil
}

func (e dummyExec) AddWorker(*install.Plan, install.Node, bool) (*install.Plan, error) {
	panic("not implemented")
}

func (e dummyExec) RunPlay(string, *install.Plan, bool) error {
	panic("not implemented")
}

func (e dummyExec) AddVolume(*install.Plan, install.StorageVolume) error {
	panic("not implemented")
}

func (e dummyExec) DeleteVolume(*install.Plan, string) error {
	panic("not implemented")
}

func (e dummyExec) UpgradeNodes(plan install.Plan, nodesToUpgrade []install.ListableNode, onlineUpgrade bool, maxParallelWorkers int, restartServices bool) error {
	panic("not implemented")
}

func (e dummyExec) ValidateControlPlane(plan install.Plan) error {
	panic("not implemented")
}

func (e dummyExec) UpgradeClusterServices(plan install.Plan) error {
	panic("not implemented")
}

type dummyProvisioner struct{}

func (p dummyProvisioner) Provision(plan install.Plan) (*install.Plan, error) {
	return &plan, nil
}

func (p dummyProvisioner) Destroy(string) error {
	return nil
}

func TestClusterControllerTriggeredByWatch(t *testing.T) {
	// TODO: the store is leaking a goroutine, so can't enable this
	// defer leaktest.Check(t)()

	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(os.Stdout, "[cluster controller] ", log.Ldate|log.Ltime)

	// Stub out dependencies
	executorCreator := func(string) (install.Executor, error) { return dummyExec{installSleep: 1 * time.Second}, nil }

	tmpFile, err := ioutil.TempFile("", "cluster-controller-tests")
	if err != nil {
		t.Fatalf("error creating temp dir for store")
	}
	s, err := store.New(tmpFile.Name(), 0600, logger)
	defer s.Close()
	bucketName := "clusters"
	if err != nil {
		t.Fatalf("error creating store")
	}
	s.CreateBucket(bucketName)

	clusterStore := store.NewClusterStore(s, bucketName)

	provisioner := func(store.Cluster) provision.Provisioner {
		return dummyProvisioner{}
	}

	// Start the controller
	clusterName := "testCluster"
	c := New(logger, executorCreator, provisioner, clusterStore, 10*time.Minute)
	go c.Run(ctx)

	// Create a new cluster in the store
	cluster := store.Cluster{CurrentState: planned, DesiredState: installed, CanContinue: true}
	err = clusterStore.Put(clusterName, cluster)
	if err != nil {
		t.Fatalf("error storing cluster")
	}

	// Assert that the cluster reaches desired state
	var done bool
	for !done {
		select {
		case <-time.Tick(time.Second):
			var cluster store.Cluster
			b, err := s.Get(bucketName, clusterName)
			if err != nil {
				t.Fatalf("got an error trying to read the cluster from the store")
			}
			err = json.Unmarshal(b, &cluster)
			if err != nil {
				t.Fatalf("error unmarshaling from store")
			}
			if cluster.CurrentState == cluster.DesiredState {
				cancel()
				done = true
				break
			}
		case <-time.After(5 * time.Second):
			fmt.Println("tick")
			cancel()
			t.Errorf("did not reach installed state")
			done = true
			break
		}
	}
}

func TestClusterControllerReconciliationLoop(t *testing.T) {
	// TODO: the store is leaking a goroutine, so can't enable this
	// defer leaktest.Check(t)()
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(os.Stdout, "[cluster controller] ", log.Ldate|log.Ltime)

	// Stub out dependencies
	executorCreator := func(string) (install.Executor, error) { return dummyExec{installSleep: 1 * time.Second}, nil }

	tmpFile, err := ioutil.TempFile("", "cluster-controller-tests")
	if err != nil {
		t.Fatalf("error creating temp dir for store")
	}
	s, err := store.New(tmpFile.Name(), 0600, logger)
	defer s.Close()
	bucketName := "clusters"
	if err != nil {
		t.Fatalf("error creating store")
	}
	s.CreateBucket(bucketName)

	clusterStore := store.NewClusterStore(s, bucketName)

	// Create a new cluster in the store before starting the controller.
	// The controller should pick it up in the reconciliation loop.
	clusterName := "testCluster"
	cluster := store.Cluster{CurrentState: planned, DesiredState: installed, CanContinue: true}
	err = clusterStore.Put(clusterName, cluster)
	if err != nil {
		t.Fatalf("error storing cluster")
	}

	provisioner := func(store.Cluster) provision.Provisioner {
		return dummyProvisioner{}
	}

	// Start the controller
	c := New(logger, executorCreator, provisioner, clusterStore, 3*time.Second)
	go c.Run(ctx)

	// Assert that the cluster reaches desired state
	var done bool
	for !done {
		select {
		case <-time.Tick(time.Second):
			var cluster store.Cluster
			b, err := s.Get(bucketName, clusterName)
			if err != nil {
				t.Fatalf("got an error trying to read the cluster from the store")
			}
			err = json.Unmarshal(b, &cluster)
			if err != nil {
				t.Fatalf("error unmarshaling from store")
			}
			if cluster.CurrentState == cluster.DesiredState {
				cancel()
				done = true
				break
			}
		case <-time.After(5 * time.Second):
			fmt.Println("tick")
			cancel()
			t.Errorf("did not reach installed state")
			done = true
			break
		}
	}
}
