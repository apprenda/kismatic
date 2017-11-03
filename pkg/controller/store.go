package controller

import (
	"context"
	"encoding/json"

	"github.com/apprenda/kismatic/pkg/store"
)

// smaller interface into the store so that the controller doesn't need to worry
// about the bucket or marshaling/unmarshaling
type clusterStore interface {
	Get(key string) (*store.Cluster, error)
	Put(key string, cluster store.Cluster) error
	GetAll() (map[string]store.Cluster, error)
	Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse
}

type cs struct {
	bucket string
	store  store.WatchedStore
}

func (s cs) Get(key string) (*store.Cluster, error) {
	b, err := s.store.Get(s.bucket, key)
	if err != nil {
		return nil, err
	}
	var c store.Cluster
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s cs) Put(key string, c store.Cluster) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return s.store.Put(s.bucket, key, b)
}

func (s cs) GetAll() (map[string]store.Cluster, error) {
	es, err := s.store.List(s.bucket)
	if err != nil {
		return nil, err
	}
	m := make(map[string]store.Cluster)
	for _, e := range es {
		var c store.Cluster
		err := json.Unmarshal(e.Value, &c)
		if err != nil {
			return nil, err
		}
		m[e.Key] = c
	}
	return m, nil
}

func (s cs) Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse {
	return s.store.Watch(ctx, s.bucket, buffer)
}
