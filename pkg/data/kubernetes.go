package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type PodLister interface {
	ListPods() (*PodList, error)
}

type PVLister interface {
	ListPersistentVolumes() (*PersistentVolumeList, error)
}

type KubernetesClient interface {
	PodLister
	PVLister
}

// RemoteKubectl
type RemoteKubectl struct {
	SSHClient ssh.Client
}

// ListPersistentVolumes returns PersistentVolume data
func (k RemoteKubectl) ListPersistentVolumes() (*PersistentVolumeList, error) {
	pvRaw, err := k.SSHClient.Output(true, "sudo kubectl get pv -o json")
	if err != nil {
		return nil, fmt.Errorf("error getting persistent volume data: %v", err)
	}

	return UnmarshalPVs(pvRaw)
}

func UnmarshalPVs(raw string) (*PersistentVolumeList, error) {
	// an empty JSON response from kubectl contains this string
	if strings.Contains(strings.TrimSpace(raw), "No resources found") {
		return nil, nil
	}
	var pvs PersistentVolumeList
	err := json.Unmarshal([]byte(raw), &pvs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent volume data: %v", err)
	}

	return &pvs, nil
}

// ListPods returns Pods data with --all-namespaces=true flag
func (k RemoteKubectl) ListPods() (*PodList, error) {
	podsRaw, err := k.SSHClient.Output(true, "sudo kubectl get pods --all-namespaces=true -o json")
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
