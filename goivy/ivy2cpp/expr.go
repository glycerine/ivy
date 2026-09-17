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
		if code, ok := g.emitBVNumeral(n); ok {
			return code, nil
		}
		if code, ok := g.emitRangeNumeral(n); ok {
			return code, nil
		}
		if code, ok, err := g.emitStringInterpConst(n); ok || err != nil {
			return code, err
		}
		if goivy.IsNumeral(n) && !goivy.IsLiteralString(n) {
			return n.Name, nil
		}
		if strings.HasPrefix(n.Name, "arg.arg") {
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
		return "!" + body, nil
	case *goivy.LogicLiteral:
		body, err := g.emitExpr(n.Atom)
		if err != nil {
			return "", err
		}
		if n.Polarity == 0 {
			return "!" + body, nil
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
		return "(!" + l + " || " + r + ")", nil
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
		return "(" + c + " ? " + t + " : " + f + ")", nil
	case *goivy.ForAll:
		return g.emitQuant(n.Variables, n.Body, true)
	case *goivy.LogicExists:
		return g.emitQuant(n.Variables, n.Body, false)
	case *goivy.LogicNativeExpr:
		return g.emitNativeExpr(n)
	case *goivy.LogicSome:
		return g.emitSome(n)
	case *goivy.LogicLet:
		return g.emitLetExpr(n)
	case *goivy.LogicNamedBinder:
		return "", fmt.Errorf("ivy2cpp: named binder %q is not supported as a C++ expression: %s", n.Name, n.String())
	case *goivy.LogicGlobally, *goivy.LogicEventually, *goivy.LogicWhenOperator:
		return "", fmt.Errorf("ivy2cpp: temporal expression %T is not supported in C++ generation yet: %s", e, e.String())
	default:
		return "", fmt.Errorf("ivy2cpp: unsupported expression %T: %s", e, e.String())
	}
}

func (g *Generator) emitExprWithHeader(w *cppWriter, e goivy.Expr) (string, error) {
	if w == nil {
		return g.emitExpr(e)
	}
	if e == nil {
		return "", fmt.Errorf("ivy2cpp: nil expression")
	}
	if repl, ok := g.aliasForSymbol(e); ok {
		return g.emitExprWithHeader(w, repl)
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
		if code, ok := g.emitBVNumeral(n); ok {
			return code, nil
		}
		if code, ok := g.emitRangeNumeral(n); ok {
			return code, nil
		}
		if code, ok, err := g.emitStringInterpConst(n); ok || err != nil {
			return code, err
		}
		if goivy.IsNumeral(n) && !goivy.IsLiteralString(n) {
			return n.Name, nil
		}
		if strings.HasPrefix(n.Name, "arg.arg") {
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
		return g.emitApplyWithHeader(w, n)
	case *goivy.Eq:
		l, err := g.emitExprWithHeader(w, n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExprWithHeader(w, n.T2)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicNot:
		body, err := g.emitExprWithHeader(w, n.Body)
		if err != nil {
			return "", err
		}
		return "!" + body, nil
	case *goivy.LogicLiteral:
		body, err := g.emitExprWithHeader(w, n.Atom)
		if err != nil {
			return "", err
		}
		if n.Polarity == 0 {
			return "!" + body, nil
		}
		return body, nil
	case *goivy.LogicAnd:
		return g.emitNaryWithHeader(w, n.Terms, "&&", "true")
	case *goivy.LogicOr:
		return g.emitNaryWithHeader(w, n.Terms, "||", "false")
	case *goivy.LogicImplies:
		l, err := g.emitExprWithHeader(w, n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExprWithHeader(w, n.T2)
		if err != nil {
			return "", err
		}
		return "(!" + l + " || " + r + ")", nil
	case *goivy.LogicIff:
		l, err := g.emitExprWithHeader(w, n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExprWithHeader(w, n.T2)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicIte:
		c, err := g.emitExprWithHeader(w, n.Cond)
		if err != nil {
			return "", err
		}
		t, err := g.emitExprWithHeader(w, n.Then)
		if err != nil {
			return "", err
		}
		f, err := g.emitExprWithHeader(w, n.Else)
		if err != nil {
			return "", err
		}
		return "(" + c + " ? " + t + " : " + f + ")", nil
	case *goivy.ForAll:
		return g.emitQuantWithHeader(w, n.Variables, n.Body, true)
	case *goivy.LogicExists:
		return g.emitQuantWithHeader(w, n.Variables, n.Body, false)
	case *goivy.LogicNativeExpr:
		return g.emitNativeExpr(n)
	case *goivy.LogicSome:
		return g.emitSome(n)
	case *goivy.LogicLet:
		return g.emitLetExprWithHeader(w, n)
	case *goivy.LogicNamedBinder:
		return "", fmt.Errorf("ivy2cpp: named binder %q is not supported as a C++ expression: %s", n.Name, n.String())
	case *goivy.LogicGlobally, *goivy.LogicEventually, *goivy.LogicWhenOperator:
		return "", fmt.Errorf("ivy2cpp: temporal expression %T is not supported in C++ generation yet: %s", e, e.String())
	default:
		return "", fmt.Errorf("ivy2cpp: unsupported expression %T: %s", e, e.String())
	}
}

func (g *Generator) emitNaryWithHeader(w *cppWriter, terms []goivy.Expr, op, ident string) (string, error) {
	if len(terms) == 0 {
		return ident, nil
	}
	parts := make([]string, len(terms))
	for i, t := range terms {
		s, err := g.emitExprWithHeader(w, t)
		if err != nil {
			return "", err
		}
		parts[i] = s
	}
	return "(" + strings.Join(parts, " "+op+" ") + ")", nil
}

func (g *Generator) emitApplyWithHeader(w *cppWriter, a *goivy.Apply) (string, error) {
	if g != nil && g.Mod != nil && g.Mod.Cfg != nil && g.Mod.Cfg.IuCfg != nil &&
		goivy.IsMacro(a, g.Mod.Cfg.IuCfg) {
		return g.emitExprWithHeader(w, goivy.ExpandMacro(a))
	}
	fnExpr := a.Func
	if repl, ok := g.aliasForSymbol(fnExpr); ok {
		fnExpr = repl
	}
	name := goivy.ExprName(fnExpr)
	if code, ok, err := g.emitCastApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitBVApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitNatMinusApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitRangeArithApply(name, a); ok || err != nil {
		return code, err
	}
	if len(a.Terms) == 2 && isInfix(name) {
		l, err := g.emitExprWithHeader(w, a.Terms[0])
		if err != nil {
			return "", err
		}
		r, err := g.emitExprWithHeader(w, a.Terms[1])
		if err != nil {
			return "", err
		}
		return "(" + l + " " + name + " " + r + ")", nil
	}
	if code, ok, err := g.emitVariantRelationWithHeader(w, name, a.Terms); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitDestructorApplyWithHeader(w, name, a.Terms); ok || err != nil {
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
		s, err := g.emitExprWithHeader(w, t)
		if err != nil {
			return "", err
		}
		args[i] = s
	}
	if g.isDefinitionName(name) {
		return fn + "(" + strings.Join(args, ",") + ")", nil
	}
	return g.cppStorageAccess(name, fnExpr.NodeSort(), args, ""), nil
}

func (g *Generator) emitLetExprWithHeader(w *cppWriter, l *goivy.LogicLet) (string, error) {
	if l == nil {
		return "", fmt.Errorf("ivy2cpp: nil let expression")
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, d := range l.Defs {
		def, ok := d.(*goivy.LogicDefinition)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: let definition has unsupported shape %T: %s", d, d.String())
		}
		lhs := def.Lhs
		switch sym := lhs.(type) {
		case *goivy.Const:
			subs[goivy.Key(sym)] = def.Rhs
		case *goivy.LogicVariable:
			subs[goivy.Key(sym)] = def.Rhs
		default:
			return "", fmt.Errorf("ivy2cpp: parametric let definition not yet supported: %s", lhs.String())
		}
	}
	body, err := goivy.Substitute(l.Body, subs)
	if err != nil {
		return "", fmt.Errorf("ivy2cpp: let substitution: %w", err)
	}
	return g.emitExprWithHeader(w, body)
}

// emitLetExpr expands `let d1, d2, ... in body` into `body` after substituting
// each definition's LHS by its RHS. Mirrors Python `ivy_logic_utils`
// behavior where `let` is removed from the AST via substitution before
// `emit_app` is ever invoked.
func (g *Generator) emitLetExpr(l *goivy.LogicLet) (string, error) {
	if l == nil {
		return "", fmt.Errorf("ivy2cpp: nil let expression")
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, d := range l.Defs {
		def, ok := d.(*goivy.LogicDefinition)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: let definition has unsupported shape %T: %s", d, d.String())
		}
		lhs := def.Lhs
		// Parameterless LHS (Const or Variable): substitute the symbol
		// by its RHS directly.
		switch sym := lhs.(type) {
		case *goivy.Const:
			subs[goivy.Key(sym)] = def.Rhs
		case *goivy.LogicVariable:
			subs[goivy.Key(sym)] = def.Rhs
		default:
			// Parametric let (let p(X,Y) := body in ...). Faithful Python
			// behavior would rewrite each occurrence p(a,b) with
			// body[X→a, Y→b]; we do not yet support this shape and refuse
			// rather than emit silently wrong C++.
			return "", fmt.Errorf("ivy2cpp: parametric let definition not yet supported: %s", lhs.String())
		}
	}
	body, err := goivy.Substitute(l.Body, subs)
	if err != nil {
		return "", fmt.Errorf("ivy2cpp: let substitution: %w", err)
	}
	return g.emitExpr(body)
}

// emitStringInterpConst mirrors Python emit_constant's strlit branch
// (ivy_to_cpp.py:3019-3023): numeral 0 on a strlit-interpreted sort
// becomes the empty C++ string literal; any other numeral on such a sort
// is an error because there is no canonical way to lower it.
func (g *Generator) emitStringInterpConst(c *goivy.Const) (string, bool, error) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false, nil
	}
	if !g.hasStringInterp(c.CSort) {
		return "", false, nil
	}
	if c.Name == "0" {
		return `""`, true, nil
	}
	return "", true, fmt.Errorf("ivy2cpp: cannot compile numeral %s of string sort %s", c.Name, sortName(c.CSort))
}

// emitCastApply handles Python emit_app's `cast` branch
// (ivy_to_cpp.py:3155-3175). Casts to BV target sorts are left to
// emitBVApply via the (false, nil) fall-through.
func (g *Generator) emitCastApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "cast" || len(a.Terms) != 1 {
		return "", false, nil
	}
	rng := a.NodeSort()
	// BV destination — delegate to emitBVApply which already handles
	// `(value & ((1<<W)-1))` masking.
	if _, ok := g.bvWidthForSort(rng); ok {
		return "", false, nil
	}
	operand, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	if rs, ok := g.rangeSortFor(rng); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", true, fmt.Errorf("ivy2cpp: cannot emit cast to non-numeric range %s", sortName(rng))
		}
		return rangeClampExpr(operand, lo, hi), true, nil
	}
	if g.hasNatInterp(rng) {
		return natSaturateExpr(operand), true, nil
	}
	if interp, ok := g.sortInterpString(rng); ok && interp == "int" {
		return operand, true, nil
	}
	return operand, true, nil
}

// emitNatMinusApply lowers `x - y` whose result sort has interp `"nat"`
// to a saturating subtraction. Mirrors Python emit_app:3148-3154.
func (g *Generator) emitNatMinusApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "-" || len(a.Terms) != 2 {
		return "", false, nil
	}
	if !g.hasNatInterp(a.NodeSort()) {
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
	return fmt.Sprintf("([&](){ auto __a = (%s); auto __b = (%s); return __a < __b ? 0 : __a - __b; })()", l, r), true, nil
}

// emitRangeArithApply lowers binary `+`, `-`, `*`, `/`, `%` whose result
// sort is a RangeSort: evaluate operands once, then clamp to [lb, ub].
// Mirrors Python emit_app:3176-3187.
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
	body := fmt.Sprintf("auto __x = (%s) %s (%s); return %s;", l, name, r, rangeClampExpr("__x", lo, hi))
	return fmt.Sprintf("([&](){ %s })()", body), true, nil
}

// rangeClampExpr returns `( x < lo ? lo : hi < x ? hi : x )`, matching
// Python's clamp shape (ivy_to_cpp.py:3030 and 3174).
func rangeClampExpr(x, lo, hi string) string {
	return fmt.Sprintf("( %s < %s ? %s : %s < %s ? %s : %s )", x, lo, lo, hi, x, hi, x)
}

// natSaturateExpr returns `( x < 0 ? 0 : x )` as a single C++ expression.
// Mirrors Python emit_app:3161-3164 cast-to-nat saturation.
func natSaturateExpr(x string) string {
	return fmt.Sprintf("([&](){ auto __x = (%s); return __x < 0 ? 0 : __x; })()", x)
}

func (g *Generator) emitRangeNumeral(c *goivy.Const) (string, bool) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false
	}
	rs, ok := g.rangeSortFor(c.CSort)
	if !ok {
		return "", false
	}
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return "", false
	}
	x := c.Name
	return fmt.Sprintf("(%s < %s ? %s : %s < %s ? %s : %s)", x, lo, lo, hi, x, hi, x), true
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
		parts[i] = s
	}
	return "(" + strings.Join(parts, " "+op+" ") + ")", nil
}

func (g *Generator) emitApply(a *goivy.Apply) (string, error) {
	// Python: `if il.is_macro(self): return il.expand_macro(self).emit(...)`
	// (ivy_to_cpp.py:3124). Expand macros before any other dispatch so the
	// expansion gets to use the regular emitter (constants, infix, etc).
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
	if code, ok, err := g.emitBVApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitNatMinusApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitRangeArithApply(name, a); ok || err != nil {
		return code, err
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
		return fn + "(" + strings.Join(args, ",") + ")", nil
	}
	return g.cppStorageAccess(name, fnExpr.NodeSort(), args, ""), nil
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
	downcast := g.variantDowncastExpr(lhs, terms[0].NodeSort(), terms[1].NodeSort(), "")
	return fmt.Sprintf("(%s.tag == %d && %s == %s)", lhs, idx, downcast, rhs), true, nil
}

func (g *Generator) emitVariantRelationWithHeader(w *cppWriter, name string, terms []goivy.Expr) (string, bool, error) {
	if name != "*>" {
		return "", false, nil
	}
	if len(terms) != 2 {
		return "", true, fmt.Errorf("ivy2cpp: variant relation *> expected 2 arguments, got %d", len(terms))
	}
	if g == nil || g.Mod == nil || !g.Mod.IsVariant(terms[0].NodeSort(), terms[1].NodeSort()) {
		return "", true, fmt.Errorf("ivy2cpp: %s *> %s is not a known variant relation", terms[0].String(), terms[1].String())
	}
	lhs, err := g.emitExprWithHeader(w, terms[0])
	if err != nil {
		return "", true, err
	}
	rhs, err := g.emitExprWithHeader(w, terms[1])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(terms[0].NodeSort(), terms[1].NodeSort())
	downcast := g.variantDowncastExpr(lhs, terms[0].NodeSort(), terms[1].NodeSort(), "")
	return fmt.Sprintf("(%s.tag == %d && %s == %s)", lhs, idx, downcast, rhs), true, nil
}

func variantPayloadField(s goivy.Sort) string {
	return "__" + varName(sortName(s))
}

func (g *Generator) emitDestructorApply(name string, terms []goivy.Expr) (string, bool, error) {
	field, ok := g.destructorFieldName(name)
	if !ok {
		return "", false, nil
	}
	if len(terms) < 1 {
		return "", true, fmt.Errorf("ivy2cpp: destructor %s expected at least 1 argument", name)
	}
	obj, err := g.emitExpr(terms[0])
	if err != nil {
		return "", true, err
	}
	dom, rng := g.destructorFieldSig(name)
	args := make([]string, 0, len(terms)-1)
	for _, t := range terms[1:] {
		s, err := g.emitExpr(t)
		if err != nil {
			return "", true, err
		}
		args = append(args, s)
	}
	return g.cppDestructorFieldAccess(field, dom, rng, args, obj), true, nil
}

func (g *Generator) emitDestructorApplyWithHeader(w *cppWriter, name string, terms []goivy.Expr) (string, bool, error) {
	field, ok := g.destructorFieldName(name)
	if !ok {
		return "", false, nil
	}
	if len(terms) < 1 {
		return "", true, fmt.Errorf("ivy2cpp: destructor %s expected at least 1 argument", name)
	}
	obj, err := g.emitExprWithHeader(w, terms[0])
	if err != nil {
		return "", true, err
	}
	dom, rng := g.destructorFieldSig(name)
	args := make([]string, 0, len(terms)-1)
	for _, t := range terms[1:] {
		s, err := g.emitExprWithHeader(w, t)
		if err != nil {
			return "", true, err
		}
		args = append(args, s)
	}
	return g.cppDestructorFieldAccess(field, dom, rng, args, obj), true, nil
}

// destructorFieldSig returns the field domain (excluding the implicit struct
// receiver) and range for a destructor symbol identified by name.
func (g *Generator) destructorFieldSig(name string) ([]goivy.Sort, goivy.Sort) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return nil, nil
	}
	owner, ok := g.Mod.DestructorSorts[name]
	if !ok {
		return nil, nil
	}
	ownerName := sortName(owner)
	destrs := g.Mod.SortDestructors.Get(ownerName)
	for _, d := range destrs {
		if d != nil && d.Name == name {
			if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
				dom := fs.Domain()
				if len(dom) > 0 {
					dom = dom[1:]
				}
				return dom, fs.Range()
			}
		}
	}
	return nil, nil
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
	if !forall {
		if code, ok, err := g.emitExistsVariantRelation(vars, body); ok || err != nil {
			return code, err
		}
	}
	// Try iterable-attribute / inequality-derived bounds before the
	// extensional-relation fallback. Python `emit_quant`
	// (ivy_to_cpp.py:3403-3435) handles iterable first, then computes
	// numeric bounds via `get_bounds` for integer-typed vars; only when
	// both fail does it look for extensional relations (line 3437-3453).
	exists := !forall
	if header, ok, err := g.quantIterableHeader(vars[0]); ok || err != nil {
		if err != nil {
			return "", err
		}
		return g.emitQuantWithHeaders(vars, body, forall, []string{header}, true)
	}
	var boundsErr error
	if cppIsAnyIntegerType(g, vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, exists); err == nil {
			headers := make([]string, len(vars))
			for i, v := range vars {
				h, herr := g.loopHeaderForSortBounds(v.VSort, varName(v.Name), bounds[i][0], bounds[i][1])
				if herr != nil {
					headers = nil
					break
				}
				headers[i] = h
			}
			if headers != nil {
				return g.emitQuantWithHeaders(vars, body, forall, headers, false)
			}
		} else {
			boundsErr = err
		}
	}
	if code, ok, err := g.emitExtensionalQuant(vars, body, forall); ok || err != nil {
		return code, err
	}
	if boundsErr != nil {
		return "", boundsErr
	}
	// Last-resort: per-variable bounded loops over finite sorts. This is
	// the same fallback path that existed before TODO 012; Python would
	// have raised BoundsError but goivy attempts a best-effort loop.
	headers := make([]string, len(vars))
	for i, v := range vars {
		h, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", err
		}
		headers[i] = h
	}
	return g.emitQuantWithHeaders(vars, body, forall, headers, false)
}

func (g *Generator) emitQuantWithHeader(w *cppWriter, vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, error) {
	if len(vars) == 0 {
		return g.emitExprWithHeader(w, body)
	}
	if !forall {
		if code, ok, err := g.emitExistsVariantRelation(vars, body); ok || err != nil {
			return code, err
		}
	}
	v0 := vars[0]
	if v0 == nil {
		return "", fmt.Errorf("ivy2cpp: nil quantified variable")
	}
	rest := vars[1:]
	exists := !forall

	res := g.nextTemp("__tmp")
	w.linef("int %s;", res)
	if exists {
		w.linef("%s = 0;", res)
	} else {
		w.linef("%s = 1;", res)
	}

	header, remaining, err := g.quantHeaderForFirstVar(v0, rest, body, exists)
	if err != nil {
		return "", err
	}
	w.line(header)
	w.indent++

	inner, err := g.emitQuantWithHeader(w, remaining, body, forall)
	if err != nil {
		return "", err
	}
	if exists {
		w.linef("if (%s) %s = 1;", inner, res)
	} else {
		w.linef("if (!%s) %s = 0;", inner, res)
	}

	w.indent--
	w.line("}")
	return res, nil
}

func (g *Generator) quantHeaderForFirstVar(v0 *goivy.LogicVariable, rest []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, []*goivy.LogicVariable, error) {
	if header, ok, err := g.quantIterableHeader(v0); ok || err != nil {
		if err != nil {
			return "", nil, err
		}
		return header, rest, nil
	}
	boundsErr := error(nil)
	if lo, hi, err := g.getBounds(v0, rest, body, exists); err == nil && cppIsAnyIntegerType(g, v0.VSort) {
		header, err := g.loopHeaderForSortBounds(v0.VSort, varName(v0.Name), lo, hi)
		if err != nil {
			return "", nil, err
		}
		return header, rest, nil
	} else if err != nil {
		boundsErr = err
	}
	if header, remaining, ok, err := g.extensionalHeaderForFirstVar(v0, rest, body, exists); ok || err != nil {
		if err != nil {
			return "", nil, err
		}
		return header, remaining, nil
	}
	if boundsErr != nil {
		return "", nil, boundsErr
	}
	return "", nil, fmt.Errorf("ivy2cpp: cannot iterate over sort %s", sortName(v0.VSort))
}

func (g *Generator) extensionalHeaderForFirstVar(v0 *goivy.LogicVariable, rest []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, []*goivy.LogicVariable, bool, error) {
	if v0 == nil || g == nil || g.Mod == nil {
		return "", nil, false, nil
	}
	var ebnds []*goivy.Apply
	g.matchExtensionalBoundExprs(v0, body, exists, &ebnds)
	if len(ebnds) == 0 {
		return "", nil, false, nil
	}
	ebnd := ebnds[0]
	fs, ok := ebnd.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok {
		return "", nil, false, nil
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	if st.Kind != cppStorageHashThunk {
		return "", nil, false, nil
	}
	rel := varName(goivy.ExprName(ebnd.Func))
	header := fmt.Sprintf("for(auto it=%s.memo.begin(),en=%s.memo.end(); it != en; ++it)if (it->second) { ", rel, rel)
	remaining := append([]*goivy.LogicVariable(nil), rest...)
	for _, term := range ebnd.Terms {
		tv, ok := term.(*goivy.LogicVariable)
		if !ok {
			continue
		}
		remaining = removeQuantVarByName(remaining, tv.Name)
	}
	return header, remaining, true, nil
}

func removeQuantVarByName(vars []*goivy.LogicVariable, name string) []*goivy.LogicVariable {
	out := vars[:0]
	for _, v := range vars {
		if v == nil || v.Name != name {
			out = append(out, v)
		}
	}
	return out
}

// emitQuantWithHeaders renders the IIFE body once the per-variable loop
// headers have been chosen by the caller. When `iterFirstOnly` is true,
// only the first header is the iterable-attribute loop; the remaining
// variables are recursively handled by emitQuant inside (matching
// Python's recursion in emit_quant after `header.append('for (...) {`).
func (g *Generator) emitQuantWithHeaders(vars []*goivy.LogicVariable, body goivy.Expr, forall bool, headers []string, iterFirstOnly bool) (string, error) {
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	emitted := headers
	if iterFirstOnly {
		emitted = headers[:1]
	}
	for _, h := range emitted {
		w.line(h)
		w.indent++
	}
	if iterFirstOnly && len(vars) > 1 {
		inner, err := g.emitQuant(vars[1:], body, forall)
		if err != nil {
			return "", err
		}
		if forall {
			w.linef("if (!(%s)) return false;", inner)
		} else {
			w.linef("if (%s) return true;", inner)
		}
	} else {
		expr, err := g.emitExpr(body)
		if err != nil {
			return "", err
		}
		if forall {
			w.linef("if (!(%s)) return false;", expr)
		} else {
			w.linef("if (%s) return true;", expr)
		}
	}
	for range emitted {
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

// quantIterableHeader emits the iterable-attribute loop header for v.
// Mirrors Python ivy_to_cpp.py:3413-3427.
func (g *Generator) quantIterableHeader(v *goivy.LogicVariable) (string, bool, error) {
	if v == nil {
		return "", false, nil
	}
	iterPrefix, iterSort, ok := g.iterableSortFor(v.VSort)
	if !ok {
		return "", false, nil
	}
	createName := varName(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "create"))
	isEndName := varName(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "is_end"))
	nextName := varName(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "next"))
	idx := varName(v.Name)
	return fmt.Sprintf("for (%s %s = %s(0); !%s(%s); %s = %s(%s)) {",
		g.cppType(iterSort), idx, createName, isEndName, idx, idx, nextName, idx), true, nil
}

func (g *Generator) emitExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (string, bool, error) {
	if len(vars) != 1 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	app, ok := body.(*goivy.Apply)
	if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 {
		return "", false, nil
	}
	bound := vars[0]
	if bound == nil || goivy.ExprName(app.Terms[1]) != bound.Name {
		return "", false, nil
	}
	if !g.Mod.IsVariant(app.Terms[0].NodeSort(), bound.VSort) {
		return "", true, fmt.Errorf("ivy2cpp: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
	}
	lhs, err := g.emitExpr(app.Terms[0])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(app.Terms[0].NodeSort(), bound.VSort)
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2cpp: no variant index for %s in %s", sortName(bound.VSort), sortName(app.Terms[0].NodeSort()))
	}
	return fmt.Sprintf("(%s.tag == %d)", lhs, idx), true, nil
}

// emitExtensionalQuant emits the `for(auto it = R.memo.begin() ...)`
// loop body used for quantifiers over relations that we know to be
// extensional (Python `the_extensional_relations`). The first
// quantified variable that has an extensional bound drives the loop;
// any other quantified variable that appears in the same extensional
// atom is also bound from it->first, matching Python's emit_quant
// (ivy_to_cpp.py:3437-3453). Remaining variables fall back to the
// numeric / per-variable loop emitted by loopHeaderForVar.
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
	if !ok {
		return "", false, nil
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	if st.Kind != cppStorageHashThunk {
		return "", false, nil
	}
	relName := goivy.ExprName(ebnd.Func)
	rel := varName(relName)

	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	w.open(fmt.Sprintf("for (auto it = %s.memo.begin(), en = %s.memo.end(); it != en; ++it) {", rel, rel))
	w.line("if (!it->second) continue;")

	// Bind every quantified variable that appears as an argument of
	// ebnd. Python: `if v == v0 or v in variables`.
	boundNames := map[string]bool{}
	for pos, term := range ebnd.Terms {
		tv, isVar := term.(*goivy.LogicVariable)
		if !isVar {
			continue
		}
		var matched *goivy.LogicVariable
		for _, v := range vars {
			if v != nil && v.Name == tv.Name {
				matched = v
				break
			}
		}
		if matched == nil || boundNames[matched.Name] {
			continue
		}
		boundNames[matched.Name] = true
		if len(ebnd.Terms) == 1 {
			w.linef("%s %s = it->first;", g.cppType(matched.VSort), varName(matched.Name))
		} else {
			w.linef("%s %s = it->first.arg%d;", g.cppType(matched.VSort), varName(matched.Name), pos)
		}
	}

	// Emit per-variable loops for any quantified variable that wasn't
	// bound by the extensional atom.
	var remaining []*goivy.LogicVariable
	for _, v := range vars {
		if v != nil && !boundNames[v.Name] {
			remaining = append(remaining, v)
		}
	}
	nestedOpened := 0
	for _, v := range remaining {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", true, err
		}
		w.open(header)
		nestedOpened++
	}

	expr, err := g.emitExpr(body)
	if err != nil {
		return "", true, err
	}
	if forall {
		w.linef("if (!(%s)) return false;", expr)
	} else {
		w.linef("if (%s) return true;", expr)
	}

	for i := 0; i < nestedOpened; i++ {
		w.close("")
	}
	w.close("") // close the extensional for-loop
	if forall {
		w.line("return true;")
	} else {
		w.line("return false;")
	}
	w.indent = 0
	w.raw("})()")
	return w.String(), true, nil
}

// matchExtensionalBoundExprs mirrors Python ivy_to_cpp.py:3351-3377
// `get_extensional_bound_exprs`. It collects extensional-relation
// applications that constrain v0, tracking quantifier polarity
// through Not / Implies / Or / And. When v0 appears inside a
// derived definition's application, the definition is unfolded
// (parameter substitution into the RHS) and the search continues.
//
// `exists` is true in an existential context. It flips through Not
// and the antecedent of Implies; under universal context (exists =
// false) only Or and the consequent of Implies recurse.
func (g *Generator) matchExtensionalBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]*goivy.Apply) {
	if v0 == nil || body == nil {
		return
	}
	// Python: if isinstance(body, il.Not):
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchExtensionalBoundExprs(v0, not.Body, !exists, res)
		return
	}
	// Go-only LogicLiteral wrapper. Polarity==0 wraps a negated atom.
	if lit, ok := body.(*goivy.LogicLiteral); ok {
		nextExists := exists
		if lit.Polarity == 0 {
			nextExists = !exists
		}
		g.matchExtensionalBoundExprs(v0, lit.Atom, nextExists, res)
		return
	}
	// Python: if il.is_app(body) and body.rep in the_extensional_relations:
	//             if v0 in body.args and exists: res.append(body)
	app, isApp := body.(*goivy.Apply)
	if isApp {
		name := goivy.ExprName(app.Func)
		if name != "" && g.extensionalRels()[name] && exists && containsVariableByName(app.Terms, v0.Name) {
			*res = append(*res, app)
		}
	}
	// Python: if isinstance(body, il.Implies) and not exists:
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchExtensionalBoundExprs(v0, imp.T1, !exists, res)
			g.matchExtensionalBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	// Python: if isinstance(body, il.Or) and not exists:
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	// Python: if isinstance(body, il.And) and exists:
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	// Python: if il.is_app(body) and body.rep in is_derived and v0 in body.args:
	if !isApp {
		return
	}
	if !containsVariableByName(app.Terms, v0.Name) {
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
	g.matchExtensionalBoundExprs(v0, substituted, exists, res)
}

// containsVariableByName returns true if any element of terms is a
// LogicVariable with the given name. Used by matchExtensionalBoundExprs
// (Python `v0 in body.args` matches variables by structural equality;
// variable scopes are unique enough that name comparison is sufficient).
func containsVariableByName(terms []goivy.Expr, name string) bool {
	for _, t := range terms {
		if v, ok := t.(*goivy.LogicVariable); ok && v.Name == name {
			return true
		}
	}
	return false
}

// boundExpr is one inequality application paired with the polarity in
// which it constrains the quantified variable. Mirrors the (expr, neg)
// tuples produced by Python `get_bound_exprs` (ivy_to_cpp.py:3264-3289).
type boundExpr struct {
	app *goivy.Apply
	neg bool
}

// matchBoundExprs mirrors Python `get_bound_exprs` (ivy_to_cpp.py:3264-3289).
// It walks the formula body collecting `<`, `<=`, `>`, `>=` applications
// that constrain v0, tracking polarity through Not / Implies / Or / And.
// `exists` is true in an existential context (the polarity flips through
// Not, and through the antecedent of Implies; under universal context
// only Or and the consequent of Implies recurse).
//
// When v0 appears inside a derived-definition application, the definition
// is unfolded by substituting its formal parameters and the search
// continues in the substituted RHS. Matches the
// matchExtensionalBoundExprs:687-767 pattern.
func (g *Generator) matchBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]boundExpr) {
	if v0 == nil || body == nil {
		return
	}
	// Python: if isinstance(body, il.Not):
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchBoundExprs(v0, not.Body, !exists, res)
		return
	}
	// Go-only LogicLiteral wrapper. Polarity==0 wraps a negated atom.
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
		// Python: if il.is_app(body) and body.rep.name in ['<','<=','>','>=']:
		name := goivy.ExprName(app.Func)
		switch name {
		case "<", "<=", ">", ">=":
			*res = append(*res, boundExpr{app: app, neg: !exists})
		}
	}
	// Python: if isinstance(body, il.Implies) and not exists:
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchBoundExprs(v0, imp.T1, !exists, res)
			g.matchBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	// Python: if isinstance(body, il.Or) and not exists:
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	// Python: if isinstance(body, il.And) and exists:
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	// Python: if il.is_app(body) and body.rep in is_derived and v0 in body.args:
	if !isApp {
		return
	}
	if !containsVariableByName(app.Terms, v0.Name) {
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

// sortHasNegativeValues mirrors Python `sort_has_negative_values`
// (ivy_to_cpp.py:3291-3292): returns true when the sort is interpreted
// as `"int"` (signed). Range/nat/bool/enumerated sorts have non-negative
// values, so we can safely lower-bound them by 0.
func (g *Generator) sortHasNegativeValues(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "int"
}

// nameOfTerm returns the variable/constant name of term, or "" if term
// is neither. Used by getBounds to match Python's `args[0] == v0`
// identity test by name.
func nameOfTerm(t goivy.Expr) string {
	switch x := t.(type) {
	case *goivy.LogicVariable:
		return x.Name
	case *goivy.Const:
		return x.Name
	}
	return ""
}

// nameIn returns true if any of vars has name n. Used to enforce
// Python's `args[1] not in variables` restriction in get_bounds.
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

// getBounds mirrors Python `get_bounds` (ivy_to_cpp.py:3301-3336).
// Returns (lo, hi) as C++ expression strings, or an error if no bound
// could be derived. `others` are sibling quantified variables that may
// not appear on the non-v0 side of an inequality (so X's bound cannot
// reference Y if Y is also quantified, matching get_all_bounds slicing).
func (g *Generator) getBounds(v0 *goivy.LogicVariable, others []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, string, error) {
	var bes []boundExpr
	g.matchBoundExprs(v0, body, exists, &bes)
	var los, his []string
	for _, be := range bes {
		op := goivy.ExprName(be.app.Func)
		strict := op == "<" || op == ">"
		// Normalize args to (lhs op rhs) with op ∈ {<, <=}.
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
		// Python: if args[0] == v0 and args[1] != v0 and args[1] not in variables:
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
		// Python: if args[1] == v0 and args[0] != v0 and args[0] not in variables:
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
	// Python: if not sort_has_negative_values(v0.sort): los.append("0")
	if !g.sortHasNegativeValues(v0.VSort) {
		los = append(los, "0")
	}
	// Python: if sort_card(v0.sort) is not None: his.append(csortcard(v0.sort))
	if card := cppSortCard(g, v0.VSort); card >= 0 {
		his = append(his, strconv.Itoa(card))
	}
	// Python: if isinstance(itp, il.RangeSort): los.append(lb), his.append(ub+1)
	if rs, ok := g.rangeSortFor(v0.VSort); ok {
		lo, hi, ok2 := numericRangeBounds(rs)
		if ok2 {
			los = append(los, lo)
			his = append(his, "("+hi+")+1")
		}
	}
	if len(los) == 0 {
		return "", "", fmt.Errorf("ivy2cpp: cannot find a lower bound for %s", quantVarDiagnostic(v0))
	}
	if len(his) == 0 {
		// Python: if il.is_uninterpreted_sort(v0.sort) and compose(name,'cardinality') in attributes:
		if hi, ok := g.sortCardinalityAttr(v0.VSort); ok {
			his = append(his, hi)
		} else {
			return "", "", fmt.Errorf("ivy2cpp: cannot find an upper bound for %s", quantVarDiagnostic(v0))
		}
	}
	return los[0], his[0], nil
}

// sortCardinalityAttr returns the upper-bound C++ expression sourced from
// a sort's `cardinality` attribute (Python ivy_to_cpp.py:3332-3334).
// Returns ("", false) if no such attribute is set.
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
	return varName(g.Mod.Cfg.IuCfg.ComposeNames(name, rep)), true
}

// getAllBounds mirrors Python `get_all_bounds` (ivy_to_cpp.py:3338-3349).
// Each variable's bound is derived using the *remaining* quantified
// variables (slice tail) as the others-set, so siblings can't appear on
// the constant side of an inequality.
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

// iterableSortFor mirrors Python's `<sort>.iterable` attribute check
// (ivy_to_cpp.py:3405-3419). When the sort carries an iterable
// attribute, returns the iter prefix (the compose(name,"iter") name) and
// the iter's sort (either compose(name,"iter") or
// compose(name,"iter","t")). Returns ok=false otherwise.
func (g *Generator) iterableSortFor(s goivy.Sort) (string, goivy.Sort, bool) {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil || g.Mod.Cfg.IuCfg == nil || g.Mod.Sig == nil {
		return "", nil, false
	}
	us, ok := s.(*goivy.UninterpretedSort)
	if !ok {
		return "", nil, false
	}
	attrKey := g.Mod.Cfg.IuCfg.ComposeNames(us.Name, "iterable")
	if _, ok := g.Mod.Attributes.Get2(attrKey); !ok {
		return "", nil, false
	}
	iterName := g.Mod.Cfg.IuCfg.ComposeNames(us.Name, "iter")
	if iterSort, ok := g.Mod.Sig.Sorts.Get2(iterName); ok {
		return iterName, iterSort, true
	}
	iterT := g.Mod.Cfg.IuCfg.ComposeNames(iterName, "t")
	if iterSort, ok := g.Mod.Sig.Sorts.Get2(iterT); ok {
		return iterName, iterSort, true
	}
	return "", nil, false
}

func (g *Generator) emitSome(s *goivy.LogicSome) (string, error) {
	if s == nil || len(s.Params) == 0 {
		return "", fmt.Errorf("ivy2cpp: empty some expression")
	}
	if code, ok, err := g.emitSomeVariantRelation(s); ok || err != nil {
		return code, err
	}
	if s.IfVal != nil || s.ElseVal != nil {
		return g.emitSomeWithElse(s)
	}
	// NOTE: goivy's LogicSome does not yet carry a min/max Kind (Python's
	// ast.SomeMinMax can appear in expression position). When the Kind
	// field is added, mirror Python emit_some:3515-3540 here.
	vars := make([]*goivy.LogicVariable, 0, len(s.Params))
	for _, p := range s.Params {
		v, ok := p.(*goivy.LogicVariable)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: some parameter %T is not a variable: %s", p, s.String())
		}
		vars = append(vars, v)
	}
	headers, err := g.someLoopHeaders(vars, s.Fmla)
	if err != nil {
		return "", err
	}
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	for _, h := range headers {
		w.line(h)
		w.indent++
	}
	cond, err := g.emitExpr(s.Fmla)
	if err != nil {
		return "", err
	}
	w.linef("if (%s) return %s;", cond, varName(vars[0].Name))
	for range headers {
		w.indent--
		w.line("}")
	}
	w.linef("return %s;", g.cppZeroValue(s.NodeSort()))
	w.indent = 0
	w.raw("})()")
	return w.String(), nil
}

// someLoopHeaders chooses the per-variable loop headers for `some`
// expressions. Mirrors Python `emit_some:3517` where bounds come from
// `get_all_bounds(header, vs, fmla, True, params)`. Falls back to
// finite-value loops when bounds extraction fails.
func (g *Generator) someLoopHeaders(vars []*goivy.LogicVariable, body goivy.Expr) ([]string, error) {
	headers := make([]string, len(vars))
	useBounds := false
	if len(vars) > 0 && cppIsAnyIntegerType(g, vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, true); err == nil {
			useBounds = true
			for i, v := range vars {
				h, herr := g.loopHeaderForSortBounds(v.VSort, varName(v.Name), bounds[i][0], bounds[i][1])
				if herr != nil {
					useBounds = false
					break
				}
				headers[i] = h
			}
		}
	}
	if useBounds {
		return headers, nil
	}
	for i, v := range vars {
		h, err := g.loopHeaderForVar(v)
		if err != nil {
			return nil, err
		}
		headers[i] = h
	}
	return headers, nil
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
		return "", true, fmt.Errorf("ivy2cpp: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
	}
	lhs, err := g.emitExpr(app.Terms[0])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(app.Terms[0].NodeSort(), bound.VSort)
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2cpp: no variant index for %s in %s", sortName(bound.VSort), sortName(app.Terms[0].NodeSort()))
	}
	boundName := varName(bound.Name)
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	w.open(fmt.Sprintf("if (%s.tag == %d) {", lhs, idx))
	w.linef("%s %s = %s;", g.cppType(bound.VSort), boundName, g.variantDowncastExpr(lhs, app.Terms[0].NodeSort(), bound.VSort, ""))
	if s.IfVal != nil || s.ElseVal != nil {
		if s.IfVal == nil || s.ElseVal == nil {
			return "", true, fmt.Errorf("ivy2cpp: some expression requires both if and else values: %s", s.String())
		}
		ifVal, err := g.emitExpr(s.IfVal)
		if err != nil {
			return "", true, err
		}
		w.linef("return %s;", ifVal)
		w.close("")
		elseVal, err := g.emitExpr(s.ElseVal)
		if err != nil {
			return "", true, err
		}
		w.linef("return %s;", elseVal)
	} else {
		w.linef("return %s;", boundName)
		w.close("")
		w.linef("return %s;", g.cppZeroValue(bound.VSort))
	}
	w.indent = 0
	w.raw("})()")
	return w.String(), true, nil
}

func (g *Generator) emitSomeWithElse(s *goivy.LogicSome) (string, error) {
	if s.IfVal == nil || s.ElseVal == nil {
		return "", fmt.Errorf("ivy2cpp: some expression requires both if and else values: %s", s.String())
	}
	vars := make([]*goivy.LogicVariable, 0, len(s.Params))
	for _, p := range s.Params {
		v, ok := p.(*goivy.LogicVariable)
		if !ok {
			return "", fmt.Errorf("ivy2cpp: some parameter %T is not a variable: %s", p, s.String())
		}
		vars = append(vars, v)
	}
	headers, err := g.someLoopHeaders(vars, s.Fmla)
	if err != nil {
		return "", err
	}
	var w cppWriter
	w.raw("([&]() {")
	w.raw("\n")
	w.indent = 1
	for _, h := range headers {
		w.line(h)
		w.indent++
	}
	cond, err := g.emitExpr(s.Fmla)
	if err != nil {
		return "", err
	}
	ifVal, err := g.emitExpr(s.IfVal)
	if err != nil {
		return "", err
	}
	w.linef("if (%s) return %s;", cond, ifVal)
	for range headers {
		w.indent--
		w.line("}")
	}
	elseVal, err := g.emitExpr(s.ElseVal)
	if err != nil {
		return "", err
	}
	w.linef("return %s;", elseVal)
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
		return fmt.Sprintf("for (%s %s : {%s}) {", g.cppType(s), name, strings.Join(vals, ", ")), nil
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if ok {
			return fmt.Sprintf("for (%s %s = %s; %s <= %s; %s++) {", g.cppType(s), name, lo, name, hi, name), nil
		}
		loExpr, hiExpr, ok := symbolicRangeLoopBounds(rs)
		if ok {
			return fmt.Sprintf("for (%s %s = %s; %s < (%s+1); %s++) {", g.cppType(s), name, loExpr, name, hiExpr, name), nil
		}
		return "", fmt.Errorf("ivy2cpp: cannot emit bounded loop over non-numeric range %s", sortName(s))
	}
	if cppIsAnyIntegerType(g, s) {
		card := cppSortCard(g, s)
		if card > 0 {
			ct := loopIntCType(g, s)
			return fmt.Sprintf("for (%s %s = 0; %s < %d; %s++) {", ct, name, name, card, name), nil
		}
	}
	return "", fmt.Errorf("ivy2cpp: cannot emit bounded loop over %s", sortName(s))
}

func symbolicRangeLoopBounds(rs *goivy.RangeSort) (string, string, bool) {
	if rs == nil || rs.Lb == nil || rs.Ub == nil {
		return "", "", false
	}
	lo, ok := rangeLoopBoundExpr(rs.Lb)
	if !ok {
		return "", "", false
	}
	hi, ok := rangeLoopBoundExpr(rs.Ub)
	if !ok {
		return "", "", false
	}
	return lo, hi, true
}

func rangeLoopBoundExpr(b goivy.NumeralOrCompiledBound) (string, bool) {
	if b == nil {
		return "", false
	}
	text := b.BoundString()
	if text == "" {
		return "", false
	}
	if b.IsNumeral() {
		return text, true
	}
	return varName(text), true
}

// loopHeaderForSortBounds mirrors Python `open_loop` when an explicit
// bounds pair is supplied (ivy_to_cpp.py:1697-1711). For enumerated
// sorts the loop variable is cast to the enum type each step
// (ivy_to_cpp.py:1707-1708); for integer-typed sorts the natural
// half-open `[lo, hi)` form is used.
func (g *Generator) loopHeaderForSortBounds(s goivy.Sort, name, lo, hi string) (string, error) {
	if lo == "" || hi == "" {
		return "", fmt.Errorf("ivy2cpp: empty bounds for %s", name)
	}
	ct := loopIntCType(g, s)
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return fmt.Sprintf("for (%s %s = (%s)%s; (int) %s < %s; %s = (%s)(((int)%s) + 1)) {",
			ct, name, ct, lo, name, hi, name, ct, name), nil
	}
	return fmt.Sprintf("for (%s %s = %s; %s < %s; %s++) {", ct, name, lo, name, hi, name), nil
}

// loopIntCType mirrors Python ivy_to_cpp.py:1705 / 3434:
//
//	ct = 'int' if ct == 'bool' else ct if ct in int_ctypes else 'int'
//
// Used to pick a loop counter type when an explicit numeric bound is
// being emitted.
func loopIntCType(g *Generator, s goivy.Sort) string {
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return g.cppType(s)
	}
	if g != nil {
		if it, ok := g.cppInterpType(s); ok && it.Kind == cppInterpBV {
			return g.cppType(s)
		}
	}
	ct := g.cppType(s)
	switch ct {
	case "bool":
		return "int"
	case "int", "long long", "unsigned", "unsigned long long", "unsigned __int128":
		return ct
	default:
		return "int"
	}
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
