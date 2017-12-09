package provision

// AWS provisioner for creating and destroying infrastructure on AWS.
type AWS struct {
	Terraform
	AccessKeyID     string
	SecretAccessKey string
}

// AWSTerraformData provider for creating and destroying infrastructure on AWS
type AWSTerraformData struct {
	Version           string `json:"version"`
	Region            string `json:"region,omitempty"`
	AvailabilityZone  string `json:"AZ,omitempty"`
	PrivateSSHKeyPath string `json:"private_ssh_key_path"`
	PublicSSHKeyPath  string `json:"public_ssh_key_path"`
	SSHUser           string `json:"ssh_user"`
	ClusterName       string `json:"cluster_name"`
	ClusterOwner      string `json:"cluster_owner"`
	MasterCount       int    `json:"master_count"`
	EtcdCount         int    `json:"etcd_count"`
	WorkerCount       int    `json:"worker_count"`
	IngressCount      int    `json:"ingress_count"`
	StorageCount      int    `json:"storage_count"`
}
