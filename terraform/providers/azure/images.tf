data "azurerm_platform_image" "ubuntu" {
  location  = "${azurerm_resource_group.kismatic.location}"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}
