resource "azurerm_public_ip" "bastion" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "bastion"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "master" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "master"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}


resource "azurerm_public_ip" "etcd" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "etcd"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "worker" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "worker"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "ingress" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "ingress"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_public_ip" "storage" {
  
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "storage"
    
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
  domain_name_label            = "${var.cluster_name}-master"

  tags {
    "Name"                  = "${var.cluster_name}-lb-master"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "lb-master"
    
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
  domain_name_label            = "${var.cluster_name}-ingress"

  tags {
    "Name"                  = "${var.cluster_name}-lb-ingress"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "lb-ingress"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}