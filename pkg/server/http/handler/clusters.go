package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/store"
	"github.com/julienschmidt/httprouter"
	"github.com/mholt/archiver"
)

// ErrClusterNotFound is the error returned by the API when a requested cluster
// is not found in the server.
var ErrClusterNotFound = errors.New("cluster details not found in the store")

// the states that can be requested through the API
var validStates = []string{"planned", "provisioned", "installed"}

// the provisioners that are supported
var validProvisionerProviders = []string{"aws"}

// The Clusters handler exposes endpoints for managing the lifecycle of clusters
type Clusters struct {
	Store     store.ClusterStore
	AssetsDir string
	Logger    *log.Logger
}

// ClusterRequest is the cluster resource defined by the user of the API
type ClusterRequest struct {
	Name         string      `json:"name"`
	DesiredState string      `json:"desiredState"`
	EtcdCount    int         `json:"etcdCount"`
	MasterCount  int         `json:"masterCount"`
	WorkerCount  int         `json:"workerCount"`
	IngressCount int         `json:"ingressCount"`
	Provisioner  Provisioner `json:"provisioner"`
}

// ClusterResponse is the cluster resource returned by the server
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

// The Provisioner defines the infrastructure provisioner that should be used
// for hosting the cluster
type Provisioner struct {
	// Options: aws
	Provider   string                 `json:"provider"`
	AWSOptions *AWSProvisionerOptions `json:"options,omitempty"`
}

// The AWSProvisionerOptions are options that are specific to the AWS provisioner
type AWSProvisionerOptions struct {
	install.AWSProvisionerOptions
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
}

// Create a cluster as described in the request body's JSON payload.
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
	cluster := buildStoreCluster(*req)
	if err := putToStore(req.Name, cluster, api.Store); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok\n"))
}

// Update the cluster with the given name
func (api Clusters) Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	fromStore, err := getFromStore(id, api.Store)
	if err != nil {
		if err == ErrClusterNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		api.Logger.Println(errorf(err.Error()))
		return
	}

	req := &ClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	patch := clusterUpdate{id: id, request: *req, inStore: *fromStore}
	// validate request
	valid, errs := patch.validate()
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

	// Update the fields that can be updated
	fromStore.Spec.DesiredState = req.DesiredState
	fromStore.Status.WaitingForManualRetry = false
	fromStore.Spec.Provisioner.Credentials.AWS.AccessKeyId = req.Provisioner.AWSOptions.AccessKeyID
	fromStore.Spec.Provisioner.Credentials.AWS.SecretAccessKey = req.Provisioner.AWSOptions.SecretAccessKey
	fromStore.Spec.MasterCount = req.MasterCount
	fromStore.Spec.WorkerCount = req.WorkerCount
	fromStore.Spec.IngressCount = req.IngressCount

	if err := putToStore(req.Name, *fromStore, api.Store); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}

	// respond with the updated request
	clusterResp := buildResponse(id, *fromStore)
	bytes, err := json.MarshalIndent(clusterResp, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not marshall response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, string(bytes))
}

// Get the cluster with the given name
func (api Clusters) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("name")
	fromStore, err := getFromStore(id, api.Store)
	if err != nil {
		if err == ErrClusterNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			api.Logger.Println(errorf(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	clusterResp := buildResponse(id, *fromStore)
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

// GetAll returns all the clusters that are defined in the API
func (api Clusters) GetAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fromStore, err := getAllFromStore(api.Store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
		return
	}

	clustersResp := make([]ClusterResponse, 0, len(fromStore))
	for key, sc := range fromStore {
		clustersResp = append(clustersResp, buildResponse(key, sc))
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
	fromStore, err := getFromStore(id, api.Store)
	if err != nil {
		if err == ErrClusterNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		api.Logger.Println(errorf(err.Error()))
		return
	}
	// update the state and put to the store
	fromStore.Spec.DesiredState = "destroyed"
	fromStore.Status.WaitingForManualRetry = false
	if err := putToStore(id, *fromStore, api.Store); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf(err.Error()))
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
	f := path.Join(api.AssetsDir, id, "assets", "kubeconfig")
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

// GetAssets creates a tarball with all the assets that were generated for the
// cluster, and return them in the response
func (api Clusters) GetAssets(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
	dir := path.Join(api.AssetsDir, id, "assets")
	if stat, err := os.Stat(dir); os.IsNotExist(err) || !stat.IsDir() {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("assets for cluster %s could not be retrieved: %v", id, err))
		return
	}
	// create a temp dir to store the tar assets
	tmpf, err := ioutil.TempFile("/tmp", id)
	defer os.Remove(tmpf.Name())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not create an assets file for cluster %s: %v", id, err))
		return
	}
	// archive the directory
	err = archiver.TarGz.Make(tmpf.Name(), []string{dir})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logger.Println(errorf("could not archive tge assets file for cluster %s: %v", id, err))
		return
	}
	attachmentName := fmt.Sprintf("attachment; filename=%s-assets.tar.gz", id)
	w.Header().Set("Content-Disposition", attachmentName)
	http.ServeFile(w, r, tmpf.Name())
}

func putToStore(name string, toStore store.Cluster, cs store.ClusterStore) error {
	if err := cs.Put(name, toStore); err != nil {
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

func getFromStore(name string, cs store.ClusterStore) (*store.Cluster, error) {
	sc, err := cs.Get(name)
	if err != nil {
		return nil, fmt.Errorf("could not get from the store: %v", err)
	}
	if sc == nil {
		return nil, ErrClusterNotFound
	}
	return sc, nil
}

func getAllFromStore(cs store.ClusterStore) (map[string]store.Cluster, error) {
	msc, err := cs.GetAll()
	if err != nil {
		return nil, fmt.Errorf("could not get from the store: %v", err)
	}
	if msc == nil {
		return make(map[string]store.Cluster, 0), nil
	}
	return msc, nil
}

func buildStoreCluster(req ClusterRequest) store.Cluster {
	spec := store.ClusterSpec{
		DesiredState: req.DesiredState,
		EtcdCount:    req.EtcdCount,
		MasterCount:  req.MasterCount,
		WorkerCount:  req.WorkerCount,
		IngressCount: req.IngressCount,
		Provisioner: store.Provisioner{
			Provider: req.Provisioner.Provider,
		},
	}
	switch req.Provisioner.Provider {
	case "aws":
		if req.Provisioner.AWSOptions != nil {
			creds := store.ProvisionerCredentials{
				AWS: store.AWSCredentials{
					AccessKeyId:     req.Provisioner.AWSOptions.AccessKeyID,
					SecretAccessKey: req.Provisioner.AWSOptions.SecretAccessKey,
				},
			}
			spec.Provisioner.Credentials = creds
		}
	}
	return store.Cluster{
		Spec: spec,
	}
}

func buildResponse(name string, sc store.Cluster) ClusterResponse {
	provisioner := Provisioner{
		Provider: sc.Spec.Provisioner.Provider,
	}
	// TODO: The user can post provisioner specific options... We need to be
	// able to show them back to the user
	return ClusterResponse{
		Name:         name,
		DesiredState: sc.Spec.DesiredState,
		EtcdCount:    sc.Spec.EtcdCount,
		MasterCount:  sc.Spec.MasterCount,
		WorkerCount:  sc.Spec.WorkerCount,
		IngressCount: sc.Spec.IngressCount,
		Provisioner:  provisioner,
		CurrentState: sc.Status.CurrentState,
		ClusterIP:    sc.Status.ClusterIP,
	}
}
