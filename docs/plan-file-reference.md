# Plan File Reference
## Index
* [cluster](#cluster)
  * [name](#clustername)
  * [admin_password](#clusteradmin_password)
  * [disable_package_installation](#clusterdisable_package_installation)
  * [allow_package_installation _(deprecated)_](#clusterallow_package_installation-deprecated)
  * [package_repository_urls](#clusterpackage_repository_urls)
  * [disconnected_installation](#clusterdisconnected_installation)
  * [disable_registry_seeding](#clusterdisable_registry_seeding)
  * [networking](#clusternetworking)
    * [type _(deprecated)_](#clusternetworkingtype-deprecated)
    * [pod_cidr_block](#clusternetworkingpod_cidr_block)
    * [service_cidr_block](#clusternetworkingservice_cidr_block)
    * [update_hosts_files](#clusternetworkingupdate_hosts_files)
    * [http_proxy](#clusternetworkinghttp_proxy)
    * [https_proxy](#clusternetworkinghttps_proxy)
    * [no_proxy](#clusternetworkingno_proxy)
  * [certificates](#clustercertificates)
    * [expiry](#clustercertificatesexpiry)
    * [ca_expiry](#clustercertificatesca_expiry)
  * [ssh](#clusterssh)
    * [user](#clustersshuser)
    * [ssh_key](#clustersshssh_key)
    * [ssh_port](#clustersshssh_port)
  * [kube_apiserver](#clusterkube_apiserver)
  * [cloud_provider](#clustercloud_provider)
    * [provider](#clustercloud_providerprovider)
    * [config](#clustercloud_providerconfig)
* [docker](#docker)
  * [storage](#dockerstorage)
    * [direct_lvm](#dockerstoragedirect_lvm)
      * [enabled](#dockerstoragedirect_lvmenabled)
      * [block_device](#dockerstoragedirect_lvmblock_device)
      * [enable_deferred_deletion](#dockerstoragedirect_lvmenable_deferred_deletion)
* [docker_registry](#docker_registry)
  * [setup_internal](#docker_registrysetup_internal)
  * [address](#docker_registryaddress)
  * [port](#docker_registryport)
  * [CA](#docker_registryCA)
* [add_ons](#add_ons)
  * [cni](#add_onscni)
    * [disable](#add_onscnidisable)
    * [provider](#add_onscniprovider)
    * [options](#add_onscnioptions)
      * [calico](#add_onscnioptionscalico)
        * [mode](#add_onscnioptionscalicomode)
  * [dns](#add_onsdns)
    * [disable](#add_onsdnsdisable)
  * [heapster](#add_onsheapster)
    * [disable](#add_onsheapsterdisable)
    * [options](#add_onsheapsteroptions)
      * [heapster](#add_onsheapsteroptionsheapster)
        * [replicas](#add_onsheapsteroptionsheapsterreplicas)
        * [service_type](#add_onsheapsteroptionsheapsterservice_type)
        * [sink](#add_onsheapsteroptionsheapstersink)
      * [influxdb](#add_onsheapsteroptionsinfluxdb)
        * [pvc_name](#add_onsheapsteroptionsinfluxdbpvc_name)
      * [heapster_replicas _(deprecated)_](#add_onsheapsteroptionsheapster_replicas-deprecated)
      * [influxdb_pvc_name _(deprecated)_](#add_onsheapsteroptionsinfluxdb_pvc_name-deprecated)
  * [dashboard](#add_onsdashboard)
    * [disable](#add_onsdashboarddisable)
  * [dashbard _(deprecated)_](#add_onsdashbard-deprecated)
    * [disable](#add_onsdashbarddisable)
  * [package_manager](#add_onspackage_manager)
    * [disable](#add_onspackage_managerdisable)
    * [provider](#add_onspackage_managerprovider)
* [features _(deprecated)_](#features-deprecated)
  * [package_manager _(deprecated)_](#featurespackage_manager-deprecated)
    * [enabled _(deprecated)_](#featurespackage_managerenabled-deprecated)
* [etcd](#etcd)
  * [expected_count](#etcdexpected_count)
  * [nodes](#etcdnodes)
    * [host](#etcdnodeshost)
    * [ip](#etcdnodesip)
    * [internalip](#etcdnodesinternalip)
* [master](#master)
  * [expected_count](#masterexpected_count)
  * [load_balanced_fqdn](#masterload_balanced_fqdn)
  * [load_balanced_short_name](#masterload_balanced_short_name)
  * [nodes](#masternodes)
    * [host](#masternodeshost)
    * [ip](#masternodesip)
    * [internalip](#masternodesinternalip)
* [worker](#worker)
  * [expected_count](#workerexpected_count)
  * [nodes](#workernodes)
    * [host](#workernodeshost)
    * [ip](#workernodesip)
    * [internalip](#workernodesinternalip)
* [ingress](#ingress)
  * [expected_count](#ingressexpected_count)
  * [nodes](#ingressnodes)
    * [host](#ingressnodeshost)
    * [ip](#ingressnodesip)
    * [internalip](#ingressnodesinternalip)
* [storage](#storage)
  * [expected_count](#storageexpected_count)
  * [nodes](#storagenodes)
    * [host](#storagenodeshost)
    * [ip](#storagenodesip)
    * [internalip](#storagenodesinternalip)
* [nfs](#nfs)
  * [nfs_volume](#nfsnfs_volume)
    * [nfs_host](#nfsnfs_volumenfs_host)
    * [mount_path](#nfsnfs_volumemount_path)
##  cluster

 Kubernetes cluster configuration 

###  cluster.name

 Name of the cluster to be used when generating assets that require a cluster name, such as kubeconfig files and certificates. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.admin_password

 The password for the admin user. This is mainly used to access the Kubernetes Dashboard. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.disable_package_installation

 Whether KET should install the packages on the cluster nodes. When true, KET will not install the required packages. Instead, it will verify that the packages have been installed by the operator. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  cluster.allow_package_installation _(deprecated)_

 Whether KET should install the packages on the cluster nodes. Use DisablePackageInstallation instead. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  cluster.package_repository_urls

 Comma-separated list of URLs of repositories that will be used for fetching the required packages. This is mainly used during a disconnected installation. In this scenario, internal package repositories that contain the KET packages and all their transitive dependencies should be listed here. Example: `http://rpm.apprenda.local:8080` 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  cluster.disconnected_installation

 Whether the cluster nodes are disconnected from the internet. When set to `true`, internal package repositories and a container image registry are required for installation. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  cluster.disable_registry_seeding

 Whether KET should seed an internal container image registry during the installation. This is mainly used during a disconnected installation. When set to true, the internal container image registry must be manually seeded before the installation. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  cluster.networking

 The Networking configuration for the cluster. 

###  cluster.networking.type _(deprecated)_

 The datapath technique that should be configured in Calico. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | `overlay` | 
| **Options** |  `overlay`, `routed`

###  cluster.networking.pod_cidr_block

 The pod network's CIDR block. For example: `172.16.0.0/16` 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.networking.service_cidr_block

 The Kubernetes service network's CIDR block. For example: `172.20.0.0/16` 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.networking.update_hosts_files

 Whether the /etc/hosts file should be updated on the cluster nodes. When set to true, KET will update the hosts file on all nodes to include entries for all other nodes in the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  cluster.networking.http_proxy

 The URL of the proxy that should be used for HTTP connections. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  cluster.networking.https_proxy

 The URL of the proxy that should be used for HTTPS connections. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  cluster.networking.no_proxy

 Comma-separated list of host names and/or IPs for which connections should not go through a proxy. All nodes' 'host' and 'IPs' are always set. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  cluster.certificates

 The Certificates configuration for the cluster. 

###  cluster.certificates.expiry

 The length of time that the generated certificates should be valid for. For example: "17520h" for 2 years. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.certificates.ca_expiry

 The length of time that the generated Certificate Authority should be valid for. For example: "17520h" for 2 years. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.ssh

 The SSH configuration for the cluster nodes. 

###  cluster.ssh.user

 The user for accessing the cluster nodes via SSH. This user requires sudo elevation privileges on the cluster nodes. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.ssh.ssh_key

 The absolute path of the SSH key that should be used for accessing the cluster nodes via SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.ssh.ssh_port

 The port number on which cluster nodes are listening for SSH connections. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  cluster.kube_apiserver

 Listing of option overrides that are to be applied to the Kubernetes API server configuration. This is an advanced feature that can prevent the API server from starting up if invalid configuration is provided. 

###  cluster.cloud_provider

 The CloudProvider configuration for the cluster. 

###  cluster.cloud_provider.provider

 The cloud provider that should be set in the Kubernetes components 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 
| **Options** |  `aws`, `azure`, `cloudstack`, `fake`, `gce`, `mesos`, `openstack`, `ovirt`, `photon`, `rackspace`, `vsphere`

###  cluster.cloud_provider.config

 Path to the cloud provider config file. This will be copied to all the machines in the cluster 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  docker

 Configuration for the docker engine installed by KET 

###  docker.storage

 Storage configuration for the docker engine 

###  docker.storage.direct_lvm

 DirectLVM is the configuration required for setting up device mapper in direct-lvm mode 

###  docker.storage.direct_lvm.enabled

 Whether the direct_lvm mode of the devicemapper storage driver should be enabled. When set to true, a dedicated block storage device must be available on each cluster node. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  docker.storage.direct_lvm.block_device

 The path to the block storage device that will be used by the devicemapper storage driver. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  docker.storage.direct_lvm.enable_deferred_deletion

 Whether deferred deletion should be enabled when using devicemapper in direct_lvm mode. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

##  docker_registry

 Docker registry configuration 

###  docker_registry.setup_internal

 Whether an internal docker registry should be installed on the cluster. When set to true, a registry will be deployed on the first master node. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  docker_registry.address

 The hostname or IP address of a private container image registry. When performing a disconnected installation, this registry will be used to fetch all the required container images. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  docker_registry.port

 The port on which the private container image registry is listening on. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  No |
| **Default** | ` ` | 

###  docker_registry.CA

 The absolute path of the Certificate Authority that should be installed on all cluster nodes that have a docker daemon. This is required to establish trust between the daemons and the private registry when the registry is using a self-signed certificate. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  add_ons

 Add on configuration 

###  add_ons.cni

 The Container Networking Interface (CNI) add-on configuration. 

###  add_ons.cni.disable

 Whether the CNI add-on is disabled. When set to true, CNI will not be installed on the cluster. Furthermore, the smoke test and any validation that depends on a functional pod network will be skipped. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.cni.provider

 The CNI provider that should be installed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | `calico` | 
| **Options** |  `calico`, `weave`, `contiv`, `custom`

###  add_ons.cni.options

 The CNI options that can be configured for each CNI provider. 

###  add_ons.cni.options.calico

 The options that can be configured for the Calico CNI provider. 

###  add_ons.cni.options.calico.mode

 The datapath technique that should be configured in Calico. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | `overlay` | 
| **Options** |  `overlay`, `routed`

###  add_ons.dns

 The DNS add-on configuration. 

###  add_ons.dns.disable

 Whether the DNS add-on should be disabled. When set to true, no DNS solution will be deployed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.heapster

 The Heapster Monitoring add-on configuration. 

###  add_ons.heapster.disable

 Whether the Heapster add-on should be disabled. When set to true, Heapster and InfluxDB will not be deployed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.heapster.options

 The options that can be configured for the Heapster add-on 

###  add_ons.heapster.options.heapster

 The Heapster configuration options. 

###  add_ons.heapster.options.heapster.replicas

 Number of Heapster replicas that should be scheduled on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  No |
| **Default** | `2` | 

###  add_ons.heapster.options.heapster.service_type

 Kubernetes service type of the Heapster service. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | `ClusterIP` | 
| **Options** |  `ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName`

###  add_ons.heapster.options.heapster.sink

 URL of the backend store that will be used as the Heapster sink. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | `influxdb:http://heapster-influxdb.kube-system.svc:8086` | 

###  add_ons.heapster.options.influxdb

 The InfluxDB configuration options. 

###  add_ons.heapster.options.influxdb.pvc_name

 Name of the Persistent Volume Claim that will be used by InfluxDB. This PVC must be created after the installation. If not set, InfluxDB will be configured with ephemeral storage. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  add_ons.heapster.options.heapster_replicas _(deprecated)_

 Number of Heapster replicas that should be scheduled on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  No |
| **Default** | ` ` | 

###  add_ons.heapster.options.influxdb_pvc_name _(deprecated)_

 Name of the Persistent Volume Claim that will be used by InfluxDB. When set, this PVC must be created after the installation. If not set, InfluxDB will be configured with ephemeral storage. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

###  add_ons.dashboard

 The Dashboard add-on configuration. 

###  add_ons.dashboard.disable

 Whether the dashboard add-on should be disabled. When set to true, the Kubernetes Dashboard will not be installed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.dashbard _(deprecated)_

 The Dashboard add-on configuration. 

###  add_ons.dashbard.disable

 Whether the dashboard add-on should be disabled. When set to true, the Kubernetes Dashboard will not be installed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.package_manager

 The PackageManager add-on configuration. 

###  add_ons.package_manager.disable

 Whether the package manager add-on should be disabled. When set to true, the package manager will not be installed on the cluster. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

###  add_ons.package_manager.provider

 This property indicates the package manager provider. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 
| **Options** |  `helm`

##  features _(deprecated)_

 Feature configuration 

###  features.package_manager _(deprecated)_

 The PackageManager feature configuration. 

###  features.package_manager.enabled _(deprecated)_

 Whether the package manager add-on should be enabled. 

| | |
|----------|-----------------|
| **Kind** |  bool |
| **Required** |  No |
| **Default** | `false` | 

##  etcd

 Etcd nodes of the cluster 

###  etcd.expected_count

 Number of nodes. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  etcd.nodes

 List of nodes. 

###  etcd.nodes.host

 The hostname of the node. The hostname is verified in the validation phase of the installation. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  etcd.nodes.ip

 The IP address of the node. This is the IP address that will be used to connect to the node over SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  etcd.nodes.internalip

 The internal (or private) IP address of the node. If set, this IP will be used when configuring cluster components. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  master

 Master nodes of the cluster 

###  master.expected_count

 Number of master nodes that are part of the cluster. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  master.load_balanced_fqdn

 The FQDN of the load balancer that is fronting multiple master nodes. In the case where there is only one master node, this can be set to the IP address of the master node. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  master.load_balanced_short_name

 The short name of the load balancer that is fronting multiple master nodes. In the case where there is only one master node, this can be set to the IP address of the master nodes. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  master.nodes

 List of master nodes that are part of the cluster. 

###  master.nodes.host

 The hostname of the node. The hostname is verified in the validation phase of the installation. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  master.nodes.ip

 The IP address of the node. This is the IP address that will be used to connect to the node over SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  master.nodes.internalip

 The internal (or private) IP address of the node. If set, this IP will be used when configuring cluster components. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  worker

 Worker nodes of the cluster 

###  worker.expected_count

 Number of nodes. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  worker.nodes

 List of nodes. 

###  worker.nodes.host

 The hostname of the node. The hostname is verified in the validation phase of the installation. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  worker.nodes.ip

 The IP address of the node. This is the IP address that will be used to connect to the node over SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  worker.nodes.internalip

 The internal (or private) IP address of the node. If set, this IP will be used when configuring cluster components. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  ingress

 Ingress nodes of the cluster 

###  ingress.expected_count

 Number of nodes. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  ingress.nodes

 List of nodes. 

###  ingress.nodes.host

 The hostname of the node. The hostname is verified in the validation phase of the installation. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  ingress.nodes.ip

 The IP address of the node. This is the IP address that will be used to connect to the node over SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  ingress.nodes.internalip

 The internal (or private) IP address of the node. If set, this IP will be used when configuring cluster components. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  storage

 Storage nodes of the cluster. 

###  storage.expected_count

 Number of nodes. 

| | |
|----------|-----------------|
| **Kind** |  int |
| **Required** |  Yes |
| **Default** | ` ` | 

###  storage.nodes

 List of nodes. 

###  storage.nodes.host

 The hostname of the node. The hostname is verified in the validation phase of the installation. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  storage.nodes.ip

 The IP address of the node. This is the IP address that will be used to connect to the node over SSH. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  storage.nodes.internalip

 The internal (or private) IP address of the node. If set, this IP will be used when configuring cluster components. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  No |
| **Default** | ` ` | 

##  nfs

 NFS volumes of the cluster. 

###  nfs.nfs_volume

 List of NFS volumes that should be attached to the cluster during the installation. 

###  nfs.nfs_volume.nfs_host

 The hostname or IP of the NFS volume. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

###  nfs.nfs_volume.mount_path

 The path where the NFS volume should be mounted. 

| | |
|----------|-----------------|
| **Kind** |  string |
| **Required** |  Yes |
| **Default** | ` ` | 

