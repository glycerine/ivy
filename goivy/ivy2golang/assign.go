package ivy2golang

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// assignBoundsExpr implements Python's bexpr trick from emit_assign
// (ivy_to_cpp.py:3717-3720). When RHS has the shape Ite(cond, then, lhs)
// and cond does not mention the modified symbol, cond can tighten the loop
// bounds used for both phases of a quantified assignment.
func (g *Generator) assignBoundsExpr(a *goivy.LogicAssignAction) goivy.Expr {
	ite, ok := a.RHS.(*goivy.LogicIte)
	if !ok {
		return nil
	}
	if !ite.Else.Equal(a.LHS) {
		return nil
	}
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

func (g *Generator) actionsCfg() *goivy.ActionsConfig {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil {
		return nil
	}
	return g.Mod.Cfg.ActCfg
}

func (g *Generator) assignmentLoopHeadersBounded(lhs goivy.Expr, body goivy.Expr) ([]string, bool) {
	vs := g.assignmentLoopVars(lhs)
	headers := make([]string, len(vs))
	for i, v := range vs {
		if v == nil {
			return nil, false
		}
		if body != nil && g.goIsAnyIntegerType(v.VSort) {
			var bes []boundExpr
			g.matchBoundExprs(v, body, true, &bes)
			if len(bes) > 0 {
				lo, hi, err := g.getBounds(v, vs[i+1:], body, true)
				if err == nil {
					header, ok, herr := g.goBoundedLoopHeaderForSort(v.VSort, goName(v.Name), lo, hi)
					if herr != nil {
						return nil, false
					}
					if ok {
						headers[i] = header
						continue
					}
				}
			}
		}
		if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, goName(v.Name)); err != nil || ok {
			if err != nil {
				return nil, false
			}
			headers[i] = header
			continue
		}
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			return nil, false
		}
		headers[i] = fmt.Sprintf("for _, %s := range []%s{%s} {", goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", "))
	}
	return headers, true
}

func (g *Generator) assignmentLoopVars(lhs goivy.Expr) []*goivy.LogicVariable {
	var out []*goivy.LogicVariable
	seen := map[string]bool{}
	var walk func(goivy.Expr)
	walk = func(e goivy.Expr) {
		switch n := e.(type) {
		case *goivy.LogicVariable:
			if n != nil && !seen[n.Name] && !g.isLocal(n.Name) {
				seen[n.Name] = true
				out = append(out, n)
			}
		case *goivy.Const:
			if n != nil && goParamConstName(n.Name) && !seen[n.Name] && !g.isLocal(n.Name) {
				seen[n.Name] = true
				out = append(out, &goivy.LogicVariable{Name: n.Name, VSort: n.CSort})
			}
		case *goivy.Apply:
			for _, term := range n.Terms {
				walk(term)
			}
		default:
			if n == nil {
				return
			}
			for _, child := range n.Children() {
				walk(child)
			}
		}
	}
	walk(lhs)
	return out
}

func (g *Generator) emitAssignLoopHeaders(w *goWriter, vs []*goivy.LogicVariable, headers []string) {
	for i, v := range vs {
		if v != nil {
			g.addLocal(v.Name)
		}
		w.open(headers[i])
	}
}

func (g *Generator) closeAssignLoopHeaders(w *goWriter, headers []string) {
	for range headers {
		w.close("")
	}
}

func (g *Generator) emitAssignTwoPhase(w *goWriter, a *goivy.LogicAssignAction, vs []*goivy.LogicVariable) {
	loc := a.GetLineno()
	sorts := make([]goivy.Sort, 0, len(vs)+1)
	for _, v := range vs {
		sorts = append(sorts, v.VSort)
	}
	sorts = append(sorts, a.RHS.NodeSort())
	tmpSort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported temp function sort: %s", err.Error())
		return
	}
	body := g.assignBoundsExpr(a)
	headers, ok := g.assignmentLoopHeadersBounded(a.LHS, body)
	if !ok {
		g.emitAssignLarge(w, a, vs)
		return
	}

	tmpName := g.nextTemp("__ivy_tmp")
	w.linef("%s := %s", goName(tmpName), g.goFunctionStorageInit(sorts[:len(sorts)-1], a.RHS.NodeSort()))

	tmpArgs := make([]goivy.Expr, len(vs))
	for i, v := range vs {
		tmpArgs[i] = v
	}
	tmpSym := goivy.NewConst(tmpName, tmpSort)
	tmpLHS := goivy.NewApplyUnchecked(tmpSym, tmpArgs...)

	g.pushScope()
	g.addLocalSort(tmpName, tmpSort)
	g.emitAssignLoopHeaders(w, vs, headers)
	rhsCode, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported assignment rhs: %s", err.Error())
		g.closeAssignLoopHeaders(w, headers)
		g.popScope()
		return
	}
	if call, ok, err := g.goStorageSet(tmpLHS, rhsCode); ok || err != nil {
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported temp lhs: %s", err.Error())
			g.closeAssignLoopHeaders(w, headers)
			g.popScope()
			return
		}
		w.line(call)
	} else {
		tmpCode, err := g.emitExpr(tmpLHS)
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported temp lhs: %s", err.Error())
			g.closeAssignLoopHeaders(w, headers)
			g.popScope()
			return
		}
		w.linef("%s = %s", tmpCode, rhsCode)
	}
	g.closeAssignLoopHeaders(w, headers)

	g.emitAssignLoopHeaders(w, vs, headers)
	tmpRHS, err := g.emitExpr(tmpLHS)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported temp rhs: %s", err.Error())
		g.closeAssignLoopHeaders(w, headers)
		g.popScope()
		return
	}
	tmpRHS = g.maybeVariantUpcastExpr(a.LHS.NodeSort(), a.RHS.NodeSort(), tmpRHS)
	if call, ok, err := g.goStorageSet(a.LHS, tmpRHS); ok || err != nil {
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported assignment lhs: %s", err.Error())
			g.closeAssignLoopHeaders(w, headers)
			g.popScope()
			return
		}
		w.line(call)
	} else {
		lhsCode, err := g.emitExpr(a.LHS)
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported assignment lhs: %s", err.Error())
			g.closeAssignLoopHeaders(w, headers)
			g.popScope()
			return
		}
		w.linef("%s = %s", lhsCode, tmpRHS)
	}
	g.closeAssignLoopHeaders(w, headers)
	g.popScope()
}

func (g *Generator) emitAssignLarge(w *goWriter, a *goivy.LogicAssignAction, lhsVs []*goivy.LogicVariable) {
	loc := a.GetLineno()
	lhsApp, ok := a.LHS.(*goivy.Apply)
	if !ok {
		g.unsupportedAt(w, loc, "emit_assign_large requires apply-shaped LHS, got %T", a.LHS)
		return
	}
	lhsFn, ok := lhsApp.Func.(*goivy.Const)
	if !ok {
		g.unsupportedAt(w, loc, "emit_assign_large requires Const function on LHS, got %T", lhsApp.Func)
		return
	}
	fnSort, ok := lhsFn.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		g.unsupportedAt(w, loc, "emit_assign_large LHS function lacks FunctionSort: %s", lhsFn.Name)
		return
	}
	dom := fnSort.Domain()
	if len(dom) != len(lhsApp.Terms) {
		g.unsupportedAt(w, loc, "emit_assign_large arity mismatch: %d vs %d", len(dom), len(lhsApp.Terms))
		return
	}
	vsPrime := make([]*goivy.LogicVariable, len(dom))
	var eqs []goivy.Expr
	seenVars := map[string]bool{}
	for i, term := range lhsApp.Terms {
		if v, ok := term.(*goivy.LogicVariable); ok {
			if !seenVars[v.Name] {
				seenVars[v.Name] = true
				vsPrime[i] = v
				continue
			}
		}
		fresh, err := goivy.NewVariable(fmt.Sprintf("X%d", i), dom[i])
		if err != nil {
			g.unsupportedAt(w, loc, "emit_assign_large failed to synthesize bound variable: %s", err.Error())
			return
		}
		vsPrime[i] = fresh
		eq, err := goivy.NewEq(term, fresh)
		if err != nil {
			g.unsupportedAt(w, loc, "emit_assign_large equality build failed: %s", err.Error())
			return
		}
		eqs = append(eqs, eq)
	}

	var expr goivy.Expr = a.RHS
	if len(eqs) > 0 {
		var cond goivy.Expr
		if len(eqs) == 1 {
			cond = eqs[0]
		} else {
			and, err := goivy.NewAnd(eqs...)
			if err != nil {
				g.unsupportedAt(w, loc, "emit_assign_large And build failed: %s", err.Error())
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
			g.unsupportedAt(w, loc, "emit_assign_large Ite build failed: %s", err.Error())
			return
		}
		expr = ite
	}

	name := lhsFn.Name
	sort, owner := g.storageSortAndOwner(name)
	if sort == nil {
		g.unsupportedAt(w, loc, "emit_assign_large unknown storage symbol %s", name)
		return
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || !g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		g.unsupportedAt(w, loc, "emit_assign_large requires large function storage: %s", name)
		return
	}
	base := goName(name)
	if owner != "" {
		base = owner + "." + base
	}
	oldName := g.nextTemp("__ivy_old_" + goName(name))
	w.linef("%s := %s", oldName, base)
	w.linef("_ = %s", oldName)
	w.linef("%s = %s", base, g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	keyName := g.nextTemp("__ivy_key")
	st := g.goFunctionStorageFor(fs.Domain(), fs.Range())
	w.open(fmt.Sprintf("%s.base = func(%s %s) %s {", base, keyName, st.KeyType, st.RangeType))
	g.pushScope()
	g.addLocalSort(oldName, sort)
	for i, v := range vsPrime {
		if v == nil {
			continue
		}
		field := keyName
		if len(vsPrime) > 1 {
			field = fmt.Sprintf("%s.A%d", keyName, i)
		}
		w.linef("%s := %s", goName(v.Name), field)
		w.linef("_ = %s", goName(v.Name))
		g.addLocal(v.Name)
	}
	prevAliases := g.exprAliases
	nextAliases := make(map[string]goivy.Expr, len(prevAliases)+1)
	for k, v := range prevAliases {
		nextAliases[k] = v
	}
	nextAliases[name] = goivy.NewConst(oldName, sort)
	g.exprAliases = nextAliases
	body, err := g.emitExpr(expr)
	g.exprAliases = prevAliases
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported thunk assignment body: %s", err.Error())
		g.popScope()
		w.close("")
		return
	}
	w.linef("return %s", body)
	g.popScope()
	w.close("")
	if summary, ok := g.sparseSupportAssignment(name, lhsApp.Terms, a.RHS); ok && g.sparseSupportRels()[name] && isBooleanSort(fs.Range()) {
		g.emitSparseSupportUpdates(w, base, oldName, st, summary)
	}
	_ = lhsVs
}

func (g *Generator) emitSparseSupportUpdates(w *goWriter, base, oldName string, st goFunctionStorage, summary sparseSupportSummary) {
	if summary.PreserveOld {
		key := g.nextTemp("__ivy_support_key")
		val := g.nextTemp("__ivy_support_val")
		w.open(fmt.Sprintf("for %s, %s := range %s.overrides {", key, val, oldName))
		w.open(fmt.Sprintf("if %s {", val))
		w.linef("%s.Set(%s, true)", base, key)
		w.close("")
		w.close("")
	}
	for _, point := range summary.Points {
		if len(point) != len(st.Domain) {
			continue
		}
		args := make([]string, len(point))
		ok := true
		for i, expr := range point {
			code, err := g.emitExpr(expr)
			if err != nil {
				g.unsupported(w, "unsupported sparse support point: %s", err.Error())
				ok = false
				break
			}
			args[i] = code
		}
		if !ok {
			continue
		}
		w.linef("%s.Set(%s, true)", base, g.goMapKeyValue(st.Domain, args))
	}
}
