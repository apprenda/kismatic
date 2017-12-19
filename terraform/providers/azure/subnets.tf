resource "azurerm_subnet" "kismatic_public" {
  depends_on            = ["azurerm_resource_group.kismatic", "azurerm_virtual_network.kismatic", "azurerm_route_table.kismatic"]
  name                      = "${var.cluster_name}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  virtual_network_name      = "${azurerm_virtual_network.kismatic.name}"
  address_prefix            = "10.0.1.0/24"
  route_table_id            = "${azurerm_route_table.kismatic.id}"
}

resource "azurerm_subnet" "kismatic_private" {
  depends_on            = ["azurerm_resource_group.kismatic", "azurerm_virtual_network.kismatic", "azurerm_route_table.kismatic"]
  name                      = "${var.cluster_name}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  virtual_network_name      = "${azurerm_virtual_network.kismatic.name}"
  address_prefix            = "10.0.2.0/24"
  route_table_id            = "${azurerm_route_table.kismatic.id}"
}

resource "azurerm_subnet" "kismatic_master" {
  depends_on            = ["azurerm_resource_group.kismatic", "azurerm_virtual_network.kismatic", "azurerm_route_table.kismatic"]
  name                      = "${var.cluster_name}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  virtual_network_name      = "${azurerm_virtual_network.kismatic.name}"
  address_prefix            = "10.0.3.0/24"
  route_table_id            = "${azurerm_route_table.kismatic.id}"
}

resource "azurerm_subnet" "kismatic_ingress" {
  depends_on            = ["azurerm_resource_group.kismatic", "azurerm_virtual_network.kismatic", "azurerm_route_table.kismatic"]
  name                      = "${var.cluster_name}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  virtual_network_name      = "${azurerm_virtual_network.kismatic.name}"
  address_prefix            = "10.0.4.0/24"
  route_table_id            = "${azurerm_route_table.kismatic.id}"
}