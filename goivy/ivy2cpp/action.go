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
	case *goivy.LogicAssignFieldAction:
		g.emitAssignField(w, a)
	case *goivy.LogicNullFieldAction:
		g.emitNullField(w, a)
	case *goivy.LogicCopyFieldAction:
		g.emitCopyField(w, a)
	default:
		w.linef("/* unsupported action %T: %s */", act, escapeComment(act.String()))
	}
}

func (g *Generator) emitHavoc(w *cppWriter, a *goivy.LogicHavocAction) {
	if a.Target == nil {
		w.line("/* unsupported havoc target: nil */")
		return
	}
	loops, ok := g.openAssignmentLoops(w, a.Target)
	if !ok {
		return
	}
	lhs, err := g.emitExpr(a.Target)
	if err != nil {
		w.linef("/* unsupported havoc target: %s */", escapeComment(err.Error()))
		g.closeAssignmentLoops(w, loops)
		return
	}
	w.linef("%s = %s;", lhs, g.cppZeroValue(a.Target.NodeSort()))
	g.closeAssignmentLoops(w, loops)
}

func (g *Generator) emitSet(w *cppWriter, a *goivy.LogicSetAction) {
	if a.Lit == nil {
		w.line("/* unsupported set literal: nil */")
		return
	}
	target, value := setTargetAndValue(a.Lit)
	loops, ok := g.openAssignmentLoops(w, target)
	if !ok {
		return
	}
	lhs, err := g.emitExpr(target)
	if err != nil {
		w.linef("/* unsupported set literal: %s */", escapeComment(err.Error()))
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
		w.linef("/* unsupported assignment lhs: %s */", escapeComment(err.Error()))
		g.closeAssignmentLoops(w, loops)
		return
	}
	rhs, err := g.emitExpr(a.RHS)
	if err != nil {
		w.linef("/* unsupported assignment rhs: %s */", escapeComment(err.Error()))
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
			w.linef("/* unsupported assignment over free variable %s:%s */", varName(v.Name), escapeComment(err.Error()))
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
		w.linef("/* unsupported assertion expression: %s */", escapeComment(err.Error()))
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef(`%s(%s, "%s");`, fn, expr, escapeString(label))
}

func (g *Generator) emitIf(w *cppWriter, a *goivy.LogicIfAction) {
	cond, err := g.emitExpr(a.GetCond())
	if err != nil {
		w.linef("/* unsupported if condition: %s */", escapeComment(err.Error()))
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

func (g *Generator) emitWhile(w *cppWriter, a *goivy.LogicWhileAction) {
	cond, err := g.emitExpr(a.Cond)
	if err != nil {
		w.linef("/* unsupported while condition: %s */", escapeComment(err.Error()))
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
		w.linef("/* unsupported call: %s */", escapeComment(err.Error()))
		return
	}
	var args []string
	if app, ok := a.Callee.(*goivy.Apply); ok {
		for _, t := range app.Terms {
			s, err := g.emitExpr(t)
			if err != nil {
				w.linef("/* unsupported call arg: %s */", escapeComment(err.Error()))
				return
			}
			args = append(args, s)
		}
	}
	if g.Mod != nil && g.Mod.Actions != nil && len(a.ActualReturns) == 1 {
		if callee, ok := g.Mod.Actions.Get2(name); ok && len(callee.GetFormalReturns()) == 1 {
			ret, err := g.emitExpr(a.ActualReturns[0])
			if err != nil {
				w.linef("/* unsupported call return: %s */", escapeComment(err.Error()))
				return
			}
			w.linef("%s = %s(%s);", ret, fn, strings.Join(args, ", "))
			return
		}
	}
	for _, r := range a.ActualReturns {
		s, err := g.emitExpr(r)
		if err != nil {
			w.linef("/* unsupported call return: %s */", escapeComment(err.Error()))
			return
		}
		args = append(args, s)
	}
	w.linef("%s(%s);", fn, strings.Join(args, ", "))
}

func (g *Generator) emitLocal(w *cppWriter, a *goivy.LogicLocalAction) {
	w.open("{")
	for _, local := range a.Locals {
		name := goivy.ExprName(local)
		w.linef("%s %s = %s;", cppType(local.NodeSort()), varName(name), g.cppZeroValue(local.NodeSort()))
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
			w.linef("/* unsupported let binding %T: %s */", binding, escapeComment(binding.String()))
			continue
		}
		name := goivy.ExprName(children[0])
		if name == "" {
			w.linef("/* unsupported let binding lhs %T: %s */", children[0], escapeComment(children[0].String()))
			continue
		}
		next[name] = children[1]
	}
	g.exprAliases = next
	w.open("{")
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
	g.exprAliases = prev
}

func (g *Generator) emitBindOlds(w *cppWriter, a *goivy.LogicBindOldsAction) {
	if inner, ok := a.Inner.(goivy.Action); ok {
		g.emitAction(w, inner)
		return
	}
	w.linef("/* unsupported bindolds inner %T */", a.Inner)
}

func (g *Generator) emitAssignField(w *cppWriter, a *goivy.LogicAssignFieldAction) {
	lhs, err := g.emitFieldRef(a.Obj, a.Field)
	if err != nil {
		w.linef("/* unsupported field assignment lhs: %s */", escapeComment(err.Error()))
		return
	}
	rhs, err := g.emitExpr(a.Value)
	if err != nil {
		w.linef("/* unsupported field assignment rhs: %s */", escapeComment(err.Error()))
		return
	}
	w.linef("%s = %s;", lhs, rhs)
}

func (g *Generator) emitNullField(w *cppWriter, a *goivy.LogicNullFieldAction) {
	lhs, err := g.emitFieldRef(a.Obj, a.Field)
	if err != nil {
		w.linef("/* unsupported null field lhs: %s */", escapeComment(err.Error()))
		return
	}
	fieldSort, err := fieldRangeSort(a.Field)
	if err != nil {
		w.linef("/* unsupported null field sort: %s */", escapeComment(err.Error()))
		return
	}
	w.linef("%s = %s;", lhs, g.cppZeroValue(fieldSort))
}

func (g *Generator) emitCopyField(w *cppWriter, a *goivy.LogicCopyFieldAction) {
	lhs, err := g.emitFieldRef(a.Dst, a.Field)
	if err != nil {
		w.linef("/* unsupported copy field lhs: %s */", escapeComment(err.Error()))
		return
	}
	rhs, err := g.emitFieldRef(a.Src, a.SrcField)
	if err != nil {
		w.linef("/* unsupported copy field rhs: %s */", escapeComment(err.Error()))
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
		w.linef("/* unsupported native action code %T */", a.Code)
		return
	}
	rendered, err := g.renderNativeTemplate(code.Code, a.Params)
	if err != nil {
		w.linef("/* unsupported native action: %s */", escapeComment(err.Error()))
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
			w.linef("/* unsupported debug expression: %s */", escapeComment(err.Error()))
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
