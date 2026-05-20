package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitExpr(e goivy.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("ivy2cpp: nil expression")
	}
	if repl, ok := g.aliasForSymbol(e); ok {
		return g.emitExpr(repl)
	}
	switch n := e.(type) {
	case *goivy.Const:
		if n.CSort == goivy.Boolean {
			switch n.Name {
			case "true":
				return "true", nil
			case "false":
				return "false", nil
			}
		}
		if n.Name != "" && (n.Name[0] >= '0' && n.Name[0] <= '9') {
			return n.Name, nil
		}
		if g.isDefinitionName(n.Name) {
			fn, err := funName(n.Name)
			if err != nil {
				return "", err
			}
			return fn + "()", nil
		}
		return varName(n.Name), nil
	case *goivy.LogicVariable:
		return varName(n.Name), nil
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
		return "((" + l + ") == (" + r + "))", nil
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
		return "((" + c + ") ? (" + t + ") : (" + f + "))", nil
	case *goivy.ForAll:
		return g.emitQuant(n.Variables, n.Body, true)
	case *goivy.LogicExists:
		return g.emitQuant(n.Variables, n.Body, false)
	case *goivy.LogicNativeExpr:
		return g.emitNativeExpr(n)
	case *goivy.LogicSome:
		return g.emitSome(n)
	case *goivy.LogicGlobally, *goivy.LogicEventually, *goivy.LogicWhenOperator:
		return "", fmt.Errorf("ivy2cpp: temporal expression %T is not supported in C++ generation yet: %s", e, e.String())
	default:
		return "", fmt.Errorf("ivy2cpp: unsupported expression %T: %s", e, e.String())
	}
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
	return strings.Join(parts, " "+op+" "), nil
}

func (g *Generator) emitApply(a *goivy.Apply) (string, error) {
	fnExpr := a.Func
	if repl, ok := g.aliasForSymbol(fnExpr); ok {
		fnExpr = repl
	}
	name := goivy.ExprName(fnExpr)
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
	if code, ok, err := g.emitVariantRelation(name, a.Terms); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitDestructorApply(name, a.Terms); ok || err != nil {
		return code, err
	}
	fn, err := funName(name)
	if err != nil {
		return "", err
	}
	if len(a.Terms) == 0 {
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
	if g.isDefinitionName(name) {
		return fn + "(" + strings.Join(args, ", ") + ")", nil
	}
	if len(args) == 1 {
		return fn + "[" + args[0] + "]", nil
	}
	return fn + "[std::make_tuple(" + strings.Join(args, ", ") + ")]", nil
}

func (g *Generator) emitVariantRelation(name string, terms []goivy.Expr) (string, bool, error) {
	if name != "*>" {
		return "", false, nil
	}
	if len(terms) != 2 {
		return "", true, fmt.Errorf("ivy2cpp: variant relation *> expected 2 arguments, got %d", len(terms))
	}
	if g == nil || g.Mod == nil || !g.Mod.IsVariant(terms[0].NodeSort(), terms[1].NodeSort()) {
		return "", true, fmt.Errorf("ivy2cpp: %s *> %s is not a known variant relation", terms[0].String(), terms[1].String())
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
	field := variantPayloadField(terms[1].NodeSort())
	return fmt.Sprintf("(%s.__tag == %d && %s.%s == %s)", lhs, idx, lhs, field, rhs), true, nil
}

func variantPayloadField(s goivy.Sort) string {
	return "__" + varName(sortName(s))
}

func (g *Generator) emitDestructorApply(name string, terms []goivy.Expr) (string, bool, error) {
	field, ok := g.destructorFieldName(name)
	if !ok {
		return "", false, nil
	}
	if len(terms) != 1 {
		return "", true, fmt.Errorf("ivy2cpp: destructor %s expected 1 argument, got %d", name, len(terms))
	}
	obj, err := g.emitExpr(terms[0])
	if err != nil {
		return "", true, err
	}
	return obj + "." + field, true, nil
}

func (g *Generator) destructorFieldName(name string) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil || name == "" {
		return "", false
	}
	if _, ok := g.Mod.DestructorSorts[name]; !ok {
		return "", false
	}
	return varName(memName(name)), true
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
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	for _, v := range vars {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", err
		}
		w.line(header)
		w.indent++
	}
	expr, err := g.emitExpr(body)
	if err != nil {
		return "", err
	}
	if forall {
		w.linef("if (!(%s)) return false;", expr)
	} else {
		w.linef("if (%s) return true;", expr)
	}
	for range vars {
		w.indent--
		w.line("}")
	}
	if forall {
		w.line("return true;")
	} else {
		w.line("return false;")
	}
	w.indent = 0
	w.raw("})()")
	return w.String(), nil
}

func (g *Generator) emitSome(s *goivy.LogicSome) (string, error) {
	if s == nil || len(s.Params) == 0 {
		return "", fmt.Errorf("ivy2cpp: empty some expression")
	}
	if s.IfVal != nil || s.ElseVal != nil {
		return "", fmt.Errorf("ivy2cpp: some expression with if/else values is not supported yet: %s", s.String())
	}
	vars := make([]*goivy.LogicVariable, 0, len(s.Params))
	for _, p := range s.Params {
		v, ok := p.(*goivy.LogicVariable)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: some parameter %T is not a variable: %s", p, s.String())
		}
		vars = append(vars, v)
	}
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	for _, v := range vars {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", err
		}
		w.line(header)
		w.indent++
	}
	cond, err := g.emitExpr(s.Fmla)
	if err != nil {
		return "", err
	}
	w.linef("if (%s) return %s;", cond, varName(vars[0].Name))
	for range vars {
		w.indent--
		w.line("}")
	}
	w.linef("return %s;", g.cppZeroValue(s.NodeSort()))
	w.indent = 0
	w.raw("})()")
	return w.String(), nil
}

func finiteValues(s goivy.Sort) ([]string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []string{"false", "true"}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			vals[i] = varName(v)
		}
		return vals, true
	default:
		return nil, false
	}
}

func (g *Generator) loopHeaderForVar(v *goivy.LogicVariable) (string, error) {
	if v == nil {
		return "", fmt.Errorf("ivy2cpp: nil loop variable")
	}
	return g.loopHeaderForSort(v.VSort, varName(v.Name))
}

func (g *Generator) loopHeaderForSort(s goivy.Sort, name string) (string, error) {
	if vals, ok := finiteValues(s); ok {
		return fmt.Sprintf("for (%s %s : {%s}) {", cppType(s), name, strings.Join(vals, ", ")), nil
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: cannot emit bounded loop over non-numeric range %s", sortName(s))
		}
		return fmt.Sprintf("for (%s %s = %s; %s <= %s; %s++) {", cppType(s), name, lo, name, hi, name), nil
	}
	return "", fmt.Errorf("ivy2cpp: cannot emit bounded loop over %s", sortName(s))
}

func (g *Generator) rangeSortFor(s goivy.Sort) (*goivy.RangeSort, bool) {
	switch st := s.(type) {
	case *goivy.RangeSort:
		return st, true
	case *goivy.UninterpretedSort:
		if g != nil && g.Mod != nil && g.Mod.Sig != nil {
			if rs, ok := g.Mod.Sig.Interp[st.Name].(*goivy.RangeSort); ok {
				return rs, true
			}
		}
	}
	return nil, false
}

func numericRangeBounds(rs *goivy.RangeSort) (string, string, bool) {
	if rs == nil || rs.Lb == nil || rs.Ub == nil || !rs.Lb.IsNumeral() || !rs.Ub.IsNumeral() {
		return "", "", false
	}
	lo, err := strconv.Atoi(rs.LbString())
	if err != nil {
		return "", "", false
	}
	hi, err := strconv.Atoi(rs.UbString())
	if err != nil {
		return "", "", false
	}
	return strconv.Itoa(lo), strconv.Itoa(hi), true
}
