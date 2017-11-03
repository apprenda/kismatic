package cli

import (
	"context"
	"io"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apprenda/kismatic/pkg/controller"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/server/http"
	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/apprenda/kismatic/pkg/server/http/service"
	"github.com/apprenda/kismatic/pkg/store"
	"github.com/spf13/cobra"
)

const (
	loggerPrefix          = "[kismatic] "
	defaultTimeout        = 10 * time.Second
	clustersServiceBucket = "kismatic"
)

type serverOptions struct {
	port     string
	certFile string
	keyFile  string
	dbFile   string
}

// NewCmdServer returns the server command
func NewCmdServer(stdout io.Writer) *cobra.Command {
	var options serverOptions
	cmd := &cobra.Command{
		Use:   "server",
		Short: "server starts an HTTP server",
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
	cmd.Flags().StringVar(&options.dbFile, "db-file", "./server.db", "path to the database file")
	return cmd
}

func doServer(stdout io.Writer, options serverOptions) error {
	logger := log.New(stdout, "[kismatic] ", log.LstdFlags|log.Lshortfile)
	genAssetsDir := "server-assets"

	// Create the store
	s, err := store.NewBoltDB(options.dbFile, 0600, logger)
	if err != nil {
		logger.Fatalf("Error creating store: %v", err)
	}
	err = s.CreateBucket("clusters")
	if err != nil {
		logger.Fatalf("Error creating bucket in store: %v", err)
	}

	// create services and handlers
	clusterService := service.NewClustersService(s, clustersServiceBucket)
	clusterAPI := handler.Clusters{Service: clusterService}

	// Setup the HTTP server
	server := http.HttpServer{
		Logger:       logger,
		Port:         options.port,
		ClustersAPI:  clusterAPI,
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
		CertFile:     options.certFile,
		KeyFile:      options.keyFile,
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

	// Setup the controller
	executorOpts := install.ExecutorOptions{
		GeneratedAssetsDirectory: genAssetsDir,
		RunsDirectory:            "server-runs",
		RestartServices:          true,
		OutputFormat:             "simple",
		Verbose:                  true,
	}
	executor, err := install.NewExecutor(stdout, os.Stderr, executorOpts)
	if err != nil {
		return err
	}

	ctrl := controller.New(logger, executor, s, genAssetsDir, 10*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Setup interrupt channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	cancel()
	if err := server.Shutdown(30 * time.Second); err != nil {
		logger.Fatalf("Error shutting down server: %v", err)
	}
	return nil
}
