package rule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// FileContentMatches is a rule that verifies that the contents of a file
// match the regular expression provided
type FreeSpace struct {
	Meta
	MinimumBytes string
	Path         string
}

// Name is the name of the rule
func (f FreeSpace) Name() string {
	return fmt.Sprintf("Contents of %q match: %s", f.Path, f.MinimumBytes)
}

// IsRemoteRule returns true if the rule is to be run from outside of the node
func (f FreeSpace) IsRemoteRule() bool { return false }

// Validate the rule
func (f FreeSpace) Validate() []error {
	errs := []error{}
	if f.Path == "" {
		errs = append(errs, errors.New("Path cannot be empty"))
	} else if !strings.HasPrefix(f.Path, "/") {
		errs = append(errs, errors.New("Path must start with /"))
	}

	if f.MinimumBytes == "" {
		errs = append(errs, errors.New("MinimumBytes cannot be empty"))
	} else {
		if _, err := f.MinimumBytesAsUint64(); err != nil {
			errs = append(errs, fmt.Errorf("MinimumBytes contains an invalid unsigned integer: %v", err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (f FreeSpace) MinimumBytesAsUint64() (uint64, error) {
	return strconv.ParseUint(f.MinimumBytes, 10, 0)
}
