package rule

import "testing"

func TestTCPPortAvailableRuleValidation(t *testing.T) {
	p := TCPPortAvailable{ProcName: "foo"}
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
	p.Port = -1
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
	p.Port = 901283
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
	p.Port = 1024
	if errs := p.Validate(); len(errs) != 0 {
		t.Errorf("expected 0 error, but got %d", len(errs))
	}
	p.ProcName = ""
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
}

func TestTCPPortAccessibleRuleValidation(t *testing.T) {
	p := TCPPortAccessible{}
	if errs := p.Validate(); len(errs) != 2 {
		t.Errorf("expected 2 error, but got %d", len(errs))
	}
	p.Port = -1
	if errs := p.Validate(); len(errs) != 2 {
		t.Errorf("expected 2 error, but got %d", len(errs))
	}
	p.Port = 901283
	if errs := p.Validate(); len(errs) != 2 {
		t.Errorf("expected 2 error, but got %d", len(errs))
	}
	p.Port = 1024
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
	p.Timeout = "nonDuration"
	if errs := p.Validate(); len(errs) != 1 {
		t.Errorf("expected 1 error, but got %d", len(errs))
	}
	p.Timeout = "3s"
	if errs := p.Validate(); len(errs) != 0 {
		t.Errorf("expected 0 error, but got %d", len(errs))
	}
}
