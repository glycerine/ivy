package ivy2go

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// action_gen.go mirrors ivy2cpp/action_gen.go. For target=test/gen,
// the C++ generator emits one action_gen_<name> class per action that
// uses Z3 to synthesise inputs satisfying the action's reverse-image
// precondition. ivy2go's equivalent emits one actionGen<Name> struct
// per action, holding a *goivy.Solver wired through the goivy facade.
//
// M9 ships a structural skeleton: each generator constructs a Solver
// (via newIvySolver from z3.go), calls pushStateIntoSolver (from
// solver_emit.go) to assert the pre-state, then chooses inputs via
// ivyChoose as the M9.0 fallback. M9.1+ will replace the input
// selection with Solver.GetSmallModel.

// emitActionGenStructs walks the module's actions and emits one
// actionGen<Name> struct per public action. Only runs for target=test
// or target=gen.
//
// Per action we emit three helpers:
//
//   - buildPrecondition_<Name>   — reified Pre Clauses for the solver
//     (OPEN 055.3)
//   - wouldFail_<Name>           — plain-Go Pre evaluator used as a
//     pre-firing check (so randomized inputs don't crash an assert)
//   - actionGen_<Name>           — the per-action generator struct
func (g *Generator) emitActionGenStructs(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	if g.Ctx != nil {
		g.Ctx.AddImport("actions", g.Config.GoivyImportPath, "")
	}
	for _, name := range g.actionGenNames() {
		g.emitPreconditionForAction(w, name)
		g.emitWouldFailHelper(w, name)
		g.emitOneActionGenStruct(w, name)
	}
}

// emitWouldFailHelper writes `wouldFail_<Name>(state *State, v0 T0, …) bool`.
// Returns true when the action's Pre clauses would hold (i.e. the
// action would fail an assert) given the supplied input values, so
// actionGen.Generate can skip the call. Mirrors the precondition
// gating ivy2cpp's test harness does via the Z3 model-feasibility
// check.
//
// We emit Pre.Fmlas as a plain-Go disjunction (any one fmla being
// true means failure), with references to formal params rewritten
// to the corresponding `v<i>` local and references to state symbols
// routed through `s.<Exported>` via the existing emitExpr machinery.
func (g *Generator) emitWouldFailHelper(w *goWriter, name string) {
	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()

	ctx := &goivy.UpdateContext{
		Domain: g.Mod,
		PVars:  map[string]bool{},
		ActCfg: g.Mod.Cfg.ActCfg,
	}
	var update *goivy.Update
	func() {
		defer func() { _ = recover() }()
		update = goivy.GetUpdate(act, ctx)
	}()

	helperName := "wouldFail_" + goExportedName(name)
	sig := "state *" + g.StateTypeName
	for i, p := range params {
		if p == nil {
			continue
		}
		sig += fmt.Sprintf(", v%d %s", i, g.goType(p.CSort))
	}
	w.linef("// %s evaluates action %q's Pre clauses with the picked", helperName, name)
	w.linef("// inputs. Returns true iff the action would fail an")
	w.linef("// ivyAssert/ivyAssume so callers can skip it.")
	w.linef("func %s(%s) bool {", helperName, sig)
	w.line("\t_ = state")
	for i := range params {
		w.linef("\t_ = v%d", i)
	}
	if update == nil || update.Pre == nil || update.Pre.IsFalse() {
		// Pre trivially-false → action never fails on this input.
		w.line("\treturn false")
		w.line("}")
		w.blank()
		return
	}
	// Lower each Pre fmla into a Go boolean expression. The body
	// runs inside a function whose receiver is `s *State`, so we
	// alias the helper's `state` param to that convention by
	// emitting wrappers below.
	//
	// Strategy: walk each Pre fmla, rewrite formal-param refs to
	// the corresponding v<i> local, then emit the Go expression.
	// State-symbol refs route through `s.X` per emitExpr — but
	// our helper has `state *State`, not `s *State`. We fix that
	// by aliasing `s := state` once at function entry.
	w.line("\ts := state")
	w.line("\t_ = s")
	// Inline Pre.Defs into every Pre.Fmla so synthetic temporaries
	// (e.g. `__ts0_a = left_player.ball`) become resolvable
	// references to state symbols / formals. Without this, fmlas
	// that mention the temporary get marked unreifiable and the
	// action is skipped, which masks the user's actual Pre and
	// breaks the test trace (cf. ivy2cpp's pingpong output that
	// fires `left_player.hit`).
	defSubs := buildDefSubstitutions(update.Pre.Defs)
	parts := []string{}
	for _, fmla := range update.Pre.Fmlas {
		inlined := applyDefSubstitutions(fmla, defSubs)
		expr, ok := g.emitPreFmlaAsGoExpr(inlined, params)
		if !ok {
			// Unreifiable fmla — conservatively assume failure so
			// the action is skipped (safer than firing and
			// panicking).
			w.linef("\t// unreifiable Pre fmla — conservatively skipping action: %s",
				strings.ReplaceAll(fmla.String(), "\n", " "))
			parts = append(parts, "true")
			continue
		}
		parts = append(parts, "("+expr+")")
	}
	if len(parts) == 0 {
		w.line("\treturn false")
	} else {
		// Pre.Fmlas semantics: a goivy *Clauses is a CONJUNCTION
		// of its fmlas. The action fails iff EVERY fmla holds.
		// Joining with `&&` (not `||`) matches that.
		w.linef("\treturn %s", strings.Join(parts, " && "))
	}
	w.line("}")
	w.blank()
}

// emitPreFmlaAsGoExpr emits a goivy Expr as a plain Go boolean
// expression by first substituting formal-param references with
// v<i>-named Consts (so they lower via emitExpr's goIdent path
// to the v<i> locals in scope), then calling emitExpr.
//
// Returns ("", false) when the Pre fmla mentions any symbol we
// can't resolve at the Go level — state symbols, formal params,
// numerals, enum constants, and definitions are OK; goivy-internal
// temporaries (e.g. `__ts0_a` from the update-analysis pass) are
// not. The caller treats this as "conservatively assume the action
// would fail" so the random scheduler skips it.
func (g *Generator) emitPreFmlaAsGoExpr(fmla goivy.Expr, params []*goivy.Const) (string, bool) {
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range params {
		if p == nil {
			continue
		}
		// Substitute by key (matches the formal's own *Const ID).
		subs[goivy.Key(p)] = &goivy.Const{Name: fmt.Sprintf("v%d", i), CSort: p.CSort}
	}
	rewritten := fmla
	if len(subs) > 0 {
		if r, err := goivy.Substitute(fmla, subs); err == nil {
			rewritten = r
		}
	}
	// Also handle name-based references (__fml:b, fml:b, b) that
	// don't share *Const identity with the formal — walk the
	// rewritten tree and rebuild Consts when their name matches.
	rewritten = rewriteFormalRefsByName(rewritten, params)
	if g.preHasUnresolvableRef(rewritten) {
		return "", false
	}
	code, err := g.emitExpr(rewritten)
	if err != nil {
		return "", false
	}
	return code, true
}

// buildDefSubstitutions converts a slice of Pre.Defs (each
// `IvyDefinition` defining a temporary or new-state symbol in
// terms of the old state) into a map keyed by the LHS symbol name.
//
// Used by applyDefSubstitutions to inline synthetic temporaries
// like `__ts0_a` before reifying a Pre fmla as Go code.
func buildDefSubstitutions(defs []*goivy.IvyDefinition) map[string]goivy.Expr {
	out := make(map[string]goivy.Expr, len(defs))
	for _, d := range defs {
		if d == nil {
			continue
		}
		lhs, ok := d.Lhs.(*goivy.Const)
		if !ok {
			continue
		}
		out[lhs.Name] = d.Rhs
	}
	return out
}

// applyDefSubstitutions walks e and replaces any *goivy.Const whose
// name is a key in subs with the corresponding RHS Expr. Used to
// inline Pre.Defs so synthetic temporaries don't appear in the
// emitted Go expression.
//
// Iterates until a fixed point in case one def references another.
// Cap at 8 passes to defend against pathological circular defs.
func applyDefSubstitutions(e goivy.Expr, subs map[string]goivy.Expr) goivy.Expr {
	if len(subs) == 0 {
		return e
	}
	for pass := 0; pass < 8; pass++ {
		next, changed := applyDefSubstitutionsOnce(e, subs)
		if !changed {
			return e
		}
		e = next
	}
	return e
}

func applyDefSubstitutionsOnce(e goivy.Expr, subs map[string]goivy.Expr) (goivy.Expr, bool) {
	if e == nil {
		return nil, false
	}
	switch n := e.(type) {
	case *goivy.Const:
		if rhs, ok := subs[n.Name]; ok {
			return rhs, true
		}
		return n, false
	case *goivy.LogicNot:
		body, ch := applyDefSubstitutionsOnce(n.Body, subs)
		if !ch {
			return n, false
		}
		return &goivy.LogicNot{Body: body}, true
	case *goivy.LogicAnd:
		anyCh := false
		terms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			nt, ch := applyDefSubstitutionsOnce(t, subs)
			if ch {
				anyCh = true
			}
			terms[i] = nt
		}
		if !anyCh {
			return n, false
		}
		return &goivy.LogicAnd{Terms: terms}, true
	case *goivy.LogicOr:
		anyCh := false
		terms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			nt, ch := applyDefSubstitutionsOnce(t, subs)
			if ch {
				anyCh = true
			}
			terms[i] = nt
		}
		if !anyCh {
			return n, false
		}
		return &goivy.LogicOr{Terms: terms}, true
	case *goivy.LogicImplies:
		l, lch := applyDefSubstitutionsOnce(n.T1, subs)
		r, rch := applyDefSubstitutionsOnce(n.T2, subs)
		if !lch && !rch {
			return n, false
		}
		return &goivy.LogicImplies{T1: l, T2: r}, true
	case *goivy.LogicIff:
		l, lch := applyDefSubstitutionsOnce(n.T1, subs)
		r, rch := applyDefSubstitutionsOnce(n.T2, subs)
		if !lch && !rch {
			return n, false
		}
		return &goivy.LogicIff{T1: l, T2: r}, true
	case *goivy.Eq:
		l, lch := applyDefSubstitutionsOnce(n.T1, subs)
		r, rch := applyDefSubstitutionsOnce(n.T2, subs)
		if !lch && !rch {
			return n, false
		}
		return &goivy.Eq{T1: l, T2: r}, true
	case *goivy.Apply:
		anyCh := false
		newTerms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			nt, ch := applyDefSubstitutionsOnce(t, subs)
			if ch {
				anyCh = true
			}
			newTerms[i] = nt
		}
		if !anyCh {
			return n, false
		}
		if a, err := goivy.NewApply(n.Func, newTerms...); err == nil {
			return a, true
		}
		return n, false
	}
	return e, false
}

// preHasUnresolvableRef walks e looking for Const references that
// can't be lowered to a sensible Go expression at runtime. Anything
// outside (state symbols, enum members, definitions, numerals, the
// known v<i> locals, plain bool literals) is flagged.
func (g *Generator) preHasUnresolvableRef(e goivy.Expr) bool {
	if e == nil {
		return false
	}
	switch n := e.(type) {
	case *goivy.Const:
		// Plain bool literals.
		if n.CSort == goivy.Boolean && (n.Name == "true" || n.Name == "false") {
			return false
		}
		// Numerals.
		if goivy.IsNumeral(n) {
			return false
		}
		// State symbols.
		if g.isStateSymbolName(n.Name) {
			return false
		}
		// Enum members.
		if g.isEnumConstantName(n.Name) {
			return false
		}
		// Definitions (parameterless).
		if g.isDefinitionName(n.Name) {
			return false
		}
		// v<i> locals minted by emitPreFmlaAsGoExpr.
		if strings.HasPrefix(n.Name, "v") {
			rest := n.Name[1:]
			allDigits := len(rest) > 0
			for _, c := range rest {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return false
			}
		}
		// Unknown — likely a goivy-internal temporary.
		return true
	}
	for _, ch := range e.Children() {
		if g.preHasUnresolvableRef(ch) {
			return true
		}
	}
	if a, ok := e.(*goivy.Apply); ok {
		if g.preHasUnresolvableRef(a.Func) {
			return true
		}
	}
	return false
}

// rewriteFormalRefsByName traverses e and replaces any *goivy.Const
// whose name matches a formal param (with the goivy naming
// conventions `b`, `fml:b`, `__fml:b`) by a fresh Const named
// `v<i>`. Used by emitPreFmlaAsGoExpr to catch references that
// goivy.Substitute missed due to non-identity-matching.
func rewriteFormalRefsByName(e goivy.Expr, params []*goivy.Const) goivy.Expr {
	if e == nil {
		return nil
	}
	switch n := e.(type) {
	case *goivy.Const:
		for i, p := range params {
			if p == nil {
				continue
			}
			short := strings.TrimPrefix(strings.TrimPrefix(p.Name, "__fml:"), "fml:")
			if n.Name == p.Name ||
				n.Name == short ||
				n.Name == "fml:"+short ||
				n.Name == "__fml:"+short {
				return &goivy.Const{Name: fmt.Sprintf("v%d", i), CSort: n.CSort}
			}
		}
		return n
	case *goivy.LogicNot:
		return &goivy.LogicNot{Body: rewriteFormalRefsByName(n.Body, params)}
	case *goivy.LogicAnd:
		terms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = rewriteFormalRefsByName(t, params)
		}
		return &goivy.LogicAnd{Terms: terms}
	case *goivy.LogicOr:
		terms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = rewriteFormalRefsByName(t, params)
		}
		return &goivy.LogicOr{Terms: terms}
	case *goivy.LogicImplies:
		return &goivy.LogicImplies{T1: rewriteFormalRefsByName(n.T1, params), T2: rewriteFormalRefsByName(n.T2, params)}
	case *goivy.LogicIff:
		return &goivy.LogicIff{T1: rewriteFormalRefsByName(n.T1, params), T2: rewriteFormalRefsByName(n.T2, params)}
	case *goivy.Eq:
		return &goivy.Eq{T1: rewriteFormalRefsByName(n.T1, params), T2: rewriteFormalRefsByName(n.T2, params)}
	case *goivy.Apply:
		// Walk args; leave the function symbol alone (its name is
		// the destructor/relation name, not a formal).
		newTerms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = rewriteFormalRefsByName(t, params)
		}
		if a, err := goivy.NewApply(n.Func, newTerms...); err == nil {
			return a
		}
		return n
	case *goivy.ForAll:
		return &goivy.ForAll{Variables: n.Variables, Body: rewriteFormalRefsByName(n.Body, params)}
	case *goivy.LogicExists:
		return &goivy.LogicExists{Variables: n.Variables, Body: rewriteFormalRefsByName(n.Body, params)}
	}
	return e
}

// emitPreconditionForAction emits a `buildPrecondition_<Name>` helper
// per action. The helper conjoins stateFactsAsClauses(state) with
// the runtime-reified Pre.Fmlas AND per-field destructor equalities
// when any param is struct-typed (OPEN 055.7).
//
// Signature includes the receiver symbol per param plus, for struct
// params, one Const per scalar field so the solver can synthesise
// each independently.
func (g *Generator) emitPreconditionForAction(w *goWriter, name string) {
	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()

	ctx := &goivy.UpdateContext{
		Domain: g.Mod,
		PVars:  map[string]bool{},
		ActCfg: g.Mod.Cfg.ActCfg,
	}
	var update *goivy.Update
	func() {
		defer func() { _ = recover() }()
		update = goivy.GetUpdate(act, ctx)
	}()

	helperName := "buildPrecondition_" + goExportedName(name)
	paramArgs := preconditionSignatureArgs(g, params)
	w.linef("// %s reifies action %q's Pre clauses at runtime.", helperName, name)
	w.linef("// Generated from goivy.GetUpdate(action, ctx).Pre at emit time.")
	w.linef("func %s(state *%s%s) *goivy.Clauses {", helperName, g.StateTypeName, optComma(len(paramArgs))+strings.Join(paramArgs, ", "))
	w.linef("\tbase := stateFactsAsClauses(state)")
	w.line("\tvar extra []goivy.Expr")

	// OPEN 055.7: for each struct-typed param, add destructor
	// equalities (destructor(p) = p_field) so the solver binds the
	// field values to per-field input symbols.
	for i, p := range params {
		if p == nil {
			continue
		}
		recName, isRecord := g.destructorStructName(p.CSort)
		if !isRecord {
			continue
		}
		for _, f := range g.destructorScalarFields(recName) {
			fnSortCode, ok := g.reifySortAsGoCode(f.DestructorC.CSort)
			if !ok {
				continue
			}
			w.linef("\textra = append(extra, &goivy.Eq{T1: mustApply(goivy.NewConst(%q, %s), __in%d), T2: __in%d_%s})",
				f.FullName, fnSortCode, i, i, goIdent(f.Name))
		}
		g.requireMustHelpers()
	}

	if update != nil && update.Pre != nil && !update.Pre.IsFalse() {
		for _, fmla := range update.Pre.Fmlas {
			code, ok := g.reifyExprAsGoCode(fmla, params)
			if !ok {
				w.linef("\t// OPEN unreifiable Pre fmla: %s", strings.ReplaceAll(fmla.String(), "\n", " "))
				continue
			}
			w.linef("\textra = append(extra, %s)", code)
		}
	}
	w.line("\treturn conjClauses(base, extra...)")
	w.line("}")
	w.blank()
}

// emitStructInputAssembly emits the per-field pick + struct-literal
// assembly for a destructor-record param (OPEN 055.7).
func (g *Generator) emitStructInputAssembly(w *goWriter, i int, typeName, recName string) {
	fields := g.destructorScalarFields(recName)
	for _, f := range fields {
		fcard := goSortCard(g, f.Sort)
		switch ft := g.goType(f.Sort); {
		case ft == "bool":
			w.linef("v%d_%s := pickBoolOrChoose(g.sol, modelResult, __in%d_%s)",
				i, goIdent(f.Name), i, goIdent(f.Name))
		case fcard > 0 && goIsAnyIntegerType(g, f.Sort):
			w.linef("v%d_%s := %s(pickUintOrChoose(g.sol, modelResult, __in%d_%s, %d))",
				i, goIdent(f.Name), ft, i, goIdent(f.Name), fcard)
		default:
			w.linef("var v%d_%s %s", i, goIdent(f.Name), ft)
			w.linef("_ = v%d_%s", i, goIdent(f.Name))
		}
	}
	assign := make([]string, 0, len(fields))
	for _, f := range fields {
		assign = append(assign, fmt.Sprintf("%s: v%d_%s", goExportedName(f.Name), i, goIdent(f.Name)))
	}
	w.linef("v%d := %s{%s}", i, typeName, strings.Join(assign, ", "))
}

// emitVariantInputAssembly emits the tag-pick + switch-by-tag + per-
// leaf constructor invocation for a variant-super-typed param
// (OPEN 055.8).
//
// Shape:
//
//	v0_tag := ivyChoose(<numLeaves>)
//	var v0 <SuperType>
//	switch v0_tag {
//	case 0: v0 = NewSuperLeafA()                  // plain leaf
//	case 1:                                         // destructor-backed leaf
//	    v0_leafB_f := pickBoolOrChoose(...)
//	    v0 = NewSuperLeafB(LeafB{F: v0_leafB_f})
//	}
//
// The solver doesn't natively know about tag selection (Ivy's
// supertype is just an UninterpretedSort) so we use ivyChoose for
// the tag. Per-leaf field synthesis still flows through the solver
// when the leaf is destructor-backed.
func (g *Generator) emitVariantInputAssembly(w *goWriter, i int, typeName, superName string) {
	leaves := g.variantLeaves(superName)
	if len(leaves) == 0 {
		w.linef("var v%d %s", i, typeName)
		w.linef("_ = v%d", i)
		return
	}
	w.linef("v%d_tag := ivyChoose(%d)", i, len(leaves))
	w.linef("var v%d %s", i, typeName)
	w.linef("switch v%d_tag {", i)
	for tag, leaf := range leaves {
		w.linef("case %d:", tag)
		ctor := "New" + goExportedName(superName) + goExportedName(leaf.Name)
		if leaf.IsPlain {
			w.linef("\tv%d = %s()", i, ctor)
			continue
		}
		// Destructor-backed leaf — pick each field, build the
		// leaf struct, pass to the constructor.
		leafType := goExportedName(leaf.Name)
		for _, f := range leaf.Fields {
			fcard := goSortCard(g, f.Sort)
			switch ft := g.goType(f.Sort); {
			case ft == "bool":
				w.linef("\tv%d_%s_%s := pickBoolOrChoose(g.sol, modelResult, __in%d_%s_%s)",
					i, goIdent(leaf.Name), goIdent(f.Name),
					i, goIdent(leaf.Name), goIdent(f.Name))
			case fcard > 0 && goIsAnyIntegerType(g, f.Sort):
				w.linef("\tv%d_%s_%s := %s(pickUintOrChoose(g.sol, modelResult, __in%d_%s_%s, %d))",
					i, goIdent(leaf.Name), goIdent(f.Name),
					ft, i, goIdent(leaf.Name), goIdent(f.Name), fcard)
			default:
				w.linef("\tvar v%d_%s_%s %s", i, goIdent(leaf.Name), goIdent(f.Name), ft)
				w.linef("\t_ = v%d_%s_%s", i, goIdent(leaf.Name), goIdent(f.Name))
			}
		}
		assign := make([]string, 0, len(leaf.Fields))
		for _, f := range leaf.Fields {
			assign = append(assign, fmt.Sprintf("%s: v%d_%s_%s",
				goExportedName(f.Name), i, goIdent(leaf.Name), goIdent(f.Name)))
		}
		w.linef("\tv%d = %s(%s{%s})", i, ctor, leafType, strings.Join(assign, ", "))
	}
	w.line("}")
}

// preconditionSignatureArgs returns the formal-param list of
// buildPrecondition_<Name>: one `__in<i> *goivy.Const` per scalar
// param plus per-field symbols for struct (OPEN 055.7) and variant
// (OPEN 055.8) params.
//
// For variant params each destructor-backed leaf contributes one
// `__in<i>_<leaf>_<field>` Const per scalar field — the receiver
// symbol __in<i> is always present too.
func preconditionSignatureArgs(g *Generator, params []*goivy.Const) []string {
	out := make([]string, 0, len(params))
	for i, p := range params {
		if p == nil {
			continue
		}
		out = append(out, fmt.Sprintf("__in%d *goivy.Const", i))
		if recName, ok := g.destructorStructName(p.CSort); ok {
			for _, f := range g.destructorScalarFields(recName) {
				out = append(out, fmt.Sprintf("__in%d_%s *goivy.Const", i, goIdent(f.Name)))
			}
			continue
		}
		if superName, ok := g.variantSuperName(p.CSort); ok {
			for _, leaf := range g.variantLeaves(superName) {
				for _, f := range leaf.Fields {
					out = append(out, fmt.Sprintf("__in%d_%s_%s *goivy.Const",
						i, goIdent(leaf.Name), goIdent(f.Name)))
				}
			}
		}
	}
	return out
}

// preconditionCallArgs returns the matching call-site identifiers for
// preconditionSignatureArgs.
func preconditionCallArgs(g *Generator, params []*goivy.Const) []string {
	out := make([]string, 0, len(params))
	for i, p := range params {
		if p == nil {
			continue
		}
		out = append(out, fmt.Sprintf("__in%d", i))
		if recName, ok := g.destructorStructName(p.CSort); ok {
			for _, f := range g.destructorScalarFields(recName) {
				out = append(out, fmt.Sprintf("__in%d_%s", i, goIdent(f.Name)))
			}
			continue
		}
		if superName, ok := g.variantSuperName(p.CSort); ok {
			for _, leaf := range g.variantLeaves(superName) {
				for _, f := range leaf.Fields {
					out = append(out, fmt.Sprintf("__in%d_%s_%s",
						i, goIdent(leaf.Name), goIdent(f.Name)))
				}
			}
		}
	}
	return out
}

// optComma returns ", " when n > 0, else "". Used to splice the
// state arg comfortably with the inputs in a function signature.
func optComma(n int) string {
	if n > 0 {
		return ", "
	}
	return ""
}

// reifyExprAsGoCode walks an Expr and returns a Go source string that,
// when evaluated at runtime, constructs the same Expr. The params
// slice lets us rewrite formal-param references (`__fml:<name>` or
// bare param name) to the input symbol __in<i> in the emitted
// helper's scope.
//
// Returns (code, true) on success; (_, false) when the Expr shape
// isn't supported (caller skips the fmla and emits a comment).
func (g *Generator) reifyExprAsGoCode(e goivy.Expr, params []*goivy.Const) (string, bool) {
	if e == nil {
		return "nil", true
	}
	switch n := e.(type) {
	case *goivy.Const:
		// Param reference: rewrite to the runtime input symbol.
		// goivy uses several name conventions for the same formal
		// parameter ("b", "fml:b", "__fml:b"). Match any of them.
		for i, p := range params {
			if p == nil {
				continue
			}
			short := strings.TrimPrefix(strings.TrimPrefix(p.Name, "__fml:"), "fml:")
			if n.Name == p.Name ||
				n.Name == short ||
				n.Name == "fml:"+short ||
				n.Name == "__fml:"+short {
				return fmt.Sprintf("__in%d", i), true
			}
		}
		// Plain constant: build a fresh goivy.Const with the
		// sort reified.
		sortCode, ok := g.reifySortAsGoCode(n.CSort)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("goivy.NewConst(%q, %s)", n.Name, sortCode), true
	case *goivy.LogicNot:
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicNot{Body: %s}", body), true
	case *goivy.LogicAnd:
		parts := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			code, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		return fmt.Sprintf("&goivy.LogicAnd{Terms: []goivy.Expr{%s}}", strings.Join(parts, ", ")), true
	case *goivy.LogicOr:
		parts := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			code, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		return fmt.Sprintf("&goivy.LogicOr{Terms: []goivy.Expr{%s}}", strings.Join(parts, ", ")), true
	case *goivy.LogicImplies:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicImplies{T1: %s, T2: %s}", l, r), true
	case *goivy.LogicIff:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicIff{T1: %s, T2: %s}", l, r), true
	case *goivy.Eq:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", l, r), true

	case *goivy.LogicVariable:
		sortCode, ok := g.reifySortAsGoCode(n.VSort)
		if !ok {
			return "", false
		}
		// LogicVariable.NewVariable validates uppercase first
		// letter; rely on the mustNewVariable runtime helper for a
		// panic-on-failure facade.
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewVariable(%q, %s)", n.Name, sortCode), true

	case *goivy.Apply:
		fnCode, ok := g.reifyExprAsGoCode(n.Func, params)
		if !ok {
			return "", false
		}
		argCodes := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			c, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			argCodes = append(argCodes, c)
		}
		g.requireMustHelpers()
		if len(argCodes) == 0 {
			return fmt.Sprintf("mustApply(%s)", fnCode), true
		}
		return fmt.Sprintf("mustApply(%s, %s)", fnCode, strings.Join(argCodes, ", ")), true

	case *goivy.ForAll:
		varCodes, sortOK := g.reifyVariableSlice(n.Variables)
		if !sortOK {
			return "", false
		}
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.ForAll{Variables: %s, Body: %s}", varCodes, body), true

	case *goivy.LogicExists:
		varCodes, sortOK := g.reifyVariableSlice(n.Variables)
		if !sortOK {
			return "", false
		}
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicExists{Variables: %s, Body: %s}", varCodes, body), true

	case *goivy.LogicIte:
		c, ok := g.reifyExprAsGoCode(n.Cond, params)
		if !ok {
			return "", false
		}
		t, ok := g.reifyExprAsGoCode(n.Then, params)
		if !ok {
			return "", false
		}
		e, ok := g.reifyExprAsGoCode(n.Else, params)
		if !ok {
			return "", false
		}
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewIte(%s, %s, %s)", c, t, e), true

	default:
		return "", false
	}
}

// reifyVariableSlice helps the ForAll/LogicExists cases by turning a
// []*goivy.LogicVariable into a Go composite-literal string.
// Returns (code, true) when every variable's sort is reifiable.
func (g *Generator) reifyVariableSlice(vars []*goivy.LogicVariable) (string, bool) {
	if len(vars) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(vars))
	for _, v := range vars {
		if v == nil {
			parts = append(parts, "nil")
			continue
		}
		sortCode, ok := g.reifySortAsGoCode(v.VSort)
		if !ok {
			return "", false
		}
		g.requireMustHelpers()
		parts = append(parts, fmt.Sprintf("mustNewVariable(%q, %s)", v.Name, sortCode))
	}
	return fmt.Sprintf("[]*goivy.LogicVariable{%s}", strings.Join(parts, ", ")), true
}

// reifySortAsGoCode returns a Go source expression that constructs s.
// Supports BooleanSort, UninterpretedSort, LogicEnumeratedSort,
// RangeSort (numeral bounds only), and LogicFunctionSort. Anything
// else returns false so the caller falls back.
func (g *Generator) reifySortAsGoCode(s goivy.Sort) (string, bool) {
	switch t := s.(type) {
	case *goivy.BooleanSort:
		return "goivy.Boolean", true
	case *goivy.UninterpretedSort:
		return fmt.Sprintf("&goivy.UninterpretedSort{Name: %q}", t.Name), true
	case *goivy.LogicEnumeratedSort:
		ext := make([]string, len(t.Extension))
		for i, v := range t.Extension {
			ext[i] = fmt.Sprintf("%q", v)
		}
		return fmt.Sprintf("&goivy.LogicEnumeratedSort{Name: %q, Extension: []string{%s}}",
			t.Name, strings.Join(ext, ", ")), true
	case *goivy.RangeSort:
		lbCode, ok := g.reifyBoundAsGoCode(t.Lb)
		if !ok {
			return "", false
		}
		ubCode, ok := g.reifyBoundAsGoCode(t.Ub)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.RangeSort{Name: %q, Lb: %s, Ub: %s}", t.Name, lbCode, ubCode), true
	case *goivy.LogicFunctionSort:
		sorts := t.Sorts
		parts := make([]string, 0, len(sorts))
		for _, sub := range sorts {
			code, ok := g.reifySortAsGoCode(sub)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewFunctionSort(%s)", strings.Join(parts, ", ")), true
	}
	return "", false
}

// reifyBoundAsGoCode reifies a NumeralOrCompiledBound. NumeralBound
// becomes a direct struct literal; CompiledBound recurses into the
// inner Expr via reifyExprAsGoCode (with no param context — bounds
// don't capture action formals).
func (g *Generator) reifyBoundAsGoCode(b goivy.NumeralOrCompiledBound) (string, bool) {
	switch v := b.(type) {
	case goivy.NumeralBound:
		sortCode := "nil"
		if v.Sort != nil {
			if c, ok := g.reifySortAsGoCode(v.Sort); ok {
				sortCode = c
			}
		}
		return fmt.Sprintf("goivy.NumeralBound{Value: %q, Sort: %s}", v.Value, sortCode), true
	case goivy.CompiledBound:
		exprCode, ok := g.reifyExprAsGoCode(v.Expr, nil)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("goivy.CompiledBound{Expr: %s}", exprCode), true
	}
	return "", false
}

// requireMustHelpers flags the runtime to emit the must* facades
// (mustApply / mustNewVariable / mustNewIte / mustNewFunctionSort).
// All four share a single emission gate; needing any one pulls them
// all in to keep the dependency graph trivial.
func (g *Generator) requireMustHelpers() {
	if g != nil && g.Ctx != nil {
		g.Ctx.OnceGlobals["__need_musthelpers"] = true
	}
}

// actionGenNames returns the actions eligible for action-gen
// emission, in deterministic order. Mirrors ivy2cpp's
// publicActionNamesSorted: an action gets a generator iff the
// module's PublicActions map flags it. Falls back to the prefix
// filter (no ext:/imp:/ivy:/__ names) when PublicActions is empty —
// e.g. when running outside an isolate context.
func (g *Generator) actionGenNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return nil
	}
	names := make([]string, 0)
	if g.Mod.PublicActions != nil && g.Mod.PublicActions.Len() > 0 {
		for name := range g.Mod.PublicActions.All() {
			// PublicActions is authoritative, but skip names that
			// the action lookup can't resolve (rare). Also keep
			// the prefix filter as a safety net.
			if hasPrefixAny(name, "imp__", "__") {
				continue
			}
			if _, ok := g.Mod.Actions.Get2(name); !ok {
				continue
			}
			names = append(names, name)
		}
	} else {
		for name := range g.Mod.Actions.All() {
			if hasPrefixAny(name, "ext:", "imp:", "ivy:", "imp__", "__") {
				continue
			}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// emitOneActionGenStruct emits a single actionGen<Name> struct + its
// constructor and Generate method.
//
// OPEN 055.1: the Generate method now runs a real solver round-trip:
//
//  1. Build a fresh `*goivy.Const` for each formal param.
//  2. Build a trivial `*goivy.Clauses` (currently asserting `true`
//     since the reverse-image precondition derivation is the next
//     phase of work).
//  3. Call `sol.GetModelClauses(clauses)` to obtain a Z3 model.
//  4. For each input, extract the model value via
//     `ModelResult.Eval` + per-sort parsing (parseZ3Bool /
//     parseZ3Uint64). Fall back to `ivyChoose` if extraction fails.
//  5. Apply the action.
//
// This is genuine solver-driven gen even though the precondition is
// still trivial — adding real preconditions only requires changing
// step 2, with the rest of the wiring already in place.
func (g *Generator) emitOneActionGenStruct(w *goWriter, name string) {
	structName := "actionGen_" + goExportedName(name)
	w.linef("// %s drives solver-backed input synthesis for the %s action.", structName, name)
	w.open(fmt.Sprintf("type %s struct {", structName))
	w.line("sol *goivy.Solver")
	w.close("")
	w.blank()

	w.open(fmt.Sprintf("func new%s(state *%s) *%s {", structName, g.StateTypeName, structName))
	w.linef("g := &%s{sol: newIvySolver()}", structName)
	w.line("_ = pushStateIntoSolver(state, g.sol)")
	w.line("return g")
	w.close("")
	w.blank()

	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()

	w.open(fmt.Sprintf("func (g *%s) Generate(state *%s) {", structName, g.StateTypeName))
	if len(params) == 0 {
		// Zero-arg action: nothing to solve for. Still gate on
		// wouldFail_<Name> so the random scheduler can't fire an
		// action whose require would assert (mirrors ivy2cpp's
		// test-harness check).
		w.linef("if wouldFail_%s(state) { return }", goExportedName(name))
		// Test-driver trace: `> name` before invocation.
		g.emitActionTraceLine(w, ">", name, nil)
		w.linef("state.%s()", goExportedName(name))
		w.close("")
		w.blank()
		return
	}
	// 1. Build per-input symbols. Scalar params get one Const;
	// struct (destructor record) params get one Const for the
	// receiver plus one Const per scalar field, so the solver
	// synthesises each field independently (OPEN 055.7).
	w.line("// OPEN 055.1/.7: declare per-param + per-field input symbols.")
	for i, p := range params {
		paramSortCode, _ := g.reifySortAsGoCode(p.CSort)
		if paramSortCode == "" {
			paramSortCode = "goivy.Boolean"
		}
		w.linef("__in%d := goivy.NewConst(\"__in%d_%s\", %s)", i, i, goIdent(p.Name), paramSortCode)
		w.linef("_ = __in%d", i)
		// Per-field input symbols for struct-typed params.
		if recName, ok := g.destructorStructName(p.CSort); ok {
			for _, f := range g.destructorScalarFields(recName) {
				fieldSortCode, _ := g.reifySortAsGoCode(f.Sort)
				if fieldSortCode == "" {
					fieldSortCode = "goivy.Boolean"
				}
				w.linef("__in%d_%s := goivy.NewConst(\"__in%d_%s_%s\", %s)",
					i, goIdent(f.Name), i, goIdent(p.Name), goIdent(f.Name), fieldSortCode)
				w.linef("_ = __in%d_%s", i, goIdent(f.Name))
			}
			continue
		}
		// Per-leaf-field input symbols for variant-typed params.
		if superName, ok := g.variantSuperName(p.CSort); ok {
			for _, leaf := range g.variantLeaves(superName) {
				for _, f := range leaf.Fields {
					fieldSortCode, _ := g.reifySortAsGoCode(f.Sort)
					if fieldSortCode == "" {
						fieldSortCode = "goivy.Boolean"
					}
					w.linef("__in%d_%s_%s := goivy.NewConst(\"__in%d_%s_%s_%s\", %s)",
						i, goIdent(leaf.Name), goIdent(f.Name),
						i, goIdent(p.Name), goIdent(leaf.Name), goIdent(f.Name),
						fieldSortCode)
					w.linef("_ = __in%d_%s_%s", i, goIdent(leaf.Name), goIdent(f.Name))
				}
			}
		}
	}
	w.line("var modelResult *goivy.ModelResult")
	w.line("if g.sol != nil {")
	w.line("\t// OPEN 055.3/.7: seed with state facts, the action's")
	w.line("\t// reified Pre, and any per-field equalities for struct")
	w.line("\t// inputs (so the solver synthesises each field too).")
	preInputs := preconditionCallArgs(g, params)
	preCall := "buildPrecondition_" + goExportedName(name) + "(state"
	if len(preInputs) > 0 {
		preCall += ", " + strings.Join(preInputs, ", ")
	}
	preCall += ")"
	w.linef("\tpreclauses := %s", preCall)
	w.line("\tmodelResult, _ = g.sol.GetModelClauses(preclauses)")
	w.line("}")
	w.line("_ = modelResult")
	// 2. Pick each input. Scalar params go through pick*Or-Choose;
	// struct params have each field picked separately, then the
	// record is assembled.
	callArgs := make([]string, len(params))
	for i, p := range params {
		if p == nil {
			callArgs[i] = ""
			continue
		}
		typeName := g.goType(p.CSort)
		card := goSortCard(g, p.CSort)
		switch {
		case typeName == "bool":
			w.linef("v%d := pickBoolOrChoose(g.sol, modelResult, __in%d)", i, i)
		case card > 0 && goIsAnyIntegerType(g, p.CSort):
			w.linef("v%d := %s(pickUintOrChoose(g.sol, modelResult, __in%d, %d))", i, typeName, i, card)
		default:
			if recName, ok := g.destructorStructName(p.CSort); ok {
				// OPEN 055.7: assemble the struct from per-field
				// model reads.
				g.emitStructInputAssembly(w, i, typeName, recName)
			} else if superName, ok := g.variantSuperName(p.CSort); ok {
				// OPEN 055.8: pick a tag, then assemble via the
				// per-leaf constructor with payload (if any).
				g.emitVariantInputAssembly(w, i, typeName, superName)
			} else {
				// Any other shape we can't synthesise: zero value.
				w.linef("var v%d %s", i, typeName)
				w.linef("_ = v%d", i)
			}
		}
		callArgs[i] = fmt.Sprintf("v%d", i)
	}
	// Pre-firing precondition check. Skip this iteration when the
	// picked inputs would make the action fail (mirrors ivy2cpp's
	// test harness, which checks Z3-model feasibility before
	// invoking the action). Without this, randomized inputs
	// frequently violate `require` clauses and crash the binary.
	w.linef("if wouldFail_%s(state%s) { return }",
		goExportedName(name), prefixCommaArgs(callArgs))
	// Test-driver trace: `> name(args)` before invocation.
	g.emitDriverTraceLine(w, name, callArgs)
	w.linef("state.%s(%s)", goExportedName(name), strings.Join(callArgs, ", "))
	w.close("")
	w.blank()
	g.Ctx.OnceGlobals["__need_pickinput"] = true
}

// prefixCommaArgs returns ", a, b, c" when args is non-empty, else
// "". Used to splice variadic args onto a function call alongside a
// fixed prefix argument.
func prefixCommaArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return ", " + strings.Join(args, ", ")
}

// emitDriverTraceLine emits the `> name(arg0,arg1,…)` test-driver
// trace at the start of actionGen.Generate. Counterpart to
// emitActionTraceLine which emits `<` inside the action body.
//
// callArgs is the list of pre-computed local names (v0, v1, …) the
// driver is about to pass to the action.
func (g *Generator) emitDriverTraceLine(w *goWriter, name string, callArgs []string) {
	display := strings.TrimPrefix(name, "ext:")
	g.Ctx.AddImport("actions", "fmt", "")
	if len(callArgs) == 0 {
		w.linef(`fmt.Fprintln(ivyTraceOut, %q)`, "> "+display)
		return
	}
	var fmtStr strings.Builder
	fmtStr.WriteString("> ")
	fmtStr.WriteString(display)
	fmtStr.WriteByte('(')
	for i := range callArgs {
		if i > 0 {
			fmtStr.WriteByte(',')
		}
		fmtStr.WriteString("%v")
	}
	fmtStr.WriteString(")\n")
	w.linef(`fmt.Fprintf(ivyTraceOut, %q, %s)`, fmtStr.String(), strings.Join(callArgs, ", "))
}

// emitCloseSolver writes a Close method on each action generator so
// callers can free the solver. Aggregated emission keeps action_gen.go
// from sprouting per-struct close methods.
func (g *Generator) emitCloseSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	for _, name := range g.actionGenNames() {
		structName := "actionGen_" + goExportedName(name)
		w.open(fmt.Sprintf("func (g *%s) Close() error {", structName))
		w.open("if g.sol != nil {")
		w.line("err := g.sol.Close()")
		w.line("g.sol = nil")
		w.line("return err")
		w.close("")
		w.line("return nil")
		w.close("")
		w.blank()
	}
}
