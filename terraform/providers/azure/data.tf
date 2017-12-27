data "azurerm_public_ip" "bastion" {
  count               = 0
  name                = "${azurerm_public_ip.bastion.name}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.bastion"]
}

data "azurerm_public_ip" "master" {
  count               = "${var.master_count}"
  name                = "${element(azurerm_public_ip.master.*.name,count.index)}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.master"]
}

data "azurerm_public_ip" "etcd" {
  count               = "${var.etcd_count}"
  name                = "${element(azurerm_public_ip.etcd.*.name,count.index)}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.etcd"]
}

data "azurerm_public_ip" "worker" {
  count               = "${var.worker_count}"
  name                = "${element(azurerm_public_ip.worker.*.name,count.index)}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.worker"]
}

data "azurerm_public_ip" "ingress" {
  count               = "${var.ingress_count}"
  name                = "${element(azurerm_public_ip.ingress.*.name,count.index)}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.ingress"]
}

data "azurerm_public_ip" "storage" {
  count               = "${var.storage_count}"
  name                = "${element(azurerm_public_ip.storage.*.name,count.index)}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  depends_on          = ["azurerm_public_ip.storage"]
}