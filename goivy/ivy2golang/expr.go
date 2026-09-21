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
	case *goivy.LogicLet:
		return g.emitLetExpr(n)
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
	if g.isLocal(c.Name) {
		return goName(c.Name), nil
	}
	if p, ok := g.isProgressName(c.Name); ok {
		return g.progressCounterLValue(p, "ivy"), nil
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
	if name == "cast" && len(a.Terms) == 1 {
		return g.emitExpr(a.Terms[0])
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
	if sort, ok := g.isStateSymbolName(name); ok {
		return g.goStorageAccess(name, sort, args, "ivy"), nil
	}
	if p, ok := g.isProgressName(name); ok {
		if len(p.Vars) != len(args) {
			return "", fmt.Errorf("ivy2golang: progress %s expected %d arguments, got %d", name, len(p.Vars), len(args))
		}
		return "ivy." + goName(name) + "[" + g.goMapKeyValue(progressDomain(p), args) + "]", nil
	}
	return fn + "(" + strings.Join(args, ", ") + ")", nil
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
	var w goWriter
	w.raw("func() bool {\n")
	w.indent++
	g.pushScope()
	for _, v := range vars {
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			g.popScope()
			return "", fmt.Errorf("ivy2golang: cannot enumerate quantified variable %s:%s", v.Name, sortName(v.VSort))
		}
		name := goName(v.Name)
		g.addLocal(v.Name)
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
