package ivy2cpp

import (
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Assignment emission. Mirrors Python ivy_to_cpp.py:3624-3764, which routes
// every AssignAction through one of three paths:
//
//   - emit_assign_simple   — no free variables in LHS (a single point write).
//   - the bounded two-phase path — quantified LHS, derive loop bounds, then
//     use a function-typed temporary to avoid aliasing for self-referential
//     RHS reads like f(X) := f(X) + 1.
//   - emit_assign_large    — quantified LHS with bounds-error; falls back to
//     a thunk via make_thunk (Phase C; see thunk.go).
//
// emitExtensionalRelationClear short-circuits the all-false extensional
// relation case before any of the above. Keep that contract.

// emitAssignSimple is the Python `emit_assign_simple` (ivy_to_cpp.py:3624-3652)
// — the LHS has no free variables, so a single C++ assignment suffices.
// When Config.Trace is set and the LHS root name is not in a `:`-namespaced
// scope (Python's `':' not in self.args[0].rep.name`), an extra trace line
// of the form `__ivy_out << "  write(<lhs>," << (<rhs>) << ")" << std::endl;`
// is emitted just before the assignment.
func (g *Generator) emitAssignSimple(w *cppWriter, a *goivy.LogicAssignAction) {
	lhs, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		return
	}
	var rhs string
	if g.Config.Target == "test" {
		rhs, err = g.emitExprWithHeader(w, a.RHS)
	} else {
		rhs, err = g.emitExpr(a.RHS)
	}
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		return
	}
	rhs = g.maybeVariantUpcast(a.LHS.NodeSort(), a.RHS.NodeSort(), rhs, "")
	if g.Config.Trace && !lhsHasNamespacedName(a.LHS) {
		nf := g.numberFormat()
		trace, err := g.emitTracedLHS(a.LHS)
		if err != nil {
			g.unsupported(w, "unsupported assignment trace lhs: %s", err.Error())
			return
		}
		w.linef(`__ivy_out%s << "  write("%s << "," << (%s) << ")" << std::endl;`, nf, trace, rhs)
	}
	w.linef("%s = %s;", lhs, rhs)
}

// lhsHasNamespacedName mirrors Python's gate `':' not in self.args[0].rep.name`
// (ivy_to_cpp.py:3627): assignment traces are suppressed when the LHS root
// symbol carries a `:`-prefixed namespace (e.g. internal helpers like
// `loc:x` or `__ts:tick`).
func lhsHasNamespacedName(lhs goivy.Expr) bool {
	name := lhsRootName(lhs)
	return strings.Contains(name, ":")
}

// lhsRootName returns the "rep.name" Python uses: the leaf symbol name at
// the root of an LHS expression. For destructor chains (App(f, App(g, x)))
// this peels until the leaf symbol.
func lhsRootName(lhs goivy.Expr) string {
	for {
		switch v := lhs.(type) {
		case *goivy.Apply:
			if fn, ok := v.Func.(goivy.Expr); ok {
				lhs = fn
				continue
			}
			return ""
		case *goivy.Const:
			return v.Name
		default:
			return goivy.ExprName(v)
		}
	}
}

// emitTracedLHS mirrors Python emit_traced_lhs: emit a stream chain that
// prints the source-level LHS name while evaluating application arguments.
// This keeps traces stable across the C++ storage choices used for arrays,
// maps, destructor-backed structs, and future storage helpers.
func (g *Generator) emitTracedLHS(lhs goivy.Expr) (string, error) {
	var b strings.Builder
	if err := g.emitTracedLHSInto(&b, lhs); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (g *Generator) emitTracedLHSInto(b *strings.Builder, lhs goivy.Expr) error {
	switch n := lhs.(type) {
	case *goivy.Const:
		b.WriteString(` << "`)
		b.WriteString(escapeString(n.Name))
		b.WriteString(`"`)
		return nil
	case *goivy.LogicVariable:
		b.WriteString(` << "`)
		b.WriteString(escapeString(n.Name))
		b.WriteString(`"`)
		return nil
	case *goivy.Apply:
		name := goivy.ExprName(n.Func)
		if name == "" {
			return nil
		}
		if g.isDestructorName(name) && len(n.Terms) > 0 {
			if err := g.emitTracedLHSInto(b, n.Terms[0]); err != nil {
				return err
			}
			b.WriteString(` << ".`)
			b.WriteString(escapeString(memName(name)))
			b.WriteString(`"`)
			if len(n.Terms) > 1 {
				if err := g.emitTraceArgs(b, n.Terms[1:]); err != nil {
					return err
				}
			}
			return nil
		}
		b.WriteString(` << "`)
		b.WriteString(escapeString(name))
		b.WriteString(`"`)
		if len(n.Terms) > 0 {
			return g.emitTraceArgs(b, n.Terms)
		}
		return nil
	default:
		code, err := g.emitExpr(lhs)
		if err != nil {
			return err
		}
		b.WriteString(` << "`)
		b.WriteString(escapeString(code))
		b.WriteString(`"`)
		return nil
	}
}

func (g *Generator) emitTraceArgs(b *strings.Builder, args []goivy.Expr) error {
	b.WriteString(` << "("`)
	for i, arg := range args {
		if i > 0 {
			b.WriteString(` << ","`)
		}
		code, err := g.emitExpr(arg)
		if err != nil {
			return err
		}
		b.WriteString(` << `)
		b.WriteString(code)
	}
	b.WriteString(` << ")"`)
	return nil
}

// assignBoundsExpr implements Python's bexpr trick from emit_assign
// (ivy_to_cpp.py:3717-3720). When RHS has the shape Ite(cond, then, lhs)
// and cond does not mention the modified symbol, cond can be used as a
// bounds-tightening expression for the loop over the LHS free variables.
//
// Returns nil when the trick does not apply (Python's `bexpr = il.And()` =
// trivially true).
func (g *Generator) assignBoundsExpr(a *goivy.LogicAssignAction) goivy.Expr {
	ite, ok := a.RHS.(*goivy.LogicIte)
	if !ok {
		return nil
	}
	if !ite.Else.Equal(a.LHS) {
		return nil
	}
	// Python uses self.modifies()[0] — the destructor-peeled root symbol.
	mods := goivy.ModifiesSingle(a, g.actionsCfg())
	if len(mods) == 0 {
		return nil
	}
	used := goivy.UsedSymbolsAst(ite.Cond)
	for _, m := range mods {
		if _, found := used.Get2(goivy.Key(m)); found {
			return nil
		}
	}
	return ite.Cond
}

// openAssignmentLoopsBounded extends openAssignmentLoops with bounds
// tightening from a body expression (Python's bexpr in emit_assign). When
// body is non-nil and provides inequalities constraining a free variable,
// the loop for that variable uses the body-derived bounds via
// loopHeaderForSortBounds (Python `open_bounded_loops` integer-loop
// branch, ivy_to_cpp.py:3671-3677).
//
// When body is nil or provides no body-derived bounds for a variable, the
// emission falls back to loopHeaderForVar, preserving the existing goivy
// loop convention (range-for over finite enums, half-open over ranges).
//
// Returns the number of opened loops and a success indicator. On failure
// (a free variable whose sort has no cardinality / iteration anchor), any
// already-opened loops are closed and (0, false) is returned. The caller
// is expected to fall back to emitAssignLarge (thunk-based assignment).
// This function does NOT call g.unsupported on failure — error reporting
// is the caller's responsibility.
func (g *Generator) openAssignmentLoopsBounded(w *cppWriter, lhs goivy.Expr, body goivy.Expr) (int, bool) {
	vars := goivy.VariablesAstList(lhs)
	opened := 0
	for _, v := range vars {
		var header string
		var err error

		if body != nil && cppIsAnyIntegerType(g, v.VSort) {
			var bes []boundExpr
			g.matchBoundExprs(v, body, true, &bes)
			if len(bes) > 0 {
				lo, hi, gerr := g.getBounds(v, nil, body, true)
				if gerr == nil {
					header, err = g.loopHeaderForSortBounds(v.VSort, varName(v.Name), lo, hi)
				}
			}
		}

		if header == "" {
			header, err = g.loopHeaderForVar(v)
		}
		if err != nil {
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return 0, false
		}
		w.open(header)
		opened++
	}
	return opened, true
}

func (g *Generator) assignmentLoopHeadersBounded(lhs goivy.Expr, body goivy.Expr) ([]string, bool) {
	vars := goivy.VariablesAstList(lhs)
	headers := make([]string, 0, len(vars))
	for _, v := range vars {
		var header string
		var err error

		if body != nil && cppIsAnyIntegerType(g, v.VSort) {
			var bes []boundExpr
			g.matchBoundExprs(v, body, true, &bes)
			if len(bes) > 0 {
				lo, hi, gerr := g.getBounds(v, nil, body, true)
				if gerr == nil {
					header, err = g.loopHeaderForSortBounds(v.VSort, varName(v.Name), lo, hi)
				}
			}
		}

		if header == "" {
			header, err = g.loopHeaderForVar(v)
		}
		if err != nil {
			return nil, false
		}
		headers = append(headers, header)
	}
	return headers, true
}

func emitPythonTestAssignmentLoopHeaders(w *cppWriter, headers []string) {
	for _, h := range headers {
		w.raw(h)
		w.raw("\n")
	}
}

func closePythonTestAssignmentLoopHeaders(w *cppWriter, headers []string) {
	for range headers {
		w.raw("}\n")
	}
}

// canOpenAssignmentLoopsBounded performs a dry-run of
// openAssignmentLoopsBounded without writing to a cppWriter. Used by
// emitAssign to decide whether to take the bounded two-phase path or
// fall back to the thunk-based emit_assign_large.
//
// Mirrors Python's open_bounded_loops returning a BoundsError sentinel —
// here we just return false. Conservative: if any free variable cannot
// open a loop header, return false so the caller picks emit_assign_large.
func (g *Generator) canOpenAssignmentLoopsBounded(lhs goivy.Expr, body goivy.Expr) bool {
	vars := goivy.VariablesAstList(lhs)
	for _, v := range vars {
		if body != nil && cppIsAnyIntegerType(g, v.VSort) {
			var bes []boundExpr
			g.matchBoundExprs(v, body, true, &bes)
			if len(bes) > 0 {
				if _, _, err := g.getBounds(v, nil, body, true); err == nil {
					continue
				}
			}
		}
		if _, err := g.loopHeaderForVar(v); err != nil {
			return false
		}
	}
	return true
}

// actionsCfg returns the ActionsConfig for ModifiesSingle. It may be nil
// when the generator is constructed without a module (e.g., in some unit
// tests); ModifiesSingle tolerates nil.
func (g *Generator) actionsCfg() *goivy.ActionsConfig {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil {
		return nil
	}
	return g.Mod.Cfg.ActCfg
}

// emitAssignTwoPhase implements the bounded two-phase pattern from Python's
// `emit_assign` (ivy_to_cpp.py:3725-3764). For a quantified LHS f(X...) and
// some RHS that may itself reference f(X...) (or any other state), we:
//
//  1. allocate a fresh function-typed temporary sym of the same domain;
//  2. loop over X..., evaluating the RHS and writing into sym;
//  3. loop over X... again, copying sym back to f(X...).
//
// The second loop is required even when RHS doesn't read the LHS, because
// `f(X) := g(X)` is fine to do in-place but `f(X) := f(X) + 1` is not — and
// we don't try to prove non-aliasing at codegen time.
func (g *Generator) emitAssignTwoPhase(w *cppWriter, a *goivy.LogicAssignAction, vs []*goivy.LogicVariable) {
	if g.Config.Target == "test" {
		g.emitAssignTwoPhasePythonTest(w, a, vs)
		return
	}

	// Build a fresh FunctionSort over the loop variables and the RHS sort.
	sorts := make([]goivy.Sort, 0, len(vs)+1)
	for _, v := range vs {
		sorts = append(sorts, v.VSort)
	}
	sorts = append(sorts, a.RHS.NodeSort())
	tsort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		g.unsupported(w, "unsupported temp function sort: %s", err.Error())
		return
	}
	tmpName := g.nextTemp("__ivy_tmp")
	tmpSym := goivy.NewConst(tmpName, tsort)

	// Declare the temp at action-body scope. cppStorageDecl picks array
	// vs hash_thunk based on cardinality.
	w.linef("%s;", g.cppStorageDecl(tmpName, tsort, ""))

	// Build the synthetic LHS expression sym(vs...) used in both phases.
	tmpArgs := make([]goivy.Expr, len(vs))
	for i, v := range vs {
		tmpArgs[i] = v
	}
	tmpLHS := goivy.NewApplyUnchecked(tmpSym, tmpArgs...)

	// Python's bexpr trick: if RHS is Ite(cond, then, lhs) and cond does
	// not mention the modified symbol, use cond to tighten loop bounds in
	// both phases (ivy_to_cpp.py:3717-3720).
	body := g.assignBoundsExpr(a)

	// Phase 1: write RHS into sym.
	loops, ok := g.openAssignmentLoopsBounded(w, a.LHS, body)
	if !ok {
		return
	}
	lhsCode, err := g.emitExpr(tmpLHS)
	if err != nil {
		g.unsupported(w, "unsupported temp lhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	rhsCode, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	w.linef("%s = %s;", lhsCode, rhsCode)
	g.closeAssignmentLoops(w, loops)

	// Phase 2: copy sym back to the LHS. Re-emit the same loops with the
	// same body so bounds-tightening is identical to phase 1.
	loops, ok = g.openAssignmentLoopsBounded(w, a.LHS, body)
	if !ok {
		return
	}
	finalLHS, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	tmpRHS, err := g.emitExpr(tmpLHS)
	if err != nil {
		g.unsupported(w, "unsupported temp rhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	tmpRHS = g.maybeVariantUpcast(a.LHS.NodeSort(), a.RHS.NodeSort(), tmpRHS, "")
	w.linef("%s = %s;", finalLHS, tmpRHS)
	g.closeAssignmentLoops(w, loops)
}

func (g *Generator) emitAssignTwoPhasePythonTest(w *cppWriter, a *goivy.LogicAssignAction, vs []*goivy.LogicVariable) {
	sorts := make([]goivy.Sort, 0, len(vs)+1)
	for _, v := range vs {
		sorts = append(sorts, v.VSort)
	}
	sorts = append(sorts, a.RHS.NodeSort())
	tsort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		g.unsupported(w, "unsupported temp function sort: %s", err.Error())
		return
	}
	body := g.assignBoundsExpr(a)
	headers, ok := g.assignmentLoopHeadersBounded(a.LHS, body)
	if !ok {
		g.emitAssignLarge(w, a, vs)
		return
	}
	tmpName := g.nextTemp("__tmp")
	tmpSym := goivy.NewConst(tmpName, tsort)
	w.linef("%s;", g.cppStorageDecl(tmpName, tsort, ""))

	tmpArgs := make([]goivy.Expr, len(vs))
	for i, v := range vs {
		tmpArgs[i] = v
	}
	tmpLHS := goivy.NewApplyUnchecked(tmpSym, tmpArgs...)

	emitPythonTestAssignmentLoopHeaders(w, headers)
	lhsCode, err := g.emitExpr(tmpLHS)
	if err != nil {
		g.unsupported(w, "unsupported temp lhs: %s", err.Error())
		closePythonTestAssignmentLoopHeaders(w, headers)
		return
	}
	rhsCode, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		closePythonTestAssignmentLoopHeaders(w, headers)
		return
	}
	w.linef("%s = %s;", lhsCode, rhsCode)
	closePythonTestAssignmentLoopHeaders(w, headers)

	emitPythonTestAssignmentLoopHeaders(w, headers)
	finalLHS, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		closePythonTestAssignmentLoopHeaders(w, headers)
		return
	}
	tmpRHS, err := g.emitExpr(tmpLHS)
	if err != nil {
		g.unsupported(w, "unsupported temp rhs: %s", err.Error())
		closePythonTestAssignmentLoopHeaders(w, headers)
		return
	}
	tmpRHS = g.maybeVariantUpcast(a.LHS.NodeSort(), a.RHS.NodeSort(), tmpRHS, "")
	w.linef("%s = %s;", finalLHS, tmpRHS)
	closePythonTestAssignmentLoopHeaders(w, headers)
}
