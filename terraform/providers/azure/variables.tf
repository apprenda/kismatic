variable "location" {
  default = "East US"
}

variable "private_ssh_key_path" {}

variable "public_ssh_key_path" {}

variable "ssh_user" {}

variable "kismatic_version" {}

variable "cluster_name" {}

variable "cluster_owner" {}

variable "instance_size" {
  default = "Standard_D2"
}

variable master_count {}

variable etcd_count {}

variable worker_count {}

variable ingress_count {}

variable storage_count {}