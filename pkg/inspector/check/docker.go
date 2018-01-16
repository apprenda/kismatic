package check

import (
	"fmt"
	"os/exec"
)

// DockerInPathCheck returns true if the docker binary is on the executable path
// The check only runs when InstallationDisabled is true and the binary should already be insalled
type DockerInPathCheck struct {
	InstallationDisabled bool
}

func (c DockerInPathCheck) Check() (bool, error) {
	if c.InstallationDisabled {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("command -v %s", "docker"))
		if err := cmd.Run(); err != nil {
			return false, nil
		}
	}
	return true, nil
}
