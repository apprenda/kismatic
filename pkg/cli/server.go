package cli

import (
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	nethttp "net/http"

	"github.com/apprenda/kismatic/pkg/server/http"

	"github.com/spf13/cobra"
)

const (
	defaultTimeout = 10 * time.Second
)

type serverOptions struct {
	port     string
	certFile string
	keyFile  string
}

func NewCmdServer(stdout io.Writer) *cobra.Command {
	var options serverOptions
	cmd := &cobra.Command{
		Use:   "server",
		Short: "server starts an http server",
		Long: `
Start an HTTP server to manage KET clusters. The API has endpoints to create, mutate, delete and view clusters.

A local datastore will be created to persist the state of the clusters managed by this server.

If cert-file or key-file are not provided, a self-signed CA will be used to create the required key-pair for TLS. 
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return cmd.Usage()
			}
			return doServer(stdout, options)
		},
	}
	cmd.Flags().StringVarP(&options.port, "port", "p", "443", "port to start the server on")
	cmd.Flags().StringVar(&options.certFile, "cert-file", "", "path to the TLS cert file")
	cmd.Flags().StringVar(&options.keyFile, "key-file", "", "path to the TLS key file")
	return cmd
}

func doServer(stdout io.Writer, options serverOptions) error {
	logger := http.DefaultLogger(stdout, "[kismatic] ")
	server, err := http.New(logger, options.port, options.certFile, options.keyFile, defaultTimeout, defaultTimeout)
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
		logger.Fatalf("Error shutting down server: %v", err)
	}
	return nil
}
