package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/apprenda/kismatic/pkg/store"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/julienschmidt/httprouter"
)

var ErrClusterNotFound = errors.New("cluster details not found in the store")

type ClusterRequest struct {
	Name         string
	DesiredState string
	AwsID        string
	AwsKey       string
	Etcd         int
	Master       int
	Worker       int
}

type ClusterResponse struct {
	Name         string
	DesiredState string
	CurrentState string
	install.Plan
}

type Clusters struct {
	Store store.ClusterStore
}

func (api Clusters) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req := &ClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	if err := putToStore(req, api.Store); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err.Error())
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok\n"))
}

func (api Clusters) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	clusterResp, err := getFromStore(p.ByName("name"), api.Store)
	if err != nil {
		if err == ErrClusterNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}
	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp, err := json.MarshalIndent(clusterResp, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "could not marshall response\n")
	}
	fmt.Fprintln(w, string(resp))
}

func putToStore(req *ClusterRequest, cs store.ClusterStore) error {
	// build the plan template
	planTemplate := install.PlanTemplateOptions{
		EtcdNodes:   req.Etcd,
		MasterNodes: req.Master,
		WorkerNodes: req.Worker,
	}
	planner := &install.BytesPlanner{}
	if err := install.WritePlanTemplate(planTemplate, planner); err != nil {
		return fmt.Errorf("could not decode request body: %v", err)
	}
	var p *install.Plan
	p, err := planner.Read()
	if err != nil {
		return fmt.Errorf("could not read plan: %v", err)
	}
	// set some defaults in the plan
	p.Cluster.Name = req.Name
	sc := store.Cluster{
		DesiredState: req.DesiredState,
		Plan:         *p,
		CanContinue:  true,
		AwsID:        req.AwsID,
		AwsKey:       req.AwsKey,
	}
	if err := cs.Put(req.Name, sc); err != nil {
		return fmt.Errorf("could not put to the store: %v", err)
	}
	return nil
}

func getFromStore(name string, cs store.ClusterStore) (*ClusterResponse, error) {
	sc, err := cs.Get(name)
	if err != nil {
		return nil, fmt.Errorf("could not get from the store: %v", err)
	}
	if sc == nil {
		return nil, ErrClusterNotFound
	}
	resp := &ClusterResponse{
		Name:         name,
		DesiredState: sc.DesiredState,
		CurrentState: sc.CurrentState,
		Plan:         sc.Plan,
	}
	return resp, nil
}
