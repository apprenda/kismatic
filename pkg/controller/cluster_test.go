package controller

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/fortytw2/leaktest"
)

type dummyExec struct {
	installSleep time.Duration
}

func (e dummyExec) Install(p *install.Plan) error {
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
	panic("not implemented")
}

func (e dummyExec) RunSmokeTest(*install.Plan) error {
	panic("not implemented")
}

func (e dummyExec) AddWorker(*install.Plan, install.Node) (*install.Plan, error) {
	panic("not implemented")
}

func (e dummyExec) RunPlay(string, *install.Plan) error {
	panic("not implemented")
}

func (e dummyExec) AddVolume(*install.Plan, install.StorageVolume) error {
	panic("not implemented")
}

func (e dummyExec) DeleteVolume(*install.Plan, string) error {
	panic("not implemented")
}

func (e dummyExec) UpgradeNodes(plan install.Plan, nodesToUpgrade []install.ListableNode, onlineUpgrade bool, maxParallelWorkers int) error {
	panic("not implemented")
}

func (e dummyExec) ValidateControlPlane(plan install.Plan) error {
	panic("not implemented")
}

func (e dummyExec) UpgradeClusterServices(plan install.Plan) error {
	panic("not implemented")
}

type dummyStore struct {
	mu    sync.Mutex
	data  map[string][]byte
	watch func() (watchChan, error)
}

func (s *dummyStore) Watch(ctx context.Context, bucket []byte) (watchChan, error) {
	return s.watch()
}

func (s *dummyStore) Put(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[string(key)] = value
	return nil
}

func (s *dummyStore) Get(key []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[string(key)], nil
}

func TestClusterController(t *testing.T) {
	defer leaktest.Check(t)()

	clusterName := "prodCluster"
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(os.Stdout, "[cluster controller] ", log.Ldate|log.Ltime)

	// Stub out dependencies
	pw := planWrapper{CurrentState: planned, DesiredState: installed}
	pwBytes, err := json.Marshal(pw)
	if err != nil {
		t.Fatalf("error marshaling plan: %v", err)
	}

	watchFunc := func() (watchChan, error) {
		c := make(chan watchResponse)
		go func(ctx context.Context) {
			// trigger a watch, and then wait until the ctx is done
			c <- watchResponse{key: []byte(clusterName), value: pwBytes}
			<-ctx.Done()
			return
		}(ctx)
		return c, nil
	}
	store := &dummyStore{
		watch: watchFunc,
		data:  map[string][]byte{clusterName: pwBytes},
	}
	executor := dummyExec{installSleep: 1 * time.Second}

	// Start the controller
	c := New(logger, executor, store)
	go func(ctx context.Context) {
		err := c.Run(ctx)
		if err != nil {
			t.Errorf("error running controller")
			cancel()
		}
	}(ctx)

	// Assert that the cluster reaches desired state
	var done bool
	for !done {
		select {
		case <-time.Tick(time.Second):
			var pw planWrapper
			b, _ := store.Get([]byte(clusterName))
			json.Unmarshal(b, &pw)
			if pw.CurrentState == pw.DesiredState {
				cancel()
				done = true
				break
			}
		case <-time.After(10 * time.Second):
			cancel()
			t.Errorf("did not reach installed state")
			done = true
			break
		}
	}
}
