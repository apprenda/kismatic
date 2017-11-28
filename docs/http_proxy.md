# Working With Proxies

During the installation or upgrade of your cluster, KET must be able to pull
packages and container images from the Internet (or your local infrastructure if
performing a [disconnected installation](./disconnected_install.md)).

When your nodes are behind a proxy server, the proxy's information must be 
set in the [plan file](./plan-file-reference.md) so that KET can send requests
through the proxy.

When the proxy configuration is entered in the plan file, KET will do the following:
* Use the proxy when downloading software packages or container images
* Use the proxy when setting up Helm, as it needs to download Chart metadata
* Configure the docker daemon to use the proxy

Additionally, KET will always set the hostnames, IPs and internal IPs of all nodes
in the `no_proxy` environment variable. This is to prevent any node-to-node communication
from going through the proxy.

