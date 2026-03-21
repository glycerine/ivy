// match_annotation: Annotation-guided recursive action decomposition.
// Walks action/annotation pairs, calling handler on matching assert/assume actions.
// Essential for proof/counterexample extraction.
//
// This is a port of Python's match_annotation in ivy_actions.py.
package actions

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// AnnotationHandler is called by MatchAnnotation to process actions during trace reconstruction.
type AnnotationHandler interface {
	// Eval evaluates a condition (given as a string symbol name) and returns true/false.
	Eval(cond string) bool
	// Handle processes an action with the given environment mapping.
	Handle(action Action, env map[string]string)
	// DoReturn processes a return action.
	DoReturn(action Action, env map[string]string)
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
	matchAnnotationRecur(action, annot, make(map[string]string), handler, -1, mod)
}

// matchAnnotationRecur is the recursive core of MatchAnnotation.
// pos is the position within a Sequence (-1 means use full length).
func matchAnnotationRecur(action Action, annot Annotation, env map[string]string, handler AnnotationHandler, pos int, mod *module.Module) {
	// Handle RenameAnnotation: update env and recurse
	if ra, ok := annot.(*RenameAnnotation); ok {
		save := make(map[string]string)
		for x, y := range ra.Map {
			if old, exists := env[x]; exists {
				save[x] = old
			}
			if mapped, exists := env[y]; exists {
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
			pos = len(seq.Children)
		}
		if pos == 0 {
			if _, ok := annot.(EmptyAnnotation); !ok {
				fmt.Println("annotation error: should be empty annotation")
			}
			return
		}

		// Check for IteAnnotation (branch failure point)
		if ite, ok := annot.(*IteAnnotation); ok {
			rncond := envGet(env, ite.Cond)
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
		childAction := extractActionFromNode(seq.Children[pos-1])
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
		rncond := envGet(env, ite.Cond)
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
			rncond := envGet(env, annots[i].Cond)
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
			if calleeIface, ok := mod.Actions[calleeName]; ok {
				if callee, ok := calleeIface.(Action); ok {
					// Build: Sequence(IgnoreAction(), callee, ReturnAction())
					seq := NewSequence(
						WrapAction(&IgnoreAction{}),
						WrapAction(callee),
						WrapAction(&ReturnAction{}),
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
	Cond string
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
			if mapped, ok := a.Map[cond]; ok {
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

// envGet looks up a key in the environment, returning the mapped value
// or the key itself if not found.
func envGet(env map[string]string, key string) string {
	if v, ok := env[key]; ok {
		return v
	}
	return key
}

// extractActionFromNode tries to extract an Action from a lg.Expr.
func extractActionFromNode(n interface{}) Action {
	if n == nil {
		return nil
	}
	if act, ok := n.(Action); ok {
		return act
	}
	// Try ActionNodeWrapper
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
	var modset []*lg.Symbol
	if mod != nil {
		bodyAction := extractActionFromNode(w.Body)
		if bodyAction != nil {
			ctx := &UpdateContext{
				Domain: mod,
				PVars:  make(map[string]bool),
				GetAction: func(name string) Action {
					if mod.Actions != nil {
						if v, ok := mod.Actions[name]; ok {
							if act, ok := v.(Action); ok {
								return act
							}
						}
					}
					return nil
				},
			}
			update := IntUpdate(bodyAction, ctx)
			if update != nil && update.Modified != nil {
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
		aux := lg.NewSymbol("$rank", rankSort)
		auxVar = aux

		// assumes.append(AssumeAction(Equals(aux, rank)))
		eqFmla := &lg.Eq{T1: aux, T2: rank}
		assumeEq := NewAssumeAction(eqFmla)
		assumeEq.SetLineno(w.GetLineno())
		assumes = append(assumes, assumeEq)

		// ltsym = Symbol('<', RelationSort([rank.sort, rank.sort]))
		ltSort := &lg.FunctionSort{Sorts: []lg.Sort{rankSort, rankSort, lg.Boolean}}
		ltSym := lg.NewSymbol("<", ltSort)

		// exit_asserts.append(AssertAction(ltsym(rank, aux)))
		ltApp := &lg.Apply{Func: ltSym, Terms: []lg.Expr{rank, aux}}
		exitAssert := NewAssertAction(ltApp)
		exitAssert.SetLineno(w.GetLineno())
		exitAsserts = append(exitAsserts, exitAssert)

		// entry_asserts.append(AssertAction(Not(ltsym(rank, Symbol('0', rank.sort)))))
		zeroSym := lg.NewSymbol("0", rankSort)
		ltZero := &lg.Apply{Func: ltSym, Terms: []lg.Expr{rank, zeroSym}}
		entryAssert := NewAssertAction(&lg.Not{Body: ltZero})
		entryAssert.SetLineno(w.GetLineno())
		entryAsserts = append(entryAsserts, entryAssert)
	}

	// Step 5: Build havocs for modified symbols.
	var havocs []Action
	if mod != nil {
		for _, modSym := range modset {
			if sym, ok := mod.Sig.Symbols[modSym.Name]; ok {
				havocTarget := lg.NewSymbol(modSym.Name, sym.Sort)
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
		thenParts = append(thenParts, WrapAction(ea))
	}
	thenParts = append(thenParts, w.Body) // the loop body
	for _, xa := range exitAsserts {
		thenParts = append(thenParts, WrapAction(xa))
	}
	for _, a := range asserts {
		thenParts = append(thenParts, WrapAction(a))
	}
	// AssumeAction(Or()) = assume false (empty disjunction)
	assumeFalse := NewAssumeAction(&lg.Or{Terms: nil})
	thenParts = append(thenParts, WrapAction(assumeFalse))

	thenSeq := NewSequence(thenParts...)
	elseSeq := NewSequence() // empty Sequence

	ifAction := NewIfAction(w.Cond, WrapAction(thenSeq), WrapAction(elseSeq))

	// Build the outer Sequence: asserts + havocs + assumes + [ifAction]
	var outerParts []lg.Expr
	for _, a := range asserts {
		outerParts = append(outerParts, WrapAction(a))
	}
	for _, h := range havocs {
		outerParts = append(outerParts, WrapAction(h))
	}
	for _, a := range assumes {
		outerParts = append(outerParts, WrapAction(a))
	}
	outerParts = append(outerParts, WrapAction(ifAction))

	var res Action = NewSequence(outerParts...)

	// Step 7: Wrap in LocalAction if decreases ranking was used.
	// Python: if decreases is not None: res = LocalAction(aux, res)
	if auxVar != nil {
		res = NewLocalAction(auxVar, WrapAction(res))
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
func (rw *RankingWrapper) Sexp() string          { return "(RankingWrapper ranking:" + rw.Ranking.String() + ")" }
func (rw *RankingWrapper) Args() []ast.Node      { return nil }
func (rw *RankingWrapper) Clone(args []ast.Node) ast.Node { return rw }

// Note: ConcatActions, AppendToAction, HasCode are defined in helpers.go
