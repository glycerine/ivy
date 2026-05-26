package ivy2go

import "github.com/glycerine/ivy/goivy"

// z3.go mirrors ivy2cpp/z3.go. The C++ version emits Z3 template
// specialisations (__from_solver, __to_solver, __randomize) inline.
// Per D1/D2, ivy2go routes ALL Z3 access through goivy.Solver — both
// at generator time and at the runtime of the emitted programs. So
// this file's role is narrower: emit import declarations and the
// small `solverSession` helper that wraps goivy.NewSolver lifecycle
// for the generated package.

// usesZ3 returns true when the resolved target produces solver-backed
// emission. Mirrors ivy2cpp Generator.usesZ3.
func (g *Generator) usesZ3() bool {
	if g == nil {
		return false
	}
	switch g.Config.Target {
	case "test", "gen":
		return true
	}
	return false
}

// actionGeneratorClassName mirrors ivy2cpp/z3.go:666. Returns the
// generator type name for an action, used by callers that reference
// the per-action generator struct by name.
func (g *Generator) actionGeneratorClassName(name string) string {
	return "actionGen_" + goExportedName(name)
}

// z3SortName mirrors ivy2cpp/z3.go:972. Returns a stable name string
// for a sort, used in solver-bridge code paths that need a textual
// sort identifier.
func z3SortName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.RangeSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.UninterpretedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	default:
		return sortName(s)
	}
}

// emitZ3Support is the runtime-emission counterpart of
// ivy2cpp/z3.go emitZ3Support. M9 emits a minimal helper that builds
// a goivy.Solver for the action generators to use.
func (g *Generator) emitZ3Support(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	if g.Ctx != nil {
		g.Ctx.AddImport("runtime", g.Config.GoivyImportPath, "")
	}
	w.line("// newIvySolver builds a goivy.Solver for use by the action")
	w.line("// generators emitted in action_gen.go. The caller is responsible")
	w.line("// for closing it when done.")
	w.linef("func newIvySolver() *goivy.Solver {")
	w.line("\treturn goivy.NewSolver(nil, goivy.DefaultSolverOptions())")
	w.line("}")
	w.blank()
}
