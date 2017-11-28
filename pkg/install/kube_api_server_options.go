package install

import (
	"fmt"
	"strings"
)

var kubeAPIServerProtectedOptions = []string{
	"advertise-address",
	"apiserver-count",
	"client-ca-file",
	"cloud-provider",
	"cloud-config",
	"etcd-cafile",
	"etcd-certfile",
	"etcd-keyfile",
	"etcd-servers",
	"insecure-port",
	"secure-port",
	"service-account-key-file",
	"service-cluster-ip-range",
	"tls-cert-file",
	"tls-private-key-file",
}

func (options *APIServerOptions) validate() (bool, []error) {
	v := newValidator()
	overrides := make([]string, 0)
	for _, protectedOption := range kubeAPIServerProtectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			overrides = append(overrides, protectedOption)
		}
	}

	if len(overrides) > 0 {
		v.addError(fmt.Errorf("Kube ApiServer Option(s) [%v] cannot be overridden", strings.Join(overrides, ", ")))
	}

	return v.valid()
}
