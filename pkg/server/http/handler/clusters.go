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
	"github.com/apprenda/kismatic/pkg/util"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/julienschmidt/httprouter"
)

var ErrClusterNotFound = errors.New("cluster details not found in the store")

// TODO should this be extracted from the install pkg?
type validatable interface {
	validate() (bool, []error)
}

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{
		errs: []error{},
	}
}

func (v *validator) addError(err ...error) {
	v.errs = append(v.errs, err...)
}

func (v *validator) validate(obj validatable) {
	if ok, err := obj.validate(); !ok {
		v.addError(err...)
	}
}

func (v *validator) valid() (bool, []error) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}

func (r *ClusterRequest) validate() (bool, []error) {
	v := newValidator()
	if r.Name == "" {
		v.addError(fmt.Errorf("name cannot be empty"))
	}
	if r.DesiredState == "" {
		v.addError(fmt.Errorf("desiredState cannot be empty"))
	} else {
		if !util.Contains(r.DesiredState, validStates) {
			v.addError(fmt.Errorf("%s is not a valid desiredState, options are: %v", r.DesiredState, validStates))
		}
	}
	if r.EtcdCount <= 0 {
		v.addError(fmt.Errorf("cluster.etcdCount must be greater than 0"))
	}
	if r.MasterCount <= 0 {
		v.addError(fmt.Errorf("cluster.masterCount must be greater than 0"))
	}
	if r.WorkerCount <= 0 {
		v.addError(fmt.Errorf("cluster.workerCount must be greater than 0"))
	}
	if r.IngressCount < 0 {
		v.addError(fmt.Errorf("cluster.ingressCount must be greater than or equal to 0"))
	}
	v.validate(&r.Provisioner)
	return v.valid()
}

func (p *Provisioner) validate() (bool, []error) {
	v := newValidator()
	if p.Provider == "" {
		v.addError(fmt.Errorf("provisioner.provider cannot be empty"))
	} else {
		if !util.Contains(p.Provider, validProvisionerProviders) {
			v.addError(fmt.Errorf("%s is not a valid provisioner.provider, options are: %v", p.Provider, validProvisionerProviders))
		}
		switch p.Provider {
		case "aws":
			if p.AWSOptions == nil || p.AWSOptions.AccessKeyID == "" {
				v.addError(fmt.Errorf("provisioner.options.accessKeyID cannot be empty"))
			}
			if p.AWSOptions == nil || p.AWSOptions.SecretAccessKey == "" {
				v.addError(fmt.Errorf("provisioner.options.secretAccessKey cannot be empty"))
			}
		}
	}
	return v.valid()
}

func formatErrs(errs []error) []string {
	out := make([]string, 0)
	for _, err := range errs {
		out = append(out, err.Error())
	}
	return out
}

type ClusterRequest struct {
	Name         string      `json:"name"`
	DesiredState string      `json:"desiredState"`
	ClusterIP    string      `json:"clusterIP"`
	EtcdCount    int         `json:"etcdCount"`
	MasterCount  int         `json:"masterCount"`
	WorkerCount  int         `json:"workerCount"`
	IngressCount int         `json:"ingressCount"`
	Provisioner  Provisioner `json:"provisioner"`
}

var validStates = []string{"installed"}
var validProvisionerProviders = []string{"aws"}

type ClusterResponse struct {
	Name         string      `json:"name"`
	DesiredState string      `json:"desiredState"`
	CurrentState string      `json:"currentState"`
	ClusterIP    string      `json:"clusterIP"`
	EtcdCount    int         `json:"etcdCount"`
	MasterCount  int         `json:"masterCount"`
	WorkerCount  int         `json:"workerCount"`
	IngressCount int         `json:"ingressCount"`
	Provisioner  Provisioner `json:"provisioner"`
}

type Provisioner struct {
	// Options: aws
	Provider   string                 `json:"provider"`
	AWSOptions *AWSProvisionerOptions `json:"options,omitempty"`
}

type Cluster struct {
}

type AWSProvisionerOptions struct {
	*install.AWSProvisionerOptions
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type Clusters struct {
	Store     store.ClusterStore
	AssetsDir string
	Logger    *log.Logger
}

func (api Clusters) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req := &ClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	// validate request
	valid, errs := req.validate()
	if !valid {
		bytes, err := json.MarshalIndent(formatErrs(errs), "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			api.Logger.Println(errorf("could not marshall response: %v", err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, string(bytes), http.StatusBadRequest)
		return
	}
	// confirm the name is unique
	exists, err := existsInStore(req.Name, api.Store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err := putToStore(req, api.Store); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		api.Logger.Println(errorf(err.Error()))
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(clusterResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not marshall response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (api Clusters) GetAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	clustersResp, err := getAllFromStore(api.Store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(clustersResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not marshall response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// Delete a cluster
// 404 is returned if the cluster is not found in the store
func (api Clusters) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	exists, err := existsInStore(id, api.Store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := api.Store.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not delete from the store: %v", err))
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	f := path.Join(api.AssetsDir, id, "generated", "kubeconfig")
	if stat, err := os.Stat(f); os.IsNotExist(err) || stat.IsDir() {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("kubeconfig for cluster %s could not be retrieved: %v", id, err))
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	f := path.Join(api.AssetsDir, id, "kismatic.log")
	if stat, err := os.Stat(f); os.IsNotExist(err) || stat.IsDir() {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("logs for cluster %s could not be retrieved: %v", id, err))
		return
	}
	http.ServeFile(w, r, f)
}

func putToStore(req *ClusterRequest, cs store.ClusterStore) error {
	// build the plan template
	planTemplate := install.PlanTemplateOptions{
		EtcdNodes:    req.EtcdCount,
		MasterNodes:  req.MasterCount,
		WorkerNodes:  req.WorkerCount,
		IngressNodes: req.IngressCount,
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
	p.Provisioner = install.Provisioner{Provider: req.Provisioner.Provider, AWSOptions: req.Provisioner.AWSOptions.AWSProvisionerOptions}
	sc := store.Cluster{
		DesiredState: req.DesiredState,
		CurrentState: "planned",
		Plan:         *p,
		CanContinue:  true,
	}
	switch p.Provisioner.Provider {
	case "aws":
		creds := store.ProvisionerCredentials{
			AWS: store.AWSCredentials{
				AccessKeyId:     req.Provisioner.AWSOptions.AccessKeyID,
				SecretAccessKey: req.Provisioner.AWSOptions.SecretAccessKey,
			},
		}
		sc.ProvisionerCredentials = creds
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
	resp := buildResponse(name, *sc)
	return &resp, nil
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
		resp = append(resp, buildResponse(key, sc))
	}
	return resp, nil
}

func buildResponse(name string, sc store.Cluster) ClusterResponse {
	provisioner := Provisioner{
		Provider: sc.Plan.Provisioner.Provider,
	}
	switch sc.Plan.Provisioner.Provider {
	case "aws":
		if sc.Plan.Provisioner.AWSOptions != nil {
			provisioner.AWSOptions = &AWSProvisionerOptions{
				AWSProvisionerOptions: sc.Plan.Provisioner.AWSOptions,
			}
		}
	}
	return ClusterResponse{
		Name:         name,
		DesiredState: sc.DesiredState,
		CurrentState: sc.CurrentState,
		ClusterIP:    sc.Plan.Master.LoadBalancedFQDN,
		EtcdCount:    sc.Plan.Etcd.ExpectedCount,
		MasterCount:  sc.Plan.Master.ExpectedCount,
		WorkerCount:  sc.Plan.Worker.ExpectedCount,
		IngressCount: sc.Plan.Ingress.ExpectedCount,
		Provisioner:  provisioner,
	}
}
