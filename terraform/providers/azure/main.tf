provider "azurerm" {
  /*
  $ export ARM_SUBSCRIPTION_ID = your_azure_sub_id
  $ export ARM_CLIENT_ID = your_client_id
  $ export ARM_CLIENT_SECRET = your_client_secret
  $ export ARM_TENANT_ID= your_tenant_id
  */
  subscription_id = "${var.sub_id}"
  client_id       = "${var.client_id}"
  client_secret   = "${var.client_secret}"
  tenant_id       = "${var.tenant_id}"
}



