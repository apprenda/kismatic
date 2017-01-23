package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type PVGetter interface {
	Get() (*PersistentVolumeList, error)
}

type PlanPVGetter struct {
	SSHClient ssh.Client
}

// Get returns PersistentVolume data
func (g PlanPVGetter) Get() (*PersistentVolumeList, error) {
	pvRaw, err := g.SSHClient.Output("sudo kubectl get pv -o json")
	if err != nil {
		return nil, fmt.Errorf("error getting persistent volume data: %v", err)
	}
	pvRaw = strings.TrimSpace(pvRaw)
	// an empty JSON response from kubectl contains this string
	if strings.Contains(pvRaw, "No resources found") {
		return nil, nil
	}
	var pvs PersistentVolumeList
	err = json.Unmarshal([]byte(pvRaw), &pvs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent volume data: %v", err)
	}

	return &pvs, nil
}
