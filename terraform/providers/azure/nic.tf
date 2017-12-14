resource "azurerm_network_interface" "bastion" {
  count                     = "${var.master_count}"
  name                      = "${var.cluster_name}-bastion-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-bastion-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_public.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.bastion.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "bastion"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "master" {
  count                     = "${var.master_count}"
  name                      = "${var.cluster_name}-master-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-master-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_master.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.master.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-master-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "master"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "etcd" {
  name                      = "${var.cluster_name}-etcd-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-etcd-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.bastion.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-etcd-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "etcd"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "worker" {
  name                      = "bastion"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "bastion"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.bastion.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "bastion"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "ingress" {
  name                      = "bastion"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "bastion"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.bastion.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "bastion"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "storage" {
  name                      = "bastion"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "bastion"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.bastion.id}"
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic/clusterName"  = "${var.cluster_name}"
    "kismatic/clusterOwner" = "${var.cluster_owner}"
    "kismatic/dateCreated"  = "${timestamp()}"
    "kismatic/version"      = "${var.version}"
    "kismatic/nic"          = "bastion"
    "kubernetes.io/cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic/dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}