// match_annotation: Annotation-guided recursive action decomposition.
// Walks action/annotation pairs, calling handler on matching assert/assume actions.
// Essential for proof/counterexample extraction.
//
// This is a port of Python's match_annotation in ivy_actions.py.
package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// AnnotationHandler is called by MatchAnnotation to process actions during trace reconstruction.
// In Python, eval receives lg.Const and env maps lg.Const → lg.Const using structural
// equality (__hash__/__eq__). In Go, eval receives lg.Expr and env uses lg.NodeKey
// (Sexp-based structural identity) as map keys, matching Python's behavior.
type AnnotationHandler interface {
	// Eval evaluates a condition (an lg.Expr, typically *lg.Const) and returns true/false.
	Eval(cond Expr) bool
	// Handle processes an action with the given environment mapping.
	// env maps lg.NodeKey → lg.Expr using structural identity.
	Handle(action ActionsAction, env map[NodeKey]Expr)
	// DoReturn processes a return action.
	DoReturn(action ActionsAction, env map[NodeKey]Expr)
	// Fail marks a failure point in the trace.
	Fail()
}

// MatchAnnotation walks an action/annotation pair, calling handler methods
// to reconstruct an execution trace from a satisfying assignment.
// Corresponds to Python's match_annotation.
// The mod parameter provides the module for resolving callee actions in CallAction.
// If mod is nil, CallAction will not be inlined.
func MatchAnnotation(action ActionsAction, annot Annotation, handler AnnotationHandler, mod *Module) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*AnnotationError); ok {
				fmt.Println("internal error: cannot convert satisfying assignment to program trace")
			} else {
				panic(r)
			}
		}
	}()
	xtracer.Trace("actions.match_annotation ENTER actionType=%s annotType=%s", actionTypeNameForTrace(action), annotationTypeName(annot))
	matchAnnotationRecur(action, annot, make(map[NodeKey]Expr), handler, -1, mod)
	xtracer.Trace("actions.match_annotation EXIT")
}

// matchAnnotationRecur is the recursive core of MatchAnnotation.
// pos is the position within a Sequence (-1 means use full length).
func matchAnnotationRecur(action ActionsAction, annot Annotation, env map[NodeKey]Expr, handler AnnotationHandler, pos int, mod *Module) {
	// Handle RenameAnnotation: update env and recurse
	// Python: for x,y in annot.map.items():
	//             if x in env: save[x] = env[x]
	//             env[x] = env.get(y,y)
	// Keys and lookups use lg.NodeKey (structural equality via Sexp).
	if ra, ok := annot.(*RenameAnnotation); ok {
		xtracer.Trace("actions.match_annotation.Rename ENTER actionType=%s annotType=%s nmap=%d env=%d pos=%d", actionTypeNameForTrace(action), annotationTypeName(ra.Arg), len(ra.Map), len(env), pos)
		save := make(map[NodeKey]Expr)
		for x, y := range ra.Map {
			if old, exists := env[x]; exists {
				save[x] = old
			}
			// Python: env[x] = env.get(y, y) — look up y by structural key, default to y itself
			yKey := Key(y)
			if mapped, exists := env[yKey]; exists {
				env[x] = mapped
			} else {
				env[x] = y
			}
		}
		matchAnnotationRecur(action, ra.Arg, env, handler, pos, mod)
		// Python only restores entries that were present before the rename.
		// Newly-added bindings remain visible to later annotation siblings.
		for x, v := range save {
			env[x] = v
		}
		xtracer.Trace("actions.match_annotation.Rename EXIT env=%d restored=%d", len(env), len(save))
		return
	}

	// Handle Sequence
	if seq, ok := action.(*LogicSequence); ok {
		if pos < 0 {
			pos = len(seq.Elems)
		}
		xtracer.Trace("actions.match_annotation.Sequence ENTER pos=%d nElems=%d annotType=%s env=%d", pos, len(seq.Elems), annotationTypeName(annot), len(env))
		if pos == 0 {
			if _, ok := annot.(EmptyAnnotation); !ok {
				fmt.Println("annotation error: should be empty annotation")
			}
			return
		}

		// Check for IteAnnotation (branch failure point)
		if ite, ok := annot.(*IteAnnotation); ok {
			rncond := envGetExpr(env, ite.Cond)
			cond := handler.Eval(rncond)
			xtracer.Trace("actions.match_annotation.SequenceIte pos=%d condType=%s result=%v cond HASH canon=%s rncond HASH canon=%s", pos, ShortTypeName(ite.Cond), cond, exprCanonForTrace(ite.Cond), exprCanonForTrace(rncond))
			if cond {
				matchAnnotationRecur(action, ite.ThenB, env, handler, pos, mod)
				return
			}
			matchAnnotationRecur(action, ite.ElseB, env, handler, pos-1, mod)
			return
		}

		// Should be ComposeAnnotation
		compose, ok := annot.(*ComposeAnnotation)
		if !ok {
			fmt.Println("annotation error: should be ComposeAnnotation")
			return
		}
		if len(compose.Args) < 2 {
			fmt.Println("annotation error: ComposeAnnotation should have 2 args")
			return
		}
		childType := "nil"
		if pos-1 >= 0 && pos-1 < len(seq.Elems) {
			childType = actionTypeNameForTrace(extractActionFromNode(seq.Elems[pos-1]))
		}
		xtracer.Trace("actions.match_annotation.SequenceCompose pos=%d childType=%s leftAnnot=%s rightAnnot=%s", pos, childType, annotationTypeName(compose.Args[0]), annotationTypeName(compose.Args[1]))
		matchAnnotationRecur(action, compose.Args[0], env, handler, pos-1, mod)
		childAction := extractActionFromNode(seq.Elems[pos-1])
		if childAction != nil {
			matchAnnotationRecur(childAction, compose.Args[1], env, handler, -1, mod)
		}
		return
	}

	// Handle IfAction
	if ifAct, ok := action.(*LogicIfAction); ok {
		ite, ok := annot.(*IteAnnotation)
		if !ok {
			fmt.Println("annotation error: IfAction should have IteAnnotation")
			return
		}
		isSome := matchAnnotationIfConditionIsSome(ifAct)
		rncond := envGetExpr(env, ite.Cond)
		cond := handler.Eval(rncond)
		xtracer.Trace("actions.match_annotation.IfAction isSome=%v condType=%s result=%v cond HASH canon=%s rncond HASH canon=%s", isSome, ShortTypeName(ite.Cond), cond, exprCanonForTrace(ite.Cond), exprCanonForTrace(rncond))
		if cond {
			thenAction := extractActionFromNode(ifAct.ThenBody)
			if thenAction != nil {
				if isSome {
					thenAction = matchAnnotationWrapSomeBranch(thenAction)
				}
				matchAnnotationRecur(thenAction, ite.ThenB, env, handler, -1, mod)
			}
		} else {
			if ifAct.ElseBody != nil {
				elseAction := extractActionFromNode(ifAct.ElseBody)
				if elseAction != nil {
					if isSome {
						elseAction = matchAnnotationWrapSomeBranch(elseAction)
					}
					matchAnnotationRecur(elseAction, ite.ElseB, env, handler, -1, mod)
				}
			}
		}
		return
	}

	// Handle ChoiceAction. EnvAction subclasses ChoiceAction in Python;
	// in Go it embeds ChoiceAction, so handle both concrete types here.
	var branches []Expr
	if choice, ok := action.(*LogicChoiceAction); ok {
		branches = choice.Branches
	} else if envAct, ok := action.(*LogicEnvAction); ok {
		branches = envAct.Branches
	}
	if branches != nil {
		ite, ok := annot.(*IteAnnotation)
		if !ok {
			fmt.Println("annotation error: ChoiceAction should have IteAnnotation")
			return
		}

		// Handle EnvAction with label
		if envAct, ok := action.(*LogicEnvAction); ok {
			if envAct.GetLabel() != "" {
				handler.Handle(envAct, env)
			}
		}

		annots := UniteAnnot(ite)
		if len(annots) != len(branches) {
			fmt.Printf("annotation error: %d annots but %d branches\n", len(annots), len(branches))
			return
		}
		xtracer.Trace("actions.match_annotation.ChoiceAction nBranches=%d nAnnots=%d env=%d", len(branches), len(annots), len(env))

		// Walk branches in reverse order, pick the first whose condition is true
		for i := len(branches) - 1; i >= 0; i-- {
			rncond := envGetExpr(env, annots[i].Cond)
			condResult := handler.Eval(rncond)
			xtracer.Trace("actions.match_annotation.ChoiceBranch idx=%d actionType=%s condType=%s result=%v cond HASH canon=%s rncond HASH canon=%s", i, actionTypeNameForTrace(extractActionFromNode(branches[i])), ShortTypeName(annots[i].Cond), condResult, exprCanonForTrace(annots[i].Cond), exprCanonForTrace(rncond))
			if condResult {
				branchAction := extractActionFromNode(branches[i])

				// Handle EnvAction without label
				if envAct, ok := action.(*LogicEnvAction); ok {
					if envAct.GetLabel() == "" && branchAction != nil {
						label := "unknown"
						if labeled, ok := branchAction.(interface{ GetLabel() string }); ok && labeled.GetLabel() != "" {
							label = labeled.GetLabel()
						}
						callAct := matchAnnotationEnvCallAction(envAct, branchAction)
						callAct.Label = "call " + label
						handler.Handle(callAct, env)
					}
				}

				if branchAction != nil {
					matchAnnotationRecur(branchAction, annots[i].Ann, env, handler, -1, mod)
				}
				return
			}
		}
		panic(&AnnotationError{Msg: "problem in match_annotation: no branch matched"})
	}

	// Handle CallAction
	// Python: handler.handle(action, env)
	//         callee = ivy_module.module.actions[action.args[0].rep]
	//         seq = LogicSequence(IgnoreAction(), callee, ReturnAction())
	//         recur(seq, annot, env, None)
	if callAct, ok := action.(*LogicCallAction); ok {
		xtracer.Trace("actions.match_annotation.CallAction callee=%s env=%d", callAct.CalleeName(), len(env))
		handler.Handle(action, env)
		if mod != nil {
			calleeName := callAct.CalleeName()
			if calleeIface, ok := mod.Actions.Get2(calleeName); ok {
				if callee, ok := calleeIface.(ActionsAction); ok {
					// Build: Sequence(IgnoreAction(), callee, ReturnAction())
					seq := NewSequence(
						&IgnoreAction{},
						callee,
						&ReturnAction{},
					)
					matchAnnotationRecur(seq, annot, env, handler, -1, mod)
				}
			}
		}
		return
	}

	// Handle WhileAction
	// Python: expanded = action.expand(ivy_module.module, [])
	//         recur(expanded, annot, env)
	if whileAct, ok := action.(*LogicWhileAction); ok {
		// Expand the while loop into an if/sequence structure
		// Python's expand creates: if cond { body; while(cond, body) } else { assume(~cond) }
		expanded := expandWhile(whileAct, mod)
		if expanded != nil {
			matchAnnotationRecur(expanded, annot, env, handler, -1, mod)
		} else {
			handler.Handle(action, env)
		}
		return
	}

	// Handle ReturnAction
	if _, ok := action.(*ReturnAction); ok {
		xtracer.Trace("actions.match_annotation.ReturnAction env=%d", len(env))
		handler.DoReturn(action, env)
		return
	}

	// Handle IgnoreAction
	if _, ok := action.(*IgnoreAction); ok {
		xtracer.Trace("actions.match_annotation.IgnoreAction")
		return
	}

	// Handle LocalAction
	if local, ok := action.(*LogicLocalAction); ok {
		xtracer.Trace("actions.match_annotation.LocalAction env=%d", len(env))
		bodyAction := extractActionFromNode(local.Body)
		if bodyAction != nil {
			matchAnnotationRecur(bodyAction, annot, env, handler, -1, mod)
		}
		return
	}

	if failed, ok := action.(interface{ FailedAction() ActionsAction }); ok {
		xtracer.Trace("actions.match_annotation.FailedAction actionType=%s env=%d", actionTypeNameForTrace(action), len(env))
		if inner := failed.FailedAction(); inner != nil {
			matchAnnotationRecur(inner, annot, env, handler, -1, mod)
		}
		handler.Fail()
		return
	}

	// Default: handle the action directly
	xtracer.Trace("actions.match_annotation.Handle actionType=%s env=%d", actionTypeNameForTrace(action), len(env))
	handler.Handle(action, env)
}

// AnnotBranch represents one branch of a flattened IteAnnotation.
type AnnotBranch struct {
	Cond Expr
	Ann  Annotation
}

// UniteAnnot flattens a nested IteAnnotation into a list of (cond, annotation) pairs.
// Corresponds to Python's unite_annot.
func UniteAnnot(annot Annotation) []AnnotBranch {
	switch a := annot.(type) {
	case *RenameAnnotation:
		inner := UniteAnnot(a.Arg)
		result := make([]AnnotBranch, len(inner))
		for i, b := range inner {
			cond := b.Cond
			// Python: cond lookup in rename map by structural identity
			if mapped, ok := a.Map[Key(cond)]; ok {
				cond = mapped
			}
			result[i] = AnnotBranch{
				Cond: cond,
				Ann:  newRenameAnnotation(b.Ann, a.Map),
			}
		}
		return result
	case *IteAnnotation:
		res := UniteAnnot(a.ElseB)
		res = append(res, AnnotBranch{Cond: a.Cond, Ann: a.ThenB})
		return res
	default:
		return nil
	}
}

// envGetExpr looks up an lg.Expr condition in the environment by structural
// identity (NodeKey), returning the mapped value or the original expression
// if not found. Matches Python: rncond = env.get(annot.cond, annot.cond)
func envGetExpr(env map[NodeKey]Expr, cond Expr) Expr {
	if v, ok := env[Key(cond)]; ok {
		return v
	}
	return cond
}

func annotationTypeName(annot Annotation) string {
	if annot == nil {
		return "nil"
	}
	return TypeName(annot)
}

func actionTypeNameForTrace(action ActionsAction) string {
	if action == nil || isNil(action) {
		return "nil"
	}
	return ActionTypeName(action)
}

func exprCanonForTrace(expr Expr) Canonical {
	if expr == nil {
		return "nil"
	}
	return expr.Canon()
}

func matchAnnotationIfConditionIsSome(ifAct *LogicIfAction) bool {
	if _, ok := ifAct.Cond.(*SomeCondition); ok {
		return true
	}
	switch ifAct.AstCond.(type) {
	case *Some, *SomeMin, *SomeMax:
		return true
	default:
		return false
	}
}

func matchAnnotationWrapSomeBranch(action ActionsAction) ActionsAction {
	var emptyAnd Expr
	emptyAnd, err := NewAnd()
	if err != nil {
		emptyAnd = True
	}
	return NewSequence(NewAssumeAction(emptyAnd), action)
}

func matchAnnotationEnvCallAction(envAct *LogicEnvAction, branchAction ActionsAction) *LogicEnvAction {
	if envAct.ActCfg != nil {
		return NewEnvActionOn(envAct.ActCfg, branchAction)
	}
	return &LogicEnvAction{LogicChoiceAction: LogicChoiceAction{Branches: []Expr{branchAction}}}
}

// extractActionFromNode tries to extract an Action from a lg.Expr.
func extractActionFromNode(n interface{}) ActionsAction {
	if n == nil {
		return nil
	}
	if act, ok := n.(ActionsAction); ok {
		return act
	}
	// Try wrapper types (e.g. TacticNodeWrapper)
	type wrapper interface {
		GetAction() ActionsAction
	}
	if w, ok := n.(wrapper); ok {
		return w.GetAction()
	}
	return nil
}

// expandWhile expands a WhileAction into a Sequence of havocs, invariant
// checks, and a conditional body+invariant re-check.
//
// This is a faithful port of Python ivy_actions.py WhileAction.expand():
//
//	def expand(self, domain, pvars):
//	    modset, pre, post = self.args[1].int_update(domain, pvars)
//	    if isinstance(self.args[-1], Ranking):
//	        asserts = self.args[2:-1]
//	        decreases = self.args[-1]
//	    else:
//	        asserts = self.args[2:]
//	        decreases = None
//	    assumes = [a.assert_to_assume([AssertAction]) for a in asserts
//	               if not isinstance(a, SubgoalAction)]
//	    asserts = [a for a in asserts if not isinstance(a, AssumeAction)]
//	    entry_asserts = []
//	    exit_asserts = []
//	    if decreases is not None:
//	        rank = decreases.args[0]
//	        aux = Symbol('$rank', rank.sort)
//	        assumes.append(AssumeAction(Equals(aux, rank)))
//	        assumes[-1].lineno = decreases.lineno
//	        ltsym = Symbol('<', LogicRelationSort([rank.sort, rank.sort]))
//	        exit_asserts.append(AssertAction(ltsym(rank, aux)))
//	        exit_asserts[-1].lineno = decreases.lineno
//	        entry_asserts.append(AssertAction(Not(ltsym(rank, Symbol('0', rank.sort)))))
//	        entry_asserts[-1].lineno = decreases.lineno
//	    havocs = [HavocAction(sym) for sym in modset]
//	    for h in havocs:
//	        h.lineno = self.lineno
//	    res = LogicSequence(*(
//	        asserts + havocs + assumes +
//	        [IfAction(self.args[0],
//	            Sequence(*(entry_asserts + [self.args[1]] + exit_asserts + asserts + [AssumeAction(Or())])),
//	            Sequence())]))
//	    if decreases is not None:
//	        res = LogicLocalAction(aux, res)
//	    return res
func expandWhile(w *LogicWhileAction, mod *Module) ActionsAction {
	// Step 1: compute the modset by getting int_update of the body.
	// We need the modified set to generate havocs.
	var modset []*Const
	if mod != nil {
		bodyAction := extractActionFromNode(w.Body)
		if bodyAction != nil {
			ctx := &UpdateContext{
				Domain:          mod,
				PVars:           make(map[string]bool),
				ActCfg:          mod.Cfg.ActCfg,
				Instantiator:    mod.Instantiator,
				CheckUnprovable: mod.Cfg.OnlyCheckUnprovable,
				CheckedAssert:   mod.Cfg.CheckLineno,
				GetAction: func(name string) ActionsAction {
					if mod.Actions != nil {
						if v, ok := mod.Actions.Get2(name); ok {
							if act, ok := v.(ActionsAction); ok {
								return act
							}
						}
					}
					return nil
				},
			}
			xtracer.Trace("actions.match calling IntUpdate type=%s", ActionTypeName(bodyAction))
			update := IntUpdate(bodyAction, ctx)
			if update != nil && !update.ModifiedAll {
				modset = update.Modified
			}
		}
	}

	// Step 2: Separate invariants from ranking.
	// In Python: self.args[2:] are invariants, last may be Ranking.
	// In Go: WhileAction.Invariants are the invariant actions.
	var asserts []ActionsAction // invariant assertions
	var decreasesRanking *LogicRanking

	for _, inv := range w.Invariants {
		// Check if it's a RankingWrapper (not an action)
		if rk, ok := inv.(*RankingWrapper); ok {
			decreasesRanking = rk.Ranking
			continue
		}
		invAction := extractActionFromNode(inv)
		if invAction == nil {
			continue
		}
		asserts = append(asserts, invAction)
	}

	// Step 3: Build assumes from asserts (assert_to_assume).
	// Python: assumes = [a.assert_to_assume([AssertAction]) for a in asserts
	//                    if not isinstance(a, SubgoalAction)]
	assertKinds := map[string]bool{"assert": true}
	var assumes []ActionsAction
	for _, a := range asserts {
		if _, isSub := a.(*LogicSubgoalAction); isSub {
			continue
		}
		assumes = append(assumes, AssertToAssume(a, assertKinds, mod.Cfg.IuCfg))
	}

	// Filter asserts: remove any that became AssumeActions
	var filteredAsserts []ActionsAction
	for _, a := range asserts {
		if _, isAssume := a.(*LogicAssumeAction); isAssume {
			continue
		}
		filteredAsserts = append(filteredAsserts, a)
	}
	asserts = filteredAsserts

	// Step 4: Handle ranking/decreases.
	var entryAsserts []ActionsAction
	var exitAsserts []ActionsAction
	var auxVar Expr // for LocalAction wrapper

	if decreasesRanking != nil && len(decreasesRanking.RArgs) > 0 {
		rank := decreasesRanking.RArgs[0]
		rankSort := rank.NodeSort()

		// aux = Symbol('$rank', rank.sort)
		aux := NewConst("$rank", rankSort)
		auxVar = aux

		// assumes.append(AssumeAction(Equals(aux, rank)))
		eqFmla := &Eq{T1: aux, T2: rank}
		assumeEq := NewAssumeAction(eqFmla)
		if decreasesRanking.HasLineno() {
			assumeEq.SetLineno(decreasesRanking.GetLineno())
		}
		assumes = append(assumes, assumeEq)

		// ltsym = Symbol('<', LogicRelationSort([rank.sort, rank.sort]))
		ltSort := &LogicFunctionSort{Sorts: []Sort{rankSort, rankSort, Boolean}}
		ltSym := NewConst("<", ltSort)

		// exit_asserts.append(AssertAction(ltsym(rank, aux)))
		ltApp := MustApply(ltSym, rank, aux)
		exitAssert := NewAssertAction(ltApp)
		if decreasesRanking.HasLineno() {
			exitAssert.SetLineno(decreasesRanking.GetLineno())
		}
		exitAsserts = append(exitAsserts, exitAssert)

		// entry_asserts.append(AssertAction(Not(ltsym(rank, Symbol('0', rank.sort)))))
		zeroSym := NewConst("0", rankSort)
		ltZero := MustApply(ltSym, rank, zeroSym)
		entryAssert := NewAssertAction(&LogicNot{Body: ltZero})
		if decreasesRanking.HasLineno() {
			entryAssert.SetLineno(decreasesRanking.GetLineno())
		}
		entryAsserts = append(entryAsserts, entryAssert)
	}

	// Step 5: Build havocs for modified symbols.
	var havocs []ActionsAction
	if mod != nil {
		for _, modSym := range modset {
			if sym, ok := mod.Sig.Symbols.Get2(modSym.Name); ok {
				havocTarget := NewConst(modSym.Name, sym.Sort)
				h := NewHavocAction(havocTarget)
				h.SetLineno(w.GetLineno())
				havocs = append(havocs, h)
			}
		}
	}

	// Step 6: Build the result Sequence.
	// Python:
	// res = LogicSequence(*(
	//     asserts + havocs + assumes +
	//     [IfAction(self.args[0],
	//         Sequence(*(entry_asserts + [self.args[1]] + exit_asserts + asserts + [AssumeAction(Or())])),
	//         Sequence())]))

	// Build the then-branch of the IfAction:
	// Sequence(entry_asserts + [body] + exit_asserts + asserts + [AssumeAction(Or())])
	var thenParts []Expr
	for _, ea := range entryAsserts {
		thenParts = append(thenParts, ea)
	}
	thenParts = append(thenParts, w.Body) // the loop body
	for _, xa := range exitAsserts {
		thenParts = append(thenParts, xa)
	}
	for _, a := range asserts {
		thenParts = append(thenParts, a)
	}
	// AssumeAction(Or()) = assume false (empty disjunction)
	assumeFalse := NewAssumeAction(&LogicOr{Terms: nil})
	assumeFalse.SetLineno(w.GetLineno())
	thenParts = append(thenParts, assumeFalse)

	thenSeq := NewSequence(thenParts...)
	thenSeq.SetLineno(w.GetLineno())
	elseSeq := NewSequence() // empty Sequence
	elseSeq.SetLineno(w.GetLineno())

	ifAction := NewIfAction(w.Cond, thenSeq, elseSeq)
	ifAction.SetLineno(w.GetLineno())

	// Build the outer Sequence: asserts + havocs + assumes + [ifAction]
	var outerParts []Expr
	for _, a := range asserts {
		outerParts = append(outerParts, a)
	}
	for _, h := range havocs {
		outerParts = append(outerParts, h)
	}
	for _, a := range assumes {
		outerParts = append(outerParts, a)
	}
	outerParts = append(outerParts, ifAction)

	outerSeq := NewSequence(outerParts...)
	outerSeq.SetLineno(w.GetLineno())
	var res ActionsAction = outerSeq

	// Step 7: Wrap in LocalAction if decreases ranking was used.
	// Python: if decreases is not None: res = LogicLocalAction(aux, res)
	if auxVar != nil {
		actCfg := mod.Cfg.ActCfg
		res = NewLocalActionOn(actCfg, "actions.WhileAction.action_update", auxVar, res)
	}

	return res
}

// RankingWrapper wraps a Ranking as a lg.Expr for storage in WhileAction.Invariants.
type RankingWrapper struct {
	Base
	Ranking *LogicRanking
}

func (rw *RankingWrapper) NodeSort() Sort    { return Boolean }
func (rw *RankingWrapper) Children() []Expr  { return nil }
func (rw *RankingWrapper) String() string    { return rw.Ranking.String() }
func (rw *RankingWrapper) Equal(n Expr) bool { return false }
func (rw *RankingWrapper) Sexp() NodeKey {
	return NodeKey("(RankingWrapper ranking:" + rw.Ranking.String() + ")")
}
func (rw *RankingWrapper) Args() []Node           { return nil }
func (rw *RankingWrapper) Clone(args []Node) Node { return rw }

// Note: ConcatActions, AppendToAction, HasCode are defined in helpers.go
