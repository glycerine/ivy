// transforms.go implements action transformation functions that correspond to
// Python's assert_to_assume, modifies, references, prefix_calls,
// drop_invariants, and unroll_loops methods on Action classes.
//
// These are implemented as standalone functions rather than interface methods
// to avoid modifying the Action interface and all its implementations.
package actions

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/xtracer"
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
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
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
				assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
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
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
			return assume
		}
		return a

	case *AssertAction:
		// Plain AssertAction (not RequiresAction, EnsuresAction, or SubgoalAction)
		if kinds["assert"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
			return assume
		}
		return a

	default:
		// Recursively transform children
		return assertToAssumeChildren(action, kinds, iuCfg...)
	}
}

// assertToAssumeChildren recursively transforms children of an action.
// Python: Action.assert_to_assume ALWAYS clones via self.clone(args).
func assertToAssumeChildren(action Action, kinds map[string]bool, iuCfg ...*iu.IvyUtilsConfig) Action {
	args := action.Args()
	newArgs := make([]ast.Node, len(args))
	for i, arg := range args {
		if child, ok := arg.(Action); ok {
			newArgs[i] = AssertToAssume(child, kinds, iuCfg...)
		} else {
			newArgs[i] = arg // LF, formulas pass through unchanged
		}
	}
	// Python ALWAYS clones — no "changed" optimization.
	res := action.Clone(newArgs).(Action)
	CopyFormalsTo(action, res)
	return res
}

// Modifies returns the list of symbols modified by an action.
// This matches Python's Action.modifies() which returns [n.rep] — a list
// of Symbol objects. Callers that need a structural-equality set build one
// via: set[lg.Key(sym)] = true. Callers that need the plain name use sym.Name.
// Modifies returns the list of symbols modified by an action.
// Uses the given ActionsConfig for destructor lookups (may be nil).
func Modifies(action Action, cfg ...*ActionsConfig) []*lg.Const {
	var acfg *ActionsConfig
	if len(cfg) > 0 {
		acfg = cfg[0]
	}
	var result []*lg.Const
	modifiesRec(action, &result, acfg)
	return result
}

func modifiesRec(action Action, result *[]*lg.Const, cfg *ActionsConfig) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *AssignAction:
		// Walk destructor chain to find root symbol.
		// Python: n = self.args[0]; while n.rep.name in destructor_sorts: n = n.args[0]; return [n.rep]
		// Python accesses n.rep (the Func/rep of an Apply), not n itself.
		// So when the chain ends on an Apply whose Func is not a destructor,
		// we return Func (the symbol), not the Apply node.
		target := a.LHS
		for {
			if app, ok := target.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					if isDestructor(c.Name, cfg) && len(app.Terms) > 0 {
						target = app.Terms[0]
						continue
					}
				}
			}
			break
		}
		if c, ok := target.(*lg.Const); ok {
			*result = append(*result, c)
		} else if app, ok := target.(*lg.Apply); ok {
			// Python: return [n.rep] — n is an Apply, n.rep is its Func symbol
			if c, ok := app.Func.(*lg.Const); ok {
				*result = append(*result, c)
			}
		}

	case *HavocAction:
		// Walk destructor chain to find root symbol, same as AssignAction.
		// Python: while n.rep.name in ivy_module.module.destructor_sorts: n = n.args[0]; return [n.rep]
		if a.Target != nil {
			target := a.Target
			for {
				if app, ok := target.(*lg.Apply); ok {
					if c, ok := app.Func.(*lg.Const); ok {
						if isDestructor(c.Name, cfg) && len(app.Terms) > 0 {
							target = app.Terms[0]
							continue
						}
					}
				}
				break
			}
			if c, ok := target.(*lg.Const); ok {
				*result = append(*result, c)
			} else if app, ok := target.(*lg.Apply); ok {
				// Python: return [n.rep] — n is an Apply, n.rep is its Func symbol
				if c, ok := app.Func.(*lg.Const); ok {
					*result = append(*result, c)
				}
			}
		}

	case *CrashAction:
		// Python: CrashAction.modifies() walks domain.hierarchy to find all
		// non-polymorphic, non-interpreted symbols under the target.
		xtracer.Trace("actions.CrashAction.modifies ENTER target_type=%T", a.Target)
		if cfg != nil && cfg.Context != nil {
			mod := cfg.Context.GetDomain()
			if mod != nil && a.Target != nil {
				// Get the name from the target's rep symbol
				var targetName string
				if app, ok := a.Target.(*lg.Apply); ok {
					if sym, ok := app.Func.(*lg.Const); ok {
						targetName = sym.Name
					}
					xtracer.Trace("actions.CrashAction.modifies ENTER target=Apply func_type=%T targetName=%s", app.Func, targetName)
				} else if sym, ok := a.Target.(*lg.Const); ok {
					targetName = sym.Name
					xtracer.Trace("actions.CrashAction.modifies ENTER target=Symbol targetName=%s", targetName)
				}
				if targetName != "" {
					// Build set of defined names
					// Python: dfnd = [ldf.formula.defines().name for ldf in domain.definitions]
					dfnd := make(map[string]bool)
					for _, ldf := range mod.Definitions {
						if def, ok := ldf.Formula.(*lg.Definition); ok {
							if sym, ok := def.Defines().(*lg.Const); ok {
								dfnd[sym.Name] = true
							}
						}
					}
					// Recurse through hierarchy
					beforeLen := len(*result)
					crashModifiesRec(mod, targetName, dfnd, result)
					xtracer.Trace("actions.CrashAction.modifies EXIT n_syms=%d", len(*result)-beforeLen)
				}
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

// crashModifiesRec walks the module hierarchy to find all symbols modified
// by a CrashAction. Corresponds to Python's CrashAction.modifies() inner recur().
func crashModifiesRec(mod *module.Module, n string, dfnd map[string]bool, result *[]*lg.Const) {
	children, inHier := mod.Hierarchy[n]
	xtracer.Trace("actions.CrashAction.modifies.recur n_type=string n_val=%s in_hierarchy=%v", n, inHier)
	if inHier {
		for child := range children {
			cname := n + "." + child
			if mod.Cfg != nil && mod.Cfg.IuCfg != nil {
				cname = mod.Cfg.IuCfg.ComposeNames(n, child)
			}
			specName := cname + ".spec"
			if mod.Cfg != nil && mod.Cfg.IuCfg != nil {
				specName = mod.Cfg.IuCfg.ComposeNames(cname, "spec")
			}
			_, hasSpec := mod.Attributes[specName]
			skipSpec := child == "spec" || hasSpec
			xtracer.Trace("actions.CrashAction.modifies.recur child=%s cname=%s skip=%v", child, cname, skipSpec)
			if child != "spec" {
				if !hasSpec {
					crashModifiesRec(mod, cname, dfnd, result)
				}
			}
		}
	} else {
		// Leaf: check if it's in the signature and not defined
		inSig := false
		if mod.Sig != nil {
			_, inSig = mod.Sig.Symbols[n]
		}
		inDfnd := dfnd[n]
		xtracer.Trace("actions.CrashAction.modifies.recur LEAF n=%s in_sig=%v in_dfnd=%v", n, inSig, inDfnd)
		if mod.Sig != nil && inSig && !inDfnd {
			for _, sym := range mod.Sig.AllSymbolsNamed(n) {
				isPoly := il.SymbolIsPolymorphic(sym.Name)
				isInterp := il.IsInterpretedSymbol(mod.Sig, sym)
				xtracer.Trace("actions.CrashAction.modifies.recur SYM name=%s poly=%v interp=%v", sym.Name, isPoly, isInterp)
				if !isPoly && !isInterp {
					*result = append(*result, sym)
				}
			}
		}
	}
}

// References returns the set of non-action symbols referenced by an action.
// Corresponds to Python's Action.references().
func References(action Action) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	referencesRec(action, result)
	return result
}

func referencesRec(action Action, result map[lg.NodeKey]lg.Expr) {
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

// ConstSymKey returns a structural identity key for a Const, suitable for
// use as a map key in symbol sets. This matches Python's set behavior where
// Const objects are distinguished by recstruct.__eq__ (compares all fields:
// name + sort). We use Sexp() which encodes full structural identity.
func ConstSymKey(c *lg.Const) lg.NodeKey {
	return c.Sexp()
}

// ConstSymDisplay returns the display string matching Python's str(const),
// which is monkey-patched by ivy_logic.py to show "name:sortname" for numerals,
// or just "name" for non-numerals.
func ConstSymDisplay(c *lg.Const) string {
	if il.IsNumeralName(c.Name) && c.CSort != nil {
		if _, isTop := c.CSort.(*lg.TopSort); !isTop {
			return c.Name + ":" + c.CSort.String()
		}
	}
	return c.Name
}

// SymKeyToDisplay converts a Sexp-based symbol key to Python's str() format.
// Extracts name from "(Symbol name:X sort:Y)" and applies numeral display.
func SymKeyToDisplay(key string) string {
	// Parse name from Sexp format: (Symbol name:FOO sort:...)
	const prefix = "(Symbol name:"
	if !strings.HasPrefix(key, prefix) {
		return key // not a Sexp key, return as-is
	}
	rest := key[len(prefix):]
	// Find " sort:" separator
	sortIdx := strings.Index(rest, " sort:")
	if sortIdx < 0 {
		return key
	}
	name := rest[:sortIdx]
	// For numerals, extract sort name and show name:sortname
	if il.IsNumeralName(name) {
		sortPart := rest[sortIdx+len(" sort:"):]
		// Extract sort name from e.g. "(UninterpretedSort name:index)"
		const usPrefix = "(UninterpretedSort name:"
		if strings.HasPrefix(sortPart, usPrefix) {
			sortName := sortPart[len(usPrefix) : len(sortPart)-2] // strip trailing ")"
			return name + ":" + sortName
		}
	}
	return name
}

func collectSymbols(node lg.Expr, result map[lg.NodeKey]lg.Expr) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Const); ok {
		result[ConstSymKey(c)] = c
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
	xtracer.Trace("actions.prefix_calls ENTER type=%s", shortTypeName(action))
	switch a := action.(type) {
	case *CallAction:
		// Python: CallAction.prefix_calls always creates new CallAction
		// with self.args[0].rename(pref(self.args[0].rep)).
		if a.AstCallee != nil {
			newAtom := a.AstCallee.Rename(renamer(a.AstCallee.Rep))
			newCallee := calleeFromAtom(newAtom)
			if a.ActCfg == nil {
				panic("we should have a.ActCfg set!")
			}
			newCall := NewCallActionOn(a.ActCfg, newCallee, a.ActualReturns...)
			newCall.ActionBase = a.ActionBase
			newCall.AstCallee = newAtom
			a.ActionBase.CopyFormalsTo(newCall)
			return newCall
		}
		// Fallback for no AstCallee (tests)
		if c, ok := a.Callee.(*lg.Const); ok {
			newCallee := lg.NewConst(renamer(c.Name), c.CSort)
			newCall := NewCallActionOn(a.ActCfg, newCallee, a.ActualReturns...)
			newCall.ActionBase = a.ActionBase
			a.ActionBase.CopyFormalsTo(newCall)
			return newCall
		}
		return a
	default:
		// Python: Action.prefix_calls ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]ast.Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newArgs[i] = PrefixCallsFunc(child, renamer)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(Action)
		CopyFormalsTo(action, res)
		return res
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
		// Python: Action.drop_invariants ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]ast.Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newArgs[i] = DropInvariants(child)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(Action)
		CopyFormalsTo(action, res)
		return res
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
		// Python: Action.unroll_loops ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]ast.Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newArgs[i] = UnrollLoops(child, card)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(Action)
		CopyFormalsTo(action, res)
		return res
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
		if sym, ok := app.Func.(*lg.Const); ok {
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
func GetReferencesInto(action Action, syms map[lg.NodeKey]lg.Expr) {
	referencesRec(action, syms)
}

// EraseUnrefed replaces assignments to unreferenced symbols with
// empty sequences. Used for cone-of-influence filtering.
// syms is the set of referenced symbols; names is a set of names
// referenced by proofs that should also be kept.
// Corresponds to Python's Action.erase_unrefed(refs, names).
func EraseUnrefed(action Action, syms map[lg.NodeKey]lg.Expr, names map[string]bool) Action {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *AssignAction:
		// If LHS symbol is not referenced, erase
		// Python: if self.modifies()[0] not in refs and self.modifies()[0].name not in ref_names
		if c, ok := rootSymbol(a.LHS); ok {
			_, inSyms := syms[ConstSymKey(c)]
			if !inSyms && !names[c.Name] {
				return NewSequence()
			}
		}
		return a
	case *HavocAction:
		if a.Target != nil {
			if c, ok := a.Target.(*lg.Const); ok {
				_, inSyms := syms[ConstSymKey(c)]
				if !inSyms && !names[c.Name] {
					return NewSequence()
				}
			}
		}
		return a
	default:
		// Trace only for non-leaf actions (matches Python: AssignAction/HavocAction override erase_unrefed)
		xtracer.Trace("actions.erase_unrefed ENTER type=%s", ActionTypeName(action))
		// Recurse into children — always clone to match Python's erase_unrefed
		args := action.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(Action); ok {
				newArgs[i] = EraseUnrefed(child, syms, names)
			} else {
				newArgs[i] = arg
			}
		}
		return action.ActionClone(newArgs)
	}
}

// rootSymbol walks destructor chains to find the root symbol of an assignment LHS.
func rootSymbol(node lg.Expr) (*lg.Const, bool) {
	for {
		if app, ok := node.(*lg.Apply); ok && len(app.Terms) > 0 {
			node = app.Terms[0]
			continue
		}
		break
	}
	c, ok := node.(*lg.Const)
	return c, ok
}
