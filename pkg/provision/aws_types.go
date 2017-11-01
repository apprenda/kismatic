package provision

type AWS struct {
	Terraform
	AWSTerraformData
}

// AWSTerraformData provider for creating and destroying infrastructure on AWS
type AWSTerraformData struct {
	Region            string `json:"region,omitempty"`
	AccessKey         string `json:"access_key"`
	SecretKey         string `json:"secret_key"`
	PrivateSSHKeyPath string `json:"private_ssh_key_path"`
	PublicSSHKey      string `json:"public_ssh_key"`
	ClusterName       string `json:"cluster_name"`
	AMI               string `json:"ami,omitempty"`
	EC2InstanceType   string `json:"instance_size,omitempty"`
	MasterCount       int    `json:"master_count"`
	EtcdCount         int    `json:"etcd_count"`
	WorkerCount       int    `json:"worker_count"`
	IngressCount      int    `json:"ingress_count"`
	StorageCount      int    `json:"storage_count"`
}
