package install

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateKubeSchedulerOptions(t *testing.T) {
	tests := []struct {
		opts            KubeSchedulerOptions
		valid           bool
		protectedFields []string
	}{
		{
			opts:  KubeSchedulerOptions{},
			valid: true,
		},
		{
			opts: KubeSchedulerOptions{
				Overrides: map[string]string{
					"foobar": "baz",
				},
			},
			valid: true,
		},
		{
			opts: KubeSchedulerOptions{
				Overrides: map[string]string{
					"kubeconfig": "/foo/.kube/config",
				},
			},
			valid:           false,
			protectedFields: []string{"kubeconfig"},
		},
		{
			opts: KubeSchedulerOptions{
				Overrides: map[string]string{
					"kubeconfig": "/foo/.kube/config",
					"v":          "3",
				},
			},
			valid:           false,
			protectedFields: []string{"kubeconfig"},
		},
	}
	for _, test := range tests {
		ok, err := test.opts.validate()
		assertEqual(t, ok, test.valid)
		if !test.valid {
			assertEqual(t, err, []error{fmt.Errorf("Kube Scheduler Option(s) [%v] cannot be overridden", strings.Join(test.protectedFields, ", "))})
		}
	}
}
