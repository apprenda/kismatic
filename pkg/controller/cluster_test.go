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

type dummyExec struct{}

func (e dummyExec) Install(p *install.Plan, restartServices bool) error {
	return nil
}

func (e dummyExec) RunPreFlightCheck(*install.Plan) error {
	return nil
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

	logger := log.New(os.Stdout, "[cluster controller] ", log.Ldate|log.Ltime)

	// Stub out dependencies
	executorCreator := func(string) (install.Executor, error) { return dummyExec{}, nil }

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clusterName := "testCluster"
	c := New(logger, executorCreator, provisioner, clusterStore, 10*time.Minute)
	go c.Run(ctx)

	// Create a new cluster in the store
	// We don't have a way to reliably wait until the controller is watching,
	// so we have to issue multiple writes
	writerDone := make(chan struct{})
	defer func() { close(writerDone) }()
	go func(done <-chan struct{}) {
		cluster := store.Cluster{CurrentState: planned, DesiredState: installed, CanContinue: true}
		err = clusterStore.Put(clusterName, cluster)
		if err != nil {
			t.Fatalf("error storing cluster")
		}
		tick := time.Tick(3 * time.Second)
		for {
			select {
			case <-tick:
				c, err := clusterStore.Get(clusterName)
				if err != nil {
					t.Fatalf("error getting cluster")
				}
				err = clusterStore.Put(clusterName, *c)
				if err != nil {
					t.Fatalf("error storing cluster")
				}
			case <-done:
				return
			}
		}
	}(writerDone)

	// Assert that the cluster reaches desired state
	tick := time.Tick(100 * time.Millisecond)
done:
	for {
		select {
		case <-tick:
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
				break done
			}
		case <-time.After(5 * time.Second):
			fmt.Println("tick")
			t.Fatalf("did not reach installed state")
		}
	}
}

func TestClusterControllerReconciliationLoop(t *testing.T) {
	// TODO: the store is leaking a goroutine, so can't enable this
	// defer leaktest.Check(t)()
	logger := log.New(os.Stdout, "[cluster controller] ", log.Ldate|log.Ltime)

	// Stub out dependencies
	executorCreator := func(string) (install.Executor, error) { return dummyExec{}, nil }

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := New(logger, executorCreator, provisioner, clusterStore, 1*time.Second)
	go c.Run(ctx)

	// Assert that the cluster reaches desired state
done:
	for {
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
				break done
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("did not reach installed state")
		}
	}
}
