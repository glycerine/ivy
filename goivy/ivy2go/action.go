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
		g.emitLet(w, a)
	case *goivy.LogicBindOldsAction:
		g.emitBindOlds(w, a)
	case *goivy.LogicNativeAction:
		g.emitNativeAction(w, a)
	case *goivy.LogicDebugAction:
		g.emitDebug(w, a)
	case *goivy.LogicCrashAction:
		// Per ivy2cpp/action.go comment: CrashAction is lowered upstream;
		// emit nothing if it survives.
		_ = a
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
		w.linef("return %s", g.actionParamName(g.currentReturns[0]))
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

// emitSet ports ivy2cpp/action.go emitSet. Lowers a constraint-style
// `set` action: open one loop per free variable of the target, then
// assign the literal's value into the indexed LHS.
func (g *Generator) emitSet(w *goWriter, a *goivy.LogicSetAction) {
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
	w.linef("%s = %s", lhs, value)
	g.closeAssignmentLoops(w, loops)
}

// setTargetAndValue ports ivy2cpp/action.go setTargetAndValue.
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

// emitLet ports ivy2cpp/action.go emitLet. Maintains exprAliases so
// references in the body lower to the bound RHS.
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
	w.open("{")
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	} else {
		g.unsupported(w, "unsupported let body %T", a.Body)
	}
	w.close("")
	g.exprAliases = prev
}

// emitBindOlds ports ivy2cpp/action.go emitBindOlds — which itself
// mirrors Python's lack of an emit_bind_olds: BindOldsAction is
// supposed to be eliminated upstream by bind_olds_action during
// int_update. Reaching emit is a bug; surface it.
func (g *Generator) emitBindOlds(w *goWriter, a *goivy.LogicBindOldsAction) {
	inner := "<nil>"
	if a.Inner != nil {
		inner = fmt.Sprintf("%T", a.Inner)
	}
	g.unsupported(w, "bindolds reached emit (Python has no emit_bind_olds): inner=%s at %s",
		inner, a.GetLineno().String())
}

// emitAssignField ports ivy2cpp/action.go emitAssignField. Synthesises
// an apply(field, obj) LHS and delegates to emitAssign so the field
// store picks up the same variant-upcast, two-phase, and thunk
// handling as a normal assignment.
func (g *Generator) emitAssignField(w *goWriter, a *goivy.LogicAssignFieldAction) {
	if a == nil || a.Field == nil || a.Obj == nil {
		g.unsupported(w, "unsupported field assignment: nil components")
		return
	}
	lhs := goivy.NewApplyUnchecked(a.Field, a.Obj)
	synth := goivy.NewAssignAction(lhs, a.Value)
	g.emitAssign(w, synth)
}

// emitNullField ports ivy2cpp/action.go emitNullField. Sets a field
// to the zero value of its range sort.
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

// emitCopyField ports ivy2cpp/action.go emitCopyField. Copies one
// destructor field into another (dst.field = src.srcField).
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

// emitFieldRef ports ivy2cpp/action.go emitFieldRef. Emits `obj.<Field>`
// using the goExportedName convention destructor structs already use.
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
	return objCode + "." + goExportedName(memName(fieldName)), nil
}

// fieldRangeSort ports ivy2cpp/action.go fieldRangeSort.
func fieldRangeSort(field goivy.Expr) (goivy.Sort, error) {
	if field == nil {
		return nil, fmt.Errorf("nil field")
	}
	if fs, ok := field.NodeSort().(*goivy.LogicFunctionSort); ok {
		return fs.Range(), nil
	}
	return nil, fmt.Errorf("field %s does not have function sort", field.String())
}

// emitDebug ports ivy2cpp/action.go:983 emitDebug. Emits a JSON-style
// trace event with the labelled with-clause values inside.
func (g *Generator) emitDebug(w *goWriter, a *goivy.LogicDebugAction) {
	event := debugEventName(a.DebugExpr)
	g.Ctx.AddImport("actions", "fmt", "")
	w.line(`fmt.Fprintln(ivyTraceOut, "{")`)
	w.linef(`fmt.Fprintln(ivyTraceOut, "    \"event\" : \"%s\",")`, escapeString(event))
	for i, e := range a.WithExprs {
		name := ""
		if i < len(a.WithNames) {
			name = a.WithNames[i]
		}
		if name == "" {
			name = goivy.ExprName(e)
		}
		w.linef(`fmt.Fprintf(ivyTraceOut, "    \"%s\" : ")`, escapeString(name))
		g.emitPrintExpr(w, e)
		w.line(`fmt.Fprintln(ivyTraceOut, ",")`)
	}
	w.line(`fmt.Fprintln(ivyTraceOut, "}")`)
}

// emitPrintExpr ports ivy2cpp/action.go:1010 emitPrintExpr. For each
// free variable in expr, opens a loop over its sort and surrounds
// the value(s) with `[…]` brackets, comma-separated across iterations.
func (g *Generator) emitPrintExpr(w *goWriter, expr goivy.Expr) {
	vs := goivy.VariablesAstList(expr)
	openedHeaders := 0
	for _, v := range vs {
		header, _, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported debug print variable %s: %s", goIdent(v.Name), err.Error())
			for i := 0; i < openedHeaders; i++ {
				w.line("}")
			}
			return
		}
		w.line(`fmt.Fprint(ivyTraceOut, "[")`)
		w.line(header)
		w.linef(`if %s != 0 { fmt.Fprint(ivyTraceOut, ",") }`, goIdent(v.Name))
		openedHeaders++
	}
	value, err := g.emitExpr(expr)
	if err != nil {
		g.unsupported(w, "unsupported debug print expression: %s", err.Error())
		for i := 0; i < openedHeaders; i++ {
			w.line("}")
		}
		return
	}
	w.linef(`fmt.Fprint(ivyTraceOut, %s)`, value)
	for i := 0; i < openedHeaders; i++ {
		w.line("}")
		w.line(`fmt.Fprint(ivyTraceOut, "]")`)
	}
}

// debugEventName mirrors ivy2cpp/action.go:1043. Extracts the event
// name from a DebugExpr (typically a quoted-string Const).
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

// escapeString mirrors ivy2cpp/action.go:1055. Escapes backslashes
// and double-quotes for embedding inside a Go string literal that
// the emitted code will pass through fmt.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// escapeComment mirrors ivy2cpp/action.go:1061. Splits `*/` inside
// a comment body so it can't terminate a `/* … */` block.
func escapeComment(s string) string {
	s = strings.ReplaceAll(s, "*/", "* /")
	return s
}

// emitCallStackPush mirrors ivy2cpp/action.go:798. cpp emits a
// `___ivy_stack.push_back(<uid>)` for runtime call-stack tracking
// when the runtime uses the gen-target generator hook. ivy2go's
// runtime doesn't use stack-based callback tracing (the test target
// drives actions via per-action `actionGen_*` structs), so this is
// a no-op — kept for parallel call-site shape with ivy2cpp.
func (g *Generator) emitCallStackPush(w *goWriter, a *goivy.LogicCallAction) bool {
	_ = w
	_ = a
	return false
}

// emitCallStackPop mirrors ivy2cpp/action.go:806. No-op for the same
// reason as emitCallStackPush.
func (g *Generator) emitCallStackPop(w *goWriter, stacked bool) {
	_ = w
	_ = stacked
}

// someConditionLoopHeaders mirrors ivy2cpp/action.go:332. Picks
// per-parameter Go `for` loop headers for an `if some` statement.
// When the params are integer-typed and getAllBounds derives
// finite bounds from the some-condition formula, emit tightly-bounded
// loops; otherwise fall back to the generic loopHeaderForSort.
func (g *Generator) someConditionLoopHeaders(some *goivy.SomeCondition) ([]string, error) {
	headers := make([]string, len(some.Params))
	useBounds := false
	if len(some.Params) > 0 && goIsAnyIntegerType(g, some.Params[0].CSort) {
		vars := make([]*goivy.LogicVariable, 0, len(some.Params))
		subs := map[goivy.NodeKey]goivy.Expr{}
		ok := true
		for _, p := range some.Params {
			v, err := goivy.NewVariable("X"+p.Name, p.CSort)
			if err != nil {
				ok = false
				break
			}
			subs[goivy.Key(p)] = v
			vars = append(vars, v)
		}
		if ok {
			fmla, err := goivy.Substitute(some.Fmla, subs)
			if err == nil {
				if bounds, berr := g.getAllBounds(vars, fmla, true); berr == nil {
					useBounds = true
					for i, p := range some.Params {
						h, herr := g.loopHeaderForSortBounds(p.CSort, goIdent(p.Name), bounds[i][0], bounds[i][1])
						if herr != nil {
							useBounds = false
							break
						}
						headers[i] = h
					}
				}
			}
		}
	}
	if useBounds {
		return headers, nil
	}
	for i, p := range some.Params {
		h, _, err := g.loopHeaderForSort(p.CSort, goIdent(p.Name))
		if err != nil {
			return nil, err
		}
		headers[i] = h
	}
	return headers, nil
}

// firstParamIsIndex mirrors ivy2cpp/action.go:474. Used by
// `if some` optimization fast-paths to detect when the first
// some-condition parameter is also the result index.
func firstParamIsIndex(some *goivy.SomeCondition) bool {
	if some == nil || len(some.Params) == 0 || some.Index == nil {
		return false
	}
	pname := some.Params[0].Name
	switch x := some.Index.(type) {
	case *goivy.LogicVariable:
		return x.Name == pname
	case *goivy.Const:
		return x.Name == pname
	}
	return false
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

// emitIf ports ivy2cpp/action.go emitIf. Dispatches `if some` to
// emitIfSome; for everything else emits a Go `if cond { ... }` chain.
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
		if elseAct, ok := a.ElseBody.(goivy.Action); ok {
			w.close(" else {")
			g.emitAction(w, elseAct)
			w.close("")
			return
		}
	}
	w.close("")
}

// emitIfSome lowers `if some X. p(X) { THEN } else { ELSE }`. Mirrors
// ivy2cpp/action.go emitIfSome (the plain-Some path; SomeMin/SomeMax
// are deferred to a follow-up).
//
// Lowering shape:
//
//	{
//	    __found := false
//	    var X T
//	    for X := T(0); X < T(card); X++ {
//	        if !__found && p(X) {
//	            __found = true
//	            // THEN with X in scope (alias rewritten via emitAction)
//	        }
//	    }
//	    if !__found {
//	        // ELSE
//	    }
//	}
func (g *Generator) emitIfSome(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) {
	if some.Kind == "some_min" || some.Kind == "some_max" {
		g.emitIfSomeMinMax(w, a, some)
		return
	}
	if len(some.Params) == 0 {
		g.unsupported(w, "if-some with no params is malformed")
		return
	}
	w.open("{")
	w.line("__found := false")
	// Declare witnesses outside the loops so the ELSE block can
	// reference them (though typically only THEN does).
	witnessNames := make([]string, len(some.Params))
	for i, p := range some.Params {
		name := goIdent(p.Name)
		witnessNames[i] = name
		w.linef("var %s %s", name, g.goType(p.CSort))
		w.linef("_ = %s", name)
	}
	// Build the loop nest. For each param we open a loop, but the
	// loop variable name shadows the witness declared above; we
	// assign the witness inside the THEN body so it survives the
	// loop exit.
	headers := make([]string, len(some.Params))
	closers := make([]string, len(some.Params))
	for i, p := range some.Params {
		// loopHeaderForSort takes the name directly; build it from
		// the param name with a "Some_" prefix so the loop variable
		// is distinct from the witness (declared above with the
		// original name) and from any user-introduced symbol.
		loopName := "__some_" + goIdent(p.Name)
		h, c, herr := g.loopHeaderForSort(p.CSort, loopName)
		if herr != nil {
			g.unsupported(w, "if-some bounds: %s", herr.Error())
			return
		}
		headers[i] = h
		closers[i] = c
	}
	for _, h := range headers {
		w.open(h)
	}
	// Build a substitution that rewrites references to the param
	// inside the cond formula to the loop variable name.
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range some.Params {
		lv := &goivy.Const{Name: "__some_" + p.Name, CSort: p.CSort}
		subs[goivy.Key(p)] = lv
		_ = i
	}
	rewritten, err := goivy.Substitute(some.Fmla, subs)
	if err != nil {
		g.unsupported(w, "if-some substitution: %s", err.Error())
		// Close the loops we opened before returning.
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	cond, err := g.emitExpr(rewritten)
	if err != nil {
		g.unsupported(w, "if-some condition: %s", err.Error())
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	w.linef("if !__found && (%s) {", cond)
	w.line("\t__found = true")
	for i, p := range some.Params {
		w.linef("\t%s = __some_%s", witnessNames[i], goIdent(p.Name))
	}
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		// Wrap in a block so the witness assignments don't dangle.
		w.line("\t{")
		// emit then body — the witness names now hold the chosen
		// values; references to the param symbols in the body see
		// them since we declared them above.
		g.emitAction(w, thenAct)
		w.line("\t}")
	}
	w.line("}")
	for range closers {
		w.close("")
	}
	if a.ElseBody != nil {
		if elseAct, ok := a.ElseBody.(goivy.Action); ok {
			w.line("if !__found {")
			g.emitAction(w, elseAct)
			w.line("}")
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

// emitIfSomeMinMax lowers `if some X. p(X) minimizing/maximizing idx`.
// Mirrors ivy2cpp/action.go emitIfSomeMinMax:
//   - declare a __found flag and a __best_idx of the index sort
//   - declare per-param witness vars
//   - scan all candidates; for each (p(X), idx_expr), if it's better
//     than the current best, update __best_idx and witnesses
//   - dispatch THEN (with witnesses bound) or ELSE based on __found
//
// "Better" means strictly smaller index for some_min, strictly larger
// for some_max.
func (g *Generator) emitIfSomeMinMax(w *goWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) {
	if some.Index == nil {
		g.unsupported(w, "if-some %s missing index expression", some.Kind)
		return
	}
	if len(some.Params) == 0 {
		g.unsupported(w, "if-some %s with no params is malformed", some.Kind)
		return
	}
	cmp := "<"
	if some.Kind == "some_max" {
		cmp = ">"
	}
	idxType := g.goType(some.Index.NodeSort())
	w.open("{")
	w.line("__found := false")
	w.linef("var __best_idx %s", idxType)
	w.line("_ = __best_idx")
	// Witnesses declared at outer scope so THEN can read them.
	for _, p := range some.Params {
		w.linef("var %s %s", goIdent(p.Name), g.goType(p.CSort))
		w.linef("_ = %s", goIdent(p.Name))
	}
	// Open loop nest using fresh loop-var names. We substitute
	// references in the cond + index expression to point at these
	// loop vars; on a winning iteration we copy them into the
	// witnesses + best_idx.
	headers := make([]string, len(some.Params))
	closers := make([]string, len(some.Params))
	for i, p := range some.Params {
		loopName := "__some_" + goIdent(p.Name)
		h, c, err := g.loopHeaderForSort(p.CSort, loopName)
		if err != nil {
			g.unsupported(w, "if-some min/max bounds: %s", err.Error())
			w.close("")
			return
		}
		headers[i] = h
		closers[i] = c
	}
	for _, h := range headers {
		w.open(h)
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, p := range some.Params {
		lv := &goivy.Const{Name: "__some_" + p.Name, CSort: p.CSort}
		subs[goivy.Key(p)] = lv
	}
	rewrittenCond, err := goivy.Substitute(some.Fmla, subs)
	if err != nil {
		g.unsupported(w, "if-some min/max cond subst: %s", err.Error())
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	rewrittenIdx, err := goivy.Substitute(some.Index, subs)
	if err != nil {
		g.unsupported(w, "if-some min/max index subst: %s", err.Error())
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	cond, err := g.emitExpr(rewrittenCond)
	if err != nil {
		g.unsupported(w, "if-some min/max cond: %s", err.Error())
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	idx, err := g.emitExpr(rewrittenIdx)
	if err != nil {
		g.unsupported(w, "if-some min/max index: %s", err.Error())
		for range closers {
			w.close("")
		}
		w.close("")
		return
	}
	// First hit always sets __found + best; later hits only update
	// when the index strictly beats the current best.
	w.linef("if (%s) {", cond)
	w.linef("\t__cur_idx := %s", idx)
	w.linef("\tif !__found || __cur_idx %s __best_idx {", cmp)
	w.line("\t\t__found = true")
	w.line("\t\t__best_idx = __cur_idx")
	for _, p := range some.Params {
		w.linef("\t\t%s = __some_%s", goIdent(p.Name), goIdent(p.Name))
	}
	w.line("\t}")
	w.line("}")
	for range closers {
		w.close("")
	}
	// Dispatch THEN/ELSE based on __found.
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		w.line("if __found {")
		g.emitAction(w, thenAct)
		w.line("}")
	}
	if a.ElseBody != nil {
		if elseAct, ok := a.ElseBody.(goivy.Action); ok {
			w.line("if !__found {")
			g.emitAction(w, elseAct)
			w.line("}")
		}
	}
	w.close("")
}

// emitNativeAction emits an inline native Go block within an action
// body. Mirrors ivy2cpp/native.go's emitNativeAction.
//
// Like module-level native blocks (M10), each in-action block's
// first non-blank line is the language tag. Blocks tagged "go" or
// "go_*" emit their body verbatim into the enclosing action method;
// blocks with non-Go tags (e.g. "cpp") are skipped.
//
// Antiquote substitution of the Params slice is a future refinement
// (it would require parsing the body for $arg references); for now
// callers should write antiquote-free Go.
func (g *Generator) emitNativeAction(w *goWriter, a *goivy.LogicNativeAction) {
	codeNode, ok := a.Code.(*goivy.NativeCode)
	if !ok {
		g.unsupported(w, "native action code is %T, not *NativeCode", a.Code)
		return
	}
	tag, body := splitNativeCode(codeNode.Code)
	if !strings.HasPrefix(tag, "go") {
		// Non-Go-tagged in-action native: skip silently so a single
		// Ivy source can carry both cpp and go native actions.
		return
	}
	body = g.renderNativeTemplate(body, a.Params)
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return
	}
	for _, line := range strings.Split(body, "\n") {
		w.line(line)
	}
}

// emitLocal ports ivy2cpp/action.go emitLocal. Each local becomes a
// Go `var name T` declaration followed by nondet initialization via
// mkNondetSym (mirrors cpp's mk_nondet_sym call after sym_decl) so
// the body observes a randomized value rather than Go's zero.
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
		g.mkNondetSym(w, c, c.Name, a.UniqueID)
	}
	if bodyAct, ok := a.Body.(goivy.Action); ok {
		g.emitAction(w, bodyAct)
	}
	w.close("")
}

// actionParamName returns the Go identifier for an action's formal
// param or return, mangled to avoid colliding with the `s` method
// receiver. Mirrors goIdent's safety check for the local Action
// context.
func (g *Generator) actionParamName(c *goivy.Const) string {
	id := goIdent(c.Name)
	if id == "s" {
		return "s_"
	}
	return id
}

// importCallers ports ivy2cpp/generator.go importCallers — which in
// turn ports Python find_import_callers. Cached on the Generator.
func (g *Generator) importCallers() map[string]bool {
	if g.importCallersCache != nil {
		return g.importCallersCache
	}
	out := map[string]bool{}
	// Python: if target.get() != "test": return  (skip all)
	// We allow gen too — both targets drive actions via the
	// action_gen mechanism.
	if g.Config.Target == "test" || g.Config.Target == "gen" {
		if g.Mod != nil {
			for _, imp := range g.Mod.Imports {
				impDef, ok := imp.(*goivy.ImportDef)
				if !ok {
					continue
				}
				// Skip imports with a non-empty scope (those are
				// scoped to a sub-module, not the test boundary).
				if atom, ok := impDef.Scope.(*goivy.Atom); ok && atom.Relname() != "" {
					continue
				}
				atom, ok := impDef.Imported.(*goivy.Atom)
				if !ok {
					continue
				}
				name := atom.Relname()
				if name == "" {
					continue
				}
				if _, ok := g.Mod.Actions.Get2(name); !ok {
					continue
				}
				// Strip the 5-char `imp__` prefix Python uses; if
				// the name already has `ext:`, strip that too.
				caller := name
				switch {
				case strings.HasPrefix(caller, "imp__"):
					caller = strings.TrimPrefix(caller, "imp__")
				case strings.HasPrefix(caller, "ext:"):
					caller = strings.TrimPrefix(caller, "ext:")
				}
				out["ext:"+caller] = true
				out[caller] = true
			}
		}
	}
	g.importCallersCache = out
	return out
}

// emitTraceActionPrologue emits a single fmt.Fprintln to ivyTraceOut
// of the form `<dir> name(arg0,arg1,…)` (or `<dir> name` for
// zero-arg actions). Mirrors ivy2cpp's emitTraceActionPrologue but
// produces a single Go call instead of a chain of stream writes.
//
// dir is "<" (action body — observed from inside) or ">" (test
// driver — about to fire from outside). name is the raw action
// name without the `ext:` prefix.
func (g *Generator) emitTraceActionPrologue(w *goWriter, dir, name string, params []*goivy.Const) {
	display := strings.TrimPrefix(name, "ext:")
	g.Ctx.AddImport("actions", "fmt", "")
	if len(params) == 0 {
		w.linef(`fmt.Fprintln(ivyTraceOut, %q)`, dir+" "+display)
		return
	}
	// Build a printf-style format string: `< name(%v,%v,…)`
	var fmtStr strings.Builder
	fmtStr.WriteString(dir)
	fmtStr.WriteByte(' ')
	fmtStr.WriteString(display)
	fmtStr.WriteByte('(')
	for i := range params {
		if i > 0 {
			fmtStr.WriteByte(',')
		}
		fmtStr.WriteString("%v")
	}
	fmtStr.WriteString(")\n")
	args := make([]string, 0, len(params))
	for _, p := range params {
		args = append(args, g.actionParamName(p))
	}
	w.linef(`fmt.Fprintf(ivyTraceOut, %q, %s)`, fmtStr.String(), strings.Join(args, ", "))
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

// emitMethods walks the module's Actions map and emits one Go
// method on *State per action. Mirrors ivy2cpp/generator.go's
// per-action emission, simplified for M4: parameters become method
// params, return values become named return values, and the action
// body is the method body.
func (g *Generator) emitMethods(w *goWriter) {
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
		g.emitSomeAction(w, name, act)
	}
}

// emitSomeAction renders a single Go method on *State for an Ivy
// action.
func (g *Generator) emitSomeAction(w *goWriter, name string, act goivy.Action) {
	methodName := goExportedName(name)
	params := act.GetFormalParams()
	returns := act.GetFormalReturns()

	// Build parameter list. Mangle any param whose lowered name
	// would collide with the method receiver `s` so the emitted
	// signature stays valid Go (`func (s *State) F(s Super)` is a
	// redeclaration).
	paramParts := make([]string, 0, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		paramParts = append(paramParts, fmt.Sprintf("%s %s", g.actionParamName(p), g.goType(p.CSort)))
	}

	// Build return list. Named returns let `return` (without args)
	// pick them up automatically — useful when the action body relies
	// on havoc-initialised returns.
	returnParts := make([]string, 0, len(returns))
	for _, r := range returns {
		if r == nil {
			continue
		}
		returnParts = append(returnParts, fmt.Sprintf("%s %s", g.actionParamName(r), g.goType(r.CSort)))
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

	// Trace prologue: emit `< name(args)` for import-caller actions
	// (Python ivy_to_cpp.py:1607-1610 / find_import_callers). These
	// represent system→env callbacks: actions the system calls
	// whose impl is owned by the environment. The display name
	// strips `ext:` per Python trace_action (line 1581-1582).
	// `<` trace prologue — emitted at the top of each import-caller
	// action body so the test driver sees the system→env callback.
	// Mirrors ivy2cpp/generator.go emitSomeAction line 1284.
	if g.importCallers()[name] {
		g.emitTraceActionPrologue(w, "<", name, params)
	}

	// Track returns so emitReturn can use the right names.
	prev := g.currentReturns
	g.currentReturns = returns

	// Install alias rewrites for any formal param / return whose
	// lowered name collided with the receiver `s`. Without this
	// the body would emit `s` references that conflict with the
	// receiver; mangling at the signature alone isn't enough.
	type aliasEntry struct {
		key  string
		orig goivy.Expr
		had  bool
	}
	if g.exprAliases == nil {
		g.exprAliases = map[string]goivy.Expr{}
	}
	saved := make([]aliasEntry, 0)
	installRewrite := func(c *goivy.Const) {
		if c == nil {
			return
		}
		mangled := g.actionParamName(c)
		orig := goIdent(c.Name)
		if mangled == orig {
			return
		}
		prev, had := g.exprAliases[c.Name]
		saved = append(saved, aliasEntry{key: c.Name, orig: prev, had: had})
		// Use a fresh Const with the mangled name as the alias
		// target; emitExpr's Const path emits goIdent(name).
		g.exprAliases[c.Name] = &goivy.Const{Name: mangled, CSort: c.CSort}
	}
	for _, p := range params {
		installRewrite(p)
	}
	for _, r := range returns {
		installRewrite(r)
	}

	g.emitAction(w, act)

	// Restore prior alias state.
	for _, e := range saved {
		if e.had {
			g.exprAliases[e.key] = e.orig
		} else {
			delete(g.exprAliases, e.key)
		}
	}
	g.currentReturns = prev

	// Named returns: append a final `return` so the function
	// closes cleanly even if the body didn't write one.
	if len(returns) > 0 {
		w.line("return")
	}

	w.close("")
	w.blank()
}
