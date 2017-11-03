package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apprenda/kismatic/pkg/server/http/model"
	"github.com/apprenda/kismatic/pkg/server/http/service"
	"github.com/julienschmidt/httprouter"
)

type Clusters struct {
	Service service.Clusters
}

func (api Clusters) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := &model.ClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	err := api.Service.Create(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err.Error())
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok\n"))
}

func (api Clusters) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	cluster, err := api.Service.Get(p.ByName("name"))
	if err != nil {
		if err == service.ErrClusterNotFound {
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
	resp, err := json.MarshalIndent(cluster, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "could not marshall response\n")
	}
	fmt.Fprintln(w, string(resp))
}
