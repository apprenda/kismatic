provider "aws" {
  /*
  $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY_ID
  $ export AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_ACCESS_KEY
  $ export AWS_DEFAULT_REGION=us-east-1
  */
  region      = "${var.region}"
  access_key  = "${var.access_key}"
  secret_key  = "${var.secret_key}"
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
  //TODO: disable when we add bastion support
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

resource "aws_subnet" "kismatic_master" {
  count       = "${var.master_count > 1 ? 1 : 0}"
  //Conditionally enable this for the load balancer.
  //Only needed if we have more than a single master, else just use public.
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.3.0/24"
  map_public_ip_on_launch = "True"
  //TODO: disable when we add bastion support
  tags {
    "Name"                  = "${var.cluster_name}-subnet-master"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/subnet"       = "master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_subnet" "kismatic_ingress" {
  count       = "${var.ingress_count > 1 ? 1 : 0}"
  //Same here, but for ingress.
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.4.0/24"
  map_public_ip_on_launch = "True"
  //TODO: disable when we add bastion support
  tags {
    "Name"                  = "${var.cluster_name}-subnet-ingress"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/subnet"       = "ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_security_group" "kismatic_ssh" {
  name        = "${var.cluster_name}-ssh"
  description = "Allow inbound SSH for kismatic."
  vpc_id      = "${aws_vpc.kismatic.id}"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-ssh"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/securityGroup"= "ssh"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_security_group" "kismatic_private" {
  name        = "${var.cluster_name}-private"
  description = "Allow all communication between nodes."
  vpc_id      = "${aws_vpc.kismatic.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = "True"
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }

  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-private"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/securityGroup"= "private"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}


resource "aws_security_group" "kismatic_lb_master" {
  name        = "${var.cluster_name}-lb-master"
  description = "Allow inbound on 6443 for kube-apiserver load balancer."
  vpc_id      = "${aws_vpc.kismatic.id}"

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-master"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/securityGroup"= "lb-master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_security_group" "kismatic_lb_ingress" {
  name        = "${var.cluster_name}-lb-ingress"
  description = "Allow inbound on 80 and 443 for ingress load balancer."
  vpc_id      = "${aws_vpc.kismatic.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-ingress"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/securityGroup"= "lb-ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_s3_bucket" "lb_logs" {
  count  = 0
  //"${var.master_count > 1 || var.ingress_count > 1 ? 1 : 0}"
  //Conditionally enable if either LB is active.
  bucket = "${var.cluster_name}-lb_logs"
  acl    = "log-delivery-write"

  tags {
    "Name"                  = "${var.cluster_name}-bucket-lb"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/bucket"       = "lb"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_elb" "kismatic_master" {
  count           = "${var.master_count > 1 ? 1 : 0}"
  //Again, only necessary in the multi-master HA cases
  name            = "${var.cluster_name}-lb-master"
  internal        = false
  security_groups = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_lb_master.id}"]
  subnets         = ["${aws_subnet.kismatic_public.id}"]

  //access_logs {
  //  bucket = "${aws_s3_bucket.lb_logs.bucket}"
  //  bucket_prefix = "${var.cluster_name}/master"
  //}

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  instances = ["${aws_instance.master.*.id}"]

  tags {
    "Name"                  = "${var.cluster_name}-lb-master"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/loadBalancer" = "master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_elb" "kismatic_ingress" {
  count           = "${var.ingress_count > 1 ? 1 : 0}"
  name            = "${var.cluster_name}-lb-ingress"
  internal        = false
  security_groups = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_lb_ingress.id}"]
  subnets         = ["${aws_subnet.kismatic_public.id}"]

  //access_logs {
  //  bucket = "${aws_s3_bucket.lb_logs.bucket}"
  //  bucket_prefix = "${var.cluster_name}/ingress"
  //}

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 443
    lb_protocol       = "tcp"
  } 

  listener {
    instance_port     = 8080
    instance_protocol = "tcp"
    lb_port           = 80
    lb_protocol       = "tcp"
  }

  instances = ["${aws_instance.ingress.*.id}"]

  tags {
    "Name"                  = "${var.cluster_name}-lb-ingress"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/loadBalancer" = "ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_instance" "bastion" {
  security_groups        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  subnet_id              = "${aws_subnet.kismatic_public.id}"
  key_name               = "${var.cluster_name}"
  count                  = 0
  // TODO: setup a bastion node for added security
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nodeRoles"    = "bastion"
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
      timeout = "2m"
    }
  }
}

resource "aws_instance" "master" {
  security_groups        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id              = "${var.master_count > 1 ? aws_subnet.kismatic_master.id : aws_subnet.kismatic_public.id}"
  key_name               = "${var.cluster_name}"
  count                  = "${var.master_count}"
  ami                    = "${data.aws_ami.ubuntu.id}"
  instance_type          = "${var.instance_size}"
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
      timeout = "2m"
    }
  }
}

resource "aws_instance" "etcd" {
  security_groups        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.etcd_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
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
      timeout = "2m"
    }
  }
}

resource "aws_instance" "worker" {
  security_groups         = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.worker_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
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
      timeout = "2m"
    }
  }
}

resource "aws_instance" "ingress" {
  security_groups         = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${var.ingress_count > 1 ? aws_subnet.kismatic_ingress.id : aws_subnet.kismatic_public.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.ingress_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
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
      timeout = "2m"
    }
  }
}

resource "aws_instance" "storage" {
  security_groups        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.storage_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
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
      timeout = "2m"
    }
  }
}