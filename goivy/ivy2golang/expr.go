package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitExpr(e goivy.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("ivy2golang: nil expression")
	}
	if repl, ok := g.aliasForSymbol(e); ok {
		return g.emitExpr(repl)
	}
	switch n := e.(type) {
	case *goivy.Const:
		return g.emitConst(n)
	case *goivy.LogicVariable:
		return goName(n.Name), nil
	case *goivy.Apply:
		return g.emitApply(n)
	case *goivy.Eq:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "(" + g.goEqualExpr(l, r, n.T1.NodeSort()) + ")", nil
	case *goivy.LogicNot:
		body, err := g.emitExpr(n.Body)
		if err != nil {
			return "", err
		}
		return "!(" + body + ")", nil
	case *goivy.LogicLiteral:
		body, err := g.emitExpr(n.Atom)
		if err != nil {
			return "", err
		}
		if n.Polarity == 0 {
			return "!(" + body + ")", nil
		}
		return body, nil
	case *goivy.LogicAnd:
		return g.emitNary(n.Terms, "&&", "true")
	case *goivy.LogicOr:
		return g.emitNary(n.Terms, "||", "false")
	case *goivy.LogicImplies:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "(!(" + l + ") || (" + r + "))", nil
	case *goivy.LogicIff:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicIte:
		c, err := g.emitExpr(n.Cond)
		if err != nil {
			return "", err
		}
		t, err := g.emitExpr(n.Then)
		if err != nil {
			return "", err
		}
		f, err := g.emitExpr(n.Else)
		if err != nil {
			return "", err
		}
		return "ivyTernary(" + c + ", " + t + ", " + f + ")", nil
	case *goivy.ForAll:
		return g.emitQuant(n.Variables, n.Body, true)
	case *goivy.LogicExists:
		return g.emitQuant(n.Variables, n.Body, false)
	case *goivy.LogicNativeExpr:
		return "", fmt.Errorf("ivy2golang: native C++ expression is not translated to Go: %s", n.String())
	case *goivy.LogicSome:
		return g.emitSome(n)
	case *goivy.LogicLet:
		return g.emitLetExpr(n)
	case *goivy.LogicNamedBinder:
		return "", fmt.Errorf("ivy2golang: named binder %q is not supported as a Go expression: %s", n.Name, n.String())
	case *goivy.LogicGlobally, *goivy.LogicEventually, *goivy.LogicWhenOperator:
		return "", fmt.Errorf("ivy2golang: temporal expression %T is not supported in Go generation yet: %s", e, e.String())
	default:
		return "", fmt.Errorf("ivy2golang: unsupported expression %T: %s", e, e.String())
	}
}

func (g *Generator) emitConst(c *goivy.Const) (string, error) {
	if c == nil {
		return "", fmt.Errorf("ivy2golang: nil constant")
	}
	if c.CSort == goivy.Boolean {
		switch c.Name {
		case "true":
			return "true", nil
		case "false":
			return "false", nil
		}
	}
	if goivy.IsLiteralString(c) {
		return c.Name, nil
	}
	if code, ok, err := g.emitStringInterpConst(c); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitBVNumeral(c); ok || err != nil {
		return code, err
	}
	if goivy.IsNumeral(c) {
		if rs, ok := g.rangeSortFor(c.CSort); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if ok {
				n, err := strconv.Atoi(c.Name)
				if err == nil {
					if n < lo {
						n = lo
					}
					if n > hi {
						n = hi
					}
					return strconv.Itoa(n), nil
				}
			}
		}
		return c.Name, nil
	}
	if enum, ok := c.CSort.(*goivy.LogicEnumeratedSort); ok && enum.Name != "" && !isNumericEnum(enum) {
		for _, name := range enum.Extension {
			if name == c.Name {
				return goName(c.Name), nil
			}
		}
	}
	if code, ok := g.exprOverride(c.Name); ok {
		return code, nil
	}
	if g.isLocal(c.Name) {
		return goName(c.Name), nil
	}
	if p, ok := g.isProgressName(c.Name); ok {
		return g.progressCounterLValue(p, "ivy"), nil
	}
	if expr, ok, err := g.expandDefinitionApplication(c.Name, nil); ok || err != nil {
		if err != nil {
			return "", err
		}
		return g.emitExpr(expr)
	}
	if g.isNativeDefinitionName(c.Name) {
		fn, err := funName(c.Name)
		if err != nil {
			return "", err
		}
		return "ivy." + fn + "()", nil
	}
	if g.isSortConstructorName(c.Name) {
		fn, err := funName(c.Name)
		if err != nil {
			return "", err
		}
		return "ivy." + fn + "()", nil
	}
	if _, ok := g.isStateSymbolName(c.Name); ok {
		return "ivy." + goName(c.Name), nil
	}
	return goName(c.Name), nil
}

func (g *Generator) emitStringInterpConst(c *goivy.Const) (string, bool, error) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false, nil
	}
	if g.hasStrBVInterp(c.CSort) {
		return strconv.Quote(c.Name), true, nil
	}
	if !g.hasStringInterp(c.CSort) {
		return "", false, nil
	}
	if c.Name == "0" {
		return `""`, true, nil
	}
	return "", true, fmt.Errorf("ivy2golang: cannot compile numeral %s of string sort %s", c.Name, sortName(c.CSort))
}

func (g *Generator) emitNary(terms []goivy.Expr, op, ident string) (string, error) {
	if len(terms) == 0 {
		return ident, nil
	}
	parts := make([]string, len(terms))
	for i, t := range terms {
		s, err := g.emitExpr(t)
		if err != nil {
			return "", err
		}
		parts[i] = "(" + s + ")"
	}
	return "(" + strings.Join(parts, " "+op+" ") + ")", nil
}

func (g *Generator) emitApply(a *goivy.Apply) (string, error) {
	if g != nil && g.Mod != nil && g.Mod.Cfg != nil && g.Mod.Cfg.IuCfg != nil &&
		goivy.IsMacro(a, g.Mod.Cfg.IuCfg) {
		return g.emitExpr(goivy.ExpandMacro(a))
	}
	fnExpr := a.Func
	if repl, ok := g.aliasForSymbol(fnExpr); ok {
		fnExpr = repl
	}
	name := goivy.ExprName(fnExpr)
	if code, ok, err := g.emitBVApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitCastApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitNatMinusApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitRangeArithApply(name, a); ok || err != nil {
		return code, err
	}
	if expr, ok, err := g.emitVariantRelation(name, a.Terms); ok || err != nil {
		return expr, err
	}
	if len(a.Terms) == 2 && isInfix(name) {
		l, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(a.Terms[1])
		if err != nil {
			return "", err
		}
		return "(" + l + " " + name + " " + r + ")", nil
	}
	if expr, ok, err := g.expandDefinitionApplication(name, a.Terms); ok || err != nil {
		if err != nil {
			return "", err
		}
		return g.emitExpr(expr)
	}
	fn, err := funName(name)
	if err != nil {
		return "", err
	}
	if len(a.Terms) == 0 {
		if g.isSortConstructorName(name) {
			return "ivy." + fn + "()", nil
		}
		if g.isLocal(name) {
			return fn, nil
		}
		if _, ok := g.isStateSymbolName(name); ok {
			return "ivy." + fn, nil
		}
		if g.isNativeDefinitionName(name) {
			return "ivy." + fn + "()", nil
		}
		return fn, nil
	}
	args := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		s, err := g.emitExpr(t)
		if err != nil {
			return "", err
		}
		args[i] = s
	}
	if g.isSortConstructorName(name) {
		return "ivy." + fn + "(" + strings.Join(args, ", ") + ")", nil
	}
	if g.isNativeDefinitionName(name) {
		return "ivy." + fn + "(" + strings.Join(args, ", ") + ")", nil
	}
	if expr, ok, err := g.destructorFieldAccess(name, args); ok || err != nil {
		return expr, err
	}
	if sort, ok := g.isStateSymbolName(name); ok {
		return g.goStorageAccess(name, sort, args, "ivy"), nil
	}
	if sort, ok := g.localSort(name); ok {
		if _, isFunc := sort.(*goivy.LogicFunctionSort); isFunc {
			return g.goStorageAccess(name, sort, args, ""), nil
		}
	}
	if p, ok := g.isProgressName(name); ok {
		if len(p.Vars) != len(args) {
			return "", fmt.Errorf("ivy2golang: progress %s expected %d arguments, got %d", name, len(p.Vars), len(args))
		}
		return "ivy." + goName(name) + "[" + g.goMapKeyValue(progressDomain(p), args) + "]", nil
	}
	return fn + "(" + strings.Join(args, ", ") + ")", nil
}

func (g *Generator) emitCastApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "cast" || len(a.Terms) != 1 {
		return "", false, nil
	}
	operand, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	rng := a.NodeSort()
	if rs, ok := g.rangeSortFor(rng); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", true, fmt.Errorf("ivy2golang: cannot emit cast to non-numeric range %s", sortName(rng))
		}
		return goRangeClampExpr(operand, lo, hi), true, nil
	}
	if g.hasNatInterp(rng) {
		return goNatSaturateExpr(operand), true, nil
	}
	return operand, true, nil
}

func (g *Generator) emitNatMinusApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "-" || len(a.Terms) != 2 || !g.hasNatInterp(a.NodeSort()) {
		return "", false, nil
	}
	l, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	r, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf("func() int {\n__a := %s\n__b := %s\nif __a < __b {\nreturn 0\n}\nreturn __a - __b\n}()", l, r), true, nil
}

func (g *Generator) emitRangeArithApply(name string, a *goivy.Apply) (string, bool, error) {
	if !isInfix(name) || len(a.Terms) != 2 {
		return "", false, nil
	}
	switch name {
	case "+", "-", "*", "/", "%":
	default:
		return "", false, nil
	}
	rs, ok := g.rangeSortFor(a.NodeSort())
	if !ok {
		return "", false, nil
	}
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return "", false, nil
	}
	l, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	r, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf("func() int {\n__x := (%s) %s (%s)\nreturn %s\n}()", l, name, r, goRangeClampExpr("__x", lo, hi)), true, nil
}

func goRangeClampExpr(x string, lo, hi int) string {
	return fmt.Sprintf("func() int {\nif %s < %d {\nreturn %d\n}\nif %d < %s {\nreturn %d\n}\nreturn %s\n}()", x, lo, lo, hi, x, hi, x)
}

func goNatSaturateExpr(x string) string {
	return fmt.Sprintf("func() int {\n__x := %s\nif __x < 0 {\nreturn 0\n}\nreturn __x\n}()", x)
}

func (g *Generator) aliasForSymbol(e goivy.Expr) (goivy.Expr, bool) {
	if g == nil || len(g.exprAliases) == 0 || e == nil {
		return nil, false
	}
	switch e.(type) {
	case *goivy.Const, *goivy.LogicVariable:
	default:
		return nil, false
	}
	name := goivy.ExprName(e)
	if name == "" {
		return nil, false
	}
	repl, ok := g.exprAliases[name]
	if !ok || repl == nil || goivy.ExprName(repl) == name {
		return nil, false
	}
	return repl, true
}

func (g *Generator) emitVariantRelation(name string, terms []goivy.Expr) (string, bool, error) {
	if name != "*>" {
		return "", false, nil
	}
	if len(terms) != 2 {
		return "", true, fmt.Errorf("ivy2golang: variant relation *> expected 2 arguments, got %d", len(terms))
	}
	if g == nil || g.Mod == nil || !g.Mod.IsVariant(terms[0].NodeSort(), terms[1].NodeSort()) {
		return "", true, fmt.Errorf("ivy2golang: %s *> %s is not a known variant relation", terms[0].String(), terms[1].String())
	}
	lhs, err := g.emitExpr(terms[0])
	if err != nil {
		return "", true, err
	}
	rhs, err := g.emitExpr(terms[1])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(terms[0].NodeSort(), terms[1].NodeSort())
	downcast := g.variantDowncastExpr(lhs, terms[1].NodeSort())
	return fmt.Sprintf("(%s.valid && %s.tag == %d && %s)", lhs, lhs, idx, g.goEqualExpr(downcast, rhs, terms[1].NodeSort())), true, nil
}

type existsVariantRelationMatch struct {
	lhs   goivy.Expr
	bound *goivy.LogicVariable
	extra []goivy.Expr
	loc   goivy.Location
}

func (g *Generator) matchExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (existsVariantRelationMatch, bool, error) {
	if len(vars) != 1 || g == nil || g.Mod == nil {
		return existsVariantRelationMatch{}, false, nil
	}
	bound := vars[0]
	if bound == nil {
		return existsVariantRelationMatch{}, false, nil
	}
	terms, ok := existsVariantRelationTerms(body)
	if !ok {
		return existsVariantRelationMatch{}, false, nil
	}
	match := existsVariantRelationMatch{bound: bound}
	found := false
	for _, term := range terms {
		app, loc, ok := variantRelationApply(term)
		if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 || goivy.ExprName(app.Terms[1]) != bound.Name {
			match.extra = append(match.extra, term)
			continue
		}
		if found {
			return existsVariantRelationMatch{}, false, nil
		}
		if !g.Mod.IsVariant(app.Terms[0].NodeSort(), bound.VSort) {
			return existsVariantRelationMatch{}, true, fmt.Errorf("ivy2golang: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
		}
		match.lhs = app.Terms[0]
		match.loc = loc
		found = true
	}
	if !found {
		return existsVariantRelationMatch{}, false, nil
	}
	return match, true, nil
}

func variantRelationApply(expr goivy.Expr) (*goivy.Apply, goivy.Location, bool) {
	switch n := expr.(type) {
	case *goivy.Apply:
		if n == nil {
			return nil, goivy.Location{}, false
		}
		return n, n.GetLineno(), true
	case *goivy.LogicLiteral:
		if n.Polarity == 0 {
			return nil, goivy.Location{}, false
		}
		app, ok := n.Atom.(*goivy.Apply)
		if !ok || app == nil {
			return nil, goivy.Location{}, false
		}
		return app, n.GetLineno(), true
	case *goivy.LogicIff:
		for _, term := range actionGeneratorIffPositiveTerms(n) {
			app, ok := term.(*goivy.Apply)
			if ok && app != nil {
				return app, n.GetLineno(), true
			}
		}
		return nil, goivy.Location{}, false
	default:
		return nil, goivy.Location{}, false
	}
}

func (g *Generator) matchNegatedExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (existsVariantRelationMatch, bool, error) {
	if len(vars) != 1 || g == nil || g.Mod == nil {
		return existsVariantRelationMatch{}, false, nil
	}
	bound := vars[0]
	if bound == nil {
		return existsVariantRelationMatch{}, false, nil
	}
	terms, ok := existsVariantRelationTerms(body)
	if !ok {
		return existsVariantRelationMatch{}, false, nil
	}
	match := existsVariantRelationMatch{bound: bound}
	found := false
	for _, term := range terms {
		app, loc, ok := negatedVariantRelationApply(term)
		if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 || goivy.ExprName(app.Terms[1]) != bound.Name {
			match.extra = append(match.extra, term)
			continue
		}
		if found {
			return existsVariantRelationMatch{}, false, nil
		}
		if !g.Mod.IsVariant(app.Terms[0].NodeSort(), bound.VSort) {
			return existsVariantRelationMatch{}, true, fmt.Errorf("ivy2golang: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
		}
		match.lhs = app.Terms[0]
		match.loc = loc
		found = true
	}
	if !found {
		return existsVariantRelationMatch{}, false, nil
	}
	return match, true, nil
}

func negatedVariantRelationApply(expr goivy.Expr) (*goivy.Apply, goivy.Location, bool) {
	switch n := expr.(type) {
	case *goivy.LogicNot:
		app, ok := n.Body.(*goivy.Apply)
		if !ok || app == nil {
			return nil, goivy.Location{}, false
		}
		return app, n.GetLineno(), true
	case *goivy.LogicLiteral:
		if n.Polarity != 0 {
			return nil, goivy.Location{}, false
		}
		app, ok := n.Atom.(*goivy.Apply)
		if !ok || app == nil {
			return nil, goivy.Location{}, false
		}
		return app, n.GetLineno(), true
	case *goivy.LogicIff:
		for _, term := range actionGeneratorIffNegativeTerms(n) {
			app, ok := term.(*goivy.Apply)
			if ok && app != nil {
				return app, n.GetLineno(), true
			}
		}
		return nil, goivy.Location{}, false
	default:
		return nil, goivy.Location{}, false
	}
}

func existsVariantRelationTerms(expr goivy.Expr) ([]goivy.Expr, bool) {
	switch n := expr.(type) {
	case *goivy.LogicLet:
		expanded, ok := actionGeneratorExpandLetExpr(n)
		if !ok {
			return nil, false
		}
		return existsVariantRelationTerms(expanded)
	case *goivy.LogicAnd:
		var out []goivy.Expr
		for _, term := range n.Terms {
			terms, ok := existsVariantRelationTerms(term)
			if !ok {
				return nil, false
			}
			out = append(out, terms...)
		}
		return out, true
	default:
		return []goivy.Expr{expr}, true
	}
}

func (g *Generator) emitExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (string, bool, error) {
	match, ok, err := g.matchExistsVariantRelation(vars, body)
	if err != nil {
		return "", true, err
	}
	if !ok {
		negatedMatch, negatedOK, negatedErr := g.matchNegatedExistsVariantRelation(vars, body)
		if negatedErr != nil {
			return "", true, negatedErr
		}
		if !negatedOK {
			return "", false, nil
		}
		lhs, err := g.emitExpr(negatedMatch.lhs)
		if err != nil {
			return "", true, err
		}
		idx := g.Mod.VariantIndex(negatedMatch.lhs.NodeSort(), negatedMatch.bound.VSort)
		if idx < 0 {
			return "", true, fmt.Errorf("ivy2golang: no variant index for %s in %s", sortName(negatedMatch.bound.VSort), sortName(negatedMatch.lhs.NodeSort()))
		}
		if len(negatedMatch.extra) != 0 {
			rhs, rhsDelta, ok := g.variantExistsUniquePayloadWitness(negatedMatch.bound, negatedMatch.extra)
			if !ok {
				return "", false, nil
			}
			rhs, ok = g.variantExistsPayloadValueWithDelta(negatedMatch.bound, rhs, rhsDelta)
			if !ok {
				return "", false, nil
			}
			rhsExpr, err := g.emitExpr(rhs)
			if err != nil {
				return "", true, err
			}
			downcast := g.variantDowncastExpr(lhs, negatedMatch.bound.VSort)
			return fmt.Sprintf("!((%s.valid && %s.tag == %d && %s))", lhs, lhs, idx, g.goEqualExpr(downcast, rhsExpr, negatedMatch.bound.VSort)), true, nil
		}
		return fmt.Sprintf("!((%s.valid && %s.tag == %d))", lhs, lhs, idx), true, nil
	}
	lhs, err := g.emitExpr(match.lhs)
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(match.lhs.NodeSort(), match.bound.VSort)
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2golang: no variant index for %s in %s", sortName(match.bound.VSort), sortName(match.lhs.NodeSort()))
	}
	if len(match.extra) == 0 {
		return fmt.Sprintf("(%s.valid && %s.tag == %d)", lhs, lhs, idx), true, nil
	}
	bodyExpr := match.extra[0]
	if len(match.extra) > 1 {
		and, err := goivy.NewAnd(match.extra...)
		if err != nil {
			return "", true, err
		}
		bodyExpr = and
	}
	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	w.open(fmt.Sprintf("if !(%s.valid && %s.tag == %d) {", lhs, lhs, idx))
	w.line("return false")
	w.close("")
	name := goName(match.bound.Name)
	w.linef("%s := %s", name, g.variantDowncastExpr(lhs, match.bound.VSort))
	w.linef("_ = %s", name)
	expr, err := g.emitExpr(bodyExpr)
	if err != nil {
		return "", true, err
	}
	w.linef("return %s", expr)
	w.indent--
	w.raw("}")
	return "(" + strings.TrimSpace(w.String()) + ")()", true, nil
}

func (g *Generator) variantExistsExactPayloadWitness(bound *goivy.LogicVariable, exprs []goivy.Expr) (goivy.Expr, bool) {
	for _, expr := range exprs {
		if value, ok := g.variantExistsExactPayloadWitnessTerm(bound, expr); ok {
			return value, true
		}
	}
	return nil, false
}

func isInfix(name string) bool {
	switch name {
	case "+", "-", "*", "/", "%", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

func (g *Generator) emitQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, error) {
	if len(vars) == 0 {
		return g.emitExpr(body)
	}
	if !forall {
		if code, ok, err := g.emitExistsVariantRelation(vars, body); ok || err != nil {
			return code, err
		}
	}
	if code, ok, err := g.emitEqualityBoundQuant(vars, body, forall); ok || err != nil {
		return code, err
	}
	exists := !forall
	if len(vars) > 0 && g.goIsAnyIntegerType(vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, exists); err == nil {
			if code, ok, err := g.emitQuantWithBounds(vars, body, forall, bounds); ok || err != nil {
				return code, err
			}
		}
	}
	if g.relationOverrideQuant {
		if code, ok, err := g.emitRelationOverrideQuant(vars, body, forall); ok || err != nil {
			return code, err
		}
	}
	if code, ok, err := g.emitExtensionalQuant(vars, body, forall); ok || err != nil {
		return code, err
	}
	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for i, v := range vars {
		name := goName(v.Name)
		g.addLocal(v.Name)
		if header, ok, err := g.goFormulaLoopHeaderForVar(v, name, vars[i+1:], body, exists); err != nil || ok {
			if err != nil {
				g.popScope()
				return "", err
			}
			w.open(header)
			continue
		}
		g.popScope()
		return "", fmt.Errorf("ivy2golang: cannot enumerate quantified variable %s:%s", v.Name, sortName(v.VSort))
	}
	expr, err := g.emitExpr(body)
	if err != nil {
		g.popScope()
		return "", err
	}
	if forall {
		w.open("if !(" + expr + ") {")
		w.line("return false")
		w.close("")
	} else {
		w.open("if " + expr + " {")
		w.line("return true")
		w.close("")
	}
	for range vars {
		w.close("")
	}
	if forall {
		w.line("return true")
	} else {
		w.line("return false")
	}
	g.popScope()
	w.close("")
	return "(" + strings.TrimSpace(w.String()) + ")()", nil
}

func (g *Generator) emitEqualityBoundQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, bool, error) {
	if len(vars) == 0 || body == nil {
		return "", false, nil
	}
	for _, v := range vars {
		if v == nil {
			return "", false, nil
		}
	}
	var subs map[goivy.NodeKey]goivy.Expr
	var rest goivy.Expr
	var ok bool
	if forall {
		subs, rest, ok = equalityBoundForallBody(vars, body)
	} else {
		subs, rest, ok = equalityBoundExistsBody(vars, body)
	}
	if !ok || len(subs) != len(vars) || rest == nil {
		return "", false, nil
	}
	substituted, err := goivy.Substitute(rest, subs)
	if err != nil {
		return "", true, err
	}
	code, err := g.emitExpr(substituted)
	return code, true, err
}

func equalityBoundForallBody(vars []*goivy.LogicVariable, body goivy.Expr) (map[goivy.NodeKey]goivy.Expr, goivy.Expr, bool) {
	imp, ok := body.(*goivy.LogicImplies)
	if !ok || imp == nil {
		return nil, nil, false
	}
	subs, extra, ok := equalityBoundsAndRemainderFromExpr(vars, imp.T1)
	if !ok {
		return nil, nil, false
	}
	if len(extra) == 0 {
		return subs, imp.T2, true
	}
	antecedent := extra[0]
	if len(extra) > 1 {
		and, err := goivy.NewAnd(extra...)
		if err != nil {
			return nil, nil, false
		}
		antecedent = and
	}
	rest, err := goivy.NewImplies(antecedent, imp.T2)
	if err != nil {
		return nil, nil, false
	}
	return subs, rest, true
}

func equalityBoundExistsBody(vars []*goivy.LogicVariable, body goivy.Expr) (map[goivy.NodeKey]goivy.Expr, goivy.Expr, bool) {
	and, ok := body.(*goivy.LogicAnd)
	if !ok || and == nil {
		subs, ok := equalityBoundsFromExpr(vars, body)
		if !ok {
			return nil, nil, false
		}
		return subs, goivy.True, true
	}
	terms := make([]goivy.Expr, 0, len(and.Terms)-1)
	boundMap := equalityBoundVarMap(vars)
	if len(boundMap) != len(vars) {
		return nil, nil, false
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, term := range and.Terms {
		if key, value, ok := equalityBoundTermForVars(boundMap, term); ok {
			if _, exists := subs[key]; exists {
				return nil, nil, false
			}
			subs[key] = value
			continue
		}
		terms = append(terms, term)
	}
	if len(subs) != len(vars) {
		return nil, nil, false
	}
	if len(terms) == 0 {
		return subs, goivy.True, true
	}
	if equalityBoundTermsContainVariantMembership(terms, vars) {
		return nil, nil, false
	}
	if len(terms) == 1 {
		return subs, terms[0], true
	}
	rest, err := goivy.NewAnd(terms...)
	if err != nil {
		return nil, nil, false
	}
	return subs, rest, true
}

func equalityBoundTermsContainVariantMembership(terms []goivy.Expr, vars []*goivy.LogicVariable) bool {
	names := map[string]bool{}
	for _, v := range vars {
		if v != nil && v.Name != "" {
			names[v.Name] = true
		}
	}
	for _, term := range terms {
		if equalityBoundExprContainsVariantMembership(term, names) {
			return true
		}
	}
	return false
}

func equalityBoundExprContainsVariantMembership(expr goivy.Expr, names map[string]bool) bool {
	if app, ok := expr.(*goivy.Apply); ok && app != nil && goivy.ExprName(app.Func) == "*>" {
		for _, term := range app.Terms {
			if exprReferencesAnyNameIncludingVariables(term, names) {
				return true
			}
		}
	}
	for _, child := range expr.Children() {
		if equalityBoundExprContainsVariantMembership(child, names) {
			return true
		}
	}
	return false
}

func equalityBoundsFromExpr(vars []*goivy.LogicVariable, expr goivy.Expr) (map[goivy.NodeKey]goivy.Expr, bool) {
	subs, extra, ok := equalityBoundsAndRemainderFromExpr(vars, expr)
	if !ok || len(extra) != 0 {
		return nil, false
	}
	return subs, true
}

func equalityBoundsAndRemainderFromExpr(vars []*goivy.LogicVariable, expr goivy.Expr) (map[goivy.NodeKey]goivy.Expr, []goivy.Expr, bool) {
	boundMap := equalityBoundVarMap(vars)
	if len(boundMap) != len(vars) {
		return nil, nil, false
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	var extra []goivy.Expr
	var terms []goivy.Expr
	if and, ok := expr.(*goivy.LogicAnd); ok && and != nil {
		terms = and.Terms
	} else {
		terms = []goivy.Expr{expr}
	}
	for _, term := range terms {
		if key, value, ok := equalityBoundTermForVars(boundMap, term); ok {
			if _, exists := subs[key]; exists {
				return nil, nil, false
			}
			subs[key] = value
			continue
		}
		extra = append(extra, term)
	}
	if len(subs) != len(vars) {
		return nil, nil, false
	}
	return subs, extra, true
}

func equalityBoundVarMap(vars []*goivy.LogicVariable) map[string]*goivy.LogicVariable {
	out := map[string]*goivy.LogicVariable{}
	for _, v := range vars {
		if v == nil || v.Name == "" || out[v.Name] != nil {
			return nil
		}
		out[v.Name] = v
	}
	return out
}

func equalityBoundTermForVars(vars map[string]*goivy.LogicVariable, expr goivy.Expr) (goivy.NodeKey, goivy.Expr, bool) {
	if lit, ok := expr.(*goivy.LogicLiteral); ok {
		if lit.Polarity == 0 {
			return goivy.NodeKey(""), nil, false
		}
		expr = lit.Atom
	}
	eq, ok := expr.(*goivy.Eq)
	if !ok || eq == nil {
		return goivy.NodeKey(""), nil, false
	}
	if key, value, ok := equalityBoundSideForVars(vars, eq.T1, eq.T2); ok {
		return key, value, true
	}
	return equalityBoundSideForVars(vars, eq.T2, eq.T1)
}

func equalityBoundSideForVars(vars map[string]*goivy.LogicVariable, lhs, rhs goivy.Expr) (goivy.NodeKey, goivy.Expr, bool) {
	v, ok := lhs.(*goivy.LogicVariable)
	if !ok || v == nil {
		return goivy.NodeKey(""), nil, false
	}
	bound := vars[v.Name]
	if bound == nil || !sortsEqual(v.VSort, bound.VSort) || !sortsEqual(rhs.NodeSort(), bound.VSort) {
		return goivy.NodeKey(""), nil, false
	}
	for name := range vars {
		if exprReferencesAnyNameIncludingVariables(rhs, map[string]bool{name: true}) {
			return goivy.NodeKey(""), nil, false
		}
	}
	return goivy.Key(bound), rhs, true
}

func equalityBoundTerm(bound *goivy.LogicVariable, expr goivy.Expr) (goivy.Expr, bool) {
	if lit, ok := expr.(*goivy.LogicLiteral); ok {
		if lit.Polarity == 0 {
			return nil, false
		}
		expr = lit.Atom
	}
	eq, ok := expr.(*goivy.Eq)
	if !ok || eq == nil || bound == nil {
		return nil, false
	}
	if equalityBoundIsVar(bound, eq.T1) && equalityBoundValueOK(bound, eq.T2) {
		return eq.T2, true
	}
	if equalityBoundIsVar(bound, eq.T2) && equalityBoundValueOK(bound, eq.T1) {
		return eq.T1, true
	}
	return nil, false
}

func equalityBoundIsVar(bound *goivy.LogicVariable, expr goivy.Expr) bool {
	v, ok := expr.(*goivy.LogicVariable)
	return ok && v != nil && bound != nil && v.Name == bound.Name && sortsEqual(v.VSort, bound.VSort)
}

func equalityBoundValueOK(bound *goivy.LogicVariable, expr goivy.Expr) bool {
	return expr != nil &&
		bound != nil &&
		sortsEqual(expr.NodeSort(), bound.VSort) &&
		!exprReferencesAnyNameIncludingVariables(expr, map[string]bool{bound.Name: true})
}

func (g *Generator) emitQuantWithBounds(vars []*goivy.LogicVariable, body goivy.Expr, forall bool, bounds [][2]string) (string, bool, error) {
	if len(vars) == 0 || len(bounds) != len(vars) {
		return "", false, nil
	}
	headers := make([]string, len(vars))
	for i, v := range vars {
		if v == nil {
			return "", false, nil
		}
		header, ok, err := g.goBoundedLoopHeaderForSort(v.VSort, goName(v.Name), bounds[i][0], bounds[i][1])
		if err != nil || !ok {
			return "", ok, err
		}
		headers[i] = header
	}
	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for i, v := range vars {
		g.addLocal(v.Name)
		w.open(headers[i])
	}
	expr, err := g.emitExpr(body)
	if err != nil {
		g.popScope()
		return "", true, err
	}
	if forall {
		w.open("if !(" + expr + ") {")
		w.line("return false")
		w.close("")
	} else {
		w.open("if " + expr + " {")
		w.line("return true")
		w.close("")
	}
	for range vars {
		w.close("")
	}
	if forall {
		w.line("return true")
	} else {
		w.line("return false")
	}
	g.popScope()
	w.close("")
	return "(" + strings.TrimSpace(w.String()) + ")()", true, nil
}

func (g *Generator) emitExtensionalQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, bool, error) {
	if len(vars) == 0 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	v0 := vars[0]
	if v0 == nil {
		return "", false, nil
	}
	exists := !forall
	var ebnds []*goivy.Apply
	g.matchExtensionalBoundExprs(v0, body, exists, &ebnds)
	if len(ebnds) == 0 {
		return "", false, nil
	}
	ebnd := ebnds[0]
	fs, ok := ebnd.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || !isBooleanSort(fs.Range()) {
		return "", false, nil
	}
	relName := goivy.ExprName(ebnd.Func)
	if relName == "" {
		return "", false, nil
	}

	boundNames := map[string]bool{}
	type boundVar struct {
		v   *goivy.LogicVariable
		pos int
	}
	var bound []boundVar
	for pos, term := range ebnd.Terms {
		tv, isVar := term.(*goivy.LogicVariable)
		if !isVar {
			continue
		}
		for _, v := range vars {
			if v == nil || v.Name != tv.Name || boundNames[v.Name] {
				continue
			}
			boundNames[v.Name] = true
			bound = append(bound, boundVar{v: v, pos: pos})
			break
		}
	}
	if !boundNames[v0.Name] {
		return "", false, nil
	}

	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for _, v := range vars {
		if v != nil {
			g.addLocal(v.Name)
		}
	}
	key := g.nextTemp("__ivy_quant_key")
	val := g.nextTemp("__ivy_quant_val")
	w.open(fmt.Sprintf("for %s, %s := range %s {", key, val, g.goStorageRangeExpr(relName, fs, "ivy")))
	w.open(fmt.Sprintf("if !%s {", val))
	w.line("continue")
	w.close("")
	for _, bv := range bound {
		name := goName(bv.v.Name)
		if len(ebnd.Terms) == 1 {
			w.linef("%s := %s", name, key)
		} else {
			w.linef("%s := %s.A%d", name, key, bv.pos)
		}
		w.linef("_ = %s", name)
	}
	nestedOpened := 0
	for _, v := range vars {
		if v == nil || boundNames[v.Name] {
			continue
		}
		name := goName(v.Name)
		if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, name); err != nil || ok {
			if err != nil {
				g.popScope()
				return "", true, err
			}
			w.open(header)
			nestedOpened++
			continue
		}
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			g.popScope()
			return "", true, fmt.Errorf("ivy2golang: cannot enumerate quantified variable %s:%s", v.Name, sortName(v.VSort))
		}
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(v.VSort), strings.Join(vals, ", ")))
		nestedOpened++
	}
	expr, err := g.emitExpr(body)
	if err != nil {
		g.popScope()
		return "", true, err
	}
	if forall {
		w.open("if !(" + expr + ") {")
		w.line("return false")
		w.close("")
	} else {
		w.open("if " + expr + " {")
		w.line("return true")
		w.close("")
	}
	for i := 0; i < nestedOpened; i++ {
		w.close("")
	}
	w.close("")
	if forall {
		w.line("return true")
	} else {
		w.line("return false")
	}
	g.popScope()
	w.close("")
	return "(" + strings.TrimSpace(w.String()) + ")()", true, nil
}

func (g *Generator) emitRelationOverrideQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, bool, error) {
	if len(vars) == 0 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	var app *goivy.Apply
	var ok bool
	if forall {
		app, ok = relationOverrideForallBoundApp(vars, body)
	} else {
		app, ok = relationOverrideExistsBoundApp(vars, body)
	}
	if !ok || app == nil {
		return "", false, nil
	}
	relName := goivy.ExprName(app.Func)
	if relName == "" {
		return "", false, nil
	}
	sort, ok := g.isStateSymbolName(relName)
	if !ok {
		return "", false, nil
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) != len(app.Terms) || !isBooleanSort(fs.Range()) {
		return "", false, nil
	}
	boundByName := map[string]int{}
	for pos, term := range app.Terms {
		v, ok := term.(*goivy.LogicVariable)
		if !ok || v == nil {
			continue
		}
		for _, qv := range vars {
			if qv != nil && qv.Name == v.Name {
				boundByName[v.Name] = pos
				break
			}
		}
	}
	for _, v := range vars {
		if v == nil {
			return "", false, nil
		}
		pos, ok := boundByName[v.Name]
		if !ok || !sortsEqual(fs.Domain()[pos], v.VSort) {
			return "", false, nil
		}
	}

	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for _, v := range vars {
		g.addLocal(v.Name)
	}
	key := g.nextTemp("__ivy_quant_key")
	val := g.nextTemp("__ivy_quant_val")
	w.open(fmt.Sprintf("for %s, %s := range %s {", key, val, g.goStorageRangeExpr(relName, sort, "ivy")))
	w.open(fmt.Sprintf("if !%s {", val))
	w.line("continue")
	w.close("")
	for _, v := range vars {
		name := goName(v.Name)
		pos := boundByName[v.Name]
		if len(app.Terms) == 1 {
			w.linef("%s := %s", name, key)
		} else {
			w.linef("%s := %s.A%d", name, key, pos)
		}
		w.linef("_ = %s", name)
	}
	expr, err := g.emitExpr(body)
	if err != nil {
		g.popScope()
		return "", true, err
	}
	if forall {
		w.open("if !(" + expr + ") {")
		w.line("return false")
		w.close("")
	} else {
		w.open("if " + expr + " {")
		w.line("return true")
		w.close("")
	}
	w.close("")
	if forall {
		w.line("return true")
	} else {
		w.line("return false")
	}
	g.popScope()
	w.close("")
	return "(" + strings.TrimSpace(w.String()) + ")()", true, nil
}

func relationOverrideExistsBoundApp(vars []*goivy.LogicVariable, body goivy.Expr) (*goivy.Apply, bool) {
	varNames := map[string]bool{}
	for _, v := range vars {
		if v == nil || v.Name == "" {
			return nil, false
		}
		varNames[v.Name] = true
	}
	var find func(goivy.Expr) (*goivy.Apply, bool)
	find = func(expr goivy.Expr) (*goivy.Apply, bool) {
		switch n := expr.(type) {
		case *goivy.Apply:
			covered := map[string]bool{}
			for _, term := range n.Terms {
				if v, ok := term.(*goivy.LogicVariable); ok && v != nil && varNames[v.Name] {
					covered[v.Name] = true
				}
			}
			for name := range varNames {
				if !covered[name] {
					return nil, false
				}
			}
			return n, true
		case *goivy.LogicLiteral:
			if n.Polarity != 0 {
				return find(n.Atom)
			}
		case *goivy.LogicAnd:
			for _, term := range n.Terms {
				if app, ok := find(term); ok {
					return app, true
				}
			}
		case *goivy.LogicOr:
			for _, term := range n.Terms {
				if app, ok := find(term); ok {
					return app, true
				}
			}
		case *goivy.LogicImplies:
			return find(n.T2)
		case *goivy.LogicIff:
			for _, term := range actionGeneratorIffPositiveTerms(n) {
				if app, ok := find(term); ok {
					return app, true
				}
			}
		case *goivy.LogicIte:
			if app, ok := find(n.Then); ok {
				return app, true
			}
			return find(n.Else)
		case *goivy.LogicLet:
			expanded, ok := actionGeneratorExpandLetExpr(n)
			if !ok {
				return nil, false
			}
			return find(expanded)
		}
		return nil, false
	}
	return find(body)
}

func relationOverrideForallBoundApp(vars []*goivy.LogicVariable, body goivy.Expr) (*goivy.Apply, bool) {
	varNames := map[string]bool{}
	for _, v := range vars {
		if v == nil || v.Name == "" {
			return nil, false
		}
		varNames[v.Name] = true
	}
	var find func(goivy.Expr, bool) (*goivy.Apply, bool)
	find = func(expr goivy.Expr, positive bool) (*goivy.Apply, bool) {
		switch n := expr.(type) {
		case *goivy.Apply:
			if !positive && relationOverrideAppCoversVars(n, varNames) {
				return n, true
			}
		case *goivy.LogicNot:
			return find(n.Body, !positive)
		case *goivy.LogicLiteral:
			nextPositive := positive
			if n.Polarity == 0 {
				nextPositive = !nextPositive
			}
			return find(n.Atom, nextPositive)
		case *goivy.LogicAnd:
			for _, term := range n.Terms {
				if app, ok := find(term, positive); ok {
					return app, true
				}
			}
		case *goivy.LogicOr:
			for _, term := range n.Terms {
				if app, ok := find(term, positive); ok {
					return app, true
				}
			}
		case *goivy.LogicImplies:
			if app, ok := find(n.T1, !positive); ok {
				return app, true
			}
			return find(n.T2, positive)
		case *goivy.LogicIff:
			for _, term := range actionGeneratorIffNegativeTerms(n) {
				if app, ok := find(term, positive); ok {
					return app, true
				}
			}
		case *goivy.LogicIte:
			if app, ok := find(n.Then, positive); ok {
				return app, true
			}
			return find(n.Else, positive)
		case *goivy.LogicLet:
			expanded, ok := actionGeneratorExpandLetExpr(n)
			if !ok {
				return nil, false
			}
			return find(expanded, positive)
		}
		return nil, false
	}
	return find(body, true)
}

func relationOverrideAppCoversVars(app *goivy.Apply, varNames map[string]bool) bool {
	if app == nil || len(varNames) == 0 {
		return false
	}
	covered := map[string]bool{}
	for _, term := range app.Terms {
		if v, ok := term.(*goivy.LogicVariable); ok && v != nil && varNames[v.Name] {
			covered[v.Name] = true
		}
	}
	for name := range varNames {
		if !covered[name] {
			return false
		}
	}
	return true
}

func (g *Generator) emitSome(s *goivy.LogicSome) (string, error) {
	if s == nil || len(s.Params) == 0 {
		return "", fmt.Errorf("ivy2golang: empty some expression")
	}
	if code, ok, err := g.emitSomeVariantRelation(s); ok || err != nil {
		return code, err
	}
	return g.emitSomeFinite(s)
}

func (g *Generator) emitSomeFinite(s *goivy.LogicSome) (string, error) {
	vars := make([]*goivy.LogicVariable, 0, len(s.Params))
	for _, p := range s.Params {
		v, ok := p.(*goivy.LogicVariable)
		if !ok {
			return "", fmt.Errorf("ivy2golang: some parameter %T is not a variable: %s", p, s.String())
		}
		vars = append(vars, v)
	}
	if len(vars) == 0 {
		return "", fmt.Errorf("ivy2golang: empty some expression")
	}
	if (s.IfVal == nil) != (s.ElseVal == nil) {
		return "", fmt.Errorf("ivy2golang: some expression requires both if and else values: %s", s.String())
	}
	resultSort := s.NodeSort()
	if s.IfVal != nil {
		resultSort = s.IfVal.NodeSort()
	}
	headers, _, err := g.goLoopHeadersForSome(vars, s.Fmla)
	if err != nil {
		return "", err
	}
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", g.goScalarType(resultSort)))
	w.indent++
	g.pushScope()
	for i, v := range vars {
		g.addLocal(v.Name)
		w.open(headers[i])
	}
	for _, v := range vars {
		w.linef("_ = %s", goName(v.Name))
	}
	cond, err := g.emitExpr(s.Fmla)
	if err != nil {
		g.popScope()
		return "", err
	}
	w.open("if " + cond + " {")
	if s.IfVal != nil {
		ifVal, err := g.emitExpr(s.IfVal)
		if err != nil {
			g.popScope()
			return "", err
		}
		w.linef("return %s", ifVal)
	} else {
		w.linef("return %s", goName(vars[0].Name))
	}
	w.close("")
	for range vars {
		w.close("")
	}
	g.popScope()
	if s.ElseVal != nil {
		elseVal, err := g.emitExpr(s.ElseVal)
		if err != nil {
			return "", err
		}
		w.linef("return %s", elseVal)
	} else {
		w.linef("return %s", g.goZeroValue(s.NodeSort()))
	}
	w.indent--
	w.raw("}()")
	return w.String(), nil
}

func (g *Generator) emitSomeVariantRelation(s *goivy.LogicSome) (string, bool, error) {
	if s == nil || len(s.Params) != 1 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	bound, ok := s.Params[0].(*goivy.LogicVariable)
	if !ok || bound == nil {
		return "", false, nil
	}
	app, ok := s.Fmla.(*goivy.Apply)
	if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 {
		return "", false, nil
	}
	if goivy.ExprName(app.Terms[1]) != bound.Name {
		return "", false, nil
	}
	if !g.Mod.IsVariant(app.Terms[0].NodeSort(), bound.VSort) {
		return "", true, fmt.Errorf("ivy2golang: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
	}
	lhs, err := g.emitExpr(app.Terms[0])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(app.Terms[0].NodeSort(), bound.VSort)
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2golang: no variant index for %s in %s", sortName(bound.VSort), sortName(app.Terms[0].NodeSort()))
	}
	resultSort := bound.VSort
	if s.IfVal != nil {
		resultSort = s.IfVal.NodeSort()
	}
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", g.goScalarType(resultSort)))
	w.indent++
	w.open(fmt.Sprintf("if %s.valid && %s.tag == %d {", lhs, lhs, idx))
	g.pushScope()
	g.addLocal(bound.Name)
	w.linef("%s := %s", goName(bound.Name), g.variantDowncastExpr(lhs, bound.VSort))
	w.linef("_ = %s", goName(bound.Name))
	if s.IfVal != nil || s.ElseVal != nil {
		if s.IfVal == nil || s.ElseVal == nil {
			g.popScope()
			return "", true, fmt.Errorf("ivy2golang: some expression requires both if and else values: %s", s.String())
		}
		ifVal, err := g.emitExpr(s.IfVal)
		if err != nil {
			g.popScope()
			return "", true, err
		}
		w.linef("return %s", ifVal)
		g.popScope()
		w.close("")
		elseVal, err := g.emitExpr(s.ElseVal)
		if err != nil {
			return "", true, err
		}
		w.linef("return %s", elseVal)
	} else {
		w.linef("return %s", goName(bound.Name))
		g.popScope()
		w.close("")
		w.linef("return %s", g.goZeroValue(bound.VSort))
	}
	w.indent--
	w.raw("}()")
	return w.String(), true, nil
}

func (g *Generator) emitLetExpr(l *goivy.LogicLet) (string, error) {
	if l == nil {
		return "", fmt.Errorf("ivy2golang: nil let expression")
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	applySubs := map[goivy.NodeKey]goivy.SubstituteApplyFunc{}
	var substErr error
	for _, d := range l.Defs {
		def, ok := d.(*goivy.LogicDefinition)
		if !ok {
			return "", fmt.Errorf("ivy2golang: let definition has unsupported shape %T: %s", d, d.String())
		}
		switch sym := def.Lhs.(type) {
		case *goivy.Const:
			subs[goivy.Key(sym)] = def.Rhs
		case *goivy.LogicVariable:
			subs[goivy.Key(sym)] = def.Rhs
		case *goivy.Apply:
			formals := append([]goivy.Expr(nil), sym.Terms...)
			rhs := def.Rhs
			applySubs[goivy.Key(sym.Func)] = func(terms []goivy.Expr) goivy.Expr {
				if substErr != nil {
					return rhs
				}
				if len(terms) != len(formals) {
					substErr = fmt.Errorf("ivy2golang: parametric let definition %s expected %d arguments, got %d", sym.String(), len(formals), len(terms))
					return rhs
				}
				localSubs := make(map[goivy.NodeKey]goivy.Expr, len(subs)+len(formals))
				for k, v := range subs {
					localSubs[k] = v
				}
				for i, formal := range formals {
					switch formal.(type) {
					case *goivy.Const, *goivy.LogicVariable:
						localSubs[goivy.Key(formal)] = terms[i]
					default:
						substErr = fmt.Errorf("ivy2golang: parametric let formal has unsupported shape %T: %s", formal, formal.String())
						return rhs
					}
				}
				out, err := goivy.Substitute(rhs, localSubs)
				if err != nil {
					substErr = fmt.Errorf("ivy2golang: parametric let substitution: %w", err)
					return rhs
				}
				return out
			}
		default:
			return "", fmt.Errorf("ivy2golang: parametric let definition not yet supported: %s", def.Lhs.String())
		}
	}
	body, err := goivy.Substitute(l.Body, subs)
	if err != nil {
		return "", fmt.Errorf("ivy2golang: let substitution: %w", err)
	}
	body = goivy.SubstituteApply(body, applySubs)
	if substErr != nil {
		return "", substErr
	}
	return g.emitExpr(body)
}
