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
	// Python ivy_actions.py:669-690 SetAction has no `emit` assignment
	// in ivy_to_cpp.py — it is lowered upstream by `action_update`
	// before code emission. Go retains an explicit lowering through
	// `emitSet` for direct callers that synthesize SetAction nodes.
	case *goivy.LogicSetAction:
		g.emitSet(w, a)
	case *goivy.LogicAssertAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicRequiresAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicEnsuresAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicSubgoalAction:
		g.emitAssertLike(w, "ivy_assert", a.Formula, linenoStr(a.GetLineno()))
	case *goivy.LogicAssumeAction:
		g.emitAssertLike(w, "ivy_assume", a.Formula, linenoStr(a.GetLineno()))
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
	// Python ivy_actions.py:1192-1206 LetAction has no `emit`
	// assignment in ivy_to_cpp.py — it is lowered upstream by
	// `int_update` via `subst_action`. Go's `emitLet` performs an
	// equivalent alias-substitution rewrite for direct callers.
	case *goivy.LogicLetAction:
		g.emitLet(w, a)
	case *goivy.LogicBindOldsAction:
		g.emitBindOlds(w, a)
	case *goivy.LogicNativeAction:
		g.emitNativeAction(w, a)
	case *goivy.LogicDebugAction:
		g.emitDebug(w, a)
	case *goivy.LogicCrashAction:
		// Mirrors Python ivy_to_cpp.py:3888-3891:
		//     def emit_crash(self,header):
		//         pass
		// CrashAction is lowered upstream by action_update
		// (ivy_actions.py:1253-1266) into a Sequence of Havoc
		// actions before reaching emit. Python emits nothing if it
		// ever survives that lowering; mirror that behavior.
		_ = a
	case *goivy.ReturnAction:
		g.emitReturn(w)
	case *goivy.IgnoreAction:
		return
	case *goivy.LogicAssignFieldAction:
		g.emitAssignField(w, a)
	// Python ivy_actions.py:756-797 NullFieldAction/CopyFieldAction
	// have no `emit` assignment in ivy_to_cpp.py — they are pre-1.3
	// legacy nodes lowered upstream into AssignAction by
	// `make_field_update`. Go keeps direct emit helpers so synthesized
	// instances do not crash code generation.
	case *goivy.LogicNullFieldAction:
		g.emitNullField(w, a)
	case *goivy.LogicCopyFieldAction:
		g.emitCopyField(w, a)
	case *goivy.LogicThunkAction:
		// Python ivy_actions.py:1271-1280 ThunkAction has no
		// `emit` assignment in ivy_to_cpp.py — reaching emit would
		// AttributeError. The class is desugared upstream (the
		// inline comment on the Python class says so explicitly).
		g.unsupported(w, "thunk reached emit (Python ThunkAction has no emit; expected to be desugared upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicInstantiateAction:
		// Python ivy_actions.py:798-832 InstantiateAction has no
		// `emit` assignment in ivy_to_cpp.py — reaching emit would
		// AttributeError. The class is inlined upstream during
		// module composition.
		g.unsupported(w, "instantiate reached emit (Python InstantiateAction has no emit; expected to be inlined upstream): %s at %s",
			a.String(), a.GetLineno().String())
	case *goivy.LogicRanking:
		// Python ivy_actions.py:1050-1052 Ranking has no `emit`
		// assignment in ivy_to_cpp.py — reaching emit would
		// AttributeError. Ranking/progress declarations are
		// enforced by separate analysis, not by code emission.
		g.unsupported(w, "ranking reached emit (Python Ranking has no emit; ranking/progress is enforced separately): %s at %s",
			a.String(), a.GetLineno().String())
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

// emitHavoc mirrors Python `emit_havoc` (ivy_to_cpp.py:3768-3773):
//
//	def emit_havoc(self,header):
//	    print(self)
//	    print(self.lineno)
//	    assert False
//
// Python aborts code generation at the `assert False`. HavocAction is
// supposed to be eliminated upstream (lowered to `mk_nondet` nondet
// init paths); reaching emit is a bug. Mirror Python by refusing to
// emit and reporting the offending target and lineno through
// `g.unsupported`, which records into `g.errs` so `Generate` returns
// the diagnostic.
func (g *Generator) emitHavoc(w *cppWriter, a *goivy.LogicHavocAction) {
	target := "<nil>"
	if a.Target != nil {
		target = a.Target.String()
	}
	g.unsupported(w, "havoc reached emit (Python emit_havoc asserts False): %s at %s", target, a.GetLineno().String())
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

// emitAssign mirrors Python `emit_assign` (ivy_to_cpp.py:3703-3764). The
// extensional-relation-clear shortcut runs first; otherwise we dispatch on
// whether the LHS has free variables, and for quantified LHSs whether the
// loops can be opened or we need the thunk-based fallback. The simple,
// two-phase, and large emission bodies live in assign.go / thunk.go.
func (g *Generator) emitAssign(w *cppWriter, a *goivy.LogicAssignAction) {
	// All-false extensional-relation reset becomes `r.memo.clear();`.
	// Python falls through to emit_assign_large for this shape; the clear
	// is observationally equivalent for hash_thunk storage.
	if g.emitExtensionalRelationClear(w, a) {
		return
	}
	vs := goivy.VariablesAstList(a.LHS)
	if len(vs) == 0 {
		g.emitAssignSimple(w, a)
		return
	}
	body := g.assignBoundsExpr(a)
	if !g.canOpenAssignmentLoopsBounded(a.LHS, body) {
		g.emitAssignLarge(w, a, vs)
		return
	}
	g.emitAssignTwoPhase(w, a, vs)
}

func (g *Generator) closeAssignmentLoops(w *cppWriter, loops int) {
	for i := 0; i < loops; i++ {
		w.close("")
	}
}

func (g *Generator) openAssignmentLoops(w *cppWriter, lhs goivy.Expr) (int, bool) {
	vars := goivy.VariablesAstList(lhs)
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
	expr, err := g.emitExpr(closeFormula(f))
	if err != nil {
		g.unsupported(w, "unsupported assertion expression: %s", err.Error())
		return
	}
	if strings.TrimSpace(label) == "" {
		label = fn
	}
	w.linef(`%s(%s, "%s");`, fn, expr, escapeString(label))
}

func closeFormula(f goivy.Expr) goivy.Expr {
	if f == nil {
		return nil
	}
	return goivy.IvyForAll(goivy.FreeVariablesList(f), f)
}

// linenoStr mirrors Python ivy_utils.lineno_str (ivy_utils.py:285-291):
// render the AST's location, then drop a trailing ": ". Goivy's
// Location.String() always appends ": " after filename and line, which
// matches Python's __str__ — but Python strips that suffix before
// substituting into ivy_assert / ivy_assume labels (ivy_to_cpp.py:3794).
func linenoStr(loc goivy.Location) string {
	return strings.TrimSuffix(loc.String(), ": ")
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
	if g.Config.Target == "test" {
		w.open("if(" + cond + "){")
	} else {
		w.open("if (" + cond + ") {")
	}
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
	if some.Kind == "some_min" || some.Kind == "some_max" {
		g.emitIfSomeMinMax(w, a, some)
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
	headers, err := g.someConditionLoopHeaders(some)
	if err != nil {
		g.unsupported(w, "unsupported some parameter: %s", err.Error())
		return
	}
	opened := 0
	for _, h := range headers {
		w.open(h)
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

// someConditionLoopHeaders chooses per-parameter loop headers for an
// `if some` statement. Like someLoopHeaders, but the params arrive as
// *goivy.Const and must be lifted to LogicVariable so the bound walker
// (which follows Python `is_variable` semantics) can match them. The
// emitted loop variable names remain the original param names so the
// THEN body sees the expected identifiers.
func (g *Generator) someConditionLoopHeaders(some *goivy.SomeCondition) ([]string, error) {
	headers := make([]string, len(some.Params))
	useBounds := false
	if len(some.Params) > 0 && cppIsAnyIntegerType(g, some.Params[0].CSort) {
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
						h, herr := g.loopHeaderForSortBounds(p.CSort, varName(p.Name), bounds[i][0], bounds[i][1])
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
		h, err := g.loopHeaderForSort(p.CSort, varName(p.Name))
		if err != nil {
			return nil, err
		}
		headers[i] = h
	}
	return headers, nil
}

// emitIfSomeMinMax lowers `if some X. fmla minimizing/maximizing idx`. Mirrors
// Python emit_some's SomeMinMax branch (ivy_to_cpp.py:3486-3555): scan all
// candidates, track the current best index in a per-loop temp, and after the
// scan dispatch the THEN/ELSE bodies with the winning witness in scope.
func (g *Generator) emitIfSomeMinMax(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) {
	if some.Index == nil {
		g.unsupported(w, "unsupported %s condition: missing index expression", some.Kind)
		return
	}
	found := g.nextTemp("__ivy_some")
	bestIdx := g.nextTemp("__ivy_some_idx")
	idxSort := some.Index.NodeSort()
	idxType := g.cppType(idxSort)
	w.linef("bool %s = false;", found)
	w.linef("%s %s = %s;", idxType, bestIdx, g.cppZeroValue(idxSort))
	// One witness temp per parameter so the THEN body can refer to the
	// chosen value.
	witnesses := make([]string, len(some.Params))
	for i, p := range some.Params {
		w.linef("%s %s = %s;", g.cppType(p.CSort), varName("__ivy_some_w"+fmt.Sprintf("%d_%s", i, p.Name)), g.cppZeroValue(p.CSort))
		witnesses[i] = varName("__ivy_some_w" + fmt.Sprintf("%d_%s", i, p.Name))
	}
	headers, herr := g.someConditionLoopHeaders(some)
	if herr != nil {
		g.unsupported(w, "unsupported some parameter: %s", herr.Error())
		return
	}
	opened := 0
	for _, h := range headers {
		w.open(h)
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
	w.open(fmt.Sprintf("if (%s) {", cond))
	curIdx := g.nextTemp("__ivy_some_cur")
	idxExpr, err := g.emitExpr(some.Index)
	if err != nil {
		g.unsupported(w, "unsupported some index: %s", err.Error())
		w.close("")
		for i := 0; i < opened; i++ {
			w.close("")
		}
		return
	}
	w.linef("%s %s = %s;", idxType, curIdx, idxExpr)
	var cmp string
	if some.Kind == "some_min" {
		cmp = fmt.Sprintf("%s < %s", curIdx, bestIdx)
	} else {
		cmp = fmt.Sprintf("%s < %s", bestIdx, curIdx)
	}
	w.open(fmt.Sprintf("if (!%s || (%s)) {", found, cmp))
	w.linef("%s = true;", found)
	w.linef("%s = %s;", bestIdx, curIdx)
	for i, p := range some.Params {
		w.linef("%s = %s;", witnesses[i], varName(p.Name))
	}
	w.close("")
	// Python emit_some:3539-3540: if minimizing the first parameter, the
	// first hit during ascending iteration is the minimum, so exit the
	// loop. Helpful when scanning a wide integer domain.
	if firstParamIsIndex(some) {
		w.line("break;")
	}
	w.close("")
	for i := 0; i < opened; i++ {
		w.close("")
	}
	if thenAct, ok := a.ThenBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if (%s) {", found))
		// Bring the witness values into scope under the original param names
		// so the THEN body emits the expected identifiers.
		for i, p := range some.Params {
			w.linef("%s %s = %s;", g.cppType(p.CSort), varName(p.Name), witnesses[i])
		}
		g.emitAction(w, thenAct)
		w.close("")
	}
	if elseAct, ok := a.ElseBody.(goivy.Action); ok {
		w.open(fmt.Sprintf("if (!%s) {", found))
		g.emitAction(w, elseAct)
		w.close("")
	}
}

// firstParamIsIndex returns true when `some.Params[0]` is the same
// variable as `some.Index`. Mirrors Python `self.params()[0] == self.index()`
// (ivy_to_cpp.py:3539). The Index may surface as either a *Const or a
// *LogicVariable; compare by name.
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

func (g *Generator) emitIfSomeExtensional(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {
	if len(some.Params) != 1 {
		return false
	}
	p := some.Params[0]
	// `some` parameters compile to local Const symbols in goivy (e.g.
	// "loc:x"), while matchExtensionalBoundExprs follows Python's
	// is_variable semantics and only matches LogicVariable. Substitute
	// the Const back to a fresh Variable in the formula before searching.
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
	if !ok {
		return false
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	if st.Kind != cppStorageHashThunk {
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
	rel := varName(relName)
	found := g.nextTemp("__ivy_some")
	w.linef("bool %s = false;", found)
	w.open(fmt.Sprintf("for (auto it = %s.memo.begin(), en = %s.memo.end(); it != en; ++it) {", rel, rel))
	w.line("if (!it->second) continue;")
	// Emit using p.Name (the original local-const name) so that the
	// downstream emitAction sees the same identifier when it expands
	// the `then` body.
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
	w.open(fmt.Sprintf("if (%s.tag == %d) {", lhs, idx))
	w.linef("%s %s = %s;", g.cppType(v.CSort), varName(v.Name), g.variantDowncastExpr(lhs, app.Terms[0].NodeSort(), v.CSort, ""))
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

// emitChoice mirrors Python `emit_choice` (ivy_to_cpp.py:3976-3994):
//
//	def emit_choice(self,header):
//	    if len(self.args) == 1:
//	        self.args[0].emit(header)
//	        return
//	    tmp = new_temp(header)
//	    mk_nondet(header,tmp,len(self.args),"___branch",self.unique_id)
//	    for idx,arg in enumerate(self.args):
//	        indent(header)
//	        if idx != 0:
//	            header.append('else ')
//	        if idx != len(self.args)-1:
//	            header.append('if(' + tmp + ' == ' + str(idx) + ')');
//	        header.append('{\n')
//	        ...
//
// We emit an `if/else if/.../else` chain over a fresh int temp populated
// by `___ivy_choose`. Python's mk_nondet hardcodes 0 into the emitted
// `___ivy_choose` call (ivy_to_cpp.py:189) even though it receives
// `rng = len(self.args)`; we mirror that exactly via mkNondet.
func (g *Generator) emitChoice(w *cppWriter, a *goivy.LogicChoiceAction) {
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
	w.linef("int %s;", tmp)
	g.mkNondet(w, tmp, len(a.Branches), "___branch", a.UniqueID, nil)
	for idx, b := range a.Branches {
		prefix := ""
		if idx != 0 {
			prefix = "else "
		}
		if idx != len(a.Branches)-1 {
			w.open(fmt.Sprintf("%sif (%s == %d) {", prefix, tmp, idx))
		} else {
			w.open(prefix + "{")
		}
		if act, ok := b.(goivy.Action); ok {
			g.emitAction(w, act)
		}
		w.close("")
	}
}

// emitCall lowers a LogicCallAction. Mirrors Python emit_call
// (ivy_to_cpp.py:3811-3886): annotation-driven argument routing,
// alias-safety temporaries when an output position aliases an input,
// variant upcast at the argument boundary, and ___ivy_stack push/pop
// for gen/test targets.
func (g *Generator) emitCall(w *cppWriter, a *goivy.LogicCallAction) {
	name := a.CalleeName()
	fn, err := funName(name)
	if err != nil {
		g.unsupported(w, "unsupported call: %s", err.Error())
		return
	}

	var calleeAction goivy.Action
	if g.Mod != nil && g.Mod.Actions != nil {
		calleeAction, _ = g.Mod.Actions.Get2(name)
	}

	// Emit positional argument strings from app.Terms and capture the
	// per-position Expr for alias-safety checks. Apply per-argument
	// variant upcasts against the formal parameter sort.
	var args []string
	var argExprs []goivy.Expr
	if app, ok := a.Callee.(*goivy.Apply); ok {
		var formals []*goivy.Const
		if calleeAction != nil {
			formals = calleeAction.GetFormalParams()
		}
		for i, t := range app.Terms {
			s, err := g.emitExpr(t)
			if err != nil {
				g.unsupported(w, "unsupported call arg: %s", err.Error())
				return
			}
			if i < len(formals) {
				s = g.maybeVariantUpcast(formals[i].CSort, t.NodeSort(), s, "")
			}
			args = append(args, s)
			argExprs = append(argExprs, t)
		}
	}

	// Without a known callee we cannot annotate. Fall back to the
	// trailing-ref convention so built-ins keep working.
	if calleeAction == nil {
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
		return
	}

	_, rtypes := g.getParamTypes(name, calleeAction)
	nargs := len(args)

	// Walk returns. For each ReturnRefType return whose Pos lies inside
	// the input args, check whether the output target aliases the input
	// at that slot or any other input — if so, save the input to a
	// temporary, schedule a post-call copy-back, and rewrite args[pos]
	// to read from the temp. For ReturnRefType returns whose Pos is
	// beyond nargs, append the return target as a trailing argument.
	type pendingCopy struct {
		lhs string
		tmp string
	}
	var copies []pendingCopy

	for rpos := 0; rpos < len(rtypes) && rpos < len(a.ActualReturns); rpos++ {
		rrt, ok := rtypes[rpos].(ReturnRefType)
		if !ok {
			continue
		}
		rv := a.ActualReturns[rpos]
		pos := rrt.Pos
		if pos < nargs {
			iparg := argExprs[pos]
			aliases := false
			for j, other := range argExprs {
				if j == pos {
					continue
				}
				if g.mayAlias(other, iparg) {
					aliases = true
					break
				}
			}
			if !exprRootEqByName(iparg, rv) || aliases {
				tmp := g.nextTemp("__tmp")
				w.linef("%s %s = %s;", g.cppType(rv.NodeSort()), tmp, args[pos])
				args[pos] = tmp
				lhsStr, err := g.emitExpr(rv)
				if err != nil {
					g.unsupported(w, "unsupported call return: %s", err.Error())
					return
				}
				copies = append(copies, pendingCopy{lhs: lhsStr, tmp: tmp})
			}
		} else {
			extra, err := g.emitExpr(rv)
			if err != nil {
				g.unsupported(w, "unsupported call return: %s", err.Error())
				return
			}
			args = append(args, extra)
		}
	}

	// Primary return assignment prefix: emitted only when rtypes[0] is
	// not a ReturnRefType (Python ivy_to_cpp.py:3862-3864).
	prefix := ""
	if len(rtypes) >= 1 && len(a.ActualReturns) >= 1 {
		if _, isRRT := rtypes[0].(ReturnRefType); !isRRT {
			lhs, err := g.emitExpr(a.ActualReturns[0])
			if err != nil {
				g.unsupported(w, "unsupported call return: %s", err.Error())
				return
			}
			prefix = lhs + " = "
		}
	}

	stacked := g.emitCallStackPush(w, a)
	w.linef("%s%s(%s);", prefix, fn, strings.Join(args, ", "))
	for _, c := range copies {
		w.linef("%s = %s;", c.lhs, c.tmp)
	}
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

// emitLocal mirrors Python `local_start` + `emit_local`
// (ivy_to_cpp.py:3893-3917):
//
//	def local_start(header,params,nondet_id=None):
//	    indent(header); header.append('{\n')
//	    indent_level += 1
//	    for p in params:
//	        indent(header); code_line(header,sym_decl(p))
//	        if nondet_id != None:
//	            mk_nondet_sym(header,p,p.name,nondet_id)
//
//	def emit_local(self,header):
//	    local_start(header,self.args[0:-1],self.unique_id)
//	    self.args[-1].emit(header)
//	    local_end(header)
//
// Each local is first declared uninitialized via `cppStorageDecl` (the
// Go equivalent of `sym_decl`), then nondet-initialized via
// `mkNondetSym` using the LocalAction's UniqueID. The body is emitted
// inside the same block scope.
func (g *Generator) emitLocal(w *cppWriter, a *goivy.LogicLocalAction) {
	w.open("{")
	for _, local := range a.Locals {
		name := goivy.ExprName(local)
		w.linef("%s;", g.cppStorageDecl(name, local.NodeSort(), ""))
		g.mkNondetSym(w, local, name, a.UniqueID)
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

// emitBindOlds mirrors Python's absence of an `emit_bind_olds`
// function (no assignment to `ia.BindOldsAction.emit` exists in
// ivy_to_cpp.py). `BindOldsAction` is supposed to be eliminated
// upstream during `int_update` via `bind_olds_action`
// (ivy_transrel.py:240); reaching emit is a bug, exactly like
// `emit_havoc`'s `assert False`. Mirror Python by refusing to emit and
// reporting the offending bindolds wrapper through `g.unsupported`.
func (g *Generator) emitBindOlds(w *cppWriter, a *goivy.LogicBindOldsAction) {
	inner := "<nil>"
	if a.Inner != nil {
		inner = fmt.Sprintf("%T", a.Inner)
	}
	g.unsupported(w, "bindolds reached emit (Python has no emit_bind_olds): inner=%s at %s", inner, a.GetLineno().String())
}

// emitAssignField handles a LogicAssignFieldAction by synthesizing the
// equivalent LogicAssignAction(Apply(field, obj), value) and routing it
// through emitAssign. This matches the modern Python representation
// where `obj.field := value` parses directly into AssignAction with a
// destructor-decomposed LHS (Python ivy_parser.py:3082 only constructs
// AssignFieldAction under iu.get_numeric_version() <= [1,2]).
//
// As a result this path picks up emitAssign's variant upcast, two-phase
// quantified-update handling, and emit_assign_large thunk fallback for
// free if a future caller synthesizes a LogicAssignFieldAction.
func (g *Generator) emitAssignField(w *cppWriter, a *goivy.LogicAssignFieldAction) {
	if a == nil || a.Field == nil || a.Obj == nil {
		g.unsupported(w, "unsupported field assignment: nil components")
		return
	}
	lhs := goivy.NewApplyUnchecked(a.Field, a.Obj)
	synth := goivy.NewAssignAction(lhs, a.Value)
	g.emitAssign(w, synth)
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

// emitDebug mirrors Python emit_debug (ivy_to_cpp.py:3998-4026). The
// compiler stores names parallel to values in LogicDebugAction
// (WithNames, populated by compiler_phase6.go). For each clause we emit
// the name as the JSON key and run emitPrintExpr on the value so
// quantified expressions open loops over their free variables and wrap
// in [...] like Python's emit_print_expr.
func (g *Generator) emitDebug(w *cppWriter, a *goivy.LogicDebugAction) {
	event := debugEventName(a.DebugExpr)
	w.line(`std::cout << "{" << std::endl;`)
	w.linef(`std::cout << "    \"event\" : \"%s\"," << std::endl;`, escapeString(event))
	for i, e := range a.WithExprs {
		name := ""
		if i < len(a.WithNames) {
			name = a.WithNames[i]
		}
		if name == "" {
			// Fallback: synthetic DebugActions (e.g. unit tests) may pass
			// bare exprs without a parallel name slot. Derive a key from
			// the expression itself so the output is still well-formed.
			name = goivy.ExprName(e)
		}
		w.linef(`std::cout << "    \"%s\" : ";`, escapeString(name))
		g.emitPrintExpr(w, e)
		w.line(`std::cout << "," << std::endl;`)
	}
	w.line(`std::cout << "}" << std::endl;`)
}

// emitPrintExpr mirrors Python emit_print_expr (ivy_to_cpp.py:3989-3996):
// for each free variable in expr, open a loop over its sort and emit
// `[` / `]` brackets around the value with `,` separators across
// iterations. With no free variables, this is a simple `std::cout <<
// (expr)`.
func (g *Generator) emitPrintExpr(w *cppWriter, expr goivy.Expr) {
	vs := goivy.VariablesAstList(expr)
	opened := 0
	for _, v := range vs {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported debug print variable %s: %s", varName(v.Name), err.Error())
			g.closeAssignmentLoops(w, opened)
			return
		}
		w.line(`std::cout << "[";`)
		w.open(header)
		w.linef(`if (%s) std::cout << ",";`, varName(v.Name))
		opened++
	}
	value, err := g.emitExpr(expr)
	if err != nil {
		g.unsupported(w, "unsupported debug print expression: %s", err.Error())
		g.closeAssignmentLoops(w, opened)
		return
	}
	w.linef(`std::cout << (%s);`, value)
	for i := 0; i < opened; i++ {
		w.close("")
		w.line(`std::cout << "]";`)
	}
}

// debugEventName extracts the event-name string from the DebugExpr. The
// parser emits the event as a quoted string literal Const (e.g.
// `Const("\"tick\"")`); we strip the surrounding quotes here and let
// the caller wrap + escape per Python's quote() helper inline in
// emit_debug (ivy_to_cpp.py:4001-4005).
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
