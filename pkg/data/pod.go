package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
	"k8s.io/kubernetes/pkg/api/v1"
)

type PodGetter interface {
	Get() (*v1.PodList, error)
}

type PlanPodGetter struct {
	SSHClient ssh.Client
	Namespace string
}

// Get returns Pods data
func (g PlanPodGetter) Get() (*v1.PodList, error) {
	ns := fmt.Sprintf("--namespace=%s", g.Namespace)
	if g.Namespace == "all" || g.Namespace == "" {
		ns = "--all-namespaces=true"
	}
	podsRaw, err := g.SSHClient.Output(fmt.Sprintf("sudo kubectl get pods %s -o json", ns))
	if err != nil {
		return nil, fmt.Errorf("error getting pod data: %v", err)
	}
	podsRaw = strings.TrimSpace(podsRaw)
	var pods v1.PodList
	err = json.Unmarshal([]byte(podsRaw), &pods)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pod data: %v", err)
	}

	return &pods, nil
}
