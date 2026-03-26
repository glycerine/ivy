package z3bridge

import (
	"fmt"
	"regexp"
	"sync/atomic"

	"github.com/glycerine/goivy/xtracer"
)

// z3VarPattern matches Z3 internal variable names like !k!0, !k!1, etc.
var z3VarPattern = regexp.MustCompile(`!k!\d+`)

// z3CheckCounter provides a global sequence number for Z3 check calls.
var z3CheckCounter int64

// NormalizeZ3VarNames replaces Z3 internal names (!k!N) with
// deterministic equivalents (!v!N) based on first-occurrence
// order. This makes Z3 sexpr output comparable across Go (cgo)
// and Python bindings which may use different internal numbering.
//
// details:
//
// !k!N is Z3's internal naming scheme for bound variables
// in quantifiers — k stands for the internal index and N
// is a number (e.g., !k!0, !k!3, !k!17). Z3 assigns
// these automatically and the numbering can differ between
// the Go cgo bindings and the Python bindings even for
// semantically identical formulas.
//
// The normalization replaces each !k!N with !v!N where
// v stands for "variable" and N is now the first-occurrence
// order (0, 1, 2, ...) within the sexpr string. So:
//
// - First unique !k! name encountered → !v!0
// - Second unique !k! name → !v!1
// - Same !k! name seen again → same !v! replacement
//
// This makes the output deterministic based on structure
// rather than Z3 internals. Both Go and Python apply the
// same regex-based renaming, so if the formulas are
// structurally identical, the normalized output matches.
func NormalizeZ3VarNames(sexpr string) string {
	counter := 0
	seen := make(map[string]string)
	return z3VarPattern.ReplaceAllStringFunc(sexpr, func(name string) string {
		if replacement, ok := seen[name]; ok {
			return replacement
		}
		replacement := fmt.Sprintf("!v!%d", counter)
		seen[name] = replacement
		counter++
		return replacement
	})
}

// TraceCheck emits Z3 solver state and check result via xtracer.
// Called automatically from Solver.Check() when xtracer is enabled.
func TraceCheck(s *Solver, result CheckResult) {
	if !xtracer.Enabled {
		return
	}
	seq := atomic.AddInt64(&z3CheckCounter, 1)
	smt2 := NormalizeZ3VarNames(s.String())

	var rs string
	switch result {
	case Sat:
		rs = "sat"
	case Unsat:
		rs = "unsat"
	default:
		rs = "unknown"
	}
	xtracer.Trace("z3.check seq=%d result=%s smt2=%s", seq, rs, smt2)
}
