package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	nethttp "net/http"

	"github.com/apprenda/kismatic/pkg/server/http"
	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/apprenda/kismatic/pkg/server/http/service"
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
	s, err := store.DefaultStore("/tmp/kismatic", 0644, logger)
	defer s.Close()
	if err != nil {
		logger.Fatalf("Error opening store: %v", err)
	}
	if err := s.CreateBucket(bucket); err != nil {
		logger.Fatalf("Error creating bucket: %v", err)
	}
	// create services and handlers
	clusterService := service.NewClustersService(s, bucket)
	clusterAPI := handler.Clusters{Service: clusterService}

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
