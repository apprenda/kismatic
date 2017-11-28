package install

import (
	"fmt"
	"strings"
)

var kubeProxyProtectedOptions = []string{
	"cluster-cidr",
	"hostname-override",
	"kubeconfig",
}

func (options *KubeProxyOptions) validate() (bool, []error) {
	v := newValidator()
	overrides := make([]string, 0)
	for _, protectedOption := range kubeProxyProtectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			overrides = append(overrides, protectedOption)
		}
	}

	if len(overrides) > 0 {
		v.addError(fmt.Errorf("Kube Proxy Option(s) [%v] cannot be overridden", strings.Join(overrides, ", ")))
	}

	return v.valid()
}
