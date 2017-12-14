variable "region" {
  default = "East US"
}

variable "sub_id" {
  default = ""
}

variable "tenant_id" {
  default = ""
}

variable "client_id" {
  default = ""
}

variable "client_secret" {
  default = ""
}

variable "private_ssh_key_path" {
  default = ""
}

variable "public_ssh_key_path" {
  default = ""
}

variable "ssh_user" {
  default = ""
}

variable "version" {
  default = ""
}

variable "cluster_name" {
  default = "kismatic-cluster"
}

variable "cluster_owner" {
  default = ""
}

variable "instance_size" {
  default = "Standard_B2s"
}

variable master_count {
  description = "Number of k8s master nodes"
  default     = 1
}

variable etcd_count {
  description = "Number of etcd nodes"
  default     = 1
}

variable worker_count {
  description = "Number of k8s worker nodes"
  default     = 1
}

variable ingress_count {
  description = "Number of k8s ingress nodes"
  default     = 1
}

variable storage_count {
  description = "Number of k8s storage nodes"
  default     = 1
}