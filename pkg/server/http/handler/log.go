package handler

import "fmt"

func errorf(format string, v ...interface{}) string {
	return fmt.Sprintf("ERROR %s", fmt.Sprintf(format, v))
}
