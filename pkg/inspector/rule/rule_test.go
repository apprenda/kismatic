package rule

import "testing"

func TestDefaultRules(t *testing.T) {
	// This will panic if there are errors in the default rule
	rules := DefaultRules(map[string]string{"kubernetes_yum_version": "1.9.3-0", "kubernetes_deb_version": "1.9.3-00"})
	if len(rules) != 76 {
		t.Errorf("expected to have %d rules, instead got %d", 76, len(rules))
	}
	for _, r := range rules {
		if errs := r.Validate(); len(errs) != 0 {
			t.Errorf("invalid default rule was found: %+v. Errors are: %v", r, errs)
		}
	}
}

func TestUpgradeRules(t *testing.T) {
	// This will panic if there are errors in the upgrade rule
	rules := UpgradeRules(map[string]string{"kubernetes_yum_version": "1.9.3-0", "kubernetes_deb_version": "1.9.3-00"})
	if len(rules) != 16 {
		t.Errorf("expected to have %d rules, instead got %d", 16, len(rules))
	}
	for _, r := range rules {
		if errs := r.Validate(); len(errs) != 0 {
			t.Errorf("invalid default rule was found: %+v. Errors are: %v", r, errs)
		}
	}
}
