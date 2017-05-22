package install

import (
	"testing"
	"fmt"
)

func TestValidateFailsForOverridingProtectedValue(t *testing.T) {
	config := APIServerConfig{
		map[string]string{
			"advertise-address": "1.2.3.4",
		},
	}

	ok, err := config.validate()

	assertEqual(t, ok, false)
	assertEqual(t, err, []error{fmt.Errorf("Api config value [%s] should not be overriden", "advertise-address")})
}

func TestValidatePassesForNoValues(t *testing.T) {

	config := APIServerConfig{
		map[string]string{},
	}

	ok, _ := config.validate()

	assertEqual(t, ok, true)
}

func TestValidatePassesForUnprotectedValues(t *testing.T) {

	config := APIServerConfig{
		map[string]string{"foobar":"baz"},
	}

	ok, _ := config.validate()

	assertEqual(t, ok, true)
}
