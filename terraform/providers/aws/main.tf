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
    values = ["${var.ami}"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
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
    Name                  = "${var.cluster_name}-vpc"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kubernetes.io/cluster = "${var.cluster_name}"
  }
}

resource "aws_internet_gateway" "kismatic_gateway" {
  vpc_id = "${aws_vpc.kismatic.id}"
  tags {
    Name                  = "${var.cluster_name}-gateway"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kubernetes.io/cluster = "${var.cluster_name}"
  }
}

resource "aws_default_route_table" "kismatic_router" {
  default_route_table_id = "${aws_vpc.kismatic.default_route_table_id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.kismatic_gateway.id}"
  }

  tags {
    Name                  = "${var.cluster_name}-router"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kubernetes.io/cluster = "${var.cluster_name}"
  }
}


resource "aws_subnet" "kismatic_public" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.1.0/24"
  map_public_ip_on_launch = "True"
  tags {
    Name                  = "${var.cluster_name}-subnet-public"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/subnet       = "public"
    kubernetes.io/cluster = "${var.cluster_name}"
  }
}

resource "aws_subnet" "kismatic_private" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.2.0/24"
  map_public_ip_on_launch = "True"
  #This needs to be false eventually
  tags {
    Name                  = "${var.cluster_name}-subnet-private"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/subnet       = "private"
    kubernetes.io/cluster = "${var.cluster_name}"
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
    Name                  = "${var.cluster_name}-securityGroup-public"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/securityGroup = "public"
    kubernetes.io/cluster = "${var.cluster_name}"
  }
}

resource "aws_instance" "master" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.master_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  tags {
    Name                  = "${var.cluster_name}-node-master"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/nodeRoles    = "master"
    kubernetes.io/cluster = "${var.cluster_name}"
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "2m"
    }
  }
}

resource "aws_instance" "etcd" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.etcd_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  tags {
    Name                  = "${var.cluster_name}-node-etcd"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/nodeRoles    = "etcd"
    kubernetes.io/cluster = "${var.cluster_name}"
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "2m"
    }
  }
}

resource "aws_instance" "worker" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.worker_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  tags {
    Name                  = "${var.cluster_name}-node-worker"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/nodeRoles    = "worker"
    kubernetes.io/cluster = "${var.cluster_name}"
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "2m"
    }
  }
}

resource "aws_instance" "ingress" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.ingress_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  tags {
    Name                  = "${var.cluster_name}-node-ingress"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/nodeRoles    = "ingress"
    kubernetes.io/cluster = "${var.cluster_name}"
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "2m"
    }
  }
}

resource "aws_instance" "storage" {
  vpc_security_group_ids  = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id               = "${aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.storage_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  tags {
    Name                  = "${var.cluster_name}-node-storage"
    kismatic/clusterName  = "${var.cluster_name}"
    kismatic/clusterOwner = "${var.cluster_owner}"
    kismatic/timestamp    = "${timestamp()}"
    kismatic/version      = "${var.version}"
    kismatic/nodeRoles    = "storage"
    kubernetes.io/cluster = "${var.cluster_name}"
  }

  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "2m"
    }
  }
}