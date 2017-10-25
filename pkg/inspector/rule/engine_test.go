package rule

import (
	"errors"
	"reflect"
	"testing"

	"github.com/apprenda/kismatic/pkg/inspector/check"
)

type fakeCheck struct {
	ok  bool
	err error
}

func (c fakeCheck) Check() (bool, error) {
	return c.ok, c.err
}

type fakeRule struct {
	Meta
	name     string
	isRemote bool
}

func (r fakeRule) Name() string       { return r.name }
func (r fakeRule) IsRemoteRule() bool { return r.isRemote }
func (r fakeRule) Validate() []error  { return nil }

type fakeRuleCheckMapper struct {
	check check.Check
	err   error
}

func (m fakeRuleCheckMapper) GetCheckForRule(Rule) (check.Check, error) {
	return m.check, m.err
}

func TestEngine(t *testing.T) {
	dummyError := errors.New("dummy error...")
	tests := []struct {
		mapper          fakeRuleCheckMapper
		rule            fakeRule
		ruleWhen        [][]string
		facts           []string
		expectedResults []Result
		expectErr       bool
	}{
		// Single rule that passes
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: true},
			},
			rule: fakeRule{
				name: "SuccessRule",
			},
			facts: []string{},
			expectedResults: []Result{
				{
					Name:    "SuccessRule",
					Success: true,
				},
			},
		},
		// Single rule that fails
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			facts: []string{},
			expectedResults: []Result{
				{
					Name:    "FailRule",
					Success: false,
					Error:   dummyError.Error(),
				},
			},
		},
		// Single rule that should run due to facts
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			ruleWhen: [][]string{[]string{"ubuntu"}, []string{"worker"}},
			facts:    []string{"ubuntu", "worker", "otherFact"},
			expectedResults: []Result{
				{
					Name:    "FailRule",
					Success: false,
					Error:   dummyError.Error(),
				},
			},
		},
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			ruleWhen: [][]string{[]string{"ubuntu"}, []string{"master", "worker"}},
			facts:    []string{"ubuntu", "worker", "otherFact"},
			expectedResults: []Result{
				{
					Name:    "FailRule",
					Success: false,
					Error:   dummyError.Error(),
				},
			},
		},
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			ruleWhen: [][]string{[]string{"centos", "rhel"}, []string{"worker"}},
			facts:    []string{"centos", "worker", "otherFact"},
			expectedResults: []Result{
				{
					Name:    "FailRule",
					Success: false,
					Error:   dummyError.Error(),
				},
			},
		},
		// Single rule that should not run due to facts
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			ruleWhen:        [][]string{[]string{"ubuntu"}},
			facts:           []string{"otherFact"},
			expectedResults: []Result{},
		},
		// Single rule that should run regardless of facts
		{
			mapper: fakeRuleCheckMapper{
				check: fakeCheck{ok: false, err: dummyError},
			},
			rule: fakeRule{
				name: "FailRule",
			},
			ruleWhen: [][]string{},
			facts:    []string{"ubuntu"},
			expectedResults: []Result{
				{
					Name:    "FailRule",
					Success: false,
					Error:   dummyError.Error(),
				},
			},
		},
		// Mapper returns an error, engine should return error
		{
			mapper: fakeRuleCheckMapper{
				err: dummyError,
			},
			rule:      fakeRule{},
			expectErr: true,
		},
	}

	for _, test := range tests {
		// Set the rule when conditions here because of embedding :(
		if test.ruleWhen != nil {
			test.rule.When = test.ruleWhen
		}
		e := Engine{
			RuleCheckMapper: test.mapper,
		}
		result, err := e.ExecuteRules([]Rule{test.rule}, test.facts)
		if test.expectErr && err == nil {
			t.Errorf("expected an error, but didn't get one")
			continue
		}
		if err != nil && !test.expectErr {
			t.Errorf("got an unexpected error: %v", err)
			continue
		}

		if !reflect.DeepEqual(test.expectedResults, result) {
			t.Errorf("expected %+v, but got %+v", test.expectedResults, result)
		}
	}
}

type fakeClosableCheck struct {
	success     bool
	closeCalled bool
}

func (c *fakeClosableCheck) Check() (bool, error) { return c.success, nil }

func (c *fakeClosableCheck) Close() error {
	c.closeCalled = true
	return nil
}

func TestEngineClosableCheckSuccess(t *testing.T) {
	fakeCheck := &fakeClosableCheck{success: true}
	mapper := fakeRuleCheckMapper{
		check: fakeCheck,
	}
	e := Engine{
		RuleCheckMapper: mapper,
	}
	rule := fakeRule{}
	_, err := e.ExecuteRules([]Rule{rule}, []string{})
	if err != nil {
		t.Errorf("unexpected error when executing closable check: %v", err)
	}

	if err := e.CloseChecks(); err != nil {
		t.Errorf("unexpected error when closing checks: %v", err)
	}

	if !fakeCheck.closeCalled {
		t.Errorf("The check was not closed")
	}
}

func TestEngineClosableCheckFail(t *testing.T) {
	fakeCheck := &fakeClosableCheck{success: false}
	mapper := fakeRuleCheckMapper{
		check: fakeCheck,
	}
	e := Engine{
		RuleCheckMapper: mapper,
	}
	rule := fakeRule{}
	_, err := e.ExecuteRules([]Rule{rule}, []string{})
	if err != nil {
		t.Errorf("unexpected error when executing closable check: %v", err)
	}

	if err := e.CloseChecks(); err != nil {
		t.Errorf("unexpected error when closing checks: %v", err)
	}

	if fakeCheck.closeCalled {
		t.Errorf("The check failed, and close was called on it")
	}
}
