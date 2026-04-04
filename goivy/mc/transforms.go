package mc

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
)

// ElimIte eliminates ITEs over non-finite sorts by introducing fresh variables
// and constraints. An ITE (x if c else y) where x has non-finite sort becomes
// a fresh variable v with added constraint (v=x if c else v=y).
// As an optimization, (z = (x if c else y)) → (z=x if c else z=y) directly.
//
// Python: ivy_mc.py:757-772
func ElimIte(expr lg.Expr, cnsts *[]lg.Expr) lg.Expr {
	switch t := expr.(type) {
	case *lg.Ite:
		if !isFiniteSort(t.Then.NodeSort()) {
			// Create fresh variable
			name := fmt.Sprintf("__ite[%d]", NextIteCtr())
			v := lg.NewConst(name, t.Then.NodeSort())
			// Add constraint: ite(c, v=x, v=y)
			cElim := ElimIte(t.Cond, cnsts)
			eqThen := ElimIte(&lg.Eq{T1: v, T2: t.Then}, cnsts)
			eqElse := ElimIte(&lg.Eq{T1: v, T2: t.Else}, cnsts)
			*cnsts = append(*cnsts, &lg.Ite{Cond: cElim, Then: eqThen, Else: eqElse})
			return v
		}
	case *lg.Eq:
		// Optimization: (z = ite(c,x,y)) → ite(c, z=x, z=y)
		if ite, ok := t.T2.(*lg.Ite); ok {
			if !isFiniteSort(ite.Then.NodeSort()) {
				cElim := ElimIte(ite.Cond, cnsts)
				eqThen := ElimIte(&lg.Eq{T1: t.T1, T2: ite.Then}, cnsts)
				eqElse := ElimIte(&lg.Eq{T1: t.T1, T2: ite.Else}, cnsts)
				return &lg.Ite{Cond: cElim, Then: eqThen, Else: eqElse}
			}
		}
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, child := range children {
		nc := ElimIte(child, cnsts)
		newChildren[i] = nc
		if nc != child {
			changed = true
		}
	}
	if !changed {
		return expr
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
		changed := false
		for i, c := range children {
			nc[i] = recur(c)
			if nc[i] != c {
				changed = true
			}
		}
		if !changed {
			return expr
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
		fApp := &lg.Apply{Func: funcSym, Terms: combo}
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

// defToConstraint converts a definition to a constraint (equality).
func defToConstraint(def *il.Definition) lg.Expr {
	return &lg.Eq{T1: def.Lhs, T2: def.Rhs}
}

// ExpandSchemata expands axiom schemata into axioms by matching schema
// premises against the sort constants and functions.
//
// Python: ivy_mc.py:637-655
func ExpandSchemata(mod *module.Module, sortConstants map[string][]*lg.Const, funs map[string]bool) []*ast.LabeledFormula {
	var result []*ast.LabeledFormula

	if module.Schemata == nil {
		return result
	}

	// For each schema, try to match its premises
	for name, lf := range module.Schemata {
		// Skip recursive/inductive schemata
		if len(name) >= 4 && (name[:4] == "rec[" || name[:4] == "lep[" || name[:4] == "ind[") {
			continue
		}

		// The schema is stored as interface{} — try to extract it
		schema, ok := extractSchemaFormula(lf)
		if !ok || schema == nil {
			continue
		}

		children := schema.Children()
		if len(children) < 2 {
			continue
		}

		conc := children[len(children)-1]
		prems := children[:len(children)-1]

		// Collect bound sorts from premises
		boundSorts := make(map[string]bool)
		for _, prem := range prems {
			if us, ok := prem.(*lg.UninterpretedSort); ok {
				boundSorts[us.Name] = true
			}
		}

		// For each matching of premises, instantiate conclusion
		matchSchemaPremsNode(prems, sortConstants, funs, boundSorts, func(mp map[string]lg.Expr) {
			inst := lu.SubstituteByName(conc, mp)
			result = append(result, module.Cfg.AstCfg.NewLabeledFormula(nil, inst))
		})
	}

	return result
}

// extractSchemaFormula attempts to extract a lg.Expr formula from a schema interface{}.
func extractSchemaFormula(lf interface{}) (lg.Expr, bool) {
	switch t := lf.(type) {
	case *ast.LabeledFormula:
		if t.Formula == nil {
			return nil, false
		}
		return t.Formula.(lg.Expr), true
	case lg.Expr:
		return t, true
	}
	return nil, false
}

// matchSchemaPremsNode tries to match schema premises against available constants/functions.
func matchSchemaPremsNode(prems []lg.Expr, sortConstants map[string][]*lg.Const, funs map[string]bool, boundSorts map[string]bool, callback func(map[string]lg.Expr)) {
	mp := make(map[string]lg.Expr)
	matchSchemaPremsRec(prems, 0, sortConstants, funs, boundSorts, mp, callback)
}

func matchSchemaPremsRec(prems []lg.Expr, idx int, sortConstants map[string][]*lg.Const, funs map[string]bool, boundSorts map[string]bool, mp map[string]lg.Expr, callback func(map[string]lg.Expr)) {
	if idx >= len(prems) {
		// All premises matched, call back with copy of map
		result := make(map[string]lg.Expr, len(mp))
		for k, v := range mp {
			result[k] = v
		}
		callback(result)
		return
	}

	prem := prems[idx]

	// If premise is an uninterpreted sort, try matching to known sorts
	if us, ok := prem.(*lg.UninterpretedSort); ok {
		for sortName := range sortConstants {
			old, hadOld := mp[us.Name]
			mp[us.Name] = lg.NewConst(sortName, nil)
			matchSchemaPremsRec(prems, idx+1, sortConstants, funs, boundSorts, mp, callback)
			if hadOld {
				mp[us.Name] = old
			} else {
				delete(mp, us.Name)
			}
		}
		return
	}

	// For other premises (variables), try matching to constants
	if v, ok := prem.(*lg.Variable); ok {
		sortKey := sortKeyStr(v.VSort)
		consts := sortConstants[sortKey]
		for _, c := range consts {
			old, hadOld := mp[v.Name]
			mp[v.Name] = c
			matchSchemaPremsRec(prems, idx+1, sortConstants, funs, boundSorts, mp, callback)
			if hadOld {
				mp[v.Name] = old
			} else {
				delete(mp, v.Name)
			}
		}
		return
	}

	// For function-typed premises, try matching to function symbols
	if c, ok := prem.(*lg.Const); ok {
		if il.IsFunctionSort(c.CSort) {
			for funName := range funs {
				old, hadOld := mp[c.Name]
				mp[c.Name] = lg.NewConst(funName, c.CSort)
				matchSchemaPremsRec(prems, idx+1, sortConstants, funs, boundSorts, mp, callback)
				if hadOld {
					mp[c.Name] = old
				} else {
					delete(mp, c.Name)
				}
			}
			return
		}
	}

	// Default: skip premise
	matchSchemaPremsRec(prems, idx+1, sortConstants, funs, boundSorts, mp, callback)
}

// InstantiateAxioms performs pattern-based eager axiom instantiation.
// For each universally quantified axiom, finds a trigger (a function application
// covering all bound variables), then matches that trigger against all
// subexpressions in the transition relation and invariant.
//
// Python: ivy_mc.py:659-745
func InstantiateAxioms(mod *module.Module, stVars []string, trans *module.Clauses, invariant lg.Expr, sortConstants map[string][]*lg.Const, funs map[string]bool) []lg.Expr {
	// Expand schemata into axioms
	expandedAxioms := ExpandSchemata(mod, sortConstants, funs)

	// Combine with existing labeled axioms
	var axioms []*ast.LabeledFormula
	axioms = append(axioms, module.LabeledAxioms...)
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

	// Match triggers against all expressions in trans and invariant
	instSet := make(map[string]bool)
	var instList []lg.Expr

	var scanExpr func(expr lg.Expr)
	scanExpr = func(expr lg.Expr) {
		for _, child := range expr.Children() {
			scanExpr(child)
		}
		for _, te := range triggers {
			mp := make(map[string]lg.Expr)
			if matchNodes(te.trigger, expr, mp) {
				inst := lu.SubstituteByName(te.axiom.Formula.(lg.Expr), mp)
				instKey := fmt.Sprint(inst)
				if !instSet[instKey] {
					instSet[instKey] = true
					instList = append(instList, inst)
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
			return fmt.Sprint(existing) == fmt.Sprint(expr)
		}
		mp[v.Name] = expr
		return true
	}

	if app, ok := pat.(*lg.Apply); ok {
		eapp, ok2 := expr.(*lg.Apply)
		if !ok2 {
			return false
		}
		// Match function symbols
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
