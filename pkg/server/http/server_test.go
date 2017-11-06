package http

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/apprenda/kismatic/pkg/store"
)

const (
	bucket = "server_test"
)

func setupStore() (store.WatchedStore, error) {
	f, err := ioutil.TempFile("/tmp", "httptests")
	if err != nil {
		return nil, err
	}
	s, err := store.DefaultStore(f.Name(), 0644, log.New(ioutil.Discard, "test ", 0))
	if err != nil {
		return nil, err
	}
	s.CreateBucket(bucket)

	return s, nil
}

func TestNewHTTPServer(t *testing.T) {
	if testing.Short() {
		return
	}
	tests := []struct {
		logger       *log.Logger
		port         string
		certFile     string
		keyFile      string
		readTimeout  time.Duration
		writeTimeout time.Duration
		valid        bool
	}{
		{
			logger:       DefaultLogger(os.Stdout, "[tester] "),
			port:         "443",
			readTimeout:  10 * time.Second,
			writeTimeout: 5 * time.Second,
			valid:        true,
		},
		{
			logger:       DefaultLogger(os.Stdout, "[tester] "),
			port:         "",
			readTimeout:  10 * time.Second,
			writeTimeout: 5 * time.Second,
			valid:        false,
		},
		{
			logger:       DefaultLogger(os.Stdout, "[tester] "),
			port:         "443",
			readTimeout:  -1,
			writeTimeout: 5 * time.Second,
			valid:        false,
		},
		{
			logger:       DefaultLogger(os.Stdout, "[tester] "),
			port:         "443",
			readTimeout:  10 * time.Second,
			writeTimeout: -1,
			valid:        false,
		},
		{
			logger:       DefaultLogger(os.Stdout, "[tester] "),
			port:         "443",
			certFile:     "foo",
			keyFile:      "bar",
			readTimeout:  10 * time.Second,
			writeTimeout: 5 * time.Second,
			valid:        false,
		},
	}
	s, err := setupStore()
	defer s.Close()
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	clusterStore := store.NewClusterStore(s, bucket)
	clusterAPI := handler.Clusters{Store: clusterStore}
	for _, test := range tests {
		server := HttpServer{
			Logger:       test.logger,
			Port:         test.port,
			ClustersAPI:  clusterAPI,
			ReadTimeout:  test.readTimeout,
			WriteTimeout: test.writeTimeout,
			CertFile:     test.certFile,
			KeyFile:      test.keyFile,
		}
		err := server.Init()
		if test.valid && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !test.valid && err == nil {
			t.Fatal("expected error, did not get one")
		}
	}
}
