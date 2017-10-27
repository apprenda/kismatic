package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestHealthzHandler(t *testing.T) {
	if testing.Short() {
		return
	}
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	// Call their ServeHTTP method directly and pass in our Request and ResponseRecorder
	r := httprouter.New()
	r.GET("/healthz", Healthz)
	r.ServeHTTP(rr, req)

	// Check the status code is as expected
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is as expected
	expected := `ok`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
