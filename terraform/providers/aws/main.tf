provider "aws" {
  /*
  $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY_ID
  $ export AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_ACCESS_KEY
  $ export AWS_DEFAULT_REGION=us-east-1
  */
  region = "${var.region}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_ami" "centos" {
  most_recent = true

  filter {
    name   = "name"
    values = ["CentOS Linux 7 x86_64 HVM EBS *"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["679593333241"] # AWS marketplace
}

resource "aws_key_pair" "kismatic" {
  key_name   = "${var.cluster_name}"
  public_key = "${file("${var.public_ssh_key_path}")}"
}

resource "aws_vpc" "kismatic" {
  cidr_block            = "10.0.0.0/16"
  enable_dns_support    = true
  enable_dns_hostnames  = true
  tags {
    "Name"                  = "${var.cluster_name}-vpc"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_internet_gateway" "kismatic_gateway" {
  vpc_id = "${aws_vpc.kismatic.id}"
  tags {
    "Name"                  = "${var.cluster_name}-gateway"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_default_route_table" "kismatic_router" {
  default_route_table_id = "${aws_vpc.kismatic.default_route_table_id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.kismatic_gateway.id}"
  }

  tags {
    "Name"                  = "${var.cluster_name}-router"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}


resource "aws_subnet" "kismatic_public" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.1.0/24"
  map_public_ip_on_launch = "True"
  tags {
    "Name"                  = "${var.cluster_name}-subnet-public"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/subnet"       = "public"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_subnet" "kismatic_private" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.2.0/24"
  map_public_ip_on_launch = "True"
  #This needs to be false eventually
  tags {
    "Name"                  = "${var.cluster_name}-subnet-private"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/subnet"       = "private"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_security_group" "kismatic_sec_group" {
  name        = "${var.cluster_name}"
  description = "Allow inbound SSH for kismatic, and all communication between nodes."
  vpc_id      = "${aws_vpc.kismatic.id}"

  ingress {
    from_port   = 8
    to_port     = 0
    protocol    = "icmp"
    self        = "True"
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = "True"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }

  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-public"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/securityGroup" = "public"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_instance" "master" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.master_count}"
  ami                     = "${var.cluster_os == "ubuntu" ? data.aws_ami.ubuntu.id : var.cluster_os == "centos" ? data.aws_ami.centos.id : ""}"
  instance_type           = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-master-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
}

resource "aws_instance" "etcd" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.etcd_count}"
  ami                     = "${var.cluster_os == "ubuntu" ? data.aws_ami.ubuntu.id : var.cluster_os == "centos" ? data.aws_ami.centos.id : ""}"
  instance_type           = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-etcd-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "etcd"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
}

resource "aws_instance" "worker" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.worker_count}"
  ami                     = "${var.cluster_os == "ubuntu" ? data.aws_ami.ubuntu.id : var.cluster_os == "centos" ? data.aws_ami.centos.id : ""}"
  instance_type           = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-worker-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "worker"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
}

resource "aws_instance" "ingress" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.ingress_count}"
  ami                     = "${var.cluster_os == "ubuntu" ? data.aws_ami.ubuntu.id : var.cluster_os == "centos" ? data.aws_ami.centos.id : ""}"
  instance_type           = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-ingress-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
}

resource "aws_instance" "storage" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.storage_count}"
  ami                     = "${var.cluster_os == "ubuntu" ? data.aws_ami.ubuntu.id : var.cluster_os == "centos" ? data.aws_ami.centos.id : ""}"
  instance_type           = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-storage-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "storage"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
}