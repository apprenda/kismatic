output "etcd_pub_ips" {
  value = ["${azurerm_public_ip.etcd.*.ip_address}"]
}

output "master_pub_ips" {
  value = ["${azurerm_public_ip.master.*.ip_address}"]
}

output "worker_pub_ips" {
  value = ["${azurerm_public_ip.worker.*.ip_address}"]
}

output "ingress_pub_ips" {
  value = ["${azurerm_public_ip.ingress.*.ip_address}"]
}

output "storage_pub_ips" {
  value = ["${azurerm_public_ip.storage.*.ip_address}"]
}

output "etcd_priv_ips" {
  value = ["${azurerm_network_interface.etcd.*.private_ip_address}"]
}

output "master_priv_ips" {
  value = ["${azurerm_network_interface.master.*.private_ip_address}"]
}

output "worker_priv_ips" {
  value = ["${azurerm_network_interface.worker.*.private_ip_address}"]
}

output "ingress_priv_ips" {
  value = ["${azurerm_network_interface.ingress.*.private_ip_address}"]
}

output "storage_priv_ips" {
  value = ["${azurerm_network_interface.storage.*.private_ip_address}"]
}

output "etcd_hosts" {
  value = ["${azurerm_network_interface.etcd.*.internal_dns_name_label}"]
}

output "master_hosts" {
  value = ["${azurerm_network_interface.master.*.internal_dns_name_label}"]
}

output "worker_hosts" {
  value = ["${azurerm_network_interface.worker.*.internal_dns_name_label}"]
}

output "ingress_hosts" {
  value = ["${azurerm_network_interface.ingress.*.internal_dns_name_label}"]
}

output "storage_hosts" {
  value = ["${azurerm_network_interface.storage.*.internal_dns_name_label}"]
}

output "master_lb" {
  value = ["${azurerm_public_ip.lb_master.*.fqdn}"]
}

output "ingress_lb" {
  value = ["${azurerm_public_ip.lb_ingress.*.fqdn}"]
}