package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
	"k8s.io/kubernetes/pkg/api/v1"
)

type PVGetter interface {
	Get() (*v1.PersistentVolumeList, error)
}

type PlanPVGetter struct {
	SSHClient ssh.Client
}

// Get returns PersistentVolume data
func (g PlanPVGetter) Get() (*v1.PersistentVolumeList, error) {
	pvRaw, err := g.SSHClient.Output("sudo kubectl get pv -o json")
	if err != nil {
		return nil, fmt.Errorf("error getting persistent volume data: %v", err)
	}
	pvRaw = strings.TrimSpace(pvRaw)
	var pvs v1.PersistentVolumeList
	err = json.Unmarshal([]byte(pvRaw), &pvs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent volume data: %v", err)
	}

	return &pvs, nil
}
