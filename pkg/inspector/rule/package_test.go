package rule

import "testing"

func TestPackageDependencyRuleValidation(t *testing.T) {
	p := PackageDependency{}
	errs := p.Validate()
	if len(errs) != 1 {
		t.Errorf("expected 1 errors, but got %d", len(errs))
	}
	p.PackageName = "foo"
	if errs := p.Validate(); len(errs) != 0 {
		t.Errorf("expected to be valid, but got %d", len(errs))
	}
	p.PackageVersion = ""
	if errs := p.Validate(); len(errs) != 0 {
		t.Errorf("expected to be valid, but got %v", errs)
	}
	p.PackageVersion = "1.0"
	if errs := p.Validate(); len(errs) != 0 {
		t.Errorf("expected to be valid, but got %d", len(errs))
	}
}
