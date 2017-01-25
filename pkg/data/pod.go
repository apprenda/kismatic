package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type PodGetter interface {
	Get() (*PodList, error)
}

type PlanPodGetter struct {
	SSHClient ssh.Client
	Namespace string
}

// Get returns Pods data
func (g PlanPodGetter) Get() (*PodList, error) {
	ns := fmt.Sprintf("--namespace=%s", g.Namespace)
	if g.Namespace == "all" || g.Namespace == "" {
		ns = "--all-namespaces=true"
	}
	podsRaw, err := g.SSHClient.Output(true, fmt.Sprintf("sudo kubectl get pods %s -o json", ns))
	if err != nil {
		return nil, fmt.Errorf("error getting pod data: %v", err)
	}

	return UnmarshalPods(podsRaw)
}

func UnmarshalPods(raw string) (*PodList, error) {
	// an empty JSON response from kubectl contains this string
	if strings.Contains(raw, "No resources found") {
		return nil, nil
	}
	var pods PodList
	err := json.Unmarshal([]byte(raw), &pods)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pod data: %v", err)
	}

	return &pods, nil
}
