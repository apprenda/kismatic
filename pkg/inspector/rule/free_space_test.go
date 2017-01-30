package rule

import "testing"

func TestFreeSpaceRuleValidation(t *testing.T) {
	f := FreeSpace{}
	if errs := f.Validate(); len(errs) != 2 {
		t.Errorf("expected 2 error, but got %d", len(errs))
	}

	f.Path = "foo"
	if errs := f.Validate(); len(errs) != 2 {
		t.Errorf("expected 2 error, but got %d", len(errs))
	}

	f.Path = "/arglebargle"
	if errs := f.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}

	f.MinimumBytes = "A stalk of corn"
	if errs := f.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}

	f.MinimumBytes = "-1"
	if errs := f.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}

	f.MinimumBytes = "909"
	if errs := f.Validate(); len(errs) != 0 {
		t.Errorf("expected 0 error, but got %d", len(errs))
	}
}
