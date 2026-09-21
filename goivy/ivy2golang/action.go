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
		g.unsupported(w, "thunk reached emit (Python ThunkAction has no emit; expected to be desugared upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicInstantiateAction:
		g.unsupported(w, "instantiate reached emit (Python InstantiateAction has no emit; expected to be inlined upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicRanking:
		g.unsupported(w, "ranking reached emit (Python Ranking has no emit; ranking/progress is enforced separately): %s at %s",
			a.String(), a.GetLineno().String())
	default:
		g.unsupported(w, "unsupported action %T: %s", act, act.String())
	}
}

func (g *Generator) emitHavoc(w *goWriter, a *goivy.LogicHavocAction) {
	target := "<nil>"
	if a != nil && a.Target != nil {
		target = a.Target.String()
	}
	lineno := ""
	if a != nil {
		lineno = a.GetLineno().String()
	}
	g.unsupported(w, "havoc reached emit (Python emit_havoc asserts False): %s at %s", target, lineno)
}

func (g *Generator) emitAssign(w *goWriter, a *goivy.LogicAssignAction) {
	if g.emitExtensionalRelationClear(w, a) {
		return
	}
	vs := g.assignmentLoopVars(a.LHS)
	if len(vs) == 0 {
		g.emitAssignOne(w, a.LHS, a.RHS)
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
	lhsCode, err := g.emitExpr(lhs)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitExpr(rhsExpr)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		return
	}
	rhs = g.maybeVariantUpcastExpr(lhs.NodeSort(), rhsExpr.NodeSort(), rhs)
	if g.Config.Trace && !lhsHasNamespacedName(lhs) {
		trace, err := g.goTraceLHSExpr(lhs)
		if err != nil {
			g.unsupported(w, "unsupported assignment trace lhs: %s", err.Error())
			return
		}
		w.linef("fmt.Fprintf(__ivy_out, %q, %s, ivyTraceValue(%s, %t))", "  write(%s,%s)\n", trace, rhs, g.traceHex())
	}
	if call, ok, err := g.goStorageSet(lhs, rhs); ok || err != nil {
		if err != nil {
			g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
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
	switch v := raw.(type) {
	case string:
		return strings.Trim(v, `"`) == "16"
	default:
		return fmt.Sprint(v) == "16"
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
		if err != nil || !ok {
			g.unsupported(w, "unsupported set over free variable %s:%s", goName(v.Name), sortName(v.VSort))
			for i := 0; i < opened; i++ {
				w.close("")
			}
			g.popScope()
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
	g.warnOnce("ivy2golang: native C++ action code is emitted as a Go no-op")
	w.line("// ivy native action omitted: C++ native code is not translated to Go")
	if a == nil {
		return
	}
	code, ok := a.Code.(*goivy.NativeCode)
	if !ok {
		return
	}
	for _, line := range strings.Split(code.Code, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		w.line("// native: " + line)
	}
}

func (g *Generator) emitDebug(w *goWriter, a *goivy.LogicDebugAction) {
	event := debugEventName(a.DebugExpr)
	w.line(`fmt.Fprintln(__ivy_out, "{")`)
	w.linef("fmt.Fprintf(__ivy_out, %q, %q)", "    \"event\" : %q,\n", event)
	for i, e := range a.WithExprs {
		name := ""
		if i < len(a.WithNames) {
			name = a.WithNames[i]
		}
		if name == "" {
			name = goivy.ExprName(e)
		}
		w.linef("fmt.Fprintf(__ivy_out, %q, %q)", "    %q : ", name)
		g.emitPrintExpr(w, e)
		w.line(`fmt.Fprintln(__ivy_out, ",")`)
	}
	w.line(`fmt.Fprintln(__ivy_out, "}")`)
}

func (g *Generator) emitPrintExpr(w *goWriter, expr goivy.Expr) {
	vs := goivy.VariablesAstList(expr)
	var closers []string
	for _, v := range vs {
		idx := g.nextTemp("__ivy_debug_idx")
		loop, ok, err := g.goIndexedLoopHeaderForVar(v, idx)
		if err != nil || !ok {
			g.unsupported(w, "unsupported debug print variable %s:%s", goName(v.Name), sortName(v.VSort))
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
		g.unsupported(w, "unsupported debug print expression: %s", err.Error())
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
	if vals, ok := goLiteralFiniteValueExprs(v.VSort); ok {
		return fmt.Sprintf("for _, %s := range []%s{%s} {", goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")), true, nil
	}
	if header, ok, err := g.goFiniteLoopHeaderForSort(v.VSort, goName(v.Name)); err != nil || ok {
		return header, ok, err
	}
	if vals, ok := g.finiteValueExprs(v.VSort); ok {
		return fmt.Sprintf("for _, %s := range []%s{%s} {", goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")), true, nil
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
		g.unsupported(w, "unsupported assertion expression: %s", err.Error())
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
		g.unsupported(w, "unsupported assumption expression: %s", err.Error())
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
		g.unsupported(w, "unsupported if condition: %s", err.Error())
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
		g.unsupported(w, "unsupported while condition: %s", err.Error())
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
			g.unsupported(w, "unsupported %s condition: %s", some.Kind, err.Error())
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
			g.unsupported(w, "unsupported some condition: %s", err.Error())
		}
		return
	}
	g.unsupported(w, "unsupported some condition: %s", some.String())
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
		g.unsupported(w, "unsupported if condition: %s", err.Error())
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
		g.unsupported(w, "unsupported variant downcast receiver: %s", err.Error())
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
			g.unsupported(w, "unsupported call argument: %s", err.Error())
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
		lhs := make([]string, len(a.ActualReturns))
		rhs := make([]string, len(a.ActualReturns))
		for i, ret := range a.ActualReturns {
			code, err := g.emitExpr(ret)
			if err != nil {
				g.unsupported(w, "unsupported call return: %s", err.Error())
				return
			}
			lhs[i] = code
			if i < len(returns) {
				rhs[i] = g.goZeroValue(returns[i].CSort)
			} else {
				rhs[i] = "0"
			}
		}
		w.linef("%s = %s", strings.Join(lhs, ", "), strings.Join(rhs, ", "))
		return
	}
	fn, err := funName(name)
	if err != nil {
		g.unsupported(w, "%s", err.Error())
		return
	}
	call := fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(argCodes, ", "))
	if len(a.ActualReturns) == 0 {
		w.line(call)
		return
	}
	lhs := make([]string, len(a.ActualReturns))
	for i, ret := range a.ActualReturns {
		code, err := g.emitExpr(ret)
		if err != nil {
			g.unsupported(w, "unsupported call return: %s", err.Error())
			return
		}
		lhs[i] = code
	}
	w.linef("%s = %s", strings.Join(lhs, ", "), call)
}

func callArgs(callee goivy.Expr) []goivy.Expr {
	if app, ok := callee.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

func (g *Generator) emitLocal(w *goWriter, a *goivy.LogicLocalAction) {
	w.open("{")
	g.pushScope()
	for _, local := range a.Locals {
		name := goivy.ExprName(local)
		if name == "" {
			g.unsupported(w, "unsupported local declaration %T: %s", local, local.String())
			continue
		}
		sort := local.NodeSort()
		g.addLocalSort(name, sort)
		if g.emitLocalFunctionNondet(w, name, sort, a.UniqueID) {
			continue
		}
		init := g.goZeroValue(sort)
		if expr, err := g.goLocalNondetValueExpr(sort, name, a.UniqueID); err == nil {
			init = expr
		}
		w.linef("%s := %s", goName(name), init)
	}
	if body, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, body)
	}
	g.popScope()
	w.close("")
}

func (g *Generator) emitLocalFunctionNondet(w *goWriter, name string, sort goivy.Sort, id int64) bool {
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		return false
	}
	w.linef("%s := %s", goName(name), g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	if !g.canEnumerateDomain(fs.Domain()) {
		g.emitNondetThunkBase(w, goName(name), fs.Domain(), fs.Range(), name, id)
		return true
	}
	g.emitDomainLoops(w, fs.Domain(), func(args []string) {
		expr, err := g.goLocalNondetValueExpr(fs.Range(), name, id)
		if err != nil {
			g.unsupported(w, "unsupported local function range: %s", err.Error())
			return
		}
		w.linef("%s = %s", g.goStorageAccess(name, sort, args, ""), expr)
	})
	return true
}

func (g *Generator) goLocalNondetValueExpr(s goivy.Sort, name string, id int64) (string, error) {
	call := fmt.Sprintf("ivy.___ivy_choose(0, %q, %d)", name, id)
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
		if g.isVariantSuperName(st.Name) || g.destructorStructFields(st.Name) != nil {
			return g.goRandomValueExpr(s, name, id)
		}
		if g.hasStringInterp(st) {
			return `""`, nil
		}
		return call, nil
	default:
		return "", fmt.Errorf("unsupported local nondet sort %s", sortName(s))
	}
}

func (g *Generator) emitLet(w *goWriter, a *goivy.LogicLetAction) {
	prev := g.exprAliases
	next := make(map[string]goivy.Expr, len(prev)+len(a.Bindings))
	for k, v := range prev {
		next[k] = v
	}
	for _, binding := range a.Bindings {
		children := binding.Children()
		if len(children) < 2 {
			g.unsupported(w, "unsupported let binding %T: %s", binding, binding.String())
			continue
		}
		name := goivy.ExprName(children[0])
		if name == "" {
			g.unsupported(w, "unsupported let binding lhs %T: %s", children[0], children[0].String())
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
	g.unsupported(w, "unsupported let action body %T", a.Body)
	w.close("")
}

func (g *Generator) emitBindOlds(w *goWriter, a *goivy.LogicBindOldsAction) {
	inner := "<nil>"
	lineno := ""
	if a != nil {
		if a.Inner != nil {
			inner = fmt.Sprintf("%T", a.Inner)
		}
		lineno = a.GetLineno().String()
	}
	g.unsupported(w, "bindolds reached emit (Python has no emit_bind_olds): inner=%s at %s", inner, lineno)
}

func debugEventName(e goivy.Expr) string {
	if e == nil {
		return "debug"
	}
	name := goivy.ExprName(e)
	name = strings.Trim(name, `"`)
	if strings.TrimSpace(name) == "" {
		return "debug"
	}
	return name
}
