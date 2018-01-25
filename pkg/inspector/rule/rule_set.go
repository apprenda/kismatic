package rule

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

/*
- kind: __RuleName__
  when:
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel", "centos"]
  ...

This rule will be executed when the node has these facts:
  ("etcd" OR "master" OR "worker" OR "ingress" OR "storage") AND ("rhel" OR "centos")
*/

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
  when:
  - ["master", "worker", "ingress", "storage"]
  executable: iptables
- kind: ExecutableInPath
  when:
  - ["master", "worker", "ingress", "storage"]
  executable: iptables-save
- kind: ExecutableInPath
  when:
  - ["master", "worker", "ingress", "storage"]
  executable: iptables-restore

# Docker should be installed when installation is disabled
- kind: DockerInPath
  when:
  - ["etcd", "master", "worker", "ingress", "storage"]
  
# Ports used by etcd are available
- kind: TCPPortAvailable
  when: 
  - ["etcd"]
  port: 2379
  procName: docker-proxy # docker sets up a proxy for the etcd container
- kind: TCPPortAvailable
  when: 
  - ["etcd"]
  port: 6666
  procName: docker-proxy # docker sets up a proxy for the etcd container
- kind: TCPPortAvailable
  when: 
  - ["etcd"]
  port: 2380
  procName: docker-proxy # docker sets up a proxy for the etcd container
- kind: TCPPortAvailable
  when: 
  - ["etcd"]
  port: 6660
  procName: docker-proxy # docker sets up a proxy for the etcd container

# Ports used by etcd are accessible
- kind: TCPPortAccessible
  when: 
  - ["etcd"]
  port: 2379
  timeout: 5s
- kind: TCPPortAccessible
  when: 
  - ["etcd"]
  port: 6666
  timeout: 5s
- kind: TCPPortAccessible
  when: 
  - ["etcd"]
  port: 2380
  timeout: 5s
- kind: TCPPortAccessible
  when: 
  - ["etcd"]
  port: 6660
  timeout: 5s

# Ports used by K8s master are available
- kind: TCPPortAvailable
  when: 
  - ["master"]
  port: 6443
  procName: kube-apiserver
- kind: TCPPortAvailable
  when: 
  - ["master"]
  port: 8080
  procName: kube-apiserver
# kube-scheduler
- kind: TCPPortAvailable
  when: 
  - ["master"]
  port: 10251
  procName: kube-scheduler
# kube-controller-manager
- kind: TCPPortAvailable
  when: 
  - ["master"]
  port: 10252
  procName: kube-controller

# Ports used by K8s master are accessible
# Port 8080 is not accessible from outside
- kind: TCPPortAccessible
  when: 
  - ["master"]
  port: 6443
  timeout: 5s
# kube-scheduler
- kind: TCPPortAccessible
  when: 
  - ["master"]
  port: 10251
  timeout: 5s
# kube-controller-manager
- kind: TCPPortAccessible
  when: 
  - ["master"]
  port: 10252
  timeout: 5s

# Ports used by K8s worker are available
# cAdvisor
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 4194
  procName: kubelet
# kubelet localhost healthz
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10248
  procName: kubelet
# kube-proxy metrics
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10249
  procName: kube-proxy
# kube-proxy health
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10256
  procName: kube-proxy
# kubelet
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10250
  procName: kubelet
# kubelet no auth
- kind: TCPPortAvailable
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10255
  procName: kubelet

# Ports used by K8s worker are accessible
# cAdvisor
- kind: TCPPortAccessible
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 4194
  timeout: 5s
# kube-proxy
- kind: TCPPortAccessible
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10256
  timeout: 5s
# kubelet
- kind: TCPPortAccessible
  when: 
  - ["master", "worker", "ingress", "storage"]
  port: 10250
  timeout: 5s

# Port used by Ingress
- kind: TCPPortAvailable
  when: 
  - ["ingress"]
  port: 80
  procName: nginx
- kind: TCPPortAccessible
  when: 
  - ["ingress"]
  port: 80
  timeout: 5s
- kind: TCPPortAvailable
  when: 
  - ["ingress"]
  port: 443
  procName: nginx
- kind: TCPPortAccessible
  when: 
  - ["ingress"]
  port: 443
  timeout: 5s
# healthz
- kind: TCPPortAvailable
  when: 
  - ["ingress"]
  port: 10254
  procName: nginx-ingress-c
- kind: TCPPortAccessible
  when: 
  - ["ingress"]
  port: 10254
  timeout: 5s

# Port required for gluster-healthz
- kind: TCPPortAvailable
  when: 
  - ["storage"]
  port: 8081
  procName: exechealthz
- kind: TCPPortAccessible
  when: 
  - ["storage"]
  port: 8081
  timeout: 5s

# Ports required for NFS
# Removed due to https://github.com/apprenda/kismatic/issues/784
#- kind: TCPPortAvailable
#  when: 
#  - ["storage"]
#  port: 111
#- kind: TCPPortAccessible
#  when: 
#  - ["storage"]
#  port: 111
#  timeout: 5s
- kind: TCPPortAvailable
  when: 
  - ["storage"]
  port: 2049
  procName: glusterfs
- kind: TCPPortAccessible
  when: 
  - ["storage"]
  port: 2049
  timeout: 5s
- kind: TCPPortAvailable
  when: 
  - ["storage"]
  port: 38465
  procName: glusterfs
- kind: TCPPortAccessible
  when: 
  - ["storage"]
  port: 38465
  timeout: 5s
- kind: TCPPortAvailable
  when: 
  - ["storage"]
  port: 38466
  procName: glusterfs
- kind: TCPPortAccessible
  when: 
  - ["storage"]
  port: 38466
  timeout: 5s
- kind: TCPPortAvailable
  when: 
  - ["storage"]
  port: 38467
  procName: glusterfs
- kind: TCPPortAccessible
  when: 
  - ["storage"]
  port: 38467
  timeout: 5s
  
- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker-ce
  packageVersion: 17.03.2~ce-0~ubuntu-xenial
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: kubelet
  packageVersion: {{.kubernetes_deb_version}}
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: nfs-common
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: kubectl
  packageVersion: {{.kubernetes_deb_version}}
# https://docs.docker.com/engine/installation/linux/docker-ee/ubuntu/#uninstall-old-versions
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker-engine
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker-ce
  acceptablePackageVersion: 17.03.2~ce-0~ubuntu-xenial
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker-ee

- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-ce
  packageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: kubelet
  packageVersion: {{.kubernetes_yum_version}}
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: nfs-utils
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: kubectl
  packageVersion: {{.kubernetes_yum_version}}
# https://docs.docker.com/engine/installation/linux/docker-ee/centos/
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-common
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-selinux
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-engine-selinux
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-engine
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-ce
  acceptablePackageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-ee

- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-ce
  packageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageDependency
  when: 
  - [master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: kubelet
  packageVersion: {{.kubernetes_yum_version}}
- kind: PackageDependency
  when: 
  - [master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: nfs-utils
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: kubectl
  packageVersion: {{.kubernetes_yum_version}}
# https://docs.docker.com/engine/installation/linux/docker-ee/rhel/#os-requirements
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-common
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-selinux
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-engine-selinux
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-engine
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-ce
  acceptablePackageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageNotInstalled
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-ee

# Gluster packages
- kind: PackageDependency
  when: 
  - ["storage"]
  - ["centos"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: 
  - ["storage"]
  - ["rhel"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: 
  - ["storage"] 
  - ["ubuntu"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-ubuntu1~xenial1
`

const upgradeRuleSet = `---
- kind: FreeSpace
  path: /
  minimumBytes: 1000000000
  
- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: docker-ce
  packageVersion: 17.03.2~ce-0~ubuntu-xenial
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: kubelet
  packageVersion: {{.kubernetes_deb_version}}
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: nfs-common
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["ubuntu"]
  packageName: kubectl
  packageVersion: {{.kubernetes_deb_version}}

- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: docker-ce
  packageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: kubelet
  packageVersion: {{.kubernetes_yum_version}}
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: nfs-utils
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["centos"]
  packageName: kubectl
  packageVersion: {{.kubernetes_yum_version}}

- kind: PackageDependency
  when: 
  - ["etcd", "master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: docker-ce
  packageVersion: 17.03.2.ce-1.el7.centos
- kind: PackageDependency
  when: 
  - [master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: kubelet
  packageVersion: {{.kubernetes_yum_version}}
- kind: PackageDependency
  when: 
  - [master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: nfs-utils
- kind: PackageDependency
  when: 
  - ["master", "worker", "ingress", "storage"]
  - ["rhel"]
  packageName: kubectl
  packageVersion: {{.kubernetes_yum_version}}

# Gluster packages
- kind: PackageDependency
  when: 
  - ["storage"]
  - ["centos"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: 
  - ["storage"]
  - ["rhel"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-2.el7
- kind: PackageDependency
  when: 
  - ["storage"] 
  - ["ubuntu"]
  packageName: glusterfs-server
  packageVersion: 3.8.15-ubuntu1~xenial1
`

// DefaultRules returns the list of rules that are built into the inspector
func DefaultRules(vars map[string]string) []Rule {
	tmpl, err := template.New("").Parse(defaultRuleSet)
	if err != nil {
		panic(fmt.Errorf("error parsing rules: %v", err))
	}
	var rawRules bytes.Buffer
	err = tmpl.Execute(&rawRules, vars)
	if err != nil {
		panic(fmt.Errorf("error reading rules from: %v", err))
	}
	rules, err := UnmarshalRulesYAML(rawRules.Bytes())
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

func UpgradeRules(vars map[string]string) []Rule {
	tmpl, err := template.New("").Parse(upgradeRuleSet)
	if err != nil {
		panic(fmt.Errorf("error parsing rules: %v", err))
	}
	fmt.Printf("template: %v+\n", tmpl.Tree)
	var rawRules bytes.Buffer
	err = tmpl.Execute(&rawRules, vars)
	if err != nil {
		panic(fmt.Errorf("error reading rules from: %v", err))
	}
	fmt.Printf("raw rules: %v\n", rawRules.String())
	rules, err := UnmarshalRulesYAML(rawRules.Bytes())
	if err != nil {
		// The upgrade rules should not contain errors
		// If they do, panic so that we catch them during tests
		panic(err)
	}
	return rules
}
