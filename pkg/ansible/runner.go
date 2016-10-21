package ansible

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	// RawFormat is the raw Ansible output formatting
	RawFormat = OutputFormat("raw")
	// JSONLinesFormat is a JSON Lines representation of Ansible events
	JSONLinesFormat = OutputFormat("json_lines")
)

// OutputFormat is used for controlling the STDOUT format of the Ansible runner
type OutputFormat string

// Runner for running Ansible playbooks
type Runner interface {
	// StartPlaybook runs the playbook asynchronously with the given inventory and extra vars.
	// It returns a read-only channel that must be consumed for the playbook execution to proceed.
	StartPlaybook(playbookFile string, inventory Inventory, vars ExtraVars) (<-chan Event, error)
	// WaitPlaybook blocks until the execution of the playbook is complete. If an error occurred,
	// it is returned. Otherwise, returns nil to signal the completion of the playbook.
	WaitPlaybook() error
	// StartPlaybookOnNode runs the playbook asynchronously with the given inventory and extra vars
	// against the specific node.
	// It returns a read-only channel that must be consumed for the playbook execution to proceed.
	StartPlaybookOnNode(playbookFile string, inventory Inventory, vars ExtraVars, node string) (<-chan Event, error)
}

type runner struct {
	// Out is the stdout writer for the Ansible process
	out io.Writer
	// ErrOut is the stderr writer for the Ansible process
	errOut io.Writer

	pythonPath   string
	ansibleDir   string
	waitPlaybook func() error
	namedPipe    string
}

// ExtraVars is a map of variables that are used when executing a playbook
type ExtraVars map[string]string

func (v ExtraVars) commandLineVars() (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("error marshaling ansible vars")
	}
	return string(b), nil
}

// NewRunner returns a new runner for running Ansible playbooks.
func NewRunner(out, errOut io.Writer, ansibleDir string) (Runner, error) {
	ppath, err := getPythonPath()
	if err != nil {
		return nil, err
	}

	return &runner{
		out:        out,
		errOut:     errOut,
		pythonPath: ppath,
		ansibleDir: ansibleDir,
	}, nil
}

// WaitPlaybook blocks until the ansible process running the playbook exits.
// If the process exits with a non-zero status, it will return an error.
func (r *runner) WaitPlaybook() error {
	if r.waitPlaybook == nil {
		return fmt.Errorf("wait called, but playbook not started")
	}
	execErr := r.waitPlaybook()
	// Process exited, we can clean up named pipe
	removeErr := os.Remove(r.namedPipe)
	if removeErr != nil && execErr != nil {
		return fmt.Errorf("an error occurred running ansible: %v. Removing named pipe at %q failed: %v", execErr, r.namedPipe, removeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("failed to clean up named pipe at %q: %v", r.namedPipe, removeErr)
	}
	if execErr != nil {
		return fmt.Errorf("error running ansible: %v", execErr)
	}
	return nil
}

// RunPlaybook with the given inventory and extra vars
func (r *runner) StartPlaybook(playbookFile string, inv Inventory, vars ExtraVars) (<-chan Event, error) {
	return r.startPlaybook(playbookFile, inv, vars, "") // Don't set the --limit arg
}

// StartPlaybookOnNode runs the playbook asynchronously with the given inventory and extra vars
// against the specific node.
// It returns a read-only channel that must be consumed for the playbook execution to proceed.
func (r *runner) StartPlaybookOnNode(playbookFile string, inv Inventory, vars ExtraVars, node string) (<-chan Event, error) {
	limitArg := node // set the --limit arg to the node we want to target
	return r.startPlaybook(playbookFile, inv, vars, limitArg)
}

func (r *runner) startPlaybook(playbookFile string, inv Inventory, vars ExtraVars, limitArg string) (<-chan Event, error) {
	extraVars, err := vars.commandLineVars()
	if err != nil {
		return nil, fmt.Errorf("error building extra vars: %v", err)
	}

	inventoryFile := filepath.Join(r.ansibleDir, "inventory.ini")
	if err = ioutil.WriteFile(inventoryFile, inv.ToINI(), 0644); err != nil {
		return nil, fmt.Errorf("error writing inventory file to %q: %v", inventoryFile, err)
	}

	playbook := filepath.Join(r.ansibleDir, "playbooks", playbookFile)
	cmd := exec.Command(filepath.Join(r.ansibleDir, "bin", "ansible-playbook"), "-i", inventoryFile, "-s", playbook, "--extra-vars", extraVars)
	cmd.Stdout = r.out
	cmd.Stderr = r.errOut

	if limitArg != "" {
		cmd.Args = append(cmd.Args, "--limit", limitArg)
	}

	os.Setenv("PYTHONPATH", r.pythonPath)
	os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")
	os.Setenv("ANSIBLE_CALLBACK_PLUGINS", filepath.Join(r.ansibleDir, "playbooks", "callback"))
	os.Setenv("ANSIBLE_CALLBACK_WHITELIST", "json_lines")

	// We always want the most verbose output from Ansible. If it's not going to
	// stdout, it's going to a log file.
	cmd.Args = append(cmd.Args, "-vvvv")

	// Create named pipe for getting JSON lines event stream
	start := time.Now()
	r.namedPipe = filepath.Join(os.TempDir(), fmt.Sprintf("ansible-pipe-%s", start.Format("2006-01-02-15-04-05.99999")))
	if err = syscall.Mkfifo(r.namedPipe, 0644); err != nil {
		return nil, fmt.Errorf("error creating named pipe %q: %v", r.namedPipe, err)
	}
	os.Setenv("ANSIBLE_JSON_LINES_PIPE", r.namedPipe)

	// Starts async execution of ansible, which will block until
	// we start reading from the named pipe
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("error running playbook: %v", err)
	}
	r.waitPlaybook = cmd.Wait

	// Create the event stream out of the named pipe
	eventStreamFile, err := os.OpenFile(r.namedPipe, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("error openning event stream pipe: %v", err)
	}
	eventStream := EventStream(eventStreamFile)
	return eventStream, nil
}

func getPythonPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working dir: %v", err)
	}
	lib := filepath.Join(wd, "ansible", "lib", "python2.7", "site-packages")
	lib64 := filepath.Join(wd, "ansible", "lib64", "python2.7", "site-packages")
	return fmt.Sprintf("%s:%s", lib, lib64), nil
}
