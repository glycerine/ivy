package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitAction(w *goWriter, act goivy.Action) {
	if act == nil {
		return
	}
	switch a := act.(type) {
	case *goivy.LogicSequence:
		for _, child := range a.Elems {
			if childAct, ok := child.(goivy.Action); ok {
				g.emitAction(w, childAct)
			} else {
				g.unsupported(w, "unsupported sequence child %T: %s", child, fmt.Sprint(child))
			}
		}
	case *goivy.LogicAssignAction:
		g.emitAssign(w, a)
	case *goivy.LogicAssignFieldAction:
		g.emitAssignField(w, a)
	case *goivy.LogicNullFieldAction:
		g.emitNullField(w, a)
	case *goivy.LogicCopyFieldAction:
		g.emitCopyField(w, a)
	case *goivy.LogicSetAction:
		g.emitSet(w, a)
	case *goivy.LogicAssertAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicRequiresAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicEnsuresAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicSubgoalAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicAssumeAction:
		g.emitAssume(w, a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicIfAction:
		g.emitIf(w, a)
	case *goivy.LogicWhileAction:
		g.emitWhile(w, a)
	case *goivy.LogicChoiceAction:
		g.emitChoice(w, a)
	case *goivy.LogicEnvAction:
		g.emitChoice(w, &a.LogicChoiceAction)
	case *goivy.LogicCallAction:
		g.emitCall(w, a)
	case *goivy.LogicLocalAction:
		g.emitLocal(w, a)
	case *goivy.LogicLetAction:
		g.emitLet(w, a)
	case *goivy.LogicBindOldsAction:
		g.emitBindOlds(w, a)
	case *goivy.LogicDebugAction:
		g.emitDebug(w, a)
	case *goivy.LogicCrashAction:
		_ = a
	case *goivy.ReturnAction:
		if len(g.currentReturns) > 0 {
			names := make([]string, len(g.currentReturns))
			for i, r := range g.currentReturns {
				names[i] = goName(r.Name)
			}
			w.linef("return %s", strings.Join(names, ", "))
		} else {
			w.line("return")
		}
	case *goivy.IgnoreAction:
		return
	case *goivy.LogicNativeAction:
		g.emitNativeAction(w, a)
	case *goivy.LogicHavocAction:
		g.emitHavoc(w, a)
	case *goivy.LogicThunkAction:
		g.unsupportedAt(w, a.GetLineno(), "thunk reached emit (Python ThunkAction has no emit; expected to be desugared upstream): %s",
			a.String())
	case *goivy.LogicInstantiateAction:
		g.unsupportedAt(w, a.GetLineno(), "instantiate reached emit (Python InstantiateAction has no emit; expected to be inlined upstream): %s",
			a.String())
	case *goivy.LogicRanking:
		g.unsupportedAt(w, a.GetLineno(), "ranking reached emit (Python Ranking has no emit; ranking/progress is enforced separately): %s",
			a.String())
	default:
		g.unsupportedAt(w, act.GetLineno(), "unsupported action %T: %s", act, act.String())
	}
}

func (g *Generator) emitHavoc(w *goWriter, a *goivy.LogicHavocAction) {
	target := "<nil>"
	if a != nil && a.Target != nil {
		target = a.Target.String()
	}
	loc := goivy.Location{}
	if a != nil {
		loc = a.GetLineno()
	}
	g.unsupportedAt(w, loc, "havoc reached emit (Python emit_havoc asserts False): %s", target)
}

func (g *Generator) emitAssign(w *goWriter, a *goivy.LogicAssignAction) {
	if g.emitExtensionalRelationClear(w, a) {
		return
	}
	vs := g.assignmentLoopVars(a.LHS)
	if len(vs) == 0 {
		g.emitAssignOneAt(w, a.LHS, a.RHS, a.GetLineno())
		return
	}
	g.emitAssignTwoPhase(w, a, vs)
}

func (g *Generator) emitAssignField(w *goWriter, a *goivy.LogicAssignFieldAction) {
	if a == nil || a.Field == nil || a.Obj == nil {
		g.unsupported(w, "unsupported field assignment: nil components")
		return
	}
	lhs := goivy.NewApplyUnchecked(a.Field, a.Obj)
	g.emitAssign(w, goivy.NewAssignAction(lhs, a.Value))
}

func (g *Generator) emitNullField(w *goWriter, a *goivy.LogicNullFieldAction) {
	lhs, err := g.emitFieldRef(a.Obj, a.Field)
	if err != nil {
		g.unsupported(w, "unsupported null field lhs: %s", err.Error())
		return
	}
	fieldSort, err := fieldRangeSort(a.Field)
	if err != nil {
		g.unsupported(w, "unsupported null field sort: %s", err.Error())
		return
	}
	w.linef("%s = %s", lhs, g.goZeroValue(fieldSort))
}

func (g *Generator) emitCopyField(w *goWriter, a *goivy.LogicCopyFieldAction) {
	lhs, err := g.emitFieldRef(a.Dst, a.Field)
	if err != nil {
		g.unsupported(w, "unsupported copy field lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitFieldRef(a.Src, a.SrcField)
	if err != nil {
		g.unsupported(w, "unsupported copy field rhs: %s", err.Error())
		return
	}
	w.linef("%s = %s", lhs, rhs)
}

func (g *Generator) emitFieldRef(obj, field goivy.Expr) (string, error) {
	if obj == nil || field == nil {
		return "", fmt.Errorf("nil field reference")
	}
	objCode, err := g.emitExpr(obj)
	if err != nil {
		return "", err
	}
	fieldName := goivy.ExprName(field)
	if fieldName == "" {
		return "", fmt.Errorf("field has no name: %T", field)
	}
	return objCode + "." + goName(memName(fieldName)), nil
}

func fieldRangeSort(field goivy.Expr) (goivy.Sort, error) {
	if field == nil {
		return nil, fmt.Errorf("nil field")
	}
	if fs, ok := field.NodeSort().(*goivy.LogicFunctionSort); ok {
		return fs.Range(), nil
	}
	return nil, fmt.Errorf("field %s does not have function sort", field.String())
}

func (g *Generator) emitAssignOne(w *goWriter, lhs, rhsExpr goivy.Expr) {
	g.emitAssignOneAt(w, lhs, rhsExpr, goivy.Location{})
}

func (g *Generator) emitAssignOneAt(w *goWriter, lhs, rhsExpr goivy.Expr, loc goivy.Location) {
	lhsCode, err := g.emitExpr(lhs)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported assignment lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitExpr(rhsExpr)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported assignment rhs: %s", err.Error())
		return
	}
	rhs = g.maybeVariantUpcastExpr(lhs.NodeSort(), rhsExpr.NodeSort(), rhs)
	if g.Config.Trace && !lhsHasNamespacedName(lhs) {
		trace, err := g.goTraceLHSExpr(lhs)
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported assignment trace lhs: %s", err.Error())
			return
		}
		w.linef("fmt.Fprintf(__ivy_out, %q, %s, ivyTraceValue(%s, %t))", "  write(%s,%s)\n", trace, rhs, g.traceHex())
	}
	if call, ok, err := g.goStorageSet(lhs, rhs); ok || err != nil {
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported assignment lhs: %s", err.Error())
			return
		}
		w.line(call)
		return
	}
	w.linef("%s = %s", lhsCode, rhs)
}

func lhsHasNamespacedName(lhs goivy.Expr) bool {
	return strings.Contains(lhsRootName(lhs), ":")
}

func lhsRootName(lhs goivy.Expr) string {
	for {
		switch n := lhs.(type) {
		case *goivy.Apply:
			lhs = n.Func
		case *goivy.Const:
			return n.Name
		default:
			return goivy.ExprName(n)
		}
	}
}

func (g *Generator) goTraceLHSExpr(lhs goivy.Expr) (string, error) {
	switch n := lhs.(type) {
	case *goivy.Const:
		return fmt.Sprintf("%q", n.Name), nil
	case *goivy.LogicVariable:
		return fmt.Sprintf("%q", n.Name), nil
	case *goivy.Apply:
		name := goivy.ExprName(n.Func)
		if name == "" {
			return fmt.Sprintf("%q", n.String()), nil
		}
		if g.isDestructorName(name) && len(n.Terms) > 0 {
			base, err := g.goTraceLHSExpr(n.Terms[0])
			if err != nil {
				return "", err
			}
			res := base + " + " + fmt.Sprintf("%q", "."+memName(name))
			if len(n.Terms) > 1 {
				args, err := g.goTraceArgsExpr(n.Terms[1:])
				if err != nil {
					return "", err
				}
				res += " + " + args
			}
			return res, nil
		}
		if len(n.Terms) == 0 {
			return fmt.Sprintf("%q", name), nil
		}
		args, err := g.goTraceArgsExpr(n.Terms)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%q", name) + " + " + args, nil
	default:
		code, err := g.emitExpr(lhs)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%q", code), nil
	}
}

func (g *Generator) goTraceArgsExpr(args []goivy.Expr) (string, error) {
	parts := make([]string, 0, len(args)*2+2)
	parts = append(parts, fmt.Sprintf("%q", "("))
	for i, arg := range args {
		if i > 0 {
			parts = append(parts, fmt.Sprintf("%q", ","))
		}
		code, err := g.emitExpr(arg)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("ivyTraceValue(%s, %t)", code, g.traceHex()))
	}
	parts = append(parts, fmt.Sprintf("%q", ")"))
	return strings.Join(parts, " + "), nil
}

func (g *Generator) traceHex() bool {
	if g == nil || g.Mod == nil || g.Mod.Attributes == nil {
		return false
	}
	raw, ok := g.Mod.Attributes.Get2("radix")
	if !ok {
		return false
	}
	return traceHexValue(raw) == "16"
}

func traceHexValue(raw any) string {
	switch v := raw.(type) {
	case string:
		return strings.Trim(v, `"`)
	case interface{ Relname() string }:
		return v.Relname()
	case goivy.Expr:
		return string(v.Sexp())
	default:
		return fmt.Sprint(v)
	}
}

func (g *Generator) emitSet(w *goWriter, a *goivy.LogicSetAction) {
	if a.Lit == nil {
		g.unsupported(w, "unsupported set literal: nil")
		return
	}
	target, value := setTargetAndValue(a.Lit)
	vs := goivy.VariablesAstList(target)
	if len(vs) == 0 {
		g.emitAssignOne(w, target, goivy.NewConst(value, goivy.Boolean))
		return
	}
	g.pushScope()
	opened := 0
	for _, v := range vs {
		header, ok, err := g.goLoopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported set over free variable %s:%s", goName(v.Name), sortName(v.VSort))
			for i := 0; i < opened; i++ {
				w.close("")
			}
			g.popScope()
			return
		}
		if !ok {
			for i := 0; i < opened; i++ {
				w.close("")
			}
			g.popScope()
			g.emitAssign(w, goivy.NewAssignAction(target, goivy.NewConst(value, goivy.Boolean)))
			return
		}
		g.addLocal(v.Name)
		w.open(header)
		opened++
	}
	g.emitAssignOne(w, target, goivy.NewConst(value, goivy.Boolean))
	for i := 0; i < opened; i++ {
		w.close("")
	}
	g.popScope()
}

func (g *Generator) emitNativeAction(w *goWriter, a *goivy.LogicNativeAction) {
	if a == nil {
		return
	}
	if len(g.currentReturns) > 0 {
		allRuntimeHandles := true
		for _, ret := range g.currentReturns {
			if ret == nil || !g.isRuntimeHandleSort(ret.CSort) {
				allRuntimeHandles = false
				break
			}
		}
		if allRuntimeHandles {
			w.line("// ivy native socket-handle action omitted: Go returns the zero runtime handle")
			return
		}
	}
	code := nativeActionCode(a)
	g.unsupportedAt(w, a.GetLineno(), "native C++ action code is not translated to Go: %s", strings.TrimSpace(code))
	for _, line := range strings.Split(code, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		w.line("// native: " + line)
	}
}

func (g *Generator) emitDebug(w *goWriter, a *goivy.LogicDebugAction) {
	event := debugEventName(a.DebugExpr)
	loc := a.GetLineno()
	w.line(`fmt.Fprintln(__ivy_out, "{")`)
	w.linef("fmt.Fprintf(__ivy_out, %q, %q)", "    \"event\" : %q,\n", event)
	for i, e := range a.WithExprs {
		name := ""
		if i < len(a.WithNames) {
			name = a.WithNames[i]
		}
		if name == "" {
			if derived, ok := exprNameOK(e); ok && derived != "" {
				name = derived
			} else {
				name = fmt.Sprintf("value%d", i)
			}
		}
		w.linef("fmt.Fprintf(__ivy_out, %q, %q)", "    %q : ", name)
		g.emitPrintExprAt(w, e, loc)
		w.line(`fmt.Fprintln(__ivy_out, ",")`)
	}
	w.line(`fmt.Fprintln(__ivy_out, "}")`)
}

func (g *Generator) emitPrintExpr(w *goWriter, expr goivy.Expr) {
	g.emitPrintExprAt(w, expr, goivy.Location{})
}

func (g *Generator) emitPrintExprAt(w *goWriter, expr goivy.Expr, loc goivy.Location) {
	vs := goivy.VariablesAstList(expr)
	var closers []string
	for _, v := range vs {
		idx := g.nextTemp("__ivy_debug_idx")
		loop, ok, err := g.goIndexedLoopHeaderForVar(v, idx)
		if err != nil || !ok {
			g.unsupportedAt(w, loc, "unsupported debug print variable %s:%s", goName(v.Name), sortName(v.VSort))
			closePrintExprLoops(w, closers)
			return
		}
		if loop.pre != "" {
			w.line(loop.pre)
		}
		w.line(`fmt.Fprint(__ivy_out, "[")`)
		w.open(loop.header)
		w.open(fmt.Sprintf("if %s > 0 {", idx))
		w.line(`fmt.Fprint(__ivy_out, ",")`)
		w.close("")
		closers = append(closers, loop.post)
	}
	value, err := g.emitExpr(expr)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported debug print expression: %s", err.Error())
		closePrintExprLoops(w, closers)
		return
	}
	w.linef("fmt.Fprint(__ivy_out, %s)", value)
	closePrintExprLoops(w, closers)
}

type goIndexedLoopHeader struct {
	pre    string
	header string
	post   string
}

func (g *Generator) goLoopHeaderForVar(v *goivy.LogicVariable) (string, bool, error) {
	if v == nil {
		return "", false, fmt.Errorf("ivy2golang: nil loop variable")
	}
	return g.goLoopHeaderForSort(v.VSort, goName(v.Name))
}

func (g *Generator) goLoopHeaderForSort(s goivy.Sort, name string) (string, bool, error) {
	if vals, ok := goLiteralFiniteValueExprs(s); ok {
		return fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(s), strings.Join(vals, ", ")), true, nil
	}
	if header, ok, err := g.goFiniteLoopHeaderForSort(s, name); err != nil || ok {
		return header, ok, err
	}
	if vals, ok := g.finiteValueExprs(s); ok {
		return fmt.Sprintf("for _, %s := range []%s{%s} {", name, g.goScalarType(s), strings.Join(vals, ", ")), true, nil
	}
	return "", false, nil
}

func (g *Generator) goIndexedLoopHeaderForVar(v *goivy.LogicVariable, idx string) (goIndexedLoopHeader, bool, error) {
	if v == nil {
		return goIndexedLoopHeader{}, false, fmt.Errorf("ivy2golang: nil loop variable")
	}
	if vals, ok := goLiteralFiniteValueExprs(v.VSort); ok {
		return goIndexedLoopHeader{
			header: fmt.Sprintf("for %s, %s := range []%s{%s} {", idx, goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")),
		}, true, nil
	}
	if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, goName(v.Name)); err != nil || ok {
		if err != nil || !ok {
			return goIndexedLoopHeader{}, ok, err
		}
		return goIndexedLoopHeader{
			pre:    fmt.Sprintf("%s := 0", idx),
			header: header,
			post:   idx + "++",
		}, true, nil
	}
	if vals, ok := g.finiteValueExprs(v.VSort); ok {
		return goIndexedLoopHeader{
			header: fmt.Sprintf("for %s, %s := range []%s{%s} {", idx, goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")),
		}, true, nil
	}
	return goIndexedLoopHeader{}, false, nil
}

func goLiteralFiniteValueExprs(s goivy.Sort) ([]string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return []string{"false", "true"}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			if isNumericEnum(st) {
				vals[i] = v
			} else {
				vals[i] = goName(v)
			}
		}
		return vals, true
	default:
		return nil, false
	}
}

func closePrintExprLoops(w *goWriter, closers []string) {
	for i := len(closers) - 1; i >= 0; i-- {
		if closers[i] != "" {
			w.line(closers[i])
		}
		w.close("")
		w.line(`fmt.Fprint(__ivy_out, "]")`)
	}
}

func setTargetAndValue(lit goivy.Expr) (goivy.Expr, string) {
	switch n := lit.(type) {
	case *goivy.LogicLiteral:
		if n.Polarity == 0 {
			return n.Atom, "false"
		}
		return n.Atom, "true"
	case *goivy.LogicNot:
		return n.Body, "false"
	default:
		return lit, "true"
	}
}

func (g *Generator) emitAssertLike(w *goWriter, fn string, f goivy.Expr, label string) {
	expr, err := g.emitExpr(closeFormulaForGo(f))
	if err != nil {
		kind := "unsupported assertion expression"
		if fn == "ivyAssume" {
			kind = "unsupported assumption expression"
		}
		g.softUnsupported(w, kind, err, label)
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef("%s(%s, %q)", fn, expr, label)
}

func (g *Generator) emitAssume(w *goWriter, f goivy.Expr, label string) {
	expr, err := g.emitExpr(closeFormulaForGo(f))
	if err != nil {
		g.softUnsupported(w, "unsupported assumption expression", err, label)
		return
	}
	if strings.TrimSpace(label) == "" {
		label = "assume"
	}
	w.open("if !ivyAssume(" + expr + ", " + strconv.Quote(label) + ") {")
	if len(g.currentReturns) > 0 {
		names := make([]string, len(g.currentReturns))
		for i, r := range g.currentReturns {
			names[i] = goName(r.Name)
		}
		w.linef("return %s", strings.Join(names, ", "))
	} else {
		w.line("return")
	}
	w.close("")
}

func closeFormulaForGo(f goivy.Expr) goivy.Expr {
	closed := goivy.CloseFormula(f)
	var vars []*goivy.LogicVariable
	seen := map[string]bool{}
	for _, sym := range goivy.UsedSymbolsAst(f).All() {
		c, ok := sym.(*goivy.Const)
		if !ok || c == nil || c.CSort == nil || !goParamConstName(c.Name) || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		vars = append(vars, &goivy.LogicVariable{Name: c.Name, VSort: c.CSort})
	}
	if len(vars) == 0 {
		return closed
	}
	return goivy.IvyForAll(vars, closed)
}

func goParamConstName(name string) bool {
	return strings.HasPrefix(name, "prm:") || strings.HasPrefix(name, "__prm:")
}

func (g *Generator) emitIf(w *goWriter, a *goivy.LogicIfAction) {
	if some, ok := a.Cond.(*goivy.SomeCondition); ok {
		g.emitIfSome(w, a, some)
		return
	}
	cond, err := g.emitExpr(a.Cond)
	if err != nil {
		g.softUnsupported(w, "unsupported if condition", err, linenoStr(a.GetLineno()))
		return
	}
	w.open("if " + cond + " {")
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	if a.ElseBody != nil {
		w.close(" else {")
		w.indent++
		if elseAct, ok := a.ElseBody.(goivy.Action); ok {
			g.emitAction(w, elseAct)
		}
		w.indent--
		w.line("}")
		return
	}
	w.close("")
}

func (g *Generator) emitWhile(w *goWriter, a *goivy.LogicWhileAction) {
	cond, err := g.emitExpr(a.Cond)
	if err != nil {
		g.softUnsupported(w, "unsupported while condition", err, linenoStr(a.GetLineno()))
		return
	}
	w.open("for " + cond + " {")
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
}

func (g *Generator) emitIfSome(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) {
	if some.Kind == "some_min" || some.Kind == "some_max" {
		if err := g.emitIfSomeMinMax(w, a, some); err != nil {
			g.softUnsupported(w, "unsupported if condition", err, linenoStr(a.GetLineno()))
		}
		return
	}
	if g.emitIfSomeVariantDowncast(w, a, some) {
		return
	}
	if g.emitIfSomeExtensional(w, a, some) {
		return
	}
	if ok, err := g.emitIfSomeFinite(w, a, some); ok || err != nil {
		if err != nil {
			g.softUnsupported(w, "unsupported if condition", err, linenoStr(a.GetLineno()))
		}
		return
	}
	g.softUnsupported(w, "unsupported if condition", fmt.Errorf("%s", some.String()), linenoStr(a.GetLineno()))
}

func (g *Generator) emitIfSomeExtensional(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {
	if some == nil || len(some.Params) != 1 {
		return false
	}
	p := some.Params[0]
	if p == nil {
		return false
	}
	boundVar, err := goivy.NewVariable("X"+p.Name, p.CSort)
	if err != nil {
		return false
	}
	subs := map[goivy.NodeKey]goivy.Expr{goivy.Key(p): boundVar}
	fmla, err := goivy.Substitute(some.Fmla, subs)
	if err != nil {
		return false
	}
	var ebnds []*goivy.Apply
	g.matchExtensionalBoundExprs(boundVar, fmla, true, &ebnds)
	if len(ebnds) == 0 {
		return false
	}
	app := ebnds[0]
	fs, ok := app.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || !isBooleanSort(fs.Range()) {
		return false
	}
	argIndex := -1
	for i, t := range app.Terms {
		if tv, ok := t.(*goivy.LogicVariable); ok && tv.Name == boundVar.Name {
			argIndex = i
			break
		}
	}
	if argIndex < 0 {
		return false
	}
	relName := goivy.ExprName(app.Func)
	if relName == "" {
		return false
	}
	g.pushScope()
	g.addLocal(p.Name)
	cond, err := g.emitExpr(some.Fmla)
	if err != nil {
		g.popScope()
		g.softUnsupported(w, "unsupported if condition", err, linenoStr(a.GetLineno()))
		return true
	}
	found := g.nextTemp("__ivy_some")
	key := g.nextTemp("__ivy_some_key")
	val := g.nextTemp("__ivy_some_val")
	w.linef("%s := false", found)
	w.open(fmt.Sprintf("for %s, %s := range %s {", key, val, g.goStorageRangeExpr(relName, fs, "ivy")))
	w.open(fmt.Sprintf("if !%s {", val))
	w.line("continue")
	w.close("")
	if len(app.Terms) == 1 {
		w.linef("%s := %s", goName(p.Name), key)
	} else {
		w.linef("%s := %s.A%d", goName(p.Name), key, argIndex)
	}
	w.linef("_ = %s", goName(p.Name))
	w.open(fmt.Sprintf("if !%s && (%s) {", found, cond))
	w.linef("%s = true", found)
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	w.close("")
	w.close("")
	g.popScope()
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if !%s {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
	return true
}

func (g *Generator) emitIfSomeVariantDowncast(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {
	if len(some.Params) != 1 {
		return false
	}
	v := some.Params[0]
	if v == nil {
		return false
	}
	app, ok := some.Fmla.(*goivy.Apply)
	if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 {
		return false
	}
	if goivy.ExprName(app.Terms[1]) != v.Name || g == nil || g.Mod == nil || !g.Mod.IsVariant(app.Terms[0].NodeSort(), v.CSort) {
		return false
	}
	lhs, err := g.emitExpr(app.Terms[0])
	if err != nil {
		g.softUnsupported(w, "unsupported if condition", err, linenoStr(a.GetLineno()))
		return true
	}
	idx := g.Mod.VariantIndex(app.Terms[0].NodeSort(), v.CSort)
	w.open(fmt.Sprintf("if %s.valid && %s.tag == %d {", lhs, lhs, idx))
	g.pushScope()
	g.addLocal(v.Name)
	w.linef("%s := %s", goName(v.Name), g.variantDowncastExpr(lhs, v.CSort))
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	g.popScope()
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.close(" else {")
		g.emitAction(w, elseAct)
		w.close("")
		return true
	}
	w.close("")
	return true
}

func (g *Generator) emitIfSomeFinite(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) (bool, error) {
	if some == nil || len(some.Params) == 0 {
		return false, nil
	}
	headers, err := g.goLoopHeadersForSomeCondition(some)
	if err != nil {
		return false, nil
	}
	g.pushScope()
	for _, p := range some.Params {
		g.addLocal(p.Name)
	}
	cond, err := g.emitExpr(some.Fmla)
	if err != nil {
		g.popScope()
		return true, err
	}
	found := g.nextTemp("__ivy_some")
	w.linef("%s := false", found)
	for i, p := range some.Params {
		_ = p
		w.open(headers[i])
	}
	for _, p := range some.Params {
		w.linef("_ = %s", goName(p.Name))
	}
	w.open(fmt.Sprintf("if !%s && (%s) {", found, cond))
	w.linef("%s = true", found)
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	w.close("")
	for range some.Params {
		w.close("")
	}
	g.popScope()
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if !%s {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
	return true, nil
}

func (g *Generator) emitIfSomeMinMax(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) error {
	if some == nil || some.Index == nil {
		return fmt.Errorf("missing index expression")
	}
	if len(some.Params) == 0 {
		return fmt.Errorf("empty some condition")
	}
	headers, err := g.goLoopHeadersForSomeCondition(some)
	if err != nil {
		return err
	}
	g.pushScope()
	for _, p := range some.Params {
		g.addLocal(p.Name)
	}
	cond, err := g.emitExpr(some.Fmla)
	if err != nil {
		g.popScope()
		return err
	}
	idxExpr, err := g.emitExpr(some.Index)
	if err != nil {
		g.popScope()
		return err
	}
	idxSort := some.Index.NodeSort()
	found := g.nextTemp("__ivy_some")
	bestIdx := g.nextTemp("__ivy_some_idx")
	curIdx := g.nextTemp("__ivy_some_cur")
	w.linef("%s := false", found)
	w.linef("var %s %s = %s", bestIdx, g.goScalarType(idxSort), g.goZeroValue(idxSort))
	witnesses := make([]string, len(some.Params))
	for i, p := range some.Params {
		witnesses[i] = goName(fmt.Sprintf("__ivy_some_w%d_%s", i, p.Name))
		w.linef("var %s %s = %s", witnesses[i], g.goScalarType(p.CSort), g.goZeroValue(p.CSort))
	}
	for i := range some.Params {
		w.open(headers[i])
	}
	for _, p := range some.Params {
		w.linef("_ = %s", goName(p.Name))
	}
	w.open("if " + cond + " {")
	w.linef("%s := %s", curIdx, idxExpr)
	cmp := g.goLessExpr(curIdx, bestIdx, idxSort)
	if some.Kind == "some_max" {
		cmp = g.goLessExpr(bestIdx, curIdx, idxSort)
	}
	w.open(fmt.Sprintf("if !%s || (%s) {", found, cmp))
	w.linef("%s = true", found)
	w.linef("%s = %s", bestIdx, curIdx)
	for i, p := range some.Params {
		w.linef("%s = %s", witnesses[i], goName(p.Name))
	}
	w.close("")
	if firstParamIsIndex(some) {
		w.line("break")
	}
	w.close("")
	for range some.Params {
		w.close("")
	}
	g.popScope()
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if %s {", found))
		g.pushScope()
		for i, p := range some.Params {
			g.addLocal(p.Name)
			w.linef("%s := %s", goName(p.Name), witnesses[i])
			w.linef("_ = %s", goName(p.Name))
		}
		g.emitAction(w, thenAct)
		g.popScope()
		w.close("")
	}
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if !%s {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
	return nil
}

func firstParamIsIndex(some *goivy.SomeCondition) bool {
	if some == nil || len(some.Params) == 0 || some.Params[0] == nil || some.Index == nil {
		return false
	}
	pname := some.Params[0].Name
	switch x := some.Index.(type) {
	case *goivy.LogicVariable:
		return x.Name == pname
	case *goivy.Const:
		return x.Name == pname
	default:
		return false
	}
}

func (g *Generator) emitChoice(w *goWriter, a *goivy.LogicChoiceAction) {
	if len(a.Branches) == 0 {
		return
	}
	if len(a.Branches) == 1 {
		if act, ok := a.Branches[0].(goivy.Action); ok {
			g.emitAction(w, act)
		}
		return
	}
	tmp := g.nextTemp("__ivy_branch")
	w.linef("%s := ivy.___ivy_choose(0, %q, %d)", tmp, "___branch", a.UniqueID)
	for i, branch := range a.Branches {
		switch {
		case i == 0:
			w.open(fmt.Sprintf("if %s == %d {", tmp, i))
		case i < len(a.Branches)-1:
			w.close(fmt.Sprintf(" else if %s == %d {", tmp, i))
			w.indent++
		default:
			w.close(" else {")
			w.indent++
		}
		if act, ok := branch.(goivy.Action); ok {
			g.emitAction(w, act)
		}
	}
	w.close("")
}

func (g *Generator) emitCall(w *goWriter, a *goivy.LogicCallAction) {
	loc := a.GetLineno()
	name := a.CalleeName()
	args := callArgs(a.Callee)
	argCodes := make([]string, len(args))
	var formals []*goivy.Const
	var returns []*goivy.Const
	if g != nil && g.Mod != nil && g.Mod.Actions != nil {
		if callee, ok := g.Mod.Actions.Get2(name); ok && callee != nil {
			formals = callee.GetFormalParams()
			returns = callee.GetFormalReturns()
		}
	}
	for i, arg := range args {
		code, err := g.emitExpr(arg)
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported call argument: %s", err.Error())
			return
		}
		if i < len(formals) {
			code = g.maybeVariantUpcastExpr(formals[i].CSort, arg.NodeSort(), code)
		}
		argCodes[i] = code
	}
	if g.isTestImportCallback(name) {
		if len(a.ActualReturns) == 0 {
			return
		}
		for i, ret := range a.ActualReturns {
			rhs := "0"
			if i < len(returns) {
				rhs = g.goZeroValue(returns[i].CSort)
			}
			if !g.emitCallReturnAssignAt(w, ret, rhs, loc) {
				return
			}
		}
		return
	}
	fn, err := funName(name)
	if err != nil {
		g.unsupportedAt(w, loc, "%s", err.Error())
		return
	}
	call := fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(argCodes, ", "))
	if len(a.ActualReturns) == 0 {
		w.linef("ivy.___ivy_push(%q)", name)
		w.line(call)
		w.line("ivy.___ivy_pop()")
		return
	}
	if len(a.ActualReturns) == 1 {
		tmp := goName(g.nextTemp("__ivy_ret"))
		w.linef("ivy.___ivy_push(%q)", name)
		w.linef("%s := %s", tmp, call)
		w.line("ivy.___ivy_pop()")
		g.emitCallReturnAssignAt(w, a.ActualReturns[0], tmp, loc)
		return
	}
	tmpNames := make([]string, len(a.ActualReturns))
	for i := range a.ActualReturns {
		tmpNames[i] = goName(g.nextTemp("__ivy_ret"))
	}
	w.linef("ivy.___ivy_push(%q)", name)
	w.linef("%s := %s", strings.Join(tmpNames, ", "), call)
	w.line("ivy.___ivy_pop()")
	for i, ret := range a.ActualReturns {
		if !g.emitCallReturnAssignAt(w, ret, tmpNames[i], loc) {
			return
		}
	}
}

func (g *Generator) emitCallReturnAssign(w *goWriter, target goivy.Expr, rhs string) bool {
	return g.emitCallReturnAssignAt(w, target, rhs, goivy.Location{})
}

func (g *Generator) emitCallReturnAssignAt(w *goWriter, target goivy.Expr, rhs string, loc goivy.Location) bool {
	if call, ok, err := g.goStorageSet(target, rhs); ok || err != nil {
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported call return: %s", err.Error())
			return false
		}
		w.line(call)
		return true
	}
	lhs, err := g.emitExpr(target)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported call return: %s", err.Error())
		return false
	}
	w.linef("%s = %s", lhs, rhs)
	return true
}

func (g *Generator) isStorageSettable(target goivy.Expr) (bool, error) {
	_, ok, err := g.goStorageSet(target, "__ivy_value")
	return ok, err
}

func callArgs(callee goivy.Expr) []goivy.Expr {
	if app, ok := callee.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

func (g *Generator) emitLocal(w *goWriter, a *goivy.LogicLocalAction) {
	loc := goivy.Location{}
	if a != nil {
		loc = a.GetLineno()
	}
	w.open("{")
	g.pushScope()
	type localDecl struct {
		name string
		sort goivy.Sort
	}
	var locals []localDecl
	for _, local := range a.Locals {
		name, ok := exprNameOK(local)
		if name == "" {
			ok = false
		}
		if !ok {
			g.unsupportedAt(w, loc, "unsupported local declaration %T: %s", local, fmt.Sprint(local))
			continue
		}
		sort := local.NodeSort()
		g.addLocalSort(name, sort)
		locals = append(locals, localDecl{name: name, sort: sort})
		if g.emitLocalFunctionNondetAt(w, name, sort, a.UniqueID, loc) {
			continue
		}
		init := g.goZeroValue(sort)
		if expr, err := g.goLocalNondetValueExprAt(sort, name, a.UniqueID, loc); err == nil {
			init = expr
		}
		localName := goName(name)
		w.linef("%s := %s", localName, init)
		w.linef("_ = %s", localName)
	}
	for _, local := range locals {
		if g.emitLocalWitnessInit(w, local.name, local.sort, a.Body) {
			continue
		}
		g.emitLocalFiniteWitnessInit(w, local.name, local.sort, a.Body)
	}
	if body, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, body)
	}
	g.popScope()
	w.close("")
}

func (g *Generator) emitLocalWitnessInit(w *goWriter, localName string, localSort goivy.Sort, body goivy.Expr) bool {
	if g == nil || w == nil || localName == "" || body == nil {
		return false
	}
	app, pos, ok := g.localWitnessBoundApply(localName, localSort, body)
	if !ok {
		return false
	}
	fs, ok := app.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || !isBooleanSort(fs.Range()) {
		return false
	}
	relName := goivy.ExprName(app.Func)
	if relName == "" {
		return false
	}
	var conds []string
	for i, term := range app.Terms {
		if i == pos {
			continue
		}
		expr, err := g.emitExpr(term)
		if err != nil {
			return false
		}
		keyExpr := "__ivy_witness_key"
		if len(app.Terms) != 1 {
			keyExpr = fmt.Sprintf("__ivy_witness_key.A%d", i)
		}
		conds = append(conds, fmt.Sprintf("(%s == %s)", keyExpr, expr))
	}
	assignExpr := "__ivy_witness_key"
	if len(app.Terms) != 1 {
		assignExpr = fmt.Sprintf("__ivy_witness_key.A%d", pos)
	}
	w.open(fmt.Sprintf("for __ivy_witness_key, __ivy_witness_val := range %s {", g.goStorageRangeExpr(relName, fs, "ivy")))
	w.open("if !__ivy_witness_val {")
	w.line("continue")
	w.close("")
	if len(conds) > 0 {
		w.open("if " + strings.Join(conds, " && ") + " {")
	}
	w.linef("%s = %s", goName(localName), assignExpr)
	w.line("break")
	if len(conds) > 0 {
		w.close("")
	}
	w.close("")
	return true
}

func (g *Generator) emitLocalFiniteWitnessInit(w *goWriter, localName string, localSort goivy.Sort, body goivy.Expr) bool {
	if g == nil || w == nil || body == nil || localName == "" {
		return false
	}
	if g.Config.Target != "test" && g.Config.Target != "gen" {
		return false
	}
	values, ok := g.finiteValueExprs(localSort)
	if !ok || len(values) == 0 || len(values) > goLargeThresh {
		return false
	}
	act, ok := body.(goivy.Action)
	if !ok {
		return false
	}
	names := map[string]bool{localName: true}
	var conds []string
	for _, f := range localWitnessAssumeFormulas(act) {
		if !exprReferencesAnyName(f, names) {
			continue
		}
		cond, err := g.emitExpr(closeFormulaForGo(f))
		if err != nil {
			return false
		}
		conds = append(conds, cond)
	}
	if len(conds) == 0 {
		return false
	}
	base := goName(localName)
	valuesName := "__ivy_local_witness_values_" + base
	candidateName := "__ivy_local_witness_" + base
	w.linef("%s := []%s{%s}", valuesName, g.goType(localSort), strings.Join(values, ", "))
	w.open(fmt.Sprintf("for _, %s := range %s {", candidateName, valuesName))
	w.linef("%s = %s", base, candidateName)
	w.open("if " + strings.Join(conds, " && ") + " {")
	w.line("break")
	w.close("")
	w.close("")
	return true
}

func (g *Generator) localWitnessBoundApply(localName string, localSort goivy.Sort, body goivy.Expr) (*goivy.Apply, int, bool) {
	act, ok := body.(goivy.Action)
	if !ok {
		return nil, -1, false
	}
	for _, f := range localWitnessAssumeFormulas(act) {
		if app, pos, ok := g.findLocalWitnessApply(localName, localSort, f); ok {
			return app, pos, true
		}
	}
	return nil, -1, false
}

func localWitnessAssumeFormulas(act goivy.Action) []goivy.Expr {
	switch a := act.(type) {
	case *goivy.LogicAssumeAction:
		if a.Formula == nil {
			return nil
		}
		return []goivy.Expr{a.Formula}
	case *goivy.LogicAssertAction:
		return nil
	case *goivy.LogicSequence:
		var guards []goivy.Expr
		for _, elem := range a.Elems {
			child, ok := elem.(goivy.Action)
			if !ok {
				break
			}
			switch child.(type) {
			case *goivy.LogicAssumeAction, *goivy.LogicAssertAction, *goivy.LogicSequence:
				guards = append(guards, localWitnessAssumeFormulas(child)...)
			default:
				return guards
			}
		}
		return guards
	default:
		return nil
	}
}

func (g *Generator) findLocalWitnessApply(localName string, localSort goivy.Sort, f goivy.Expr) (*goivy.Apply, int, bool) {
	switch n := f.(type) {
	case *goivy.LogicAnd:
		for _, term := range n.Terms {
			if app, pos, ok := g.findLocalWitnessApply(localName, localSort, term); ok {
				return app, pos, true
			}
		}
	case *goivy.LogicLiteral:
		if n.Polarity != 0 {
			return g.findLocalWitnessApply(localName, localSort, n.Atom)
		}
	case *goivy.Apply:
		name := goivy.ExprName(n.Func)
		if name == "" || !g.quantifierSupportRels()[name] {
			return nil, -1, false
		}
		fs, ok := n.Func.NodeSort().(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) != len(n.Terms) || !isBooleanSort(fs.Range()) {
			return nil, -1, false
		}
		pos := -1
		for i, term := range n.Terms {
			if !exprHasName(term, localName) {
				continue
			}
			if pos >= 0 || !sortsEqual(fs.Domain()[i], localSort) {
				return nil, -1, false
			}
			pos = i
		}
		if pos >= 0 {
			return n, pos, true
		}
	}
	return nil, -1, false
}

func exprHasName(e goivy.Expr, name string) bool {
	switch n := e.(type) {
	case *goivy.Const:
		return n.Name == name
	case *goivy.LogicVariable:
		return n.Name == name
	default:
		return false
	}
}

func sortsEqual(a, b goivy.Sort) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.String() == b.String()
}

func (g *Generator) emitLocalFunctionNondet(w *goWriter, name string, sort goivy.Sort, id int64) bool {
	return g.emitLocalFunctionNondetAt(w, name, sort, id, goivy.Location{})
}

func (g *Generator) emitLocalFunctionNondetAt(w *goWriter, name string, sort goivy.Sort, id int64, loc goivy.Location) bool {
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		return false
	}
	localName := goName(name)
	w.linef("%s := %s", localName, g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	w.linef("_ = %s", localName)
	if g.localNondetSkipsSort(fs.Range()) {
		return true
	}
	if !g.canEnumerateDomain(fs.Domain()) {
		g.emitLocalNondetThunkBaseSeenAt(w, localName, fs.Domain(), fs.Range(), name, id, map[string]bool{}, loc)
		return true
	}
	g.emitDomainLoops(w, fs.Domain(), func(args []string) {
		expr, err := g.goLocalNondetValueExprAt(fs.Range(), name, id, loc)
		if err != nil {
			g.unsupportedAt(w, loc, "unsupported local function range: %s", err.Error())
			return
		}
		w.linef("%s = %s", g.goStorageAccess(name, sort, args, ""), expr)
	})
	return true
}

func (g *Generator) goLocalNondetValueExpr(s goivy.Sort, name string, id int64) (string, error) {
	return g.goLocalNondetValueExprAt(s, name, id, goivy.Location{})
}

func (g *Generator) goLocalNondetValueExprAt(s goivy.Sort, name string, id int64, loc goivy.Location) (string, error) {
	return g.goLocalNondetValueExprSeenAt(s, name, id, map[string]bool{}, loc)
}

func (g *Generator) goLocalNondetValueExprSeen(s goivy.Sort, name string, id int64, seen map[string]bool) (string, error) {
	return g.goLocalNondetValueExprSeenAt(s, name, id, seen, goivy.Location{})
}

func (g *Generator) goLocalNondetValueExprSeenAt(s goivy.Sort, name string, id int64, seen map[string]bool, loc goivy.Location) (string, error) {
	call := fmt.Sprintf("ivy.___ivy_choose(0, %q, %d)", name, id)
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.goLocalNondetFunctionValueExprSeenAt(fs, name, id, seen, loc)
	}
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return call + " != 0", nil
	case *goivy.LogicEnumeratedSort:
		if st.Name != "" && !isNumericEnum(st) {
			return fmt.Sprintf("%s(%s)", goName(st.Name), call), nil
		}
		return call, nil
	case *goivy.RangeSort:
		return call, nil
	case *goivy.UninterpretedSort:
		if g.isVariantSuperName(st.Name) {
			return g.goZeroValue(s), nil
		}
		if fields := g.destructorStructFieldInfos(st.Name); len(fields) > 0 {
			key := "struct:" + st.Name
			if seen[key] {
				return g.goZeroValue(s), nil
			}
			seen[key] = true
			defer delete(seen, key)
			return g.goLocalNondetStructValueExprWithSeenAt(s, fields, name, id, seen, loc)
		}
		if g.localNondetSkipsSort(st) {
			return g.goZeroValue(s), nil
		}
		return call, nil
	default:
		return "", fmt.Errorf("unsupported local nondet sort %s", sortName(s))
	}
}

func (g *Generator) localNondetSkipsSort(s goivy.Sort) bool {
	if g == nil || s == nil {
		return false
	}
	if g.isNativeTypeSort(s) || g.hasStringInterp(s) {
		return true
	}
	if it, ok := g.goInterpType(s); ok {
		return it.Kind == goInterpStrBV || it.Kind == goInterpIntBV
	}
	return false
}

func (g *Generator) goLocalNondetStructValueExprWithSeen(s goivy.Sort, fields []goDestructorField, name string, id int64, seen map[string]bool) (string, error) {
	return g.goLocalNondetStructValueExprWithSeenAt(s, fields, name, id, seen, goivy.Location{})
}

func (g *Generator) goLocalNondetStructValueExprWithSeenAt(s goivy.Sort, fields []goDestructorField, name string, id int64, seen map[string]bool, loc goivy.Location) (string, error) {
	if g.destructorStructNeedsInitFunc(fields) {
		tmp := g.nextTemp("__ivy_rec")
		var w goWriter
		w.raw(fmt.Sprintf("func() %s {\n", g.goScalarType(s)))
		w.indent++
		w.linef("var %s %s", tmp, g.goScalarType(s))
		for i, field := range fields {
			label := name + "." + memName(field.Const.Name)
			domain := field.Sort.Domain()
			if g.destructorFieldIsLarge(field) {
				extra := domain[1:]
				tmpThunk := g.nextTemp("__ivy_thunk")
				w.linef("%s := %s", tmpThunk, g.goFunctionStorageInit(extra, field.Sort.Range()))
				g.emitLocalNondetThunkBaseSeenAt(&w, tmpThunk, extra, field.Sort.Range(), label, int64(i), seen, loc)
				w.linef("%s.%s = &%s", tmp, field.FieldName, tmpThunk)
				continue
			}
			if len(domain) == 1 {
				expr, err := g.goLocalNondetValueExprSeenAt(field.Sort.Range(), label, int64(i), seen, loc)
				if err != nil {
					return "", err
				}
				w.linef("%s.%s = %s", tmp, field.FieldName, expr)
				continue
			}
			extra := domain[1:]
			g.emitDomainLoops(&w, extra, func(args []string) {
				expr, err := g.goLocalNondetValueExprSeenAt(field.Sort.Range(), label, int64(i), seen, loc)
				if err != nil {
					g.unsupportedAt(&w, loc, "unsupported destructor field local nondet value: %s", err.Error())
					return
				}
				w.linef("%s.%s%s = %s", tmp, field.FieldName, goIndexSuffix(args), expr)
			})
		}
		w.linef("return %s", tmp)
		w.indent--
		w.raw("}()")
		return w.String(), nil
	}
	inits := make([]string, 0, len(fields))
	for i, field := range fields {
		expr, err := g.goLocalNondetValueExprSeenAt(field.Sort.Range(), name+"."+memName(field.Const.Name), int64(i), seen, loc)
		if err != nil {
			return "", err
		}
		inits = append(inits, field.FieldName+": "+expr)
	}
	return fmt.Sprintf("%s{%s}", g.goScalarType(s), strings.Join(inits, ", ")), nil
}

func (g *Generator) goLocalNondetFunctionValueExprSeen(fs *goivy.LogicFunctionSort, name string, id int64, seen map[string]bool) (string, error) {
	return g.goLocalNondetFunctionValueExprSeenAt(fs, name, id, seen, goivy.Location{})
}

func (g *Generator) goLocalNondetFunctionValueExprSeenAt(fs *goivy.LogicFunctionSort, name string, id int64, seen map[string]bool, loc goivy.Location) (string, error) {
	domain := fs.Domain()
	rng := fs.Range()
	if len(domain) == 0 {
		return g.goLocalNondetValueExprSeenAt(rng, name, id, seen, loc)
	}
	st := g.goFunctionStorageFor(domain, rng)
	tmp := g.nextTemp("__ivy_fn")
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", st.Type))
	w.indent++
	w.linef("%s := %s", tmp, g.goFunctionStorageInit(domain, rng))
	if st.Large {
		g.emitLocalNondetThunkBaseSeenAt(&w, tmp, domain, rng, name, id, seen, loc)
	} else {
		g.emitDomainLoops(&w, domain, func(args []string) {
			expr, err := g.goLocalNondetValueExprSeenAt(rng, name, id, seen, loc)
			if err != nil {
				g.unsupportedAt(&w, loc, "unsupported function-sorted local nondet value range: %s", err.Error())
				return
			}
			w.linef("%s[%s] = %s", tmp, g.goMapKeyValue(domain, args), expr)
		})
	}
	w.linef("return %s", tmp)
	w.indent--
	w.raw("}()")
	return w.String(), nil
}

func (g *Generator) emitLocalNondetThunkBaseSeen(w *goWriter, base string, domain []goivy.Sort, rng goivy.Sort, label string, id int64, seen map[string]bool) {
	g.emitLocalNondetThunkBaseSeenAt(w, base, domain, rng, label, id, seen, goivy.Location{})
}

func (g *Generator) emitLocalNondetThunkBaseSeenAt(w *goWriter, base string, domain []goivy.Sort, rng goivy.Sort, label string, id int64, seen map[string]bool, loc goivy.Location) {
	st := g.goFunctionStorageFor(domain, rng)
	if !st.Large {
		return
	}
	expr, err := g.goLocalNondetValueExprSeenAt(rng, label, id, seen, loc)
	if err != nil {
		g.unsupportedAt(w, loc, "unsupported nondet thunk range: %s", err.Error())
		return
	}
	w.open(fmt.Sprintf("%s.base = func(__ivy_key %s) %s {", base, st.KeyType, st.RangeType))
	w.line("_ = __ivy_key")
	w.linef("return %s", expr)
	w.close("")
}

func (g *Generator) emitLet(w *goWriter, a *goivy.LogicLetAction) {
	loc := a.GetLineno()
	prev := g.exprAliases
	next := make(map[string]goivy.Expr, len(prev)+len(a.Bindings))
	for k, v := range prev {
		next[k] = v
	}
	for _, binding := range a.Bindings {
		children := binding.Children()
		if len(children) < 2 {
			g.unsupportedAt(w, loc, "unsupported let binding %T: %s", binding, binding.String())
			continue
		}
		name := goivy.ExprName(children[0])
		if name == "" {
			g.unsupportedAt(w, loc, "unsupported let binding lhs %T: %s", children[0], children[0].String())
			continue
		}
		next[name] = children[1]
	}
	g.exprAliases = next
	defer func() {
		g.exprAliases = prev
	}()
	w.open("{")
	if body, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, body)
		w.close("")
		return
	}
	g.unsupportedAt(w, loc, "unsupported let action body %T", a.Body)
	w.close("")
}

func (g *Generator) emitBindOlds(w *goWriter, a *goivy.LogicBindOldsAction) {
	inner := "<nil>"
	loc := goivy.Location{}
	if a != nil {
		if a.Inner != nil {
			inner = fmt.Sprintf("%T", a.Inner)
		}
		loc = a.GetLineno()
	}
	g.unsupportedAt(w, loc, "bindolds reached emit (Python has no emit_bind_olds): inner=%s", inner)
}

func debugEventName(e goivy.Expr) string {
	if e == nil {
		return "debug"
	}
	name, ok := exprNameOK(e)
	if !ok {
		return "debug"
	}
	name = strings.Trim(name, `"`)
	if strings.TrimSpace(name) == "" {
		return "debug"
	}
	return name
}

func exprNameOK(e goivy.Expr) (string, bool) {
	switch t := e.(type) {
	case *goivy.Const:
		return t.Name, true
	case *goivy.LogicVariable:
		return t.Name, true
	case *goivy.UninterpretedSort:
		return t.Name, true
	case *goivy.BooleanSort:
		return "bool", true
	case *goivy.LogicEnumeratedSort:
		return t.Name, true
	case *goivy.RangeSort:
		return t.Name, true
	case *goivy.TopSort:
		return t.Name, true
	case *goivy.LogicWhenOperator:
		return t.Name, true
	case *goivy.LogicNamedBinder:
		return t.Name, true
	case *goivy.Apply:
		return exprNameOK(t.Func)
	default:
		return "", false
	}
}
