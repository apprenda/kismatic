package check

import (
	"fmt"
	"syscall"
)

// FreeSpaceCheck checks the available disk space on a path
type FreeSpaceCheck struct {
	MinimumBytes uint64
	Path         string
}

// Check returns true if file contents match the regular expression. Otherwise,
// returns false. If an error occurrs, returns false and the error.
func (c FreeSpaceCheck) Check() (bool, error) {
	var stat syscall.Statfs_t

	syscall.Statfs(c.Path, &stat)

	// Available blocks * size per block = available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	fmt.Printf("%v>=%v", availableBytes, c.MinimumBytes)

	return availableBytes >= c.MinimumBytes, nil
}
