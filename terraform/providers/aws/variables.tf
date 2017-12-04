variable "region" {
  default = "us-east-1"
}

variable "access_key" {
  default = ""
}

variable "secret_key" {
  default = ""
}

variable "private_ssh_key_path" {
  default = ""
}

variable "public_ssh_key_path" {
  description = "SSH Public Key"
  default = ""
}

variable "ssh_user" {
  default = "ubuntu"
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

variable "ami" {
  default = "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"
  //These will have to change when we want to also support RHEL/CentOS
}

variable "instance_size" {
  default = "t2.medium"
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