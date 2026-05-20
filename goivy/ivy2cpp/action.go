package ivy2cpp

import (
	"fmt"
	"strconv"
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
		g.emitNativeAction(w, a)
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
		w.linef("%s %s = %s;", cppType(local.NodeSort()), varName(name), cppZeroValue(local.NodeSort()))
	}
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
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

func (g *Generator) renderNativeTemplate(code string, params []goivy.Expr) (string, error) {
	fields := strings.Split(code, "`")
	for i := 1; i < len(fields); i += 2 {
		idx, err := strconv.Atoi(fields[i])
		if err != nil {
			return "", fmt.Errorf("bad native antiquote index %q", fields[i])
		}
		if idx < 0 || idx >= len(params) {
			return "", fmt.Errorf("native antiquote index %d out of range", idx)
		}
		prev := fields[i-1]
		var repl string
		switch {
		case strings.HasSuffix(prev, "%"):
			repl, err = g.nativeTypeOf(params[idx])
		case strings.HasSuffix(prev, `"`):
			repl, err = g.nativeZ3Name(params[idx])
		default:
			repl, err = g.nativeReference(params[idx])
		}
		if err != nil {
			return "", err
		}
		fields[i] = repl
	}
	for i := 0; i < len(fields); i += 2 {
		if strings.HasSuffix(fields[i], "%") {
			fields[i] = strings.TrimSuffix(fields[i], "%")
		}
	}
	return strings.Join(fields, ""), nil
}

func (g *Generator) nativeReference(arg goivy.Expr) (string, error) {
	switch a := arg.(type) {
	case *goivy.Const:
		if s, ok := g.sortByName(a.Name); ok {
			return cppType(s), nil
		}
		return varName(a.Name), nil
	case *goivy.LogicVariable:
		return varName(a.Name), nil
	case *goivy.Apply:
		return g.emitExpr(a)
	case *goivy.UninterpretedSort:
		return cppType(a), nil
	case *goivy.LogicEnumeratedSort:
		return cppType(a), nil
	case *goivy.RangeSort:
		return cppType(a), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		name := rn.Relname()
		if s, ok := g.sortByName(name); ok {
			return cppType(s), nil
		}
		res := varName(name)
		for _, child := range arg.Args() {
			expr, ok := child.(goivy.Expr)
			if !ok {
				return "", fmt.Errorf("native reference %s has non-expression argument %T", name, child)
			}
			idx, err := g.nativeReference(expr)
			if err != nil {
				return "", err
			}
			res += "[" + idx + "]"
		}
		return res, nil
	}
	return g.emitExpr(arg)
}

func (g *Generator) nativeTypeOf(arg goivy.Expr) (string, error) {
	if rn, ok := arg.(interface{ Relname() string }); ok {
		name := rn.Relname()
		if g.Mod != nil && g.Mod.Actions != nil {
			if _, ok := g.Mod.Actions.Get2(name); ok {
				return "thunk__" + varName(name), nil
			}
		}
	}
	return cppType(arg.NodeSort()), nil
}

func (g *Generator) nativeZ3Name(arg goivy.Expr) (string, error) {
	if v, ok := arg.(*goivy.LogicVariable); ok {
		return sortName(v.VSort), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		return rn.Relname(), nil
	}
	if c, ok := arg.(*goivy.Const); ok {
		return c.Name, nil
	}
	return "", fmt.Errorf("cannot emit native z3 name for %T", arg)
}

func (g *Generator) sortByName(name string) (goivy.Sort, bool) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil, false
	}
	return g.Mod.Sig.Sorts.Get2(name)
}

func emitNativeLines(w *cppWriter, code string) {
	code = strings.TrimRight(code, " \t\r\n")
	lines := strings.Split(code, "\n")
	base := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := nativeIndent(line)
		if base == -1 || indent < base {
			base = indent
		}
	}
	if base < 0 {
		return
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			w.line("")
			continue
		}
		indent := nativeIndent(line) - base
		if indent < 0 {
			indent = 0
		}
		w.line(strings.Repeat(" ", indent) + strings.TrimSpace(line))
	}
}

func nativeIndent(line string) int {
	indent := 0
	for _, r := range line {
		switch r {
		case ' ':
			indent++
		case '\t':
			indent = ((indent + 8) / 8) * 8
		default:
			return indent
		}
	}
	return indent
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
