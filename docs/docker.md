# Docker Configuration

Kismatic supports docker as the underlying container runtime for your cluster. During the installation of your cluster, 
Kismatic will download and install docker on all nodes except `etcd` nodes.

## Storage in RHEL/CentOS (KET v1.3.1+)
The default storage driver that gets installed on RHEL/CentOS is `devicemapper` in `loop-lvm` mode. However, 
this is not recommended for production setups. Instead, `devicemapper` must be setup in `direct-lvm` mode. 

The `direct-lvm` mode requires a block storage device that will be used exclusively for docker's storage needs.
For this reason, this is an opt-in feature that can be enabled in the plan file:

```
docker:
  storage:
    direct_lvm:                          # Configure devicemapper in direct-lvm mode (RHEL/CentOS only).
      enabled: false
      block_device: ""                   # Path to the block device that will be used for direct-lvm mode. This device will be wiped and used exclusively by docker.
      enable_deferred_deletion: false    # Set to true if you want to enable deferred deletion when using direct-lvm mode.
```

When `direct-lvm` is enabled, Kismatic will create a logical volume configured as a thin pool to use as the backing storage
for docker. Creating a thin pool logical volume requires a "data" logical volume and a "metadata" logical volume.

The provided block storage device will be split 95% for data, and 1% for metadata, as recommended by Docker. The remaining 
space allows for auto-extending the logical volumes when they start to fill up as a temporary stop gap. 
The threshold for auto-extension is set to 80% of the size of the volume.

For more information on `devicemapper` see https://docs.docker.com/engine/userguide/storagedriver/device-mapper-driver/#configure-direct-lvm-mode-for-production
