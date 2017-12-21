resource "azurerm_lb" "kismatic_master" {
  
  name                         = "${var.cluster_name}-lb-master"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"

  frontend_ip_configuration {
    name                          = "${var.cluster_name}-lb-master"
    public_ip_address_id          = "${azurerm_public_ip.lb_master.id}"
    private_ip_address_allocation = "dynamic"
  }
  tags {
    "Name"                  = "${var.cluster_name}-lb-master"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.loadBalancer" = "master"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_lb_rule" "kismatic_master" {
  
  name                    = "${var.cluster_name}-lb-master"
  resource_group_name     = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id         = "${azurerm_lb.kismatic_master.id}"
  backend_address_pool_id = "${azurerm_lb_backend_address_pool.kismatic_master.id}"
  probe_id                = "${azurerm_lb_probe.kismatic_master.id}"

  protocol                       = "tcp"
  frontend_port                  = 6443
  backend_port                   = 6443
  frontend_ip_configuration_name = "${var.cluster_name}-lb-master"
}

resource "azurerm_lb_probe" "kismatic_master" {
  
  name                = "${var.cluster_name}-lb-master"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id     = "${azurerm_lb.kismatic_master.id}"
  protocol            = "tcp"
  port                = 6443
}

resource "azurerm_lb_backend_address_pool" "kismatic_master" {
  
  name                = "${var.cluster_name}-lb-master"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id     = "${azurerm_lb.kismatic_master.id}"
}

resource "azurerm_lb" "kismatic_ingress" {
  
  name                         = "${var.cluster_name}-lb-ingress"
  location                     = "${azurerm_resource_group.kismatic.location}"
  resource_group_name          = "${azurerm_resource_group.kismatic.name}"

  frontend_ip_configuration {
    name                          = "${var.cluster_name}-lb-ingress"
    public_ip_address_id          = "${azurerm_public_ip.lb_ingress.id}"
    private_ip_address_allocation = "dynamic"
  }
  tags {
    "Name"                  = "${var.cluster_name}-lb-ingress"
    "kismatic.clusterName"  = "${var.cluster_name}"
    "kismatic.clusterOwner" = "${var.cluster_owner}"
    "kismatic.dateCreated"  = "${timestamp()}"
    "kismatic.version"      = "${var.kismatic_version}"
    "kismatic.loadBalancer" = "ingress"
    
  }
  lifecycle {
    ignore_changes = ["tags.kismatic.dateCreated", "tags.Owner", "tags.PrincipalID"]
  }
}

resource "azurerm_lb_rule" "kismatic_ingress_443" {
  
  name                    = "${var.cluster_name}-lb-ingress"
  resource_group_name     = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id         = "${azurerm_lb.kismatic_ingress.id}"
  backend_address_pool_id = "${azurerm_lb_backend_address_pool.kismatic_ingress.id}"
  probe_id                = "${azurerm_lb_probe.kismatic_ingress.id}"

  protocol                       = "tcp"
  frontend_port                  = 443
  backend_port                   = 443
  frontend_ip_configuration_name = "${var.cluster_name}-lb-ingress"
}

resource "azurerm_lb_rule" "kismatic_ingress_80" {
  
  name                    = "${var.cluster_name}-lb-ingress"
  resource_group_name     = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id         = "${azurerm_lb.kismatic_ingress.id}"
  backend_address_pool_id = "${azurerm_lb_backend_address_pool.kismatic_ingress.id}"
  probe_id                = "${azurerm_lb_probe.kismatic_ingress.id}"

  protocol                       = "tcp"
  frontend_port                  = 80
  backend_port                   = 80
  frontend_ip_configuration_name = "${var.cluster_name}-lb-ingress"
}

resource "azurerm_lb_probe" "kismatic_ingress" {
  
  name                = "${var.cluster_name}-lb-ingress"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id     = "${azurerm_lb.kismatic_ingress.id}"
  protocol            = "tcp"
  port                = 443
}

resource "azurerm_lb_backend_address_pool" "kismatic_ingress" {
  
  name                = "${var.cluster_name}-lb-ingress"
  resource_group_name = "${azurerm_resource_group.kismatic.name}"
  loadbalancer_id     = "${azurerm_lb.kismatic_ingress.id}"
}