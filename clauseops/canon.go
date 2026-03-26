package clauseops

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/goivy/ivyutils"
)

// Canon returns a canonical s-expression for the Clauses.
// Format: (clauses fmlas:[sexp1 sexp2 ...] defs:[sexp1 sexp2 ...])
// Both Go and Python must produce identical output for identical state.
func (c *Clauses) Canon() iu.Canonical {
	if c == nil {
		return iu.Canonical("nil")
	}
	var fmlaStrs []string
	for _, f := range c.Fmlas {
		fmlaStrs = append(fmlaStrs, string(f.Sexp()))
	}
	var defStrs []string
	for _, d := range c.Defs {
		defStrs = append(defStrs, string(d.Sexp()))
	}
	return iu.Canonical(fmt.Sprintf("(clauses fmlas:[%s] defs:[%s])",
		strings.Join(fmlaStrs, " "), strings.Join(defStrs, " ")))
}
