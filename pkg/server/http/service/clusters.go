package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/server/http/model"
	"github.com/apprenda/kismatic/pkg/store"
)

var ErrClusterNotFound = errors.New("cluster details not found in the store")

type Clusters interface {
	Create(c *model.ClusterRequest) error
	Get(name string) (*model.ClusterResponse, error)
}

type clustersService struct {
	store  store.WatchedStore
	bucket string
}

func NewClustersService(store store.WatchedStore, bucket string) Clusters {
	return clustersService{store: store, bucket: bucket}
}

func (cs clustersService) Get(name string) (*model.ClusterResponse, error) {
	v, err := cs.store.Get(cs.bucket, name)
	if err != nil {
		return nil, fmt.Errorf("could not get the cluster details from the store: %v", err)
	}
	if v == nil || len(v) == 0 {
		return nil, ErrClusterNotFound
	}
	return UnmarshalFromStore(name, v)
}

func (cs clustersService) Create(c *model.ClusterRequest) error {
	b, err := MarshalForStore(c)
	if err != nil {
		return err
	}
	if err := cs.store.Put(cs.bucket, c.Name, b); err != nil {
		return fmt.Errorf("could not persist the cluster in the store: %v", err)
	}
	return nil
}

func MarshalForStore(c *model.ClusterRequest) ([]byte, error) {
	// build the plan template
	planTemplate := install.PlanTemplateOptions{
		EtcdNodes:   c.Etcd,
		MasterNodes: c.Master,
		WorkerNodes: c.Worker,
	}
	planner := &install.BytesPlanner{}
	if err := install.WritePlanTemplate(planTemplate, planner); err != nil {
		return nil, fmt.Errorf("could not decode body: %v", err)
	}
	var p *install.Plan
	p, err := planner.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read plan: %v", err)
	}
	// set some defaults in the plan
	p.Cluster.Name = c.Name
	storeCluster := store.Cluster{
		DesiredState: c.DesiredState,
		Plan:         *p,
		CanContinue:  true,
		AwsID:        c.AwsID,
		AwsKey:       c.AwsKey,
	}
	b, err := json.Marshal(storeCluster)
	if err != nil {
		return nil, fmt.Errorf("could not encode plan into store format: %v", err)
	}
	return b, nil
}

func UnmarshalFromStore(name string, v []byte) (*model.ClusterResponse, error) {
	storeCluster := &store.Cluster{}
	err := json.Unmarshal(v, storeCluster)
	if err != nil {
		return nil, fmt.Errorf("could not decode plan from store format: %v", err)
	}
	resp := &model.ClusterResponse{
		Name:         name,
		DesiredState: storeCluster.DesiredState,
		CurrentState: storeCluster.CurrentState,
		Plan:         storeCluster.Plan,
	}
	return resp, nil
}
