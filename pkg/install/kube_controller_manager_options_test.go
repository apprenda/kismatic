package install

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateKubeControllerManagerOptions(t *testing.T) {
	tests := []struct {
		opts            KubeControllerManagerOptions
		valid           bool
		protectedFields []string
	}{
		{
			opts:  KubeControllerManagerOptions{},
			valid: true,
		},
		{
			opts: KubeControllerManagerOptions{
				Overrides: map[string]string{
					"foobar": "baz",
				},
			},
			valid: true,
		},
		{
			opts: KubeControllerManagerOptions{
				Overrides: map[string]string{
					"cluster-cidr": "1.2.3.4",
				},
			},
			valid:           false,
			protectedFields: []string{"cluster-cidr"},
		},
		{
			opts: KubeControllerManagerOptions{
				Overrides: map[string]string{
					"cluster-cidr": "1.2.3.4",
					"kubeconfig":   "/foo/.kube/config",
				},
			},
			valid:           false,
			protectedFields: []string{"cluster-cidr", "kubeconfig"},
		},
		{
			opts: KubeControllerManagerOptions{
				Overrides: map[string]string{
					"cluster-cidr": "1.2.3.4",
					"kubeconfig":   "/foo/.kube/config",
					"v":            "3",
				},
			},
			valid:           false,
			protectedFields: []string{"cluster-cidr", "kubeconfig"},
		},
	}
	for _, test := range tests {
		ok, err := test.opts.validate()
		assertEqual(t, ok, test.valid)
		if !test.valid {
			assertEqual(t, err, []error{fmt.Errorf("Kube Controller Manager Option(s) [%v] cannot be overridden", strings.Join(test.protectedFields, ", "))})
		}
	}
}
