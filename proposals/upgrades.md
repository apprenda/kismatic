# Cluster Upgrade

Status: Implemented.

This document describes the initial implementation of upgrades in Kismatic. An upgrade
is defined as the replacement of _binaries_ or _configuration_ of a cluster created by Kismatic. 
An upgrade does not include the operating system, or any packages that are not managed by Kismatic.

This upgrade implementation is concerned with upgrading the following:
* Etcd clusters (Kubernetes and Networking)
* Kubernetes components
* Docker (if we decide to support a newer version)
* Calico
* On-Cluster services (e.g. DNS, Dashboard, etc)

## Versions
Every component on a cluster has a *current* version, and may be transitioning to 
exactly one *target version.*

Every cluster has components with some number of current versions, and may be transitioning 
to exactly one target version. The cluster is said to be “at version X” only 
if all components are at that version. There may be operations that aren’t 
performed on a cluster in transition because of the complexity of applying cluster-level 
decisions to a system in an unknown state.

Each Kismatic release has a single target version. We will attempt to support { some number } 
of Kismatic versions back in time.

## Can I add a feature during an upgrade?
No. You need a cluster at a consistent, known version to add a feature.

## Safety
Safety is the first concern of upgrading Kubernetes. An unsafe upgrade is one that results in 
loss of data or critical functionality.

Kubernetes does not have a concept of a stable workload installation, but by default it won’t 
move pods unless there is a problem with them. This relative stability means it’s possible to 
use Kubernetes to stand up workloads in configurations that aren’t at all safe, 
such as a database accepting writes that’s running in a single pod with emptyDir.

It is not Kismatic’s responsibility to fix these workloads. Also, it is not okay for us to 
identify that a workload is unsafe and perform an action that would cause it to lose data, 
which could be as simple as disconnecting Kubelet long enough that Kubernetes re-schedules the pod.

## Availability
Availability is the second concern of upgrading Kubernetes. An upgrade interrupts 
cluster availability if it results in the loss of a global cluster function 
(such as removing the last master, ingress or breaking etcd cluster quorum) 
and it interrupts workload availability if it results in the reduction of a service to 0 active pods.

## Upgrade Safety and Availability
The cluster operator will be able to choose between two upgrade modalities. 

### Offline upgrade
The offline upgrade is the most basic upgrade supported by Kismatic. In this mode, the cluster
will be upgraded regardless of potential safety or availability issues. More specifically,
Kismatic will not perform any safety or availability checks before performing the upgrade, nor will it
cordon or drain nodes during the upgrade. This is suitable for clusters that are not housing production
workloads.

### Online upgrade
The online upgrade is gated by safety and availability checks. In this mode, Kismatic will report 
any upgrade condition that is potentially unsafe or likely to cause a loss of availability. 
When faced with this situation, the upgrade will not proceed.

Operators may address the safety and availability concerns by:
* Manually scaling out nodes or pods, where applicable
* Manually scaling down or removing the unsafe workload
* Forcing the upgrade by using the offline modality

The following table outlines the conditions Kismatic will use to determine upgrade
safety and availability.

| Detected condition                         | Reasoning                                                                 |
|--------------------------------------------|---------------------------------------------------------------------------|
| Pod not managed by RC, RS,  Job, DS, or SS | Potentially unsafe: unmanaged pod will not be rescheduled                 |
| Pods without peers (i.e. replicas = 1)     | Potentially unavailable: singleton pod will be unavailable during upgrade |
| DaemonSet scheduled on a single node       | Potentially unavailable: singleton pod will be unavailable during upgrade |
| Pod using EmptyDir volume                  | Potentially unsafe: pod will loose the data in this volume                |
| Pod using HostPath volume                  | Potentially unsafe: pod will loose the data in this volume                |
| Pod using HostPath persistent volume       | Potentially unsafe: pod will loose the data in this volume                |
| Master node in a cluster with <2 masters   | Unavailable: upgrading the master will bring the control plane down       |
| Worker node in a cluster with <2 workers   | Unavailable: upgrading the worker will bring all workloads down           |
| Ingress node                               | Unavailable: we can't ensure that ingress nodes are load balanced         |
| Gluster node                               | Potentially unavailable: brick on node will become unavailable            |

## Readiness
Validation (aka. Preflight) during an upgrade is about node readiness. In other words,
validation is about answering the question: Can the node be expected to safely install
the new software and configuration?

The following checks are performed on each node to determine readiness:
1. Disk space: Ensure that there is enough disk space on the node for upgrading.
2. Packages: When package installation is disabled, ensure that the new packages are installed.

## Order of upgrade
All etcd nodes
Then
All master nodes
Then
All worker nodes (regardless of specialization)
Then
“On-cluster” Docker Registry
Then
Other on-cluster systems (DNS, dashboard, etc)

Nodes are upgraded one node at a time.

# Partial upgrade
Both the offline and online upgrade modalities will allow for partially upgrading a cluster.
A partial upgrade involves upgrading the nodes that did not report a problem. This enables 
the ability for an operator to upgrade part of the cluster online and to upgrade the rest of 
the cluster in a smaller downtime window.

## User Experience
```
kismatic info [-f planfile]
```
Prints the version of the cluster and all nodes.

```
kismatic upgrade online [-f plan]
```
Computes the upgrade plan for a cluster, detecting nodes NOT at the target version. 
Checks node readiness for upgrade. If no unsafe/unavailable workloads are detected, 
the plan is immediately executed. If any nodes are unready, unsafe or unavailable, 
the command will print and quit.

```
kismatic upgrade online [-f plan] --partial-ok
```
Computes the upgrade plan for a cluster, detecting nodes NOT at the target version. 
Checks node readiness for upgrade. All unready/unsafe/unavailable workloads detected are pruned 
from the plan. It is then immediately executed. All unready, unsafe or unavailable nodes are printed, 
then the command will quit.

```
kismatic upgrade offline [-f plan]
```
Computes the upgrade plan for a cluster, detecting nodes NOT at the target version. 
Checks node readiness for upgrade. If no unready workloads are detected, the plan is 
immediately executed. If any nodes are unready, the command will print and quit.

```
kismatic upgrade offline [-f plan] --partial-ok
```
Computes the upgrade plan for a cluster, detecting nodes NOT at the target version. 
Checks node readiness for upgrade. All unready nodes are pruned from the plan. 
It is then immediately executed. All unready nodes are printed, then the command will quit.
