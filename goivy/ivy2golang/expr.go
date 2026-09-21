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
		return "(" + l + " == " + r + ")", nil
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
	if _, ok := g.isStateSymbolName(c.Name); ok {
		return "ivy." + goName(c.Name), nil
	}
	return goName(c.Name), nil
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
		if g.isLocal(name) {
			return fn, nil
		}
		if _, ok := g.isStateSymbolName(name); ok {
			return "ivy." + fn, nil
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
	if field, ok := g.destructorFieldName(name); ok && len(args) == 1 {
		return args[0] + "." + field, nil
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
	return fmt.Sprintf("(%s.valid && %s.tag == %d && %s == %s)", lhs, lhs, idx, downcast, rhs), true, nil
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
	if len(vars) > 0 && g.goIsAnyIntegerType(vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, !forall); err == nil {
			if code, ok, err := g.emitQuantWithBounds(vars, body, forall, bounds); ok || err != nil {
				return code, err
			}
		}
	}
	if code, ok, err := g.emitExtensionalQuant(vars, body, forall); ok || err != nil {
		return code, err
	}
	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for _, v := range vars {
		name := goName(v.Name)
		g.addLocal(v.Name)
		if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, name); err != nil || ok {
			if err != nil {
				g.popScope()
				return "", err
			}
			w.open(header)
			continue
		}
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			g.popScope()
			return "", fmt.Errorf("ivy2golang: cannot enumerate quantified variable %s:%s", v.Name, sortName(v.VSort))
		}
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(v.VSort), strings.Join(vals, ", ")))
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

	remainingVals := map[string][]string{}
	for _, v := range vars {
		if v == nil || boundNames[v.Name] {
			continue
		}
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			return "", true, fmt.Errorf("ivy2golang: cannot enumerate quantified variable %s:%s", v.Name, sortName(v.VSort))
		}
		remainingVals[v.Name] = vals
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
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(v.VSort), strings.Join(remainingVals[v.Name], ", ")))
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
		default:
			return "", fmt.Errorf("ivy2golang: parametric let definition not yet supported: %s", def.Lhs.String())
		}
	}
	body, err := goivy.Substitute(l.Body, subs)
	if err != nil {
		return "", fmt.Errorf("ivy2golang: let substitution: %w", err)
	}
	return g.emitExpr(body)
}
