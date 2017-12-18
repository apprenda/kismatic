package check

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// TCPPortClientCheck verifies that a given port on a remote node
// is accessible through the network
type TCPPortClientCheck struct {
	// IPAddress is the IP of the remote node
	IPAddress string
	// PortNumber is the target service port
	PortNumber int
	// Timeout is the maximum amount of time the check will
	// wait when connecting to the server before bailing out
	Timeout time.Duration
}

// Check returns true if the TCP connection is established and the server
// returns the expected response. Otherwise, returns false and an error message
func (c *TCPPortClientCheck) Check() (bool, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.IPAddress, c.PortNumber), timeout)
	if err != nil {
		return false, fmt.Errorf("Port %d on host %q is unreachable. Error was: %v", c.PortNumber, c.IPAddress, err)
	}
	conn.Close()
	return true, nil
}

// TCPPortServerCheck ensures that the given port is free, or bound to the right
// process. In the case that it is free, it stands up a TCP server that can be
// used to check TCP connectivity to the host using TCPPortClientCheck
type TCPPortServerCheck struct {
	PortNumber     int
	ProcName       string
	started        bool
	closeListener  func() error
	listenerClosed chan interface{}
}

// Check returns true if the port is free, or taken by the expected process.
func (c *TCPPortServerCheck) Check() (bool, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", c.PortNumber))
	if err != nil && strings.Contains(err.Error(), "address already in use") {
		return portTakenByProc(c.PortNumber, c.ProcName)
	}
	if err != nil {
		return false, fmt.Errorf("error listening on port %d: %v", c.PortNumber, err)
	}
	c.closeListener = ln.Close
	// Setup go routine for accepting connections
	c.listenerClosed = make(chan interface{})
	go func(closed <-chan interface{}) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-closed:
					// don't log the error, as we have closed the server and the error
					// is related to that.
					return
				default:
					log.Println(fmt.Sprintf("error occurred accepting request: %v", err))
					continue
				}
			}
			// Setup go routine that behaves as an echo server
			go func(c net.Conn) {
				io.Copy(c, c)
				c.Close()
			}(conn)
		}
	}(c.listenerClosed)
	c.started = true
	return true, nil
}

// Close the TCP server if it was started. Otherwise this is a noop.
func (c *TCPPortServerCheck) Close() error {
	if c.started {
		close(c.listenerClosed)
		return c.closeListener()
	}
	return nil
}

// Returns true if the port is taken by a process with the given name.
func portTakenByProc(port int, procName string) (bool, error) {
	// Use `ss` (sockstat) to find the process listening on the given port.
	// Sample output:
	// ~# ss -tpln state listening src :6443  | strings
	// Recv-Q Send-Q Local Address:Port               Peer Address:Port
	// 0      128              :::6443                         :::*                   users:(("kube-apiserver",pid=21199,fd=59))
	//
	// ~# ss -tpln state listening src :80  | strings
	// Recv-Q Send-Q Local Address:Port               Peer Address:Port
	// 0      128               *:80                            *:*                   users:(("nginx",pid=30729,fd=10),("nginx",pid=30728,fd=10),("nginx",pid=30721,fd=10))
	//
	cmd := exec.Command("ss", "-tpln", "state", "listening", "src", fmt.Sprintf(":%d", port))
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error running ss: %v", err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return false, fmt.Errorf("expected ss to return at least 2 lines, but returned %d", len(lines))
	}
	boundProc, err := getProcNameFromTCPSockStatLine(lines[1])
	if err != nil {
		return false, err
	}
	return boundProc == procName, nil
}

// given an entry returned by sockstat (ss), return the name of the process using the port.
// assumes ss was run with flags: -tpln
func getProcNameFromTCPSockStatLine(line string) (string, error) {
	ssFields := strings.Fields(line)
	// The fifth field includes information about the process using the port
	if len(ssFields) != 5 {
		return "", fmt.Errorf("unexpected output returned from ss command. output was: %s", line)
	}
	// users:(("nginx",pid=30729,fd=10),("nginx",pid=30728,fd=10),("nginx",pid=30721,fd=10))
	usersField := ssFields[4]

	// This regular expression contains a single capturing group that will fish
	// out the process name from the `ss` output. In the case that `ss` returns
	// a list with multiple users, the left-most user will be matched.
	re := regexp.MustCompile(`^users:\(\("([^"]+)",pid=\d+,fd=\d+\)`)
	matched := re.FindSubmatch([]byte(usersField))
	if len(matched) < 2 {
		return "", fmt.Errorf("unable to determine the process from ss line %q", usersField)
	}
	// We are interested in the subexpression (capturing group). The first item
	// in the matched list is the match of the full regexp, not the capturing
	// group.
	return string(matched[1]), nil
}
