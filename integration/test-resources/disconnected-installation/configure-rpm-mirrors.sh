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

mQINBFWln24BEADrBl5p99uKh8+rpvqJ48u4eTtjeXAWbslJotmC/CakbNSqOb9o
ddfzRvGVeJVERt/Q/mlvEqgnyTQy+e6oEYN2Y2kqXceUhXagThnqCoxcEJ3+KM4R
mYdoe/BJ/J/6rHOjq7Omk24z2qB3RU1uAv57iY5VGw5p45uZB4C4pNNsBJXoCvPn
TGAs/7IrekFZDDgVraPx/hdiwopQ8NltSfZCyu/jPpWFK28TR8yfVlzYFwibj5WK
dHM7ZTqlA1tHIG+agyPf3Rae0jPMsHR6q+arXVwMccyOi+ULU0z8mHUJ3iEMIrpT
X+80KaN/ZjibfsBOCjcfiJSB/acn4nxQQgNZigna32velafhQivsNREFeJpzENiG
HOoyC6qVeOgKrRiKxzymj0FIMLru/iFF5pSWcBQB7PYlt8J0G80lAcPr6VCiN+4c
NKv03SdvA69dCOj79PuO9IIvQsJXsSq96HB+TeEmmL+xSdpGtGdCJHHM1fDeCqkZ
hT+RtBGQL2SEdWjxbF43oQopocT8cHvyX6Zaltn0svoGs+wX3Z/H6/8P5anog43U
65c0A+64Jj00rNDr8j31izhtQMRo892kGeQAaaxg4Pz6HnS7hRC+cOMHUU4HA7iM
zHrouAdYeTZeZEQOA7SxtCME9ZnGwe2grxPXh/U/80WJGkzLFNcTKdv+rwARAQAB
tDdEb2NrZXIgUmVsZWFzZSBUb29sIChyZWxlYXNlZG9ja2VyKSA8ZG9ja2VyQGRv
Y2tlci5jb20+iQI4BBMBAgAiBQJVpZ9uAhsvBgsJCAcDAgYVCAIJCgsEFgIDAQIe
AQIXgAAKCRD3YiFXLFJgnbRfEAC9Uai7Rv20QIDlDogRzd+Vebg4ahyoUdj0CH+n
Ak40RIoq6G26u1e+sdgjpCa8jF6vrx+smpgd1HeJdmpahUX0XN3X9f9qU9oj9A4I
1WDalRWJh+tP5WNv2ySy6AwcP9QnjuBMRTnTK27pk1sEMg9oJHK5p+ts8hlSC4Sl
uyMKH5NMVy9c+A9yqq9NF6M6d6/ehKfBFFLG9BX+XLBATvf1ZemGVHQusCQebTGv
0C0V9yqtdPdRWVIEhHxyNHATaVYOafTj/EF0lDxLl6zDT6trRV5n9F1VCEh4Aal8
L5MxVPcIZVO7NHT2EkQgn8CvWjV3oKl2GopZF8V4XdJRl90U/WDv/6cmfI08GkzD
YBHhS8ULWRFwGKobsSTyIvnbk4NtKdnTGyTJCQ8+6i52s+C54PiNgfj2ieNn6oOR
7d+bNCcG1CdOYY+ZXVOcsjl73UYvtJrO0Rl/NpYERkZ5d/tzw4jZ6FCXgggA/Zxc
jk6Y1ZvIm8Mt8wLRFH9Nww+FVsCtaCXJLP8DlJLASMD9rl5QS9Ku3u7ZNrr5HWXP
HXITX660jglyshch6CWeiUATqjIAzkEQom/kEnOrvJAtkypRJ59vYQOedZ1sFVEL
MXg2UCkD/FwojfnVtjzYaTCeGwFQeqzHmM241iuOmBYPeyTY5veF49aBJA1gEJOQ
TvBR8Q==
=Fm3p
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