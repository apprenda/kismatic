package integration_tests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/apprenda/kismatic/pkg/retry"
	. "github.com/onsi/ginkgo"
)

func verifyIngressNodes(master NodeDeets, ingressNodes []NodeDeets, sshKey string) error {
	By("Adding a service and an ingress resource")
	addIngressResource(master, sshKey)

	By("Verifying the service is accessible via the ingress point(s)")
	for _, ingNode := range ingressNodes {
		if err := verifyIngressPoint(ingNode); err != nil {
			// For debugging purposes...
			runViaSSH([]string{"sudo kubectl --kubeconfig /root/.kube/config describe -f /tmp/ingress.yaml", "sudo kubectl --kubeconfig /root/.kube/config describe pods"}, []NodeDeets{master}, sshKey, 1*time.Minute)
			return err
		}
	}

	return nil
}

func addIngressResource(node NodeDeets, sshKey string) {
	err := copyFileToRemote("test-resources/ingress.yaml", "/tmp/ingress.yaml", node, sshKey, 1*time.Minute)
	FailIfError(err, "Error copying ingress test file")

	err = runViaSSH([]string{"sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/tls.key -out /tmp/tls.crt -subj \"/CN=kismaticintegration.com\""}, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Error creating certificates for HTTPs")

	err = runViaSSH([]string{"sudo kubectl --kubeconfig /root/.kube/config create secret tls kismaticintegration-tls --cert=/tmp/tls.crt --key=/tmp/tls.key"}, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Error creating tls secret")

	err = runViaSSH([]string{"sudo kubectl --kubeconfig /root/.kube/config apply -f /tmp/ingress.yaml"}, []NodeDeets{node}, sshKey, 1*time.Minute)
	FailIfError(err, "Error creating ingress resources")
}

func verifyIngressPoint(node NodeDeets) error {
	// HTTP ingress
	url := "http://" + node.PublicIP + "/echo"
	if err := retry.WithBackoff(func() error { return ingressRequest(url) }, 7); err != nil {
		return err
	}
	// HTTPS ingress
	url = "https://" + node.PublicIP + "/echo-tls"
	if err := retry.WithBackoff(func() error { return ingressRequest(url) }, 7); err != nil {
		return err
	}
	return nil
}

func ingressRequest(url string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   1000 * time.Millisecond,
		Transport: tr,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Could not create request for ingress via %s, %v", url, err)
	}
	// Set the host header since this is not a real domain, curl $IP/echo -H 'Host: kismaticintegration.com'
	req.Host = "kismaticintegration.com"
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not reach ingress via %s, %v", url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Ingress status code is not 200, got %d vi %s", resp.StatusCode, url)
	}

	return nil
}
