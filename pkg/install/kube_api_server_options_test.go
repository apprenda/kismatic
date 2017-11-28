package install

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestValidateKubeApiServerOptions(t *testing.T) {
	tests := []struct {
		opts            APIServerOptions
		valid           bool
		protectedFields []string
	}{
		{
			opts:  APIServerOptions{},
			valid: true,
		},
		{
			opts: APIServerOptions{
				Overrides: map[string]string{
					"foobar": "baz",
				},
			},
			valid: true,
		},
		{
			opts: APIServerOptions{
				Overrides: map[string]string{
					"advertise-address": "1.2.3.4",
				},
			},
			valid:           false,
			protectedFields: []string{"advertise-address"},
		},
		{
			opts: APIServerOptions{
				Overrides: map[string]string{
					"advertise-address": "1.2.3.4",
					"secure-port":       "123",
				},
			},
			valid:           false,
			protectedFields: []string{"advertise-address", "secure-port"},
		},
		{
			opts: APIServerOptions{
				Overrides: map[string]string{
					"advertise-address": "1.2.3.4",
					"secure-port":       "123",
					"v":                 "3",
				},
			},
			valid:           false,
			protectedFields: []string{"advertise-address", "secure-port"},
		},
	}
	for _, test := range tests {
		ok, err := test.opts.validate()
		assertEqual(t, ok, test.valid)
		if !test.valid {
			assertEqual(t, err, []error{fmt.Errorf("Kube ApiServer Option(s) [%v] cannot be overridden", strings.Join(test.protectedFields, ", "))})
		}
	}
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}
