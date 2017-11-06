package install

import (
	"errors"
	"regexp"
	"strings"
)

func (dl DockerLogs) validate() (bool, []error) {
	v := newValidator()
	if dl.MaxSize != "-1" {
		if emptyString(dl.MaxSize) {
			v.addError(errors.New("Max size cannot be empty"))
		} else if valid, _ := regexp.MatchString("[0-9]+[kmg]", dl.MaxSize); !valid {
			v.addError(errors.New("Max size must be numberic followed by either k, m or g (lowercase)"))
		}
	}
	if dl.MaxFile < 1 {
		v.addError(errors.New("Max file must be greater than or equal to 1"))
	}
	return v.valid()
}

func emptyString(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
