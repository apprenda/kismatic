# Kismatic packages

Starting with KET v1.6, the tool installs the "official" packages from Kubernetes and Docker. For any cluster built with a previous version of KET, the packages will be upgraded to use the new repos during the regular upgrade process and should have no impact on the users. 


By default, Kismatic will install the required repos onto machines and use them to install the packages. This may not be acceptable, for example, if you want to adopt a "golden image" prior to rolling out a many-node cluster, if you need to install a cluster in a lab where most machines are disconnected from the internet, or if you simply want to save bandwidth. If this is your use case, please view the [instructions below](#synclocal).

## Installing via RPM (Redhat, CentOS)

#### Add the Docker repo to the machine
```
sudo bash -c 'cat <<EOF > /etc/yum.repos.d/docker.repo
[docker]
name=Docker
baseurl=https://yum.dockerproject.org/repo/main/centos/7/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://yum.dockerproject.org/gpg
EOF'
```

#### Add the Kubernetes repo to the machine
```
sudo bash -c 'cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
        https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF'
```

#### Install the RPMs for the type of node you want to create

| Component | Install Command |
| ---- | ---- |
| Etcd Node | `sudo yum -y install docker-engine-1.12.6-1.el7.centos` |
| Kubernetes Node | `sudo yum -y install docker-engine-1.12.6-1.el7.centos kubelet-1.7.4-0 kubectl-1.7.4-0` |

## Installing via DEB (Ubuntu Xenial)

#### Add the Docker repo to the machine

1. Install prerequisites
```
sudo apt-get update && sudo apt-get install -y apt-transport-https curl
```

2. Add the Docker public key to apt

```
curl -s https://apt.dockerproject.org/gpg | sudo apt-key add -
```

3.  Add the Kubernetes repo

```
sudo bash -c 'cat <<EOF >/etc/apt/sources.list.d/docker.list
deb https://apt.dockerproject.org/repo/ ubuntu-xenial main
EOF'
```

#### Add the Kubernetes repo to the machine
1. Add the Docker public key to apt

```
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
```

2. Add the Kubernetes repo

```
sudo bash -c 'cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://packages.cloud.google.com/apt/ kubernetes-xenial main
EOF'
```

#### Refresh the machine's repo cache

```
sudo apt-get update
```

#### Install the RPMs for the type of node you want to create

| Component | Install Command |
| ---- | ---- |
| Etcd Node | `sudo apt-get install -y docker-engine=1.12.6-0~ubuntu-xenial` |
| Kubernetes Node | `sudo apt-get install -y docker-engine=1.12.6-0~ubuntu-xenial kubelet=1.7.4-00 kubectl=1.7.4-00` |


# <a name="synclocal"></a>Synchronizing a local repo

If you maintain a package repository, you should not perform step 1 in the instructions above. Instead, you should point machines to your own package repository and keep it in sync with Docker and Kubernetes.

Each package will also have many transitive dependencies. To be able to install nodes fully disconnected from the internet, you will need to synchronize your repo with these packages' repos as well.

Dependencies and their versions will change over time and as such they are not listed here. Instead, they can be derived from any machine that is integrated with our repos using the commands linked below.

One way to ensure you've correctly synchronized your repo is to install a test cluster.

1. Provision 1 node of each role
2. Install the packages
3. Run `kismatic plan` to generate a new Plan file
4. Update the Plan file to identify your nodes and using the configuration `disable_package_installation=true`.
5. Run `kismatic validate`
6. During validation, the Kismatic inspector will check your packages to be sure they installed correctly and will fail if any of them are missing.

Changes to dependencies should be called out in the notes that accompany a release.

## yum

Listing dependencies of a package: `yum deplist $PACKAGE`

[Syncing with a repo](http://bencane.com/2013/04/15/creating-a-local-yum-repository/)

## apt


Listing dependencies of a package: `apt-cache depends $PACKAGE`

[Syncing with a repo](http://www.tecmint.com/setup-local-repositories-in-ubuntu/)
