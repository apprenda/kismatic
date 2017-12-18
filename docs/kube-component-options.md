# Configuring Kubernetes Components

Under certain scenarios, users might want to set flags that are not set 
by default, or they might want to override the value of certain flags that are passed
to the Kubernetes components. 

Each Kubernetes component has a corresponding section in the plan file, through which
specific configuration flags can be set or overridden. There is a subset of protected flags 
that cannot be overridden, as they depend on configuration that is managed by KET.

When using this feature, you must keep in mind that an invalid configuration could
prevent the cluster from functioning properly.

## Configuring the API Server

The Kubernetes API Server options can be set or overridden in the plan file using the 
[cluster.kube_apiserver.option_overrides](./plan-file-reference.md#clusterkube_apiserveroption_overrides) field.

For example:
```
cluster:
...
  kube_apiserver:
    option_overrides:
      "audit-log-path": "/var/log/kube-apiserver.log"
      "event-ttl": "2h0m0s"
      "runtime-config": "batch/v2alpha1=true"
```

## Configuring the Controller Manager
The Kubernetes Controller Manager options can be set or overridden in the plan file 
using the [cluster.kube_controller_manager.option_overrides](./plan-file-reference.md#clusterkube_controller_manageroption_overrides) field.

## Configuring the Scheduler
The Kubernetes Scheduler options can be set or overridden in the plan file using the 
[cluster.kube_scheduler.option_overrides](./plan-file-reference.md#clusterkube_scheduleroption_overrides) field.

## Configuring the Kubelet
The Kubelet options can be set or overridden in the plan file using the 
[cluster.kubelet.option_overrides](./plan-file-reference.md#clusterkubeletoption_overrides) field.
This configuration is applied to all Kubelets in the cluster.

The Kubelet options can also be set at the node level using the `kubelet.option_overrides` 
field of each node. 

For example:
```
cluster:
  kubelet:
    option_overrides:
      max-pods: 50
...
...
worker:
  expected_count: 1
  nodes:
    host: my-worker
    ip: 10.0.20.1
    kubelet:
      option_overrides:
        max-pods: 120
```

## Configuring the Kube Proxy
The Kube Proxy options can be set or overridden in the plan file using the 
[cluster.kube_proxy.option_overrides](./plan-file-reference.md#clusterkube_proxyoption_overrides) field.
