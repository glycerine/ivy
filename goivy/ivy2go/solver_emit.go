package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// solver_emit.go mirrors ivy2cpp/solver_emit.go. The C++ file emits
// `add(__to_solver(...))` calls per state cell plus forall-quantified
// thunks for large functions; ivy2go's equivalent path goes through
// goivy.Translator and Solver.Assert.
//
// M9 ships a skeleton: per-state-symbol assertion emission that
// generates a comment marker. The real solver assertions land in
// later increments as fixtures exercise them.

// emitSetSolver writes pushStateIntoSolver and the per-action
// buildPrecondition helpers. Together they prepare a goivy.Solver
// with the current State as facts and an action-specific Clauses
// describing the inputs whose values we want the solver to choose.
//
// OPEN 055 / 055.1 / 055.2 progression:
//
//   - 055   wired the Solver round-trip with a `true` sanity check.
//   - 055.1 added GetModelClauses + per-input value extraction.
//   - 055.2 (this) emits a per-action precondition helper that
//     asserts state-symbol facts for every scalar bool state symbol
//     into the returned Clauses. Function-sorted symbols + the full
//     action.Pre derivation are the next sub-step (OPEN 055.3).
func (g *Generator) emitSetSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	w.line("// pushStateIntoSolver verifies the State is consistent with")
	w.line("// the module's axioms by running a SAT check through the")
	w.line("// provided goivy.Solver.")
	w.linef("func pushStateIntoSolver(_ *%s, sol *goivy.Solver) error {", g.StateTypeName)
	w.line("\tif sol == nil { return nil }")
	w.line("\ttrueExpr := goivy.NewConst(\"true\", goivy.Boolean)")
	w.line("\tok, err := sol.IsSat(trueExpr)")
	w.line("\tif err != nil { return err }")
	w.line("\tif !ok { return fmt.Errorf(\"ivy: state is unsatisfiable\") }")
	w.line("\treturn nil")
	w.line("}")
	w.blank()
	g.Ctx.AddImport("runtime", "fmt", "")

	// Shared helper that builds a Clauses asserting per-symbol state
	// facts. OPEN 055.3 extends 055.2 to cover function-sorted
	// (array-storage) state symbols via per-cell facts.
	w.line("// stateFactsAsClauses builds a *goivy.Clauses asserting that")
	w.line("// each scalar and array-storage state symbol matches its")
	w.line("// current value. Map-storage (large-domain) symbols are")
	w.line("// skipped because their domain is unbounded; the thunk-slot")
	w.line("// wiring (OPEN 061.1) handles those at read time.")
	w.linef("func stateFactsAsClauses(state *%s) *goivy.Clauses {", g.StateTypeName)
	w.line("\tvar fmlas []goivy.Expr")
	for _, sym := range g.stateSymbols() {
		g.emitStateSymbolFacts(w, sym)
	}
	w.line("\treturn goivy.NewClauses(fmlas, nil, goivy.EmptyAnnotation{})")
	w.line("}")
	w.blank()

	w.line("// mkBoolFact builds the formula (sym = val) where sym is a")
	w.line("// fresh bool-sorted goivy.Const named `name`. We return the")
	w.line("// LHS-only literal when val is true and its negation when false,")
	w.line("// matching the encoding goivy.Solver expects for bool facts.")
	w.line("func mkBoolFact(name string, val bool) goivy.Expr {")
	w.line("\tsym := goivy.NewConst(name, goivy.Boolean)")
	w.line("\tif val {")
	w.line("\t\treturn sym")
	w.line("\t}")
	w.line("\treturn &goivy.LogicNot{Body: sym}")
	w.line("}")
	w.blank()

	// Per-action precondition reifier — emitted via emitActionGen.
	// Helpers used by the reified expressions live here.
	g.emitPreconditionHelpers(w)
}

// emitStateSymbolFacts appends per-symbol fact assertions to the
// stateFactsAsClauses body. Dispatches on sort shape:
//   - bool scalar              → mkBoolFact
//   - enum scalar              → Eq to extension constant
//   - range/integer scalar     → Eq to integer-named constant
//   - bool-valued array        → nested loops + per-cell mkBoolFact
//   - enum-valued array        → nested loops + per-cell Eq
//   - hash-thunk function      → skip (handled by thunk-slot at read time)
func (g *Generator) emitStateSymbolFacts(w *goWriter, sym stateSymbol) {
	exported := goExportedName(sym.Name)
	if _, ok := sym.Sort.(*goivy.BooleanSort); ok {
		w.linef("\tfmlas = append(fmlas, mkBoolFact(%q, state.%s))", sym.Name, exported)
		return
	}
	if es, ok := sym.Sort.(*goivy.LogicEnumeratedSort); ok {
		g.emitEnumScalarFact(w, sym.Name, "state."+exported, es)
		return
	}
	// Range/uninterpreted scalar with a known cardinality → pin via
	// an integer-named constant. The solver sees `<name> = <int>`
	// which collides correctly with formula-side references via the
	// goivy translator's integer-sort handling.
	if !isFunctionSort(sym.Sort) {
		if card := goSortCard(g, sym.Sort); card > 0 {
			g.emitIntScalarFact(w, sym.Name, "state."+exported, sym.Sort)
		}
		return
	}
	fs := sym.Sort.(*goivy.LogicFunctionSort)
	domain := fs.Domain()
	st := goFunctionStorageFor(g, domain, fs.Range())
	if st.Kind != goStorageArray {
		// Hash-thunk: skip (handled by thunk-slot wiring at read time).
		return
	}
	switch rng := fs.Range().(type) {
	case *goivy.BooleanSort:
		_ = rng
		w.linef("\t// Per-cell facts for array-storage symbol %q.", sym.Name)
		for i, d := range st.Dims {
			w.linef("\tfor __i%d := 0; __i%d < %d; __i%d++ {", i, i, d, i)
		}
		w.line("\t\tcellName := " + cellNameExpr(sym.Name, len(st.Dims)))
		cellAcc := "state." + exported
		for i := range st.Dims {
			cellAcc += "[__i" + fmt.Sprintf("%d", i) + "]"
		}
		w.linef("\t\tfmlas = append(fmlas, mkBoolFact(cellName, %s))", cellAcc)
		for range st.Dims {
			w.line("\t}")
		}
	case *goivy.LogicEnumeratedSort:
		w.linef("\t// Per-cell facts (enum-valued) for array-storage symbol %q.", sym.Name)
		for i, d := range st.Dims {
			w.linef("\tfor __i%d := 0; __i%d < %d; __i%d++ {", i, i, d, i)
		}
		w.line("\t\tcellName := " + cellNameExpr(sym.Name, len(st.Dims)))
		cellAcc := "state." + exported
		for i := range st.Dims {
			cellAcc += "[__i" + fmt.Sprintf("%d", i) + "]"
		}
		g.emitEnumScalarFactIndented(w, "cellName", cellAcc, rng, "\t\t")
		for range st.Dims {
			w.line("\t}")
		}
	default:
		// Integer/range/uninterp-valued cells: pin via integer
		// equality if the range has a known cardinality.
		if card := goSortCard(g, fs.Range()); card > 0 {
			w.linef("\t// Per-cell facts (int-valued) for array-storage symbol %q.", sym.Name)
			for i, d := range st.Dims {
				w.linef("\tfor __i%d := 0; __i%d < %d; __i%d++ {", i, i, d, i)
			}
			w.line("\t\tcellName := " + cellNameExpr(sym.Name, len(st.Dims)))
			cellAcc := "state." + exported
			for i := range st.Dims {
				cellAcc += "[__i" + fmt.Sprintf("%d", i) + "]"
			}
			g.emitIntScalarFactIndented(w, "cellName", cellAcc, fs.Range(), "\t\t")
			for range st.Dims {
				w.line("\t}")
			}
		}
	}
}

// emitEnumScalarFact emits an `<name> = <extension>` equality for
// an enum-sorted scalar state symbol whose current Go value is
// stored at `lhsAcc`.
func (g *Generator) emitEnumScalarFact(w *goWriter, name, lhsAcc string, es *goivy.LogicEnumeratedSort) {
	sortCode, ok := g.reifySortAsGoCode(es)
	if !ok {
		return
	}
	extLits := make([]string, len(es.Extension))
	for i, ext := range es.Extension {
		extLits[i] = fmt.Sprintf("%q", ext)
	}
	w.linef("\t{")
	w.linef("\t\textNames := []string{%s}", strings.Join(extLits, ", "))
	w.linef("\t\tsortDecl := %s", sortCode)
	w.linef("\t\tfmlas = append(fmlas, &goivy.Eq{T1: goivy.NewConst(%q, sortDecl), T2: goivy.NewConst(extNames[int(%s)], sortDecl)})",
		name, lhsAcc)
	w.linef("\t}")
}

// emitEnumScalarFactIndented is the per-cell-loop variant: `name`
// is a Go expression (not a string literal), and `lhsAcc` indexes
// into the cell's storage. `indent` is the surrounding indent
// (e.g. "\t\t" inside two open loops).
func (g *Generator) emitEnumScalarFactIndented(w *goWriter, nameExpr, lhsAcc string, es *goivy.LogicEnumeratedSort, indent string) {
	sortCode, ok := g.reifySortAsGoCode(es)
	if !ok {
		return
	}
	extLits := make([]string, len(es.Extension))
	for i, ext := range es.Extension {
		extLits[i] = fmt.Sprintf("%q", ext)
	}
	w.linef("%s{", indent)
	w.linef("%s\textNames := []string{%s}", indent, strings.Join(extLits, ", "))
	w.linef("%s\tsortDecl := %s", indent, sortCode)
	w.linef("%s\tfmlas = append(fmlas, &goivy.Eq{T1: goivy.NewConst(%s, sortDecl), T2: goivy.NewConst(extNames[int(%s)], sortDecl)})",
		indent, nameExpr, lhsAcc)
	w.linef("%s}", indent)
}

// emitIntScalarFact emits an `<name> = <integer>` equality for a
// scalar state symbol of an integer-like sort (range / numeric enum
// / uninterpreted with int interp). The runtime resolves the int via
// `strconv.Itoa(int(lhsAcc))`.
func (g *Generator) emitIntScalarFact(w *goWriter, name, lhsAcc string, s goivy.Sort) {
	sortCode, ok := g.reifySortAsGoCode(s)
	if !ok {
		return
	}
	w.linef("\tfmlas = append(fmlas, &goivy.Eq{T1: goivy.NewConst(%q, %s), T2: goivy.NewConst(strconv.Itoa(int(%s)), %s)})",
		name, sortCode, lhsAcc, sortCode)
	g.Ctx.AddImport("runtime", "strconv", "")
}

// emitIntScalarFactIndented is the per-cell variant.
func (g *Generator) emitIntScalarFactIndented(w *goWriter, nameExpr, lhsAcc string, s goivy.Sort, indent string) {
	sortCode, ok := g.reifySortAsGoCode(s)
	if !ok {
		return
	}
	w.linef("%sfmlas = append(fmlas, &goivy.Eq{T1: goivy.NewConst(%s, %s), T2: goivy.NewConst(strconv.Itoa(int(%s)), %s)})",
		indent, nameExpr, sortCode, lhsAcc, sortCode)
	g.Ctx.AddImport("runtime", "strconv", "")
}

// cellNameExpr returns a Go expression that builds the synthetic
// per-cell symbol name like "link(0,1)" for `sym` over `arity` dims.
func cellNameExpr(sym string, arity int) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%q", sym+"("))
	for i := 0; i < arity; i++ {
		if i > 0 {
			parts = append(parts, `","`)
		}
		parts = append(parts, fmt.Sprintf("strconv.Itoa(__i%d)", i))
	}
	parts = append(parts, `")"`)
	return strings.Join(parts, " + ")
}

// emitPreconditionHelpers writes the small helpers each per-action
// buildPrecondition function leans on for reifying Pre.Fmlas.
// strconv is added lazily by finalize() when cellNameExpr usage
// in stateFactsAsClauses references it (array-storage symbols only).
func (g *Generator) emitPreconditionHelpers(w *goWriter) {
	w.line("// mkAnd builds the conjunction of fmlas, simplifying the")
	w.line("// trivial 0/1-term cases.")
	w.line("func mkAnd(fmlas []goivy.Expr) goivy.Expr {")
	w.line("\tswitch len(fmlas) {")
	w.line(`	case 0:`)
	w.line(`		return goivy.NewConst("true", goivy.Boolean)`)
	w.line("\tcase 1:")
	w.line("\t\treturn fmlas[0]")
	w.line("\t}")
	w.line("\treturn &goivy.LogicAnd{Terms: fmlas}")
	w.line("}")
	w.blank()
	w.line("// conjClauses returns a fresh Clauses whose fmlas are the")
	w.line("// concatenation of a.Fmlas and extra.")
	w.line("func conjClauses(a *goivy.Clauses, extra ...goivy.Expr) *goivy.Clauses {")
	w.line("\tif a == nil {")
	w.line("\t\treturn goivy.NewClauses(extra, nil, goivy.EmptyAnnotation{})")
	w.line("\t}")
	w.line("\tfmlas := append([]goivy.Expr{}, a.Fmlas...)")
	w.line("\tfmlas = append(fmlas, extra...)")
	w.line("\treturn goivy.NewClauses(fmlas, a.Defs, goivy.EmptyAnnotation{})")
	w.line("}")
	w.blank()
}
