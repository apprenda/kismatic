# This is what needs to be generated from the plan file.
# Plan will need to prompt the user if they need to provision infrastructure.
# If yes, plan will generate the terraform template instead of the full plan file.
# Prompt for provider (only aws for now)
# prompt for OS (only ubuntu for now)
# prompt for instance size (exact definitions will vary from provider to provider)
  /*
  $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY_ID
  $ export AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_ACCESS_KEY
  $ export AWS_DEFAULT_REGION=us-east-1

  can be used instead of the following variables
  */

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

variable "public_ssh_key" {
  description = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"
  #As an example.
  /*
    When importing an existing key pair the public key material may be in any format supported by AWS. Supported formats (per the AWS documentation) are:

    OpenSSH public key format (the format in ~/.ssh/authorized_keys)
    Base64 encoded DER format
    SSH public key file format as specified in RFC4716
  */
  #If being used explicitly through KET, this should already be filled in.

  default = ""
}
#This will have to be populated from the plan.

variable "cluster_name" {
  default = "kismatic-cluster"
}

variable "ami" {
  default = "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"
  #This might be a pain to do extensibly.
}

variable "instance_size" {
  default = "t2.medium"
  #This too.
  #Recommend at least a t2.medium
}

# Reference architecture.
# This can be reused from the plan code that already exists.
# It'll just have to be marshalled differently 
# (terraform accepts JSON by default as well)
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