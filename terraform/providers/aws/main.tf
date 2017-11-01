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
  public_key = "${var.public_ssh_key}"
}

resource "aws_vpc" "kismatic" {
  cidr_block            = "10.0.0.0/16"
  enable_dns_support    = true
  enable_dns_hostnames  = true
  tags {
    Name = "kismatic - cluster"
  }
}

resource "aws_internet_gateway" "kismatic_gateway" {
  vpc_id = "${aws_vpc.kismatic.id}"
  tags {
    Name = "kismatic - cluster"
  }
}

resource "aws_default_route_table" "kismatic_router" {
  default_route_table_id = "${aws_vpc.kismatic.default_route_table_id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.kismatic_gateway.id}"
  }

  tags {
    Name = "kismatic - cluster"
  }
}


resource "aws_subnet" "kismatic_public" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.1.0/24"
  map_public_ip_on_launch = "True"
}

resource "aws_subnet" "kismatic_private" {
  vpc_id      = "${aws_vpc.kismatic.id}"
  cidr_block  = "10.0.2.0/24"
  map_public_ip_on_launch = "True"
  #This needs to be false eventually
}

resource "aws_security_group" "kismatic_sec_group" {
  name        = "kismatic - cluster"
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
    Name = "kismatic"
  }
}

resource "aws_instance" "master" {
  vpc_security_group_ids = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id       = "${aws_subnet.kismatic_public.id}"
  key_name        = "${var.cluster_name}"
  count           = "${var.master_count}"
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    Name = "kismatic - master"
  }
}

resource "aws_instance" "etcd" {
  vpc_security_group_ids = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id       = "${aws_subnet.kismatic_public.id}"
  key_name        = "${var.cluster_name}"
  count           = "${var.etcd_count}"
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    Name = "kismatic - etcd"
  }
}

resource "aws_instance" "worker" {
  vpc_security_group_ids = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id       = "${aws_subnet.kismatic_public.id}"
  key_name        = "${var.cluster_name}"
  count           = "${var.worker_count}"
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    Name = "kismatic - worker"
  }
}

resource "aws_instance" "ingress" {
  vpc_security_group_ids = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id       = "${aws_subnet.kismatic_public.id}"
  key_name        = "${var.cluster_name}"
  count           = "${var.ingress_count}"
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    Name = "kismatic - ingress"
  }
}

resource "aws_instance" "storage" {
  vpc_security_group_ids = ["${aws_security_group.kismatic_sec_group.id}"]
  subnet_id       = "${aws_subnet.kismatic_public.id}"
  key_name        = "${var.cluster_name}"
  count           = "${var.storage_count}"
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  tags {
    Name = "kismatic - storage"
  }
}


# data "template_file" "kismatic_cluster" {
#   template = "${file("${path.module}/../../clusters/dev/kismatic-cluster.yaml.tpl")}"
#   vars {
#     etcd_pub_ip = "${aws_instance.etcd.0.public_ip}"
#     master_pub_ip = "${aws_instance.master.0.public_ip}"
#     worker_pub_ip = "${aws_instance.worker.0.public_ip}"
#     ingress_pub_ip = "${aws_instance.ingress.0.public_ip}"

#     etcd_priv_ip = "${aws_instance.etcd.0.private_ip}"
#     master_priv_ip = "${aws_instance.master.0.private_ip}"
#     worker_priv_ip = "${aws_instance.worker.0.private_ip}"
#     ingress_priv_ip = "${aws_instance.ingress.0.private_ip}"
       
#     etcd_host = "${aws_instance.etcd.0.private_dns}"
#     master_host = "${aws_instance.master.0.private_dns}"
#     worker_host = "${aws_instance.worker.0.private_dns}"
#     ingress_host = "${aws_instance.ingress.0.private_dns}"

#     ssh_key = "${var.private_ssh_key_path}"
#   }
# }


