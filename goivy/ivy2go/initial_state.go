package ivy2go

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// initial_state.go mirrors ivy2cpp/initial_state.go. M5 implements the
// nondeterministic (impl/repl/class) path: every unconstrained scalar
// gets ivyChoose; named maps stay empty until first write.
//
// The solver-driven path for target=test/gen is deferred to M9 and
// will plug into the same emitOneInitialState hook.

// emitOneInitialState writes statements into `init.go` that populate the
// receiver `s` with sensible nondeterministic defaults. For the
// non-solver targets it's enough to set scalar booleans / range values
// to a randomised choice (mirroring Python's `mk_nondet` path); map
// storage is left empty since map reads on missing keys return the
// element zero value (matching ivy2cpp's hash_thunk semantics).
func (g *Generator) emitOneInitialState(w *goWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	// Mirror ivy2cpp/initial_state.go emitDefaultInitialState: every
	// scalar state symbol gets a nondeterministic choice. Solver-
	// driven init for axiom-constrained symbols lands in M9; mark
	// the gap only when an actual InitCond formula would be missed,
	// not for every symbol.
	if g.Mod.InitCond != nil && !g.Mod.InitCond.IsTrue() && (g.Config.Target == "test" || g.Config.Target == "gen") {
		w.linef("// TODO(M9): solver-driven init for InitCond (%d fmlas)", len(g.Mod.InitCond.Fmlas))
	}
	for _, sym := range g.stateSymbols() {
		if isFunctionSort(sym.Sort) {
			// Function-sorted symbols (arrays / maps) are handled
			// either by NewState (map allocation) or by individual
			// per-cell init steps lowered upstream. M5 leaves them
			// alone; refine in M8/M9.
			continue
		}
		g.emitScalarChoice(w, sym)
	}
}

// emitScalarChoice writes a single nondet initialization for a scalar
// state symbol. Mirrors Python's `mk_nondet_sym` (ivy_to_cpp.py:3019).
func (g *Generator) emitScalarChoice(w *goWriter, sym stateSymbol) {
	field := "s." + goExportedName(sym.Name)
	switch sym.Sort.(type) {
	case *goivy.BooleanSort:
		w.linef("%s = ivyChoose(2) == 1", field)
	default:
		card := goSortCard(g, sym.Sort)
		if card <= 0 {
			// Fall back to zero-initialisation when cardinality is
			// unknown; Go's zero value is sensible for scalars.
			return
		}
		typeName := g.goType(sym.Sort)
		w.linef("%s = %s(ivyChoose(%d))", field, typeName, card)
	}
	_ = fmt.Sprintf
}

// isFunctionSort returns true when s is a function sort. Local helper
// so callers don't need to type-assert inline.
func isFunctionSort(s goivy.Sort) bool {
	_, ok := s.(*goivy.LogicFunctionSort)
	return ok
}
