# Windows Nodes

Add support for configuring Windows worker nodes with Kismatic alongside Linux nodes.

## Motivation

Windows Server Containers on Kubernetes is a Beta feature in Kubernetes v1.9.

Kubernetes version 1.5 introduced Alpha support for Windows Server Containers based on the Windows Server 2016 operating system. With the release of Windows Server version 1709 and Kubernetes v1.9 users are able to deploy a Kubernetes cluster that contain both Linux and Windows nodes.

Apprenda has been leading the development of Windows support in Kubernetes. Apprenda has had over 10 years of experience developing a platform to run Winwods applications and is bringing that experience to Kubernetes. It is only natural for Kismatic, an open-source tool to manage Kubernetes clusters, developed by Apprenda to also support intallation on Windows nodes.

## Prerequisites

In Kubernetes version 1.9 or later, Windows Server Containers for Kubernetes are supported using the following:

* Kubernetes control plane running on existing Linux infrastructure (version `1.9` or later).
* Kubenet network plugin setup on the Linux nodes.
* Windows Server 2016 RTM or later. Windows Server version `1709` or later is preferred; it unlocks key capabilities like shared network namespace.
* Docker Version `17.06.1-ee-2` or later for Windows Server nodes (Linux nodes and Kubernetes control plane can run any Kubernetes supported Docker Version).

## Technical Challenges

* Up to this point Kismatic has only installed and configured Linux nodes. As the tool uses Ansible to perform the node mutation, we will need to determine Ansible support on Windows as many modules are Linux specific.
* Many CNI solutions that work on Linux nodes require a simple `kubectl apply -f` style deployment. Unfortunately much more configuration is required to have networking on Windows nodes.
* Both the pre-install `kismatic-inspector` tool and `kuberang` the post-install smoke test tool are designed to work on Linux, the projects will need to be modified and built to supprt Windows nodes.

## Plan File Changes

There should not be any changes to the plan file indicating that a worker node is running Windows. The installation of Kubernetes should be transparent to the user and they should not need to worry about treating Windows nodes in any special manner.

Because current CNI providers (Calico and Weave) are not supported on Windows, there will be additional option(s) that are supported. - TODO determine what that option is. [Available options](https://kubernetes.io/docs/getting-started-guides/windows/#networking)


## Implementation

### Validation

* Windows nodes will only function as `worker` nodes, if a user uses a Windows node as any other role the validation should fail.
* When Kismatic determines that any of the nodes are Windows nodes it will need to validate that a supported CNI provider was selected.
  * TODO consider if the plan file can contain an indicator so that validation can fail early before running Ansible.
* Modify `kismatic-inspector` to properly work on Windows

### Installation

* Install `docker`, `kubelet` and `kube-proxy` on Windows nodes
* Install a CNI provider

### Post Installation

* Test and modify `kuberang` the smoke-test tool to support Winwods worker nodes.

### Other Changes

Below changes should be worked on after adding support for a basic installation.

* Support `cluster.disable_package_installation`
* Support `cluster.disconnected_installation`
* Support `cluster.networking.update_hosts_files`
* Support `kubelet.option_overrides`
* Support `docker`
* Support `docker_registry`

## Links

* https://kubernetes.io/docs/getting-started-guides/windows/#networking
* https://github.com/apprenda/kubernetes-ovn-heterogeneous-cluster
* https://github.com/openvswitch/ovn-kubernetes
