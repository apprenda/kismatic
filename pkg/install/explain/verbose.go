package explain

import (
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/ansible"
	"github.com/apprenda/kismatic/pkg/util"
)

type verboseExplainer struct {
	out              io.Writer
	printPlayMessage bool
	printPlayStatus  bool
	lastPlay         string
	currentTask      string
}

func (explainer *verboseExplainer) writePlayStatus(buf io.Writer) {
	// Do not print message before first play
	if explainer.printPlayMessage {
		// No tasks were printed, no nodes match the selector
		// This is OK and a valid scenario
		if explainer.printPlayStatus {
			fmt.Fprintln(buf)
			util.PrintColor(buf, util.Green, "%s Finished With No Tasks\n", explainer.lastPlay)
		} else {
			util.PrintColor(buf, util.Green, "%s Finished\n", explainer.lastPlay)
		}
	}
}

// ExplainEvent writes the verbose explanation of the ansible event
func (explainer *verboseExplainer) ExplainEvent(e ansible.Event) {
	out := explainer.out
	switch event := e.(type) {
	case *ansible.PlayStartEvent:
		// On a play start the previous play ends
		// Print a success status, but only when there were no errors
		explainer.writePlayStatus(out)
		fmt.Fprintf(out, "%s", event.Name)
		// Set default state for the play
		explainer.lastPlay = event.Name
		explainer.printPlayStatus = true
		explainer.printPlayMessage = true
	case *ansible.RunnerFailedEvent:
		// Print newline before first task status
		if explainer.printPlayStatus {
			fmt.Fprintln(out)
			// Dont print play success status on error
			explainer.printPlayStatus = false
		}
		// Tasks only print at verbose level, on ERROR also print task name
		if event.IgnoreErrors {
			util.PrettyPrintErrorIgnored(out, "  %s", event.Host)
		} else {
			util.PrettyPrintErr(out, "  %s %s", event.Host, event.Result.Message)
		}
		if event.Result.Stdout != "" {
			util.PrintColor(out, util.Red, "---- STDOUT ----\n%s\n", event.Result.Stdout)
		}
		if event.Result.Stderr != "" {
			util.PrintColor(out, util.Red, "---- STDERR ----\n%s\n", event.Result.Stderr)
		}
		if event.Result.Stderr != "" || event.Result.Stdout != "" {
			util.PrintColor(out, util.Red, "---------------\n")
		}
	case *ansible.RunnerUnreachableEvent:
		// Host is unreachable
		// Print newline before first task
		if explainer.printPlayStatus {
			fmt.Fprintln(out)
			// Dont print play success status on error
			explainer.printPlayStatus = false
		}
		util.PrettyPrintUnreachable(out, "  %s", event.Host)
	case *ansible.TaskStartEvent:
		// Print newline before first task status
		if explainer.printPlayStatus {
			fmt.Fprintln(out)
			// Dont print play success status on error
			explainer.printPlayStatus = false
		}
		fmt.Fprintf(out, "- Running task: %s\n", event.Name)
		// Set current task name
		explainer.currentTask = event.Name
	case *ansible.HandlerTaskStartEvent:
		// Print newline before first task
		if explainer.printPlayStatus {
			fmt.Fprintln(out)
			// Dont print play success status on error
			explainer.printPlayStatus = false
		}
		fmt.Fprintf(out, "- Running task: %s\n", event.Name)
		// Set current task name
		explainer.currentTask = event.Name
	case *ansible.PlaybookEndEvent:
		// Playbook ends, print the last play status
		explainer.writePlayStatus(out)
	case *ansible.RunnerSkippedEvent:
		util.PrettyPrintSkipped(out, "  %s", event.Host)
	case *ansible.RunnerOKEvent:
		util.PrettyPrintOk(out, "  %s", event.Host)
	case *ansible.RunnerItemOKEvent:
		msg := fmt.Sprintf("  %s", event.Host)
		if event.Result.Item != "" {
			msg = msg + fmt.Sprintf(" with %q", event.Result.Item)
		}
		util.PrettyPrintOk(out, msg)
	case *ansible.RunnerItemFailedEvent:
		msg := fmt.Sprintf("  %s", event.Host)
		if event.Result.Item != "" {
			msg = msg + fmt.Sprintf(" with %q", event.Result.Item)
		}
		// Print newline before first task status
		if explainer.printPlayStatus {
			fmt.Fprintln(out)
			// Dont print play success status on error
			explainer.printPlayStatus = false
		}
		// Tasks only print at verbose level, on ERROR also print task name
		if event.IgnoreErrors {
			util.PrettyPrintErrorIgnored(out, msg)
		} else {
			util.PrettyPrintErr(out, "  %s %s", msg, event.Result.Message)
		}
		if event.Result.Stdout != "" {
			util.PrintColor(out, util.Red, "---- STDOUT ----\n%s\n", event.Result.Stdout)
		}
		if event.Result.Stderr != "" {
			util.PrintColor(out, util.Red, "---- STDERR ----\n%s\n", event.Result.Stderr)
		}
		if event.Result.Stderr != "" || event.Result.Stdout != "" {
			util.PrintColor(out, util.Red, "---------------\n")
		}

	// Ignored events
	case *ansible.RunnerItemRetryEvent:
		fmt.Fprintf(out, " %s Retrying: %s (%d/%d attempts)\n", event.Host, explainer.currentTask, event.Result.Attempts, event.Result.MaxRetries)
	case *ansible.PlaybookStartEvent:
	default:
		util.PrintColor(out, util.Orange, "Unhandled event: %T\n", event)
	}
}
