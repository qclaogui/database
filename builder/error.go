package builder

import (
	"strings"
)

var (
// ErrQueryError = errors.New("query Error")
)

type queryError struct {
	PSql  string
	PArgs []interface{}
	Err   error
}

// Error query error
func (e *queryError) Error() string {
	var errString strings.Builder
	errString.Grow(1024)
	errString.WriteString("\x1b[91mOops🔥\x1b[39m \n\t")
	errString.WriteString(e.Err.Error())
	errString.WriteString(";\n\tError SQL: ")
	errString.WriteString(e.PSql)
	errString.WriteString("\n\tBindings:[")
	for k, v := range e.PArgs {
		switch s := v.(type) {
		case string:
			if k == 0 {
				errString.WriteString(s)
			} else {
				errString.WriteString(", ")
				errString.WriteString(s)
			}
		case int:
			if k == 0 {
				errString.WriteString(itoa(s))
			} else {
				errString.WriteString(", ")
				errString.WriteString(itoa(s))
			}
		default:
			if k == 0 {
				errString.WriteString("???")
			} else {
				errString.WriteString(", ???")
			}
		}
	}
	errString.WriteString("]")
	return errString.String()
}
