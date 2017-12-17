resource "azurerm_virtual_machine" "bastion" {
  count                 = 0
  name                  = "${var.cluster_name}-bastion-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${element(azurerm_network_interface.bastion.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-bastion-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-bastion-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "bastion"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_machine" "master" {
  count                 = "${var.master_count}"
  name                  = "${var.cluster_name}-master-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${azurerm_network_interface.master.*.id}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-master-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-master-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-master-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "master"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_machine" "etcd" {
  count                 = "${var.etcd_count}"
  name                  = "${var.cluster_name}-etcd-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${azurerm_network_interface.etcd.*.id}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-etcd-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-etcd-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-etcd-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "etcd"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_machine" "worker" {
  count                 = "${var.worker_count}"
  name                  = "${var.cluster_name}-worker-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${azurerm_network_interface.worker.*.id}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-worker-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-worker-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-worker-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "worker"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_machine" "ingress" {
  count                 = "${var.ingress_count}"
  name                  = "${var.cluster_name}-ingress-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${azurerm_network_interface.ingress.*.id}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-ingress-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-ingress-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-ingress-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "ingress"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_virtual_machine" "storage" {
  count                 = "${var.storage_count}"
  name                  = "${var.cluster_name}-storage-${count.index}"
  location              = "${azurerm_resource_group.kismatic.location}"
  resource_group_name   = "${azurerm_resource_group.kismatic.name}"
  network_interface_ids = ["${azurerm_network_interface.storage.*.id}"]
  vm_size               = "${var.instance_size}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    id = "${data.azurerm_platform_image.ubuntu.id}"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-storage-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    admin_username = "${var.ssh_user}"
    computer_name  = "${var.cluster_name}-storage-${count.index}"
  }

  os_profile_linux_config {
    disable_password_authentication = true
    ssh_keys {
        path     = "/home/${var.ssh_user}/.ssh/authorized_keys"
        key_data = "${file("${var.public_ssh_key_path}")}"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-storage-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.version}"
    "kismatic.nodeRoles"    = "storage"
    "kubernetes.io.cluster" = "${var.cluster_name}"
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}