package ivy2cpp

import (
	"fmt"
	"strings"

	goivy "github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitExpr(e goivy.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("ivy2cpp: nil expression")
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
	name := goivy.ExprName(a.Func)
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
	if len(args) == 1 {
		return fn + "[" + args[0] + "]", nil
	}
	return fn + "[std::make_tuple(" + strings.Join(args, ", ") + ")]", nil
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
		vals, ok := finiteValues(v.VSort)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: cannot emit bounded quantifier over %s", sortName(v.VSort))
		}
		w.linef("for (%s %s : {%s}) {", cppType(v.VSort), varName(v.Name), strings.Join(vals, ", "))
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
