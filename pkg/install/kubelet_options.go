package install

import (
	"fmt"
	"strings"
)

var kubeletProtectedOptions = []string{
	"cloud-provider",
	"cloud-config",
	"cluster-dns",
	"container-runtime",
	"cni-bin-dir",
	"cni-conf-dir",
	"network-plugin",
	"docker",
	"hostname-override",
	"require-kubeconfig",
	"kubeconfig",
	"node-labels",
	"node-ip",
	"pod-manifest-path",
	"tls-cert-file",
	"tls-private-key-file",
}

func (options *KubeletOptions) validate() (bool, []error) {
	v := newValidator()
	overrides := make([]string, 0)
	for _, protectedOption := range kubeletProtectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			overrides = append(overrides, protectedOption)
		}
	}

	if len(overrides) > 0 {
		v.addError(fmt.Errorf("Kubelet Option(s) [%v] cannot be overridden", strings.Join(overrides, ", ")))
	}

	return v.valid()
}
