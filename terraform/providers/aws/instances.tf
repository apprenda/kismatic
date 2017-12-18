resource "aws_instance" "bastion" {
  vpc_security_group_ids        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  subnet_id              = "${aws_subnet.kismatic_public.id}"
  key_name               = "${var.cluster_name}"
  count                  = 0
  // TODO: setup a bastion node for added security
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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
  vpc_security_group_ids        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id              = "${aws_subnet.kismatic_master.id}"
  key_name               = "${var.cluster_name}"
  count                  = "${var.master_count}"
  ami                    = "${data.aws_ami.ubuntu.id}"
  instance_type          = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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
  vpc_security_group_ids        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.etcd_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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
  vpc_security_group_ids         = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.worker_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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
  vpc_security_group_ids         = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_ingress.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.ingress_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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
  vpc_security_group_ids        = ["${aws_security_group.kismatic_private.id}", "${aws_security_group.kismatic_ssh.id}"]
  // TODO: remove from public when bastion is set up
  subnet_id               = "${aws_subnet.kismatic_private.id}"
  key_name                = "${var.cluster_name}"
  count                   = "${var.storage_count}"
  ami                     = "${data.aws_ami.ubuntu.id}"
  instance_type           = "${var.instance_size}"
  availability_zone = "${var.AZ}"
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