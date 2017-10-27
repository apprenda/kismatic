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

type loggerHttpServer struct {
	httpServer *http.Server
	logger     *log.Logger
}

// New creates a configured http server
// If certificates are not provided, a self signed CA will be used
// Use 0 for no read and write timeouts
func New(logger *log.Logger, port string, certFile string, keyFile string, readTimeout time.Duration, writeTimeout time.Duration) (Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if port == "" {
		return nil, fmt.Errorf("port cannot be empty")
	}
	addr := fmt.Sprintf(":%s", port)
	if readTimeout < 0 {
		return nil, fmt.Errorf("readTimeout cannot be negative")
	}
	if readTimeout == 0 {
		logger.Printf("ReadTimeout is set to 0 and will never timeout, you may want to provide a timeout value\n")
	}
	if writeTimeout < 0 {
		return nil, fmt.Errorf("writeTimeout cannot be negative")
	}
	if writeTimeout == 0 {
		logger.Printf("WriteTimeout is set to 0 and will never timeout, you may want to provide a timeout value\n")
	}
	// use self signed CA
	var keyPair tls.Certificate
	if certFile == "" || keyFile == "" {
		logger.Printf("Using self-signed certificate\n")
		key, cert, err := selfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("could not get self-signed certificate key-pair: %v", err)
		}
		if keyPair, err = tls.X509KeyPair(cert, key); err != nil {
			return nil, fmt.Errorf("could not parse key-pair: %v", err)
		}
	} else {
		var err error
		if keyPair, err = tls.LoadX509KeyPair(certFile, keyFile); err != nil {
			return nil, fmt.Errorf("could not load provided key-pair: %v", err)
		}
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{keyPair}}

	// setup routes
	router := httprouter.New()
	router.GET("/healthz", handler.Healthz)

	// use our own logger format
	l := negroni.NewLogger()
	l.Logger = logger
	// use our own logger format
	r := negroni.NewRecovery()
	r.Logger = logger
	r.PrintStack = false
	h := negroni.New(r, l)
	h.UseHandler(router)

	s := &loggerHttpServer{
		httpServer: &http.Server{
			Addr:         addr,
			TLSConfig:    tlsConfig,
			Handler:      h,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		logger: logger,
	}

	return s, nil
}

// RunTLS support
func (s *loggerHttpServer) RunTLS() error {
	s.logger.Printf("Listening on 0.0.0.0%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServeTLS("", "")
}

// Shutdown will gracefully shutdown the server
func (s *loggerHttpServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.logger.Println("Shutting down the server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	s.logger.Println("Server stopped")

	return nil
}

// DefaultLogger returns a logger the specified writer and prefix
func DefaultLogger(out io.Writer, prefix string) *log.Logger {
	return log.New(out, prefix, log.Ldate|log.Ltime)
}
