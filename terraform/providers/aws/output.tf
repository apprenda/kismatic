# Used for rendering 

# output "master_node_ips" {
#   value = "${join(",",aws_instance.master.*.ipv4_address)}"
# }

# output "etcd_node_ips" {
#   value = "${join(",",aws_instance.etcd.*.ipv4_address)}"
# }

# output "worker_node_ips" {
#   value = "${join(",",aws_instance.worker.*.ipv4_address)}"
# }

# output "ingress_node_ips" {
#   value = "${join(",",aws_instance.ingress.*.ipv4_address)}"
# }

# output "storage_node_ips" {
#   value = "${join(",",aws_instance.storage.*.ipv4_address)}"
# }

# output "rendered_template" {
#     value = "${data.template_file.kismatic_cluster.rendered}"
# }

# Used for rendering 

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
