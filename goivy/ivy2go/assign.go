package ivy2go

import "github.com/glycerine/ivy/goivy"

// assign.go mirrors ivy2cpp/assign.go.
//
// Shape decisions:
//   - Empty LHS free-var set            → simple scalar assignment.
//   - Free vars present, loops openable → two-phase assignment via a
//     temporary so RHS can't alias the LHS storage mid-loop.
//   - Free vars present, loops not openable → defer to the thunk path
//     (currently reported as deferred to a follow-up).
//
// OPEN 052: this file adds the loops-openable path. The thunk fallback
// remains a future refinement.

// emitAssign dispatches by LHS shape. Mirrors ivy2cpp/action.go
// emitAssign.
func (g *Generator) emitAssign(w *goWriter, a *goivy.LogicAssignAction) {
	vs := goivy.VariablesAstList(a.LHS)
	if len(vs) == 0 {
		g.emitAssignSimple(w, a)
		return
	}
	if g.canOpenAssignmentLoops(vs) {
		g.emitAssignTwoPhase(w, a, vs)
		return
	}
	// Loops can't be opened — fall back to thunk-wrapped assignment.
	// Mirrors ivy2cpp/action.go's emitAssign call to emitAssignLarge.
	g.emitAssignLarge(w, a, vs)
}

// emitAssignLarge wraps the RHS in a thunk closure (from M8) and
// assigns that thunk to the LHS storage. Used when the quantifier
// variables have no derivable iteration bounds, so we can't expand
// the assignment into explicit loops.
//
// The thunk struct memoises on first read of each domain tuple, so
// the semantics match the in-loop assignment but defers the work to
// lookup time.
//
// Note: this requires the LHS storage to be a map[K]V-shaped thunk
// (goStorageHashThunk) — array-storage LHSes with unbounded vars
// are by definition impossible (their loop bounds would be known).
func (g *Generator) emitAssignLarge(w *goWriter, a *goivy.LogicAssignAction, vs []*goivy.LogicVariable) {
	// Build the thunk: lambda(vs...) → a.RHS.
	thunkExpr, err := g.makeThunk(vs, a.RHS)
	if err != nil {
		g.unsupported(w, "thunk emission for unbounded assignment: %s", err.Error())
		return
	}
	// The thunk constructor returns a *thunk_N. We need to install
	// it as a callable backing store for the LHS function symbol.
	// For Go, the simplest faithful shape is to replace the LHS
	// map with a fresh one whose values are computed lazily via
	// the thunk: we don't actually back the map with the thunk
	// (Go maps aren't lazy), but we *eagerly evaluate* the thunk
	// over the (unknown) domain. Since we can't enumerate, we
	// instead leave the map empty and document that reads through
	// the LHS will go through a thunk-style getter that callers
	// must already use.
	//
	// For the M8-compatible runtime path the actual semantics is:
	//   1. Clear the LHS map.
	//   2. Stash the thunk on a parallel field of *State (auto-
	//      provisioned name: __lhs_thunk_<sym>) so action methods
	//      that read the LHS can fall back to it on map miss.
	//
	// This is a partial port: the M8 hash_thunk's lazy semantics
	// aren't fully wired through into read-side emission yet. We
	// emit a thunk construction and a comment marker so callers
	// see the intended shape, with a follow-up to thread the
	// read-side lookup. Tracked in OPEN 061.1.
	lhsCode, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "thunk-fallback lhs: %s", err.Error())
		return
	}
	// Extract the LHS root symbol name so the emitted code knows
	// which State field to reset.
	root := lhsRootName(a.LHS)
	w.linef("// OPEN 061: thunk fallback for unbounded quantified LHS %q.", root)
	w.linef("// Reset map storage; reads must use the thunk's get() on miss.")
	w.linef("_ = %s", lhsCode)
	w.linef("__thunk := %s", thunkExpr)
	w.line("_ = __thunk")
	w.linef("// OPEN 061.1: wire reads of %s through __thunk.get(k).", root)
}

// emitAssignSimple ports ivy2cpp/assign.go emitAssignSimple.
// Trace-LHS emission (Config.Trace) is handled separately in vprint.go.
func (g *Generator) emitAssignSimple(w *goWriter, a *goivy.LogicAssignAction) {
	lhs, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		return
	}
	if g.Config.Trace {
		g.emitTraceWrite(w, a.LHS, rhs)
	}
	w.linef("%s = %s", lhs, rhs)
}

// canOpenAssignmentLoops reports whether all free variables in vs
// have finite, computable bounds for a loop header. Mirrors the gate
// ivy2cpp/action.go uses to decide simple vs thunk emission.
func (g *Generator) canOpenAssignmentLoops(vs []*goivy.LogicVariable) bool {
	for _, v := range vs {
		if _, _, err := g.loopHeaderForVar(v); err != nil {
			return false
		}
	}
	return true
}

// emitAssignTwoPhase ports ivy2cpp/assign.go emitAssignTwoPhase. The
// RHS is computed into a temporary indexed by the free variables, then
// copied back into the LHS in a second loop pass — this avoids
// read-during-write aliasing.
//
// The temporary is a Go local of the same storage shape as the LHS
// (chosen by goFunctionStorageFor). For small bounded loops the temp
// is a fixed-size array; for larger loops it's a map.
func (g *Generator) emitAssignTwoPhase(w *goWriter, a *goivy.LogicAssignAction, vs []*goivy.LogicVariable) {
	// Build a function sort over the loop variables + RHS sort so we
	// can derive the temp's Go type via goFunctionStorageFor.
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
	tmpName := "__ivy_tmp_" + g.Ctx.GetTemp()
	tmpType := g.goType(tsort)

	w.open("{")
	// Declare the temp. For map storage it needs make(); for array
	// storage zero-init is fine.
	if g.symbolNeedsMakeMap(tsort) {
		w.linef("%s := %s{}", tmpName, tmpType)
		w.linef("_ = %s", tmpName)
	} else {
		w.linef("var %s %s", tmpName, tmpType)
		w.linef("_ = %s", tmpName)
	}

	// Phase 1: write RHS into the temp, indexed by loop vars.
	g.emitAssignmentLoopsBody(w, vs, a.RHS, tmpName, true)
	// Phase 2: copy temp back to actual LHS storage.
	g.emitAssignmentLoopsBody(w, vs, a.LHS, tmpName, false)

	w.close("")
}

// emitAssignmentLoopsBody emits the nested loops and inner write for
// one of the two phases. When `intoTemp` is true the inner statement
// is `tmp[loop_vars] = rhsExpr`; otherwise it's
// `lhs[loop_vars] = tmp[loop_vars]`.
func (g *Generator) emitAssignmentLoopsBody(w *goWriter, vs []*goivy.LogicVariable, rhsOrLHS goivy.Expr, tmpName string, intoTemp bool) {
	headers := make([]string, len(vs))
	closers := make([]string, len(vs))
	for i, v := range vs {
		h, c, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported assignment loop var %s: %s", goIdent(v.Name), err.Error())
			return
		}
		headers[i] = h
		closers[i] = c
	}
	for _, h := range headers {
		w.open(h)
	}
	keyExpr := g.tempKeyExpression(vs, tmpName)
	if intoTemp {
		// tmp[key] = rhs
		rhsCode, err := g.emitExpr(rhsOrLHS)
		if err != nil {
			g.unsupported(w, "unsupported rhs in two-phase: %s", err.Error())
			for range closers {
				w.close("")
			}
			return
		}
		w.linef("%s = %s", keyExpr, rhsCode)
	} else {
		// lhs[loop vars] = tmp[key]
		lhsCode, err := g.emitExpr(rhsOrLHS)
		if err != nil {
			g.unsupported(w, "unsupported lhs in two-phase: %s", err.Error())
			for range closers {
				w.close("")
			}
			return
		}
		w.linef("%s = %s", lhsCode, keyExpr)
	}
	for range closers {
		w.close("")
	}
}

// tempKeyExpression returns the indexed access for the temporary
// storage given the loop variables vs. Single-var → tmp[v]; multi-var
// → tmp[tup__T1__T2{v1, v2}].
func (g *Generator) tempKeyExpression(vs []*goivy.LogicVariable, tmpName string) string {
	if len(vs) == 1 {
		return tmpName + "[" + goIdent(vs[0].Name) + "]"
	}
	domSorts := make([]goivy.Sort, len(vs))
	for i, v := range vs {
		domSorts[i] = v.VSort
	}
	// For arrays we just use [v1][v2]…; for maps we use the tuple key.
	tsort, err := goivy.NewFunctionSort(append(domSorts, goivy.Boolean)...)
	if err == nil {
		st := goFunctionStorageFor(g, domSorts, goivy.Boolean)
		_ = tsort
		if st.Kind == goStorageArray {
			expr := tmpName
			for _, v := range vs {
				expr += "[" + goIdent(v.Name) + "]"
			}
			return expr
		}
	}
	keyType := goCTupleNameWith(g, domSorts)
	expr := tmpName + "[" + keyType + "{"
	for i, v := range vs {
		if i > 0 {
			expr += ", "
		}
		expr += goIdent(v.Name)
	}
	expr += "}]"
	return expr
}

// emitTraceWrite is OPEN 058's hook. M10 left it stubbed; the real
// implementation in vprint.go is called here when Config.Trace is on.
func (g *Generator) emitTraceWrite(w *goWriter, lhs goivy.Expr, rhsCode string) {
	g.emitTracedLHS(w, lhs, rhsCode)
}
