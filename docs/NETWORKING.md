# Kubernetes Networking
Kismatic installs [Calico](https://www.projectcalico.org/) as the networking solution for the cluster.
Detailed documentation for Calico may be found [here](http://docs.projectcalico.org/).

The main Calico components are deployed as static pods on all nodes except `etcd` nodes.

## Calicoctl
Calicoctl is the command-line utility for managing the Calico network.

If you need to troubleshoot calico, using calicoctl will be useful. This is
a quick command that you can use to run calicoctl:
```
docker run -i \
    --net host \
    -v /etc/kubernetes:/etc/kubernetes \
    -v /etc/calico/calicoctl.cfg:/etc/calico/calicoctl.cfg \
    calico/ctl:v1.1.0
```

You may find calicoctl's reference guide [here](http://docs.projectcalico.org/v2.1/reference/calicoctl/)

## Network Policy (Advanced Feature)
Calico supports defining Network Policy in your cluster.

More detailed documentation on policy can be found [here](http://docs.projectcalico.org/v2.1/getting-started/kubernetes/tutorials/simple-policy)

## Useful links
* Kubernetes + Calico overview: http://docs.projectcalico.org/v2.1/getting-started/kubernetes/
* Troubleshooting: http://docs.projectcalico.org/v2.1/getting-started/kubernetes/troubleshooting
* Logging: http://docs.projectcalico.org/v2.1/usage/troubleshooting/logging
