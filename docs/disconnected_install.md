# Disconnected Installation

Certain organizations need to run Kubernetes clusters in air-gapped environments, and thus need to perform an installation that is completely disconnected from the internet. The process of performing an installation on nodes with no internet access is called a disconnected installation.

Being disconnected means that you will not use public repositories or registries to get binaries to your nodes. Instead, before performing the installation, you will sync a local package repository and container image registry with the packages and images required to operate a Kubernetes cluster.

- [Prerequisites](#prerequisites)
- [Planning the installation](#planning-the-installation)
- [Installing the cluster](#installing-the-cluster)
- [Upgrading your cluster](#upgrading-your-cluster)
- [Creating a local package repository](#creating-a-local-package-repository)
  - [CentOS](#centos-7)
  - [RHEL 7](#rhel-7)
  - [Ubuntu 16.04](#ubuntu-1604)
- [Seeding a local container registry](#seeding-a-local-container-registry)

## Prerequisites

* Local package repository that is accessible from all nodes. This repository must include the Kubernetes software packages and their transitive dependencies.

* The local package repository must be configured on all nodes. 

* Package repositories that are not accessible should be disabled or removed.
Otherwise, the package manager will attempt to download metadata from these
inaccessible repositories, and the installation wil fail.

* Local docker registry that is accessible from all nodes. 
This registry must be seeded with the images required for the installation. See [Seeding a local container registry](#seeding-a-local-container-registry).

## Planning the installation
Before executing the validation or installation stages, you must let KET know that
it should perform a disconnected installation. The following plan file options
must be considered:

**disconnected_installation**: This field must be set to `true` when performing a
disconnected installation. When `true`, KET will:
1. Not configure the upstream package repositories. Instead, KET will assume that the 
internla repositories have been configured on all nodes.
2. Use the local image registry for cluster components, instead of pulling them from
Docker Hub, GCR, or other public registries.

**disable_package_installation**: In most cases, KET is responsible for installing the required packages onto the cluster nodes. If, however, you want to control the installation of the packages, you can set this flag to `true` to prevent KET from installing the packages. More importantly, disabling package installation will enable a set of preflight checks that will ensure the packages have been installed on all nodes.

## Installing the cluster

Once the relevant options in the plan file have been set, and the local repository and local registry have been stood up, you are ready to perform the disconnected installation. 

At this point, you can run `kismatic install apply` to initiate the installation.

## Upgrading your cluster
Before performing a cluster upgrade, you must:
- Update your local package repository to include the new packages.
- Seed your local registry using the new version of KET.

# Creating a local package repository

## CentOS 7

### Install required utilities
We will use `reposync` to download the packages from upstream repositories, and `httpd` to expose our local repository over HTTP.

```
yum install yum-utils httpd createrepo
```

### Setup the upstream repositories

The kubernetes, docker and gluster RPM repositories must be configured on the node to pull the packages.

```
# Add docker repo
sudo bash -c 'cat <<EOF > /etc/yum.repos.d/docker.repo
[docker]
name=Docker
baseurl=https://download.docker.com/linux/centos/7/x86_64/stable/
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://download.docker.com/linux/centos/gpg
EOF'

# Add Kubernetes repo
sudo bash -c 'cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
        https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF'

# Add Gluster repo
sudo bash -c 'cat <<EOF > /etc/yum.repos.d/gluster.repo
[gluster]
name=Gluster
baseurl=http://buildlogs.centos.org/centos/7/storage/x86_64/gluster-3.8/
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://download.gluster.org/pub/gluster/glusterfs/3.8/3.8.7/rsa.pub
EOF'

# Clean yum cache
yum clean all
```

### Download the RPMs using reposync
Sync the desired packages to the local machine, and place them in `/var/www/html`.

```
reposync -l -p /var/www/html/ -r base -r updates -r docker -r gluster

# The kubernetes repo is special as it places the packages in an unexpected location.
reposync -l -p /var/www/html -r kubernetes
mv /var/www/pool/* /var/www/html/kubernetes/
rmdir /var/www/pool
```

### Create a repository
Now that we have the RPMs locally, we must generate the metadata files required for the repository.

```
for repo in `ls /var/www/html`
do 
    createrepo /var/www/html/$repo
done
```

### Start Apache server
We will use the Apache HTTP server for exposing the repository over HTTP.
```
systemctl enable httpd
systemctl start httpd
```

### Configure nodes
With this approach, we created five mirrors on the same machine 
that must be configured on the nodes:
* `/base`
* `/updates`
* `/docker`
* `/kubernetes`
* `/gluster`

For example, to configure the base repository that has been created on a machine with hostname
`rpm-mirror.example.com`, you can create the file `/etc/yum.repos.d/base.repo`
with the following:
```
[base]
name=Base
baseurl=http://rpm-mirror.example.com/base
enabled=1
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
```

The above configuration file must be created for all the repository mirrors listed.

## RHEL 7
Creating a mirror for nodes that are running RHEL is fairly similar to the process
described for CentOS. However, depending on your RHEL distribution, the "base" and
"updates" mirror will differ.

The RHEL 7 AMI on AWS, for example, uses `rhui-REGION-rhel-server-releases` as the
repo ID for the RedHat repository.

## Ubuntu 16.04

### Pre-requisites
* GPG Private Key is required to sign the repositories. The generation and management
of the key is outside of the scope of this document. This is a [handy cheatsheet](http://irtfweb.ifa.hawaii.edu/~lockhart/gpg/).

### Install required utilities
We will use [aptly](https://www.aptly.info) to mirror and serve the repository.

```
echo "deb http://repo.aptly.info/ squeeze main" >> /etc/apt/sources.list
wget -qO - https://www.aptly.info/pubkey.txt | sudo apt-key add -

apt-get -y update
apt-get -y install aptly
```

### Create snapshost of the docker repository
```
wget -O - https://download.docker.com/linux/ubuntu/gpg | gpg --no-default-keyring --keyring trustedkeys.gpg --import
aptly -architectures="amd64" mirror create docker https://download.docker.com/linux/ubuntu xenial stable
aptly mirror update docker
aptly snapshot create docker from mirror docker
```

### Create snapshost of the kubernetes repository
```
wget -O - https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --no-default-keyring --keyring trustedkeys.gpg --import
aptly mirror create kubernetes https://packages.cloud.google.com/apt/ kubernetes-xenial main
aptly mirror update kubernetes
aptly snapshot create kubernetes from mirror kubernetes
aptly publish snapshot kubernetes
```

### Create snapshost of the Ubuntu repository

Note: The filter parameter might have to change, depending on what is available on the
Ubuntu image you are using.

```
gpg --no-default-keyring --keyring trustedkeys.gpg --keyserver keys.gnupg.net --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32
aptly mirror create \
  -architectures=amd64 \
  -filter="bridge-utils|nfs-common|socat|libltdl7|python2.7|python-apt|ebtables|libaio1|libibverbs1|libpython2.7|librdmacm1|liburcu4|attr" \
  -filter-with-deps \
  ubuntu-main http://archive.ubuntu.com/ubuntu xenial main universe
aptly mirror update ubuntu-main
aptly snapshot create ubuntu-main from mirror ubuntu-main
```

### Create a snapshot of the Gluster repository
```
gpg --no-default-keyring --keyring trustedkeys.gpg --keyserver keyserver.ubuntu.com --recv-keys 3FE869A9
aptly mirror create gluster ppa:gluster/glusterfs-3.8
aptly mirror update gluster
aptly snapshot create gluster from mirror gluster
```

### Merge ubuntu, gluster and docker snapshots
```
aptly snapshot merge xenial-repo ubuntu-main gluster docker
aptly publish snapshot xenial-repo
```

### Serve the mirrors
```
# Serve the repositories
cat <<EOF > /etc/systemd/system/aptly.service
[Service]
Type=simple
ExecStart=/usr/bin/aptly serve -listen=:80
User=root
EOF

systemctl daemon-reload
systemctl enable aptly
systemctl start aptly
```

### Configure nodes
The mirror must be configured on all nodes of the cluster, and any repository
that is not available from the node must be disabled.

Sample `/etc/apt/sources.list`:
```
deb http://mirror.example.com xenial main
deb http://mirror.example.com kubernetes-xenial main
deb [arch=amd64] http://mirror.example.com xenial stable
```

## Seeding a local container registry

The local registry must contain all the required images before installing the cluster.
The `seed-registry` command can be used to seed the registry with the images, or to
obtain a list of all the required images.

For more information about using a local registry, see the [Container Image Registry](./container-registry.md)
documentation.
