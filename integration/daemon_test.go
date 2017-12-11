package integration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/apprenda/kismatic/pkg/server/http/handler"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
var secretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
var tlsIgnoringClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

var daemonProcess *os.Process
var daemonPort int

func skipIfAWSCredsMissing() {
	if accessKeyID == "" || secretAccessKey == "" {
		Skip("AWS environment variables are not defined")
	}
}

var _ = Describe("Daemon", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)

		// Each ginkgo runner will get it's own daemon started. Once the spec is
		// done, the daemon is stopped in AfterEach.
		daemonPort = 8080 + config.GinkgoConfig.ParallelNode
		cmd := exec.Command("./kismatic", "server", "-p", strconv.Itoa(daemonPort))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		Expect(err).ToNot(HaveOccurred())
		daemonProcess = cmd.Process
		// wait until server is up
		time.Sleep(3 * time.Second)
	})

	AfterEach(func() {
		// Destroy all clusters
		err := destroyAllClusters(daemonPort)
		if err != nil {
			fmt.Printf(`+++++++++++++++++++++++++++++++++++++

ERROR DESTROYING CLUSTERS ON AWS. MUST BE CLEANED UP MANUALLY.

The error: %v

+++++++++++++++++++++++++++++++++++++`, err)
		}
		err = daemonProcess.Kill()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Posting a cluster resource", func() {
		Context("using AWS as the infrastructure provider", func() {
			It("should create a working cluster on AWS", func() {
				skipIfAWSCredsMissing()

				clusterName := "test-cluster-" + generateRandomString(8)
				payload := handler.ClusterRequest{
					Name:         clusterName,
					DesiredState: "installed",
					EtcdCount:    1,
					MasterCount:  1,
					WorkerCount:  1,
					IngressCount: 1,
					Provisioner: handler.Provisioner{
						Provider: "aws",
						AWSOptions: &handler.AWSProvisionerOptions{
							AccessKeyID:     accessKeyID,
							SecretAccessKey: secretAccessKey,
						},
					},
				}

				pb, err := json.Marshal(payload)
				Expect(err).ToNot(HaveOccurred())

				clustersEndpoint := fmt.Sprintf("https://localhost:%d/clusters", daemonPort)
				resp, err := tlsIgnoringClient.Post(clustersEndpoint, "application/json", bytes.NewBuffer(pb))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))

				// Destroy the cluster once the test is done
				defer requestClusterDeletion(clustersEndpoint, clusterName)

				// Wait up to 20 minutes for the cluster to be up
				deadline := time.After(20 * time.Minute)
				tick := time.Tick(10 * time.Second)
				for {
					select {
					case <-tick:
						By("Checking if the cluster is in state = installed")
						resp, err := tlsIgnoringClient.Get(clustersEndpoint + "/" + clusterName)
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusOK))

						var c handler.ClusterResponse
						err = json.NewDecoder(resp.Body).Decode(&c)
						resp.Body.Close()
						Expect(err).ToNot(HaveOccurred())

						if c.CurrentState == "provisionFailed" || c.CurrentState == "installFailed" {
							printClusterLogs(os.Stdout, clustersEndpoint, clusterName)
							Fail("cluster entered a failure state: " + c.CurrentState)
						}

						if c.CurrentState == "installed" {
							return
						}
					case <-deadline:
						printClusterLogs(os.Stdout, clustersEndpoint, clusterName)
						Fail("timed out waiting for the cluster to be up")
					}
				}
			})
		})
	})
})

func destroyAllClusters(port int) error {
	// Get all clusters
	By("Getting all clusters to destroy them")
	clustersEndpoint := fmt.Sprintf("https://localhost:%d/clusters", port)
	resp, err := tlsIgnoringClient.Get(clustersEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get list of all clusters: %v", err)
	}
	defer resp.Body.Close()

	var clusters []handler.ClusterResponse
	err = json.NewDecoder(resp.Body).Decode(&clusters)
	if err != nil {
		return fmt.Errorf("error decoding server response: %v", err)
	}

	// Short circuit if no clusters are out there
	if len(clusters) == 0 {
		return nil
	}

	// Issue delete request for each cluster
	for _, c := range clusters {
		err := requestClusterDeletion(clustersEndpoint, c.Name)
		if err != nil {
			return fmt.Errorf("error destroying cluster %q: %v", c.Name, err)
		}
	}

	// Wait until all clusters have been deleted
	deadline := time.After(10 * time.Minute)
	tick := time.Tick(10 * time.Second)
	// The only way to exit this loop is to hit the timeout, or for all clusters
	// to be destroyed
	for {
		var clusters []handler.ClusterResponse
		select {
		case <-tick:
			resp, err := tlsIgnoringClient.Get(clustersEndpoint)
			if err != nil {
				return fmt.Errorf("failed to get list of all clusters: %v", err)
			}
			defer resp.Body.Close()

			err = json.NewDecoder(resp.Body).Decode(&clusters)
			if err != nil {
				return fmt.Errorf("error decoding server response: %v", err)
			}

			if len(clusters) == 0 {
				return nil
			}
		case <-deadline:
			return fmt.Errorf("timed out waiting for all clusters to be destroyed. remaining clusters last time we checked: %v", clusters)
		}
	}
}

func requestClusterDeletion(clustersEndpoint, clusterName string) error {
	By(fmt.Sprintf("Destroying cluster %q", clusterName))
	req, err := http.NewRequest("DELETE", clustersEndpoint+"/"+clusterName, nil)
	if err != nil {
		return fmt.Errorf("error building delete request: %v", err)
	}
	resp, err := tlsIgnoringClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending DELETE request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code from DELETE endpoint: %s", resp.Status)
	}
	return nil
}

func printClusterLogs(out io.Writer, clustersEndpoint, clusterName string) error {
	resp, err := tlsIgnoringClient.Get(clustersEndpoint + "/" + clusterName + "/logs")
	if err != nil {
		return fmt.Errorf("error getting logs for cluster %q: %v", clusterName, err)
	}
	fmt.Fprintf(out, "Logs for cluster %s", clusterName)
	io.Copy(out, resp.Body)
	resp.Body.Close()
	fmt.Fprintf(out, "End logs for %s", clusterName)
	return nil
}

func generateRandomString(n int) string {
	// removed 1, l, o, 0 to prevent confusion
	chars := []rune("abcdefghijkmnpqrstuvwxyz23456789")
	res := make([]rune, n)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}
