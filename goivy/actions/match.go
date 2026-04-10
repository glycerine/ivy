// match_annotation: Annotation-guided recursive action decomposition.
// Walks action/annotation pairs, calling handler on matching assert/assume actions.
// Essential for proof/counterexample extraction.
//
// This is a port of Python's match_annotation in ivy_actions.py.
package actions

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/xtracer"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// AnnotationHandler is called by MatchAnnotation to process actions during trace reconstruction.
// In Python, eval receives lg.Const and env maps lg.Const → lg.Const using structural
// equality (__hash__/__eq__). In Go, eval receives lg.Expr and env uses lg.NodeKey
// (Sexp-based structural identity) as map keys, matching Python's behavior.
type AnnotationHandler interface {
	// Eval evaluates a condition (an lg.Expr, typically *lg.Const) and returns true/false.
	Eval(cond lg.Expr) bool
	// Handle processes an action with the given environment mapping.
	// env maps lg.NodeKey → lg.Expr using structural identity.
	Handle(action Action, env map[lg.NodeKey]lg.Expr)
	// DoReturn processes a return action.
	DoReturn(action Action, env map[lg.NodeKey]lg.Expr)
	// Fail marks a failure point in the trace.
	Fail()
}

// MatchAnnotation walks an action/annotation pair, calling handler methods
// to reconstruct an execution trace from a satisfying assignment.
// Corresponds to Python's match_annotation.
// The mod parameter provides the module for resolving callee actions in CallAction.
// If mod is nil, CallAction will not be inlined.
func MatchAnnotation(action Action, annot Annotation, handler AnnotationHandler, mod *module.Module) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*AnnotationError); ok {
				fmt.Println("internal error: cannot convert satisfying assignment to program trace")
			} else {
				panic(r)
			}
		}
	}()
	matchAnnotationRecur(action, annot, make(map[lg.NodeKey]lg.Expr), handler, -1, mod)
}

// matchAnnotationRecur is the recursive core of MatchAnnotation.
// pos is the position within a Sequence (-1 means use full length).
func matchAnnotationRecur(action Action, annot Annotation, env map[lg.NodeKey]lg.Expr, handler AnnotationHandler, pos int, mod *module.Module) {
	// Handle RenameAnnotation: update env and recurse
	// Python: for x,y in annot.map.items():
	//             if x in env: save[x] = env[x]
	//             env[x] = env.get(y,y)
	// Keys and lookups use lg.NodeKey (structural equality via Sexp).
	if ra, ok := annot.(*RenameAnnotation); ok {
		save := make(map[lg.NodeKey]lg.Expr)
		for x, y := range ra.Map {
			if old, exists := env[x]; exists {
				save[x] = old
			}
			// Python: env[x] = env.get(y, y) — look up y by structural key, default to y itself
			yKey := lg.Key(y)
			if mapped, exists := env[yKey]; exists {
				env[x] = mapped
			} else {
				env[x] = y
			}
		}
		matchAnnotationRecur(action, ra.Arg, env, handler, pos, mod)
		// Restore saved env
		for x := range ra.Map {
			delete(env, x)
		}
		for x, v := range save {
			env[x] = v
		}
		return
	}

	// Handle Sequence
	if seq, ok := action.(*Sequence); ok {
		if pos < 0 {
			pos = len(seq.Elems)
		}
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
		matchAnnotationRecur(action, compose.Args[0], env, handler, pos-1, mod)
		childAction := extractActionFromNode(seq.Elems[pos-1])
		if childAction != nil {
			matchAnnotationRecur(childAction, compose.Args[1], env, handler, -1, mod)
		}
		return
	}

	// Handle IfAction
	if ifAct, ok := action.(*IfAction); ok {
		ite, ok := annot.(*IteAnnotation)
		if !ok {
			fmt.Println("annotation error: IfAction should have IteAnnotation")
			return
		}
		rncond := envGetExpr(env, ite.Cond)
		cond := handler.Eval(rncond)
		if cond {
			thenAction := extractActionFromNode(ifAct.ThenBody)
			if thenAction != nil {
				matchAnnotationRecur(thenAction, ite.ThenB, env, handler, -1, mod)
			}
		} else {
			if ifAct.ElseBody != nil {
				elseAction := extractActionFromNode(ifAct.ElseBody)
				if elseAction != nil {
					matchAnnotationRecur(elseAction, ite.ElseB, env, handler, -1, mod)
				}
			}
		}
		return
	}

	// Handle ChoiceAction
	if choice, ok := action.(*ChoiceAction); ok {
		ite, ok := annot.(*IteAnnotation)
		if !ok {
			fmt.Println("annotation error: ChoiceAction should have IteAnnotation")
			return
		}

		// Handle EnvAction with label
		if envAct, ok := action.(*EnvAction); ok {
			if len(envAct.Labels) > 0 {
				handler.Handle(envAct, env)
			}
		}

		annots := UniteAnnot(ite)
		if len(annots) != len(choice.Branches) {
			fmt.Printf("annotation error: %d annots but %d branches\n", len(annots), len(choice.Branches))
			return
		}

		// Walk branches in reverse order, pick the first whose condition is true
		for i := len(choice.Branches) - 1; i >= 0; i-- {
			rncond := envGetExpr(env, annots[i].Cond)
			if handler.Eval(rncond) {
				branchAction := extractActionFromNode(choice.Branches[i])

				// Handle EnvAction without label
				if envAct, ok := action.(*EnvAction); ok {
					if len(envAct.Labels) == 0 && branchAction != nil {
						callAct := &EnvAction{}
						callAct.Labels = []string{"call"}
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
	//         seq = Sequence(IgnoreAction(), callee, ReturnAction())
	//         recur(seq, annot, env, None)
	if callAct, ok := action.(*CallAction); ok {
		handler.Handle(action, env)
		if mod != nil {
			calleeName := callAct.CalleeName()
			if calleeIface, ok := mod.Actions.Get2(calleeName); ok {
				if callee, ok := calleeIface.(Action); ok {
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
	if whileAct, ok := action.(*WhileAction); ok {
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
		handler.DoReturn(action, env)
		return
	}

	// Handle IgnoreAction
	if _, ok := action.(*IgnoreAction); ok {
		return
	}

	// Handle LocalAction
	if local, ok := action.(*LocalAction); ok {
		bodyAction := extractActionFromNode(local.Body)
		if bodyAction != nil {
			matchAnnotationRecur(bodyAction, annot, env, handler, -1, mod)
		}
		return
	}

	// Default: handle the action directly
	handler.Handle(action, env)
}

// AnnotBranch represents one branch of a flattened IteAnnotation.
type AnnotBranch struct {
	Cond lg.Expr
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
			if mapped, ok := a.Map[lg.Key(cond)]; ok {
				cond = mapped
			}
			result[i] = AnnotBranch{
				Cond: cond,
				Ann:  &RenameAnnotation{Arg: b.Ann, Map: a.Map},
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
func envGetExpr(env map[lg.NodeKey]lg.Expr, cond lg.Expr) lg.Expr {
	if v, ok := env[lg.Key(cond)]; ok {
		return v
	}
	return cond
}

// extractActionFromNode tries to extract an Action from a lg.Expr.
func extractActionFromNode(n interface{}) Action {
	if n == nil {
		return nil
	}
	if act, ok := n.(Action); ok {
		return act
	}
	// Try wrapper types (e.g. TacticNodeWrapper)
	type wrapper interface {
		GetAction() Action
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
//   def expand(self, domain, pvars):
//       modset, pre, post = self.args[1].int_update(domain, pvars)
//       if isinstance(self.args[-1], Ranking):
//           asserts = self.args[2:-1]
//           decreases = self.args[-1]
//       else:
//           asserts = self.args[2:]
//           decreases = None
//       assumes = [a.assert_to_assume([AssertAction]) for a in asserts
//                  if not isinstance(a, SubgoalAction)]
//       asserts = [a for a in asserts if not isinstance(a, AssumeAction)]
//       entry_asserts = []
//       exit_asserts = []
//       if decreases is not None:
//           rank = decreases.args[0]
//           aux = Symbol('$rank', rank.sort)
//           assumes.append(AssumeAction(Equals(aux, rank)))
//           ltsym = Symbol('<', RelationSort([rank.sort, rank.sort]))
//           exit_asserts.append(AssertAction(ltsym(rank, aux)))
//           entry_asserts.append(AssertAction(Not(ltsym(rank, Symbol('0', rank.sort)))))
//       havocs = [HavocAction(sym) for sym in modset]
//       res = Sequence(*(
//           asserts + havocs + assumes +
//           [IfAction(self.args[0],
//               Sequence(*(entry_asserts + [self.args[1]] + exit_asserts + asserts + [AssumeAction(Or())])),
//               Sequence())]))
//       if decreases is not None:
//           res = LocalAction(aux, res)
//       return res
func expandWhile(w *WhileAction, mod *module.Module) Action {
	// Step 1: compute the modset by getting int_update of the body.
	// We need the modified set to generate havocs.
	var modset []*lg.Const
	if mod != nil {
		bodyAction := extractActionFromNode(w.Body)
		if bodyAction != nil {
			ctx := &UpdateContext{
				Domain:       mod,
				PVars:        make(map[string]bool),
				ActCfg:       mod.Cfg.ActCfg,
				Instantiator: mod.Instantiator,
				GetAction: func(name string) Action {
					if mod.Actions != nil {
						if v, ok := mod.Actions.Get2(name); ok {
							if act, ok := v.(Action); ok {
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
	var asserts []Action // invariant assertions
	var decreasesRanking *Ranking

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
	var assumes []Action
	for _, a := range asserts {
		if _, isSub := a.(*SubgoalAction); isSub {
			continue
		}
		assumes = append(assumes, AssertToAssume(a, assertKinds))
	}

	// Filter asserts: remove any that became AssumeActions
	var filteredAsserts []Action
	for _, a := range asserts {
		if _, isAssume := a.(*AssumeAction); isAssume {
			continue
		}
		filteredAsserts = append(filteredAsserts, a)
	}
	asserts = filteredAsserts

	// Step 4: Handle ranking/decreases.
	var entryAsserts []Action
	var exitAsserts []Action
	var auxVar lg.Expr // for LocalAction wrapper

	if decreasesRanking != nil && len(decreasesRanking.RArgs) > 0 {
		rank := decreasesRanking.RArgs[0]
		rankSort := rank.NodeSort()

		// aux = Symbol('$rank', rank.sort)
		aux := lg.NewConst("$rank", rankSort)
		auxVar = aux

		// assumes.append(AssumeAction(Equals(aux, rank)))
		eqFmla := &lg.Eq{T1: aux, T2: rank}
		assumeEq := NewAssumeAction(eqFmla)
		assumeEq.SetLineno(w.GetLineno())
		assumes = append(assumes, assumeEq)

		// ltsym = Symbol('<', RelationSort([rank.sort, rank.sort]))
		ltSort := &lg.FunctionSort{Sorts: []lg.Sort{rankSort, rankSort, lg.Boolean}}
		ltSym := lg.NewConst("<", ltSort)

		// exit_asserts.append(AssertAction(ltsym(rank, aux)))
		ltApp := lg.MustApply(ltSym, rank, aux)
		exitAssert := NewAssertAction(ltApp)
		exitAssert.SetLineno(w.GetLineno())
		exitAsserts = append(exitAsserts, exitAssert)

		// entry_asserts.append(AssertAction(Not(ltsym(rank, Symbol('0', rank.sort)))))
		zeroSym := lg.NewConst("0", rankSort)
		ltZero := lg.MustApply(ltSym, rank, zeroSym)
		entryAssert := NewAssertAction(&lg.Not{Body: ltZero})
		entryAssert.SetLineno(w.GetLineno())
		entryAsserts = append(entryAsserts, entryAssert)
	}

	// Step 5: Build havocs for modified symbols.
	var havocs []Action
	if mod != nil {
		for _, modSym := range modset {
			if sym, ok := mod.Sig.Symbols[modSym.Name]; ok {
				havocTarget := lg.NewConst(modSym.Name, sym.Sort)
				h := NewHavocAction(havocTarget)
				h.SetLineno(w.GetLineno())
				havocs = append(havocs, h)
			}
		}
	}

	// Step 6: Build the result Sequence.
	// Python:
	// res = Sequence(*(
	//     asserts + havocs + assumes +
	//     [IfAction(self.args[0],
	//         Sequence(*(entry_asserts + [self.args[1]] + exit_asserts + asserts + [AssumeAction(Or())])),
	//         Sequence())]))

	// Build the then-branch of the IfAction:
	// Sequence(entry_asserts + [body] + exit_asserts + asserts + [AssumeAction(Or())])
	var thenParts []lg.Expr
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
	assumeFalse := NewAssumeAction(&lg.Or{Terms: nil})
	assumeFalse.SetLineno(w.GetLineno())
	thenParts = append(thenParts, assumeFalse)

	thenSeq := NewSequence(thenParts...)
	thenSeq.SetLineno(w.GetLineno())
	elseSeq := NewSequence() // empty Sequence
	elseSeq.SetLineno(w.GetLineno())

	ifAction := NewIfAction(w.Cond, thenSeq, elseSeq)
	ifAction.SetLineno(w.GetLineno())

	// Build the outer Sequence: asserts + havocs + assumes + [ifAction]
	var outerParts []lg.Expr
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
	var res Action = outerSeq

	// Step 7: Wrap in LocalAction if decreases ranking was used.
	// Python: if decreases is not None: res = LocalAction(aux, res)
	if auxVar != nil {
		actCfg := mod.Cfg.ActCfg
		res = NewLocalActionOn(actCfg, "actions.WhileAction.action_update", auxVar, res)
	}

	return res
}

// RankingWrapper wraps a Ranking as a lg.Expr for storage in WhileAction.Invariants.
type RankingWrapper struct {
	ast.Base
	Ranking *Ranking
}

func (rw *RankingWrapper) NodeSort() lg.Sort   { return lg.Boolean }
func (rw *RankingWrapper) Children() []lg.Expr  { return nil }
func (rw *RankingWrapper) String() string        { return rw.Ranking.String() }
func (rw *RankingWrapper) Equal(n lg.Expr) bool { return false }
func (rw *RankingWrapper) Sexp() lg.NodeKey       { return lg.NodeKey("(RankingWrapper ranking:" + rw.Ranking.String() + ")") }
func (rw *RankingWrapper) Args() []ast.Node      { return nil }
func (rw *RankingWrapper) Clone(args []ast.Node) ast.Node { return rw }

// Note: ConcatActions, AppendToAction, HasCode are defined in helpers.go
