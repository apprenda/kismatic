package rule

import (
	"fmt"
)

// DockerInPath is a rule that ensures the docker executable is in
// the system's path
type DockerInPath struct {
	Meta
}

func (d DockerInPath) Name() string {
	return fmt.Sprintf("Docker Executable In Path")
}

func (d DockerInPath) IsRemoteRule() bool { return false }

func (d DockerInPath) Validate() []error {
	return nil
}
