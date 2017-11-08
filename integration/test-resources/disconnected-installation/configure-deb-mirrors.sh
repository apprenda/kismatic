#!/bin/bash
#
# Configures DEB mirrors on a cluster node for a disconnected installation.
# It also removes any pre-configured mirrors to avoid trying to "refresh" them.
# Usage: ./configure-mirror-debs.sh MIRROR_BASE_URL
#     where MIRROR_BASE_URL is the URL to where the mirror is running.
#
set -o errexit
set -o pipefail
set -o nounset

# Remove all pre-existing repos
mv /etc/apt/sources.list.d/ /etc/apt/sources.list.d.backup
mv /etc/apt/sources.list /etc/apt/sources.list.backup

# Add the gpg public key to the apt keyring
cat <<EOF | apt-key add -
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mI0EWcjz0gEEAK5A5hU37D/qzgFRD8TfBktfsYqFNw45XPwamtqpWRANi8RhZA1z
3SoI1c/C45LYfQNxtz6H8ZlzTTvOP2DyjefunDta3jd35bbabwV+EY98ZTmfjHRz
zUXhrR8YATQ/XtzCk13M5c+zlrCX2ZOv3LWhTg8h7OvoCoxbIAKkovx1ABEBAAG0
GWR1bW15IDxkdW1teUBleGFtcGxlLmNvbT6IuAQTAQIAIgUCWcjz0gIbAwYLCQgH
AwIGFQgCCQoLBBYCAwECHgECF4AACgkQsJ2tkWLvIyjjjQP9FnmlogJxyU3+wP/u
RRjfSKUF3aLiEFor22Uq/M8bG819VDHlN/UcAtzADliUjesvTN/DiE8haDZ6fq1L
sK7VrUvRW0sRidk0ejseA0PEjv7V+yEByZBIUS8+1UkF3WT0cPzNjXGXYkpO56vy
/xQl+K1hvnALKVuxDKaGerVZVoq4jQRZyPPSAQQA3VXvEH4ERVG9E4bENyqUiTfB
rRo6X0pVZEloXxTS55ZZm6kGlHEqwqqxEvb6ihT+gUubr0R10VWGq9NY6UFR9bo2
A1IKmAd66JPgAcSW1UYz4205D46zMqcojMaFS46qY7IjhF0lMOQ63qGAom3iBolM
HirFZcfazw63MqejLb0AEQEAAYifBBgBAgAJBQJZyPPSAhsMAAoJELCdrZFi7yMo
H00EAI4uuXeGucuvNvRi7zv66tVuJa5IiVNQ4V6FKSDMMtmAmPCrU7wJbb7fBhE1
hkPzJyZULrZo5kruYot8BVA5LYAGed6b35q7JeBCe7gjKJpsIHEFYto4nK/JDYiA
otBzq4EUjuNgaiy3Xt3kFlXWFvZIMEmJWA/fHakxzU+TjUbc
=1KmT
-----END PGP PUBLIC KEY BLOCK-----
EOF

cat <<EOF > /etc/apt/sources.list
deb $1 xenial main
deb $1 kubernetes-xenial main
deb [arch=amd64] $1 xenial stable
EOF

apt-get update -y
