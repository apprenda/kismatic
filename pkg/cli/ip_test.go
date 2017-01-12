package cli

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIPCmdMissingPlan(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists: false,
	}
	opts := &ipOpts{
		planFilename: "planFile",
	}

	_, err := doIP(out, fp, opts)
	if err == nil {
		t.Errorf("ip did not return an error when the plan does not exist")
	}
}

func TestIPCmdEmptyAddress(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists:            true,
		clusterAddress:    "",
		clusterAddressErr: fmt.Errorf("FQDN empty"),
	}
	opts := &ipOpts{
		planFilename: "planFile",
	}

	_, err := doIP(out, fp, opts)
	if err == nil {
		t.Errorf("ip did not return an error when LoadBalancedFQDN is empty")
	}
}

func TestIPCmdValidAddress(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists:         true,
		clusterAddress: "10.0.0.10",
	}
	opts := &ipOpts{
		planFilename: "planFile",
	}

	ip, err := doIP(out, fp, opts)
	if err != nil {
		t.Errorf("ip returned an error %v", err)
	}

	if len(ip) <= 0 {
		t.Errorf("ip value returned is empty")
	}
}
