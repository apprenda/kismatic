# Persistent Storage

## Background

A container is, at its heart, just a process. When that process ends, whatever state it was managing ends along with it.

To be stateful, a container must write to a disk, and for that state to have any real value, the disk must something that another container can find and access again later, should the original process die.

This is obviously a challenge with a container cluster, where containers are starting and stopping and scaling all over the place.

Kubernetes offers three main techniques to tie stateful disks to containers:

* Pod **[Volumes](http://kubernetes.io/docs/user-guide/volumes/)**, in which a specific network share is created and described specifically in the specification of a Pod.
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
  * The addition of future shares on this storage cluster is left up to cluster operators. A single command, such as `kismatic volume add 10 storage01`, can be used to provision a new storage volume and also add that volume to Kubernetes as an unclaimed PersistentVolume.
  * The storage cluster will be set up using GlusterFS

## Using GlusterFS storage cluster for your workloads

1. Use Kismatic to configure `nodes` for GlusterFS by providing their details in the plan file, ie.
   ```
   ...
   storage:
     expected_count: 2
     nodes:
     - host: storage1.somehost.com
       ip: 8.8.8.1
       internalip: 8.8.8.1
     - host: storage2.somehost.com
       ip: 8.8.8.1
       internalip: 8.8.8.1
   ```

 If you have an existing kubernetes cluster setup with Kismatic you can still have the tool configure a GlusterFS cluster by adding to your plan file similar to above and running:
   ```
   kismatic install step _storage.yaml
   ```

 This will setup a 2 node GlusterFS cluster and expose it as a kubernetes service with the name of `kismatic-storage`
2. Create a new GlusterFS volume and expose it in kubernetes as a PersistentVolume use:
   ```
   kismatic volume add 10 storage01 -r 2 -d 1 -c="durable" -a 10.10.*.*
   ```

  * `10` represents the volume size in GB, in this example a GlusterFS volume with a `10GB` quota and a kubernetes PersistentVolume with a capacity of `10Gi` will be created
  * `storage01` is the volume name used when creating the GlusterFS volume name, the GlusterFS brick directories and Kubernetes PersistentVolume name. All GlusterFS bricks will be created under the `/data` directory on the node, using the logical disk mounted under `/`
  * `-r (replica-count)` the number of duplicates to make of each file stored when writing data
  * `-d (distribution-count)` the degree to which files will be distributed across the cluster. A count of 1 means that all files will exist at every replica. A count of 2 means that each set of replicas will have half of all files.
  * **NOTE**: the GlusterFS cluster must have at least `replica-count * distribution-count` storage nodes available for a volume to be created. In this example the storage cluster must have 2 or more nodes with at least 10GB free disk space on each of the machines
  * `-c (storage-class)` the name of the StorageClass that will be added when creating the PersistentVolume. Use this name when creating your PersistentVolumeClaims.
  * `-a allow-address` is comma separated list of off-cluster IP ranges that are permitted to mount and access the GlusterFS network volumes. Include any addresses you use for data management. Nodes in the Kubernetes cluster and the pods CIDR range will always have access.
  * **NOTE**: IP address is the only credential used to authorize a storage connection. All nodes and pods will be able to access these shares.
3. Create a new PersistentVolumeClaim
   ```
   kind: PersistentVolumeClaim
   apiVersion: v1
   metadata:
     name: my-app-frontend-claim
     annotations:
       volume.beta.kubernetes.io/storage-class: "durable"
   spec:
     accessModes:
       - ReadWriteMany
     resources:
       requests:
         storage: 10Gi
   ```

 Use the `volume.beta.kubernetes.io/storage-class: "durable"` annotation for the PersistentVolumeClaim to bind to the newly created PersistentVolume

4. Use the claim as a pod volume
   ```
   kind: Pod
   apiVersion: v1
   metadata:
     name: my-app-frontend
   spec:
     containers:
       - name: my-app-frontend
         image: nginx
         volumeMounts:
         - mountPath: "/var/www/html"
           name: html
     volumes:
       - name: html
         persistentVolumeClaim:
           claimName: my-app-frontend-claim
   ```

5. Your pod will now have access to the `/var/www/html` directory that is backed by a GlusterFS volume. If you scale this pod out, each instance of the pod should have access to that directory.
