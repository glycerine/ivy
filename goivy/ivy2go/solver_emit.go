package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// solver_emit.go mirrors ivy2cpp/solver_emit.go. The C++ file emits
// `add(__to_solver(...))` calls per state cell plus forall-quantified
// facts for large functions; ivy2go's equivalent path builds goivy.Expr
// trees and hands them to goivy.Solver.

// emitSetSolver writes pushStateIntoSolver and the per-action
// buildPrecondition helpers. Together they prepare a goivy.Solver
// with the current State as facts and an action-specific Clauses
// describing the inputs whose values we want the solver to choose.
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

	w.line("// stateFactsAsClauses builds a *goivy.Clauses asserting that")
	w.line("// each scalar, array-storage, and map-storage state symbol")
	w.line("// matches its current value. Function symbols are encoded with")
	w.line("// real Apply terms so formula-side preconditions and state facts")
	w.line("// refer to the same solver terms.")
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
	w.line("func mkBoolFactExpr(lhs goivy.Expr, val bool) goivy.Expr {")
	w.line("\tif val {")
	w.line("\t\treturn lhs")
	w.line("\t}")
	w.line("\treturn &goivy.LogicNot{Body: lhs}")
	w.line("}")
	w.blank()
	w.line("func mkBoolFact(name string, val bool) goivy.Expr {")
	w.line("\tsym := goivy.NewConst(name, goivy.Boolean)")
	w.line("\treturn mkBoolFactExpr(sym, val)")
	w.line("}")
	w.blank()
	w.line("func boolValueExpr(val bool) goivy.Expr {")
	w.line("\tif val {")
	w.line(`		return goivy.NewConst("true", goivy.Boolean)`)
	w.line("\t}")
	w.line(`	return goivy.NewConst("false", goivy.Boolean)`)
	w.line("}")
	w.blank()

	// Per-action precondition reifier — emitted via emitActionGen.
	// Helpers used by the reified expressions live here.
	g.emitPreconditionHelpers(w)
}

// isLargeType mirrors ivy2cpp/solver_emit.go:255. Returns true when
// the function sort's domain is too large for cell-by-cell solver
// assertions — either because some domain sort isn't enumerable or
// because the product of cardinalities exceeds largeThresh.
func (g *Generator) isLargeType(s goivy.Sort) bool {
	fs, ok := s.(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	for _, d := range dom {
		if !goIsAnyIntegerType(g, d) {
			return true
		}
	}
	product := 1
	for _, d := range dom {
		c := goSortCard(g, d)
		if c <= 0 {
			return true
		}
		if product <= largeThresh {
			product *= c
		}
	}
	return product > largeThresh
}

// emitStateSymbolFacts appends per-symbol fact assertions to the
// stateFactsAsClauses body. Dispatches on sort shape:
//   - bool scalar              → mkBoolFact
//   - enum scalar              → Eq to extension constant
//   - range/integer scalar     → Eq to integer-named constant
//   - array-storage function   → nested loops + per-cell Apply facts
//   - map-storage function     → forall Apply facts with thunk/map RHS
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
	// Record-valued (destructor) scalar — pin each scalar field via
	// the destructor accessor. Mirrors the cpp recursion in
	// emitSetSolverCustom for destructor record state symbols.
	if recName, ok := g.destructorStructName(sym.Sort); ok {
		g.emitRecordScalarFacts(w, sym.Name, "state."+exported, recName)
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
		g.emitMapStorageFunctionFacts(w, sym, fs, st)
		return
	}
	g.emitArrayStorageFunctionFacts(w, sym, fs, st)
}

func (g *Generator) emitArrayStorageFunctionFacts(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort, st goFunctionStorage) {
	ident, fnVar, sortVars, _, ok := g.emitFunctionFactPreamble(w, sym, fs)
	if !ok {
		return
	}
	exported := goExportedName(sym.Name)
	for i, d := range st.Dims {
		w.linef("\tfor __i%d := 0; __i%d < %d; __i%d++ {", i, i, d, i)
	}
	argExprs := make([]string, len(st.Dims))
	for i := range st.Dims {
		argExprs[i] = g.domainConstFromCompactIndex(fs.Domain()[i], sortVars[i], fmt.Sprintf("__i%d", i))
	}
	w.linef("\t\t__lhs_%s := mustApply(%s, %s)", ident, fnVar, strings.Join(argExprs, ", "))
	cellAcc := "state." + exported
	for i := range st.Dims {
		cellAcc += "[__i" + fmt.Sprintf("%d", i) + "]"
	}
	switch rng := fs.Range().(type) {
	case *goivy.BooleanSort:
		_ = rng
		w.linef("\t// Per-cell facts for array-storage symbol %q.", sym.Name)
		w.linef("\t\tfmlas = append(fmlas, mkBoolFactExpr(__lhs_%s, %s))", ident, cellAcc)
	case *goivy.LogicEnumeratedSort:
		w.linef("\t// Per-cell facts (enum-valued) for array-storage symbol %q.", sym.Name)
		g.emitEnumExprFactIndented(w, "__lhs_"+ident, cellAcc, rng, "\t\t")
	default:
		// Record-valued cells: per-cell per-field equalities so
		// the solver sees `<field-accessor>(<f>(<cell-args>)) =
		// <state.<F>[…].<Field>>` and can resolve formula
		// references through the destructor accessors.
		if recName, ok := g.destructorStructName(fs.Range()); ok {
			w.linef("\t// Per-cell record-field facts for array-storage symbol %q.", sym.Name)
			g.emitRecordCellFacts(w, "__lhs_"+ident, cellAcc, recName)
		} else if card := goSortCard(g, fs.Range()); card > 0 {
			w.linef("\t// Per-cell facts (int-valued) for array-storage symbol %q.", sym.Name)
			g.emitIntExprFactIndented(w, "__lhs_"+ident, cellAcc, fs.Range(), "\t\t")
		}
	}
	for range st.Dims {
		w.line("\t}")
	}
}

func (g *Generator) emitMapStorageFunctionFacts(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort, st goFunctionStorage) {
	ident, fnVar, sortVars, rangeSortCode, ok := g.emitFunctionFactPreamble(w, sym, fs)
	if !ok {
		return
	}
	exported := goExportedName(sym.Name)
	w.linef("\t// Quantified facts for map-storage symbol %q.", sym.Name)
	vars := make([]string, len(fs.Domain()))
	for i := range fs.Domain() {
		vars[i] = fmt.Sprintf("__v%d_%s", i, ident)
		w.linef("\t%s := mustNewVariable(%q, %s)", vars[i], fmt.Sprintf("X__%d", i), sortVars[i])
	}
	w.linef("\t__lhs_%s := mustApply(%s, %s)", ident, fnVar, strings.Join(vars, ", "))
	w.linef("\tvar __rhs_%s goivy.Expr = %s", ident, g.zeroValueExprCode(fs.Range(), rangeSortCode))
	w.linef("\tif state.__thunk_%s != nil {", exported)
	w.linef("\t\t__rhs_%s = state.__thunk_%s.toZ3Value([]goivy.Expr{%s})", ident, exported, strings.Join(vars, ", "))
	w.line("\t}")
	keyVar := "__key_" + ident
	valVar := "__val_" + ident
	w.linef("\tfor %s, %s := range state.%s {", keyVar, valVar, exported)
	condParts := make([]string, len(fs.Domain()))
	for i, d := range fs.Domain() {
		keyAccess := keyVar
		if len(fs.Domain()) > 1 {
			keyAccess = fmt.Sprintf("%s.Arg%d", keyVar, i)
		}
		w.linef("\t\t__keyExpr_%s_%d := %s", ident, i, g.domainConstFromGoValue(d, sortVars[i], keyAccess))
		condParts[i] = fmt.Sprintf("__condPart_%s_%d", ident, i)
		w.linef("\t\t%s := &goivy.Eq{T1: %s, T2: __keyExpr_%s_%d}", condParts[i], vars[i], ident, i)
	}
	condVar := "__cond_" + ident
	if len(condParts) == 1 {
		w.linef("\t\t%s := %s", condVar, condParts[0])
	} else {
		w.linef("\t\t%s := &goivy.LogicAnd{Terms: []goivy.Expr{%s}}", condVar, strings.Join(condParts, ", "))
	}
	g.emitValueExprBinding(w, "__valExpr_"+ident, valVar, fs.Range(), rangeSortCode, "\t\t")
	w.linef("\t\t__rhs_%s = mustNewIte(%s, __valExpr_%s, __rhs_%s)", ident, condVar, ident, ident)
	w.line("\t}")
	w.linef("\tfmlas = append(fmlas, &goivy.ForAll{Variables: []*goivy.LogicVariable{%s}, Body: &goivy.Eq{T1: __lhs_%s, T2: __rhs_%s}})",
		strings.Join(vars, ", "), ident, ident)
	_ = st
}

func (g *Generator) emitFunctionFactPreamble(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort) (ident, fnVar string, sortVars []string, rangeSortCode string, ok bool) {
	g.requireMustHelpers()
	g.Ctx.AddImport("runtime", "strconv", "")
	ident = goIdent(sym.Name)
	sortVars = make([]string, len(fs.Domain()))
	for i, d := range fs.Domain() {
		sortCode, sortOK := g.reifySortAsGoCode(d)
		if !sortOK {
			return "", "", nil, "", false
		}
		sortVars[i] = fmt.Sprintf("__sort_%s_%d", ident, i)
		w.linef("\t%s := %s", sortVars[i], sortCode)
	}
	rangeSortCode, ok = g.reifySortAsGoCode(fs.Range())
	if !ok {
		return "", "", nil, "", false
	}
	fnVar = "__fn_" + ident
	sortArgs := append([]string{}, sortVars...)
	sortArgs = append(sortArgs, rangeSortCode)
	w.linef("\t%s := goivy.NewConst(%q, mustNewFunctionSort(%s))", fnVar, sym.Name, strings.Join(sortArgs, ", "))
	return ident, fnVar, sortVars, rangeSortCode, true
}

func (g *Generator) domainConstFromGoValue(s goivy.Sort, sortVar, value string) string {
	if es, ok := s.(*goivy.LogicEnumeratedSort); ok && !isNumericEnum(es) {
		names := make([]string, len(es.Extension))
		for i, ext := range es.Extension {
			names[i] = fmt.Sprintf("%q", ext)
		}
		return fmt.Sprintf("goivy.NewConst([]string{%s}[int(%s)], %s)", strings.Join(names, ", "), value, sortVar)
	}
	return fmt.Sprintf("goivy.NewConst(strconv.Itoa(int(%s)), %s)", value, sortVar)
}

func (g *Generator) domainConstFromCompactIndex(s goivy.Sort, sortVar, value string) string {
	return g.domainConstFromGoValue(s, sortVar, g.rangeValueFromCompactIndexExpr(s, value))
}

func (g *Generator) zeroValueExprCode(s goivy.Sort, sortCode string) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return `goivy.NewConst("false", goivy.Boolean)`
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) > 0 && !isNumericEnum(st) {
			return fmt.Sprintf("goivy.NewConst(%q, %s)", st.Extension[0], sortCode)
		}
		return fmt.Sprintf("goivy.NewConst(%q, %s)", "0", sortCode)
	default:
		return fmt.Sprintf("goivy.NewConst(%q, %s)", "0", sortCode)
	}
}

func (g *Generator) emitValueExprBinding(w *goWriter, name, value string, sort goivy.Sort, sortCode string, indent string) {
	switch st := sort.(type) {
	case *goivy.BooleanSort:
		_ = st
		w.linef("%s%s := boolValueExpr(%s)", indent, name, value)
	case *goivy.LogicEnumeratedSort:
		if !isNumericEnum(st) {
			extLits := make([]string, len(st.Extension))
			for i, ext := range st.Extension {
				extLits[i] = fmt.Sprintf("%q", ext)
			}
			w.linef("%s__names_%s := []string{%s}", indent, name, strings.Join(extLits, ", "))
			w.linef("%s%s := goivy.NewConst(__names_%s[int(%s)], %s)", indent, name, name, value, sortCode)
			return
		}
		w.linef("%s%s := goivy.NewConst(strconv.Itoa(int(%s)), %s)", indent, name, value, sortCode)
		g.Ctx.AddImport("runtime", "strconv", "")
	default:
		w.linef("%s%s := goivy.NewConst(strconv.Itoa(int(%s)), %s)", indent, name, value, sortCode)
		g.Ctx.AddImport("runtime", "strconv", "")
	}
}

func (g *Generator) emitEnumExprFactIndented(w *goWriter, lhsExpr, lhsAcc string, es *goivy.LogicEnumeratedSort, indent string) {
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
	w.linef("%s\tfmlas = append(fmlas, &goivy.Eq{T1: %s, T2: goivy.NewConst(extNames[int(%s)], sortDecl)})",
		indent, lhsExpr, lhsAcc)
	w.linef("%s}", indent)
}

func (g *Generator) emitIntExprFactIndented(w *goWriter, lhsExpr, lhsAcc string, s goivy.Sort, indent string) {
	sortCode, ok := g.reifySortAsGoCode(s)
	if !ok {
		return
	}
	w.linef("%sfmlas = append(fmlas, &goivy.Eq{T1: %s, T2: goivy.NewConst(strconv.Itoa(int(%s)), %s)})",
		indent, lhsExpr, lhsAcc, sortCode)
	g.Ctx.AddImport("runtime", "strconv", "")
}

// emitRecordCellFacts is the per-cell-loop variant of
// emitRecordScalarFacts. `nameExpr` is a Go expression for the
// function application term, and `lhsAcc` indexes into the cell's storage.
func (g *Generator) emitRecordCellFacts(w *goWriter, nameExpr, lhsAcc, recName string) {
	for _, f := range g.destructorScalarFields(recName) {
		field := goExportedName(memName(f.Name))
		access := lhsAcc + "." + field
		destrSort, ok := g.reifySortAsGoCode(f.DestructorC.CSort)
		if !ok {
			continue
		}
		destrVar := "__destr_" + goIdent(f.FullName)
		w.linef("\t\t%s := goivy.NewConst(%q, %s)", destrVar, f.FullName, destrSort)
		w.linef("\t\t__field_%s := mustApply(%s, %s)", goIdent(f.FullName), destrVar, nameExpr)
		switch fs := f.Sort.(type) {
		case *goivy.BooleanSort:
			_ = fs
			w.linef("\t\tfmlas = append(fmlas, mkBoolFactExpr(__field_%s, %s))", goIdent(f.FullName), access)
		case *goivy.LogicEnumeratedSort:
			sortCode, ok := g.reifySortAsGoCode(fs)
			if !ok {
				continue
			}
			extLits := make([]string, len(fs.Extension))
			for i, ext := range fs.Extension {
				extLits[i] = fmt.Sprintf("%q", ext)
			}
			w.line("\t\t{")
			w.linef("\t\t\textNames := []string{%s}", strings.Join(extLits, ", "))
			w.linef("\t\t\tsortDecl := %s", sortCode)
			w.linef("\t\t\tfmlas = append(fmlas, &goivy.Eq{T1: __field_%s, T2: goivy.NewConst(extNames[int(%s)], sortDecl)})",
				goIdent(f.FullName), access)
			w.line("\t\t}")
		default:
			if goSortCard(g, f.Sort) > 0 {
				sortCode, ok := g.reifySortAsGoCode(f.Sort)
				if !ok {
					continue
				}
				w.linef("\t\tfmlas = append(fmlas, &goivy.Eq{T1: __field_%s, T2: goivy.NewConst(strconv.Itoa(int(%s)), %s)})",
					goIdent(f.FullName), access, sortCode)
				g.Ctx.AddImport("runtime", "strconv", "")
			}
		}
	}
}

// emitRecordScalarFacts emits per-destructor-field facts for a
// record-valued state symbol. Mirrors the cpp recursion in
// emitSetSolverCustom's destructor branch (ivy2cpp/solver_emit.go).
// For each scalar destructor field of `recName`, emits an equality
// pinning `<destr-name>(<sym>) = <state.<Field>>` so the solver
// sees the record's current contents.
func (g *Generator) emitRecordScalarFacts(w *goWriter, symName, lhsAcc, recName string) {
	for _, f := range g.destructorScalarFields(recName) {
		field := goExportedName(memName(f.Name))
		access := lhsAcc + "." + field
		// Destructor-side name for the field accessor (e.g. "foo.x")
		// so the solver formula references match formula-side
		// `Apply(foo.x, <sym>)` references.
		destrName := f.FullName
		switch fs := f.Sort.(type) {
		case *goivy.BooleanSort:
			_ = fs
			w.linef("\tfmlas = append(fmlas, mkBoolFact(%q, %s))",
				symName+"."+destrName, access)
		case *goivy.LogicEnumeratedSort:
			g.emitEnumScalarFact(w, symName+"."+destrName, access, fs)
		default:
			if goSortCard(g, f.Sort) > 0 {
				g.emitIntScalarFact(w, symName+"."+destrName, access, f.Sort)
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

// emitPreconditionHelpers writes the small helpers each per-action
// buildPrecondition function leans on for reifying Pre.Fmlas.
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
