package main

import (
	"os"

	"github.com/apprenda/kismatic/pkg/cli"
	"github.com/spf13/cobra/doc"
)

var version string
var buildDate string

func main() {
	cmd, _ := cli.NewKismaticCommand(version, buildDate, os.Stdin, os.Stdout, os.Stderr)
	doc.GenMarkdownTree(cmd, "./docs/kismatic-cli")
}
