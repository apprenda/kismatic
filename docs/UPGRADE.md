# Upgrade

In order to keep your Kubernetes cluster up to date, and to take advantage of new
features in Kismatic, you may upgrade an existing cluster that was previously
stood up using Kismatic. The upgrade functionality is available through the
`kismatic upgrade` command.

The upgrade process is applied to each node, one node at a time. If a private docker registry
is being used, the new container images will be pushed by Kismatic before starting to upgrade
nodes.

Nodes in the cluster are upgraded in the following order:

1. Etcd nodes
2. Master nodes
3. Worker nodes (regardless of specialization)

It is important to keep in mind that if a node has multiple roles, all components will be upgraded.
For example, if we are in the process of upgrading etcd nodes, and a node is both an etcd node and
a master node, both the etcd and kubernetes master components will be upgraded in a single pass.

Cluster level services are upgraded once all nodes have been successfully upgraded.
Cluster services include the pod network provider (e.g. Calico), Dashboard, cluster DNS, etc.

**Upgrade pre-requisites**:
- Plan file used to install existing cluster
- Generated assets directory ("generated")
- SSH access to cluster nodes
- Cluster in a healthy state

## Readiness
Before performing an upgrade, Kismatic ensures that the nodes are ready to be upgraded.
The following checks are performed on each node to determine readiness:

1. Disk space: Ensure that there is enough disk space on the root drive of the node.
2. Packages: When package installation is disabled, ensure that the new packages are installed.

## Online Upgrade
With the goal of preventing workload data or availability loss, you might opt for doing
an online upgrade. In this mode, Kismatic will run safety and availability checks (see table below) against the
existing cluster before performing the upgrade. If any unsafe condition is detected, a report will
be printed, and the upgrade will not proceed.

Once all nodes are deemed ready for upgrade, it will proceed one node at a time.
If the node under upgrade is a Kubernetes node, it is cordoned and drained of workloads
before any changes are applied. In order to prevent workloads from being forcefully killed,
it is important that they handle termination signals to perform any clean up if required.
Once the node has been upgraded successfully, it is uncordoned and reinstated to the pool
of available nodes.

To perform an online upgrade, use the `kismatic upgrade online` command.

### Safety
Safety is the first concern of upgrading Kubernetes. An unsafe upgrade is one that results in
loss of data or critical functionality, or the potential for this loss.
For example, upgrading a node that hosts a pod which writes to an empty dir volume is considered unsafe.

### Availability
Availability is the second concern of upgrading Kubernetes. An upgrade interrupts
**cluster availability** if it results in the loss of a global cluster function
(such as removing the last master, ingress or breaking etcd quorum). An upgrade
interrupts **workload availability** if it results in the reduction of a service
to 0 active pods.

### Safety and Availability checks
The following list contains the conditions that are checked during an online upgrade, and the reason
why the upgrade is blocked if the condition is detected.

| Condition                                  | Reasoning                                                                 |
|--------------------------------------------|---------------------------------------------------------------------------|
| Pod not managed by RC, RS,  Job, DS, or SS | Potentially unsafe: unmanaged pod will not be rescheduled                 |
| Pods without peers (i.e. replicas = 1)     | Potentially unavailable: singleton pod will be unavailable during upgrade |
| DaemonSet scheduled on a single node       | Potentially unavailable: singleton pod will be unavailable during upgrade |
| Pod using EmptyDir volume                  | Potentially unsafe: pod will loose the data in this volume                |
| Pod using HostPath volume                  | Potentially unsafe: pod will loose the data in this volume                |
| Pod using HostPath persistent volume       | Potentially unsafe: pod will loose the data in this volume                |
| Master node in a cluster with < 2 masters  | Unavailable: upgrading the master will bring the control plane down       |
| Worker node in a cluster with < 2 workers  | Unavailable: upgrading the worker will bring all workloads down           |
| Ingress node                               | Unavailable: we can't ensure that ingress nodes are load balanced         |
| Storage node                               | Potentially unavailable: brick on node will become unavailable            |


## Offline Upgrade
The offline upgrade is available for those clusters in which safety and availabilty are not a concern.
In this mode, the safety and availability checks will not be performed, nor will the nodes in the cluster
be drained of workloads.

Performing an offline upgrade could result in loss of critical data and reduced service
availability. For this reason, this method should not be used for clusters that are housing
production workloads.

To perform an offline upgrade, use the `kismatic upgrade offline` command.

## Partial Upgrade
Kismatic is able to perform a partial upgrade, in which the subset of nodes that
reported readiness, safety or availability problems are not upgraded. A partial upgrade
can only be performed when all etcd and master nodes are ready for upgrading. In other words,
performing a partial upgrade is not supported if any etcd or master node reports issues.

The partial upgrade can be useful in the case where addressing these problems is not feasible. 
For example, one could decide to upgrade most of the nodes under an online upgrade, and then schedule
a downtime window for upgrading the rest of the nodes under an offline upgrade.

This mode can be enabled in both the online and offline upgrades by using the `--partial-ok` flag.

## Etcd upgrade
The etcd clusters should be backed up before performing an upgrade. Even though Kismatic will 
backup the clusters during an upgrade, it is recommended that you perform and maintain your own backups.
If you don't have an automated backup solution in place, it is recommended that you perform a manual backup of 
both the Kubernetes and networking etcd clusters before upgrading your cluster, and store 
the backup on persistent storage off cluster.

Kismatic will backup the etcd data before performing an upgrade. If necessary, you may find the
backups in the following locations:

* Kubernetes etcd cluster: `/etc/etcd_k8s/backup/$timestamp`
* Networking etcd cluster: `/etc/etcd_networking/backup/$timestamp`

For safety reasons, Kismatic does not remove the backups after the cluster has been
successfully upgraded.

## Version-specific notes
The following list contains links to upgrade notes that are specific to a given
Kismatic version.

- [Kismatic v1.3.0](./upgrade/v1.3.0)