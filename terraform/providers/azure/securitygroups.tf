resource "azurerm_network_security_group" "kismatic_private" {
  
  name                = "${var.cluster_name}-private"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  security_rule {
    name                       = "${var.cluster_name}-ssh"
    description                = "Allow inbound SSH for kismatic."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-in"
    description                = "Allow inbound traffic from all other nodes."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-private"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.securityGroup"= "private"
     
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_security_group" "kismatic_master" {
  
  name                = "${var.cluster_name}-master"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  security_rule {
    name                       = "${var.cluster_name}-ssh"
    description                = "Allow inbound SSH for kismatic."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-in"
    description                = "Allow inbound traffic from all other nodes."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-lb-master"
    description                = "Allow inbound on 6443 for kube-apiserver load balancer."
    priority                   = 102
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "*"
    destination_port_range     = "6443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-master"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.securityGroup"= "lb-master"
     
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_security_group" "kismatic_ingress" {
  
  name                = "${var.cluster_name}-ingress"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  security_rule {
    name                       = "${var.cluster_name}-ssh"
    description                = "Allow inbound SSH for kismatic."
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-in"
    description                = "Allow inbound traffic from all other nodes."
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "10.0.0.0/16"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "${var.cluster_name}-lb-ingress-80"
    description                = "Allow inbound on 80 for nginx."
    priority                   = 103
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
    security_rule {
    name                       = "${var.cluster_name}-lb-ingress-443"
    description                = "Allow inbound on 443 for nginx."
    priority                   = 102
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  tags {
    "Name"                  = "${var.cluster_name}-securityGroup-lb-ingress"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.securityGroup"= "lb-ingress"
     
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}