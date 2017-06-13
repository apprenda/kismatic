package install

import (
	"testing"
	"fmt"
	"reflect"
)

func TestValidateFailsForOverridingProtectedValue(t *testing.T) {
	config := APIServerConfig{
			"advertise-address": "1.2.3.4",
	}

	ok, err := config.validate()

	assertEqual(t, ok, false)
	assertEqual(t, err, []error{fmt.Errorf("Api config value [%s] should not be overriden", "advertise-address")})
}

func TestValidatePassesForNoValues(t *testing.T) {

	config := APIServerConfig{
	}

	ok, _ := config.validate()

	assertEqual(t, ok, true)
}

func TestValidatePassesForUnprotectedValues(t *testing.T) {
	config := APIServerConfig{
		"foobar":"baz",
	}

	ok, _ := config.validate()

	assertEqual(t, ok, true)
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}
