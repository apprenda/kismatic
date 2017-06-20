package install

import (
	"fmt"
)

type APIServerOptions struct {
  Overrides  map[string]string `yaml:"option_overrides"`
}

var protectedOptions = []string{
	"advertise-address",
	"apiserver-count",
	"client-ca-file",
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
	for _, protectedOption := range protectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			v.addError(fmt.Errorf("APIServer option [%s] should not be overriden", protectedOption))
		}
	}
	return v.valid();
}