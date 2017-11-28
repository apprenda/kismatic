# Etcd Nodes

## Action Required
Block storage device must be added to etcd nodes when using DeviceMapper storage
driver in direct-lvm mode.

## Summary

In previous versions, Etcd was installed on the nodes using an RPM or DEB package
that was maintained by the Kismatic team. As KET moves away from the Kismatic packages
to the upstream Kubernetes packages, KET is no longer able to install etcd using
these RPMs or DEBs.

Starting with KET v1.6.0, etcd is executed as a container on the etcd nodes, which 
means docker will be installed during the upgrade. If you are running docker
with the DeviceMapper storage driver in direct-lvm mode, you will need to add a
new block storage device that will be used for direct-lvm. Furthermore, the device
must be attached at the same location as every other node on the cluster.