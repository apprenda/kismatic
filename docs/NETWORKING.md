# Kubernetes Networking
Kismatic installs [Calico](https://www.projectcalico.org/) as the networking solution for the cluster.
Detailed documentation for Calico may be found [here](http://docs.projectcalico.org/).

The main Calico components are running on each master and worker node of the cluster. The components themselves are
running inside a container called `calico-node`.

## Calicoctl
Calicoctl is the command-line utility for managing the Calico network. This utility is installed on all master nodes of your cluster.

You may find calicoctl's reference guide [here](http://docs.projectcalico.org/v1.6/reference/calicoctl/)

## Network Policy (Advanced Feature)
Calico supports defining Network Policy in your cluster. This is an advanced feature that is disabled by default in
the installation plan file.

More detailed documentation on policy can be found [here](http://docs.projectcalico.org/v1.6/getting-started/kubernetes/tutorials/simple-policy)

## Useful links
* Kubernetes + Calico overview: http://docs.projectcalico.org/v1.6/getting-started/kubernetes/
* Troubleshooting: http://docs.projectcalico.org/v1.6/getting-started/kubernetes/troubleshooting
* Logging: http://docs.projectcalico.org/v1.6/usage/troubleshooting/logging
