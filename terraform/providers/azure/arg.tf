resource "azurerm_resource_group" "kismatic" {
  name     = "${var.cluster_name}"
  location = "${var.location}"
  tags {
    "Name"                  = "${var.cluster_name}-rg"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_network" "kismatic" {
  
  name                = "${var.cluster_name}"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  tags {
    "Name"                  = "${var.cluster_name}-vnet"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_route_table" "kismatic" {
  
  name                = "${var.cluster_name}"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  tags {
    "Name"                  = "${var.cluster_name}-router"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

