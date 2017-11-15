package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"crypto/tls"

	"github.com/apprenda/kismatic/pkg/server/http/handler"
	"github.com/julienschmidt/httprouter"
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
	router.GET("/healthz", handler.Healthz)
	router.GET("/clusters", s.ClustersAPI.GetAll)
	router.GET("/clusters/:name", s.ClustersAPI.Get)
	router.DELETE("/clusters/:name", s.ClustersAPI.Delete)
	router.POST("/clusters", s.ClustersAPI.Create)
	router.GET("/clusters/:name/kubeconfig", s.ClustersAPI.GetKubeconfig)
	router.GET("/clusters/:name/logs", s.ClustersAPI.GetLogs)
	router.GET("/clusters/:name/assets", s.ClustersAPI.GetAssets)

	// use our own logger format
	l := negroni.NewLogger()
	l.ALogger = s.Logger
	// use our own logger format
	r := negroni.NewRecovery()
	r.Logger = s.Logger
	r.PrintStack = false
	h := negroni.New(r, l)
	h.UseHandler(router)

	s.httpServer = &http.Server{
		Addr:         addr,
		TLSConfig:    tlsConfig,
		Handler:      h,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	return nil
}

// RunTLS support
func (s *HttpServer) RunTLS() error {
	s.Logger.Printf("Listening on 0.0.0.0%s\n", s.httpServer.Addr)
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
