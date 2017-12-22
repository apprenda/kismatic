resource "azurerm_subnet" "kismatic" {
  
  name                      = "${var.cluster_name}"
  resource_group_name       = "${azurerm_resource_group.kismatic.name}"
  virtual_network_name      = "${azurerm_virtual_network.kismatic.name}"
  address_prefix            = "10.0.1.0/24"
  route_table_id            = "${azurerm_route_table.kismatic.id}"
  network_security_group_id = "${azurerm_network_security_group.kismatic_private.id}"
}
