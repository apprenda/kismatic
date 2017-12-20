package provision

// Azure provisioner for creating and destroying infrastructure on Azure.
type Azure struct {
	Terraform
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string
}

// AzureTerraformData provider for creating and destroying infrastructure on Azure
type AzureTerraformData struct {
	KismaticVersion   string `json:"kismatic_version"`
	Location          string `json:"location,omitempty"`
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
