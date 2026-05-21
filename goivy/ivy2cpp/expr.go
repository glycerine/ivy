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
		if goivy.IsNumeral(n) && !goivy.IsLiteralString(n) {
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
	if code, ok, err := g.emitBVApply(name, a); ok || err != nil {
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
		return fn + "(" + strings.Join(args, ", ") + ")", nil
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
	if code, ok, err := g.emitExtensionalQuant(vars, body, forall); ok || err != nil {
		return code, err
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
	ifVal, err := g.emitExpr(s.IfVal)
	if err != nil {
		return "", err
	}
	w.linef("if (%s) return %s;", cond, ifVal)
	for range vars {
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
		if !ok {
			return "", fmt.Errorf("ivy2cpp: cannot emit bounded loop over non-numeric range %s", sortName(s))
		}
		return fmt.Sprintf("for (%s %s = %s; %s <= %s; %s++) {", g.cppType(s), name, lo, name, hi, name), nil
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
