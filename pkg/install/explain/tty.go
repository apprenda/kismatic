package explain

import (
	"io"

	isatty "github.com/mattn/go-isatty"
)

func isTerminal(out io.Writer) bool {
	type fd interface {
		Fd() uintptr
	}
	switch w := out.(type) {
	case fd:
		return isatty.IsTerminal(w.Fd())
	default:
		return false
	}
}
