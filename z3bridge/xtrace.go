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

// NormalizeZ3VarNames replaces Z3 internal names (!k!N) with deterministic
// equivalents (!v!N) based on first-occurrence order. This makes Z3 sexpr
// output comparable across Go (cgo) and Python bindings which may use
// different internal numbering.
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
