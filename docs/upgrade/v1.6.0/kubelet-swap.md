# Kubelet and Swap

Starting with Kubernetes v1.8.0, the Kubelet will fail to start up if the nodes
have swap memory enabled. Discussion around why swap is not supported can be 
found in [this issue](https://github.com/kubernetes/kubernetes/issues/7294).

Before performing an installation, you must disable swap memory on your nodes. 
If you want to run with swap memory enabled, you can override the Kubelet 
configuration in the plan file.

If you are performing an upgrade and you have swap enabled, you will have to
decide whether you want to disable swap on all your nodes. If not, you must
override the kubelet configuration to allow swap.

## Override Kubelet Configuration
If you want to run your cluster nodes with swap memory enabled, you can override
the Kubelet configuration in the plan file:
```
cluster:
  # ...
  kubelet:
    option_overrides:
      fail-swap-on: false
```
