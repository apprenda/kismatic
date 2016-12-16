# Docker Registry

By default the Kismatic tools is configured to use the public [Docker Hub](https://hub.docker.com/) when pulling images during installation.
If the cluster does not have access to the Internet or the Hub, there are 2 options available in the plan to accommodate file when performing the initial Kubernetes cluster setup with Kismatic.

## Self-Configured Docker Registry
```
# plan file
docker_registry:
    setup_internal: true
```
Setup a private Docker registry running on the first of the master nodes(from the plan file) on port `8443`.   
The registry is configured to use TLS certificates that are generated during the install, similar to the approach from the [official docs](https://docs.docker.com/registry/deploying/#/running-a-domain-registry).

The Docker daemons running on all of the nodes in the cluster are automatically configured to be able to pull from the registry,
however if you need to setup any other Docker daemon to push or pull Docker images to/from this registry you can do so by following the [self-signed certificate instructions](https://docs.docker.com/registry/insecure/#/using-self-signed-certificates).  

ie. `/etc/docker/certs.d/$FQDN:8443/ca.crt`, where `$FQDN` is `master.load_balanced_fqdn` from the plan file, and `ca.crt` is just renamed `ca.pem` file from the `generated/keys` directory

*No other information is required from you to setup this option*

**NOTE: this option is NOT suitable for production, if the master node dies it could lead to cluster issues or complete loss of all of your Docker images**

## User Configured Docker Registry
```
# plan file
docker_registry:                         
    setup_internal: false                
    address: myprivateregistry.com                               
    port: 8443              
    CA: /certs/ca.rt    
```
If you already have a Docker registry running and want to use that in the cluster, it is easiest to let the Kismatic tool configure all of the Docker daemons in that cluster.   

It is recommended to follow the [official docs](https://docs.docker.com/registry/deploying/#/running-a-domain-registry) when setting up your own docker registry.  

You will need to provide 3 pieces of information in the plan file:
* `address` the **reachable**(from all of the Kubernetes nodes) URL or IP where the Docker registry is running
* `port` number of the Docker registry
* `CA` that was used to sign the TLS certificates provided when starting the Docker registry

---
**If either option is selected, the Kismatic tool will attempt to push all of the images used during the install to the registry;
this allows for an installation that is completely self-contained and does not require an internet connection.**
