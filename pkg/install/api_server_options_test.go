package install

import (
	"testing"
	"fmt"
	"reflect"
)

func TestValidateFailsForOverridingProtectedValue(t *testing.T) {
	options := APIServerOptions{
		Overrides: map[string]string{
			"advertise-address": "1.2.3.4",
		},
	}

	ok, err := options.validate()

	assertEqual(t, ok, false)
	assertEqual(t, err, []error{fmt.Errorf("APIServer option [%s] should not be overriden", "advertise-address")})
}

func TestValidatePassesForNoValues(t *testing.T) {

	options := APIServerOptions{
	}

	ok, _ := options.validate()

	assertEqual(t, ok, true)
}

func TestValidatePassesForUnprotectedValues(t *testing.T) {
	options := APIServerOptions{
		Overrides: map[string]string{
			"foobar":"baz",
		},
	}

	ok, _ := options.validate()

	assertEqual(t, ok, true)
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}
