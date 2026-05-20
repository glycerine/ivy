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
	case *goivy.LogicCallAction:
		g.emitCall(w, a)
	case *goivy.LogicLocalAction:
		g.emitLocal(w, a)
	case *goivy.LogicNativeAction:
		w.line("/* native action omitted by ivy2cpp v1 */")
	case *goivy.LogicCrashAction:
		w.line("std::abort();")
	default:
		w.linef("/* unsupported action %T: %s */", act, escapeComment(act.String()))
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
		return
	}
	rhs, err := g.emitExpr(a.RHS)
	if err != nil {
		w.linef("/* unsupported assignment rhs: %s */", escapeComment(err.Error()))
		return
	}
	w.linef("%s = %s;", lhs, rhs)
	for i := 0; i < loops; i++ {
		w.close("")
	}
}

func (g *Generator) openAssignmentLoops(w *cppWriter, lhs goivy.Expr) (int, bool) {
	vars := goivy.FreeVariablesList(lhs)
	opened := 0
	for _, v := range vars {
		vals, ok := finiteValues(v.VSort)
		if !ok {
			w.linef("/* unsupported assignment over free variable %s:%s */", varName(v.Name), sortName(v.VSort))
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return 0, false
		}
		w.open(fmt.Sprintf("for (%s %s : {%s}) {", cppType(v.VSort), varName(v.Name), strings.Join(vals, ", ")))
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
	fn, err := funName(a.CalleeName())
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
		w.linef("%s %s = %s;", cppType(local.NodeSort()), varName(name), cppZeroValue(local.NodeSort()))
	}
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
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
