#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

yum install -y yum-utils httpd createrepo

# Add docker repo
cat <<EOF > /etc/yum.repos.d/docker.repo
[docker]
name=Docker
baseurl=https://download.docker.com/linux/centos/7/x86_64/stable/
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://download.docker.com/linux/centos/gpg
EOF

# Add Kubernetes repo
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
        https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

# Add Gluster repo
cat <<EOF > /etc/yum.repos.d/gluster.repo
[gluster]
name=Gluster
baseurl=http://buildlogs.centos.org/centos/7/storage/x86_64/gluster-3.8/
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://download.gluster.org/pub/gluster/glusterfs/3.8/3.8.7/rsa.pub
EOF

reposync -p /var/www/html/ -r base -r updates -r docker -r gluster

# The kubernetes repo is special as it places the packages in an unexpected location.
reposync -p /var/www/html -r kubernetes
mv /var/www/pool/* /var/www/html/kubernetes/
rmdir /var/www/pool

for repo in `ls /var/www/html`
do 
    createrepo /var/www/html/$repo
done

systemctl enable httpd
systemctl start httpd
