package util

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/fatih/color"
)

const (
	noType          = ""
	okType          = "[OK]"
	errType         = "[ERROR]"
	skippedType     = "[SKIPPED]"
	warnType        = "[WARNING]"
	unreachableType = "[UNREACHABLE]"
	errIgnoredType  = "[ERROR IGNORED]"
)

var Green = color.New(color.FgGreen)
var Red = color.New(color.FgRed)
var Orange = color.New(color.FgRed, color.FgYellow)
var Blue = color.New(color.FgCyan)
var White = color.New(color.FgHiWhite)

// PrettyPrintOk [OK](Green) with formatted string
func PrettyPrintOk(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, okType, a...)
}

// PrettyPrintErr [ERROR](Red) with formatted string
func PrettyPrintErr(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, errType, a...)
}

// PrettyPrint no type will be displayed, used for just single line printing
func PrettyPrint(out io.Writer, msg string, a ...interface{}) {
	print(out, msg+"\t", noType, a...)
}

// PrettyPrintWarn [WARNING](Orange) with formatted string
func PrettyPrintWarn(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, warnType, a...)
}

// PrettyPrintErrorIgnored [ERROR IGNORED](Red) with formatted string
func PrettyPrintErrorIgnored(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, errIgnoredType, a...)
}

// PrettyPrintUnreachable [UNREACHABLE](Red) with formatted string
func PrettyPrintUnreachable(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, unreachableType, a...)
}

// PrettyPrintSkipped [SKIPPED](blue) with formatted string
func PrettyPrintSkipped(out io.Writer, msg string, a ...interface{}) {
	print(out, msg, skippedType, a...)
}

// PrintOk print whole message in green(Red) format
func PrintOk(out io.Writer) {
	PrintColor(out, Green, okType)
}

// PrintOkln print whole message in green(Red) format
func PrintOkln(out io.Writer) {
	PrintColor(out, Green, okType+"\n")
}

// PrintError print whole message in error(Red) format
func PrintError(out io.Writer) {
	PrintColor(out, Red, errType)
}

// PrintWarn print whole message in warn(Orange) format
func PrintWarn(out io.Writer) {
	PrintColor(out, Orange, warnType)
}

// PrintSkipped print whole message in green(Red) format
func PrintSkipped(out io.Writer) {
	PrintColor(out, Blue, skippedType)
}

// PrintHeader will print header with predifined width
func PrintHeader(out io.Writer, msg string, padding byte) {
	w := tabwriter.NewWriter(out, 84, 0, 0, padding, 0)
	fmt.Fprintln(w, "")
	format := msg + "\t\n"
	fmt.Fprintf(w, format)
	w.Flush()
}

func PrintTable(out io.Writer, msgMap map[string][]string) {
	w := tabwriter.NewWriter(out, 84, 0, 0, ' ', 0)
	for k, v := range msgMap {
		format := fmt.Sprintf("- %s %s", k, v) + "\t\n"
		fmt.Fprintf(w, format)
	}
	w.Flush()
}

// PrintColor prints text in color
func PrintColor(out io.Writer, clr *color.Color, msg string, a ...interface{}) {
	// Remove any newline, results in only one \n
	line := fmt.Sprintf("%s", clr.SprintfFunc()(msg, a...))
	fmt.Fprint(out, line)
}

func print(out io.Writer, msg, status string, a ...interface{}) {
	w := tabwriter.NewWriter(out, 80, 0, 0, ' ', 0)
	// print message
	format := msg + "\t"
	fmt.Fprintf(w, format, a...)

	// print status
	if status != noType {
		// get correct color
		var clr *color.Color
		switch status {
		case okType:
			clr = Green
		case errType, unreachableType:
			clr = Red
		case warnType, errIgnoredType:
			clr = Orange
		case skippedType:
			clr = Blue
		}

		sformat := "%s\n"
		fmt.Fprintf(w, sformat, clr.SprintFunc()(status))

	}
	w.Flush()
}

// PrintValidationErrors loops through the errors
func PrintValidationErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		PrintColor(out, Red, "- %v\n", err)
	}
}
