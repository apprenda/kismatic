package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	nethttp "net/http"

	"github.com/apprenda/kismatic/pkg/server/http"
	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/apprenda/kismatic/pkg/store"
)

const (
	defaultTimeout = 10 * time.Second
	defaultPort    = "8443"
	bucket         = "kismatic"
)

func main() {
	port := defaultPort
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	logger := http.DefaultLogger(os.Stdout, "[kismatic] ")

	storeFile, err := ioutil.TempFile("/tmp", "ket-server-store")
	if err != nil {
		logger.Fatalf("error creating temp directory: %v", err)
	}
	s, err := store.New(storeFile.Name(), 0644, logger)
	if err != nil {
		logger.Fatalf("Error opening store: %v", err)
	}
	defer s.Close()
	if err := s.CreateBucket(bucket); err != nil {
		logger.Fatalf("Error creating bucket: %v", err)
	}

	clusterStore := store.NewClusterStore(s, bucket)

	assetsDir, err := ioutil.TempDir("/tmp", "ket-server-assets")
	if err != nil {
		logger.Fatalf("error creating assets directory %q: %v", assetsDir, err)
	}

	// create handlers
	clusterAPI := handler.Clusters{Store: clusterStore, AssetsDir: assetsDir, Logger: logger}

	// Setup the HTTP server
	server := http.HttpServer{
		Logger:       logger,
		Port:         port,
		ClustersAPI:  clusterAPI,
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	}

	if err := server.Init(); err != nil {
		logger.Fatalf("Error creating server: %v", err)
	}

	go func() {
		logger.Println("Starting server...")
		if err := server.RunTLS(); err != nethttp.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	// setup interrupt channgel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	if err := server.Shutdown(30 * time.Second); err != nil {
		logger.Fatalf("Error shutting down server: %v\n", err)
	}
}
