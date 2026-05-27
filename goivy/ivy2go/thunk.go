package ivy2go

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// thunk.go mirrors ivy2cpp/thunk.go. A "thunk" wraps a large-domain
// function assignment in a memoized closure: the Ivy program said
// `f(X) := expr(X)` where the domain is too large to materialise into
// an array, so we compute each entry on demand.

// makeThunk emits a Go thunk struct definition into the package's
// thunks stream and returns the construction expression that
// instantiates it. Mirrors ivy2cpp/thunk.go makeThunk, simplified:
// we always use the file-scope memoization path so equivalent thunks
// deduplicate. Each thunk also emits a solver-side `toZ3Value` method
// so large-domain function state can be constrained during action input
// generation.
//
// vs are the loop variables (the domain of the new function value).
// expr is the body that produces a range value when each v in vs is
// substituted by the corresponding key tuple field.
func (g *Generator) makeThunk(vs []*goivy.LogicVariable, expr goivy.Expr) (string, error) {
	domSorts := make([]goivy.Sort, len(vs))
	for i, v := range vs {
		domSorts[i] = v.VSort
	}
	domT := goCTupleNameWith(g, domSorts)
	rangeT := g.goType(expr.NodeSort())

	envSyms := g.thunkEnvSymbols(vs, expr)

	if g.thunkMemo == nil {
		g.thunkMemo = map[string]string{}
	}
	key := g.thunkMemoKey(domT, rangeT, vs, envSyms, expr)
	name, ok := g.thunkMemo[key]
	if !ok {
		name = g.nextThunkName()
		g.thunkMemo[key] = name
		// Emit the struct definition into the thunks stream.
		g.emitThunkStruct(&g.thunks, name, domT, rangeT, vs, expr, envSyms)
	}

	// Construction expression: a fresh instance with env fields
	// populated. Returned to the caller (typically emitAssignLarge),
	// which assigns it to the destination state field.
	envArgs := make([]string, len(envSyms))
	for i, sym := range envSyms {
		envArgs[i] = "s." + goExportedName(sym.Name)
	}
	return fmt.Sprintf("new%s(%s)", name, strings.Join(envArgs, ", ")), nil
}

// nextThunkName mirrors ivy2cpp/thunk.go nextThunkName. Counter naming
// (`thunk_N`) is preserved verbatim modulo the leading underscores so
// cross-package audits stay readable.
func (g *Generator) nextThunkName() string {
	name := fmt.Sprintf("thunk_%d", g.thunkCtr)
	g.thunkCtr++
	return name
}

// thunkMemoKey ports ivy2cpp/thunk.go thunkMemoKey.
func (g *Generator) thunkMemoKey(domT, rangeT string, vs []*goivy.LogicVariable, envSyms []*goivy.Const, expr goivy.Expr) string {
	parts := []string{domT, rangeT, goivy.ReprExpr(expr)}
	for _, v := range vs {
		parts = append(parts, "v:"+v.Name+":"+sortName(v.VSort))
	}
	for _, sym := range envSyms {
		parts = append(parts, "env:"+sym.Name+":"+sortName(sym.CSort))
	}
	return strings.Join(parts, "|")
}

// emitThunkStruct writes a Go struct + constructor + get method into
// the thunks stream. Mirrors ivy2cpp/thunk.go emitThunkStruct.
//
//	type thunk_0 struct {
//	    memo map[<domT>]<rangeT>
//	    env_<sym1> <type1>   // captured state field
//	    env_<sym2> <type2>
//	}
//
//	func newthunk_0(env_sym1 t1, env_sym2 t2) *thunk_0 {
//	    return &thunk_0{memo: map[domT]rangeT{}, env_<sym1>: env_sym1, env_<sym2>: env_sym2}
//	}
//
//	func (t *thunk_0) get(k <domT>) <rangeT> {
//	    if v, ok := t.memo[k]; ok { return v }
//	    v := <body in terms of t.env_* and k>
//	    t.memo[k] = v
//	    return v
//	}
func (g *Generator) emitThunkStruct(w *goWriter, name, domT, rangeT string, vs []*goivy.LogicVariable, expr goivy.Expr, envSyms []*goivy.Const) {
	w.linef("// %s is a thunked large-domain function value.", name)
	w.open(fmt.Sprintf("type %s struct {", name))
	w.linef("memo map[%s]%s", domT, rangeT)
	for _, sym := range envSyms {
		w.linef("env_%s %s", goIdent(sym.Name), g.goType(sym.CSort))
	}
	w.close("")
	w.blank()

	// Constructor.
	params := make([]string, len(envSyms))
	for i, sym := range envSyms {
		params[i] = fmt.Sprintf("env_%s %s", goIdent(sym.Name), g.goType(sym.CSort))
	}
	w.linef("func new%s(%s) *%s {", name, strings.Join(params, ", "), name)
	w.linef("\tt := &%s{memo: map[%s]%s{}}", name, domT, rangeT)
	for _, sym := range envSyms {
		w.linef("\tt.env_%s = env_%s", goIdent(sym.Name), goIdent(sym.Name))
	}
	w.line("\treturn t")
	w.line("}")
	w.blank()

	// get method: memoized body. We rewrite the body so any
	// reference to a loop variable v becomes a field-access into
	// the key tuple, and any reference to an env symbol becomes
	// t.env_<sym>. Emitting the body without these rewrites would
	// produce invalid Go that doesn't know about k or t.
	w.linef("func (t *%s) get(k %s) %s {", name, domT, rangeT)
	w.line("\tif v, ok := t.memo[k]; ok {")
	w.line("\t\treturn v")
	w.line("\t}")
	body, err := g.emitThunkBody(vs, expr, envSyms)
	if err != nil {
		// Record the error and fall back to zero-value so the
		// emitted file remains gofmt-clean.
		g.errs = append(g.errs, fmt.Errorf("thunk body: %w", err))
		w.linef("\tvar v %s", rangeT)
		w.line("\t_ = k")
		w.line("\tt.memo[k] = v")
		w.line("\treturn v")
		w.line("}")
		w.blank()
		return
	}
	w.linef("\tv := %s", body)
	w.line("\tt.memo[k] = v")
	w.line("\treturn v")
	w.line("}")
	w.blank()

	// ToZ3: emits a goivy.Expr constraint saying `result ==
	// <body-with-args-substituted>`. Used by action_gen preconditions
	// when a hash-thunk-stored function symbol appears in the
	// formula — without this the solver sees the symbol as a free
	// variable and would find any model. Mirrors the role of
	// ivy2cpp/thunk.go emitThunkToZ3.
	g.emitThunkToZ3(w, name, vs, expr, envSyms)
}

// emitThunkToZ3 mirrors the role of ivy2cpp/thunk.go emitThunkToZ3.
// The cpp version emits a `to_z3` method that pushes an assertion
// into Z3's C++ solver; the Go equivalent returns a goivy.Expr
// constraint the caller can hand to goivy.Solver.
//
// Shape of the emitted Go methods:
//
//	func (t *thunk0) toZ3Value(args []goivy.Expr) goivy.Expr { ... }
//	func (t *thunk0) ToZ3(args []goivy.Expr, res goivy.Expr) goivy.Expr { ... }
//
// For simple bodies that don't reference env state, the reified body
// is a literal goivy.Expr tree built via reifyExprAsGoCode.
func (g *Generator) emitThunkToZ3(w *goWriter, name string, vs []*goivy.LogicVariable, expr goivy.Expr, envSyms []*goivy.Const) {
	g.Ctx.AddImport("thunks", g.Config.GoivyImportPath, "")
	// Build a substitution that renames loop vars to args[i] and
	// env syms to t.env_<name> by structural placeholder Consts.
	// reifyExprAsGoCode will lower these Consts to bare names; we
	// post-rewrite the emitted text to wire them through.
	subs := map[goivy.NodeKey]goivy.Expr{}
	argSentinels := make([]string, len(vs))
	for i, v := range vs {
		sentinel := fmt.Sprintf("__thunk_z3_arg_%d_%s", i, v.Name)
		subs[goivy.Key(v)] = &goivy.Const{Name: sentinel, CSort: v.VSort}
		argSentinels[i] = sentinel
	}
	envSentinels := make([]string, len(envSyms))
	for i, sym := range envSyms {
		sentinel := fmt.Sprintf("__thunk_z3_env_%d_%s", i, sym.Name)
		subs[goivy.Key(sym)] = &goivy.Const{Name: sentinel, CSort: sym.CSort}
		envSentinels[i] = sentinel
	}
	substituted := expr
	if len(subs) > 0 {
		if r, err := goivy.Substitute(expr, subs); err == nil {
			substituted = r
		}
	}
	prevAliases := g.reifyExprCodeAlias
	g.reifyExprCodeAlias = map[string]string{}
	for i, sent := range argSentinels {
		g.reifyExprCodeAlias[sent] = fmt.Sprintf("args[%d]", i)
	}
	for i, sent := range envSentinels {
		sortCode, ok := g.reifySortAsGoCode(envSyms[i].CSort)
		if !ok {
			sortCode = "goivy.Boolean"
		}
		g.reifyExprCodeAlias[sent] = fmt.Sprintf("goivy.NewConst(%q, %s)", envSyms[i].Name, sortCode)
	}
	bodyCode, ok := g.reifyExprAsGoCode(substituted, nil)
	g.reifyExprCodeAlias = prevAliases
	if !ok {
		// Body can't be reified — emit a stub that returns the
		// vacuous `true` constraint so callers don't crash.
		w.linef("// toZ3Value stub: thunk body not reifiable as runtime goivy.Expr.")
		w.linef("func (t *%s) toZ3Value(args []goivy.Expr) goivy.Expr {", name)
		w.line("\t_ = t")
		w.line("\t_ = args")
		w.line(`	return goivy.NewConst("true", goivy.Boolean)`)
		w.line("}")
		w.blank()
		w.linef("func (t *%s) ToZ3(args []goivy.Expr, res goivy.Expr) goivy.Expr {", name)
		w.line("\treturn &goivy.Eq{T1: res, T2: t.toZ3Value(args)}")
		w.line("}")
		w.blank()
		return
	}
	w.linef("// toZ3Value emits the thunk body as a solver expression.")
	w.linef("func (t *%s) toZ3Value(args []goivy.Expr) goivy.Expr {", name)
	w.line("\t_ = t")
	w.line("\t_ = args")
	w.linef("\treturn %s", bodyCode)
	w.line("}")
	w.blank()
	w.linef("// ToZ3 emits the constraint `res == <body>` for the solver.")
	w.linef("func (t *%s) ToZ3(args []goivy.Expr, res goivy.Expr) goivy.Expr {", name)
	w.line("\treturn &goivy.Eq{T1: res, T2: t.toZ3Value(args)}")
	w.line("}")
	w.blank()
}

// emitThunkBody substitutes references to loop variables (vs) by key
// tuple field accesses (k.Arg0, k.Arg1, …) and references to env
// symbols by t.env_<name>, then emits the expression.
func (g *Generator) emitThunkBody(vs []*goivy.LogicVariable, expr goivy.Expr, envSyms []*goivy.Const) (string, error) {
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, v := range vs {
		// Loop variable v of sort S → emit as accessor on k. When
		// there's only one variable, we use k directly (no struct
		// wrapping); when there are multiple, k is a tup__… struct
		// with fields Arg0, Arg1, …
		fieldName := "k"
		if len(vs) > 1 {
			fieldName = fmt.Sprintf("k.Arg%d", i)
		}
		// Mint a Const standing in for the field access. emitExpr
		// will lower this Const via its `goIdent` branch; we want
		// the literal Go expression to flow through, so install a
		// per-thunk alias.
		fakeName := "__thunk_var_" + v.Name
		subs[goivy.Key(v)] = &goivy.Const{Name: fakeName, CSort: v.VSort}
		g.exprAliases[fakeName] = nil
		_ = fieldName
	}
	// Env-symbol rewrite mirrors above; we use a unique fake name so
	// emitExpr's alias machinery routes us to t.env_<name>.
	applySubs := map[goivy.NodeKey]goivy.SubstituteApplyFunc{}
	for _, sym := range envSyms {
		fakeName := "__thunk_env_" + sym.Name
		subs[goivy.Key(sym)] = &goivy.Const{Name: fakeName, CSort: sym.CSort}
		if _, ok := sym.CSort.(*goivy.LogicFunctionSort); ok {
			fakeFunc := &goivy.Const{Name: fakeName, CSort: sym.CSort}
			applySubs[goivy.Key(sym)] = func(terms []goivy.Expr) goivy.Expr {
				return goivy.MustApply(fakeFunc, terms...)
			}
		}
		_ = fakeName
	}
	if len(applySubs) > 0 {
		expr = goivy.SubstituteApply(expr, applySubs)
	}
	if len(subs) > 0 {
		body, err := goivy.Substitute(expr, subs)
		if err != nil {
			return "", err
		}
		expr = body
	}
	// Emit the substituted body with a temporary alias map that
	// remaps __thunk_var_* → k or k.Arg* and __thunk_env_* →
	// t.env_*. We splice these in via a per-call alias hook.
	prev := g.exprAliases
	g.exprAliases = map[string]goivy.Expr{}
	// Use string replacement after emitExpr: simpler than threading
	// through aliasForSymbol since aliases there resolve to
	// goivy.Expr values, not arbitrary text. The replacement is
	// purely on tokens we minted ourselves, so no risk of collision.
	out, err := g.emitExpr(expr)
	g.exprAliases = prev
	if err != nil {
		return "", err
	}
	for i, v := range vs {
		fake := "__thunk_var_" + v.Name
		// emitExpr lowers a bare Const through goIdent → fake.
		ident := goIdent(fake)
		var replacement string
		if len(vs) == 1 {
			replacement = "k"
		} else {
			replacement = fmt.Sprintf("k.Arg%d", i)
		}
		out = strings.ReplaceAll(out, ident, replacement)
	}
	for _, sym := range envSyms {
		fake := "__thunk_env_" + sym.Name
		ident := goIdent(fake)
		out = strings.ReplaceAll(out, ident, "t.env_"+goIdent(sym.Name))
	}
	return out, nil
}

// thunkEnvSymbols mirrors ivy2cpp/thunk.go thunkEnvSymbols. Returns
// the Const symbols referenced in expr that aren't loop variables
// themselves — these become env fields on the thunk struct.
//
// Note: goivy.Apply.Children returns only Terms, not the Func slot.
// We descend into Apply.Func explicitly so the called symbol gets
// captured as env.
func (g *Generator) thunkEnvSymbols(vs []*goivy.LogicVariable, expr goivy.Expr) []*goivy.Const {
	loopNames := map[string]bool{}
	for _, v := range vs {
		loopNames[v.Name] = true
	}
	seen := map[string]bool{}
	var env []*goivy.Const
	var walk func(e goivy.Expr)
	walk = func(e goivy.Expr) {
		if e == nil {
			return
		}
		if c, ok := e.(*goivy.Const); ok {
			if !loopNames[c.Name] && !goivy.IsNumeral(c) && !seen[c.Name] {
				// Skip enum constants — they're package-level Go
				// constants, not state captures.
				if g.isEnumConstantName(c.Name) {
					return
				}
				if !g.isStateSymbolName(c.Name) {
					return
				}
				seen[c.Name] = true
				env = append(env, c)
			}
			return
		}
		if a, ok := e.(*goivy.Apply); ok {
			walk(a.Func)
			for _, t := range a.Terms {
				walk(t)
			}
			return
		}
		for _, ch := range e.Children() {
			walk(ch)
		}
	}
	walk(expr)
	sort.Slice(env, func(i, j int) bool { return env[i].Name < env[j].Name })
	return env
}

// thunkSubstituteArgs mirrors ivy2cpp/thunk.go:164. Replaces each
// loop variable in vs with a Const named `arg` (single-var) or
// `arg.arg<i>` (multi-var) so the body can lower through the
// standard emitExpr path.
func (g *Generator) thunkSubstituteArgs(vs []*goivy.LogicVariable, expr goivy.Expr) (goivy.Expr, error) {
	subs := map[goivy.NodeKey]goivy.Expr{}
	if len(vs) == 1 {
		v := vs[0]
		subs[goivy.Key(v)] = goivy.NewConst("arg", v.VSort)
	} else {
		for i, v := range vs {
			subs[goivy.Key(v)] = goivy.NewConst(fmt.Sprintf("arg.arg%d", i), v.VSort)
		}
	}
	return goivy.Substitute(expr, subs)
}

// appliedFunctionConstKeys mirrors ivy2cpp/thunk.go:233. Walks expr
// and returns the set of NodeKeys for function-sorted Const symbols
// used in Apply positions — i.e. function symbols that are actually
// invoked (rather than referenced as values).
func appliedFunctionConstKeys(expr goivy.Expr) map[goivy.NodeKey]bool {
	out := map[goivy.NodeKey]bool{}
	var walk func(goivy.Expr)
	walk = func(e goivy.Expr) {
		if e == nil {
			return
		}
		if app, ok := e.(*goivy.Apply); ok {
			if c, ok := app.Func.(*goivy.Const); ok {
				if _, isFn := c.CSort.(*goivy.LogicFunctionSort); isFn {
					out[goivy.Key(c)] = true
				}
			}
		}
		for _, child := range e.Children() {
			walk(child)
		}
	}
	walk(expr)
	return out
}

// expandDerivedForThunk mirrors ivy2cpp/thunk.go:255. Unfolds all
// derived definitions in expr, iterating to a fixed point (capped at
// definitions.Len()+4 passes to defend against pathological loops).
func (g *Generator) expandDerivedForThunk(expr goivy.Expr) goivy.Expr {
	defs := g.allDefinitions()
	if len(defs) == 0 || expr == nil {
		return expr
	}
	out := expr
	for i := 0; i < len(defs)+4; i++ {
		next := g.expandDerivedForThunkOnce(out, defs)
		if next == nil || next.Equal(out) {
			return out
		}
		out = next
	}
	return out
}

// expandDerivedForThunkOnce mirrors ivy2cpp/thunk.go:271. Single
// pass over expr applying each derived definition's substitution.
func (g *Generator) expandDerivedForThunkOnce(expr goivy.Expr, defs []derivedDefinition) goivy.Expr {
	applySubs := map[goivy.NodeKey]goivy.SubstituteApplyFunc{}
	constSubs := map[goivy.NodeKey]goivy.Expr{}
	for _, d := range defs {
		if d.Head == nil || d.RHS == nil {
			continue
		}
		if len(d.Params) == 0 {
			constSubs[goivy.Key(d.Head)] = d.RHS
			continue
		}
		def := d
		if !allDefinitionParamsVariables(def.Params) {
			continue
		}
		applySubs[goivy.Key(def.Head)] = func(terms []goivy.Expr) goivy.Expr {
			if len(terms) != len(def.Params) {
				return goivy.MustApply(def.Head, terms...)
			}
			subs := make(map[goivy.NodeKey]goivy.Expr, len(def.Params))
			for i, p := range def.Params {
				subs[goivy.Key(p)] = terms[i]
			}
			rhs, err := goivy.Substitute(def.RHS, subs)
			if err != nil {
				return def.RHS
			}
			return rhs
		}
	}
	out := expr
	if len(applySubs) > 0 {
		out = goivy.SubstituteApply(out, applySubs)
	}
	if len(constSubs) > 0 {
		out = goivy.SubstituteConstantsExpr(out, constSubs)
	}
	return out
}

// allDefinitionParamsVariables mirrors ivy2cpp/thunk.go:311.
func allDefinitionParamsVariables(params []goivy.Expr) bool {
	for _, p := range params {
		if _, ok := p.(*goivy.LogicVariable); !ok {
			return false
		}
	}
	return true
}

// allNumericOrEnumeratedConstants mirrors ivy2cpp/thunk.go:449.
func (g *Generator) allNumericOrEnumeratedConstants(used *goivy.InsMap[goivy.NodeKey, goivy.Expr]) bool {
	if used == nil {
		return true
	}
	for _, sym := range used.All() {
		if !g.isNumericOrEnumeratedConstant(sym) {
			return false
		}
	}
	return true
}

// isNumericOrEnumeratedConstant mirrors ivy2cpp/thunk.go:461.
func (g *Generator) isNumericOrEnumeratedConstant(sym goivy.Expr) bool {
	c, ok := sym.(*goivy.Const)
	if !ok {
		return false
	}
	if goivy.IsNumeral(c) {
		return true
	}
	if _, ok := c.CSort.(*goivy.LogicEnumeratedSort); ok {
		return true
	}
	if g != nil && g.Mod != nil && g.Mod.Sig != nil {
		if itp, ok := g.Mod.Sig.Interp[sortName(c.CSort)]; ok {
			if _, isEnum := itp.(*goivy.LogicEnumeratedSort); isEnum {
				return true
			}
		}
	}
	return false
}
