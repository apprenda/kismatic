#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

echo "deb http://repo.aptly.info/ squeeze main" >> /etc/apt/sources.list
wget -qO - https://www.aptly.info/pubkey.txt | sudo apt-key add -

apt-get -y update
apt-get -y install aptly

### THIS IS A DUMMY PRIVATE KEY USED FOR TESTING PURPOSES. 
### DO NOT USE THIS FOR ANYTHING ELSE.
cat <<EOF | gpg --import
-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

lQIGBFnI89IBBACuQOYVN+w/6s4BUQ/E3wZLX7GKhTcOOVz8GpraqVkQDYvEYWQN
c90qCNXPwuOS2H0Dcbc+h/GZc007zj9g8o3n7pw7Wt43d+W22m8FfhGPfGU5n4x0
c81F4a0fGAE0P17cwpNdzOXPs5awl9mTr9y1oU4PIezr6AqMWyACpKL8dQARAQAB
/gcDAuzPHLdImiEOYBQkG6TAguPQAA5bGD6lMtocBa0A4sbZrq5/2noJ1bzDFZ/t
busTm4D7dlmjFft5uT/Lu0L+Qv1v1pB61B7H/jYDiisPcQyl/N0nJSmeI/c+7EnH
s3BFdpYKEsQxM2U3HAJO+81RkjO4aAUISN5SeVzyRY3nHO3rf8SLna1aJ38EODIl
OD190ZLvo3RaGuX6NuqwtBbjAweDokKeH27L/QhSjZorVg3Kvccd0inC6mJgDdg9
gOWE9vyKkUYfvP+7YoU1CldZ2/bnPixG9Dj9PvmKv1edE1EdwugQbCUFjIi9/hIo
MURPz4CQWr5CRtYkILpy/sU0t1ym5Rglw25TWx3++ilofdSAPbsyF/gG+W6jQaF8
aZZX6zzN0PvB4p/d3Semnt6uVJlSzEnZiukYvTL9q0uMs74gAWOxJqn6+bpFdPEu
OevVZzGovmep7DmkjJnao8aQC4/9nZpV5Ir0kV2DrWfD35eW/kYba2O0GWR1bW15
IDxkdW1teUBleGFtcGxlLmNvbT6IuAQTAQIAIgUCWcjz0gIbAwYLCQgHAwIGFQgC
CQoLBBYCAwECHgECF4AACgkQsJ2tkWLvIyjjjQP9FnmlogJxyU3+wP/uRRjfSKUF
3aLiEFor22Uq/M8bG819VDHlN/UcAtzADliUjesvTN/DiE8haDZ6fq1LsK7VrUvR
W0sRidk0ejseA0PEjv7V+yEByZBIUS8+1UkF3WT0cPzNjXGXYkpO56vy/xQl+K1h
vnALKVuxDKaGerVZVoqdAgYEWcjz0gEEAN1V7xB+BEVRvROGxDcqlIk3wa0aOl9K
VWRJaF8U0ueWWZupBpRxKsKqsRL2+ooU/oFLm69EddFVhqvTWOlBUfW6NgNSCpgH
euiT4AHEltVGM+NtOQ+OszKnKIzGhUuOqmOyI4RdJTDkOt6hgKJt4gaJTB4qxWXH
2s8OtzKnoy29ABEBAAH+BwMC7M8ct0iaIQ5gTB3hv/vSeQB20EgaBGN4rku9Ddlk
5Rn5WeS31uXYNPtjat1K6o96JlNjDYOlAn+La6vbuKfXce695ZIpE/JyY6VSnkzk
kg9sr0215jHs0eZSAN0ECkXAOaVwUS0RM/x6HPKC7dRFmnkZa+TxtlifR8CXfhcj
v0z2Z7G/eoOPejPBZdNeE+zVj+dNSYCyBx4QqS/1LOvq0iwqTLB+pMcnR7z6EjP9
8UgxbDfXcoYy0dE8IUnAqSzpOkWv/lT9tQ3CqjsTl3KyBw1tt9f/6EuU9X/LO0tn
uPJ9DOnpGhbSzl491uvJH7nfcmMVYJtLaUY5n6seW77Gr7oaVyEwfQRyHKgFL4aj
lCCRT/OFMAEu/xY3EbpTMTtOvaCRVey0Kr6KDZK1ZBSbDDegnJFfDFebbWe2y0BQ
pM0MoO2UFaIXGXn7GAslcWSRTrDh5kjg08hC1Ilgp06OgyZqM2aR2GOE+H5vCztn
7Ap6tYXu94ifBBgBAgAJBQJZyPPSAhsMAAoJELCdrZFi7yMoH00EAI4uuXeGucuv
NvRi7zv66tVuJa5IiVNQ4V6FKSDMMtmAmPCrU7wJbb7fBhE1hkPzJyZULrZo5kru
Yot8BVA5LYAGed6b35q7JeBCe7gjKJpsIHEFYto4nK/JDYiAotBzq4EUjuNgaiy3
Xt3kFlXWFvZIMEmJWA/fHakxzU+TjUbc
=fkrb
-----END PGP PRIVATE KEY BLOCK-----
EOF

### Create snapshost of the docker repository
wget -O - https://apt.dockerproject.org/gpg | gpg --no-default-keyring --keyring trustedkeys.gpg --import
aptly mirror create docker https://apt.dockerproject.org/repo/ ubuntu-xenial main
aptly mirror update docker
aptly snapshot create docker from mirror docker
aptly publish snapshot --passphrase dummy docker

### Create snapshost of the kubernetes repository
wget -O - https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --no-default-keyring --keyring trustedkeys.gpg --import
aptly mirror create kubernetes https://packages.cloud.google.com/apt/ kubernetes-xenial main
aptly mirror update kubernetes
aptly snapshot create kubernetes from mirror kubernetes
aptly publish snapshot --passphrase dummy kubernetes

### Create snapshost of the Ubuntu repository
### Retry loop: Adding this gpg fails with networking blips on occasion. 
n=0
while true
do
  gpg --no-default-keyring --keyring trustedkeys.gpg --keyserver keys.gnupg.net --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32 && break || true
  n=$((n+1))
  if [ $n -ge 3 ]; then exit 1; fi
  echo "Retrying..."
  sleep 5
done

aptly mirror create \
  -architectures=amd64 \
  -filter="bridge-utils|nfs-common|socat|libltdl7|python2.7|python-apt|ebtables|libaio1|libibverbs1|libpython2.7|librdmacm1|liburcu4|attr" \
  -filter-with-deps \
  ubuntu-main http://archive.ubuntu.com/ubuntu xenial main universe
aptly mirror update ubuntu-main
aptly snapshot create ubuntu-main from mirror ubuntu-main

### Create a snapshot of the Gluster repository
gpg --no-default-keyring --keyring trustedkeys.gpg --keyserver keyserver.ubuntu.com --recv-keys 3FE869A9
aptly mirror create gluster ppa:gluster/glusterfs-3.8
aptly mirror update gluster
aptly snapshot create gluster from mirror gluster

### Merge ubuntu and gluster snapshots into
aptly snapshot merge xenial-repo ubuntu-main gluster
aptly publish snapshot --passphrase dummy xenial-repo

### Serve the mirrors
nohup aptly serve -listen=:80 &

sleep 3
cat nohup.out
