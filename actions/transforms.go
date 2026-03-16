// transforms.go implements action transformation functions that correspond to
// Python's assert_to_assume, modifies, references, prefix_calls,
// drop_invariants, and unroll_loops methods on Action classes.
//
// These are implemented as standalone functions rather than interface methods
// to avoid modifying the Action interface and all its implementations.
package actions

import (
	lg "github.com/glycerine/goivy/logic"
)

// AssertToAssume recursively transforms an action tree, converting
// AssertAction nodes whose kind matches one of the given kinds into
// AssumeAction nodes. This is used during isolate extraction to convert
// assertions into assumptions for specification actions.
//
// Corresponds to Python's Action.assert_to_assume(kinds).
func AssertToAssume(action Action, kinds map[string]bool) Action {
	if action == nil {
		return nil
	}

	switch a := action.(type) {
	case *AssertAction:
		// Check if this assert's kind matches
		kind := ""
		if a.Kind != "" {
			kind = a.Kind
		}
		if kinds[kind] || kinds[""] {
			// Convert to assume
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			return assume
		}
		return a

	case *EnsureAction:
		// EnsureAction has version-dependent logic
		// For simplicity, convert to assume if kinds contains "ensure"
		if kinds["ensure"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			return assume
		}
		return a

	default:
		// Recursively transform children
		args := action.Args()
		changed := false
		newArgs := make([]lg.Node, len(args))
		for i, arg := range args {
			if child := UnwrapAction(arg); child != nil {
				newChild := AssertToAssume(child, kinds)
				if newChild != child {
					changed = true
					newArgs[i] = WrapAction(newChild)
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.Clone(newArgs)
		}
		return action
	}
}

// Modifies returns the set of symbol names modified by an action.
// Corresponds to Python's Action.modifies().
func Modifies(action Action) map[string]bool {
	result := make(map[string]bool)
	modifiesRec(action, result)
	return result
}

func modifiesRec(action Action, result map[string]bool) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *AssignAction:
		// Walk destructor chain to find root symbol
		target := a.LHS
		for {
			if app, ok := target.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					if isDestructor(c.Name) && len(app.Terms) > 0 {
						target = app.Terms[0]
						continue
					}
				}
			}
			break
		}
		if c, ok := target.(*lg.Const); ok {
			result[c.Name] = true
		}

	case *HavocAction:
		if a.Target != nil {
			if c, ok := a.Target.(*lg.Const); ok {
				result[c.Name] = true
			}
		}

	default:
		// Recurse into children
		for _, arg := range action.Args() {
			if child := UnwrapAction(arg); child != nil {
				modifiesRec(child, result)
			}
		}
	}
}

func isDestructor(name string) bool {
	// A destructor in Ivy is typically a field accessor
	// This is a simplified check
	return false
}

// References returns the set of non-action symbols referenced by an action.
// Corresponds to Python's Action.references().
func References(action Action) map[string]bool {
	result := make(map[string]bool)
	referencesRec(action, result)
	return result
}

func referencesRec(action Action, result map[string]bool) {
	if action == nil {
		return
	}
	for _, arg := range action.Args() {
		if child := UnwrapAction(arg); child != nil {
			referencesRec(child, result)
		} else if arg != nil {
			// Collect symbols from non-action nodes
			collectSymbols(arg, result)
		}
	}
}

func collectSymbols(node lg.Node, result map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Const); ok {
		result[c.Name] = true
	}
	for _, child := range node.Children() {
		collectSymbols(child, result)
	}
}

// PrefixCalls renames call targets by prepending a prefix.
// Used during isolate composition.
// Corresponds to Python's Action.prefix_calls(pref).
func PrefixCalls(action Action, prefix string) Action {
	if action == nil || prefix == "" {
		return action
	}
	switch a := action.(type) {
	case *CallAction:
		if a.Callee != nil {
			if c, ok := a.Callee.(*lg.Const); ok {
				newName := prefix + c.Name
				newConst := lg.NewConst(newName, c.CSort)
				newCall := NewCallAction(newConst)
				newCall.ActionBase = a.ActionBase
				return newCall
			}
		}
		return a
	default:
		args := action.Args()
		changed := false
		newArgs := make([]lg.Node, len(args))
		for i, arg := range args {
			if child := UnwrapAction(arg); child != nil {
				newChild := PrefixCalls(child, prefix)
				if newChild != child {
					changed = true
					newArgs[i] = WrapAction(newChild)
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.Clone(newArgs)
		}
		return action
	}
}

// DropInvariants strips loop invariants from while loops.
// Corresponds to Python's Action.drop_invariants().
func DropInvariants(action Action) Action {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *WhileAction:
		// Remove invariant by setting it to nil
		newWhile := &WhileAction{
			ActionBase: a.ActionBase,
		}
		newWhile.Cond = a.Cond
		newWhile.Body = a.Body
		// Don't copy invariant
		return newWhile
	default:
		args := action.Args()
		changed := false
		newArgs := make([]lg.Node, len(args))
		for i, arg := range args {
			if child := UnwrapAction(arg); child != nil {
				newChild := DropInvariants(child)
				if newChild != child {
					changed = true
					newArgs[i] = WrapAction(newChild)
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.Clone(newArgs)
		}
		return action
	}
}

// UnrollLoops converts while loops to bounded if-then-else chains.
// The bound parameter controls how many times to unroll.
// Corresponds to Python's Action.unroll_loops(bound).
func UnrollLoops(action Action, bound int) Action {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *WhileAction:
		if bound <= 0 {
			return NewSequence()
		}
		// Unroll: if cond then { body; unroll(n-1) } else { skip }
		body := a.Body
		if bodyAct := UnwrapAction(body); bodyAct != nil {
			bodyAct = UnrollLoops(bodyAct, bound)
			// Build: if cond then { body; unroll(bound-1) }
			innerUnroll := UnrollLoops(a, bound-1)
			seq := NewSequence(WrapAction(bodyAct), WrapAction(innerUnroll))
			result := NewIfAction(a.Cond, WrapAction(seq))
			result.ActionBase = a.ActionBase
			return result
		}
		return NewSequence()
	default:
		args := action.Args()
		changed := false
		newArgs := make([]lg.Node, len(args))
		for i, arg := range args {
			if child := UnwrapAction(arg); child != nil {
				newChild := UnrollLoops(child, bound)
				if newChild != child {
					changed = true
					newArgs[i] = WrapAction(newChild)
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.Clone(newArgs)
		}
		return action
	}
}

// EraseUnrefed replaces unreferenced actions with empty sequences.
// Used for cone-of-influence filtering.
// Corresponds to Python's erase_unrefed(action, refs).
func EraseUnrefed(action Action, refs map[string]bool) Action {
	if action == nil {
		return nil
	}
	// Check if this action references any symbol in refs
	actionRefs := References(action)
	hasRef := false
	for sym := range actionRefs {
		if refs[sym] {
			hasRef = true
			break
		}
	}
	if !hasRef {
		return NewSequence()
	}
	return action
}
