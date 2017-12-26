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
    "kismatic/version"      = "${var.kismatic_version}"
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
    "kismatic/version"      = "${var.kismatic_version}"
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
    "kismatic/version"      = "${var.kismatic_version}"
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
    "kismatic/version"      = "${var.kismatic_version}"
    "kismatic/securityGroup"= "lb-ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}
