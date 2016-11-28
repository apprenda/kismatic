## Plan

To get started with Kubernetes quickly, you can use Kismatic to stand up a small cluster in AWS or virtualized on a personal computer.

But setting up a proper cluster takes a little forethought. Depending on your intent, you may need to engage multiple teams within your organization to correctly provision the required infrastructure. Planning will also help you identify provisioning tasks and know what infromation will be needed to proceed with installation.

Planning focuses mainly on three areas of concern:

* The machines that will form a cluster
* The network the cluster will operate on
* Other services the cluster be interacting with

## <a name="compute"></a>Compute resources

<table>
  <tr>
    <td>Etcd Nodes <br />
Suggested: 3</td>
    <td>1      <b>3</b>     5     7</td>
  </tr>
  <tr>
    <td>Master Nodes <br />
Suggested: 2</td>
    <td>1      <b>2</b> </td>
  </tr>
</table>

Kubernetes is installed on multiple physical or virtual machines running Linux. These machines become **nodes** of the Kubernetes **cluster**.

In a Kismatic installation of Kubernetes, nodes are specialized to one of three distinct roles within the cluster: **etcd**, **master** or **worker**.

* etcd
  * These nodes provide data storage for the master.
* master
  * These nodes provide API endpoints and manage the Pods installed on workers.
* worker
  * These nodes are where your Pods are instantiated.

Nodes within a cluster should have latencies between them of 10ms or lower to prevent instability. If you would like to host workloads at multiple data centers, or in a hybrid cloud scenario, you should expect to set up at least one cluster in each geographically seperated region.

### Hardware & Operating System

Infrastructure supported:

* bare metal
* virtual machines
* AWS EC2
* Packet.net

If using VMs or IaaS, we suggest avoiding virtualization strategies that rely on the assignment of partial CPUs to your VM. This includes avoiding AWS T2 instances or CPU oversubscription with VMs.

Operating Systems supported:

* RHEL 7
* Centos 7
* Ubuntu 16.04

Minimum hardware requirements:

<table>
  <tr>
    <td>Node Role</td>
    <td>CPU</td>
    <td>Ram</td>
    <td>Disk (Prototyping<sup>1</sup>)</td>
    <td>Disk (Production<sup>1</sup>)</td>
  </tr>
  <tr>
    <td>etcd</td>
    <td>1 CPU Core, 2 GHz</td>
    <td>1 GB</td>
    <td>8 GB</td>
    <td>50 GB</td>
  </tr>
  <tr>
    <td>master</td>
    <td>1 CPU Cores, 2 GHz</td>
    <td>2 GB</td>
    <td>8 GB</td>
    <td>50 GB</td>
  </tr>
  <tr>
    <td>worker</td>
    <td>1 CPU Core, 2 GHz</td>
    <td>1 GB</td>
    <td>8 GB</td>
    <td>200 GB</td>
  </tr>
</table>

<sup>1</sup>A Prototype cluster is one you build for a short term use case (less than a week or so). It can have smaller drives, but you wouldn't want to run like this for extended use.

[Recommended Master sizing:](http://kubernetes.io/docs/admin/cluster-large/#size-of-master-and-master-components)

Worker Count | CPUs | RAM
---          | ---  | ---
< 5          | 1    | 3.75
< 10	       | 2	  | 7.5
< 100	       | 4	  | 15
< 250	       | 8	  | 30
< 500	       | 16	  | 30
< 1000	     | 32	  | 60

### Planning for etcd nodes:

Each etcd node receives all the data for a cluster to help protect against data loss in the event that something happens to one of the nodes. A Kubernetes cluster is able to operate as long as more than 50% of its etcd nodes are online. Always use an odd number of etcd nodes. Count of etcd nodes is primarily an availability concern, as adding etcd nodes can decrease Kubernetes performance.

<table>
  <tr>
    <td>Node Count</td>
    <td>Safe for</td>
  </tr>
  <tr>
    <td>1</td>
    <td>Unsafe. Use only for small development clusters</td>
  </tr>
  <tr>
    <td>3</td>
    <td>Failure of any one node</td>
  </tr>
  <tr>
    <td>5</td>
    <td>Simultaneous failure of two nodes</td>
  </tr>
  <tr>
    <td>7</td>
    <td>Simultaneous failure of three nodes</td>
  </tr>
</table>

### Planning for master nodes:

Master nodes provide API endpoints and keep Kubernetes workloads running. A Kubernetes cluster is able to operate as long as one of its master nodes is online. We suggest at least two master nodes for availability.

<table>
  <tr>
    <td>Node Count</td>
    <td>Safe for</td>
  </tr>
  <tr>
    <td>1</td>
    <td>Unsafe. Use only for small development clusters.</td>
  </tr>
  <tr>
    <td>2</td>
    <td>Failure of any one node.</td>
  </tr>
</table>

Both users of Kubernetes and Kubernetes itself occasionally attempt to communicate with master via a URL. With two or more masters, we suggest introducing a load balanced url (via a virtual IP or DNS CNAME). This is required to allow clients and components with Kubernetes to balance between them or to provide uninterrupted operation in the event that a master node goes offline.

### Planning for worker nodes:

Worker nodes are where your applications will run. your initial worker count should be large enough to hold all the workloads you intend to deploy to it plus enough slack to handle a partial failure. You can add more as necessary after the initial setup without interrupting operation of the cluster.

## Network

<table>
  <tr>
    <td>Networking Technique</td>
    <td>Routed<br />
Overlay</td>
  </tr>
  <tr>
    <td>How hostnames will be resolved for nodes</td>
    <td>Use DNS<br />
Let Kismatic Manage Hosts Files on nodes</td>
  </tr>
  <tr>
    <td>Network Policy Control</td>
    <td>No network policy<br />
Calico-managed network policy</td>
  </tr>
  <tr>
    <td>Pod Network CIDR Block</td>
    <td></td>
  </tr>
  <tr>
    <td>Services Network CIDR Block</td>
    <td></td>
  </tr>
  <tr>
    <td>Load-Balanced URL for Master Nodes</td>
    <td></td>
  </tr>
</table>


Kubernetes allocates a unique IP address for every Pod created on a cluster. Within a cluster, all Pods are visible to all other Pods and directly addressable by IP, simplifying point to point communications.

Similarly, Kubernetes uses a special network for Services, allowing them to talk to each other via an address that is stable even as the underlying cluster topology changes.

For this to work, Kubernetes makes use of technologies built in to Docker and Linux (including iptables, bridge networking and the Container Networking Interface). We tie these together with a network technology from Tigera networks called Calico.

### Pod and Service CIDR blocks

To provide these behaviors, Kubernetes needs to be able to issue IP addresses from two IP ranges: a **pod network** and a **services network**. This is in addition to the IP addresses nodes will be assigned on their **local network**.

The pod and service network ranges each need to be assigned a single contiguous CIDR block large enough to handle your workloads and any future scaling. With Calico, Worker and Master nodes are assigned IP addresses for allocation in blocks of 64 IPs; newly created pods will receive an address from this block until all IPs are consumed, at which point an additional block will be allocated to the node.

Thus, your pod network must be sized so that:

`Pod Network IP Block Size >= (Worker Node Count + Master Node Count) * 64`

Our default CIDR block for a pod is **172.16.0.0/16**, which would allow for a maximum of roughly 65k pods in total or roughly 1000 nodes with 64 pods per node or fewer.

Similarly, the service network needs to be large enough to handle all of the Services that might be created on the cluster. Our default is **172.17.0.0/16**, which would allow for 65k services and that ought to be enough for anybody.

Care should be taken that the IP addresses under management by Kubernetes do not collide with IP addresses on the local network, including omitting these ranges from control of  DHCP.

### Pod Networking

There are two techniques we support for pod networking on Kubernetes: **overlay** and **routed**.

In an **overlay** network, communications between pods happen on a virtual network that is only visible to machines that are running an agent. This agent communicates with other agents via the node's local network and establishes IP-over-IP tunnels through which Kubernetes Pod traffic is routed.

In this model, no work has to be done to allow pods to communicate with each other (other than ensuring that you are not blocking IP-over-IP traffic). Two or more Kubernetes clusters might even operate on the same pod and services IP ranges, without being able to see each others’ traffic.

However, work does need to be done to expose pods to the local network. This role is usually filled by a Kubernetes Ingress Controller.

Overlay networks work best for development clusters.

In a **routed** network, communications between pods happen on a network that is accessible to all machines on their local network. In this model, each node acts as a router for the ip ranges of the pods that it hosts. The cluster communicates with existing network routers via BGP to establish the responsibility of nodes for routing these addresses. Once routing is in place, a request to a pod or service IP is treated the same as any other request on the network. There is no tunnel or IP wrapping involved. This may also make it easier to inspect traffic via tools like wire shark and tcpdump.

In a routed model, cluster communications often work out of the box. Sometimes routers need to be configured to expect and acknowledge BGP messages from a cluster.

Routed networks work best when you want a majority of your workloads to be automatically visible to clients that aren't on Kubernetes, including other systems on the local network.

Sometimes, it is valuable to peer nodes in the cluster with a network router that is physically near to them. For this purpose, the cluster announces its BGP messages with an **AS Number** that may be specified when Kubernetes is installed. Our default AS Number is 64511.

### Pod Network Policy Enforcement

By default, Pods can talk to any port and any other Pod, Service or node on its network. Pod to pod network access is a requirement of Kubernetes, but this degree of openness is not.

When policy is enabled, access to all Pods is restricted and managed in part by Kubernetes and the Calico networking plugin. When adding new Pods, any ports that are identified within the definition will be made accessible to other pods. Access can be further opened or closed using the Calico command line tools installed on every Master node -- for example, you may grant access to a pod, or a namespace of pods, to a developer’s machine.

Network policy is an experimental feature that can make prototyping the cluster more difficult. It’s turned off by default.

### DNS & Load Balancing

All nodes in the cluster will need a short name with which they can communicate with each other. DNS is one way to provide this.

It's also valuable to have a load balanced alias for the master servers in your cluster, allowing for transparent failover if a master node goes offline. This can be performed either via DNS load balancing or via a Virtual IP if your network has a load balancer already. Pick a FQDN and short name for this alias to master that defines your cluster's intent -- for example, if this is the only Kubernetes cluster on your network, [kubernetes.yourdomain.com](http://kubernetes.yourdomain.com) would be ideal.

If you do not wish to run DNS, you may optionally allow the Kismatic installer to manage hosts files on all of your nodes. Be aware that this option will not scale beyond a few dozen nodes, as adding or removing nodes through the installer will force a hosts file update to all nodes on the cluster.

### Firewall Rules

Kubernetes must be allowed to manage network policy for any IP range it manages.

Network policies for the local network on which nodes reside will need to be set up prior to construction of the cluster, or installation will fail.

<table>
  <tr>
    <td>Purpose for rule</td>
    <td>Target node types</td>
    <td>Source IP range</td>
    <td>Allow Rules</td>
  </tr>
  <tr>
    <td>To allow acces to the API server</td>
    <td>worker</td>
    <td>worker nodes<br/>
        master nodes<br/>
        The IP ranges of any machines you want to be able to manage Kubernetes workloads </td>
    <td>tcp:6443</td>
  </tr>
  <tr>
    <td>To allow all internal traffic between Kubernetes nodes</td>
    <td>all</td>
    <td>All nodes in the Kubernetes cluster</td>
    <td>tcp:0-65535<br />
udp:0-65535</td>
  </tr>
  <tr>
    <td>To allow SSH</td>
    <td>all</td>
    <td>worker nodes<br/>
        master nodes<br/>
        The IP ranges of any machines you want to be able to manage Kubernetes nodes
    <td>tcp:22</td>
  </tr> 
  <tr>
  	<td>To allow communications between ETCD nodes</td>
    <td>etcd</td>
    <td>etcd nodes</td>
    <td>tcp:2380<br/>
 tcp:6660</td>
  </tr>
  <tr>
  	<td>To allow communications between Kubernetes nodes and ETCD</td>
    <td>etcd</td>
    <td>master nodes</td>
    <td>tcp:2379</td>
  </tr>
  <td>To allow communications between Calico networking and ETCD</td>
    <td>etcd</td>
    <td>etcd nodes</td>
    <td>tcp:6666</td>
  </tr>
</table>

## Certificates and Keys

<table>
  <tr>
    <td>Expiration period for certificates<br/>
<i>default 17520h</i></td>
    <td></td>
  </tr>
</table>

Kismatic will automate generation and installation of TLS certificates and keys used for intra-cluster security. It does this using the open source CloudFlare SSL library. These certificates and keys are exclusively used to encrypt and authorize traffic between Kubernetes components; they are not presented to end-users.

The default expiry period for certificates is **17520h** (2 years). Certificates must be updated prior to expiration or the cluster will cease to operate without warning. Replacing certificates will cause momentary downtime with Kubernetes as of version 1.4; future versions should allow for certificate "rolling" without downtime.
