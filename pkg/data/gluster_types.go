package data

// gluster volume quota $VOLUME list --xml
//==============================================================================
type GlusterVolumeQuotaCliOutput struct {
	VolumeQuota *GlusterVolumeQuota `xml:" volQuota,omitempty" json:"volQuota,omitempty"`
}

type GlusterVolumeLimit struct {
	AvailSpace       float64 `xml:" avail_space,omitempty" json:"avail_space,omitempty"`
	HardLimit        float64 `xml:" hard_limit,omitempty" json:"hard_limit,omitempty"`
	HlExceeded       string  `xml:" hl_exceeded,omitempty" json:"hl_exceeded,omitempty"`
	SlExceeded       string  `xml:" sl_exceeded,omitempty" json:"sl_exceeded,omitempty"`
	SoftLimitPercent string  `xml:" soft_limit_percent,omitempty" json:"soft_limit_percent,omitempty"`
	SoftLimitValue   float64 `xml:" soft_limit_value,omitempty" json:"soft_limit_value,omitempty"`
	UsedSpace        float64 `xml:" used_space,omitempty" json:"used_space,omitempty"`
}

type GlusterVolumeQuota struct {
	Limit *GlusterVolumeLimit `xml:" limit,omitempty" json:"limit,omitempty"`
}

// gluster volume info all --xml
//==============================================================================
type GlusterVolumeInfoCliOutput struct {
	VolumeInfo *GlusterVolumeInfo `xml:" volInfo,omitempty" json:"volInfo,omitempty"`
}
type GlusterBrick struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type GlusterBricks struct {
	Brick []*GlusterBrick `xml:" brick,omitempty" json:"brick,omitempty"`
}

type GlusterVolumeInfo struct {
	Volumes *GlusterVolumes `xml:" volumes,omitempty" json:"volumes,omitempty"`
}

type GlusterVolume struct {
	BrickCount   uint           `xml:" brickCount,omitempty" json:"brickCount,omitempty"`
	Bricks       *GlusterBricks `xml:" bricks,omitempty" json:"bricks,omitempty"`
	DistCount    uint           `xml:" distCount,omitempty" json:"distCount,omitempty"`
	Name         string         `xml:" name,omitempty" json:"name,omitempty"`
	ReplicaCount uint           `xml:" replicaCount,omitempty" json:"replicaCount,omitempty"`
}

type GlusterVolumes struct {
	Count  uint             `xml:" count,omitempty" json:"count,omitempty"`
	Volume []*GlusterVolume `xml:" volume,omitempty" json:"volume,omitempty"`
}
