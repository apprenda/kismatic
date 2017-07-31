# Troubleshooting

This document outlines troubleshooting steps for specific issues that may arise
when setting up a Kubernetes cluster using Kismatic.

- [Timed out waiting for control plane component to start up](#timed-out-waiting-for-control-plane-component-to-start-up)
- [Timed out waiting for Calico to start up](#timed-out-waiting-for-calico-to-start-up)
- [Timed out waiting for DNS to start up](#timed-out-waiting-for-dns-to-start-up)
- [Failure during installation](#failure-during-installation)

## Timed out waiting for control plane component to start up
The Kubernetes control plane components are deployed inside Kubernetes itself as 
static pods on each master node. Due to the asynchronous nature of deploying workloads
with Kubernetes, we must wait until the pods start up before we can continue with the installation. 
Kismatic is configured to wait a specific amount of time for these pods to start up.
If the pods don't start up in time, a timeout error is reported.

Hitting this timeout usually means there is an underlying issue with the cluster. However,
slow infrastructure or network might also be the culprit.

### Getting Control Plane logs
If the API server is running, you may use `kubectl` to get logs and troubleshoot:
1. Find the component that is failing: `kubectl get pods -n kube-system`
1. Get the logs for the component: `kubectl logs -n kube-system kube-controller-manager-node001`

If the API server is not running, you must use `docker` to get the container logs directly:
1. Find the container that is failing: `docker ps -a`
1. Get the logs for the container: `docker logs $CONTAINER_ID`

## Timed out waiting for Calico to start up
Calico is deployed as a daemonset on the cluster. For this reason, we must wait until all
pods start up successfully on all nodes. Kismatic will wait a specific amount of time for the pod
before it reports a timeout error. Hitting this timeout might mean there is an underlying networking issue with the cluster. 

### Getting Calico logs
Each Calico component logs to a specific directory on the nodes:
* Felix: `/var/log/calico/felix.log`
* BIRD: `/var/log/calico/bird/current`
* confd: `/var/log/calico/confd/current`

You may also use `kubectl` to get logs for the `calico-node` pod.

## Timed out waiting for DNS to start up
The cluster DNS workload is deployed as a pod on the cluster. For this reason, we must wait
until the pod starts up successfully. Kismatic will wait a specific amount of time for the pod
before it reports a timeout error. Hitting this timeout might mean there is an underlying networking issue with the cluster.

### Getting DNS logs
1. Find the DNS pod that is failing: `kubectl get pods -n kube-system`
1. The DNS pod is made up of three containers. You will most likely be interested in
the logs from the `kube-dns` container: `kubectl logs -n kube-system $KUBE_DNS_POD_NAME kube-dns`

## Failure during installation
Kismatic keeps a record for all command executions in the `runs` directory.
Inside `runs`, kismatic creates subdirectories that map to actions performed. 
For example, when running `kismatic install apply`, 
an `apply` directory is created in the `runs`, and a timestamped directory is created inside `runs/apply`
for each execution of the command.

```
ls -l runs/apply
total 0
drwxr-xr-x  6 root  root  204 Mar 15 15:06 2017-03-15-15-06-23
drwxr-xr-x  6 root  root  204 Mar 15 15:07 2017-03-15-15-07-05
drwxr-xr-x  6 root  root  204 Mar 15 15:07 2017-03-15-15-07-10
drwxr-xr-x  6 root  root  204 Mar 15 15:07 2017-03-15-15-07-35
drwxr-xr-x  6 root  root  204 Mar 15 15:08 2017-03-15-15-08-08
drwxr-xr-x  6 root  root  204 Mar 15 15:09 2017-03-15-15-09-00
drwxr-xr-x  6 root  root  204 Mar 15 15:10 2017-03-15-15-10-59
```

Each of these directories contains the following files:
* ansible.log: Verbose ansible logs
* clustercatalog.yaml: Listing of all variables passed to ansible
* inventory.ini: The ansible inventory that was generated from the plan file
* kismatic-cluster.yaml: The plan file that was used in the execution
