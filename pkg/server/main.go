package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	nethttp "net/http"

	"github.com/apprenda/kismatic/pkg/server/http"
)

const (
	defaultTimeout = 10 * time.Second
	defaultPort    = "8443"
)

func main() {
	port := defaultPort
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	logger := http.DefaultLogger(os.Stdout, "[kismatic] ")
	server, err := http.New(logger, port, "", "", defaultTimeout, defaultTimeout)
	if err != nil {
		logger.Fatalf("Error creating server: %v", err)
	}

	// setup interrupt channgel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Println("Starting server...")
		if err := server.RunTLS(); err != nethttp.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	<-stop

	if err := server.Shutdown(30 * time.Second); err != nil {
		logger.Fatalf("Error shutting down server: %v\n", err)
	}
}
