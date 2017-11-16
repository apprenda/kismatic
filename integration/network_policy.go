package integration

import (
	"fmt"
	"time"

	"github.com/apprenda/kismatic/pkg/retry"
	. "github.com/onsi/ginkgo"
)

// pre17 uses annotations and the extensions/v1beta1 API
func verifyNetworkPolicy(node NodeDeets, sshKey string, pre17 bool) error {
	By("deplying test pods")
	if err := copyFileToRemote("test-resources/network-policy/tester.yaml", "/tmp/tester.yaml", node, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("could not copy network-policy tester to remote: %v", err)
	}
	if err := runViaSSH([]string{"sudo kubectl apply -f /tmp/tester.yaml"}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("could not deploy network-policy tester to remote: %v", err)
	}

	By("testing connection with policy disabled")
	if err := testPodAccess(node, sshKey, 5); err != nil {
		return fmt.Errorf("could not connect to pod: %v", err)
	}

	By("enabling global network policy on the policy-tester namespace")
	if pre17 {
		if err := retry.WithBackoff(func() error {
			return runViaSSH([]string{`sudo kubectl annotate ns policy-tester "net.beta.kubernetes.io/network-policy={\"ingress\": {\"isolation\": \"DefaultDeny\"}}" --overwrite`}, []NodeDeets{node}, sshKey, 1*time.Minute)
		}, 3); err != nil {
			return fmt.Errorf("could not set deny policy: %v", err)
		}
	} else {
		if err := copyFileToRemote("test-resources/network-policy/default-deny.yaml", "/tmp/default-deny.yaml", node, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("could not copy default-deny network-policy resource to remote: %v", err)
		}
		if err := runViaSSH([]string{"sudo kubectl apply -f /tmp/default-deny.yaml"}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("could not deploy default-deny network-policy resource to remote: %v", err)
		}
	}

	By("testing connection with global policy enabled")
	if err := testPodAccess(node, sshKey, 1); err == nil {
		return fmt.Errorf("expected connection to fail and it did not")
	}

	policyFile := "policy.yaml"
	if pre17 {
		policyFile = "policy-pre17.yaml"
	}
	By("applying a policy to allow test pods communication")
	if err := copyFileToRemote("test-resources/network-policy/"+policyFile, "/tmp/policy.yaml", node, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("could not copy pod network-policy resources to remote: %v", err)
	}
	if err := runViaSSH([]string{"sudo kubectl apply -f /tmp/policy.yaml"}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
		return fmt.Errorf("could not deploy pod network-policy resources to remote: %v", err)
	}

	By("testing connection with global policy enabled and pod policy deployed")
	if err := testPodAccess(node, sshKey, 5); err != nil {
		return fmt.Errorf("could not connect to pod after allowing traffic: %v", err)
	}

	// always try to disable global policy
	By("disabling global network policy on the policy-tester namespace")
	if pre17 {
		if err := retry.WithBackoff(func() error {
			return runViaSSH([]string{`sudo kubectl annotate ns policy-tester "net.beta.kubernetes.io/network-policy={\"ingress\": {\"isolation\": \"DefaultAllow\"}}" --overwrite`}, []NodeDeets{node}, sshKey, 1*time.Minute)
		}, 3); err != nil {
			return fmt.Errorf("could not unset deny policy: %v\n", err)
		}
	} else {
		if err := runViaSSH([]string{"sudo kubectl delete -f /tmp/default-deny.yaml"}, []NodeDeets{node}, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("could not deploy default-deny network-policy resource to remote: %v", err)
		}
	}

	return nil
}

func testPodAccess(node NodeDeets, sshKey string, tries uint) error {
	return retry.WithBackoff(func() error {
		return runViaSSH([]string{"sudo kubectl exec -n policy-tester -it network-policy-tester -- wget --spider --timeout=1 network-policy-echoserver"}, []NodeDeets{node}, sshKey, 1*time.Minute)
	}, tries)
}
