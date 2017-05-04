package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	tillerImg = "gcr.io/kubernetes-helm/tiller"
)

type HelmClient struct {
	Binary          string
	ClientDirectory string
	TillerImage     string
	ServiceAccount  string
	Kubeconfig      string
	Stdout          io.Writer
	Stderr          io.Writer
}

func DefaultHelmClient() (*HelmClient, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("Could not determine $HOME directory: %v", err)
	}
	helm := HelmClient{
		Binary:          "./helm",
		ClientDirectory: path.Join(home, ".helm"),
		Kubeconfig:      "~/.kube/config",
		TillerImage:     fmt.Sprintf("%s:v%s", tillerImg, HelmVersion),
		ServiceAccount:  "tiller",
	}

	return &helm, nil
}

func (h *HelmClient) BackupDirectory() string {
	return fmt.Sprintf("%s.backup-%s", h.ClientDirectory, time.Now().Format("2006-01-02-15-04-05"))
}

// BackupClient checks for existance of the $HOME/.helm directory and backs it up to $HOME/.helm.backup-$DATETIME
func (h *HelmClient) BackupClient() (bool, error) {
	_, err := os.Stat(h.ClientDirectory)
	var backedup bool
	// Directory exists
	if err == nil {
		if err = os.Rename(h.ClientDirectory, h.BackupDirectory()); err != nil {
			return true, fmt.Errorf("Could not back up %q directory: %v", h.ClientDirectory, err)
		}
		return true, nil
	} else if !os.IsNotExist(err) { // Directory does not exist but got some other error
		return backedup, fmt.Errorf("Could not determine if $HOME/.helm directory exists: %v", err)
	}
	// Directory does not already exist, nothing to do
	return backedup, nil
}

// Init exectues 'helm init' using the local binary
func (h *HelmClient) Init() error {
	cmd := exec.Command(h.Binary, "init", "--service-account", h.ServiceAccount)
	if h.TillerImage != "" {
		cmd.Args = append(cmd.Args, "-i", h.TillerImage)
	}
	cmd.Stdout = h.Stdout
	cmd.Stderr = h.Stderr
	// append KUBECONFIG env var with the generated kubeconfig file path
	env := os.Environ()
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", h.Kubeconfig))
	cmd.Env = env

	return cmd.Run()
}
