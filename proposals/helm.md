# Helm

Status: Proposal

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
  * Use the included binary to run `helm init` (explore including it as a Go dependency instead)
* Include `tiller` docker images in the offline package
* Plan file option to disable *NOTE* This will block installation of any charts that KET would configure, ie logging and monitoring and any others in the future     
