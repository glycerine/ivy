package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type boundExpr struct {
	app *goivy.Apply
	neg bool
}

// matchBoundExprs mirrors ivy2cpp's port of Python get_bound_exprs. It
// collects inequality atoms that may bound v0, while tracking quantifier
// polarity through Not, literals, Implies, Or, and And.
func (g *Generator) matchBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]boundExpr) {
	if v0 == nil || body == nil {
		return
	}
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchBoundExprs(v0, not.Body, !exists, res)
		return
	}
	if lit, ok := body.(*goivy.LogicLiteral); ok {
		nextExists := exists
		if lit.Polarity == 0 {
			nextExists = !exists
		}
		g.matchBoundExprs(v0, lit.Atom, nextExists, res)
		return
	}
	app, isApp := body.(*goivy.Apply)
	if isApp {
		switch goivy.ExprName(app.Func) {
		case "<", "<=", ">", ">=":
			*res = append(*res, boundExpr{app: app, neg: !exists})
		}
	}
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchBoundExprs(v0, imp.T1, !exists, res)
			g.matchBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if !isApp || !containsVariableByName(app.Terms, v0.Name) {
		return
	}
	def, ok := g.definitionByName(goivy.ExprName(app.Func))
	if !ok || def.RHS == nil {
		return
	}
	if !allArgsVariable(def.Params) || len(def.Params) != len(app.Terms) {
		return
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range def.Params {
		pv, isVar := p.(*goivy.LogicVariable)
		if !isVar {
			return
		}
		subs[goivy.Key(pv)] = app.Terms[i]
	}
	substituted, err := goivy.Substitute(def.RHS, subs)
	if err != nil {
		return
	}
	g.matchBoundExprs(v0, substituted, exists, res)
}

func (g *Generator) sortHasNegativeValues(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "int"
}

func nameOfTerm(t goivy.Expr) string {
	switch x := t.(type) {
	case *goivy.LogicVariable:
		return x.Name
	case *goivy.Const:
		return x.Name
	default:
		return ""
	}
}

func nameIn(vars []*goivy.LogicVariable, n string) bool {
	for _, v := range vars {
		if v != nil && v.Name == n {
			return true
		}
	}
	return false
}

func quantVarDiagnostic(v *goivy.LogicVariable) string {
	if v == nil {
		return "<nil>"
	}
	if name := sortName(v.VSort); name != "" {
		return v.Name + ":" + name
	}
	return v.Name
}

func (g *Generator) getBounds(v0 *goivy.LogicVariable, others []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, string, error) {
	var bes []boundExpr
	g.matchBoundExprs(v0, body, exists, &bes)
	var los, his []string
	for _, be := range bes {
		op := goivy.ExprName(be.app.Func)
		strict := op == "<" || op == ">"
		args := be.app.Terms
		if op == ">" || op == ">=" {
			args = []goivy.Expr{args[1], args[0]}
		}
		if be.neg {
			strict = !strict
			args = []goivy.Expr{args[1], args[0]}
		}
		if len(args) != 2 {
			continue
		}
		ln := nameOfTerm(args[0])
		rn := nameOfTerm(args[1])
		if ln == v0.Name && rn != v0.Name && !nameIn(others, rn) {
			e, err := g.emitExpr(args[1])
			if err != nil {
				return "", "", err
			}
			if strict {
				his = append(his, e)
			} else {
				his = append(his, "("+e+")+1")
			}
		}
		if rn == v0.Name && ln != v0.Name && !nameIn(others, ln) {
			e, err := g.emitExpr(args[0])
			if err != nil {
				return "", "", err
			}
			if strict {
				los = append(los, "("+e+")+1")
			} else {
				los = append(los, e)
			}
		}
	}
	if !g.sortHasNegativeValues(v0.VSort) {
		los = append(los, "0")
	}
	if card := g.sortCard(v0.VSort); card > 0 {
		his = append(his, strconv.Itoa(card))
	}
	if rs, ok := g.rangeSortFor(v0.VSort); ok {
		lo, hi, ok2 := numericRangeBounds(rs)
		if ok2 {
			los = append(los, strconv.Itoa(lo))
			his = append(his, "("+strconv.Itoa(hi)+")+1")
		}
	}
	if len(los) == 0 {
		return "", "", fmt.Errorf("ivy2golang: cannot find a lower bound for %s", quantVarDiagnostic(v0))
	}
	if len(his) == 0 {
		if hi, ok := g.sortCardinalityAttr(v0.VSort); ok {
			his = append(his, hi)
		} else {
			return "", "", fmt.Errorf("ivy2golang: cannot find an upper bound for %s", quantVarDiagnostic(v0))
		}
	}
	return los[0], his[0], nil
}

func (g *Generator) sortCardinalityAttr(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil || g.Mod.Cfg.IuCfg == nil {
		return "", false
	}
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return "", false
	}
	name := sortName(s)
	if name == "" {
		return "", false
	}
	attrKey := g.Mod.Cfg.IuCfg.ComposeNames(name, "cardinality")
	val, ok := g.Mod.Attributes.Get2(attrKey)
	if !ok {
		return "", false
	}
	var rep string
	switch v := val.(type) {
	case goivy.Expr:
		rep = goivy.ExprName(v)
		if rep == "" {
			rep = string(v.Sexp())
		}
	case interface{ Relname() string }:
		rep = v.Relname()
	case string:
		rep = v
	default:
		return "", false
	}
	if rep == "" {
		return "", false
	}
	return goName(g.Mod.Cfg.IuCfg.ComposeNames(name, rep)), true
}

func (g *Generator) getAllBounds(vars []*goivy.LogicVariable, body goivy.Expr, exists bool) ([][2]string, error) {
	if len(vars) == 0 {
		return nil, nil
	}
	res := make([][2]string, 0, len(vars))
	for i, v := range vars {
		lo, hi, err := g.getBounds(v, vars[i+1:], body, exists)
		if err != nil {
			return nil, err
		}
		res = append(res, [2]string{lo, hi})
	}
	return res, nil
}

func (g *Generator) goIsAnyIntegerType(s goivy.Sort) bool {
	switch st := s.(type) {
	case *goivy.RangeSort:
		return true
	case *goivy.LogicEnumeratedSort:
		return isNumericEnum(st)
	case *goivy.UninterpretedSort:
		if _, ok := g.rangeSortFor(st); ok {
			return true
		}
		text, ok := g.sortInterpString(st)
		return ok && (text == "int" || text == "nat")
	default:
		return false
	}
}

func (g *Generator) goBoundedLoopHeaderForSort(s goivy.Sort, name, lo, hi string) (string, bool, error) {
	if lo == "" || hi == "" {
		return "", false, fmt.Errorf("ivy2golang: empty bounds for %s", name)
	}
	switch st := s.(type) {
	case *goivy.LogicEnumeratedSort:
		if st.Name != "" && !isNumericEnum(st) {
			return "", false, nil
		}
	}
	return fmt.Sprintf("for %s := %s; %s < %s; %s++ {", name, lo, name, hi, name), true, nil
}

func (g *Generator) goFiniteLoopHeaderForSort(s goivy.Sort, name string) (string, bool, error) {
	if st, ok := s.(*goivy.LogicEnumeratedSort); ok && st.Name != "" {
		return "", false, nil
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if ok {
			if hi < lo {
				return "", false, nil
			}
			return g.goBoundedLoopHeaderForSort(s, name, strconv.Itoa(lo), "("+strconv.Itoa(hi)+")+1")
		}
		loExpr, hiExpr, ok, err := g.goSymbolicRangeLoopBounds(rs)
		if err != nil {
			return "", false, err
		}
		if !ok {
			return "", false, nil
		}
		return g.goBoundedLoopHeaderForSort(s, name, loExpr, "("+hiExpr+" + 1)")
	}
	if g.goScalarType(s) != "int" {
		return "", false, nil
	}
	if !g.sortHasNegativeValues(s) {
		if card := g.sortCard(s); card > 0 {
			return g.goBoundedLoopHeaderForSort(s, name, "0", strconv.Itoa(card))
		}
		if hi, ok := g.sortCardinalityAttr(s); ok {
			return g.goBoundedLoopHeaderForSort(s, name, "0", hi)
		}
	}
	return "", false, nil
}

func (g *Generator) goSymbolicRangeLoopBounds(rs *goivy.RangeSort) (string, string, bool, error) {
	if rs == nil || rs.Lb == nil || rs.Ub == nil {
		return "", "", false, nil
	}
	lo, ok, err := g.goRangeLoopBoundExpr(rs.Lb)
	if err != nil || !ok {
		return "", "", false, err
	}
	hi, ok, err := g.goRangeLoopBoundExpr(rs.Ub)
	if err != nil || !ok {
		return "", "", false, err
	}
	return lo, hi, true, nil
}

func (g *Generator) goRangeLoopBoundExpr(b goivy.NumeralOrCompiledBound) (string, bool, error) {
	if b == nil {
		return "", false, nil
	}
	text := b.BoundString()
	if text == "" {
		return "", false, nil
	}
	if b.IsNumeral() {
		return text, true, nil
	}
	if cb, ok := b.(goivy.CompiledBound); ok && cb.Expr != nil {
		code, err := g.emitExpr(cb.Expr)
		if err != nil {
			return "", false, err
		}
		return code, true, nil
	}
	return goName(text), true, nil
}

func (g *Generator) goLoopHeadersForSome(vars []*goivy.LogicVariable, body goivy.Expr) ([]string, bool, error) {
	headers := make([]string, len(vars))
	if len(vars) > 0 && g.goIsAnyIntegerType(vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, true); err == nil {
			ok := true
			for i, v := range vars {
				h, boundOK, herr := g.goBoundedLoopHeaderForSort(v.VSort, goName(v.Name), bounds[i][0], bounds[i][1])
				if herr != nil {
					return nil, false, herr
				}
				if !boundOK {
					ok = false
					break
				}
				headers[i] = h
			}
			if ok {
				return headers, true, nil
			}
		}
	}
	for i, v := range vars {
		name := goName(v.Name)
		if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, name); err != nil || ok {
			if err != nil {
				return nil, false, err
			}
			headers[i] = header
			continue
		}
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			return nil, false, fmt.Errorf("ivy2golang: cannot enumerate some variable %s:%s", v.Name, sortName(v.VSort))
		}
		headers[i] = fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(v.VSort), strings.Join(vals, ", "))
	}
	return headers, false, nil
}

func (g *Generator) goLoopHeadersForSomeCondition(some *goivy.SomeCondition) ([]string, error) {
	if some == nil {
		return nil, fmt.Errorf("ivy2golang: nil some condition")
	}
	headers := make([]string, len(some.Params))
	if len(some.Params) > 0 && some.Params[0] != nil && g.goIsAnyIntegerType(some.Params[0].CSort) {
		vars := make([]*goivy.LogicVariable, 0, len(some.Params))
		subs := map[goivy.NodeKey]goivy.Expr{}
		ok := true
		for _, p := range some.Params {
			if p == nil {
				ok = false
				break
			}
			v, err := goivy.NewVariable("X"+p.Name, p.CSort)
			if err != nil {
				ok = false
				break
			}
			subs[goivy.Key(p)] = v
			vars = append(vars, v)
		}
		if ok {
			if fmla, err := goivy.Substitute(some.Fmla, subs); err == nil {
				if bounds, berr := g.getAllBounds(vars, fmla, true); berr == nil {
					allBounded := true
					for i, p := range some.Params {
						h, boundOK, herr := g.goBoundedLoopHeaderForSort(p.CSort, goName(p.Name), bounds[i][0], bounds[i][1])
						if herr != nil {
							return nil, herr
						}
						if !boundOK {
							allBounded = false
							break
						}
						headers[i] = h
					}
					if allBounded {
						return headers, nil
					}
				}
			}
		}
	}
	for i, p := range some.Params {
		if p == nil {
			return nil, fmt.Errorf("ivy2golang: nil some parameter")
		}
		name := goName(p.Name)
		if header, ok, err := g.goFiniteLoopHeaderForSort(p.CSort, name); err != nil || ok {
			if err != nil {
				return nil, err
			}
			headers[i] = header
			continue
		}
		vals, ok := g.finiteValueExprs(p.CSort)
		if !ok {
			return nil, fmt.Errorf("ivy2golang: cannot enumerate some variable %s:%s", p.Name, sortName(p.CSort))
		}
		headers[i] = fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(p.CSort), strings.Join(vals, ", "))
	}
	return headers, nil
}
