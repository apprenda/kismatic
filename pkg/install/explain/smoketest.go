package explain

import (
	"bytes"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/util"
)

type SmokeTestEventExplainer struct {
	DefaultEventExplainer *DefaultEventExplainer
	smokeTestStarted      bool
}

// ExplainEvent detects when Kuberang has started execution, and then analyses
// Kuberang's output.
func (explainer *SmokeTestEventExplainer) ExplainEvent(e ansible.Event, verbose bool) string {
	switch event := e.(type) {
	case *ansible.TaskStartEvent:
		// Flip the bool when the kuberang ansible task starts
		if event.Name == "run smoke test checks using Kuberang" {
			explainer.smokeTestStarted = true
		}
		return explainer.DefaultEventExplainer.ExplainEvent(e, verbose)
	case *ansible.RunnerOKEvent:
		// delegate to the default explainer if we are not looking at the smoke test's results
		if !explainer.smokeTestStarted {
			return explainer.DefaultEventExplainer.ExplainEvent(e, verbose)
		}
		buf := &bytes.Buffer{}
		util.PrintOkln(buf)
		buf.WriteString(event.Result.Stdout)
		buf.WriteString("\n")
		return buf.String()
	case *ansible.PlaybookEndEvent:
		// Override so that we don't print the last green [OK]
		return ""
	default:
		return explainer.DefaultEventExplainer.ExplainEvent(e, verbose)
	}
}
