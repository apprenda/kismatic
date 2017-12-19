resource "azurerm_network_security_group" "kismatic_private" {
  depends_on            = ["azurerm_resource_group.kismatic"]  
  name                = "${var.cluster_name}-private"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  security_rule {
    name                       = "${var.cluster_name}-ssh"
    description                = "Allow inbound SSH for kismatic."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "22"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-private-in"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    // The address space of the resource group
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-private-out"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-private"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.securityGroup"= "private"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_security_group" "kismatic_lb_master" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  name                = "${var.cluster_name}-lb-master"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"

  security_rule {
    name                       = "${var.cluster_name}-lb-master"
    description                = "Allow inbound on 6443 for kube-apiserver load balancer."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "6443"
    destination_port_range     = "6443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-private-in"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    // The address space of the resource group
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-private-out"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-master"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.securityGroup"= "lb-master"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_security_group" "kismatic_lb_ingress" {
  depends_on            = ["azurerm_resource_group.kismatic"]
  name                = "${var.cluster_name}-lb-ingress"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"

  security_rule {
    name                       = "${var.cluster_name}-lb-ingress-80"
    description                = "Allow inbound on 80 for nginx."
    priority                   = 102
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "80"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
    security_rule {
    name                       = "${var.cluster_name}-lb-ingress-443"
    description                = "Allow inbound on 443 for nginx."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "443"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
    security_rule {
    name                       = "${var.cluster_name}-private-in"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    // The address space of the resource group
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-private-out"
    description                = "Allow all communication between nodes."
    priority                   = 100
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-ingress"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.securityGroup"= "lb-ingress"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}