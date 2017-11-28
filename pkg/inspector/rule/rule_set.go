package rule

import (
	"io"
	"strings"
)

// DefaultRuleSet is the list of rules that are built into the inspector
const defaultRuleSet = `---
- kind: FreeSpace
  path: /
  minimumBytes: 1000000000

# Python 2.5+ is installed on all nodes
# This is required by ansible
- kind: Python2Version
  when: []
  supportedVersions:
   - Python 2.5
   - Python 2.6
   - Python 2.7

# Executables required by kubelet
- kind: ExecutableInPath
  when: ["master","worker"]
  executable: iptables
- kind: ExecutableInPath
  when: ["master","worker"]
  executable: iptables-save
- kind: ExecutableInPath
  when: ["master","worker"]
  executable: iptables-restore

# Ports used by etcd are available
- kind: TCPPortAvailable
  when: ["etcd"]
  port: 2379
- kind: TCPPortAvailable
  when: ["etcd"]
  port: 6666
- kind: TCPPortAvailable
  when: ["etcd"]
  port: 2380
- kind: TCPPortAvailable
  when: ["etcd"]
  port: 6660

# Ports used by etcd are accessible
- kind: TCPPortAccessible
  when: ["etcd"]
  port: 2379
  timeout: 5s
- kind: TCPPortAccessible
  when: ["etcd"]
  port: 6666
  timeout: 5s
- kind: TCPPortAccessible
  when: ["etcd"]
  port: 2380
  timeout: 5s
- kind: TCPPortAccessible
  when: ["etcd"]
  port: 6660
  timeout: 5s

# Ports used by K8s master are available
- kind: TCPPortAvailable
  when: ["master"]
  port: 6443
- kind: TCPPortAvailable
  when: ["master"]
  port: 8080
# kube-scheduler
- kind: TCPPortAvailable
  when: ["master"]
  port: 10251
# kube-controller-manager
- kind: TCPPortAvailable
  when: ["master"]
  port: 10252

# Ports used by K8s master are accessible
# Port 8080 is not accessible from outside
- kind: TCPPortAccessible
  when: ["master"]
  port: 6443
  timeout: 5s
# kube-scheduler
- kind: TCPPortAccessible
  when: ["master"]
  port: 10251
  timeout: 5s
# kube-controller-manager
- kind: TCPPortAccessible
  when: ["master"]
  port: 10252
  timeout: 5s

# Ports used by K8s worker are available
# cAdvisor
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 4194
# kubelet localhost healthz
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 10248
# kube-proxy metrics
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 10249
# kube-proxy health
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 10256
# kubelet
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 10250
# kubelet no auth
- kind: TCPPortAvailable
  when: ["master","worker","ingress","storage"]
  port: 10255

# Ports used by K8s worker are accessible
# cAdvisor
- kind: TCPPortAccessible
  when: ["master","worker","ingress","storage"]
  port: 4194
  timeout: 5s
# kube-proxy
- kind: TCPPortAccessible
  when: ["master","worker","ingress","storage"]
  port: 10256
  timeout: 5s
# kubelet
- kind: TCPPortAccessible
  when: ["master","worker","ingress","storage"]
  port: 10250
  timeout: 5s

# Port used by Ingress
- kind: TCPPortAvailable
  when: ["ingress"]
  port: 80
- kind: TCPPortAccessible
  when: ["ingress"]
  port: 80
  timeout: 5s
- kind: TCPPortAvailable
  when: ["ingress"]
  port: 443
- kind: TCPPortAccessible
  when: ["ingress"]
  port: 443
  timeout: 5s
# healthz
- kind: TCPPortAvailable
  when: ["ingress"]
  port: 10254
- kind: TCPPortAccessible
  when: ["ingress"]
  port: 10254
  timeout: 5s


- kind: PackageDependency
  when: ["etcd","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00


- kind: PackageDependency
  when: ["etcd","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["master","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["master","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0


- kind: PackageDependency
  when: ["etcd","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0


# Gluster packages
- kind: PackageDependency
  when: ["storage", "centos"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: ["storage", "rhel"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: ["storage", "ubuntu"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-ubuntu1~xenial1

# Port required for gluster-healthz
- kind: TCPPortAvailable
  when: ["storage"]
  port: 8081
- kind: TCPPortAccessible
  when: ["storage"]
  port: 8081
  timeout: 5s

# Ports required for NFS
# Removed due to https://github.com/apprenda/kismatic/issues/784
#- kind: TCPPortAvailable
#  when: ["storage"]
#  port: 111
#- kind: TCPPortAccessible
#  when: ["storage"]
#  port: 111
#  timeout: 5s
- kind: TCPPortAvailable
  when: ["storage"]
  port: 2049
- kind: TCPPortAccessible
  when: ["storage"]
  port: 2049
  timeout: 5s
- kind: TCPPortAvailable
  when: ["storage"]
  port: 38465
- kind: TCPPortAccessible
  when: ["storage"]
  port: 38465
  timeout: 5s
- kind: TCPPortAvailable
  when: ["storage"]
  port: 38466
- kind: TCPPortAccessible
  when: ["storage"]
  port: 38466
  timeout: 5s
- kind: TCPPortAvailable
  when: ["storage"]
  port: 38467
- kind: TCPPortAccessible
  when: ["storage"]
  port: 38467
  timeout: 5s
`

const upgradeRuleSet = `---
- kind: FreeSpace
  path: /
  minimumBytes: 1000000000

- kind: PackageDependency
  when: ["etcd","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["master","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: docker-engine
  packageVersion: 1.12.6-0~ubuntu-xenial
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: kubelet
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: nfs-common
  anyVersion: true
- kind: PackageDependency
  when: ["worker","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["ingress","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00
- kind: PackageDependency
  when: ["storage","ubuntu"]
  packageName: kubectl
  packageVersion: 1.8.4-00

- kind: PackageDependency
  when: ["etcd","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["master","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["master","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: kubectl
  packageVersion: 1.8.4-0

- kind: PackageDependency
  when: ["etcd","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["master","rhel"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["ingress","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["storage","centos"]
  packageName: docker-engine
  packageVersion: 1.12.6-1.el7.centos
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: kubelet
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: nfs-utils
  anyVersion: true
- kind: PackageDependency
  when: ["worker","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["ingress","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0
- kind: PackageDependency
  when: ["storage","rhel"]
  packageName: kubectl
  packageVersion: 1.8.4-0

# Gluster packages
- kind: PackageDependency
  when: ["storage", "centos"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: ["storage", "rhel"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: ["storage", "ubuntu"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-ubuntu1~xenial1
`

// DefaultRules returns the list of rules that are built into the inspector
func DefaultRules() []Rule {
	rules, err := UnmarshalRulesYAML([]byte(defaultRuleSet))
	if err != nil {
		// The default rules should not contain errors
		// If they do, panic so that we catch them during tests
		panic(err)
	}
	return rules
}

// DumpDefaultRules writes the default rule set to a file
func DumpDefaultRules(writer io.Writer) error {
	_, err := io.Copy(writer, strings.NewReader(defaultRuleSet))
	if err != nil {
		return err
	}
	return nil
}

func UpgradeRules() []Rule {
	rules, err := UnmarshalRulesYAML([]byte(upgradeRuleSet))
	if err != nil {
		// The upgrade rules should not contain errors
		// If they do, panic so that we catch them during tests
		panic(err)
	}
	return rules
}
