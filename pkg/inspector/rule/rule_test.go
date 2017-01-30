package rule

import "testing"

func TestDefaultRules(t *testing.T) {
	// This will panic if there are errors in the default rule
	rules := DefaultRules()
	for _, r := range rules {
		if errs := r.Validate(); len(errs) != 0 {
			t.Errorf("invalid default rule was found: %+v. Errors are: %v", r, errs)
		}
	}
}

func TestUpgradeRules(t *testing.T) {
	// This will panic if there are errors in the upgrade rule
	rules := UpgradeRules()
	for _, r := range rules {
		if errs := r.Validate(); len(errs) != 0 {
			t.Errorf("invalid default rule was found: %+v. Errors are: %v", r, errs)
		}
	}
}
