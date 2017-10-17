package cli

import "strings"

//ListResponse contains a slice of volumes with corresponding json fields
type ListResponse struct {
	Volumes []Volume `json:"items"`
}

//Volume contains name, storage class, labels, capacity, availability, replicacount, distribution count, bricks, status, claim, and pods
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

//Brick contains Host and Path information
type Brick struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

//Claim contains Name and Namespace information
type Claim struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

//Pod contains Name and Namespace information, as well as associated containers
type Pod struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Containers []Container `json:"containers"`
}

//Container contains Name and mounting information (name and path)
type Container struct {
	Name      string `json:"name"`
	MountName string `json:"mountName"`
	MountPath string `json:"mountPath"`
}

//Readable joins a claim's namespace and name in the format "namespace/name"
func (c *Claim) Readable() string {
	if c != nil {
		return strings.Join([]string{c.Namespace, c.Name}, "/")
	}
	return ""
}

//Readable joins a pod's namespace and name in the format "namespace/name"
func (p *Pod) Readable() string {
	if p != nil {
		return strings.Join([]string{p.Namespace, p.Name}, "/")
	}
	return ""
}

//Readable joins a brick's host and path in the format "host:path"
func (b *Brick) Readable() string {
	if b != nil {
		return strings.Join([]string{b.Host, b.Path}, ":")
	}
	return ""
}

//VolumeBrickToString joins the volume bricks into a CSV string
func VolumeBrickToString(bricks []Brick) string {
	var vbList []string
	for _, brick := range bricks {
		vbList = append(vbList, brick.Readable())
	}
	return strings.Join(vbList, ",")
}
