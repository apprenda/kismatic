# Centralized Logging 

Status: Proposal

## Overview
Centralized logging is one of the key components of a production cluster. Operators
and application developers can search and monitor all logs from a single place. With the goal of 
providing a logging experience out of the box, kismatic will deploy a centralized
logging pipeline based on the EFK stack.

## Use Cases
* As a cluster operator, I want to view and analyze logs from the kubernetes components
* As a developer, I want to view and analyze logs from the services I deploy to the cluster

## Components
### Fluentd
Fluentd is the log collector that will serve as the pipeline's entrypoint. It is 
deployed as a daemonset on all nodes of the cluster, configured with specific volume mounts
that allow fluentd to grab logs from the filesystem and ship them to elastic search.

Fluentd is built with a pluggable architecture that allows users to customize its
functionality. The community maintains a kubernetes _filter_ plugin for enriching the fluentd
events with kubernetes metadata such as pod ID, pod name, namepsace, etc. The plugin
can be found [here](https://github.com/fabric8io/fluent-plugin-kubernetes_metadata_filter).
This plugin will be included in the deployment of fluentd on the cluster.

### Elasticsearch
Elasticsearch is the sink for all the logs that are being shipped by fluentd. Storage
is the most critical consideration that must be made. In the simplest deployment,
an `emptyDir` pod volume will be used, which means the data is basically not persistent.

When setting up a more advanced scenario, the user will be responsible for providing 
a persistent volume claim name that will be used in the deployment descriptors. It is the
user's responsibility to configure storage according to their needs and create 
the expected PVC.

### Kibana
Kibana is the data visualization dashboard for viewing the logs stored in elasticsearch. 
Kibana will be deployed as a service on the cluster, and will be accessible to anyone
that has access to the cluster. Authentication or authorization is out of scope for this
proposal.

## Deployment
### Fluentd
The kubernetes team has historically maintained an add-on for `fluentd-elasticsearch`, 
including a container image for fluentd. However, it is unclear how much investment
is being made for keeping the image up to date with the fluentd project. For this reason,
the official fluentd container will be used for the deployment. Furthermore, 
the fluentd team maintains a [daemonset deployment](https://github.com/fluent/fluentd-kubernetes-daemonset)
that we can leverage as well.

### Elasticsearch
At the time of writing, there was an abandonded (last update was 3 months ago) helm 
chart in the [incubator](https://github.com/kubernetes/charts/tree/master/incubator/elasticsearch).
For this reason, Elasticsearch will be deployed on the cluster using the deployment technique maintained
by @pires [here](https://github.com/pires/kubernetes-elasticsearch-cluster), which seems
to be the one that the community has ralied around. There is an operator [available](https://github.com/upmc-enterprises/elasticsearch-operator) but it requires a cloud integration for provisioning storage volumes and is fairly new.

### Kibana
Kibana will be deployed on the cluster using a Deployment with a single replica.
The deployment can be scaled when deemed necessary by the user. A service will also
be deployed to expose Kibana to the user.

## Security
Kismatic has always taken a stance of "secure by default". All communications between
the log pipeline components will be secured using TLS:
* Fluentd -> Elasticsearch
* Fluentd -> Kubernetes API Server
* Kibana -> Elasticsearch

This will require new certificates to be generated during the cert generation phase.

## Feature in plan file
The centralized logging solution will be controlled by a feature in the plan file:
```
cluster:
...
features:
  logging:
    enabled: true|false
    elastic_search:
      persistent_volume_claim: "" # Name of persistent volume claim that will be provisioned after cluster is up.
```

## Log Rotation
Kismatic will configure the docker daemon to perform log rotation. This is inline with the 
approach taken by upstream Kubernetes ([PR](https://github.com/kubernetes/kubernetes/pull/40634)). This takes care of log rotation for applications and cluster components running in containers. Log
rotation will also be necessary for anything running outside of containers, such as
etcd, the kubelet, gluster services, etc. We might be able to use journald's log rotation
mechanism, or setup `logrotate` for these services.

Given that the logs are being shipped to a central data store, the log rotation parameters 
will not be exposed as options in the plan file. The maximum log file size will be 25mb 
and we will keep up to 5 log files in rotation. The `logrotate` script will be configured
to run daily on all nodes.