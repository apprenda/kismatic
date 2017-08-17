# Plan File Reference

## Index
* [cluster](#cluster)
  * [name](#clustername)
  * [admin_password](#clusteradmin_password)
  * [disable_package_installation](#clusterdisable_package_installation)
  * [package_repository_urls](#clusterpackage_repository_urls)
  * [disconnected_installation](#clusterdisconnected_installation)
  * [disable_registry_seeding](#clusterdisable_registry_seeding)
  * [networking](#clusternetworkingpod_cidr_block)
    * [pod_cidr_block](#clusternetworkingpod_cidr_block)
    * [service_cidr_block](#clusternetworkingservice_cidr_block)
    * [update_hosts_files](#clusternetworkingupdate_hosts_files)
    * [http_proxy](#clusternetworkinghttp_proxy)
    * [https_proxy](#clusternetworkinghttps_proxy)
    * [no_proxy](#clusternetworkingno_proxy)
  * [certificates](#clustercertificatesexpiry)
    * [expiry](#clustercertificatesexpiry)
    * [ca_expiry](#clustercertificatesca_expiry)
  * [ssh](#clustersshuser)
    * [user](#clustersshuser)
    * [ssh_key](#clustersshssh_key)
    * [ssh_port](#clustersshssh_port)
  * [kube_apiserver](#clusterkube_apiserveroption_overrides)
    * [option_overrides](#clusterkube_apiserveroption_overrides)
* [docker](#docker)
  * [storage](#dockerstoragedirect_lvmenabled)
    * [direct_lvm](#dockerstoragedirect_lvmenabled)
      * [enabled](#dockerstoragedirect_lvmenabled)
      * [block_device](#dockerstoragedirect_lvmblock_device)
      * [enable_deferred_deletion](#dockerstoragedirect_lvmenable_deferred_deletion)
* [docker_registry](#docker-registry)
  * [setup_internal](#docker_registrysetup_internal)
  * [address](#docker_registryaddress)
  * [port](#docker_registryport)
  * [ca](#docker_registryca)
* [add-ons](#add-ons)
  * [cni](#cni)
    * [disable](#add_onscnidisable)
    * [provider](#add_onscniprovider)
    * [options](#add_onsoptionscalicomode)
      * [calico](#add_onsoptionscalicomode)
        * [mode](#add_onsoptionscalicomode)
  * [dns](#dns)
    * [disable](#add_onsdnsdisable)
  * [heapster](#heapster)
    * [disable](#add_onsheapsterdisable)
    * [options](#add_onsheapsteroptionsheapsterreplicas)
      * [heapster](#add_onsheapsteroptionsheapsterreplicas)
        * [replicas](#add_onsheapsteroptionsheapsterreplicas)
        * [service_type](#add_onsheapsteroptionsheapsterservice_type)
        * [sink](#add_onsheapsteroptionsheapstersink)
      * [influxdb](#add_onsheapsteroptionsinfluxdbpvc_name)
        * [pvc_name](#add_onsheapsteroptionsinfluxdbpvc_name)
  * [dashboard](#dashboard)
    * [disable](#add_onsdashboarddisable)
  * [package_manager](#package-manager)
    * [disable](#add_onspackage_managerdisable)
    * [provider](#add_onspackage_managerprovider)
* [etcd](#etcd)
  * [expected_count](#etcdexpected_count)
  * [nodes](#etcdnodes)
    * [host](#etcdnodesihost)
    * [ip](#etcdnodesiip)
    * [internalip](#etcdnodesiinternalip)
* [master](#master)
  * [expected_count](#masterexpected_count)
  * [load_balanced_fqdn](#masterload_balanced_fqdn)
  * [load_balanced_short_name](#masterload_balanced_short_name)
  * [nodes](#masternodes)
    * [host](#masternodesihost)
    * [ip](#masternodesiip)
    * [internalip](#masternodesiinternalip)
* [worker](#worker)
  * [expected_count](#workerexpected_count)
  * [nodes](#workernodes)
    * [host](#workernodesihost)
    * [ip](#workernodesiip)
    * [internalip](#workernodesiinternalip)
* [ingress](#ingress)
  * [expected_count](#ingressexpected_count)
  * [nodes](#ingressnodes)
    * [host](#ingressnodesihost)
    * [ip](#ingressnodesiip)
    * [internalip](#ingressnodesiinternalip)
* [storage](#storage)
  * [expected_count](#storageexpected_count)
  * [nodes](#storagenodes)
    * [host](#storagenodesihost)
    * [ip](#storagenodesiip)
    * [internalip](#storagenodesiinternalip)

## cluster

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `name`   | Yes | ` ` | _custom_ | This property indicates the name of the cluster. It is used when generating assets that require a cluster name, such as kubeconfig files and certificates. |
| `admin_password` | Yes | ` ` | _custom_ | This property sets the password for the admin user. |
| `disable_package_installation` | No | `false` | `true`, `false` | This property disables the package installation process on the cluster nodes. When set to true, KET will not install the required packages. Instead, it will verify that the packages have been installed by the operator. |
| `package_repository_urls` | No | ` ` | _Comma-separated list of URLs_ | This property contains a comma-separated list of URLs of repositories that will be used for fetching the required packages. This is mainly used during a disconnected installation. In this scenario, internal package repositories that contain the KET packages and all their transitive dependencies should be listed here. Example: `http://rpm.apprenda.local:8080` | 
| `disconnected_installation` | No | `false` | `true`, `false` | This property indicates whether the cluster nodes are disconnected from the internet. When set to `true`, internal package repositories and container image registries are required for installation. See the [disconnected installation](./disconnected_install.md) documentation for more information. |
| `disable_registry_seeding` | No | `false` | `true`, `false` | This property disables the seeding of an internal container image registry during the installation. This is mainly used during a disconnected installation. When set to `true`, the internal container image registry must be seeded before performing the installation. |
| `networking.pod_cidr_block` | Yes | ` ` | _Network CIDR_ | This property indicates the pod network's CIDR block. For example: `172.16.0.0/16` |
| `networking.service_cidr_block` | Yes | ` ` | _Network CIDR_ | This property indicates the Kubernetes service network's CIDR block. For example: `172.20.0.0/16` |
| `networking.update_hosts_files` | No | `false` | `true`, `false` | This property controls whether the `/etc/hosts` file should be updated on the cluster nodes. When set to `true`, KET will updated the hosts file on all nodes to include entries for all other nodes in the cluster. |
| <a id="clusternetworkinghttp_proxy"></a> `networking.http_proxy` | No | ` ` | _URL_ | This property indicates the URL of the proxy that should be used for HTTP connections. For example: `http://secure.apprenda.local:3128` |
| `networking.https_proxy` | No | ` ` | _URL_ | This property indicates the URL of the proxy that should be used for HTTPS connections. For example: `https://secure.apprenda.local:3128` |
| `networking.no_proxy` | No | ` ` | _Comma-separated list_ | This property contains a comma-separated list of host names and/or IPs for which connections will not go through a proxy. For example: `ignore-proxy.com,8.8.8.8,8.8.4.4` |
| `certificates.expiry` | Yes | ` ` | _Integer followed by unit of time (`h`,`m`, or `s`)_ | This property indicates the length of time that generated certificates should be valid for. For example: `17520h` for 2 years. |
| `certificates.ca_expiry` | Yes | ` ` | _Integer followed by unit of time (`h`,`m`, or `s`)_ | This property indicates the length of time that the generated Certificate Authority should be valid for. For example: `17520h` for 2 years. |
| `ssh.user` | Yes | ` ` | _custom_ | This property indicates the user for accessing the cluster nodes via SSH. This user requires sudo elevation privileges on the cluster nodes. |
| `ssh.ssh_key` | Yes | ` ` | _custom_ | This property indicates the absolute path of the SSH key that should be used for accessing the cluster nodes.
| `ssh.ssh_port` | Yes | ` ` | _integer_ | This property indicates the port number on which cluster nodes are listening for SSH connections.
| `kube_apiserver.option_overrides` | No | ` ` | _Map with string keys and string values_ | This property contains a listing of option overrides that should be applied to the Kubernetes API server configuration. This is an advanced feature that can prevent the API server from starting up if invalid configuration is provided. For example: `{"event_ttl": "2h0m0s"}` |

## Docker

#### docker.storage.direct_lvm.enabled

This property controls whether the `direct_lvm` mode of the `devicemapper` storage driver should be enabled. When set to `true`, a dedicated block storage device must be available on each cluster node. See the [documentation](https://github.com/apprenda/kismatic/blob/master/docs/docker.md#storage-in-rhelcentos-ket-v131) on this KET feature for more information.

| Required | No |
|----------|-----|
| **Default** | `false` |
| **Options** | `true`, `false` |

#### docker.storage.direct_lvm.block_device

This property indicates the path to the block storage device that will be used by the `devicemapper` storage driver.

| Required | No |
|----------|-----|
| **Default** | ` ` |
| **Format** | Absolute path to the block storage device |
| **Example** | `/dev/sdb1` |

#### docker.storage.direct_lvm.enable_deferred_deletion

This property indicates whether deferred deletion should be enabled when using `devicemapper` in `direct_lvm` mode.

| Required | No |
|----------|-----|
| **Default** | `false` |

## Docker Registry

#### docker_registry.setup_internal

This property indicates whether an internal docker registry should be installed on the cluster. When set to `true`, a registry will be deployed on the first master node. See the [documentation](https://github.com/apprenda/kismatic/blob/master/docs/docker_registry.md) on this KET feature for more information.

| Required | No |
|----------|-----|
| **Default** | `false` |

#### docker_registry.address

This property indicates the hostname or IP of a private registry that is accessible from the cluster nodes. When performing a disconnected installation, this registry will be used to fetch all the required container images.

| Required | No |
|----------|-----|
| **Default** | ` ` |

#### docker_registry.port

This property indicates the port on which your private registry is listening for incomming connections.

| Required | No |
|----------|-----|
| **Default** | ` ` |

#### docker_registry.CA

This property indicates the absolute path of the Certificate Authority that should be installed on all cluster nodes that have a docker daemon. This is required to establish trust between the daemons and the private registry when the registry is using a self-signed certificate.

| Required | No |
|----------|-----|
| **Default** | ` ` |

## Add-Ons

The complete documentation on add-ons can be found [here](./add-ons.md).

### CNI

#### add_ons.cni.disable

This property indicates whether the CNI add-on is disabled. When set to `true`, CNI will not be installed on the cluster. Furthermore, the smoke test and any validation that depends on a functional pod network will be skipped.

| Required | No |
|----------|-----|
| **Default** | `false` |

#### add_ons.cni.provider

This property indicates the CNI provider that should be installed.

| Required | No |
|----------|-----|
| **Default** | `calico` |
| **Options** | `calico`, `weave`, `contiv`, `custom` |

#### add_ons.options.calico.mode

This property indicates the datapath technique that should be configured in Calico.

| Required | No |
|----------|-----|
| **Default** | `overlay` |
| **Options** | `overlay`, `routed` |

### DNS

#### add_ons.dns.disable

This property indicates whether the DNS add-on should be disabled. When set to `true`, no DNS solution will be deployed on the cluster.

| Required | No |
|----------|-----|
| **Default** | `false` |

### Heapster

#### add_ons.heapster.disable

This property indicates whether the Heapster add-on should be disabled. When set to `true`, Heapster and InfluxDB will not be deployed on the cluster.

| Required | No |
|----------|-----|
| **Default** | `false` |

#### add_ons.heapster.options.heapster.replicas

This property indicates the number of Heapster replicas that should be scheduled on the cluster.

| Required | No |
|----------|-----|
| **Default** | `2` |

#### add_ons.heapster.options.heapster.service_type

This property indicates the Kubernetes service type that should be used when creating the Heapster service.

| Required | No |
|----------|----|
| **Default** | `ClusterIP` |
| **Options** | `ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName` |

#### add_ons.heapster.options.heapster.sink

This property indicates the URL of the backend store that should be configured as the Heapster sink. More information on Heapster can be found [here](https://github.com/kubernetes/heapster).

| Required | No |
|----------|----|
| **Default** | `http://heapster-influxdb.kube-system.svc:8086` |

#### add_ons.heapster.options.influxdb.pvc_name

This property indicates the name of the Persistent Volume Claim that will be used by InfluxDB. When set, this PVC must be created after the installation.

If not set, InfluxDB will be configured with ephemeral storage.

| Required | No |
|----------|----|
| **Default** | ` ` |

### Dashboard 

#### add_ons.dashboard.disable

This property indicates whether the dashboard add-on should be disabled. When set to `true`, the Kubernetes Dashboard will not be installed on the cluster.

| Required | No |
|----------|-----|
| **Default** | `false` |

#### add_ons.dashbard.disable (deprecated)

This property indicates whether the dashboard add-on should be disabled. When set to `true`, the Kubernetes Dashboard will not be installed on the cluster.

| Required | No |
|----------|-----|
| **Default** | `false` |

### Package Manager

#### add_ons.package_manager.disable

This property indicates whether the package manager add-on should be disabled. When set to `true`, the package manager will not be installed on the cluster.

| Required | No |
|----------|-----|
| **Default** | `false` |

#### add_ons.package_manager.provider

This property indicates the package manager provider. 

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

## Etcd

#### etcd.expected_count

This property indicates the number of etcd nodes that will be part of the cluster.

| Required | Yes |
|----------|-----|
| **Default** | `0` |

#### etcd.nodes

This property lists the etcd nodes that will be part of the cluster.

#### etcd.nodes[i].host

This property indicates the hostname of the etcd node. The hostname is verified in the validation phase of the installation.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### etcd.nodes[i].ip

This property indicates the IP address of the etcd node. The installation node will access the node over SSH using this IP address.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### etcd.nodes[i].internalip

This property indicates the internal (or private) IP address of the etcd node. If not empty, this IP will be used when configuring cluster components.

| Required | No |
|----------|-----|
| **Default** | ` ` |

## Master

#### master.expected_count

This property indicates the number of master nodes that will be part of the cluster.

| Required | Yes |
|----------|-----|
| **Default** | `0` |

#### master.load_balanced_fqdn

This property indicates the FQDN of the load balancer that is fronting multiple master nodes.

In the case where there is only one master node, this can be set to the IP address of the master node.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### master.load_balanced_short_name

This property indicates the short name of the load balancer that is fronting multiple master nodes.

In the case where there is only one master node, this can be set to the IP address of the master node.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### master.nodes

This property lists the master nodes that will be part of the cluster.

#### master.nodes[i].host

This property indicates the hostname of the master node. The hostname is verified in the validation phase of the installation.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### master.nodes[i].ip

This property indicates the IP address of the master node. The installation node will access the node over SSH using this IP address.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### master.nodes[i].internalip

This property indicates the internal (or private) IP address of the master node. If not empty, this IP will be used when configuring cluster components.

| Required | No |
|----------|-----|
| **Default** | ` ` |

## Worker

#### worker.expected_count

This property indicates the number of worker nodes that will be part of the cluster.

| Required | Yes |
|----------|-----|
| **Default** | `0` |

#### worker.nodes

This property lists the worker nodes that will be part of the cluster.

#### worker.nodes[i].host

This property indicates the hostname of the worker node. The hostname is verified in the validation phase of the installation.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### worker.nodes[i].ip

This property indicates the IP address of the worker node. The installation node will access the node over SSH using this IP address.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### worker.nodes[i].internalip

This property indicates the internal (or private) IP address of the worker node. If not empty, this IP will be used when configuring cluster components.

| Required | No |
|----------|-----|
| **Default** | ` ` |

## Ingress

#### ingress.expected_count

This property indicates the number of ingress nodes that will be part of the cluster.

| Required | Yes |
|----------|-----|
| **Default** | `0` |

#### ingress.nodes

This property lists the ingress nodes that will be part of the cluster.

#### ingress.nodes[i].host

This property indicates the hostname of the ingress node. The hostname is verified in the validation phase of the installation.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### ingress.nodes[i].ip

This property indicates the IP address of the ingress node. The installation node will access the node over SSH using this IP address.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### ingress.nodes[i].internalip

This property indicates the internal (or private) IP address of the ingress node. If not empty, this IP will be used when configuring cluster components.

| Required | No |
|----------|-----|
| **Default** | ` ` |

## Storage

#### storage.expected_count

This property indicates the number of storage nodes that will be part of the cluster.

| Required | Yes |
|----------|-----|
| **Default** | `0` |

#### storage.nodes

This property lists the storage nodes that will be part of the cluster.

#### storage.nodes[i].host

This property indicates the hostname of the storage node. The hostname is verified in the validation phase of the installation.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### storage.nodes[i].ip

This property indicates the IP address of the storage node. The installation node will access the node over SSH using this IP address.

| Required | Yes |
|----------|-----|
| **Default** | ` ` |

#### storage.nodes[i].internalip

This property indicates the internal (or private) IP address of the storage node. If not empty, this IP will be used when configuring cluster components.

| Required | No |
|----------|-----|
| **Default** | ` ` |