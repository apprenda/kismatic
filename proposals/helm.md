# Helm

Status: Implemented

Helm is a tool for managing Kubernetes charts. Charts are packages of pre-configured Kubernetes resources.

Helm is in the kubernetes Github repository, has a large community support and seems to be the de facto tool to deploy complex common applications in a repeated way.  

Helm has two parts: a client (`helm`) and a server (`tiller`).
Helm can be installed by running `helm init`, this deploys `tiller` in the cluster and sets up the local `~/.helm/` directory on the machine where the command was run on.

Although current k8s clusters built with KET _support_ Helm and can be configured to run `tiller` with `helm init` there are benefits of configuring Helm as part of the initial cluster installation:
* Use helm to install monitoring and logging charts
* Allow a user to have a cluster with working helm already installed
* Deploy a predictable and tested version of helm on the cluster

# Required Changes
* Include the `helm` binary as part of the KET tar ball
* Create a new _phase_ to deploy Helm
  * Use the included binary to run `helm init`
  * `helm` will be executed from the install machine and not part of the regular install, the new UI would looks something like:
```
Installing Cluster==================================================================
Generating Kubeconfig File==========================================================
Installing Helm on the Cluster======================================================
Running Smoke Test==================================================================
```
* All helm charts for cluster components KET is installing will be included in the distribution and used by the installer. This allows to lock down and test specific versions.
* Include `tiller` docker images in the offline package
* Plan file option to disable *NOTE* This will block installation of any charts that KET would configure, ie logging and monitoring and any others in the future     
```
cluster:
    features:
      package_manager:
        enabled: true|false
        provider: helm
```
* If an existing `./helm` directory is detected, it will be backed up prior to running `helm init`
