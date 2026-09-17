package ivy2cpp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Thunk emission for emit_assign_large (Python ivy_to_cpp.py:499-578 and
// 3654-3662). When the LHS of an assignment has free variables whose
// sorts have no derivable iteration bounds (and no extensional anchor),
// we cannot emit explicit loops. Instead we wrap the new function value
// in a hash_thunk that lazily evaluates each entry on demand.
//
// Full Generate mirrors Python's `thunks = impl` behavior: thunk structs
// are collected in a file-scope impl buffer and emitted before any method
// bodies that instantiate them. Direct white-box unit calls keep the
// legacy inline writer unless Generator.fileScopeThunks is set.

// makeThunk emits a C++ thunk struct definition and returns the
// construction expression that wraps a new instance in a hash_thunk.
// Mirrors Python `make_thunk` (ivy_to_cpp.py:499-604). For gen/test
// targets it emits a z3_thunk subclass with Python's to_z3 background
// encoding, so hash_thunk values can be sent back to the solver.
//
// vs are the loop variables (the domain of the new function value).
// expr is the body that produces a range value when each v in vs is
// substituted by the corresponding arg field of the domain tuple.
func (g *Generator) makeThunk(w *cppWriter, vs []*goivy.LogicVariable, expr goivy.Expr) string {
	expr = g.expandDerivedForThunk(expr)

	domSorts := make([]goivy.Sort, len(vs))
	for i, v := range vs {
		domSorts[i] = v.VSort
	}
	domT := cppCTupleNameWith(g, domSorts, g.ClassName)
	rangeT := g.cppQualifiedType(expr.NodeSort(), g.ClassName)
	defDomT := domT
	defRangeT := rangeT
	defClassName := g.ClassName
	if g.fileScopeThunks {
		defClassName = g.ClassName
		defDomT = cppCTupleNameWith(g, domSorts, g.ClassName)
		defRangeT = g.cppQualifiedType(expr.NodeSort(), g.ClassName)
	}

	// Collect captured environment symbols — any Const symbols
	// referenced in expr that aren't loop variables themselves.
	envSyms := g.thunkEnvSymbols(w, vs, expr)

	thunkClass := "thunk"
	if g.usesZ3() {
		thunkClass = "z3_thunk"
	}
	defW := w
	name := ""
	emitDefinition := true
	if g.fileScopeThunks {
		if g.thunkMemo == nil {
			g.thunkMemo = map[string]string{}
		}
		key := g.thunkMemoKey(thunkClass, defDomT, defRangeT, vs, envSyms, expr)
		if memoName, ok := g.thunkMemo[key]; ok {
			name = memoName
			emitDefinition = false
		} else {
			name = g.nextThunkName()
			g.thunkMemo[key] = name
		}
		defW = &g.thunkDefs
	} else {
		name = g.nextThunkName()
		if g.thunkWriter != nil {
			defW = g.thunkWriter
		}
	}

	if emitDefinition {
		g.emitThunkStruct(defW, name, thunkClass, defDomT, defRangeT, defClassName, vs, expr, envSyms)
	}

	// Construction expression: hash_thunk<D, R>(new __thunk__N(envs...))
	envArgs := make([]string, len(envSyms))
	for i, sym := range envSyms {
		envArgs[i] = varName(sym.Name)
	}
	return fmt.Sprintf("hash_thunk<%s, %s>(new %s(%s))",
		domT, rangeT, name, strings.Join(envArgs, ", "))
}

func (g *Generator) nextThunkName() string {
	name := fmt.Sprintf("__thunk__%d", g.thunkCtr)
	g.thunkCtr++
	return name
}

func (g *Generator) thunkMemoKey(thunkClass, domT, rangeT string, vs []*goivy.LogicVariable, envSyms []*goivy.Const, expr goivy.Expr) string {
	parts := []string{thunkClass, domT, rangeT, goivy.ReprExpr(expr)}
	for _, v := range vs {
		parts = append(parts, "v:"+v.Name+":"+sortName(v.VSort))
	}
	for _, sym := range envSyms {
		parts = append(parts, "env:"+sym.Name+":"+sortName(sym.CSort))
	}
	return strings.Join(parts, "|")
}

func (g *Generator) emitThunkStruct(w *cppWriter, name, thunkClass, domT, rangeT, className string, vs []*goivy.LogicVariable, expr goivy.Expr, envSyms []*goivy.Const) {
	// Header: struct __thunk__N : thunk/z3_thunk<D, R> {
	w.open(fmt.Sprintf("struct %s : %s<%s, %s> {", name, thunkClass, domT, rangeT))
	if g.usesZ3() {
		w.line("int __ident;")
	}
	// Env fields
	for _, sym := range envSyms {
		w.linef("%s;", g.cppStorageDecl(sym.Name, sym.CSort, className))
	}
	// Constructor: __thunk__N(<envDecls>) : env1(env1), env2(env2), ... {}
	if len(envSyms) > 0 {
		decls := make([]string, len(envSyms))
		inits := make([]string, len(envSyms))
		for i, sym := range envSyms {
			decls[i] = g.cppStorageDecl(sym.Name, sym.CSort, className)
			inits[i] = fmt.Sprintf("%s(%s)", varName(sym.Name), varName(sym.Name))
		}
		w.open(fmt.Sprintf("%s(%s) : %s {", name, strings.Join(decls, ", "), strings.Join(inits, ", ")))
	} else {
		w.open(fmt.Sprintf("%s() {", name))
	}
	if g.usesZ3() {
		w.line("__ident = z3_thunk_counter;")
		w.line("z3_thunk_counter++;")
	}
	w.close("")
	// operator()(const D &arg) returns R
	w.open(fmt.Sprintf("%s operator()(const %s &arg) {", rangeT, domT))
	body, err := g.emitThunkBody(vs, expr)
	if err != nil {
		g.unsupported(w, "unsupported thunk body: %s", err.Error())
		w.close("")
		w.close(";")
		return
	}
	w.linef("return %s;", body)
	w.close("")
	if g.usesZ3() {
		g.emitThunkToZ3(w, name, vs, expr, envSyms)
	}
	w.close(";")
}

// emitThunkBody emits the C++ expression for `expr` with each loop
// variable v in vs substituted by `arg` (single-var case) or
// `arg.arg<idx>` (multi-var case). Mirrors Python make_thunk's
// substitute_ast at ivy_to_cpp.py:533-535.
func (g *Generator) emitThunkBody(vs []*goivy.LogicVariable, expr goivy.Expr) (string, error) {
	sub, err := g.thunkSubstituteArgs(vs, expr)
	if err != nil {
		return "", err
	}
	return g.emitExpr(sub)
}

func (g *Generator) thunkSubstituteArgs(vs []*goivy.LogicVariable, expr goivy.Expr) (goivy.Expr, error) {
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
	return goivy.Substitute(expr, subs)
}

// thunkEnvSymbols returns the Const symbols referenced by expr that are
// captured by the thunk as environment. Excludes loop variables, literals,
// and derived definitions, matching Python make_thunk's
// `env = [sym for sym in syms if sym not in is_derived]`
// (AUDIT2 DONE 032). Function-typed symbols are captured when they are
// used as ordinary applications; a bare higher-order function value is
// reported as unsupported and not captured.
func (g *Generator) thunkEnvSymbols(w *cppWriter, vs []*goivy.LogicVariable, expr goivy.Expr) []*goivy.Const {
	used := goivy.UsedSymbolsAst(expr)
	loopNames := map[string]bool{}
	for _, v := range vs {
		loopNames[v.Name] = true
	}
	defNames := g.definitionNames()
	appliedFunctions := appliedFunctionConstKeys(expr)
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
		if defNames[c.Name] {
			continue
		}
		if _, isFn := c.CSort.(*goivy.LogicFunctionSort); isFn && !appliedFunctions[goivy.Key(c)] {
			if w != nil {
				g.unsupported(w, "thunk environment cannot capture bare function symbol %s", c.Name)
			}
			continue
		}
		seen[c.Name] = true
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := pythonThunkEnvRank(out[i]), pythonThunkEnvRank(out[j])
		if ri != rj {
			return ri < rj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func pythonThunkEnvRank(c *goivy.Const) int {
	if c == nil {
		return 4
	}
	if strings.HasPrefix(c.Name, "loc:") {
		return 0
	}
	if _, ok := c.CSort.(*goivy.LogicFunctionSort); ok {
		return 1
	}
	return 2
}

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

func allDefinitionParamsVariables(params []goivy.Expr) bool {
	for _, p := range params {
		if _, ok := p.(*goivy.LogicVariable); !ok {
			return false
		}
	}
	return true
}

func (g *Generator) emitThunkToZ3(w *cppWriter, name string, vs []*goivy.LogicVariable, origExpr goivy.Expr, envSyms []*goivy.Const) {
	w.open("z3::expr to_z3(gen &g, const z3::expr &v) {")
	expr, err := g.thunkSubstituteArgs(vs, origExpr)
	if err != nil {
		w.line("return g.ctx.bool_val(true);")
		w.close("")
		return
	}
	if g.isPrimitiveSort(expr.NodeSort()) {
		w.line("return g.ctx.bool_val(true);")
		w.close("")
		return
	}
	if len(goivy.VariablesAST(expr)) == 0 && g.allNumericOrEnumeratedConstants(goivy.UsedSymbolsAst(expr)) {
		code, err := g.emitExpr(expr)
		if err != nil {
			w.line("return g.ctx.bool_val(true);")
			w.close("")
			return
		}
		if g.thunkFastPathUsesToSolver(expr.NodeSort()) {
			w.linef("z3::expr res = __to_solver(g, v, %s);", code)
		} else {
			cty := "int"
			if g.hasStringInterp(expr.NodeSort()) {
				cty = "__strlit"
			}
			w.linef("z3::expr res = v == g.int_to_z3(g.sort(%s), (%s)(%s));", strconv.Quote(z3SortName(expr.NodeSort())), cty, code)
		}
		w.line("return res;")
		w.close("")
		return
	}

	w.line("std::ostringstream __ss;")
	w.line("__ss << __ident;")
	vsyms := make([]*goivy.Const, len(vs))
	for i, v := range vs {
		vsyms[i] = goivy.NewConst(fmt.Sprintf("%s_arg_%d", name, i), v.VSort)
	}
	rsymIdx := 0
	if len(vs) > 0 {
		rsymIdx = len(vs) - 1
	}
	rsym := goivy.NewConst(fmt.Sprintf("%s_res_%d", name, rsymIdx), expr.NodeSort())
	envSymsLocal := make([]*goivy.Const, len(envSyms))
	for i, sym := range envSyms {
		envSymsLocal[i] = goivy.NewConst(fmt.Sprintf("%s_env_%d", name, i), sym.CSort)
	}
	for _, c := range append(append([]*goivy.Const{}, vsyms...), append(envSymsLocal, rsym)...) {
		w.open(fmt.Sprintf("if (g.decls_by_name.find(%s) == g.decls_by_name.end()) {", strconv.Quote(c.Name)))
		g.emitDeclSolverWithName(w, stateSymbol{Name: c.Name, Sort: c.CSort}, "", "g.")
		w.close("")
	}

	subst := make(map[goivy.NodeKey]goivy.Expr, len(vs))
	for i, v := range vs {
		subst[goivy.Key(v)] = vsyms[i]
	}
	renamedExpr, err := goivy.Substitute(origExpr, subst)
	if err != nil {
		w.line("return g.ctx.bool_val(true);")
		w.close("")
		return
	}
	rename := make(map[goivy.NodeKey]*goivy.Const, len(envSyms))
	for i, sym := range envSyms {
		rename[goivy.Key(sym)] = envSymsLocal[i]
	}
	renamedExpr = goivy.RenameAST(renamedExpr, rename)
	eq, err := goivy.NewEq(rsym, renamedExpr)
	if err != nil {
		w.line("return g.ctx.bool_val(true);")
		w.close("")
		return
	}
	smt, ok := g.formulaToSmtlib(eq)
	if !ok {
		w.line("return g.ctx.bool_val(true);")
		w.close("")
		return
	}

	w.line("z3::expr res = g.ctx.bool_val(true);")
	if g.Config.Target == "test" {
		w.line("hash_map<std::string, std::string> rn;")
	} else {
		w.line("std::map<std::string, std::string> rn;")
	}
	for i, sym := range envSyms {
		locv := g.emitThunkLocalZ3Symbol(w, sym)
		g.emitSetSolverCustom(w, stateSymbol{Name: sym.Name, Sort: sym.CSort}, emitSetSolverOptions{
			prefix: "g.",
			gen:    "g",
			sname:  locv + ".c_str()",
			cvalue: varName(sym.Name),
			add: func(w *cppWriter, text string) {
				w.linef("g.slvr.add(%s);", text)
			},
		})
		w.linef("rn[%s] = %s.c_str();", strconv.Quote(envSymsLocal[i].Name), locv)
	}
	w.linef("z3::expr the_expr = z3::expr(g.ctx,Z3_parse_smtlib2_string(g.ctx, %s, g.sort_names.size(), &g.sort_names[0], &g.sorts[0], g.decl_names.size(), &g.decl_names[0], &g.decls[0]));", pythonZ3StringLiteral("(assert "+smt+")"))
	w.line("the_expr = __z3_rename(the_expr, rn);")
	w.line("g.ctx.check_error();")
	w.line("z3::expr_vector src(g.ctx);")
	w.line("z3::expr_vector dst(g.ctx);")
	for i, v := range vs {
		w.linef("src.push_back(g.ctx.constant(%s,g.sort(%s)));;", strconv.Quote(vsyms[i].Name), strconv.Quote(z3SortName(v.VSort)))
		w.linef("dst.push_back(v.arg(%d));;", i)
	}
	w.linef("src.push_back(g.ctx.constant(%s,g.sort(%s)));;", strconv.Quote(rsym.Name), strconv.Quote(z3SortName(rsym.CSort)))
	w.line("dst.push_back(v);;")
	w.line("g.ctx.check_error();")
	w.line("res = the_expr.substitute(src, dst);")
	w.line("g.ctx.check_error();")
	w.line("return res;")
	w.close("")
}

func pythonZ3StringLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " \"\n\"")
	return `"` + s + `"`
}

func (g *Generator) emitThunkLocalZ3Symbol(w *cppWriter, sym *goivy.Const) string {
	locv := "loc_" + varName(sym.Name)
	w.linef("std::string %s = std::string(\"__loc_\") + __ss.str() + std::string(\"__\") + %s;", locv, strconv.Quote(sym.Name))
	w.open(fmt.Sprintf("if (g.decls_by_name.find(%s) == g.decls_by_name.end()) {", locv))
	g.emitDeclSolverWithName(w, stateSymbol{Name: sym.Name, Sort: sym.CSort}, locv+".c_str()", "g.")
	w.close("")
	return locv
}

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

func (g *Generator) thunkFastPathUsesToSolver(s goivy.Sort) bool {
	if g == nil || g.Mod == nil {
		return false
	}
	if it, ok := g.cppInterpType(s); ok {
		// Python excludes `interpret T -> bv[N]` from sort_to_cpptype, so
		// primitive BV constants use the same int_to_z3 equality path as
		// numeric ranges. Helper-class interpreted types still use their
		// generated __to_solver specializations.
		return it.Kind != cppInterpBV
	}
	if g.isDestructorRecordRange(s) {
		return true
	}
	name := sortName(s)
	if name != "" && g.Mod.NativeTypes != nil {
		if _, ok := g.Mod.NativeTypes[name]; ok {
			return true
		}
	}
	if name != "" && g.isVariantSuperName(name) {
		return true
	}
	return false
}

func (g *Generator) isPrimitiveSort(s goivy.Sort) bool {
	if _, isFn := s.(*goivy.LogicFunctionSort); isFn {
		return false
	}
	if g == nil || g.Mod == nil || g.Mod.NativeTypes == nil {
		return false
	}
	nt, ok := g.Mod.NativeTypes[sortName(s)]
	if !ok || nt == nil {
		return false
	}
	code, err := g.nativeTypeFull(nt)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(code), "primitive ")
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
