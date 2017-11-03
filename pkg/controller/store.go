package controller

import (
	"context"
	"encoding/json"

	"github.com/apprenda/kismatic/pkg/store"
)

// smaller interface into the store so that the controller doesn't need to worry
// about the bucket or marshaling/unmarshaling
type clusterStore interface {
	Get(key string) (*planWrapper, error)
	Put(key string, pw planWrapper) error
	GetAll() (map[string]planWrapper, error)
	Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse
}

type cs struct {
	bucket string
	store  store.WatchedStore
}

func (s cs) Get(key string) (*planWrapper, error) {
	b, err := s.store.Get(s.bucket, key)
	if err != nil {
		return nil, err
	}
	var pw planWrapper
	err = json.Unmarshal(b, &pw)
	if err != nil {
		return nil, err
	}
	return &pw, nil
}

func (s cs) Put(key string, pw planWrapper) error {
	b, err := json.Marshal(pw)
	if err != nil {
		return err
	}
	return s.store.Put(s.bucket, key, b)
}

func (s cs) GetAll() (map[string]planWrapper, error) {
	es, err := s.store.List(s.bucket)
	if err != nil {
		return nil, err
	}
	m := make(map[string]planWrapper)
	for _, e := range es {
		var pw planWrapper
		err := json.Unmarshal(e.Value, &pw)
		if err != nil {
			return nil, err
		}
		m[e.Key] = pw
	}
	return m, nil
}

func (s cs) Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse {
	return s.store.Watch(ctx, s.bucket, buffer)
}
