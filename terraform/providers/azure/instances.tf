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
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.bastion.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-bastion-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "bastion"
    
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
  network_interface_ids = ["${element(azurerm_network_interface.master.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"
  availability_set_id   = "${azurerm_availability_set.master.id}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.master.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-master-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "master"
    
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
  network_interface_ids = ["${element(azurerm_network_interface.etcd.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"
  availability_set_id   = "${azurerm_availability_set.etcd.id}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.etcd.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-etcd-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "etcd"
    
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
  network_interface_ids = ["${element(azurerm_network_interface.worker.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"
  availability_set_id   = "${azurerm_availability_set.worker.id}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.worker.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-worker-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "worker"
    
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
  network_interface_ids = ["${element(azurerm_network_interface.ingress.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"
  availability_set_id   = "${azurerm_availability_set.ingress.id}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.ingress.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-ingress-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "ingress"
    
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
  network_interface_ids = ["${element(azurerm_network_interface.storage.*.id, count.index)}"]
  vm_size               = "${var.instance_size}"
  availability_set_id   = "${azurerm_availability_set.storage.id}"

  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
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
  provisioner "remote-exec" {
    inline = ["echo ready"]

    connection {
      host = "${element(azurerm_public_ip.storage.*.ip_address, count.index)}"
      type = "ssh"
      user = "${var.ssh_user}"
      private_key = "${file("${var.private_ssh_key_path}")}"
      timeout = "5m"
    }
  }
  tags {
    "Name"                  = "${var.cluster_name}-storage-${count.index}"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.nodeRoles"    = "storage"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}