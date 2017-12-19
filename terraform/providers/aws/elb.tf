resource "aws_s3_bucket" "lb_logs" {
  count  = 0
  bucket = "${var.cluster_name}-lb_logs"
  acl    = "log-delivery-write"

  tags {
    "Name"                  = "${var.cluster_name}-bucket-lb"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.kismatic_version}"
    "kismatic/bucket"       = "lb"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_elb" "kismatic_master" {
  name            = "${var.cluster_name}-lb-master"
  internal        = false
  security_groups = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_lb_master.id}"]
  subnets         = ["${aws_subnet.kismatic_public.id}"]
  connection_draining = "True"
  cross_zone_load_balancing = "True"
  //access_logs {
  //  bucket = "${aws_s3_bucket.lb_logs.bucket}"
  //  bucket_prefix = "${var.cluster_name}/master"
  //}
  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:6443"
    interval            = 10
  }
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
    "kismatic/version"      = "${var.kismatic_version}"
    "kismatic/loadBalancer" = "master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "aws_elb" "kismatic_ingress" {
  name            = "${var.cluster_name}-lb-ingress"
  internal        = false
  security_groups = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_lb_ingress.id}"]
  subnets         = ["${aws_subnet.kismatic_public.id}"]
  connection_draining = "True"
  cross_zone_load_balancing = "True"
  //access_logs {
  //  bucket = "${aws_s3_bucket.lb_logs.bucket}"
  //  bucket_prefix = "${var.cluster_name}/ingress"
  //}

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:443"
    interval            = 10
  }

  listener {
    instance_port     = 443
    instance_protocol = "tcp"
    lb_port           = 443
    lb_protocol       = "tcp"
  } 

  listener {
    instance_port     = 80
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
    "kismatic/version"      = "${var.kismatic_version}"
    "kismatic/loadBalancer" = "ingress"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}
