resource "azurerm_network_interface" "bastion" {
  internal_dns_name_label   = "${var.cluster_name}-bastion"
  count                     = 0
  name                      = "${var.cluster_name}-bastion-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-bastion-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_private.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.bastion.*.id, count.index)}"
  }
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

resource "azurerm_network_interface" "master" {
  internal_dns_name_label   = "${var.cluster_name}-master-${count.index}"
  count                     = "${var.master_count}"
  name                      = "${var.cluster_name}-master-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_master.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-master-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_master.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.master.*.id, count.index)}"
    load_balancer_backend_address_pools_ids = ["${azurerm_lb_backend_address_pool.kismatic_master.id}"]
  }
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

resource "azurerm_network_interface" "etcd" {
  internal_dns_name_label   = "${var.cluster_name}-etcd-${count.index}"
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "etcd"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "worker" {
  internal_dns_name_label   = "${var.cluster_name}-worker-${count.index}"
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "worker"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_network_interface" "ingress" {
  internal_dns_name_label   = "${var.cluster_name}-ingress-${count.index}"
  count                     = "${var.ingress_count}"
  name                      = "${var.cluster_name}-ingress-${count.index}"
  location                  = "${azurerm_resource_group.kismatic.location}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_ingress.id}"

  ip_configuration {
    name                          = "${var.cluster_name}-ingress-${count.index}"
    subnet_id                     = "${azurerm_subnet.kismatic_ingress.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.ingress.*.id, count.index)}"
    load_balancer_backend_address_pools_ids = ["${azurerm_lb_backend_address_pool.kismatic_ingress.id}"]
  }
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

resource "azurerm_network_interface" "storage" {
  internal_dns_name_label   = "${var.cluster_name}-storage-${count.index}"
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
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nic"          = "storage"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}