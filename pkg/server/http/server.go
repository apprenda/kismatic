// Package classification Kismatic API.
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: http, https
//     APIPath: /api/v1
//     Version: 0.0.1
//     License: Apache 2.0 http://www.apache.org/licenses/
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - text/plain
//     - application/json
//
// swagger:meta
package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"crypto/tls"

	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// Server to run an HTTPs server
type Server interface {
	RunTLS() error
	Shutdown(timeout time.Duration) error
}

type HttpServer struct {
	httpServer   *http.Server
	CertFile     string
	KeyFile      string
	Logger       *log.Logger
	Port         string
	ClustersAPI  handler.Clusters
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

const (
	version = "v1"
	//APIPath is the path to interact with the business logic of the server
	APIPath = "/api/" + version
	//DocsPath is where the swagger UI will be served to
	DocsPath = "/docs/" + version
	//SpecPath is where the raw swagger.json will be served to.
	SpecPath = "/spec/" + version
)

// Init creates a configured http server
// If certificates are not provided, a self signed CA will be used
// Use 0 for no read and write timeouts
func (s *HttpServer) Init() error {

	if s.Logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}
	if s.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	addr := fmt.Sprintf(":%s", s.Port)
	if s.ReadTimeout < 0 {
		return fmt.Errorf("readTimeout cannot be negative")
	}
	if s.ReadTimeout == 0 {
		s.Logger.Printf("ReadTimeout is set to 0 and will never timeout, you may want to provide a timeout value\n")
	}
	if s.WriteTimeout < 0 {
		return fmt.Errorf("writeTimeout cannot be negative")
	}
	if s.WriteTimeout == 0 {
		s.Logger.Printf("WriteTimeout is set to 0 and will never timeout, you may want to provide a timeout value\n")
	}
	// use self signed CA
	var keyPair tls.Certificate
	if s.CertFile == "" || s.KeyFile == "" {
		s.Logger.Printf("Using self-signed certificate\n")
		key, cert, err := selfSignedCert()
		if err != nil {
			return fmt.Errorf("could not get self-signed certificate key-pair: %v", err)
		}
		if keyPair, err = tls.X509KeyPair(cert, key); err != nil {
			return fmt.Errorf("could not parse key-pair: %v", err)
		}
	} else {
		var err error
		if keyPair, err = tls.LoadX509KeyPair(s.CertFile, s.KeyFile); err != nil {
			return fmt.Errorf("could not load provided key-pair: %v", err)
		}
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{keyPair}}

	// setup routes
	router := httprouter.New()
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// swagger:operation GET /docs/v1/ getDocs
	//
	// Returns the swagger ui for the API documentation
	//
	// ---
	// responses:
	//   '200':
	//     description: "ok"
	//   '404':
	//	   description: "file not found"
	router.ServeFiles(DocsPath+"/*filepath", http.Dir(filepath.Join(wd, "swagger-ui/dist")))
	// swagger:operation GET /spec/v1/swagger.json getSpec
	//
	// Returns the swagger.json spec file
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: "ok"
	//   '404':
	//     description: "file not found"
	router.ServeFiles(SpecPath+"/*filepath", http.Dir(filepath.Join(wd, "swagger-ui/spec")))
	// swagger:operation GET /healthz getHealthz
	//
	// Returns server health
	//
	// ---
	// produces:
	// - text/plain
	// responses:
	//   '200':
	//     description: "Server ok"
	//     schema:
	//       type: string
	router.GET("/healthz", handler.Healthz)
	// swagger:operation GET /api/v1/clusters getAllClusters
	//
	// Returns all clusters the server is managing
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: Body contains JSON spec for all clusters
	//   '500':
	//     description: Marshalling/fetching error
	router.GET(APIPath+"/clusters", s.ClustersAPI.GetAll)
	// swagger:operation GET /api/v1/clusters/{name} getCluster
	//
	// Returns cluster with {name}
	//
	// Could be any cluster being managed
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the cluster
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: OK
	//   '404':
	//     description: cluster with {name} not found
	//   '500':
	//     description: marshalling/fetching error
	router.GET(APIPath+"/clusters/:name", s.ClustersAPI.Get)
	// swagger:operation DELETE /api/v1/clusters/{name} deleteCluster
	//
	// Deletes cluster with {name}
	//
	// Could be any cluster being managed
	//
	// ---
	// produces:
	// - text/plain
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the cluster
	//   required: true
	//   type: string
	// responses:
	//   '202':
	//     description: OK
	//   '404':
	//     description: cluster with {name} not found
	//   '500':
	//     description: marshalling/fetching error
	router.DELETE(APIPath+"/clusters/:name", s.ClustersAPI.Delete)
	// swagger:operation POST /api/v1/clusters createCluster
	//
	// Creates cluster with {name}
	//
	// Creates cluster according to provided spec
	//
	// ---
	// produces:
	// - text/plain
	// consumes:
	// - application/json
	// parameters:
	// - name: cluster
	//   in: body
	//   description: the cluster specification to create
	//   required: true
	// responses:
	//   '202':
	//     description: Accepted
	//   '400':
	//     description: malformed request
	//   '409':
	//     description: cluster with {name} already exists
	//   '500':
	//     description: marshalling/fetching error
	router.POST(APIPath+"/clusters", s.ClustersAPI.Create)
	// swagger:operation PUT /api/v1/clusters/{name} updateCluster
	//
	// Updates cluster with {name}
	//
	// Updates cluster according to provided spec - still requiring a full request.
	//
	// ---
	// produces:
	// - application/json
	// consumes:
	// - application/json
	// parameters:
	// - name: cluster
	//   in: body
	//   description: the cluster specification to create
	//   required: true
	// responses:
	//   '202':
	//     description: Accepted
	//   '400':
	//     description: malformed request
	//   '404':
	//     description: cluster with {name} not found
	//   '500':
	//     description: marshalling/fetching error
	router.PUT(APIPath+"/clusters/:name", s.ClustersAPI.Update)
	// swagger:operation GET /api/v1/clusters/{name}/kubeconfig getKubeconfig
	//
	// Returns kubeconfig for cluster with {name}
	//
	// Downloads the kubeconfig for the specified cluster
	//
	// ---
	// produces:
	// - text/plain
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the cluster
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: ok
	//   '404':
	//     description: cluster with {name} not found
	//   '500':
	//     description: couldn't find kubeconfig
	router.GET(APIPath+"/clusters/:name/kubeconfig", s.ClustersAPI.GetKubeconfig)
	// swagger:operation GET /api/v1/clusters/{name}/logs getLogs
	//
	// Returns the Kismatic installer logs for the cluster
	//
	// Displays the plaintext Kismatic logs
	//
	// ---
	// produces:
	// - text/plain
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the cluster
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: The plain text logs for the cluster installation
	//   '404':
	//     description: couldn't find logs
	//   '500':
	//     description: marshalling/fetching error
	router.GET(APIPath+"/clusters/:name/logs", s.ClustersAPI.GetLogs)
	// swagger:operation GET /api/v1/clusters/{name}/assets getAssets
	//
	// Returns all assets for cluster with {name}
	//
	// Downloads assets as a tarball
	//
	// ---
	// produces:
	// - application/gzip
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the cluster
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: ''
	//   '404':
	//     description: couldn't find assets
	//   '500':
	//     description: marshalling/fetching error
	router.GET(APIPath+"/clusters/:name/assets", s.ClustersAPI.GetAssets)

	// use our own logger format
	l := negroni.NewLogger()
	l.ALogger = s.Logger
	// use our own logger format
	r := negroni.NewRecovery()
	r.Logger = s.Logger
	r.PrintStack = false
	h := negroni.New(r, l)
	h.UseHandler(router)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080", "https://localhost:8443"},
	})
	h.Use(c)
	s.httpServer = &http.Server{
		Addr:         addr,
		TLSConfig:    tlsConfig,
		Handler:      h,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	return nil
}

// Run starts the HTTP server
func (s *HttpServer) Run(disableTLS bool) error {
	s.Logger.Printf("Listening on 0.0.0.0%s\n", s.httpServer.Addr)
	if disableTLS {
		return s.httpServer.ListenAndServe()
	}
	return s.httpServer.ListenAndServeTLS("", "")
}

// Shutdown will gracefully shutdown the server
func (s *HttpServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.Logger.Println("Shutting down the server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	s.Logger.Println("Server stopped")

	return nil
}

// DefaultLogger returns a logger the specified writer and prefix
func DefaultLogger(out io.Writer, prefix string) *log.Logger {
	return log.New(out, prefix, log.Ldate|log.Ltime)
}
