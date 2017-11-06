package store

import (
	"context"
	"encoding/json"

	"github.com/apprenda/kismatic/pkg/install"
)

type Cluster struct {
	DesiredState string
	CurrentState string
	CanContinue  bool
	Plan         install.Plan
	AwsID        string
	AwsKey       string
}

// ClusterStore is a smaller interface into the store
// so that the clients don't need to worry about the bucket
// or marshaling/unmarshaling
type ClusterStore interface {
	Get(key string) (*Cluster, error)
	Put(key string, cluster Cluster) error
	GetAll() (map[string]Cluster, error)
	Watch(ctx context.Context, buffer uint) <-chan WatchResponse
}

type cs struct {
	Bucket string
	Store  WatchedStore
}

func NewClusterStore(store WatchedStore, bucket string) ClusterStore {
	return cs{Store: store, Bucket: bucket}
}

func (s cs) Get(key string) (*Cluster, error) {
	b, err := s.Store.Get(s.Bucket, key)
	if err != nil {
		return nil, err
	}
	var c Cluster
	if b == nil || len(b) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s cs) Put(key string, c Cluster) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return s.Store.Put(s.Bucket, key, b)
}

func (s cs) GetAll() (map[string]Cluster, error) {
	es, err := s.Store.List(s.Bucket)
	if err != nil {
		return nil, err
	}
	m := make(map[string]Cluster)
	for _, e := range es {
		var c Cluster
		err := json.Unmarshal(e.Value, &c)
		if err != nil {
			return nil, err
		}
		m[e.Key] = c
	}
	return m, nil
}

func (s cs) Watch(ctx context.Context, buffer uint) <-chan WatchResponse {
	return s.Store.Watch(ctx, s.Bucket, buffer)
}
