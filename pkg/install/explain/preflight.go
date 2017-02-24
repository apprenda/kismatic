package explain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/inspector/rule"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/gosuri/uilive"
)

// PreflightExplainer is an explainer to be used when running preflight checks.
func PreflightExplainer(verbose bool, out io.Writer) AnsibleEventExplainer {
	if verbose || !isTerminal(out) {
		return &verbosePreflightExplainer{
			out:       out,
			explainer: verboseExplainer{out: out},
		}
	}
	w := uilive.New()
	w.Out = out
	w.Start()
	return &updatingPreflightExplainer{
		out:       w,
		explainer: updatingExplainer{out: w},
	}
}

type updatingPreflightExplainer struct {
	out       *uilive.Writer
	explainer updatingExplainer
}

func (exp *updatingPreflightExplainer) ExplainEvent(ansibleEvent ansible.Event) {
	switch event := ansibleEvent.(type) {
	default:
		exp.explainer.ExplainEvent(ansibleEvent)
	case *ansible.RunnerFailedEvent:
		buf := &bytes.Buffer{}
		// only print this header this is the first failure
		if !exp.explainer.failureOccurred {
			util.PrettyPrintErr(buf, "%s", exp.explainer.currentPlayName)
			fmt.Fprintln(buf, "- Task: "+exp.explainer.currentTask)
		}
		results := []rule.Result{}
		if err := json.Unmarshal([]byte(event.Result.Stdout), &results); err != nil {
			exp.explainer.ExplainEvent(event)
			return
		}
		// print info about pre-flight checks that failed
		util.PrintColor(buf, util.Red, "=> The following checks failed on %q:\n", event.Host)
		for _, r := range results {
			if !r.Success && r.Error != "" {
				util.PrintColor(buf, util.Red, "   - %s: %v\n", r.Name, r.Error)
			} else if !r.Success {
				util.PrintColor(buf, util.Red, "   - %s\n", r.Name)
			}
		}
		fmt.Fprintf(exp.out.Bypass(), buf.String())
		exp.explainer.failureOccurred = true
	}
}

type verbosePreflightExplainer struct {
	out       io.Writer
	explainer verboseExplainer
}

func (exp *verbosePreflightExplainer) ExplainEvent(ansibleEvent ansible.Event) {
	switch event := ansibleEvent.(type) {
	default:
		exp.explainer.ExplainEvent(ansibleEvent)
	case *ansible.RunnerFailedEvent:
		results := []rule.Result{}
		if err := json.Unmarshal([]byte(event.Result.Stdout), &results); err != nil {
			exp.explainer.ExplainEvent(event)
			return
		}
		// print info about pre-flight checks that failed
		util.PrintColor(exp.out, util.Red, "=> The following checks failed on %q:\n", event.Host)
		for _, r := range results {
			if !r.Success && r.Error != "" {
				util.PrintColor(exp.out, util.Red, "   - %s: %v\n", r.Name, r.Error)
			} else if !r.Success {
				util.PrintColor(exp.out, util.Red, "   - %s\n", r.Name)
			}
		}
		util.PrintColor(exp.out, util.Green, "=> Successful pre-flight checks:\n")
		for _, r := range results {
			if r.Success {
				util.PrintColor(exp.out, util.Green, "   - %s\n", r.Name)
			}
		}
		exp.explainer.printPlayStatus = false
	}
}
