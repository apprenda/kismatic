package data

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

type GlusterInfoGetter interface {
	GetVolumes() (*GlusterVolumeInfoCliOutput, error)
	GetQuota(volume string) (*GlusterVolumeQuotaCliOutput, error)
}

type GlusterCLIGetter struct {
	SSHClient ssh.Client
}

// GetVolumes returns gluster volume data using gluster command on the first sotrage node
func (g GlusterCLIGetter) GetVolumes() (*GlusterVolumeInfoCliOutput, error) {
	glusterVolumeInfoRaw, err := g.SSHClient.Output("sudo gluster volume info all --xml")
	if err != nil {
		return nil, fmt.Errorf("error getting volume info data: %v", err)
	}
	glusterVolumeInfoRaw = strings.TrimSpace(glusterVolumeInfoRaw)
	var glusterVolumeInfo GlusterVolumeInfoCliOutput
	err = xml.Unmarshal([]byte(glusterVolumeInfoRaw), &glusterVolumeInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling volume info data: %v", err)
	}
	if &glusterVolumeInfo == nil || glusterVolumeInfo.VolumeInfo == nil || glusterVolumeInfo.VolumeInfo.Volumes == nil || glusterVolumeInfo.VolumeInfo.Volumes.Volume == nil || len(glusterVolumeInfo.VolumeInfo.Volumes.Volume) == 0 {
		return nil, nil
	}

	return &glusterVolumeInfo, nil
}

// GetQuota returns gluster volume quota data using gluster command on the first sotrage node
func (g GlusterCLIGetter) GetQuota(volume string) (*GlusterVolumeQuotaCliOutput, error) {
	glusterVolumeQuotaRaw, err2 := g.SSHClient.Output(fmt.Sprintf("sudo gluster volume quota %s list --xml", volume))
	if err2 != nil {
		return nil, fmt.Errorf("error getting volume quota data for %s: %v", volume, err2)
	}
	var glusterVolumeQuota GlusterVolumeQuotaCliOutput
	err2 = xml.Unmarshal([]byte(glusterVolumeQuotaRaw), &glusterVolumeQuota)
	if err2 != nil {
		fmt.Printf("error unmarshalling volume quota data: %v", err2)
	}
	if &glusterVolumeQuota == nil || glusterVolumeQuota.VolumeQuota == nil || glusterVolumeQuota.VolumeQuota.Limit == nil {
		return nil, fmt.Errorf("error getting volume quota data for %s", volume)
	}

	return &glusterVolumeQuota, nil
}
