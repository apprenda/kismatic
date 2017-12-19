resource "azurerm_network_interface" "bastion" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_public", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.bastion"] 
  count                     = 0
  name                      = "${var.cluster_name}-bastion-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-bastion-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_public.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.bastion.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "master" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_master", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.master"] 
  count                     = "${var.master_count}"
  name                      = "${var.cluster_name}-master-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-master-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_master.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.master.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "etcd" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_private", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.etcd"] 
  count                     = "${var.etcd_count}"
  name                      = "${var.cluster_name}-etcd-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-etcd-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.etcd.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "worker" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_private", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.worker"] 
  count                     = "${var.worker_count}"
  name                      = "${var.cluster_name}-worker-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-worker-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.worker.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "ingress" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_private", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.ingress"] 
  count                     = "${var.ingress_count}"
  name                      = "${var.cluster_name}-ingress-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-ingress-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.ingress.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "storage" {
  depends_on                = ["azurerm_resource_group.kismatic", "azurerm_subnet.kismatic_private", "azurerm_network_security_group.kismatic_private", "azurerm_public_ip.storage"] 
  count                     = "${var.storage_count}"
  name                      = "${var.cluster_name}-storage-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-storage-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.storage.*.id, count.index)}"
  }
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