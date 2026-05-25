package ivy2go

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// emitAction is the top-level statement dispatcher. Mirrors
// ivy2cpp/action.go emitAction case-by-case; each case substitutes Go
// syntax for C++.
//
// M4 ships the common cases (Sequence, AssertLike, simple If, simple
// Assign with no quantified LHS, While, Local, Return, simple Call,
// IgnoreAction). Advanced cases (if-some, quantified assign, choice,
// native action, debug, bind-olds, assign-field, copy-field) are
// deferred to later milestones; their dispatch entries record a
// `g.unsupported` error so any test that exercises the path fails
// loudly rather than emitting silently wrong Go.
func (g *Generator) emitAction(w *goWriter, act goivy.Action) {
	if act == nil {
		return
	}
	switch a := act.(type) {
	case *goivy.LogicSequence:
		w.open("{")
		for _, child := range a.Elems {
			if childAct, ok := child.(goivy.Action); ok {
				g.emitAction(w, childAct)
			} else {
				g.unsupported(w, "unsupported sequence child %T: %s", child, fmt.Sprint(child))
			}
		}
		w.close("")
	case *goivy.LogicAssignAction:
		g.emitAssign(w, a)
	case *goivy.LogicHavocAction:
		g.emitHavoc(w, a)
	case *goivy.LogicSetAction:
		g.unsupported(w, "set-action emission deferred (M5)")
	case *goivy.LogicAssertAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicRequiresAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicEnsuresAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicSubgoalAction:
		g.emitAssertLike(w, "ivyAssert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicAssumeAction:
		g.emitAssertLike(w, "ivyAssume", a.Formula, linenoStr(a.GetLineno()))
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
		g.unsupported(w, "let-action emission deferred (M5)")
	case *goivy.LogicBindOldsAction:
		g.unsupported(w, "bind-olds emission deferred (M5)")
	case *goivy.LogicNativeAction:
		g.unsupported(w, "native-action emission deferred (M10)")
	case *goivy.LogicDebugAction:
		// Python's emit_debug is a no-op; mirror that.
		_ = a
	case *goivy.LogicCrashAction:
		// Per ivy2cpp/action.go comment: CrashAction is lowered upstream;
		// emit nothing if it survives.
		_ = a
	case *goivy.ReturnAction:
		g.emitReturn(w)
	case *goivy.IgnoreAction:
		return
	case *goivy.LogicAssignFieldAction:
		g.unsupported(w, "assign-field emission deferred (M8)")
	case *goivy.LogicNullFieldAction:
		g.unsupported(w, "null-field emission deferred (M8)")
	case *goivy.LogicCopyFieldAction:
		g.unsupported(w, "copy-field emission deferred (M8)")
	case *goivy.LogicThunkAction:
		g.unsupported(w, "thunk reached emit (expected desugared upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicInstantiateAction:
		g.unsupported(w, "instantiate reached emit (expected inlined upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicRanking:
		g.unsupported(w, "ranking reached emit (handled by analysis, not codegen): %s at %s",
			a.String(), a.GetLineno().String())
	default:
		g.unsupported(w, "unsupported action %T: %s", act, act.String())
	}
}

// emitReturn ports ivy2cpp/action.go emitReturn.
func (g *Generator) emitReturn(w *goWriter) {
	if g != nil && len(g.currentReturns) == 1 {
		w.linef("return %s", goIdent(g.currentReturns[0].Name))
		return
	}
	w.line("return")
}

// emitHavoc ports ivy2cpp/action.go emitHavoc: a havoc that survives
// upstream lowering is a bug; report and continue rather than crash.
func (g *Generator) emitHavoc(w *goWriter, a *goivy.LogicHavocAction) {
	target := "<nil>"
	if a.Target != nil {
		target = a.Target.String()
	}
	g.unsupported(w, "havoc reached emit (expected lowered upstream): %s at %s", target, a.GetLineno().String())
}

// emitAssertLike ports ivy2cpp/action.go emitAssertLike.
// Emits `ivyAssert(expr, "filename:lineno")` or `ivyAssume(...)`.
func (g *Generator) emitAssertLike(w *goWriter, fn string, f goivy.Expr, label string) {
	expr, err := g.emitExpr(f)
	if err != nil {
		g.unsupported(w, "unsupported assertion expression: %s", err.Error())
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef(`%s(%s, %q)`, fn, expr, label)
}

// linenoStr ports ivy2cpp/action.go linenoStr.
func linenoStr(loc goivy.Location) string {
	return strings.TrimSuffix(loc.String(), ": ")
}

// emitIf ports ivy2cpp/action.go emitIf, simplified for M4.
// `if some` (existential conditions) is deferred to M5.
func (g *Generator) emitIf(w *goWriter, a *goivy.LogicIfAction) {
	if _, ok := a.Cond.(*goivy.SomeCondition); ok {
		g.unsupported(w, "if-some emission deferred (M5)")
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
		if elseAct, ok := a.ElseBody.(goivy.Action); ok {
			w.close(" else {")
			g.emitAction(w, elseAct)
			w.close("")
			return
		}
	}
	w.close("")
}

// emitWhile ports ivy2cpp/action.go emitWhile, simplified for M4.
// Go's `for` doubles as `while`; we drop invariants (they're handled by
// upstream verification, not codegen).
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

// emitChoice ports ivy2cpp/action.go emitChoice: pick one of N branches
// uniformly at random, then execute it. In Go we use ivyChoose(N) and
// a switch.
func (g *Generator) emitChoice(w *goWriter, a *goivy.LogicChoiceAction) {
	if len(a.Branches) == 0 {
		return
	}
	w.linef("switch ivyChoose(%d) {", len(a.Branches))
	for i, b := range a.Branches {
		if i == len(a.Branches)-1 {
			w.line("default:")
		} else {
			w.linef("case %d:", i)
		}
		if act, ok := b.(goivy.Action); ok {
			g.emitAction(w, act)
		} else {
			g.unsupported(w, "unsupported choice branch %T", b)
		}
	}
	w.line("}")
}

// emitCall ports a small subset of ivy2cpp/action.go emitCall. M4
// handles only `call f(args...)` and `call ret := f(args...)` against
// a named action method on *State; the call-stack push/pop, callback
// thunk paths, and import-callback routing are deferred to M5/M7.
func (g *Generator) emitCall(w *goWriter, a *goivy.LogicCallAction) {
	apply, ok := a.Callee.(*goivy.Apply)
	if !ok {
		name := goivy.ExprName(a.Callee)
		if name == "" {
			g.unsupported(w, "unsupported call callee %T", a.Callee)
			return
		}
		if len(a.ActualReturns) == 0 {
			w.linef("s.%s()", goExportedName(name))
		} else {
			rets := make([]string, len(a.ActualReturns))
			for i, r := range a.ActualReturns {
				code, err := g.emitExpr(r)
				if err != nil {
					g.unsupported(w, "unsupported call return: %s", err.Error())
					return
				}
				rets[i] = code
			}
			w.linef("%s = s.%s()", strings.Join(rets, ", "), goExportedName(name))
		}
		return
	}
	name := goivy.ExprName(apply.Func)
	args := make([]string, len(apply.Terms))
	for i, t := range apply.Terms {
		code, err := g.emitExpr(t)
		if err != nil {
			g.unsupported(w, "unsupported call arg: %s", err.Error())
			return
		}
		args[i] = code
	}
	if len(a.ActualReturns) == 0 {
		w.linef("s.%s(%s)", goExportedName(name), strings.Join(args, ", "))
		return
	}
	rets := make([]string, len(a.ActualReturns))
	for i, r := range a.ActualReturns {
		code, err := g.emitExpr(r)
		if err != nil {
			g.unsupported(w, "unsupported call return: %s", err.Error())
			return
		}
		rets[i] = code
	}
	w.linef("%s = s.%s(%s)", strings.Join(rets, ", "), goExportedName(name), strings.Join(args, ", "))
}

// emitLocal ports ivy2cpp/action.go emitLocal, simplified: each local
// becomes a Go `var name T` declaration, then the body runs.
func (g *Generator) emitLocal(w *goWriter, a *goivy.LogicLocalAction) {
	w.open("{")
	for _, loc := range a.Locals {
		c, ok := loc.(*goivy.Const)
		if !ok {
			g.unsupported(w, "unsupported local declaration %T", loc)
			continue
		}
		w.linef("var %s %s", goIdent(c.Name), g.goType(c.CSort))
		w.linef("_ = %s", goIdent(c.Name)) // silence "declared and not used"
	}
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
}

// unsupported records a generator-time error and emits a marker
// comment so the failure is visible in the source. Mirrors
// ivy2cpp/generator.go unsupported.
func (g *Generator) unsupported(w *goWriter, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	g.errs = append(g.errs, fmt.Errorf("%s", msg))
	if w != nil {
		w.linef("// unsupported: %s", msg)
	}
}

// emitActionMethods walks the module's Actions map and emits one Go
// method on *State per action. Mirrors ivy2cpp/generator.go's
// per-action emission, simplified for M4: parameters become method
// params, return values become named return values, and the action
// body is the method body.
func (g *Generator) emitActionMethods(w *goWriter) {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return
	}
	// Stable iteration order so emission is deterministic.
	names := make([]string, 0)
	for name := range g.Mod.Actions.All() {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		act, _ := g.Mod.Actions.Get2(name)
		if act == nil {
			continue
		}
		g.emitActionMethod(w, name, act)
	}
}

// emitActionMethod renders a single Go method on *State for an Ivy
// action.
func (g *Generator) emitActionMethod(w *goWriter, name string, act goivy.Action) {
	methodName := goExportedName(name)
	params := act.GetFormalParams()
	returns := act.GetFormalReturns()

	// Build parameter list.
	paramParts := make([]string, 0, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		paramParts = append(paramParts, fmt.Sprintf("%s %s", goIdent(p.Name), g.goType(p.CSort)))
	}

	// Build return list. Named returns let `return` (without args)
	// pick them up automatically — useful when the action body relies
	// on havoc-initialised returns.
	returnParts := make([]string, 0, len(returns))
	for _, r := range returns {
		if r == nil {
			continue
		}
		returnParts = append(returnParts, fmt.Sprintf("%s %s", goIdent(r.Name), g.goType(r.CSort)))
	}

	header := fmt.Sprintf("func (s *%s) %s(%s)", g.StateTypeName, methodName, strings.Join(paramParts, ", "))
	switch len(returnParts) {
	case 0:
	case 1:
		header += " (" + returnParts[0] + ")"
	default:
		header += " (" + strings.Join(returnParts, ", ") + ")"
	}
	w.open(header + " {")

	// Track returns so emitReturn can use the right names.
	prev := g.currentReturns
	g.currentReturns = returns
	g.emitAction(w, act)
	g.currentReturns = prev

	w.close("")
	w.blank()
}
