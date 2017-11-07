package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apprenda/kismatic/pkg/store"
	"github.com/julienschmidt/httprouter"
)

type mockClustersStore struct {
	store map[string]store.Cluster
}

func (cs mockClustersStore) Get(key string) (*store.Cluster, error) {
	c, ok := cs.store[key]
	if !ok {
		return nil, nil
	}
	return &c, nil
}
func (cs *mockClustersStore) Put(key string, cluster store.Cluster) error {
	if cs.store == nil {
		cs.store = make(map[string]store.Cluster)
	}
	cs.store[key] = cluster
	return nil
}

func (cs mockClustersStore) GetAll() (map[string]store.Cluster, error) {
	return cs.store, nil
}

func (cs mockClustersStore) Delete(key string) error {
	delete(cs.store, key)
	return nil
}

func (cs mockClustersStore) Watch(ctx context.Context, buffer uint) <-chan store.WatchResponse {
	return nil
}

func TestCreateGetGetandDelete(t *testing.T) {
	if testing.Short() {
		return
	}
	c := &ClusterRequest{
		Name:         "foo",
		DesiredState: "running",
		AwsID:        "",
		AwsKey:       "",
		Etcd:         3,
		Master:       2,
		Worker:       5,
	}
	encoded, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("could not encode body to json %v", err)
	}
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/clusters", bytes.NewBuffer(encoded))
	if err != nil {
		t.Fatal(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()

	cs := &mockClustersStore{}
	clustersAPI := Clusters{Store: cs}
	r.POST("/clusters", clustersAPI.Create)
	r.ServeHTTP(rr, req)

	// Check the status code is as expected
	if status := rr.Code; status != http.StatusAccepted {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusAccepted, rr.Body.String())
	}

	// Check the response body is as expected
	expected := "ok\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// should get 404
	req, err = http.NewRequest("GET", "/clusters/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters/:name", clustersAPI.Get)
	r.ServeHTTP(rr, req)
	// Check the status code is 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusNotFound, rr.Body.String())
	}

	// should get a response
	req, err = http.NewRequest("GET", "/clusters/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters/:name", clustersAPI.Get)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// should getAll
	req, err = http.NewRequest("GET", "/clusters", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters", clustersAPI.GetAll)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// should delete
	req, err = http.NewRequest("DELETE", "/clusters/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.DELETE("/clusters/:name", clustersAPI.Delete)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusAccepted {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusAccepted, rr.Code)
	}
	expected = "ok\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// should getAll
	req, err = http.NewRequest("GET", "/clusters", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = httprouter.New()
	r.GET("/clusters", clustersAPI.GetAll)
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v: %s",
			status, http.StatusOK, rr.Body.String())
	}
	expected = "[]\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
