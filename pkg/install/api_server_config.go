package install

import (
	"fmt"
)

type APIServerConfig map[string]string

var protectedValues = []string{
	"advertise-address",
	"apiserver-count",
	"authorization-policy-file",
	"basic-auth-file",
	"client-ca-file",
	"etcd-cafile",
	"etcd-certfile",
	"etcd-keyfile",
	"etcd-servers",
	"service-account-key-file",
	"service-cluster-ip-range",
	"tls-cert-file",
	"tls-private-key-file",
	"v",
}

func (config *APIServerConfig) validate() (bool, []error) {
	v := newValidator()
	for _, protectedItem := range protectedValues {
		_, found := (*config)[protectedItem]
		if found {
			v.addError(fmt.Errorf("Api config value [%s] should not be overriden", protectedItem))
		}
	}
	return v.valid();
}