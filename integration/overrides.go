package integration

import (
	"fmt"
	"time"
)

func ContainsOverrides(nodes provisionedNodes, sshKey string) error {
	// The installer defaults to --v=2, check if was overridden to --v=3
	tests := []struct {
		nodes []NodeDeets
		text  string
	}{
		{
			nodes: nodes.master,
			text:  "v=3",
		},
		{
			nodes: nodes.worker,
			text:  "v=3",
		},
		{
			nodes: nodes.ingress,
			text:  "v=3",
		},
		{
			nodes: nodes.storage,
			text:  "v=3",
		},
	}
	manifests := []string{"kube-apiserver.yaml", "kube-controller-manager.yaml", "kube-scheduler.yaml"}
	// validate master components overrides
	for _, m := range manifests {
		if err := runViaSSH([]string{fmt.Sprintf("sudo cat /etc/kubernetes/manifests/%s | grep \"v=3\"", m)}, []NodeDeets{nodes.master[0]}, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("error validating file %q label: %v", m, err)
		}
	}

	for _, role := range tests {
		for _, n := range role.nodes {
			if err := runViaSSH([]string{fmt.Sprintf("sudo cat /etc/systemd/system/kubelet.service | grep \"%s\"", role.text)}, []NodeDeets{n}, sshKey, 1*time.Minute); err != nil {
				return fmt.Errorf("error validating kubelet override on %q: %v", n.Hostname, err)
			}
		}
	}

	return nil
}
