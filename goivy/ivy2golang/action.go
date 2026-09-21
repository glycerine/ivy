package ivy2golang

import (
	"fmt"
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
		g.emitAssume(w, a.Formula)
	case *goivy.LogicIfAction:
		g.emitIf(w, a)
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
		if inner, ok := a.Inner.(goivy.Action); ok {
			g.emitAction(w, inner)
		}
	case *goivy.LogicDebugAction:
		w.line("_ = fmt.Sprintf")
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
		g.unsupported(w, "native actions are not yet supported in ivy2golang")
	case *goivy.LogicHavocAction:
		g.unsupported(w, "havoc reached emit; expected upstream lowering: %s", a.String())
	default:
		g.unsupported(w, "unsupported action %T: %s", act, act.String())
	}
}

func (g *Generator) emitAssign(w *goWriter, a *goivy.LogicAssignAction) {
	vs := goivy.VariablesAstList(a.LHS)
	if len(vs) == 0 {
		g.emitAssignOne(w, a.LHS, a.RHS)
		return
	}
	g.pushScope()
	for _, v := range vs {
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			g.unsupported(w, "unsupported assignment over free variable %s:%s", goName(v.Name), sortName(v.VSort))
			g.popScope()
			return
		}
		g.addLocal(v.Name)
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")))
	}
	g.emitAssignOne(w, a.LHS, a.RHS)
	for range vs {
		w.close("")
	}
	g.popScope()
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
	if g.Config.Trace {
		w.linef("fmt.Fprintf(__ivy_out, %q, %s)", "  write("+lhsCode+",%v)\n", rhs)
	}
	w.linef("%s = %s", lhsCode, rhs)
}

func (g *Generator) emitSet(w *goWriter, a *goivy.LogicSetAction) {
	if a.Lit == nil {
		g.unsupported(w, "unsupported set literal: nil")
		return
	}
	target, value := setTargetAndValue(a.Lit)
	g.emitAssignOne(w, target, goivy.NewConst(value, goivy.Boolean))
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
	expr, err := g.emitExpr(goivy.CloseFormula(f))
	if err != nil {
		g.unsupported(w, "unsupported assertion expression: %s", err.Error())
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef("%s(%s, %q)", fn, expr, label)
}

func (g *Generator) emitAssume(w *goWriter, f goivy.Expr) {
	expr, err := g.emitExpr(goivy.CloseFormula(f))
	if err != nil {
		g.unsupported(w, "unsupported assumption expression: %s", err.Error())
		return
	}
	w.open("if !ivyAssume(" + expr + ", \"assume\") {")
	w.line("return")
	w.close("")
}

func (g *Generator) emitIf(w *goWriter, a *goivy.LogicIfAction) {
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

func (g *Generator) emitChoice(w *goWriter, a *goivy.LogicChoiceAction) {
	if len(a.Branches) == 0 {
		return
	}
	w.linef("switch ivy.___ivy_choose(%d, %q, %d) {", len(a.Branches), "choice", a.UniqueID)
	w.indent++
	for i, branch := range a.Branches {
		w.linef("case %d:", i)
		w.indent++
		if act, ok := branch.(goivy.Action); ok {
			g.emitAction(w, act)
		}
		w.indent--
	}
	w.indent--
	w.line("}")
}

func (g *Generator) emitCall(w *goWriter, a *goivy.LogicCallAction) {
	name := a.CalleeName()
	args := callArgs(a.Callee)
	argCodes := make([]string, len(args))
	for i, arg := range args {
		code, err := g.emitExpr(arg)
		if err != nil {
			g.unsupported(w, "unsupported call argument: %s", err.Error())
			return
		}
		argCodes[i] = code
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
	g.pushScope()
	for _, local := range a.Locals {
		name := goivy.ExprName(local)
		if name == "" {
			g.unsupported(w, "unsupported local declaration %T: %s", local, local.String())
			continue
		}
		g.addLocal(name)
		sort := local.NodeSort()
		init := g.goZeroValue(sort)
		if expr, err := g.goRandomValueExpr(sort, name, 0); err == nil {
			init = expr
		}
		w.linef("%s := %s", goName(name), init)
	}
	if body, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, body)
	}
	g.popScope()
}

func (g *Generator) emitLet(w *goWriter, a *goivy.LogicLetAction) {
	if body, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, body)
		return
	}
	g.unsupported(w, "unsupported let action body %T", a.Body)
}
