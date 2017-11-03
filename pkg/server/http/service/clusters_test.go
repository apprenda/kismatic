package service

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/apprenda/kismatic/pkg/server/http/model"
	"github.com/apprenda/kismatic/pkg/store"
)

const (
	bucket = "cluster_test"
)

func setupStore() (store.WatchedStore, error) {
	f, err := ioutil.TempFile("/tmp", "httptests")
	if err != nil {
		return nil, err
	}
	s, err := store.NewBoltDB(f.Name(), 0644, log.New(ioutil.Discard, "test ", 0))
	if err != nil {
		return nil, err
	}
	s.CreateBucket(bucket)

	return s, nil
}

func TestCreateAndGet(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}
	cs := NewClustersService(s, bucket)

	c := &model.ClusterRequest{
		Name:         "foo",
		DesiredState: "running",
		AwsID:        "",
		AwsKey:       "",
		Etcd:         3,
		Master:       2,
		Worker:       5,
	}
	err = cs.Create(c)
	if err != nil {
		t.Fatalf("unexpected error creating cluster: %v", err)
	}
	_, err = cs.Get("bar")
	if err == nil {
		t.Errorf("expected an error")
	}
	if err != ErrClusterNotFound {
		t.Errorf("expected a cluster not found")
	}

	resp, err := cs.Get("foo")
	if err != nil {
		t.Errorf("unexpected error getting cluster: %v", err)
	}

	if resp.Name != "foo" {
		t.Errorf("expected cluster.name to be %s, got %s", "foo", resp.Name)
	}
	if resp.Plan.Etcd.ExpectedCount != 3 {
		t.Errorf("expected etcd.ExpectedCount to be %d, got %d", 3, resp.Plan.Etcd.ExpectedCount)
	}
	if resp.Plan.Master.ExpectedCount != 2 {
		t.Errorf("expected etcd.ExpectedCount to be %d, got %d", 2, resp.Plan.Master.ExpectedCount)
	}
	if resp.Plan.Worker.ExpectedCount != 5 {
		t.Errorf("expected etcd.ExpectedCount to be %d, got %d", 5, resp.Plan.Worker.ExpectedCount)
	}
}
