package ansible

import (
	"io/ioutil"
	"testing"
)

func TestWaitPlaybook(t *testing.T) {
	r, err := NewRunner(ioutil.Discard, ioutil.Discard, "", "/tmp")
	if err != nil {
		t.Fatalf("Error creating runner: %v", err)
	}
	err = r.WaitPlaybook()
	if err.Error() != "wait called, but playbook not started" {
		t.Error("Did not get the expected error when calling WaitPlaybook")
	}
}
