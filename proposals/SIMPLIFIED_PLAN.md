# Simplify Required Info For A Nodes
Status: Proposal

## Motivation
Currently KET requires 3 pieces of information for a node:
```
nodes:
- host: etcd001
  ip: 192.168.42.2
  internalIP: 192.168.42.2
```
`host`: the hostname of the node, must equal to what the `hostname` command and `ansible_nodename` return  
`ip`: the IP used to SSH into the node  
`internalIP`: (optional) the IP used for the cluster components. If not provided will equal to `ip`

With different infrastructure not all of this information is available, ie in AWS the public IP is not stable and can change on a node restart.  
This is also a lot of information for a node, and in many configurations the same cluster can be set up with just a single `address` field.  

This new `address` field would replace all the fields above. An optional `ssh_address` can be used by Ansible to establish a connection(if not provided will default to `address`).  
Both the `address` and `ssh_address` can be an _IP_ or _FQDN_

### FQDN
When `address` is an FQDN:
* the address needs to be resolvable from the other nodes in the cluster
* the `kube-apiserver` flag `--advertise-address` will not be set and the IP will be auto-detected

### IP
When `address` is an IP:
* Ansible requires a unique name for each node and the _hostname_ seems like a sane choice
  * a new process will collect the _hostnames_ of all nodes prior to installation, this will also minimize the required changes in the playbooks
* If the _hostnames_ are not resolvable, it is still possible to use the `networking.update_hosts_files: true` option, KET will update the `/etc/hosts` files with the required entries

## Upgrade considerations
This would be a fairly intrusive change and I do not see a clear upgrade path from clusters/plan files with the old interface moving to the new interface.  
The old plan file field and the new fields will need to coexist. For any new clusters the generated file will only contain the new `address` and `ssh_address` fields. For backwards compatibility KET will also support reading old plan files.
