package ansible

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	StartPlaybook(playbookFile string, inventory Inventory, cc ClusterCatalog) (<-chan Event, error)
	// WaitPlaybook blocks until the execution of the playbook is complete. If an error occurred,
	// it is returned. Otherwise, returns nil to signal the completion of the playbook.
	WaitPlaybook() error
	// StartPlaybookOnNode runs the playbook asynchronously with the given inventory and extra vars
	// against the specific node.
	// It returns a read-only channel that must be consumed for the playbook execution to proceed.
	StartPlaybookOnNode(playbookFile string, inventory Inventory, cc ClusterCatalog, node ...string) (<-chan Event, error)
}

type runner struct {
	// Out is the stdout writer for the Ansible process
	out io.Writer
	// ErrOut is the stderr writer for the Ansible process
	errOut io.Writer

	pythonPath   string
	ansibleDir   string
	runDir       string
	waitPlaybook func() error
	namedPipe    string
}

// NewRunner returns a new runner for running Ansible playbooks.
func NewRunner(out, errOut io.Writer, ansibleDir string, runDir string) (Runner, error) {
	// Ansible depends on python 2.7 being installed and on the path as "python".
	// Validate that it is available
	if _, err := exec.LookPath("python"); err != nil {
		return nil, fmt.Errorf("Could not find 'python' in the PATH. Ensure that python 2.7 is installed and in the path as 'python'.")
	}

	ppath, err := getPythonPath()
	if err != nil {
		return nil, err
	}

	return &runner{
		out:        out,
		errOut:     errOut,
		pythonPath: ppath,
		ansibleDir: ansibleDir,
		runDir:     runDir,
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
func (r *runner) StartPlaybook(playbookFile string, inv Inventory, cc ClusterCatalog) (<-chan Event, error) {
	return r.startPlaybook(playbookFile, inv, cc) // Don't set the --limit arg
}

// StartPlaybookOnNode runs the playbook asynchronously with the given inventory and extra vars
// against the specific node.
// It returns a read-only channel that must be consumed for the playbook execution to proceed.
func (r *runner) StartPlaybookOnNode(playbookFile string, inv Inventory, cc ClusterCatalog, nodes ...string) (<-chan Event, error) {
	// set the --limit arg to the node we want to target
	return r.startPlaybook(playbookFile, inv, cc, nodes...)
}

func (r *runner) startPlaybook(playbookFile string, inv Inventory, cc ClusterCatalog, nodes ...string) (<-chan Event, error) {
	playbook := filepath.Join(r.ansibleDir, "playbooks", playbookFile)
	if _, err := os.Stat(playbook); os.IsNotExist(err) {
		return nil, fmt.Errorf("playbook %q does not exist", playbook)
	}

	yamlBytes, err := cc.ToYAML()
	if err != nil {
		return nil, fmt.Errorf("error writing cluster catalog data to yaml: %v", err)
	}
	clusterCatalogFile := filepath.Join(r.ansibleDir, "clustercatalog.yaml")
	if err = ioutil.WriteFile(clusterCatalogFile, yamlBytes, 0644); err != nil {
		return nil, fmt.Errorf("error writing cluster catalog file to %q: %v", clusterCatalogFile, err)
	}

	inventoryFile := filepath.Join(r.ansibleDir, "inventory.ini")
	if err := ioutil.WriteFile(inventoryFile, inv.ToINI(), 0644); err != nil {
		return nil, fmt.Errorf("error writing inventory file to %q: %v", inventoryFile, err)
	}

	if err := copyFileContents(clusterCatalogFile, filepath.Join(r.runDir, "clustercatalog.yaml")); err != nil {
		return nil, fmt.Errorf("error copying clustercatalog.yaml to %q: %v", r.runDir, err)
	}
	if err := copyFileContents(inventoryFile, filepath.Join(r.runDir, "inventory.ini")); err != nil {
		return nil, fmt.Errorf("error copying inventory.ini to %q: %v", r.runDir, err)
	}

	cmd := exec.Command(filepath.Join(r.ansibleDir, "bin", "ansible-playbook"), "-i", inventoryFile, "-s", playbook, "--extra-vars", "@"+clusterCatalogFile)
	cmd.Stdout = r.out
	cmd.Stderr = r.errOut

	log.SetOutput(r.out)

	limitArg := strings.Join(nodes, ",")
	if limitArg != "" {
		cmd.Args = append(cmd.Args, "--limit", limitArg)
	}

	// We always want the most verbose output from Ansible. If it's not going to
	// stdout, it's going to a log file.
	cmd.Args = append(cmd.Args, "-vvvv")

	// Create named pipe
	np, err := createTempNamedPipe()
	if err != nil {
		return nil, err
	}
	r.namedPipe = np

	os.Setenv("PYTHONPATH", r.pythonPath)
	os.Setenv("ANSIBLE_CALLBACK_PLUGINS", filepath.Join(r.ansibleDir, "playbooks", "callback"))
	os.Setenv("ANSIBLE_CALLBACK_WHITELIST", "json_lines")
	os.Setenv("ANSIBLE_CONFIG", filepath.Join(r.ansibleDir, "playbooks", "ansible.cfg"))
	os.Setenv("ANSIBLE_JSON_LINES_PIPE", r.namedPipe)

	// Print Ansible command
	fmt.Fprintf(r.out, "export PYTHONPATH=%v\n", os.Getenv("PYTHONPATH"))
	fmt.Fprintf(r.out, "export ANSIBLE_CALLBACK_PLUGINS=%v\n", os.Getenv("ANSIBLE_CALLBACK_PLUGINS"))
	fmt.Fprintf(r.out, "export ANSIBLE_CALLBACK_WHITELIST=%v\n", os.Getenv("ANSIBLE_CALLBACK_WHITELIST"))
	fmt.Fprintf(r.out, "export ANSIBLE_CONFIG=%v\n", os.Getenv("ANSIBLE_CONFIG"))
	fmt.Fprintf(r.out, "export ANSIBLE_JSON_LINES_PIPE=%v\n", os.Getenv("ANSIBLE_JSON_LINES_PIPE"))
	fmt.Fprintln(r.out, strings.Join(cmd.Args, " "))

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

// create a named pipe for getting json events out of ansible.
// add random int to file name to avoid collision.
func createTempNamedPipe() (string, error) {
	start := time.Now()
	np := filepath.Join(os.TempDir(), fmt.Sprintf("ansible-pipe-%d-%s", rand.Int(), start.Format("2006-01-02-15-04-05.99999")))
	if err := syscall.Mkfifo(np, 0644); err != nil {
		return "", fmt.Errorf("error creating named pipe %q: %v", np, err)
	}
	return np, nil
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

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
