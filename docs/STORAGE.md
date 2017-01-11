# Kismatic and Persistent Storage

## Background

A container is, at its heart, just a process. When that process ends, whatever state it was managing ends along with it.

To be stateful, a container must write to a disk, and for that state to have any real value, the disk must something that another container can find and access again later, should the original process die.

This is obviously a challenge with a container cluster, where containers are starting and stopping and scaling all over the place.

Kubernetes offers three main techniques to tie stateful disks to containers: 

* **[Volumes](http://kubernetes.io/docs/user-guide/volumes/)**, in which a specific network share is created and described specifically in the specification of a Pod.
* Staticly provisioned **[Persistent Volumes](http://kubernetes.io/docs/user-guide/persistent-volumes/)**, in which a specific network share is created and described generically to Kubernetes. Specific volumes are "claimed" by cluster users for use by a single Kubernetes namespace along with a name. Any pod within this namespace may consume a claimed PV by its name. 
* Dynamic provisioned **[Persistent Volumes](http://kubernetes.io/docs/user-guide/persistent-volumes/)**, in which Kubernetes interfaces with a storage layer, building new network shares on demand.

To use Volumes, you need to have provisioned your storage with a specific use case in mind and will to push the storage details along with your pod spec. Since the implementation of storage is tied to the pod spec, the spec is less portable -- you can't run the same spec on another cluster.

To use staticly provisioned PersistentVolumes, you sill need to provision your own storage; however, this can be accomplished more generally. For example, a storage engineer might create a big batch 100GB shares with various performance characteristics and engineers on various teams could make claims against these PVs without having to relay their specific needs to the storage team ahead of time.

To use dynamically provisioned PersistentVolumes, you need to grant Kubernetes permission to make new storage volumes on your behalf. Any new claims that come in and aren't covered by a statically provisioned PV will result in a new volume being created for that claim. This is likely the ideal solution for the (functionally) infinite storage available in a public cloud -- however, for most private clouds, unbounded on-demand resource allocation is effectively a run on the bank. If you have a very large amount of available storage, or very small claims, a provisioner combined with a Resource Quota should serve your needs. Those with less elastic storage growth limits will likely have more success with occasional static provisioning, and Kismatic is currently focused on making this easier.

## Storage options for Kismatic-managed features

There are currently three storage options with Kismatic:

1. **None**. In this case, you won't be able to responsibly use any stateful features of Kismatic.
  * New shares may be added after creation using kubectl
2. **Bring-your-own NFS shares**. When building a Kubernetes cluster with Kismatic, you will first provision one or more NFS shares on an off-cluster file server or SAN, open access to these shares from the Kubernetes network and provide their details.
  * New shares may be added after creation using kubectl
  * Only multi-reader, multi-writer NFSv3 volumes are supported
3. **Kismatic manages a storage cluster**. When building a Kubernetes cluster with Kismatic, you will identify machines that will be used as part of a storage cluster. These may be dedicated to the task of storage or may duplicate other cluster roles, such as Worker and Ingress. 
  * Kismatic will automatically create and claim one replicated NFS share of 10 GB automatically the first time a stateful feature is included on a storage cluster. Future features will use this same volume.
  * The addition of future shares on this storage cluster is left up to cluster operators. A single command, such as `kismatic storage new-volume -r2 10G storage01`, can be used to provision a new storage volume and also add that volume to Kubernetes as an unclaimed PersistentVolume.
  * The storage cluster will be set up using GlusterFS
