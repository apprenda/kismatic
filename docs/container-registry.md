# Container Image Registry

By default, KET will pull the required container images from the internet during installation
and upgrade. If the cluster does not have access to the internet, or you would rather
use an internal container image registry, you must provide the details of the registry
in the plan file.


## Configuring KET
The following information must be provided in the plan file to use an internal
image registry:
* `server`: The hostname or IP address of the registry and the port. This must be reachable from
all the nodes in the cluster.
* `CA`: The absolute path to the certificate that should be trusted when connecting
to the registry. This is optional. When set, KET will configure the docker daemon
on all nodes to trust this certificate.

Sample:
```
# plan file
docker_registry:                         
    server: registry.example.com:8443        
    CA: /certs/ca.rt    
```

## Seeding a registry
Before being able to use an internal registry for installing or upgrading your cluster,
the required container images must be available in the registry.

KET provides the `seed-registry` command to seed the internal registry with the
required images. With this command, you can download, tag and push the required
images to your internal registry. Alternatively, you can obtain the list of images
and perform the seeding without KET.

In order to seed the registry with KET, your machine must:
* Have internet access
* Have docker installed
* Trust the registry's certificate
* Have enough disk space to pull all the required images

For more information about this command, see the [reference documentation](./kismatic-cli/kismatic_seed-registry.md)
or use `./kismatic seed-registry --help`. 
