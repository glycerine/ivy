package goivy

import (
	"fmt"
	"strings"
)

// ActionCanon returns a canonical s-expression for any Action using the
// generic interface methods Name() and ActionArgs(). This avoids needing
// a Canon() method on every individual action type.
// Format: (actionName args:[arg1.Sexp() arg2.Sexp() ...])
func ActionCanon(a Action) Canonical {
	if a == nil {
		return "nil"
	}
	var argStrs []string
	for _, arg := range a.ActionArgs() {
		if arg == nil {
			argStrs = append(argStrs, "nil")
		} else {
			argStrs = append(argStrs, string(arg.Sexp()))
		}
	}
	return Canonical(fmt.Sprintf("(%s args:[%s])", a.Name(), strings.Join(argStrs, " ")))
}
