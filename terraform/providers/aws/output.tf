output "etcd_pub_ips" {
  value = ["${aws_instance.etcd.*.public_ip}"]
}

output "master_pub_ips" {
  value = ["${aws_instance.master.*.public_ip}"]
}

output "worker_pub_ips" {
  value = ["${aws_instance.worker.*.public_ip}"]
}

output "ingress_pub_ips" {
  value = ["${aws_instance.ingress.*.public_ip}"]
}

output "storage_pub_ips" {
  value = ["${aws_instance.storage.*.public_ip}"]
}

output "etcd_priv_ips" {
  value = ["${aws_instance.etcd.*.private_ip}"]
}

output "master_priv_ips" {
  value = ["${aws_instance.master.*.private_ip}"]
}

output "worker_priv_ips" {
  value = ["${aws_instance.worker.*.private_ip}"]
}

output "ingress_priv_ips" {
  value = ["${aws_instance.ingress.*.private_ip}"]
}

output "storage_priv_ips" {
  value = ["${aws_instance.storage.*.private_ip}"]
}

output "etcd_hosts" {
  value = ["${aws_instance.etcd.*.private_dns}"]
}

output "master_hosts" {
  value = ["${aws_instance.master.*.private_dns}"]
}

output "worker_hosts" {
  value = ["${aws_instance.worker.*.private_dns}"]
}

output "ingress_hosts" {
  value = ["${aws_instance.ingress.*.private_dns}"]
}

output "storage_hosts" {
  value = ["${aws_instance.storage.*.private_dns}"]
}

output "master_lb" {
  value = ["${aws_elb.kismatic_master.*.dns_name}"]
}

output "ingress_lb" {
  value = ["${aws_elb.kismatic_ingress.*.dns_name}"]
}