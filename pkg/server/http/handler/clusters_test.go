package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apprenda/kismatic/pkg/server/http/model"
	"github.com/apprenda/kismatic/pkg/server/http/service"
	"github.com/julienschmidt/httprouter"
)

type mockClustersService struct {
	store map[string][]byte
}

func (cs *mockClustersService) Create(c *model.ClusterRequest) error {
	if cs.store == nil {
		cs.store = make(map[string][]byte)
	}
	b, err := service.MarshalForStore(c)
	if err != nil {
		return err
	}
	cs.store[c.Name] = b
	return nil
}

func (cs *mockClustersService) Get(name string) (*model.ClusterResponse, error) {
	v, ok := cs.store[name]
	if !ok {
		return nil, service.ErrClusterNotFound
	}
	return service.UnmarshalFromStore(name, v)
}

func TestCreateAndGet(t *testing.T) {
	if testing.Short() {
		return
	}
	c := &model.ClusterRequest{
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

	cs := &mockClustersService{}
	clustersAPI := Clusters{Service: cs}
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
	req, err = http.NewRequest("GET", "/clusters/bar", bytes.NewBuffer(encoded))
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
	req, err = http.NewRequest("GET", "/clusters/foo", bytes.NewBuffer(encoded))
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
}
