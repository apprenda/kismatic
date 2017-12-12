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



resource "azurerm_resource_group" "kismatic" {
  name     = "${var.cluster_name}"
  location = "East US"
}

resource "azurerm_virtual_network" "kubenet" {
  name                = "kubenet"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurerm_resource_group.ket.location}"
  resource_group_name = "${azurerm_resource_group.ket.name}"
}

# Required for cloud provider integration
resource "azurerm_route_table" "ket" {
  name                = "kubernetes"
  location            = "${azurerm_resource_group.ket.location}"
  resource_group_name = "${azurerm_resource_group.ket.name}"
}

resource "azurerm_subnet" "kubenodes" {
  name                      = "kubenodes"
  resource_group_name       = "${azurerm_resource_group.ket.name}"
  virtual_network_name      = "${azurerm_virtual_network.kubenet.name}"
  address_prefix            = "${var.nodes_subnet}"
  route_table_id            = "${azurerm_route_table.ket.id}"
}