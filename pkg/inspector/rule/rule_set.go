package rule

import (
	"io"
	"strings"
)

// DefaultRuleSet is the list of rules that are built into the inspector
const defaultRuleSet = `---
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

# Ports used by K8s master are accessible
# Port 8080 is not accessible from outside
- kind: TCPPortAccessible
  when: ["master"]
  port: 6443
  timeout: 5s

# Port used by Docker registry
- kind: TCPPortAvailable
  when: ["master"]
  port: 8443
- kind: TCPPortAccessible
  when: ["master"]
  port: 8443
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

- kind: PackageAvailable
  when: ["etcd", "ubuntu"]
  packageName: kismatic-etcd
  packageVersion: 1.5.2-3
- kind: PackageAvailable
  when: ["master","ubuntu"]
  packageName: kismatic-kubernetes-master
  packageVersion: 1.5.2-3
- kind: PackageAvailable
  when: ["master","ubuntu", "disconnected"]
  packageName: kismatic-offline
  packageVersion: 1.5.2-3
- kind: PackageAvailable
  when: ["worker","ubuntu"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2-3
- kind: PackageAvailable
  when: ["ingress","ubuntu"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2-3
- kind: PackageAvailable
  when: ["storage","ubuntu"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2-3

- kind: PackageAvailable
  when: ["etcd", "centos"]
  packageName: kismatic-etcd
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["master","centos"]
  packageName: kismatic-kubernetes-master
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["master","centos", "disconnected"]
  packageName: kismatic-offline
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["worker","centos"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["ingress","centos"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["storage","centos"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1

- kind: PackageAvailable
  when: ["etcd", "rhel"]
  packageName: kismatic-etcd
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["master","rhel"]
  packageName: kismatic-kubernetes-master
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["master","rhel", "disconnected"]
  packageName: kismatic-offline
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["worker","rhel"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["ingress","rhel"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1
- kind: PackageAvailable
  when: ["storage","rhel"]
  packageName: kismatic-kubernetes-node
  packageVersion: 1.5.2_3-1

# Gluster packages
- kind: PackageAvailable
  when: ["storage", "centos"]
  packageName: glusterfs-server
  packageVersion: 3.8.7-1.el7
- kind: PackageAvailable
  when: ["storage", "rhel"]
  packageName: glusterfs-server
  packageVersion: 3.8.7-1.el7
- kind: PackageAvailable
  when: ["storage", "ubuntu"]
  packageName: glusterfs-server
  packageVersion: 3.8.7-ubuntu1~xenial1

# Port required for gluster-healthz
- kind: TCPPortAvailable
  when: ["storage"]
  port: 8081
- kind: TCPPortAccessible
  when: ["storage"]
  port: 8081
  timeout: 5s

# Ports required for NFS
- kind: TCPPortAvailable
  when: ["storage"]
  port: 111
- kind: TCPPortAccessible
  when: ["storage"]
  port: 111
  timeout: 5s
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
