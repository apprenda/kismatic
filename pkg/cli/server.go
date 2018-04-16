package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/apprenda/kismatic/pkg/controller"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/plan"
	"github.com/apprenda/kismatic/pkg/provision"
	"github.com/apprenda/kismatic/pkg/server/http"
	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/apprenda/kismatic/pkg/store"
	"github.com/spf13/cobra"
)

const (
	defaultInsecurePort = "8080"
	defaultSecurePort   = "8443"
	defaultTimeout      = 10 * time.Second
)

type serverOptions struct {
	port       string
	certFile   string
	keyFile    string
	dbFile     string
	disableTLS bool
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
	cmd.Flags().StringVarP(&options.port, "port", "p", "", "port to start the server on. Defaults to 8443, or 8080 when TLS is disabled.")
	cmd.Flags().StringVar(&options.certFile, "cert-file", "", "path to the TLS cert file")
	cmd.Flags().StringVar(&options.keyFile, "key-file", "", "path to the TLS key file")
	cmd.Flags().StringVar(&options.dbFile, "db-file", filepath.Join(assetsFolder, "server.db"), "path to the database file")
	cmd.Flags().BoolVar(&options.disableTLS, "insecure-disable-tls", false, "set to true to disable TLS")
	return cmd
}

func doServer(stdout io.Writer, options serverOptions) error {
	logger := log.New(stdout, "[kismatic] ", log.LstdFlags|log.Lshortfile)
	parent, _ := filepath.Split(options.dbFile)
	if err := os.MkdirAll(parent, 0700); err != nil {
		logger.Fatalf("Error creating store directory structure: %v", err)
	}
	// Create the store
	s, err := store.New(options.dbFile, 0600, logger)
	if err != nil {
		logger.Fatalf("Error creating store: %v", err)
	}
	defer s.Close()
	err = s.CreateBucket(clustersBucket)
	if err != nil {
		logger.Fatalf("Error creating bucket in store: %v", err)
	}

	clusterStore := store.NewClusterStore(s, clustersBucket)

	// Create a dir where all the controller-related files will be stored
	// Needs to be an absoulute path for the HTTP server
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Could not get current directory for assets: %v", err)
	}
	assetsDir := path.Join(pwd, assetsFolder)
	err = os.MkdirAll(assetsDir, 0700)
	if err != nil {
		logger.Fatalf("Error creating assets directory %q: %v", assetsDir, err)
	}

	// create handlers
	clusterAPI := handler.Clusters{Store: clusterStore, AssetsDir: assetsDir, Logger: logger}

	port := defaultSecurePort
	if options.disableTLS {
		port = defaultInsecurePort
	}
	if options.port != "" {
		port = options.port
	}

	// Setup the HTTP server
	server := http.HttpServer{
		Logger:       logger,
		Port:         port,
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
		if err := server.Run(options.disableTLS); err != nethttp.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	provisionerCreator := func(out io.Writer) provision.Provisioner {
		return provision.AnyTerraform{
			Output:          out,
			BinaryPath:      filepath.Join(pwd, "terraform"),
			KismaticVersion: install.KismaticVersion.String(),
			StateDir:        assetsDir,
			ProvidersDir:    filepath.Join(pwd, "providers"),
			SecretsGetter:   storeSecretsGetter{store: clusterStore},
		}
	}

	planner := plan.ProviderTemplatePlanner{
		ProvidersDir: filepath.Join(pwd, "providers"),
	}

	ctrl := controller.New(
		logger,
		planner,
		controller.DefaultExecutorCreator(),
		provisionerCreator,
		clusterStore,
		10*time.Minute,
		controller.AssetsDir(assetsDir),
	)
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

type storeSecretsGetter struct {
	store store.ClusterStore
}

// GetAsEnvironmentVariables returns the expected environment variables sourcing
// them from the cluster store.
func (ssg storeSecretsGetter) GetAsEnvironmentVariables(clusterName string, expected map[string]string) ([]string, error) {
	c, err := ssg.store.Get(clusterName)
	if err != nil {
		return nil, err
	}
	storedSecrets := c.Spec.Provisioner.Secrets
	var envVars []string
	for expectedKey, envVar := range expected {
		for key, value := range storedSecrets {
			if key == expectedKey {
				envVars = append(envVars, fmt.Sprintf("%s=%s", envVar, value))
				continue
			}
		}
		if value := os.Getenv(envVar); value != "" {
			envVars = append(envVars, fmt.Sprintf("%s=%s", envVar, value))
			continue
		}
		return nil, fmt.Errorf("required option %q was not provided", expectedKey)
	}
	return envVars, nil
}
