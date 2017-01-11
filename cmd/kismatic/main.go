package main

import (
	"os"

	"github.com/apprenda/kismatic/pkg/cli"
	"github.com/apprenda/kismatic/pkg/util"
	"time"
	"math/rand"
)

// Set via linker flag
var version string
var buildDate string

func main() {
	rand.Seed(time.Now().UnixNano())
	cmd, err := cli.NewKismaticCommand(version, buildDate, os.Stdin, os.Stdout)
	if err != nil {
		util.PrintColor(os.Stderr, util.Red, "Error initializing command: %v\n", err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		util.PrintColor(os.Stderr, util.Red, "%v\n", err)
		os.Exit(1)
	}
}
