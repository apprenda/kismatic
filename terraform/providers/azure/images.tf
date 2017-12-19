data "azurerm_platform_image" "ubuntu" {
  location  = "West US"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}
