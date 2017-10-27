package http

import (
	"log"
	"os"
	"testing"
	"time"
)

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
	for _, test := range tests {
		_, err := New(test.logger, test.port, test.certFile, test.keyFile, test.readTimeout, test.writeTimeout)
		if test.valid && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !test.valid && err == nil {
			t.Fatal("expected error, did not get one")
		}
	}
}
