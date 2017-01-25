package data

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type GlusterClient interface {
	ListVolumes() (*GlusterVolumeInfoCliOutput, error)
	GetQuota(volume string) (*GlusterVolumeQuotaCliOutput, error)
}

type RemoteGlusterCLI struct {
	SSHClient ssh.Client
}

// ListVolumes returns gluster volume data using gluster command on the first sotrage node
func (g RemoteGlusterCLI) ListVolumes() (*GlusterVolumeInfoCliOutput, error) {
	glusterVolumeInfoRaw, err := g.SSHClient.Output(true, "sudo gluster volume info all --xml")
	if err != nil {
		return nil, fmt.Errorf("error getting volume info data: %v", err)
	}

	return UnmarshalVolumeData(glusterVolumeInfoRaw)
}

func UnmarshalVolumeData(raw string) (*GlusterVolumeInfoCliOutput, error) {
	var glusterVolumeInfo GlusterVolumeInfoCliOutput
	err := xml.Unmarshal([]byte(strings.TrimSpace(raw)), &glusterVolumeInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling volume info data: %v", err)
	}
	if &glusterVolumeInfo == nil || glusterVolumeInfo.VolumeInfo == nil {
		return nil, fmt.Errorf("error getting volume info data")
	}
	if glusterVolumeInfo.VolumeInfo.Volumes == nil || glusterVolumeInfo.VolumeInfo.Volumes.Volume == nil || len(glusterVolumeInfo.VolumeInfo.Volumes.Volume) == 0 {
		return nil, nil
	}

	return &glusterVolumeInfo, nil
}

// GetQuota returns gluster volume quota data using gluster command on the first sotrage node
func (g RemoteGlusterCLI) GetQuota(volume string) (*GlusterVolumeQuotaCliOutput, error) {
	glusterVolumeQuotaRaw, err := g.SSHClient.Output(true, fmt.Sprintf("sudo gluster volume quota %s list --xml", volume))
	if err != nil {
		return nil, fmt.Errorf("error getting volume quota data for %s: %v", volume, err)
	}

	return UnmarshalVolumeQuota(glusterVolumeQuotaRaw)
}

func UnmarshalVolumeQuota(raw string) (*GlusterVolumeQuotaCliOutput, error) {
	if raw == "" {
		return nil, nil
	}
	var glusterVolumeQuota GlusterVolumeQuotaCliOutput
	err := xml.Unmarshal([]byte(strings.TrimSpace(raw)), &glusterVolumeQuota)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling volume quota data: %v", err)
	}

	return &glusterVolumeQuota, nil
}
