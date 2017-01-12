package cli

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDashboardCmdMissingPlan(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists: false,
	}
	opts := &dashboardOpts{
		planFilename:     "planFile",
		dashboardURLMode: true,
	}

	_, err := doDashboard(out, fp, opts)
	if err == nil {
		t.Errorf("dashboard did not return an error when the plan does not exist")
	}
}

func TestDashboardCmdEmptyAddress(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists:          true,
		dashboardURL:    "",
		dashboardURLErr: fmt.Errorf("FQDN empty"),
	}
	opts := &dashboardOpts{
		planFilename:     "planFile",
		dashboardURLMode: true,
	}

	_, err := doDashboard(out, fp, opts)
	if err == nil {
		t.Errorf("dashboard did not return an error when LoadBalancedFQDN is empty")
	}
}

func TestDashboardCmdTimeoutAddress(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists:       true,
		dashboardURL: "http://httpbin.org/delay/5",
	}
	opts := &dashboardOpts{
		planFilename:     "planFile",
		dashboardURLMode: true,
	}

	_, err := doDashboard(out, fp, opts)
	if err == nil {
		t.Errorf("ip returned an error %v", err)
	}
}

func TestDashboardCmdValidAddress(t *testing.T) {
	out := &bytes.Buffer{}
	fp := &fakePlanner{
		exists:       true,
		dashboardURL: "http://httpbin.org/delay/1",
	}
	opts := &dashboardOpts{
		planFilename:     "planFile",
		dashboardURLMode: true,
	}

	url, err := doDashboard(out, fp, opts)
	if err != nil {
		t.Errorf("dashboard returned an error %v", err)
	}

	if len(url) <= 0 {
		t.Errorf("dashboard url value returned is empty")
	}
}
