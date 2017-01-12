package util

import (
	"crypto/tls"
	"net/http"
	"time"
)

// HTTPGet a proxy to http.Get with timeout and insecure cert
// timeout of 0 means no timeout
func HTTPGet(url string, timeout time.Duration, insecure bool) (int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
