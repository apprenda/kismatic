package cli

import "strings"

// ListResponse
type ListResponse struct {
	Volumes []Volume `json:"items"`
}

// Volume
type Volume struct {
	Name              string            `json:"name"`
	StorageClass      string            `json:"storageClass,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Capacity          string            `json:"capacity"`
	Available         string            `json:"available"`
	ReplicaCount      uint              `json:"replicaCount"`
	DistributionCount uint              `json:"distributionCount"`
	Bricks            []Brick           `json:"bricks"`
	Status            string            `json:"status"`
	Claim             *Claim            `json:"claim,omitempty"`
	Pods              []Pod             `json:"pods,omitempty"`
}

// Brick
type Brick struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

// Claim
type Claim struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Pod
type Pod struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Containers []Container `json:"containers"`
}

// Container
type Container struct {
	Name      string `json:"name"`
	MountName string `json:"mountName"`
	MountPath string `json:"mountPath"`
}

func (c *Claim) Readable() string {
	if c != nil {
		return strings.Join([]string{c.Namespace, c.Name}, "/")
	}
	return ""
}

func (p *Pod) Readable() string {
	if p != nil {
		return strings.Join([]string{p.Namespace, p.Name}, "/")
	}
	return ""
}

func (b *Brick) Readable() string {
	if b != nil {
		return strings.Join([]string{b.Host, b.Path}, ":")
	}
	return ""
}

func VolumeBrickToString(bricks []Brick) string {
	var vbList []string
	for _, brick := range bricks {
		vbList = append(vbList, brick.Readable())
	}
	return strings.Join(vbList, ",")
}
