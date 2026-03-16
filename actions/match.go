// match_annotation: Annotation-guided recursive action decomposition.
// Walks action/annotation pairs, calling handler on matching assert/assume actions.
// Essential for proof/counterexample extraction.
//
// This is a port of Python's match_annotation in ivy_actions.py.
package actions

import (
	"fmt"
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
func MatchAnnotation(action Action, annot Annotation, handler AnnotationHandler) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*AnnotationError); ok {
				fmt.Println("internal error: cannot convert satisfying assignment to program trace")
			} else {
				panic(r)
			}
		}
	}()
	matchAnnotationRecur(action, annot, make(map[string]string), handler, -1)
}

// matchAnnotationRecur is the recursive core of MatchAnnotation.
// pos is the position within a Sequence (-1 means use full length).
func matchAnnotationRecur(action Action, annot Annotation, env map[string]string, handler AnnotationHandler, pos int) {
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
		matchAnnotationRecur(action, ra.Arg, env, handler, pos)
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
				matchAnnotationRecur(action, ite.ThenB, env, handler, pos)
				return
			}
			matchAnnotationRecur(action, ite.ElseB, env, handler, pos-1)
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
		matchAnnotationRecur(action, compose.Args[0], env, handler, pos-1)
		childAction := extractActionFromNode(seq.Children[pos-1])
		if childAction != nil {
			matchAnnotationRecur(childAction, compose.Args[1], env, handler, -1)
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
				matchAnnotationRecur(thenAction, ite.ThenB, env, handler, -1)
			}
		} else {
			if ifAct.ElseBody != nil {
				elseAction := extractActionFromNode(ifAct.ElseBody)
				if elseAction != nil {
					matchAnnotationRecur(elseAction, ite.ElseB, env, handler, -1)
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
					matchAnnotationRecur(branchAction, annots[i].Ann, env, handler, -1)
				}
				return
			}
		}
		panic(&AnnotationError{Msg: "problem in match_annotation: no branch matched"})
	}

	// Handle CallAction
	if _, ok := action.(*CallAction); ok {
		handler.Handle(action, env)
		// In a full implementation, we'd look up the callee in the module
		// and recurse into it. For now, just handle the call action.
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
			matchAnnotationRecur(bodyAction, annot, env, handler, -1)
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

// extractActionFromNode tries to extract an Action from a lg.Node.
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

// Note: ConcatActions, AppendToAction, HasCode are defined in helpers.go
