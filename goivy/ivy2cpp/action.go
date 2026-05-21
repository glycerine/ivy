package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitAction(w *cppWriter, act goivy.Action) {
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
	case *goivy.LogicHavocAction:
		g.emitHavoc(w, a)
	case *goivy.LogicSetAction:
		g.emitSet(w, a)
	case *goivy.LogicAssertAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, a.GetLineno().String())
	case *goivy.LogicRequiresAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, a.GetLineno().String())
	case *goivy.LogicEnsuresAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, a.GetLineno().String())
	case *goivy.LogicSubgoalAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, a.GetLineno().String())
	case *goivy.LogicAssumeAction:
		g.emitAssertLike(w, "ivy_assume", a.Formula, a.GetLineno().String())
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
	case *goivy.LogicNativeAction:
		g.emitNativeAction(w, a)
	case *goivy.LogicDebugAction:
		g.emitDebug(w, a)
	case *goivy.LogicCrashAction:
		w.line("std::abort();")
	case *goivy.ReturnAction:
		g.emitReturn(w)
	case *goivy.IgnoreAction:
		return
	case *goivy.LogicAssignFieldAction:
		g.emitAssignField(w, a)
	case *goivy.LogicNullFieldAction:
		g.emitNullField(w, a)
	case *goivy.LogicCopyFieldAction:
		g.emitCopyField(w, a)
	default:
		g.unsupported(w, "unsupported action %T: %s", act, act.String())
	}
}

func (g *Generator) emitReturn(w *cppWriter) {
	if g != nil && len(g.currentReturns) == 1 {
		w.linef("return %s;", varName(g.currentReturns[0].Name))
		return
	}
	w.line("return;")
}

func (g *Generator) emitHavoc(w *cppWriter, a *goivy.LogicHavocAction) {
	if a.Target == nil {
		g.unsupported(w, "unsupported havoc target: nil")
		return
	}
	loops, ok := g.openAssignmentLoops(w, a.Target)
	if !ok {
		return
	}
	lhs, err := g.emitExpr(a.Target)
	if err != nil {
		g.unsupported(w, "unsupported havoc target: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	w.linef("%s = %s;", lhs, g.cppZeroValue(a.Target.NodeSort()))
	g.closeAssignmentLoops(w, loops)
}

func (g *Generator) emitSet(w *cppWriter, a *goivy.LogicSetAction) {
	if a.Lit == nil {
		g.unsupported(w, "unsupported set literal: nil")
		return
	}
	target, value := setTargetAndValue(a.Lit)
	loops, ok := g.openAssignmentLoops(w, target)
	if !ok {
		return
	}
	lhs, err := g.emitExpr(target)
	if err != nil {
		g.unsupported(w, "unsupported set literal: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	w.linef("%s = %s;", lhs, value)
	g.closeAssignmentLoops(w, loops)
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

func (g *Generator) emitAssign(w *cppWriter, a *goivy.LogicAssignAction) {
	loops, ok := g.openAssignmentLoops(w, a.LHS)
	if !ok {
		return
	}
	lhs, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	rhs, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		g.closeAssignmentLoops(w, loops)
		return
	}
	w.linef("%s = %s;", lhs, rhs)
	g.closeAssignmentLoops(w, loops)
}

func (g *Generator) closeAssignmentLoops(w *cppWriter, loops int) {
	for i := 0; i < loops; i++ {
		w.close("")
	}
}

func (g *Generator) openAssignmentLoops(w *cppWriter, lhs goivy.Expr) (int, bool) {
	vars := goivy.FreeVariablesList(lhs)
	opened := 0
	for _, v := range vars {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported assignment over free variable %s:%s", varName(v.Name), err.Error())
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return 0, false
		}
		w.open(header)
		opened++
	}
	return opened, true
}

func (g *Generator) emitAssertLike(w *cppWriter, fn string, f goivy.Expr, label string) {
	expr, err := g.emitExpr(f)
	if err != nil {
		g.unsupported(w, "unsupported assertion expression: %s", err.Error())
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef(`%s(%s, "%s");`, fn, expr, escapeString(label))
}

func (g *Generator) emitIf(w *cppWriter, a *goivy.LogicIfAction) {
	if some, ok := a.Cond.(*goivy.SomeCondition); ok {
		g.emitIfSome(w, a, some)
		return
	}
	cond, err := g.emitExpr(a.GetCond())
	if err != nil {
		g.unsupported(w, "unsupported if condition: %s", err.Error())
		return
	}
	w.open("if (" + cond + ") {")
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.close(" else {")
		g.emitAction(w, elseAct)
		w.close("")
		return
	}
	w.close("")
}

func (g *Generator) emitIfSome(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) {
	if some.Kind != "some" {
		g.unsupported(w, "unsupported if %s condition: minimizing/maximizing some is not supported yet", some.Kind)
		return
	}
	if g.emitIfSomeVariantDowncast(w, a, some) {
		return
	}
	if g.emitIfSomeExtensional(w, a, some) {
		return
	}
	found := g.nextTemp("__ivy_some")
	w.linef("bool %s = false;", found)
	opened := 0
	for _, p := range some.Params {
		header, err := g.loopHeaderForSort(p.CSort, varName(p.Name))
		if err != nil {
			g.unsupported(w, "unsupported some parameter %s:%s", varName(p.Name), err.Error())
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return
		}
		w.open(header)
		opened++
	}
	cond, err := g.emitExpr(some.Fmla)
	if err != nil {
		g.unsupported(w, "unsupported some condition: %s", err.Error())
		for i := 0; i < opened; i++ {
			w.close("")
		}
		return
	}
	w.open(fmt.Sprintf("if (!%s && (%s)) {", found, cond))
	w.linef("%s = true;", found)
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	w.close("")
	for i := 0; i < opened; i++ {
		w.close("")
	}
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if (!%s) {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
}

func (g *Generator) emitIfSomeExtensional(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {
	if len(some.Params) != 1 {
		return false
	}
	p := some.Params[0]
	bound := &goivy.LogicVariable{Name: p.Name, VSort: p.CSort}
	app, argIndex, ok := g.findPositiveExtensionalRelationBound(some.Fmla, bound)
	if !ok {
		return false
	}
	relName := goivy.ExprName(app.Func)
	fs, ok := app.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	if st.Kind != cppStorageHashThunk {
		return false
	}
	rel := varName(relName)
	found := g.nextTemp("__ivy_some")
	w.linef("bool %s = false;", found)
	w.open(fmt.Sprintf("for (auto it = %s.memo.begin(), en = %s.memo.end(); it != en; ++it) {", rel, rel))
	w.line("if (!it->second) continue;")
	if len(app.Terms) == 1 {
		w.linef("%s %s = it->first;", g.cppType(p.CSort), varName(p.Name))
	} else {
		w.linef("%s %s = it->first.arg%d;", g.cppType(p.CSort), varName(p.Name), argIndex)
	}
	cond, err := g.emitExpr(some.Fmla)
	if err != nil {
		g.unsupported(w, "unsupported some condition: %s", err.Error())
		w.close("")
		return true
	}
	w.open(fmt.Sprintf("if (!%s && (%s)) {", found, cond))
	w.linef("%s = true;", found)
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	w.close("")
	w.close("")
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if (!%s) {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
	return true
}

func (g *Generator) emitIfSomeVariantDowncast(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {
	if len(some.Params) != 1 {
		return false
	}
	v := some.Params[0]
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
	w.open(fmt.Sprintf("if (%s.__tag == %d) {", lhs, idx))
	w.linef("%s %s = %s.%s;", g.cppType(v.CSort), varName(v.Name), lhs, variantPayloadField(v.CSort))
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		g.emitAction(w, thenAct)
	}
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.close(" else {")
		g.emitAction(w, elseAct)
		w.close("")
		return true
	}
	w.close("")
	return true
}

func (g *Generator) emitWhile(w *cppWriter, a *goivy.LogicWhileAction) {
	cond, err := g.emitExpr(a.Cond)
	if err != nil {
		g.unsupported(w, "unsupported while condition: %s", err.Error())
		return
	}
	w.open("while (" + cond + ") {")
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
}

func (g *Generator) emitChoice(w *cppWriter, a *goivy.LogicChoiceAction) {
	if len(a.Branches) == 0 {
		return
	}
	w.open(fmt.Sprintf("switch (___ivy_choose(%d, \"___branch\", %d)) {", len(a.Branches), a.UniqueID))
	for i, b := range a.Branches {
		if i == len(a.Branches)-1 {
			w.line("default:")
		} else {
			w.linef("case %d:", i)
		}
		w.indent++
		if act, ok := b.(goivy.Action); ok {
			g.emitAction(w, act)
		}
		w.line("break;")
		w.indent--
	}
	w.close("")
}

func (g *Generator) emitCall(w *cppWriter, a *goivy.LogicCallAction) {
	name := a.CalleeName()
	fn, err := funName(name)
	if err != nil {
		g.unsupported(w, "unsupported call: %s", err.Error())
		return
	}
	var args []string
	if app, ok := a.Callee.(*goivy.Apply); ok {
		for _, t := range app.Terms {
			s, err := g.emitExpr(t)
			if err != nil {
				g.unsupported(w, "unsupported call arg: %s", err.Error())
				return
			}
			args = append(args, s)
		}
	}
	if g.Mod != nil && g.Mod.Actions != nil && len(a.ActualReturns) == 1 {
		if callee, ok := g.Mod.Actions.Get2(name); ok && len(callee.GetFormalReturns()) == 1 {
			ret, err := g.emitExpr(a.ActualReturns[0])
			if err != nil {
				g.unsupported(w, "unsupported call return: %s", err.Error())
				return
			}
			stacked := g.emitCallStackPush(w, a)
			w.linef("%s = %s(%s);", ret, fn, strings.Join(args, ", "))
			g.emitCallStackPop(w, stacked)
			return
		}
	}
	for _, r := range a.ActualReturns {
		s, err := g.emitExpr(r)
		if err != nil {
			g.unsupported(w, "unsupported call return: %s", err.Error())
			return
		}
		args = append(args, s)
	}
	stacked := g.emitCallStackPush(w, a)
	w.linef("%s(%s);", fn, strings.Join(args, ", "))
	g.emitCallStackPop(w, stacked)
}

func (g *Generator) emitCallStackPush(w *cppWriter, a *goivy.LogicCallAction) bool {
	if g == nil || a == nil || !g.runtimeUsesGenerator() {
		return false
	}
	w.linef("___ivy_stack.push_back(%d);", a.UniqueID)
	return true
}

func (g *Generator) emitCallStackPop(w *cppWriter, stacked bool) {
	if stacked {
		w.line("___ivy_stack.pop_back();")
	}
}

func (g *Generator) emitLocal(w *cppWriter, a *goivy.LogicLocalAction) {
	w.open("{")
	for _, local := range a.Locals {
		name := goivy.ExprName(local)
		w.linef("%s %s = %s;", g.cppType(local.NodeSort()), varName(name), g.cppZeroValue(local.NodeSort()))
	}
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
}

func (g *Generator) emitLet(w *cppWriter, a *goivy.LogicLetAction) {
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
	w.open("{")
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	} else {
		g.unsupported(w, "unsupported let body %T", a.Body)
	}
	w.close("")
	g.exprAliases = prev
}

func (g *Generator) emitBindOlds(w *cppWriter, a *goivy.LogicBindOldsAction) {
	if inner, ok := a.Inner.(goivy.Action); ok {
		g.emitAction(w, inner)
		return
	}
	g.unsupported(w, "unsupported bindolds inner %T", a.Inner)
}

func (g *Generator) emitAssignField(w *cppWriter, a *goivy.LogicAssignFieldAction) {
	lhs, err := g.emitFieldRef(a.Obj, a.Field)
	if err != nil {
		g.unsupported(w, "unsupported field assignment lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitExpr(a.Value)
	if err != nil {
		g.unsupported(w, "unsupported field assignment rhs: %s", err.Error())
		return
	}
	w.linef("%s = %s;", lhs, rhs)
}

func (g *Generator) emitNullField(w *cppWriter, a *goivy.LogicNullFieldAction) {
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
	w.linef("%s = %s;", lhs, g.cppZeroValue(fieldSort))
}

func (g *Generator) emitCopyField(w *cppWriter, a *goivy.LogicCopyFieldAction) {
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
	w.linef("%s = %s;", lhs, rhs)
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
	return objCode + "." + varName(memName(fieldName)), nil
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

func (g *Generator) emitNativeAction(w *cppWriter, a *goivy.LogicNativeAction) {
	code, ok := a.Code.(*goivy.NativeCode)
	if !ok {
		g.unsupported(w, "unsupported native action code %T", a.Code)
		return
	}
	rendered, err := g.renderNativeTemplate(code.Code, a.Params)
	if err != nil {
		g.unsupported(w, "unsupported native action: %s", err.Error())
		return
	}
	emitNativeLines(w, rendered)
}

func (g *Generator) emitDebug(w *cppWriter, a *goivy.LogicDebugAction) {
	event := debugEventName(a.DebugExpr)
	w.line(`std::cout << "{" << std::endl;`)
	w.linef(`std::cout << "    \"event\" : \"%s\"," << std::endl;`, escapeString(event))
	for i, e := range a.WithExprs {
		expr, err := g.emitExpr(e)
		if err != nil {
			g.unsupported(w, "unsupported debug expression: %s", err.Error())
			continue
		}
		w.linef(`std::cout << "    \"value%d\" : " << (%s) << "," << std::endl;`, i, expr)
	}
	w.line(`std::cout << "}" << std::endl;`)
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

func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func escapeComment(s string) string {
	s = strings.ReplaceAll(s, "*/", "* /")
	return s
}
