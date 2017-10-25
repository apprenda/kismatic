package rule

import (
	"sync"

	"github.com/apprenda/kismatic/pkg/inspector/check"
)

// The Engine executes rules and reports the results
type Engine struct {
	RuleCheckMapper CheckMapper
	mu              sync.Mutex
	closableChecks  []check.ClosableCheck
}

// ExecuteRules runs the rules that should be executed according to the facts,
// and returns a collection of results. The number of results is not guaranteed
// to equal the number of rules.
func (e *Engine) ExecuteRules(rules []Rule, facts []string) ([]Result, error) {
	results := []Result{}
	for _, rule := range rules {
		if !shouldExecuteRule(rule, facts) {
			continue
		}

		// Map the rule to a check
		c, err := e.RuleCheckMapper.GetCheckForRule(rule)
		if err != nil {
			return nil, err
		}

		// Run the check and report result
		ok, err := c.Check()
		res := Result{
			Name:        rule.Name(),
			Success:     ok,
			Remediation: "",
		}
		if err != nil {
			res.Error = err.Error()
		}

		// We update the closables as we go to avoid leaking closables
		// in the event where we have to return an error from within the loop.
		if closeable, ok := c.(check.ClosableCheck); ok && res.Success {
			e.mu.Lock()
			e.closableChecks = append(e.closableChecks, closeable)
			e.mu.Unlock()
		}

		results = append(results, res)
	}
	return results, nil
}

// CloseChecks that need to be closed
func (e *Engine) CloseChecks() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, c := range e.closableChecks {
		if err := c.Close(); err != nil {
			// TODO: Figure out what to do with the error here
		}
	}
	e.closableChecks = []check.ClosableCheck{}
	return nil
}

func shouldExecuteRule(rule Rule, facts []string) bool {
	// Run if and only if the all the conditions on the rule are
	// satisfied by the facts
	for _, whenSlice := range rule.GetRuleMeta().When {
		found := false
		for _, whenCondition := range whenSlice {
			for _, l := range facts {
				if whenCondition == l {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
