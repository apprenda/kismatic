resource "azurerm_availability_set" "master" {
  name                = "master"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  managed             = "true"
}

resource "azurerm_availability_set" "etcd" {
  name                = "etcd"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  managed             = "true"
}

resource "azurerm_availability_set" "worker" {
  name                = "worker"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  managed             = "true"
}

resource "azurerm_availability_set" "ingress" {
  name                = "ingress"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  managed             = "true"
}

resource "azurerm_availability_set" "storage" {
  name                = "storage"
  location            = "${azurerm_resource_group.kismatic.location}"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  managed             = "true"
}