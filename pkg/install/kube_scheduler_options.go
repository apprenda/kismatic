package install

import (
	"fmt"
	"strings"
)

var kubeSchedulerProtectedOptions = []string{
	"kubeconfig",
}

func (options *KubeSchedulerOptions) validate() (bool, []error) {
	v := newValidator()
	overrides := make([]string, 0)
	for _, protectedOption := range kubeSchedulerProtectedOptions {
		_, found := options.Overrides[protectedOption]
		if found {
			overrides = append(overrides, protectedOption)
		}
	}

	if len(overrides) > 0 {
		v.addError(fmt.Errorf("Kube Scheduler Option(s) [%v] cannot be overridden", strings.Join(overrides, ", ")))
	}

	return v.valid()
}
