// transforms.go implements action transformation functions that correspond to
// Python's assert_to_assume, modifies, references, prefix_calls,
// drop_invariants, and unroll_loops methods on Action classes.
//
// These are implemented as standalone functions rather than interface methods
// to avoid modifying the Action interface and all its implementations.
package actions

import (
	"fmt"

	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
)

// AssertToAssume recursively transforms an action tree, converting
// action nodes whose type name matches one of the given kinds into
// AssumeAction nodes. This is used during isolate extraction to convert
// assertions into assumptions for specification actions.
//
// The kinds map uses action type names as keys: "assert", "require", "ensure".
// This matches Python's assert_to_assume(kinds) which checks type(self) in kinds.
//
// Corresponds to Python's Action.assert_to_assume(kinds).
func AssertToAssume(action Action, kinds map[string]bool, iuCfg ...*iu.IvyUtilsConfig) Action {
	if action == nil {
		return nil
	}

	switch a := action.(type) {
	case *RequiresAction:
		// RequiresAction must be checked before AssertAction since it embeds it
		if kinds["require"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			return assume
		}
		return a

	case *EnsuresAction:
		// EnsuresAction must be checked before AssertAction since it embeds it
		// Python: checks iu.get_numeric_version() <= [1,6] before converting
		if kinds["ensure"] {
			var ver []int
			if len(iuCfg) > 0 && iuCfg[0] != nil {
				ver = iuCfg[0].GetNumericVersion()
			} else {
				panic("AssertToAssume: EnsuresAction requires IvyUtilsConfig for version check")
			}
			if len(ver) >= 2 && (ver[0] < 1 || (ver[0] == 1 && ver[1] <= 6)) {
				assume := NewAssumeAction(a.Formula)
				assume.ActionBase = a.ActionBase
				return assume
			}
		}
		return a

	case *SubgoalAction:
		// SubgoalAction embeds AssertAction — match when "assert" is in kinds
		// Mirrors Python's class hierarchy where SubgoalAction inherits AssertAction
		if kinds["assert"] || kinds["subgoal"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			return assume
		}
		return a

	case *AssertAction:
		// Plain AssertAction (not RequiresAction, EnsuresAction, or SubgoalAction)
		if kinds["assert"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			return assume
		}
		return a

	default:
		// Recursively transform children
		return assertToAssumeChildren(action, kinds, iuCfg...)
	}
}

// assertToAssumeChildren recursively transforms children of an action.
func assertToAssumeChildren(action Action, kinds map[string]bool, iuCfg ...*iu.IvyUtilsConfig) Action {
	args := action.ActionArgs()
	changed := false
	newArgs := make([]lg.Expr, len(args))
	for i, arg := range args {
		if child, ok := arg.(Action); ok {
			newChild := AssertToAssume(child, kinds, iuCfg...)
			if newChild != child {
				changed = true
				newArgs[i] = newChild
			} else {
				newArgs[i] = arg
			}
		} else {
			newArgs[i] = arg
		}
	}
	if changed {
		return action.ActionClone(newArgs)
	}
	return action
}

// Modifies returns the list of symbols modified by an action.
// This matches Python's Action.modifies() which returns [n.rep] — a list
// of Symbol objects. Callers that need a structural-equality set build one
// via: set[lg.Key(sym)] = true. Callers that need the plain name use sym.Name.
// Modifies returns the list of symbols modified by an action.
// Uses the given ActionsConfig for destructor lookups (may be nil).
func Modifies(action Action, cfg ...*ActionsConfig) []*lg.Symbol {
	var acfg *ActionsConfig
	if len(cfg) > 0 {
		acfg = cfg[0]
	}
	var result []*lg.Symbol
	modifiesRec(action, &result, acfg)
	return result
}

func modifiesRec(action Action, result *[]*lg.Symbol, cfg *ActionsConfig) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *AssignAction:
		// Walk destructor chain to find root symbol
		target := a.LHS
		for {
			if app, ok := target.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Symbol); ok {
					if isDestructor(c.Name, cfg) && len(app.Terms) > 0 {
						target = app.Terms[0]
						continue
					}
				}
			}
			break
		}
		if c, ok := target.(*lg.Symbol); ok {
			*result = append(*result, c)
		}

	case *HavocAction:
		if a.Target != nil {
			if c, ok := a.Target.(*lg.Symbol); ok {
				*result = append(*result, c)
			}
		}

	default:
		// Recurse into children
		for _, arg := range action.ActionArgs() {
			if child, ok := arg.(Action); ok {
				modifiesRec(child, result, cfg)
			}
		}
	}
}

func isDestructor(name string, cfg *ActionsConfig) bool {
	// Python: return symbol.name in im.module.destructor_sorts
	if cfg == nil || cfg.Context == nil {
		return false
	}
	mod := cfg.Context.GetDomain()
	if mod != nil {
		_, found := mod.DestructorSorts[name]
		return found
	}
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
	for _, arg := range action.ActionArgs() {
		if child, ok := arg.(Action); ok {
			referencesRec(child, result)
		} else if arg != nil {
			// Collect symbols from non-action nodes
			collectSymbols(arg, result)
		}
	}
}

func collectSymbols(node lg.Expr, result map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Symbol); ok {
		result[c.Name] = true
	}
	if app, ok := node.(*lg.Apply); ok {
		collectSymbols(app.Func, result)
	}
	for _, child := range node.Children() {
		collectSymbols(child, result)
	}
}

// PrefixCalls renames call targets by prepending a prefix.
// Used during isolate composition.
// Corresponds to Python's Action.prefix_calls(pref) when pref is a string.
func PrefixCalls(action Action, prefix string) Action {
	if action == nil || prefix == "" {
		return action
	}
	return PrefixCallsFunc(action, func(name string) string {
		return prefix + name
	})
}

// PrefixCallsFunc renames call targets using a callable renamer.
// The renamer receives the current callee name and returns the new name.
// Python: Action.prefix_calls(pref) when pref is callable.
func PrefixCallsFunc(action Action, renamer func(string) string) Action {
	if action == nil || renamer == nil {
		return action
	}
	switch a := action.(type) {
	case *CallAction:
		if a.Callee != nil {
			if c, ok := a.Callee.(*lg.Symbol); ok {
				newName := renamer(c.Name)
				newConst := lg.NewSymbol(newName, c.CSort)
				newCall := NewCallAction(newConst, a.ActualReturns...)
				newCall.ActionBase = a.ActionBase
				a.ActionBase.CopyFormalsTo(newCall)
				return newCall
			}
		}
		return a
	default:
		args := action.ActionArgs()
		changed := false
		newArgs := make([]lg.Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newChild := PrefixCallsFunc(child, renamer)
				if newChild != child {
					changed = true
					newArgs[i] = newChild
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.ActionClone(newArgs)
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
		args := action.ActionArgs()
		changed := false
		newArgs := make([]lg.Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newChild := DropInvariants(child)
				if newChild != child {
					changed = true
					newArgs[i] = newChild
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.ActionClone(newArgs)
		}
		return action
	}
}

// CardFunc computes the cardinality of a sort for loop unrolling.
// Returns -1 if the cardinality cannot be determined.
// Matches Python's card parameter to unroll_loops/unroll.
type CardFunc func(s lg.Sort) int

// UnrollLoops converts while loops to bounded if-then-else chains.
// The card function determines the iteration bound from the loop's index sort.
// Corresponds to Python's Action.unroll_loops(card) and WhileAction.unroll(card,body).
func UnrollLoops(action Action, card CardFunc) Action {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *WhileAction:
		// Python: WhileAction.unroll_loops first recurses into body,
		// then calls self.unroll(card, body)
		bodyAct, _ := a.Body.(Action)
		if bodyAct != nil {
			bodyAct = UnrollLoops(bodyAct, card)
		}
		return unrollWhile(a, card, bodyAct)
	default:
		args := action.ActionArgs()
		changed := false
		newArgs := make([]lg.Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newChild := UnrollLoops(child, card)
				if newChild != child {
					changed = true
					newArgs[i] = newChild
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.ActionClone(newArgs)
		}
		return action
	}
}

// unrollWhile implements Python's WhileAction.unroll(card, body).
// Examines the condition to determine an index sort, computes cardinality,
// and builds the unrolled if-then-else chain.
func unrollWhile(a *WhileAction, card CardFunc, body Action) Action {
	cond := a.Cond
	// Peel through And to find the comparison (Python lines 1027-1028)
	for {
		if andNode, ok := cond.(*lg.And); ok && len(andNode.Terms) > 0 {
			cond = andNode.Terms[0]
		} else {
			break
		}
	}
	// Determine index sort from condition (Python lines 1029-1033)
	var idxSort lg.Sort
	if app, ok := cond.(*lg.Apply); ok {
		if sym, ok := app.Func.(*lg.Symbol); ok {
			name := sym.Name
			if name == "<" || name == ">" || name == "<=" || name == ">=" {
				if len(app.Terms) > 0 {
					idxSort = app.Terms[0].NodeSort()
				}
			}
		}
	} else if notNode, ok := cond.(*lg.Not); ok {
		if eq, ok := notNode.Body.(*lg.Eq); ok {
			idxSort = eq.T1.NodeSort()
		}
	}
	cardsort := card(idxSort)
	sortName := "unknown sort"
	if idxSort != nil {
		sortName = idxSort.String()
	}
	if cardsort < 0 {
		panic(fmt.Sprintf("cannot determine an iteration bound for loop over %s", sortName))
	}
	if cardsort > 100 {
		panic(fmt.Sprintf("cowardly refusing to unroll loop over %s %d times", sortName, cardsort))
	}
	// Build unrolled if-then-else chain (Python lines 1041-1044)
	// Base case: if cond then AssumeAction(Or()) — equivalent to assume false
	orExpr := &lg.Or{}
	res := NewIfAction(a.Cond, NewAssumeAction(orExpr))
	for i := 0; i < cardsort; i++ {
		var bodyExpr lg.Expr
		if body != nil {
			bodyExpr = body
		} else {
			bodyExpr = a.Body
		}
		seq := NewSequence(bodyExpr, res)
		res = NewIfAction(a.Cond, seq)
	}
	CopyFormalsTo(a, res)
	return res
}

// GetReferencesInto accumulates non-action symbol references from an
// action into the given set. Corresponds to Python's get_references().
func GetReferencesInto(action Action, syms map[string]bool) {
	referencesRec(action, syms)
}

// EraseUnrefed replaces assignments to unreferenced symbols with
// empty sequences. Used for cone-of-influence filtering.
// syms is the set of referenced symbols; names is a set of names
// referenced by proofs that should also be kept.
// Corresponds to Python's Action.erase_unrefed(refs, names).
func EraseUnrefed(action Action, syms map[string]bool, names map[string]bool) Action {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *AssignAction:
		// If LHS symbol is not referenced, erase
		if c, ok := rootSymbol(a.LHS); ok {
			if !syms[c.Name] && !names[c.Name] {
				return NewSequence()
			}
		}
		return a
	case *HavocAction:
		if a.Target != nil {
			if c, ok := a.Target.(*lg.Symbol); ok {
				if !syms[c.Name] && !names[c.Name] {
					return NewSequence()
				}
			}
		}
		return a
	default:
		// Recurse into children
		args := action.ActionArgs()
		changed := false
		newArgs := make([]lg.Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newChild := EraseUnrefed(child, syms, names)
				if newChild != child {
					changed = true
					newArgs[i] = newChild
				} else {
					newArgs[i] = arg
				}
			} else {
				newArgs[i] = arg
			}
		}
		if changed {
			return action.ActionClone(newArgs)
		}
		return action
	}
}

// rootSymbol walks destructor chains to find the root symbol of an assignment LHS.
func rootSymbol(node lg.Expr) (*lg.Symbol, bool) {
	for {
		if app, ok := node.(*lg.Apply); ok && len(app.Terms) > 0 {
			node = app.Terms[0]
			continue
		}
		break
	}
	c, ok := node.(*lg.Symbol)
	return c, ok
}
