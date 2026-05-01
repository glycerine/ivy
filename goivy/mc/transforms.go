package mc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// ElimIte eliminates ITEs over non-finite sorts by introducing fresh variables
// and constraints. An ITE (x if c else y) where x has non-finite sort becomes
// a fresh variable v with added constraint (v=x if c else v=y).
// As an optimization, (z = (x if c else y)) → (z=x if c else z=y) directly.
//
// Python: ivy_mc.py:757-772
func ElimIte(expr lg.Expr, cnsts *[]lg.Expr, iteCtr *int64) lg.Expr {
	switch t := expr.(type) {
	case *lg.Ite:
		if !isFiniteSort(t.Then.NodeSort()) {
			name := fmt.Sprintf("__ite[%d]", nextIteCtr(iteCtr))
			v := lg.NewConst(name, t.Then.NodeSort())
			cElim := ElimIte(t.Cond, cnsts, iteCtr)
			eqThen := ElimIte(&lg.Eq{T1: v, T2: t.Then}, cnsts, iteCtr)
			eqElse := ElimIte(&lg.Eq{T1: v, T2: t.Else}, cnsts, iteCtr)
			*cnsts = append(*cnsts, &lg.Ite{Cond: cElim, Then: eqThen, Else: eqElse})
			return v
		}
	case *lg.Eq:
		if ite, ok := t.T2.(*lg.Ite); ok {
			if !isFiniteSort(ite.Then.NodeSort()) {
				cElim := ElimIte(ite.Cond, cnsts, iteCtr)
				eqThen := ElimIte(&lg.Eq{T1: t.T1, T2: ite.Then}, cnsts, iteCtr)
				eqElse := ElimIte(&lg.Eq{T1: t.T1, T2: ite.Else}, cnsts, iteCtr)
				return &lg.Ite{Cond: cElim, Then: eqThen, Else: eqElse}
			}
		}
	}

	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Expr, len(children))
	for i, child := range children {
		newChildren[i] = ElimIte(child, cnsts, iteCtr)
	}
	return il.CloneNode(expr, newChildren)
}

// ToTableLookup translates finite function applications to table lookups.
// This is preferable to using axioms if the argument types are large, since
// it prevents copying the argument expressions. Also, the result is a
// circuit (as opposed to a set of constraints) which might be helpful to ABC.
//
// Python: ivy_mc.py:1063-1114
func ToTableLookup(trans *module.Clauses, invariant lg.Expr) (*module.Clauses, lg.Expr) {
	var newDefs []lg.Expr
	counter := 0

	argSym := func(sort lg.Sort) *lg.Const {
		name := fmt.Sprintf("__arg[%d]", counter)
		counter++
		return lg.NewConst(name, sort)
	}

	var recur func(expr lg.Expr) lg.Expr
	recur = func(expr lg.Expr) lg.Expr {
		// Check if this is an application with all finite-sort args
		if app, ok := expr.(*lg.Apply); ok && len(app.Terms) > 0 {
			if funcSym, ok := app.Func.(*lg.Const); ok && !isInterpretedSymbol(funcSym) {
				allFinite := true
				for _, arg := range app.Terms {
					if !isFiniteSort(arg.NodeSort()) {
						allFinite = false
						break
					}
				}
				if allFinite {
					return tableLookupApp(app, funcSym, argSym, &newDefs, recur)
				}
			}
		}

		// Recurse into children
		children := expr.Children()
		if len(children) == 0 {
			return expr
		}
		nc := make([]lg.Expr, len(children))
		for i, c := range children {
			nc[i] = recur(c)
		}
		return il.CloneNode(expr, nc)
	}

	// Check if there are any finite-domain functions
	// (For simplicity, we check during processing and return unchanged if nothing matched)
	defs := make([]lg.Expr, len(trans.Defs))
	changed := false
	for i, df := range trans.Defs {
		defs[i] = recur(df)
		if defs[i] != df {
			changed = true
		}
	}
	fmlas := make([]lg.Expr, len(trans.Fmlas))
	for i, fmla := range trans.Fmlas {
		fmlas[i] = recur(fmla)
		if fmlas[i] != fmla {
			changed = true
		}
	}

	if !changed && len(newDefs) == 0 {
		return trans, invariant
	}

	// Build new defs including the freshly introduced arg definitions
	allDefs := make([]*il.Definition, 0, len(defs)+len(newDefs))
	for _, d := range defs {
		if def, ok := d.(*il.Definition); ok {
			allDefs = append(allDefs, def)
		}
	}
	for _, d := range newDefs {
		if def, ok := d.(*il.Definition); ok {
			allDefs = append(allDefs, def)
		}
	}
	newTrans := module.NewClauses(fmlas, allDefs, trans.Annot)

	// Process invariant
	newDefs = nil
	newInv := recur(invariant)
	if len(newDefs) > 0 {
		// invariant becomes: implies(and(new_def_constraints...), invariant)
		var constraints []lg.Expr
		for _, d := range newDefs {
			if def, ok := d.(*il.Definition); ok {
				constraints = append(constraints, defToConstraint(def))
			}
		}
		if len(constraints) > 0 {
			newInv = &lg.Implies{T1: &lg.And{Terms: constraints}, T2: newInv}
		}
	}

	return newTrans, newInv
}

// tableLookupApp converts a function application f(a1,...,an) with finite-sort
// args into a table-lookup ITE chain.
func tableLookupApp(app *lg.Apply, funcSym *lg.Const, argSym func(lg.Sort) *lg.Const, newDefs *[]lg.Expr, recur func(lg.Expr) lg.Expr) lg.Expr {
	// For each argument, either use it directly or introduce a fresh symbol
	argSyms := make([]lg.Expr, len(app.Terms))
	constSets := make([][]lg.Expr, len(app.Terms))

	for i, x := range app.Terms {
		vals, err := SortValues(x.NodeSort())
		if err != nil {
			// Shouldn't happen since we checked isFiniteSort
			return app
		}
		constNodes := make([]lg.Expr, len(vals))
		for j, v := range vals {
			constNodes[j] = lg.NewConst(v, x.NodeSort())
		}
		constSets[i] = constNodes

		// Check if arg is already a constant in the sort values
		if c, ok := x.(*lg.Const); ok {
			isVal := false
			for _, v := range vals {
				if c.Name == v {
					isVal = true
					break
				}
			}
			if isVal {
				argSyms[i] = x
				constSets[i] = []lg.Expr{x}
				continue
			}
			argSyms[i] = x
		} else {
			sym := argSym(x.NodeSort())
			argSyms[i] = sym
			*newDefs = append(*newDefs, il.NewDefinition(sym, recur(x)))
		}
	}

	// Generate cartesian product of all argument values
	combos := cartesianProductNodes(constSets)
	if len(combos) == 0 {
		return app
	}

	// Build ITE chain: f(v0) for first combo, then ite(args==v, f(v), rest)
	var result lg.Expr
	for i, combo := range combos {
		fApp := lg.MustApply(funcSym, combo...)
		if i == 0 {
			result = fApp
		} else {
			// Build equality condition: args[0]==combo[0] && args[1]==combo[1] && ...
			var eqs []lg.Expr
			for j, arg := range argSyms {
				eqs = append(eqs, &lg.Eq{T1: arg, T2: combo[j]})
			}
			var cond lg.Expr
			if len(eqs) == 1 {
				cond = eqs[0]
			} else {
				cond = &lg.And{Terms: eqs}
			}
			result = &lg.Ite{Cond: cond, Then: fApp, Else: result}
		}
	}
	return result
}

// cartesianProductNodes computes the cartesian product of multiple slices of lg.Expr.
func cartesianProductNodes(sets [][]lg.Expr) [][]lg.Expr {
	if len(sets) == 0 {
		return [][]lg.Expr{{}}
	}
	first := sets[0]
	rest := cartesianProductNodes(sets[1:])
	var result [][]lg.Expr
	for _, v := range first {
		for _, r := range rest {
			combo := make([]lg.Expr, 1+len(r))
			combo[0] = v
			copy(combo[1:], r)
			result = append(result, combo)
		}
	}
	return result
}

// defToConstraint converts a definition to a constraint formula.
func defToConstraint(def *il.Definition) lg.Expr {
	result := il.DefinitionToConstraint(def)
	if xtracer.Enabled {
		lhsSort := def.Lhs.NodeSort()
		xtracer.Trace("module/clauses.go:262 defToConstraint lhsSort=%v resultType=%v", lhsSort, iu.ShortTypeName(result))
	}
	return result
}

// schemaMatch implements Python's Match class (ivy_mc.py:551-576).
// It supports backtracking unification of sorts and symbols.
type schemaMatch struct {
	mp    map[string]lg.Expr
	stack [][]string
}

func newSchemaMatch() *schemaMatch {
	return &schemaMatch{
		mp:    make(map[string]lg.Expr),
		stack: [][]string{{}},
	}
}

func (m *schemaMatch) add(key string, val lg.Expr) {
	m.mp[key] = val
	m.stack[len(m.stack)-1] = append(m.stack[len(m.stack)-1], key)
}

func (m *schemaMatch) push() {
	m.stack = append(m.stack, []string{})
}

func (m *schemaMatch) pop() {
	top := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	for _, k := range top {
		delete(m.mp, k)
	}
}

// unify tries to unify sort x with sort y.
// Python: if x not in self.map: self.add(x,y); return True
//
//	return self.map[x] == y
func (m *schemaMatch) unify(x, y lg.Sort) bool {
	key := sortName(x)
	if _, has := m.mp[key]; !has {
		if strings.HasSuffix(key, "_finite") && !isFiniteSort(y) {
			return false
		}
		m.add(key, y)
		return true
	}
	return sortName(m.mp[key].(lg.Sort)) == sortName(y)
}

func (m *schemaMatch) unifyLists(xl, yl []lg.Sort) bool {
	if len(xl) != len(yl) {
		return false
	}
	for i := range xl {
		if !m.unify(xl[i], yl[i]) {
			return false
		}
	}
	return true
}

// unwrapSort extracts an UninterpretedSort from a node, handling
// CompiledNode wrapping that occurs after schema compilation.
func unwrapSort(n ast.Node) (*lg.UninterpretedSort, bool) {
	if us, ok := n.(*lg.UninterpretedSort); ok {
		return us, true
	}
	if cn, ok := n.(*ast.CompiledNode); ok {
		if us, ok2 := cn.Node.(*lg.UninterpretedSort); ok2 {
			return us, true
		}
		if s, ok2 := cn.Node.(lg.Sort); ok2 {
			if us, ok3 := s.(*lg.UninterpretedSort); ok3 {
				return us, true
			}
		}
	}
	return nil, false
}

func sortName(s lg.Sort) string {
	if s == nil {
		return ""
	}
	return s.String()
}

// ExpandSchemata expands axiom schemata into axioms by matching schema
// premises against the sort constants and functions.
//
// Python: ivy_mc.py:638-656
func ExpandSchemata(mod *module.Module, sortConstants map[string][]*lg.Const, funs *iu.InsMap[string, *lg.Const]) []*ast.LabeledFormula {
	var result []*ast.LabeledFormula

	if mod.Schemata == nil {
		xtracer.Trace("mc.ExpandSchemata nSchemata=0 (nil) nSortConstants=%d nFuns=%d", len(sortConstants), funs.Len())
		return result
	}
	xtracer.Trace("mc.ExpandSchemata nSchemata=%d nSorts=%d nSortConstants=%d nFuns=%d", mod.Schemata.Len(), mod.Sig.Sorts.Len(), len(sortConstants), funs.Len())

	match := newSchemaMatch()
	// Python: for s in list(mod.sig.sorts.values()):
	//             if not il.is_function_sort(s): match.add(s, s)
	for _, s := range mod.Sig.Sorts.All() {
		if !il.IsFunctionSort(s) {
			match.add(sortName(s), s)
		}
	}

	for name, lfNode := range mod.Schemata.All() {
		xtracer.Trace("mc.ExpandSchemata entry name=%s", name)
		if strings.HasPrefix(name, "rec[") || strings.HasPrefix(name, "lep[") || strings.HasPrefix(name, "ind[") {
			continue
		}

		lf, ok := lfNode.(*ast.LabeledFormula)
		if !ok {
			continue
		}

		// Python: schema = lf.formula; conc = schema.args[-1]; prems = list(schema.args[:-1])
		sb, ok := lf.Formula.(*ast.SchemaBody)
		if !ok {
			continue
		}
		prems := sb.Prems()
		conc := sb.Conc()
		if conc == nil {
			continue
		}
		xtracer.Trace("mc.ExpandSchemata schema name=%s nPrems=%d", name, len(prems))

		boundSorts := make(map[string]bool)
		for _, p := range prems {
			if us, ok := unwrapSort(p); ok {
				boundSorts[sortName(us)] = true
			}
		}

		matchSchemaPrems(prems, sortConstants, funs, match, boundSorts, func(mp map[string]lg.Expr) {
			inst := applyMatch(mp, conc)
			result = append(result, mod.Cfg.AstCfg.NewLabeledFormula(
				mod.Cfg.AstCfg.NewAtom(name), inst))
		})
	}

	xtracer.Trace("mc.ExpandSchemata result nExpanded=%d", len(result))
	return result
}

// matchSchemaPrems is the faithful port of Python's match_schema_prems
// (ivy_mc.py:582-617). It uses a stack-based premise list (pop/append)
// matching Python's generator pattern.
func matchSchemaPrems(prems []ast.Node, sortConstants map[string][]*lg.Const, funs *iu.InsMap[string, *lg.Const], match *schemaMatch, boundSorts map[string]bool, callback func(map[string]lg.Expr)) {
	if len(prems) == 0 {
		result := make(map[string]lg.Expr, len(match.mp))
		for k, v := range match.mp {
			result[k] = v
		}
		callback(result)
		return
	}

	// Python: prem = prems.pop()
	prem := prems[len(prems)-1]
	prems = prems[:len(prems)-1]
	premForType := prem
	if cn, ok2 := prem.(*ast.CompiledNode); ok2 && cn.Node != nil {
		if n, ok3 := cn.Node.(ast.Node); ok3 {
			premForType = n
		}
	}
	xtracer.Trace("mc.matchSchemaPrems premType=%s nPremsLeft=%d", iu.ShortTypeName(premForType), len(prems))
	if cd, ok := prem.(*ast.ConstantDecl); ok {
		args := cd.Args()
		if len(args) == 0 {
			prems = append(prems, prem)
			return
		}
		var sym *lg.Const
		switch a := args[0].(type) {
		case *lg.Const:
			sym = a
		case *ast.CompiledNode:
			if c, ok2 := a.Node.(*lg.Const); ok2 {
				sym = c
			}
		}
		if sym == nil {
			prems = append(prems, prem)
			return
		}
		xtracer.Trace("mc.matchSchemaPrems sym=%s isFuncSort=%v", sym.Name, il.IsFunctionSort(sym.CSort))

		if il.IsFunctionSort(sym.CSort) {
			// Python: sorts = sym.sort.dom + (sym.sort.rng,)
			fs := sym.CSort.(*lg.FunctionSort)
			sorts := append(fs.Domain(), fs.Range())

			for _, f := range funs.All() {
				if !il.IsFunctionSort(f.CSort) {
					continue
				}
				ffs := f.CSort.(*lg.FunctionSort)
				fsorts := append(ffs.Domain(), ffs.Range())

				match.push()
				for _, s := range sorts {
					key := sortName(s)
					if !boundSorts[key] {
						if _, has := match.mp[key]; !has {
							match.add(key, s)
						}
					}
				}
				if match.unifyLists(sorts, fsorts) {
					match.add(sym.Name, f)
					matchSchemaPrems(prems, sortConstants, funs, match, boundSorts, callback)
				}
				match.pop()
			}
		} else {
			// Non-function constant premise
			symSortKey := sortName(sym.CSort)
			var cands []*lg.Const
			if _, mapped := match.mp[symSortKey]; mapped || !boundSorts[symSortKey] {
				lookupKey := symSortKey
				if mapped, ok := match.mp[symSortKey]; ok {
					lookupKey = sortName(mapped.(lg.Sort))
				}
				cands = sortConstants[lookupKey]
				xtracer.Trace("mc.matchSchemaPrems nonFunc sortKey=%s inMap=%v isBound=%v nCands=%d", symSortKey, match.mp[symSortKey] != nil, boundSorts[symSortKey], len(cands))
			} else {
				scKeys := make([]string, 0, len(sortConstants))
				for k := range sortConstants {
					scKeys = append(scKeys, k)
				}
				sort.Strings(scKeys)
				for _, k := range scKeys {
					cands = append(cands, sortConstants[k]...)
				}
				xtracer.Trace("mc.matchSchemaPrems nonFunc sortKey=%s inMap=%v isBound=%v nCands=%d (allSorts)", symSortKey, match.mp[symSortKey] != nil, boundSorts[symSortKey], len(cands))
			}
			for _, cand := range cands {
				match.push()
				if match.unify(sym.CSort, cand.CSort) {
					match.add(sym.Name, cand)
					matchSchemaPrems(prems, sortConstants, funs, match, boundSorts, callback)
				}
				match.pop()
			}
		}
	} else if _, ok := unwrapSort(prem); ok {
		// Python: just recurse without consuming
		matchSchemaPrems(prems, sortConstants, funs, match, boundSorts, callback)
	}

	// Python: prems.append(prem) — restore for backtracking
	prems = append(prems, prem)
}

// applyMatch applies a match substitution to a formula, performing beta
// reduction for matched function symbols.
// Python: ivy_mc.py:619-636
func applyMatch(mp map[string]lg.Expr, fmla ast.Node) lg.Expr {
	if cn, ok := fmla.(*ast.CompiledNode); ok {
		if e, ok2 := cn.Node.(lg.Expr); ok2 {
			return applyMatch(mp, e)
		}
		if n, ok2 := cn.Node.(ast.Node); ok2 {
			return applyMatch(mp, n)
		}
	}
	expr, ok := fmla.(lg.Expr)
	if !ok {
		return nil
	}

	children := expr.Children()
	args := make([]lg.Expr, len(children))
	for i, c := range children {
		args[i] = applyMatch(mp, c)
	}

	// Python: if il.is_app(fmla) and fmla.rep in match: func = match[fmla.rep]; return func(*args)
	if app, ok := expr.(*lg.Apply); ok {
		if f, ok := app.Func.(*lg.Const); ok {
			if repl, has := mp[f.Name]; has {
				if rc, ok := repl.(*lg.Const); ok {
					return lg.NewApplyUnchecked(rc, args...)
				}
			}
		}
		return lg.CloneApplyTerms(app, args)
	}

	// Python: elif il.is_binder(fmla): vs = [apply_match(match,v) for v in fmla.variables]
	//          return fmla.clone_binder(vs, apply_match(match, fmla.body))
	if fa, ok := expr.(*lg.ForAll); ok {
		newVars := make([]*lg.Variable, len(fa.Variables))
		for i, v := range fa.Variables {
			nv := applyMatch(mp, v)
			if rv, ok := nv.(*lg.Variable); ok {
				newVars[i] = rv
			} else {
				newVars[i] = v
			}
		}
		return &lg.ForAll{Variables: newVars, Body: args[0]}
	}
	if ex, ok := expr.(*lg.Exists); ok {
		newVars := make([]*lg.Variable, len(ex.Variables))
		for i, v := range ex.Variables {
			nv := applyMatch(mp, v)
			if rv, ok := nv.(*lg.Variable); ok {
				newVars[i] = rv
			} else {
				newVars[i] = v
			}
		}
		return &lg.Exists{Variables: newVars, Body: args[0]}
	}

	// Python: elif il.is_variable(fmla): return Variable(fmla.name, match.get(fmla.sort, fmla.sort))
	if v, ok := expr.(*lg.Variable); ok {
		newSort := v.VSort
		if mapped, has := mp[sortName(v.VSort)]; has {
			if ms, ok := mapped.(lg.Sort); ok {
				newSort = ms
			}
		}
		return &lg.Variable{Name: v.Name, VSort: newSort}
	}

	// Bare zero-arity constants.
	// Python: zero-arity constants are Apply(rep=x, args=[]) and hit the
	// is_app branch. In Go they are bare *lg.Const nodes.
	if c, ok := expr.(*lg.Const); ok {
		if repl, has := mp[c.Name]; has {
			return repl
		}
		return c
	}

	// Default: clone with recursed args
	astArgs := make([]ast.Node, len(args))
	for i, a := range args {
		astArgs[i] = a
	}
	result := expr.Clone(astArgs)
	if re, ok := result.(lg.Expr); ok {
		return re
	}
	return expr
}

// InstantiateAxioms performs pattern-based eager axiom instantiation.
// For each universally quantified axiom, finds a trigger (a function application
// covering all bound variables), then matches that trigger against all
// subexpressions in the transition relation and invariant.
//
// Python: ivy_mc.py:659-745
func InstantiateAxioms(mod *module.Module, stVars []string, trans *module.Clauses, invariant lg.Expr, sortConstants map[string][]*lg.Const, funs *iu.InsMap[string, *lg.Const], iuCfg *iu.IvyUtilsConfig) []lg.Expr {
	// Expand schemata into axioms
	expandedAxioms := ExpandSchemata(mod, sortConstants, funs)

	// Combine with existing labeled axioms
	var axioms []*ast.LabeledFormula
	axioms = append(axioms, mod.LabeledAxioms...)
	axioms = append(axioms, expandedAxioms...)

	// Get triggers for each quantified axiom
	type trigEntry struct {
		trigger lg.Expr
		axiom   *ast.LabeledFormula
	}
	var triggers []trigEntry

	for _, ax := range axioms {
		fmla := ax.Formula.(lg.Expr)
		vars := lu.FreeVariablesList(fmla)
		if len(vars) > 0 {
			trig := getTrigger(fmla, vars)
			if trig != nil {
				triggers = append(triggers, trigEntry{trigger: trig, axiom: ax})
			}
		}
	}

	if xtracer.Enabled {
		xtracer.Trace("mc.InstantiateAxioms nAxioms=%d nTriggers=%d", len(axioms), len(triggers))

		for i, te := range triggers {
			xtracer.Trace("mc.InstantiateAxioms trigger[%d] trig=%s axiom=%s", i, te.trigger, te.axiom.Formula)
		}
	}

	// Match triggers against all expressions in trans and invariant
	instSet := make(map[lg.NodeKey]bool)
	var instList []lg.Expr
	instCount := 0

	var scanExpr func(expr lg.Expr)
	scanExpr = func(expr lg.Expr) {
		for _, child := range expr.Children() {
			scanExpr(child)
		}
		for ti, te := range triggers {
			mp := make(map[string]lg.Expr)
			if matchNodes(te.trigger, expr, mp) {
				raw := lu.SubstituteByName(te.axiom.Formula.(lg.Expr), mp)
				inst := normalize(raw, iuCfg)
				instKey := inst.Sexp()
				if !instSet[instKey] {
					instSet[instKey] = true
					instList = append(instList, inst)
					mpStr := fmtMp(mp)
					xtracer.Trace("mc.InstantiateAxioms addUnique[%d] trig[%d] axiom=%s mp={%s} preNorm=%s postNorm=%s", instCount, ti, te.axiom.Formula, mpStr, raw, inst)
					instCount++
				}
			}
		}
	}

	// Scan transition relation defs and fmlas
	if trans != nil {
		for _, df := range trans.Defs {
			scanExpr(df)
		}
		for _, fmla := range trans.Fmlas {
			scanExpr(fmla)
		}
	}
	scanExpr(invariant)

	xtracer.Trace("mc.InstantiateAxioms result nUnique=%d", len(instList))
	for i, inst := range instList {
		xtracer.Trace("mc.InstantiateAxioms inst[%d] %s", i, inst)
	}
	return instList
}

// getTrigger finds a trigger expression in a formula that covers all the given variables.
// A trigger is a function application or equality that contains all bound variables.
//
// Python: ivy_mc.py:674-684
func getTrigger(expr lg.Expr, vars []*lg.Variable) lg.Expr {
	if il.IsQuantifier(expr) || il.IsVariable(expr) {
		return nil
	}

	// Check children first (depth-first)
	for _, child := range expr.Children() {
		r := getTrigger(child, vars)
		if r != nil {
			return r
		}
	}

	// Check if this expression is a trigger
	if il.IsApp(expr) || isEq(expr) {
		exprVars := lu.FreeVariablesList(expr)
		if containsAllVars(exprVars, vars) {
			return expr
		}
	}
	return nil
}

// matchNodes attempts to match pattern against expr, filling in the substitution map.
// Variables in the pattern are matched to subexpressions in expr.
//
// Python: ivy_mc.py:701-725
func matchNodes(pat, expr lg.Expr, mp map[string]lg.Expr) bool {
	if v, ok := pat.(*lg.Variable); ok {
		if existing, ok := mp[v.Name]; ok {
			return existing.Sexp() == expr.Sexp()
		}
		mp[v.Name] = expr
		return true
	}

	if app, ok := pat.(*lg.Apply); ok {
		eapp, ok2 := expr.(*lg.Apply)
		if !ok2 {
			return false
		}
		pFunc, ok3 := app.Func.(*lg.Const)
		eFunc, ok4 := eapp.Func.(*lg.Const)
		if !ok3 || !ok4 || pFunc.Name != eFunc.Name {
			return false
		}
		if len(app.Terms) != len(eapp.Terms) {
			return false
		}
		for i := range app.Terms {
			if !matchNodes(app.Terms[i], eapp.Terms[i], mp) {
				return false
			}
		}
		return true
	}

	if il.IsQuantifier(pat) {
		return false
	}

	// For equality, try both orderings
	if peq, ok := pat.(*lg.Eq); ok {
		eeq, ok2 := expr.(*lg.Eq)
		if !ok2 {
			return false
		}
		// Check sort compatibility
		if !sameSortKey(peq.T1.NodeSort(), eeq.T1.NodeSort()) {
			return false
		}
		// Try forward
		save := copyMap(mp)
		if matchNodes(peq.T1, eeq.T1, mp) && matchNodes(peq.T2, eeq.T2, mp) {
			return true
		}
		// Try reversed
		for k := range mp {
			delete(mp, k)
		}
		for k, v := range save {
			mp[k] = v
		}
		return matchNodes(peq.T1, eeq.T2, mp) && matchNodes(peq.T2, eeq.T1, mp)
	}

	// Generic: same type, same number of children, match all children
	if fmt.Sprintf("%T", pat) != fmt.Sprintf("%T", expr) {
		return false
	}
	pc := pat.Children()
	ec := expr.Children()
	if len(pc) != len(ec) {
		return false
	}
	for i := range pc {
		if !matchNodes(pc[i], ec[i], mp) {
			return false
		}
	}
	return true
}

func fmtMp(mp map[string]lg.Expr) string {
	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s->%s", k, mp[k])
	}
	return strings.Join(parts, " ")
}

// Helper functions

func isEq(n lg.Expr) bool {
	_, ok := n.(*lg.Eq)
	return ok
}

func containsAllVars(have []*lg.Variable, need []*lg.Variable) bool {
	haveSet := make(map[string]bool, len(have))
	for _, v := range have {
		haveSet[v.Name] = true
	}
	for _, v := range need {
		if !haveSet[v.Name] {
			return false
		}
	}
	return true
}

func sameSortKey(a, b lg.Sort) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.String() == b.String()
}

func copyMap(m map[string]lg.Expr) map[string]lg.Expr {
	cp := make(map[string]lg.Expr, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
