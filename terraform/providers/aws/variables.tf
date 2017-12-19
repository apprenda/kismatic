variable "region" {
  default = "us-east-1"
}

variable "availability_zone" {
  default = "us-east-1c"
}

variable "private_ssh_key_path" {
  description = "Path to the SSH private key"
}

variable "public_ssh_key_path" {
  description = "Path to the SSH public key"
}

variable "ssh_user" {
  description = "The SSH user that should be used for accessing the nodes over SSH"
}

variable "kismatic_version" {}

variable "cluster_name" {}

variable "cluster_owner" {}

variable "ami" {
  // These will have to change when we want to also support RHEL/CentOS
  default = "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"
}

variable "instance_size" {
  default = "t2.medium"
}

variable master_count {
  description = "Number of k8s master nodes"
}

variable etcd_count {
  description = "Number of etcd nodes"
}

variable worker_count {
  description = "Number of k8s worker nodes"
}

variable ingress_count {
  description = "Number of k8s ingress nodes"
}

variable storage_count {
  description = "Number of k8s storage nodes"
}
