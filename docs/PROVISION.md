# Provision

## <a name="get"></a>Get the distribution

You will need to run the installer either from a Linux machine or from a Darwin (OSX) machine. This machine will need to be able to access via SSH all the machines that will become nodes of the Kubernetes cluster.

The installer can run from a machine that will become a node on the cluster, but since the installer's machine holds secrets (such as SSH/SSL keys and certificates), it's best to run from a machine with limited user access and an encrypted disk.

The machine the installer is run from should be available for future modifications to the cluster (adding and removing nodes, upgrading Kubernetes).

## Download the installer

### To unpack the installer from Linux

From an ssh session, type:

`curl -L https://github.com/apprenda/kismatic/releases/download/v1.2.0/kismatic-v1.2.0-linux-amd64.tar.gz | tar -zx`

### To unpack from Darwin (Mac OSX)

From a terminal window, type

`wget -O - https://github.com/apprenda/kismatic/releases/download/v1.2.0/kismatic-v1.2.0-darwin-amd64.tar.gz | tar -zx`

## Generate A Plan File

From the machine you installed Kismatic to, run the following:

`./kismatic install plan`

You will be asked a few questions regarding the decisions you made in the Plan section above. The kismatic installer will then generate a **kismatic-cluster.yaml** file.

As machines are being provisioned, you must record their identity and credentials in this file.

## Create Machines

### <a name="access"></a>Providing access to the Installer

Kismatic deploys packages on each node, so you will need a user with remote passwordless sudo access and an ssh public key added to each node. The same username and keypair must be used for all nodes. This account should only be used by the kismatic installer.

We suggest a default user **kismaticuser** via:
```
sudo useradd -d /home/kismaticuser -m kismaticuser
sudo passwd kismaticuser
```

The user can be given full, passwordless sudo privileges via:
```
echo "kismaticuser ALL = (root) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/kismaticuser
sudo chmod 0440 /etc/sudoers.d/kismaticuser
```

We also suggest a corresponding private key **kismaticuser.key** added to the directory you're running the installer from. This would be the ideal spot to generate the keypair via:

`ssh-keygen -t rsa -b 4096 -f kismaticuser.key -P ""`

The resulting **kismaticuser.pub** will need to be copied to each node. ssh-copy-id can be convenient for this, or you can simply copy its contents to ~/.ssh/idrsa

There are four pieces of information we will need to be able to address each node:

<table>
  <tr>
    <td><b>hostname</b></td>
    <td>This is a short name that machines can access each other with. If you opt for Kismatic to manage host files for your cluster, this name will be copied to host files.</td>
  </tr>
  <tr>
    <td><b>ip</b></td>
    <td>This is the ip that the installer should connect to your node with. If you don't specify a separate internal_ip for the node, the ip will be used for cluster traffic as well</td>
  </tr>
  <tr>
    <td><b>internal_ip</b><br/> (optional)</td>
    <td>In many cases nodes will have more than one physical network card or more than one ip address. Specifying an internal IP address allows you to route traffic over a specific network. It's best for Kubernetes components to communicate with each other via a local network, rather than over the internet.</td>
  </tr>
  <tr>
    <td><b>labels</b> <br/> (optional)</td>
    <td>With worker nodes, labels allow you to identify details of the hardware that you may want to be available to Kubernetes to aid in scheduling decisions. For example, if you have worker nodes with GPUs and worker nodes without, you may want to tag the nodes with GPUs.</td>
  </tr>
</table>


### Pre-Install Configuration

By default, Kismatic will attempt to install any of the software packages below if missing. This can cause the installer to take significantly more time and use more bandwidth than pre-baking an image and will require internet access to the Kismatic package repository, DockerHub and a repository for your operating system.

If you are building a large cluster or one that won't have access to these repositories, you will want to [cache the necessary packages](PACKAGES.md) in a repository on your network.

<table>
  <tr>
    <th>Requirement</th>
    <th>Required for</th>
    <th>etcd</th>
    <th>master</th>
    <th>worker</th>
  </tr>
  <tr>
    <td>user and public key installed on all nodes</td>
    <td>Access from kismatic to manage node</td>
    <td>yes</td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>/etc/ssh/sshd_config contains `PubkeyAuthentication yes`</td>
    <td>Access from kismatic to manage node</td>
    <td>yes</td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Access to an apt or yum repository</td>
    <td>Retrieving binaries over the internet during installation</td>
    <td>yes</td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Python 2.7</td>
    <td>Kismatic management of nodes</td>
    <td>yes</td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Kismatic package of Docker 1.11.2</td>
    <td>hosting containers</td>
    <td></td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Kismatic package of Calico 2.0.0</td>
    <td>inter-pod networking</td>
    <td></td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Kismatic package of Etcd 3.1.0</td>
    <td>inter-pod networking</td>
    <td>yes</td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Kismatic package of Kubernetes kubelet 1.5.2-4</td>
    <td>Kubernetes</td>
    <td></td>
    <td>yes</td>
    <td>yes</td>
  </tr>
  <tr>
    <td>Kismatic package of Kubernetes kubectl 1.5.2-4</td>
    <td>Kubernetes</td>
    <td></td>
    <td>yes</td>
    <td></td>
  </tr>
</table>

### Inspector

To double check that your nodes are fit for purpose, you can run the kismatic inspector. This tool will be run on each node as part of validating your cluster and network fitness prior to installation.

## Networking

Enter your network settings in the plan file, including

* pod networking technique (**routed** or **overlay**)
* CIDR ranges for pod and services networks
* whether the Kismatic installer should manage hosts files for your cluster

Create your DNS CNAME or load balancer alias for your Kubernetes master nodes based on their hostnames.
