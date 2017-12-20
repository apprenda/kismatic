package store

import (
	"context"
	"encoding/json"
)

// Cluster defines a Kubernetes cluster in KET
type Cluster struct {
	Spec   ClusterSpec
	Status ClusterStatus
}

// The ClusterSpec describes the specifications for the cluster. In other words,
// the desired state and desired configuration of the cluster.
type ClusterSpec struct {
	DesiredState string
	EtcdCount    int
	MasterCount  int
	WorkerCount  int
	IngressCount int
	Provisioner  Provisioner
}

// The ClusterStatus is the current status of the cluster.
type ClusterStatus struct {
	CurrentState          string
	WaitingForManualRetry bool
	ClusterIP             string
}

// The Provisioner specifies the infrastructure provisioner that should be used
// for the cluster.
type Provisioner struct {
	Provider    string
	Credentials ProvisionerCredentials
}

// ProvisionerCredentials are the credentials necessary for connecting to the
// infrastructure provisioner.
type ProvisionerCredentials struct {
	AWS AWSCredentials
}

// AWSCredentials are the credentials for interacting with the AWS API.
type AWSCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
}

// ClusterStore is a smaller interface into the store
// so that the clients don't need to worry about the bucket
// or marshaling/unmarshaling
type ClusterStore interface {
	Get(key string) (*Cluster, error)
	Put(key string, cluster Cluster) error
	GetAll() (map[string]Cluster, error)
	Delete(key string) error
	Watch(ctx context.Context, buffer uint) <-chan WatchResponse
}

type cs struct {
	Bucket string
	Store  WatchedStore
}

// NewClusterStore returns a ClusterStore that manages clusters under the given
// bucket.
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

func (s cs) Delete(key string) error {
	return s.Store.Delete(s.Bucket, key)
}

func (s cs) Watch(ctx context.Context, buffer uint) <-chan WatchResponse {
	return s.Store.Watch(ctx, s.Bucket, buffer)
}
