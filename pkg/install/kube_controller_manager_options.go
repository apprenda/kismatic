package install

import (
	"fmt"
	"strings"
)

var kubeControllerManagerProtectedOptions = []string{
	"cloud-provider",
	"cloud-config",
	"cluster-cidr",
	"cluster-name",
	"kubeconfig",
	"root-ca-file",
	"service-account-private-key-file",
	"service-cluster-ip-range",
}

func (options *KubeControllerManagerOptions) validate() (bool, []error) {
	v := newValidator()
	overrides := make([]string, 0)
	for _, protectedOption := range kubeControllerManagerProtectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			overrides = append(overrides, protectedOption)
		}
	}

	if len(overrides) > 0 {
		v.addError(fmt.Errorf("Kube Controller Manager Option(s) [%v] cannot be overridden", strings.Join(overrides, ", ")))
	}

	return v.valid()
}
