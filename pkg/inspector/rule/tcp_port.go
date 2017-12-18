package rule

import (
	"errors"
	"fmt"
	"time"
)

// TCPPortAvailable is a rule that ensures that a given port is available
// on the node. The port is considered available if:
// - The port is free and ready to be bound by a new process, or
// - The port is bound to the process defined in ProcName
type TCPPortAvailable struct {
	Meta
	// The port number to verify
	Port int
	// The name of the process that owns this port after KET installation
	ProcName string
}

// Name is the name of the rule
func (p TCPPortAvailable) Name() string {
	return fmt.Sprintf("Port Available: %d", p.Port)
}

// IsRemoteRule returns true if the rule is to be run from outside the node
func (p TCPPortAvailable) IsRemoteRule() bool { return false }

// Validate the rule
func (p TCPPortAvailable) Validate() []error {
	var errs []error
	if p.Port < 1 || p.Port > 65535 {
		errs = append(errs, fmt.Errorf("Invalid port number %d specified", p.Port))
	}
	if p.ProcName == "" {
		errs = append(errs, fmt.Errorf("ProcName cannot be empty"))
	}
	return errs
}

// TCPPortAccessible is a rule that ensures the given port on a remote node
// is accessible from the network
type TCPPortAccessible struct {
	Meta
	Port    int
	Timeout string
}

// Name returns the name of the rule
func (p TCPPortAccessible) Name() string {
	return fmt.Sprintf("Port Accessible: %d", p.Port)
}

// IsRemoteRule returns true if the rule is to be run from a remote node
func (p TCPPortAccessible) IsRemoteRule() bool { return true }

// Validate the rule
func (p TCPPortAccessible) Validate() []error {
	errs := []error{}
	if p.Port < 1 || p.Port > 65535 {
		errs = append(errs, fmt.Errorf("Invalid port number %d specified", p.Port))
	}
	if p.Timeout == "" {
		errs = append(errs, errors.New("Timeout cannot be empty"))
	}
	if p.Timeout != "" {
		if _, err := time.ParseDuration(p.Timeout); err != nil {
			errs = append(errs, fmt.Errorf("Invalid duration provided %q", p.Timeout))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
