package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Thunk emission for emit_assign_large (Python ivy_to_cpp.py:499-578 and
// 3654-3662). When the LHS of an assignment has free variables whose
// sorts have no derivable iteration bounds (and no extensional anchor),
// we cannot emit explicit loops. Instead we wrap the new function value
// in a hash_thunk that lazily evaluates each entry on demand.
//
// Python emits the thunk struct at file scope (via `thunks = impl` so
// make_thunk writes directly to the impl buffer). C++11 also permits
// local class definitions inside function bodies, with the restriction
// that local classes cannot be passed as template arguments. Our usage
// passes the thunk pointer to hash_thunk's constructor by value/pointer,
// which is fine for local classes. So we emit the struct inline at the
// assignment site for simplicity, sidestepping the impl-buffer ordering
// dance.

// makeThunk emits a C++ thunk struct definition and returns the
// construction expression that wraps a new instance in a hash_thunk.
// Mirrors Python `make_thunk` (ivy_to_cpp.py:499-604) for the non-z3
// target. The z3 / gen-mode path (Python lines 538-602) is currently
// stubbed — when target is gen/test we still emit a working thunk but
// without the to_z3 method; this is sufficient for `impl`/`repl`/`class`
// targets and leaves Z3 thunks as follow-up work documented inline.
//
// vs are the loop variables (the domain of the new function value).
// expr is the body that produces a range value when each v in vs is
// substituted by the corresponding arg field of the domain tuple.
func (g *Generator) makeThunk(w *cppWriter, vs []*goivy.LogicVariable, expr goivy.Expr) string {
	name := fmt.Sprintf("__thunk__%d", g.thunkCtr)
	g.thunkCtr++

	domSorts := make([]goivy.Sort, len(vs))
	for i, v := range vs {
		domSorts[i] = v.VSort
	}
	domT := cppCTupleNameWith(g, domSorts, "")
	rangeT := g.cppType(expr.NodeSort())

	// Collect captured environment symbols — any Const symbols
	// referenced in expr that aren't loop variables themselves.
	envSyms := g.thunkEnvSymbols(vs, expr)

	// Header: struct __thunk__N : thunk<D, R> {
	w.open(fmt.Sprintf("struct %s : thunk<%s, %s> {", name, domT, rangeT))
	// Env fields
	for _, sym := range envSyms {
		w.linef("%s;", g.cppStorageDecl(sym.Name, sym.CSort, ""))
	}
	// Constructor: __thunk__N(<envDecls>) : env1(env1), env2(env2), ... {}
	if len(envSyms) > 0 {
		decls := make([]string, len(envSyms))
		inits := make([]string, len(envSyms))
		for i, sym := range envSyms {
			decls[i] = g.cppStorageDecl(sym.Name, sym.CSort, "")
			inits[i] = fmt.Sprintf("%s(%s)", varName(sym.Name), varName(sym.Name))
		}
		w.linef("%s(%s) : %s {}", name, strings.Join(decls, ", "), strings.Join(inits, ", "))
	} else {
		w.linef("%s() {}", name)
	}
	// operator()(const D &arg) returns R
	w.open(fmt.Sprintf("%s operator()(const %s &arg) {", rangeT, domT))
	body, err := g.emitThunkBody(vs, expr)
	if err != nil {
		g.unsupported(w, "unsupported thunk body: %s", err.Error())
		w.close("")
		w.close(";")
		return fmt.Sprintf("hash_thunk<%s, %s>()", domT, rangeT)
	}
	w.linef("return %s;", body)
	w.close("")
	w.close(";")

	// Construction expression: hash_thunk<D, R>(new __thunk__N(envs...))
	envArgs := make([]string, len(envSyms))
	for i, sym := range envSyms {
		envArgs[i] = varName(sym.Name)
	}
	return fmt.Sprintf("hash_thunk<%s, %s>(new %s(%s))",
		domT, rangeT, name, strings.Join(envArgs, ", "))
}

// emitThunkBody emits the C++ expression for `expr` with each loop
// variable v in vs substituted by `arg` (single-var case) or
// `arg.arg<idx>` (multi-var case). Mirrors Python make_thunk's
// substitute_ast at ivy_to_cpp.py:533-535.
func (g *Generator) emitThunkBody(vs []*goivy.LogicVariable, expr goivy.Expr) (string, error) {
	subs := map[goivy.NodeKey]goivy.Expr{}
	if len(vs) == 1 {
		// Single-variable: v → Const("arg", v.VSort)
		v := vs[0]
		// We synthesize a Const because the substitution map keys on
		// node identity and the body emits Const by name lookup,
		// producing `arg` directly.
		subs[goivy.Key(v)] = goivy.NewConst("arg", v.VSort)
	} else {
		// Multi-var: v_i → Const("arg.arg<i>", v_i.VSort)
		for i, v := range vs {
			subs[goivy.Key(v)] = goivy.NewConst(fmt.Sprintf("arg.arg%d", i), v.VSort)
		}
	}
	sub, err := goivy.Substitute(expr, subs)
	if err != nil {
		return "", err
	}
	return g.emitExpr(sub)
}

// thunkEnvSymbols returns the Const symbols referenced by expr that are
// captured by the thunk as environment. Excludes the loop variables
// themselves (they're substituted by arg/arg.argN in emitThunkBody).
// Mirrors Python's `gather_referenced_symbols(expr, syms)` then
// `env = [sym for sym in syms if sym not in is_derived]` at
// ivy_to_cpp.py:511-514.
//
// For simplicity, this Go port treats every Const in the expr's symbol
// set as part of the environment (derived/function filtering is a
// follow-up). Loop variables are LogicVariables, not Consts, so they
// naturally don't appear here.
func (g *Generator) thunkEnvSymbols(vs []*goivy.LogicVariable, expr goivy.Expr) []*goivy.Const {
	used := goivy.UsedSymbolsAst(expr)
	loopNames := map[string]bool{}
	for _, v := range vs {
		loopNames[v.Name] = true
	}
	seen := map[string]bool{}
	var out []*goivy.Const
	for _, sym := range used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if loopNames[c.Name] {
			continue
		}
		if seen[c.Name] {
			continue
		}
		// Skip numerals and obvious literals — they aren't state.
		if goivy.IsNumeral(c) {
			continue
		}
		// Skip Boolean true/false.
		if c.CSort == goivy.Boolean && (c.Name == "true" || c.Name == "false") {
			continue
		}
		seen[c.Name] = true
		out = append(out, c)
	}
	return out
}

// emitAssignLarge mirrors Python `emit_assign_large` (ivy_to_cpp.py:3654-3662).
// When the LHS f(args...) has free variables that cannot be iterated, we
// build a guarded RHS:
//
//	expr = Ite(And(eqs...), originalRHS, f(vs'...))
//
// where vs' are fresh domain variables and eqs are equalities binding each
// concrete (non-variable) arg of the original LHS. Then we assign
// f = makeThunk(vs', expr).
func (g *Generator) emitAssignLarge(w *cppWriter, a *goivy.LogicAssignAction, lhsVs []*goivy.LogicVariable) {
	lhsApp, ok := a.LHS.(*goivy.Apply)
	if !ok {
		g.unsupported(w, "emit_assign_large requires apply-shaped LHS, got %T", a.LHS)
		return
	}
	lhsFn, ok := lhsApp.Func.(*goivy.Const)
	if !ok {
		g.unsupported(w, "emit_assign_large requires Const function on LHS, got %T", lhsApp.Func)
		return
	}
	fnSort, ok := lhsFn.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		g.unsupported(w, "emit_assign_large LHS function lacks FunctionSort: %s", lhsFn.Name)
		return
	}
	dom := fnSort.Domain()
	if len(dom) != len(lhsApp.Terms) {
		g.unsupported(w, "emit_assign_large arity mismatch: %d vs %d", len(dom), len(lhsApp.Terms))
		return
	}
	// Build fresh vs' replacing non-variable args with synthesized
	// variables. Preserve LogicVariable args unchanged so RHS reads of
	// the original args (e.g., self-referential LHS) still resolve.
	vsPrime := make([]*goivy.LogicVariable, len(dom))
	var eqs []goivy.Expr
	for i, t := range lhsApp.Terms {
		if v, ok := t.(*goivy.LogicVariable); ok {
			vsPrime[i] = v
			continue
		}
		// Synthesize a fresh variable for slot i.
		fresh, err := goivy.NewVariable(fmt.Sprintf("X%d", i), dom[i])
		if err != nil {
			g.unsupported(w, "emit_assign_large failed to synthesize bound variable: %s", err.Error())
			return
		}
		vsPrime[i] = fresh
		eq, err := goivy.NewEq(t, fresh)
		if err != nil {
			g.unsupported(w, "emit_assign_large equality build failed: %s", err.Error())
			return
		}
		eqs = append(eqs, eq)
	}
	// Build expr = Ite(And(eqs...), a.RHS, lhsFn(vsPrime...)) if eqs.
	var expr goivy.Expr = a.RHS
	if len(eqs) > 0 {
		var cond goivy.Expr
		if len(eqs) == 1 {
			cond = eqs[0]
		} else {
			and, err := goivy.NewAnd(eqs...)
			if err != nil {
				g.unsupported(w, "emit_assign_large And build failed: %s", err.Error())
				return
			}
			cond = and
		}
		elseArgs := make([]goivy.Expr, len(vsPrime))
		for i, v := range vsPrime {
			elseArgs[i] = v
		}
		elseExpr := goivy.NewApplyUnchecked(lhsFn, elseArgs...)
		ite, err := goivy.NewIte(cond, a.RHS, elseExpr)
		if err != nil {
			g.unsupported(w, "emit_assign_large Ite build failed: %s", err.Error())
			return
		}
		expr = ite
	}
	thunkExpr := g.makeThunk(w, vsPrime, expr)
	w.linef("%s = %s;", varName(lhsFn.Name), thunkExpr)
	_ = lhsVs // (unused for now — kept for future emit_assign_large variants)
}
