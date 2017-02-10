# Kismatic packages

We provide our own packages to ensure that all services have versions of their dependencies that will work together. Where possible, our packages simply declare dependencies on a vendor's own packages (Docker, for example).

With Kubernetes, packages are still in incubation; we will continue to produce our own for now and repackage as new Kubernetes releases come out.

By default, Kismatic will install its own repos onto machines and use them to download Kismatic packages. This may not be acceptable, for example, if you want to adopt a "golden image" prior to rolling out a many-node cluster, if you need to install a cluster in a lab where most machines are disconnected from the internet, or if you simply want to save bandwidth. If this is your use case, please view the [instructions below](#synclocal).

## Installing via RPM (Redhat, CentOS, Fedora)

1. Add the Kismatic repo to the machine

`sudo curl https://kismatic-packages-rpm.s3-accelerate.amazonaws.com/kismatic.repo -o /etc/yum.repos.d/kismatic.repo`

2. Install the RPMs for the type of node you want to create

| Product | Install Command |
| --- | --- | --- |
| Etcd | `sudo yum -y install etcd-3.1.0-1` |
| Kubernetes Master | `sudo yum -y install docker-engine-1.11.2-1.el7.centos kubelet-1.5.2_4-1 kubectl-1.5.2_4-1` |
| Kubernetes Worker | `sudo yum -y install docker-engine-1.11.2-1.el7.centos kubelet-1.5.2_4-1` |

## Installing via DEB (Ubuntu Xenial)

1. Add the Kismatic repo to the machine
   1. Add the Kismatic public key to apt

`wget -qO - https://kismatic-packages-deb.s3-accelerate.amazonaws.com/public.key | sudo apt-key add -`

   2. Add the Kismatic repo

`cat <<EOF > /etc/apt/sources.list.d/kubernetes.list
deb https://kismatic-packages-deb.s3-accelerate.amazonaws.com kismatic-xenial main
EOF
`

2. Refresh the machine's repo cache

`sudo apt-get update`

3. Install the RPMs for the type of node you want to create

| Product | Install Command |
| --- | --- | --- |
| Etcd | `sudo apt-get -y -t=kismatic-xenial  install etcd=3.1.0` |
| Kubernetes Master | `sudo apt-get -y -t=kismatic-xenial install docker=1.11.2-0~xenial kubelet=1.5.2-4 kubectl=1.5.2-4` |
| Kubernetes Worker | `sudo apt-get -y -t=kismatic-xenial  install docker=1.11.2-0~xenial kubelet=1.5.2-4` |

# <a name="synclocal"></a>Synchronizing a local repo

If you maintain a package repository, you should not perform step 1 in the instructions above. Instead, you should point machines to your own package repository and keep it in sync with Kismatic.

Each Kismatic package will also have many transitive dependencies. To be able to install nodes fully disconnected from the internet, you will need to synchronize your repo with these packages' repos as well.

Dependencies and their versions will change over time and as such they are not listed here. Instead, they can be derived from any machine that is integrated with our repos using the commands linked below.

One way to ensure you've correctly synchronized your repo is to install a test cluster.

1. Provision 1 node of each role
2. Install the Kismatic packages
3. Run `kismatic plan` to generate a new Plan file
4. Update the Plan file to identify your nodes and using the configuration `allow_package_installation=false`.
5. Run `kismatic validate`
6. During validation, the Kismatic inspector will check your packages to be sure they installed correctly and will fail if any of them are missing.

Changes to dependencies should be called out in the notes that accompany a release.

## yum

Listing dependencies of a package: `yum deplist $PACKAGE`

[Syncing with a repo](http://bencane.com/2013/04/15/creating-a-local-yum-repository/)

## apt


Listing dependencies of a package: `apt-cache depends $PACKAGE`

[Syncing with a repo](http://www.tecmint.com/setup-local-repositories-in-ubuntu/)
