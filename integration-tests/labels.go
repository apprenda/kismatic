package integration_tests

import (
	"fmt"
	"time"
)

func containsLabels(nodes provisionedNodes, sshKey string) error {
	// labels were hardcoded in th plan pattern
	tests := []struct {
		nodes []NodeDeets
		label string
	}{
		{
			nodes: nodes.master,
			label: "com.integrationtest/master:true",
		},
		{
			nodes: nodes.worker,
			label: "com.integrationtest/worker:true",
		},
		{
			nodes: nodes.ingress,
			label: "com.integrationtest/ingress:true",
		},
		{
			nodes: nodes.storage,
			label: "com.integrationtest/storage:true",
		},
	}
	for _, role := range tests {
		for _, n := range role.nodes {
			if err := runViaSSH([]string{fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get nodes %s -o jsonpath='{.metadata.labels}' | grep %q", n.Hostname, role.label)}, []NodeDeets{nodes.master[0]}, sshKey, 1*time.Minute); err != nil {
				return fmt.Errorf("error validating node %q label: %v", n.Hostname, err)
			}
		}
	}

	return nil
}
