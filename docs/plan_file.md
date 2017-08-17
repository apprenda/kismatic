# Plan File Reference

## Index
* [cluster](#cluster)
  * [name](#cluster_name)
  * [admin_password](#cluster_admin_password)
  * [disable_package_installation](#cluster_disable_package_installation)
  * [package_repository_urls](#cluster_package_repository_urls)
  * [disconnected_installation](#cluster_disconnected_installation)
  * [disable_registry_seeding](#cluster_disable_registry_seeding)
  * [networking](#cluster_networking_pod_cidr_block)
    * [pod_cidr_block](#cluster_networking_pod_cidr_block)
    * [service_cidr_block](#cluster_networking_service_cidr_block)
    * [update_hosts_files](#cluster_networking_update_hosts_files)
    * [http_proxy](#cluster_networking_http_proxy)
    * [https_proxy](#cluster_networking_https_proxy)
    * [no_proxy](#cluster_networking_no_proxy)
  * [certificates](#cluster_certificates_expiry)
    * [expiry](#cluster_certificates_expiry)
    * [ca_expiry](#cluster_certificates_ca_expiry)
  * [ssh](#cluster_ssh_user)
    * [user](#cluster_ssh_user)
    * [ssh_key](#cluster_ssh_ssh_key)
    * [ssh_port](#cluster_ssh_ssh_port)
  * [kube_apiserver](#cluster_kube_apiserver_option_overrides)
    * [option_overrides](#cluster_kube_apiserver_option_overrides)
* [docker](#docker)
  * [storage](#docker_storage_direct_lvm_enabled)
    * [direct_lvm](#docker_storage_direct_lvm_enabled)
      * [enabled](#docker_storage_direct_lvm_enabled)
      * [block_device](#docker_storage_direct_lvm_block_device)
      * [enable_deferred_deletion](#docker_storage_direct_lvm_enable_deferred_deletion)
* [docker_registry](#docker_registry)
  * [setup_internal](#docker_registry_setup_internal)
  * [address](#docker_registry_address)
  * [port](#docker_registry_port)
  * [ca](#docker_registry_ca)
* [add_ons](#add_ons)
  * [cni](#cni)
    * [disable](#add_ons_cni_disable)
    * [provider](#add_ons_cni_provider)
    * [options](#add_ons_cni_options_calico_mode)
      * [calico](#add_ons_cni_options_calico_mode)
        * [mode](#add_ons_cni_options_calico_mode)
  * [dns](#dns)
    * [disable](#add_ons_dns_disable)
  * [heapster](#heapster)
    * [disable](#add_ons_heapster_disable)
    * [options](#add_ons_heapster_options_heapster_replicas)
      * [heapster](#add_ons_heapster_options_heapster_replicas)
        * [replicas](#add_ons_heapster_options_heapster_replicas)
        * [service_type](#add_ons_heapster_options_heapster_service_type)
        * [sink](#add_ons_heapster_options_heapster_sink)
      * [influxdb](#add_ons_heapster_options_influxdb_pvc_name)
        * [pvc_name](#add_ons_heapster_options_influxdb_pvc_name)
  * [dashboard](#dashboard)
    * [disable](#add_ons_dashboard_disable)
  * [package_manager](#package_manager)
    * [disable](#add_ons_package_manager_disable)
    * [provider](#add_ons_package_manager_provider)
* [etcd](#etcd)
  * [expected_count](#etc_dexpected_count)
  * [nodes](#etcd_nodes)
    * [host](#etcd_nodes_host)
    * [ip](#etcd_nodes_ip)
    * [internalip](#etcd_nodes_internalip)
* [master](#master)
  * [expected_count](#master_expected_count)
  * [load_balanced_fqdn](#master_load_balanced_fqdn)
  * [load_balanced_short_name](#master_load_balanced_short_name)
  * [nodes](#master_nodes)
    * [host](#master_nodes_i_host)
    * [ip](#master_nodes_i_ip)
    * [internalip](#master_nodes_i_internalip)
* [worker](#worker)
  * [expected_count](#workerexpected_count)
  * [nodes](#worker_nodes)
    * [host](#worker_nodes_i_host)
    * [ip](#worker_nodes_i_ip)
    * [internalip](#worker_nodes_i_internalip)
* [ingress](#ingress)
  * [expected_count](#ingress_expected_count)
  * [nodes](#ingress_nodes)
    * [host](#ingress_nodes_i_host)
    * [ip](#ingress_nodes_i_ip)
    * [internalip](#ingress_nodes_i_internalip)
* [storage](#storage)
  * [expected_count](#storage_expected_count)
  * [nodes](#storage_nodes)
    * [host](#storage_nodes_i_host)
    * [ip](#storage_nodes_i_ip)
    * [internalip](#storage_nodes_i_internalip)

## cluster

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `name`   | Yes | ` ` | _string_ | <a id="cluster_name"></a>  This property indicates the name of the cluster. It is used when generating assets that require a cluster name, such as kubeconfig files and certificates. |
| `admin_password` | Yes | ` ` | _string_ | <a id="cluster_admin_password"></a> This property sets the password for the admin user.  |
| `disable_package_installation` | No | `false` | `true`, `false` | <a id="cluster_disable_package_installation"></a> This property disables the package installation process on the cluster nodes. When set to true, KET will not install the required packages. Instead, it will verify that the packages have been installed by the operator.  |
| `package_repository_urls` | No | ` ` | _Comma-separated list of URLs_ | <a id="cluster_package_repository_urls"></a> This property contains a comma-separated list of URLs of repositories that will be used for fetching the required packages. This is mainly used during a disconnected installation. In this scenario, internal package repositories that contain the KET packages and all their transitive dependencies should be listed here. Example: `http://rpm.apprenda.local:8080` | 
| `disconnected_installation` | No | `false` | `true`, `false` | <a id="cluster_disconnected_installation"></a> This property indicates whether the cluster nodes are disconnected from the internet. When set to `true`, internal package repositories and container image registries are required for installation. See the [disconnected installation](./disconnected_install.md) documentation for more information. |
| `disable_registry_seeding` | No | `false` | `true`, `false` | <a id="cluster_disable_registry_seeding"></a> This property disables the seeding of an internal container image registry during the installation. This is mainly used during a disconnected installation. When set to `true`, the internal container image registry must be seeded before performing the installation. |
| `networking.pod_cidr_block` | Yes | ` ` | _Network CIDR_ | <a id="cluster_networking_pod_cidr_block"></a> This property indicates the pod network's CIDR block. For example: `172.16.0.0/16` |
| `networking.service_cidr_block` | Yes | ` ` | _Network CIDR_ | <a id="cluster_networking_service_cidr_block"></a>  This property indicates the Kubernetes service network's CIDR block. For example: `172.20.0.0/16` |
| `networking.update_hosts_files` | No | `false` | `true`, `false` | <a id="cluster_networking_update_hosts_files"></a> This property controls whether the `/etc/hosts` file should be updated on the cluster nodes. When set to `true`, KET will updated the hosts file on all nodes to include entries for all other nodes in the cluster. |
| `networking.http_proxy` | No | ` ` | _URL_ | <a id="cluster_networking_http_proxy"></a> This property indicates the URL of the proxy that should be used for HTTP connections. For example: `http://secure.apprenda.local:3128` |
| `networking.https_proxy` | No | ` ` | _URL_ | <a id="cluster_networking_https_proxy"></a> This property indicates the URL of the proxy that should be used for HTTPS connections. For example: `https://secure.apprenda.local:3128` |
| `networking.no_proxy` | No | ` ` | _Comma-separated list_ | <a id="cluster_networking_no_proxy"></a> This property contains a comma-separated list of host names and/or IPs for which connections will not go through a proxy. For example: `ignore-proxy.com,8.8.8.8,8.8.4.4` |
| `certificates.expiry` | Yes | ` ` | _Integer followed by unit of time (`h`,`m`, or `s`)_ | <a id="cluster_certificates_expiry"></a> This property indicates the length of time that generated certificates should be valid for. For example: `17520h` for 2 years. |
| `certificates.ca_expiry` | Yes | ` ` | _Integer followed by unit of time (`h`,`m`, or `s`)_ | <a id="cluster_certificates_ca_expiry"></a> This property indicates the length of time that the generated Certificate Authority should be valid for. For example: `17520h` for 2 years. |
| `ssh.user` | Yes | ` ` | _string_ | <a id="cluster_ssh_ssh_user"></a> This property indicates the user for accessing the cluster nodes via SSH. This user requires sudo elevation privileges on the cluster nodes. |
| `ssh.ssh_key` | Yes | ` ` | _string_ | <a id="cluster_ssh_ssh_key"></a> This property indicates the absolute path of the SSH key that should be used for accessing the cluster nodes.
| `ssh.ssh_port` | Yes | ` ` | _integer_ | <a id="cluster_ssh_ssh_port"></a> This property indicates the port number on which cluster nodes are listening for SSH connections.
| `kube_apiserver.option_overrides` | No | ` ` | _map with string keys and string values_ | <a id="cluster_kube_apiserver_option_overrides"></a> This property contains a listing of option overrides that should be applied to the Kubernetes API server configuration. This is an advanced feature that can prevent the API server from starting up if invalid configuration is provided. For example: `{"event_ttl": "2h0m0s"}` |

## docker

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `storage.direct_lvm.enabled` | No | `false` | `true`, `false` | <a id="docker_storage_direct_lvm_enabled"></a> This property controls whether the `direct_lvm` mode of the `devicemapper` storage driver should be enabled. When set to `true`, a dedicated block storage device must be available on each cluster node. See the [documentation](https://github.com/apprenda/kismatic/blob/master/docs/docker.md#storage-in-rhelcentos-ket-v131) on this KET feature for more information. |
| `storage.direct_lvm.block_device` | No | ` ` | _Absolute path_ | <a id="docker_storage_direct_lvm_block_device"></a> This property indicates the path to the block storage device that will be used by the `devicemapper` storage driver. |
| `storage.direct_lvm.enable_deferred_deletion` | No | `false` | `true`, `false` | <a id="docker_storage_direct_lvm_enable_deferred_deletion"></a> This property indicates whether deferred deletion should be enabled when using `devicemapper` in `direct_lvm` mode. |


## docker_registry

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `setup_internal` | No | `false` | `true`, `false` | <a id="docker_registry_setup_internal"></a> This property indicates whether an internal docker registry should be installed on the cluster. When set to `true`, a registry will be deployed on the first master node. See the [documentation](https://github.com/apprenda/kismatic/blob/master/docs/docker_registry.md) on this KET feature for more information. |
| `address` | No | ` ` | _string_ | <a id="docker_registry_address"></a> This property indicates the hostname or IP of a private registry that is accessible from the cluster nodes. When performing a disconnected installation, this registry will be used to fetch all the required container images.
| `port` | No | `0` | _integer_ | <a id="docker_registry_port"></a> This property indicates the port on which your private registry is listening for incomming connections. | 
| `ca` | No | ` ` | _absolute path_ | <a id="docker_registry_ca"></a> This property indicates the absolute path of the Certificate Authority that should be installed on all cluster nodes that have a docker daemon. This is required to establish trust between the daemons and the private registry when the registry is using a self-signed certificate. |

## add_ons

The complete documentation on add-ons can be found [here](./add-ons.md).

### cni

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `disable` | No | `false` | `true`, `false` | <a id="add_ons_cni_disable"></a> This property indicates whether the CNI add-on is disabled. When set to `true`, CNI will not be installed on the cluster. Furthermore, the smoke test and any validation that depends on a functional pod network will be skipped. |
| `provider` | No | `calico` | `calico`, `weave`, `contiv`, `custom` | <a id="add_ons_cni_provider"></a> This property indicates the CNI provider that should be installed. |
| `options.calico.mode` | No | `overlay` | `overlay`, `routed` | <a id="add_ons_cni_options_calico_mode"></a> This property indicates the datapath technique that should be configured in Calico. |

### dns

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `disable` | No | `false` | `true`, `false` | <a id="add_ons_dns_disable"></a> This property indicates whether the DNS add-on should be disabled. When set to `true`, no DNS solution will be deployed on the cluster. |

### heapster

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `disable` | No | `false` | `true`, `false` | <a id="add_ons_heapster_disable"></a> This property indicates whether the Heapster add-on should be disabled. When set to `true`, Heapster and InfluxDB will not be deployed on the cluster. |
| `options.heapster.replicas` | No | `2` | _integer_ | <a id="add_ons_heapster_options_heapster_replicas"></a> This property indicates the number of Heapster replicas that should be scheduled on the cluster.|
| `options.heapster.service_type` | No | `ClusterIP` | `ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName` | <a id="add_ons_heapster_options_heapster_service_type"></a> This property indicates the Kubernetes service type that should be used when creating the Heapster service. |
| `options.heapster.sink` | No | `http://heapster-influxdb.kube-system.svc:8086` | _URL_ | <a id="add_ons_heapster_options_heapster_sink"></a> This property indicates the URL of the backend store that should be configured as the Heapster sink. More information on Heapster can be found [here](https://github.com/kubernetes/heapster).|
| `options.influxdb.pvc_name` | No | ` ` | _string_ | <a id="add_ons_heapster_options_influxdb_pvc_name"></a> This property indicates the name of the Persistent Volume Claim that will be used by InfluxDB. When set, this PVC must be created after the installation. If not set, InfluxDB will be configured with ephemeral storage. |

### dashboard 

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `disable` | No | `false` | `true`, `false` | <a id="add_ons_dashboard_disable"></a> This property indicates whether the dashboard add-on should be disabled. When set to `true`, the Kubernetes Dashboard will not be installed on the cluster. |

### package_manager

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `disable` | No | `false` | `true`, `false` | <a id="add_ons_package_manager_disable"></a> This property indicates whether the package manager add-on should be disabled. When set to `true`, the package manager will not be installed on the cluster. |
| `provider` | Yes | ` ` | `helm` | <a id="add_ons_package_manager_provider"></a> This property indicates the package manager provider.|

## etcd

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `expected_count` | Yes | `0` | _integer_ | This property indicates the number of etcd nodes that will be part of the cluster. |
| `nodes` | Yes | `[]` | _list_ | This property lists the etcd nodes that will be part of the cluster. |
| `nodes[i].host` | Yes | ` ` | _string_ | This property indicates the hostname of the etcd node. The hostname is verified in the validation phase of the installation. |
| `nodes[i].ip` | Yes | ` ` | _ip_ | This property indicates the IP address of the etcd node. The installation node will access the node over SSH using this IP address.|
| `nodes[i].internalip` | No | ` ` | _ip_ | This property indicates the internal (or private) IP address of the etcd node. If not empty, this IP will be used when configuring cluster components.|


## master

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `expected_count` | Yes | `0` | _integer_ | <a id="master_expected_count"></a> This property indicates the number of master nodes that will be part of the cluster. |
| `load_balanced_fqdn` | Yes | ` ` | _string_ | <a id="master_load_balanced_fqdn"></a> This property indicates the FQDN of the load balancer that is fronting multiple master nodes. In the case where there is only one master node, this can be set to the IP address of the master node. |
| `load_balanced_short_name` | Yes | ` ` | _string_ | <a id="master_load_balanced_short_name"></a> This property indicates the short name of the load balancer that is fronting multiple master nodes. In the case where there is only one master node, this can be set to the IP address of the master node. |
| `nodes` | Yes | `[]` | _list_ | <a id="master_nodes"></a> This property lists the master nodes that will be part of the cluster. |
| `nodes[i].host` | Yes | ` ` | _string_ | <a id="master_nodes_i_host"></a> This property indicates the hostname of the master node. The hostname is verified in the validation phase of the installation. |
| `nodes[i].ip` | Yes | ` ` | _ip_ | <a id="master_nodes_i_ip"></a> This property indicates the IP address of the master node. The installation node will access the node over SSH using this IP address.|
| `nodes[i].internalip` | No | ` ` | _ip_ | <a id="master_nodes_i_internalip"></a> This property indicates the internal (or private) IP address of the master node. If not empty, this IP will be used when configuring cluster components.|

## worker

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `expected_count` | Yes | `0` | _integer_ | <a id="worker_expected_count"></a> This property indicates the number of worker nodes that will be part of the cluster. |
| `nodes` | Yes | `[]` | _list_ | <a id="worker_nodes"></a> This property lists the worker nodes that will be part of the cluster. |
| `nodes[i].host` | Yes | ` ` | _string_ | <a id="worker_nodes_i_host"></a> This property indicates the hostname of the worker node. The hostname is verified in the validation phase of the installation. |
| `nodes[i].ip` | Yes | ` ` | _ip_ | <a id="worker_nodes_i_ip"></a> This property indicates the IP address of the worker node. The installation node will access the node over SSH using this IP address.|
| `nodes[i].internalip` | No | ` ` | _ip_ | <a id="worker_nodes_i_internalip"></a> This property indicates the internal (or private) IP address of the worker node. If not empty, this IP will be used when configuring cluster components.|

## ingress

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `expected_count` | Yes | `0` | _integer_ | <a id="ingress_expected_count"></a> This property indicates the number of ingress nodes that will be part of the cluster. |
| `nodes` | Yes | `[]` | _list_ | <a id="ingress_nodes"></a> This property lists the ingress nodes that will be part of the cluster. |
| `nodes[i].host` | Yes | ` ` | _string_ | <a id="ingress_nodes_i_host"></a> This property indicates the hostname of the ingress node. The hostname is verified in the validation phase of the installation. |
| `nodes[i].ip` | Yes | ` ` | _ip_ | <a id="ingress_nodes_i_ip"></a> This property indicates the IP address of the ingress node. The installation node will access the node over SSH using this IP address.|
| `nodes[i].internalip` | No | ` ` | _ip_ | <a id="ingress_nodes_i_internalip"></a> This property indicates the internal (or private) IP address of the ingress node. If not empty, this IP will be used when configuring cluster components.|

## storage

| Property | Required | Default Value | Allowed Value | Description |
|----------|----------|---------|----------------|-------------|
| `expected_count` | Yes | `0` | _integer_ | <a id="storage_expected_count"></a> This property indicates the number of storage nodes that will be part of the cluster. |
| `nodes` | Yes | `[]` | _list_ | <a id="storage_nodes"></a> This property lists the storage nodes that will be part of the cluster. |
| `nodes[i].host` | Yes | ` ` | _string_ | <a id="storage_nodes_i_host"></a> This property indicates the hostname of the storage node. The hostname is verified in the validation phase of the installation. |
| `nodes[i].ip` | Yes | ` ` | _ip_ | <a id="storage_nodes_i_ip"></a> This property indicates the IP address of the storage node. The installation node will access the node over SSH using this IP address.|
| `nodes[i].internalip` | No | ` ` | _ip_ | <a id="storage_nodes_i_internalip"></a> This property indicates the internal (or private) IP address of the storage node. If not empty, this IP will be used when configuring cluster components.|
