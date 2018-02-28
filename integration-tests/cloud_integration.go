package integration_tests

import (
	"fmt"
	"time"

	"github.com/apprenda/kismatic/pkg/retry"
)

func testAWSCloudProvider(node NodeDeets, sshKey string) error {
	// uses an ECR image to test cloud-provider
	if err := runViaSSH([]string{`sudo kubectl --kubeconfig /root/.kube/config run cloud-provider-nginx --image=633668368853.dkr.ecr.us-east-1.amazonaws.com/kismatic/nginx --replicas=2 --port=80`}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("error creating nginx deployment: %v", err)
	}

	if err := runViaSSH([]string{`sudo kubectl --kubeconfig /root/.kube/config expose deployment cloud-provider-nginx --port=80 --type=LoadBalancer`}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("error creating exposing nginx deployment with a LoadBalancer: %v", err)
	}

	testErr := retry.WithBackoff(func() error {
		return runViaSSH([]string{"curl `sudo kubectl --kubeconfig /root/.kube/config get svc cloud-provider-nginx -o jsonpath={.status.loadBalancer.ingress[0].hostname}`"}, []NodeDeets{node}, sshKey, 1*time.Minute)
	}, 8)

	if testErr != nil {
		// get info for diagnosing failure
		runViaSSH([]string{"sudo kubectl --kubeconfig /root/.kube/config describe svc cloud-provider-nginx"}, []NodeDeets{node}, sshKey, 1*time.Minute)
	}

	if err := runViaSSH([]string{`sudo kubectl --kubeconfig /root/.kube/config delete svc cloud-provider-nginx`}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("error deleting service: %v", err)
	}

	if testErr != nil {
		return fmt.Errorf("error curling LoadBalancer endpoint: %v", testErr)
	}

	return nil
}
