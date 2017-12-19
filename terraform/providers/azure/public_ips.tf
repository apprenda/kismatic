resource "azurerm_public_ip" "bastion" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = 0 
  name                         = "${var.cluster_name}-bastion-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "bastion"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "master" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = "${var.master_count}"
  name                         = "${var.cluster_name}-master-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-master-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "master"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}


resource "azurerm_public_ip" "etcd" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = "${var.etcd_count}"
  name                         = "${var.cluster_name}-etcd-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-etcd-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "etcd"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "worker" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = "${var.worker_count}"
  name                         = "${var.cluster_name}-worker-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-worker-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "worker"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "ingress" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = "${var.ingress_count}"
  name                         = "${var.cluster_name}-ingress-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-ingress-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "ingress"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "storage" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  count                        = "${var.storage_count}"
  name                         = "${var.cluster_name}-storage-${count.index}"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-storage-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "storage"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "lb_master" {
  name                         = "${var.cluster_name}-lb-master"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-lb-master"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "lb-master"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "lb_ingress" {
  name                         = "${var.cluster_name}-lb-ingress"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"
  public_ip_address_allocation = "static"

  tags {
    "Name"                  = "${var.cluster_name}-lb-ingress"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nic"          = "lb-ingress"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}