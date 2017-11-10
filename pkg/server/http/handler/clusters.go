package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

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
	Store     store.ClusterStore
	AssetsDir string
	Logger    *log.Logger
}

// TODO add validation to all requests
func (api Clusters) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req := &ClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	if err := putToStore(req, api.Store); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok\n"))
}

func (api Clusters) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	clusterResp, err := getFromStore(p.ByName("name"), api.Store)
	if err != nil {
		if err == ErrClusterNotFound {
			http.Error(w, "", http.StatusNotFound)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
		api.Logger.Println(errorf(err.Error()))
		return
	}

	err = json.NewEncoder(w).Encode(clusterResp)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf("could not marshall response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (api Clusters) GetAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	clustersResp, err := getAllFromStore(api.Store)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}

	err = json.NewEncoder(w).Encode(clustersResp)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf("could not marshall response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (api Clusters) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	if err := api.Store.Delete(id); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf("could not delete from the store: %v", err))
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok\n"))
}

// GetKubeconfig will return the kubeconfig file for a cluster :name
// 404 is returned if the cluster is not found in the store
// 500 is returned when the cluster is in the store but the file does not exist in the assets
func (api Clusters) GetKubeconfig(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	exists, err := existsInStore(id, api.Store)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if !exists {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	f := path.Join(api.AssetsDir, id, "generated", "kubeconfig")
	if stat, err := os.Stat(f); os.IsNotExist(err) || stat.IsDir() {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf("kubeconfig for cluster %s could not be retrieved: %v", id, err))
	}
	// set so the browser downloads it instead of displaying it
	w.Header().Set("Content-Disposition", "attachment; filename=config")
	http.ServeFile(w, r, f)
}

// GetLogs will return the log file for a cluster :name
// A 404 is returned if a file is not found in the store
// 500 is returned when the cluster is in the store but the file does not exist in the assets
func (api Clusters) GetLogs(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	exists, err := existsInStore(id, api.Store)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if !exists {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	f := path.Join(api.AssetsDir, id, "kismatic.log")
	if stat, err := os.Stat(f); os.IsNotExist(err) || stat.IsDir() {
		http.Error(w, "", http.StatusInternalServerError)
		api.Logger.Println(errorf("logs for cluster %s could not be retrieved: %v", id, err))
	}
	// set so the browser downloads it instead of displaying it
	w.Header().Set("Content-Disposition", "attachment; filename=kismatic.log")
	http.ServeFile(w, r, f)
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
	p.Provisioner.Provider = "aws"
	sc := store.Cluster{
		DesiredState: req.DesiredState,
		CurrentState: "planned",
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

func existsInStore(name string, cs store.ClusterStore) (bool, error) {
	sc, err := cs.Get(name)
	if err != nil {
		return false, fmt.Errorf("could not get from the store: %v", err)
	}
	return sc != nil, nil
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

func getAllFromStore(cs store.ClusterStore) ([]ClusterResponse, error) {
	msc, err := cs.GetAll()
	if err != nil {
		return nil, fmt.Errorf("could not get from the store: %v", err)
	}
	resp := make([]ClusterResponse, 0)
	if msc == nil {
		return resp, nil
	}
	for key, sc := range msc {
		r := ClusterResponse{
			Name:         key,
			DesiredState: sc.DesiredState,
			CurrentState: sc.CurrentState,
			Plan:         sc.Plan,
		}
		resp = append(resp, r)
	}
	return resp, nil
}
