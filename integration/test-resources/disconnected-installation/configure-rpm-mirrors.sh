#!/bin/bash
#
# Configures RPM mirrors on a cluster node for a disconnected installation.
# It also removes any pre-configured mirrors to avoid trying to "refresh" them.
# Usage: ./configure-mirror-rpms.sh MIRROR_BASE_URL
#     where MIRROR_BASE_URL is the URL to where the mirror is running.
#
set -o errexit
set -o pipefail
set -o nounset

# Remove all pre-existing repos
mv /etc/yum.repos.d /etc/yum.repos.d.backup
mkdir /etc/yum.repos.d

# Add base repo
cat <<EOF > /etc/yum.repos.d/base.repo
[base]
name=Base
baseurl=$1/base
enabled=1
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
EOF

# Add updates repo
cat <<EOF > /etc/yum.repos.d/updates.repo
[updates]
name=updates
baseurl=$1/updates
enabled=1
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
EOF

cat <<EOF > /tmp/docker.gpg
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBFit5IEBEADDt86QpYKz5flnCsOyZ/fk3WwBKxfDjwHf/GIflo+4GWAXS7wJ
1PSzPsvSDATV10J44i5WQzh99q+lZvFCVRFiNhRmlmcXG+rk1QmDh3fsCCj9Q/yP
w8jn3Hx0zDtz8PIB/18ReftYJzUo34COLiHn8WiY20uGCF2pjdPgfxE+K454c4G7
gKFqVUFYgPug2CS0quaBB5b0rpFUdzTeI5RCStd27nHCpuSDCvRYAfdv+4Y1yiVh
KKdoe3Smj+RnXeVMgDxtH9FJibZ3DK7WnMN2yeob6VqXox+FvKYJCCLkbQgQmE50
uVK0uN71A1mQDcTRKQ2q3fFGlMTqJbbzr3LwnCBE6hV0a36t+DABtZTmz5O69xdJ
WGdBeePCnWVqtDb/BdEYz7hPKskcZBarygCCe2Xi7sZieoFZuq6ltPoCsdfEdfbO
+VBVKJnExqNZCcFUTEnbH4CldWROOzMS8BGUlkGpa59Sl1t0QcmWlw1EbkeMQNrN
spdR8lobcdNS9bpAJQqSHRZh3cAM9mA3Yq/bssUS/P2quRXLjJ9mIv3dky9C3udM
+q2unvnbNpPtIUly76FJ3s8g8sHeOnmYcKqNGqHq2Q3kMdA2eIbI0MqfOIo2+Xk0
rNt3ctq3g+cQiorcN3rdHPsTRSAcp+NCz1QF9TwXYtH1XV24A6QMO0+CZwARAQAB
tCtEb2NrZXIgUmVsZWFzZSAoQ0UgcnBtKSA8ZG9ja2VyQGRvY2tlci5jb20+iQI3
BBMBCgAhBQJYrep4AhsvBQsJCAcDBRUKCQgLBRYCAwEAAh4BAheAAAoJEMUv62ti
Hp816C0P/iP+1uhSa6Qq3TIc5sIFE5JHxOO6y0R97cUdAmCbEqBiJHUPNQDQaaRG
VYBm0K013Q1gcJeUJvS32gthmIvhkstw7KTodwOM8Kl11CCqZ07NPFef1b2SaJ7l
TYpyUsT9+e343ph+O4C1oUQw6flaAJe+8ATCmI/4KxfhIjD2a/Q1voR5tUIxfexC
/LZTx05gyf2mAgEWlRm/cGTStNfqDN1uoKMlV+WFuB1j2oTUuO1/dr8mL+FgZAM3
ntWFo9gQCllNV9ahYOON2gkoZoNuPUnHsf4Bj6BQJnIXbAhMk9H2sZzwUi9bgObZ
XO8+OrP4D4B9kCAKqqaQqA+O46LzO2vhN74lm/Fy6PumHuviqDBdN+HgtRPMUuao
xnuVJSvBu9sPdgT/pR1N9u/KnfAnnLtR6g+fx4mWz+ts/riB/KRHzXd+44jGKZra
IhTMfniguMJNsyEOO0AN8Tqcl0eRBxcOArcri7xu8HFvvl+e+ILymu4buusbYEVL
GBkYP5YMmScfKn+jnDVN4mWoN1Bq2yMhMGx6PA3hOvzPNsUoYy2BwDxNZyflzuAi
g59mgJm2NXtzNbSRJbMamKpQ69mzLWGdFNsRd4aH7PT7uPAURaf7B5BVp3UyjERW
5alSGnBqsZmvlRnVH5BDUhYsWZMPRQS9rRr4iGW0l+TH+O2VJ8aQ
=0Zqq
-----END PGP PUBLIC KEY BLOCK-----
EOF

# Add docker repo
cat <<EOF > /etc/yum.repos.d/docker.repo
[docker]
name=Docker
baseurl=$1/docker
enabled=1
gpgcheck=1
gpgkey=file:///tmp/docker.gpg
EOF

cat <<EOF > /tmp/kubernetes-yum-key.gpg
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFUd6rIBCAD6mhKRHDn3UrCeLDp7U5IE7AhhrOCPpqGF7mfTemZYHf/5Jdjx
cOxoSFlK7zwmFr3lVqJ+tJ9L1wd1K6P7RrtaNwCiZyeNPf/Y86AJ5NJwBe0VD0xH
TXzPNTqRSByVYtdN94NoltXUYFAAPZYQls0x0nUD1hLMlOlC2HdTPrD1PMCnYq/N
uL/Vk8sWrcUt4DIS+0RDQ8tKKe5PSV0+PnmaJvdF5CKawhh0qGTklS2MXTyKFoqj
XgYDfY2EodI9ogT/LGr9Lm/+u4OFPvmN9VN6UG+s0DgJjWvpbmuHL/ZIRwMEn/tp
uneaLTO7h1dCrXC849PiJ8wSkGzBnuJQUbXnABEBAAG0QEdvb2dsZSBDbG91ZCBQ
YWNrYWdlcyBBdXRvbWF0aWMgU2lnbmluZyBLZXkgPGdjLXRlYW1AZ29vZ2xlLmNv
bT6JAT4EEwECACgFAlUd6rICGy8FCQWjmoAGCwkIBwMCBhUIAgkKCwQWAgMBAh4B
AheAAAoJEDdGwginMXsPcLcIAKi2yNhJMbu4zWQ2tM/rJFovazcY28MF2rDWGOnc
9giHXOH0/BoMBcd8rw0lgjmOosBdM2JT0HWZIxC/Gdt7NSRA0WOlJe04u82/o3OH
WDgTdm9MS42noSP0mvNzNALBbQnlZHU0kvt3sV1YsnrxljoIuvxKWLLwren/GVsh
FLPwONjw3f9Fan6GWxJyn/dkX3OSUGaduzcygw51vksBQiUZLCD2Tlxyr9NvkZYT
qiaWW78L6regvATsLc9L/dQUiSMQZIK6NglmHE+cuSaoK0H4ruNKeTiQUw/EGFaL
ecay6Qy/s3Hk7K0QLd+gl0hZ1w1VzIeXLo2BRlqnjOYFX4A=
=HVTm
-----END PGP PUBLIC KEY BLOCK-----
EOF

cat <<EOF > /tmp/kubernetes-rpm-package.gpg
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFWKtqgBCADmKQWYQF9YoPxLEQZ5XA6DFVg9ZHG4HIuehsSJETMPQ+W9K5c5
Us5assCZBjG/k5i62SmWb09eHtWsbbEgexURBWJ7IxA8kM3kpTo7bx+LqySDsSC3
/8JRkiyibVV0dDNv/EzRQsGDxmk5Xl8SbQJ/C2ECSUT2ok225f079m2VJsUGHG+5
RpyHHgoMaRNedYP8ksYBPSD6sA3Xqpsh/0cF4sm8QtmsxkBmCCIjBa0B0LybDtdX
XIq5kPJsIrC2zvERIPm1ez/9FyGmZKEFnBGeFC45z5U//pHdB1z03dYKGrKdDpID
17kNbC5wl24k/IeYyTY9IutMXvuNbVSXaVtRABEBAAG0Okdvb2dsZSBDbG91ZCBQ
YWNrYWdlcyBSUE0gU2lnbmluZyBLZXkgPGdjLXRlYW1AZ29vZ2xlLmNvbT6JATgE
EwECACIFAlWKtqgCGy8GCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEPCcOUw+
G6jV+QwH/0wRH+XovIwLGfkg6kYLEvNPvOIYNQWnrT6zZ+XcV47WkJ+i5SR+QpUI
udMSWVf4nkv+XVHruxydafRIeocaXY0E8EuIHGBSB2KR3HxG6JbgUiWlCVRNt4Qd
6udC6Ep7maKEIpO40M8UHRuKrp4iLGIhPm3ELGO6uc8rks8qOBMH4ozU+3PB9a0b
GnPBEsZdOBI1phyftLyyuEvG8PeUYD+uzSx8jp9xbMg66gQRMP9XGzcCkD+b8w1o
7v3J3juKKpgvx5Lqwvwv2ywqn/Wr5d5OBCHEw8KtU/tfxycz/oo6XUIshgEbS/+P
6yKDuYhRp6qxrYXjmAszIT25cftb4d4=
=/PbX
-----END PGP PUBLIC KEY BLOCK-----
EOF

# Add Kubernetes repo
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=$1/kubernetes
enabled=1
gpgcheck=1
gpgkey=file:///tmp/kubernetes-rpm-package.gpg
        file:///tmp/kubernetes-yum-key.gpg
EOF

cat <<EOF > /tmp/gluster.gpg
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFYEjzYBCADJJQNhYyCPmOwpYuJVk4ywq3bsWU9oDy/6OguoNheSobobVsP6
cEi5CVYIB6SYJbdj27w7uSOgCWQM58wGBdQEH09P2cbYlqEhdqRzmw9B0wlcZbB9
Kg2eiBVLH0wnWi3pHgtaltsSHI01qyyfS1cEXVZewmkrqcmXgjaChy8SYUPey43K
MJOe0TRL02PaPvXvX3jG1+J4XGTt/fb8slZrIdcUcO3W+mnINg1fut/mbD2RDSJH
yoexyQD1AP96oqxksS/EaCsUsjLgQ5BSiV9XerieDv+vMIBb/sKuhjoMcxtAJFwq
J/rqHXDxUrwo+Zoo4e2FQOw1E2DOwkO3t56tABEBAAG0N0dsdXN0ZXIgUGFja2Fn
ZXIgPGdsdXN0ZXJwYWNrYWdlckBkb3dubG9hZC5nbHVzdGVyLm9yZz6JATgEEwEC
ACIFAlYEjzYCGwMGCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEP55u1LV3FLc
amwH/0v2rEpY57tPu+bd4VvvnjO3rYI0Uy0G4cMN7ogMaOYnlOkKsNBNmfi5Nkg5
I26KC4zHstg/n4ViuZjVklqidbCqaPjeAIZdO6M84EaPPOoVBSUB6QTvN5NnlUsG
WywAVHaeOFoU5m6zqlf4mkg58rnFKOXgLkXpWg0h0bbczhAkP4HtnQqnt4V4GRPl
JiZfy0eCe+gUaonrVplfUQ9hcrxi8ZE0vvhaq10RK8V3Q/HZiZ+2izWzMJYViNKw
pVhPpTbgk+b1Csc+s1cWs1Sv8vId6EoTxldPeAcJNcqcNV6dbe5Ewy3HOUc1ydI+
cCHVq/LghVOyWLS05HRtfN76fCq5AQ0EVgSPNgEIAMW84kVtosYYpo1U9oiGD6Ji
5jTHYSEn4VdmHByXVuRvHmwhgAIAHX/0vBhhOogulKbR1g0t+E6lRQqbFSSKE77j
8udATv2VS9J/ApwDiChLGz6TUq9qjIvXKbVddtO57WE9MD2DpMwVAyPPeZeLE6Qz
vT8bVRx/UteOkMsfPKn3xPtYLgcz1WF1hpk0Efwwi9wjpNChH/qvPdLmvr6PZTci
ux791RdlUrpY6yQUPR+PZjfdbCZEgZsRkIIon4VVKGrIaWaABv4rH9RBT0IPfMgX
zsE1wE6Y5icGb2MPhFj05zikkDby3Gl6MBMd87M4x/e2Sfe/nf0rfu+n2kdxwXkA
EQEAAYkBHwQYAQIACQUCVgSPNgIbDAAKCRD+ebtS1dxS3BgMB/wMYq6wXrfC8LCB
130cZPLa3Aq5xP0IL0dtr+ur4MIffZcj97Lush/GnEJf0UWLZjFpsKbhlXt+cVPV
FgHp/Weo0qDQ6cxk8wpGfgRfmA9u1Wz4iZpDYcC/g9E1Z0uSvxuDz4d3+iO+yc1e
0i8D2bzatSSeXs5FwVRdgEZTUcnIZl+Wvn7J9P70gQkUn+rL791AUpIixhB/k5QN
alz10wyGWq+IROH3KvFE6MQuk4062M99+wjHjokHF/FdZUznCFJxECIoAPsLv5OX
bgKtHEB6OHVOzAdO0yAO39BTup8Wk6jF1tYT+ovNcxKEsUYsAWUgmHa4JVq+tvaD
Cb9v1xIj
=O+RH
-----END PGP PUBLIC KEY BLOCK-----
EOF

# Add Gluster repo
cat <<EOF > /etc/yum.repos.d/gluster.repo
[gluster]
name=Gluster
baseurl=$1/gluster
enabled=1
gpgcheck=1
gpgkey=file:///tmp/gluster.gpg
EOF

# Need to clean cache to download metadata again
yum clean all
yum makecache