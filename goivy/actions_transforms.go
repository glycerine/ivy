// transforms.go implements action transformation functions that correspond to
// Python's assert_to_assume, modifies, references, prefix_calls,
// drop_invariants, and unroll_loops methods on Action classes.
//
// These are implemented as standalone functions rather than interface methods
// to avoid modifying the Action interface and all its implementations.
package goivy

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// AssertLocEnabled gates AssertEveryActionHasLoc. Default false so
// production builds pay zero cost. Flip to true (via go test, a build
// tag init, or a one-shot debug session) when validating Loc-propagation
// invariants. Once a clean baseline is confirmed, leave it off again.
//
// Exception to the no-package-globals rule (CLAUDE.md C): this is a
// debug-only diagnostic flag, not mutable production state. If
// multi-tenant correctness ever becomes a concern, move it to
// module.Config.
const AssertLocEnabled = false

// AssertToAssume recursively transforms an action tree, converting
// action nodes whose type name matches one of the given kinds into
// AssumeAction nodes. This is used during isolate extraction to convert
// assertions into assumptions for specification actions.
//
// The kinds map uses action type names as keys: "assert", "require", "ensure".
// This matches Python's assert_to_assume(kinds) which checks type(self) in kinds.
//
// Corresponds to Python's Action.assert_to_assume(kinds).
func AssertToAssume(action ActionsAction, kinds map[string]bool, iuCfg *IvyUtilsConfig) ActionsAction {
	if action == nil {
		return nil
	}

	switch a := action.(type) {
	case *LogicRequiresAction:
		// RequiresAction must be checked before AssertAction since it embeds it
		if kinds["require"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
			return assume
		}
		return a

	case *LogicEnsuresAction:
		// EnsuresAction must be checked before AssertAction since it embeds it.
		// Python ivy_actions.py:407-411:
		//   if iu.get_numeric_version() <= [1,6]:
		//       return Action.assert_to_assume(self,kinds)   (recurse, NO class-convert)
		//   return AssertAction.assert_to_assume(self,kinds) (class-check, convert iff in kinds)
		var ver []int
		if iuCfg != nil {
			ver = iuCfg.GetNumericVersion()
		}
		isLE16 := len(ver) >= 2 && (ver[0] < 1 || (ver[0] == 1 && ver[1] <= 6))
		if isLE16 {
			// version <= 1.6: never class-convert, just recurse-and-clone
			return assertToAssumeChildren(a, kinds, iuCfg)
		}
		// version > 1.6: AssertAction semantics — convert iff class in kinds
		if kinds["ensure"] {
			if ver == nil {
				panicf("AssertToAssume: EnsuresAction (version>1.6 branch) requires IvyUtilsConfig for version check. iuCfg='%#v'", iuCfg)
			}
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
			return assume
		}
		return assertToAssumeChildren(a, kinds, iuCfg)

	case *LogicSubgoalAction:
		// Python checks class identity: AssertAction.assert_to_assume(kinds)
		// computes mykind = type(self) and only converts when mykind is in
		// the kinds list (ivy_actions.py:396-403). Since SubgoalAction is its
		// own class (ivy_actions.py:416-421), type(SubgoalAction()) is NOT
		// AssertAction — passing ['assert'] (or [AssertAction]) does NOT
		// downgrade SubgoalActions. Match Python by only converting when the
		// caller explicitly opts in via "subgoal".
		if kinds["subgoal"] {
			assume := NewAssumeAction(a.Formula)
			assume.ActionBase = a.ActionBase
			assume.LF = a.LF // Python: AssumeAction(*self.args) preserves LF
			return assume
		}
		return a

	case *LogicAssertAction:
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
		return assertToAssumeChildren(action, kinds, iuCfg)
	}
}

// assertToAssumeChildren recursively transforms children of an action.
// Python: Action.assert_to_assume ALWAYS clones via self.clone(args).
func assertToAssumeChildren(action ActionsAction, kinds map[string]bool, iuCfg *IvyUtilsConfig) ActionsAction {
	args := action.Args()
	newArgs := make([]Node, len(args))
	for i, arg := range args {
		if child, ok := arg.(ActionsAction); ok {
			newArgs[i] = AssertToAssume(child, kinds, iuCfg)
		} else {
			newArgs[i] = arg // LF, formulas pass through unchanged
		}
	}
	// Python ALWAYS clones — no "changed" optimization.
	res := action.Clone(newArgs).(ActionsAction)
	CopyFormalsTo(action, res)
	return res
}

// ModifiesSingle returns the symbols modified by a single action, matching
// Python's per-class Action.modifies() method — NO recursion into children.
// AssignAction and HavocAction walk destructor chains.
// CrashAction walks module hierarchy.
// All other types (including SetAction and Sequence) return nil.
func ModifiesSingle(action ActionsAction, cfg ...*ActionsConfig) []*Const {
	if action == nil {
		return nil
	}
	var acfg *ActionsConfig
	if len(cfg) > 0 {
		acfg = cfg[0]
	}
	var result []*Const
	switch action.(type) {
	case *LogicAssignAction, *LogicHavocAction, *LogicCrashAction:
		modifiesRec(action, &result, acfg)
	}
	return result
}

// Modifies returns the list of symbols modified by an action, recursing
// into children. Uses the given ActionsConfig for destructor lookups (may be nil).
func Modifies(action ActionsAction, cfg ...*ActionsConfig) []*Const {
	var acfg *ActionsConfig
	if len(cfg) > 0 {
		acfg = cfg[0]
	}
	var result []*Const
	modifiesRec(action, &result, acfg)
	return result
}

func modifiesRec(action ActionsAction, result *[]*Const, cfg *ActionsConfig) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *LogicAssignAction:
		// Walk destructor chain to find root symbol.
		// Python: n = self.args[0]; while n.rep.name in destructor_sorts: n = n.args[0]; return [n.rep]
		// Python accesses n.rep (the Func/rep of an Apply), not n itself.
		// So when the chain ends on an Apply whose Func is not a destructor,
		// we return Func (the symbol), not the Apply node.
		target := a.LHS
		for {
			if app, ok := target.(*Apply); ok {
				if c, ok := app.Func.(*Const); ok {
					if isDestructor(c.Name, cfg) && len(app.Terms) > 0 {
						target = app.Terms[0]
						continue
					}
				}
			}
			break
		}
		if c, ok := target.(*Const); ok {
			*result = append(*result, c)
		} else if app, ok := target.(*Apply); ok {
			// Python: return [n.rep] — n is an Apply, n.rep is its Func symbol
			if c, ok := app.Func.(*Const); ok {
				*result = append(*result, c)
			}
		}

	case *LogicHavocAction:
		// Walk destructor chain to find root symbol, same as AssignAction.
		// Python: while n.rep.name in ivy_module.module.destructor_sorts: n = n.args[0]; return [n.rep]
		if a.Target != nil {
			target := a.Target
			for {
				if app, ok := target.(*Apply); ok {
					if c, ok := app.Func.(*Const); ok {
						if isDestructor(c.Name, cfg) && len(app.Terms) > 0 {
							target = app.Terms[0]
							continue
						}
					}
				}
				break
			}
			if c, ok := target.(*Const); ok {
				*result = append(*result, c)
			} else if app, ok := target.(*Apply); ok {
				// Python: return [n.rep] — n is an Apply, n.rep is its Func symbol
				if c, ok := app.Func.(*Const); ok {
					*result = append(*result, c)
				}
			}
		}

	case *LogicCrashAction:
		// Python: CrashAction.modifies() walks domain.hierarchy to find all
		// non-polymorphic, non-interpreted symbols under the target.
		targetName := constName(a.Target)
		xtracer.Trace("actions.CrashAction.modifies ENTER lhs_type=Atom n_type=str n_val=%s", targetName)
		if cfg != nil && cfg.Context != nil {
			mod := cfg.Context.GetDomain()
			if mod != nil && a.Target != nil {
				if targetName != "" {
					// Build set of defined names
					// Python: dfnd = [ldf.formula.defines().name for ldf in domain.definitions]
					dfnd := make(map[string]bool)
					for _, ldf := range mod.Definitions {
						if def, ok := ldf.Formula.(*LogicDefinition); ok {
							if sym, ok := def.Defines().(*Const); ok {
								dfnd[sym.Name] = true
							}
						}
					}
					// Recurse through hierarchy
					beforeLen := len(*result)
					crashModifiesRec(mod, targetName, dfnd, result)
					added := (*result)[beforeLen:]
					symNames := make([]string, len(added))
					for i, sym := range added {
						symNames[i] = fmt.Sprint(sym)
					}
					xtracer.Trace("actions.CrashAction.modifies EXIT n_syms=%d syms=%s", len(added), strings.Join(symNames, ","))
				}
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
func crashModifiesRec(mod *Module, n string, dfnd map[string]bool, result *[]*Const) {
	children, inHier := mod.Hierarchy.Get2(n)
	xtracer.Trace("actions.CrashAction.modifies.recur n_type=str n_val=%s in_hierarchy=%v", n, inHier)
	if inHier {
		for child := range children.All() {
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
			_, inSig = mod.Sig.Symbols.Get2(n)
		}
		inDfnd := dfnd[n]
		xtracer.Trace("actions.CrashAction.modifies.recur LEAF n=%s in_sig=%v in_dfnd=%v", n, inSig, inDfnd)
		if mod.Sig != nil && inSig && !inDfnd {
			for _, sym := range mod.Sig.AllSymbolsNamed(n) {
				isPoly := SymbolIsPolymorphic(sym.Name)
				isInterp := IsInterpretedSymbol(mod.Sig, sym)
				xtracer.Trace("actions.CrashAction.modifies.recur SYM name=%s poly=%v interp=%v", sym.Name, isPoly, isInterp)
				if !isPoly && !isInterp {
					*result = append(*result, sym)
				}
			}
		}
	}
}

// References returns the set of non-action symbols referenced by an action.
// Corresponds to Python's Action.references() + get_references().
func References(action ActionsAction, destructorSorts map[string]Sort) *InsMap[NodeKey, Expr] {
	result := NewInsMap[NodeKey, Expr]()
	referencesRec(action, result, destructorSorts)
	return result
}

// referencesRec matches Python's get_references (ivy_actions.py:302-306).
// Python has specialized references() methods per action type:
//   - AssignAction: only RHS + assign_refs(LHS)
//   - HavocAction: only assign_refs(target)
//   - Base Action: all non-Action args
func referencesRec(action ActionsAction, result *InsMap[NodeKey, Expr], destructorSorts map[string]Sort) {
	if action == nil {
		return
	}
	if xtracer.Enabled {
		xtracer.Trace("actions.referencesRec ENTER type=%s n_before=%d", action.Name(), result.Len())
	}
	// Dispatch: matches Python's specialized references() overrides
	switch a := action.(type) {
	case *LogicAssignAction:
		// Python AssignAction.references (ivy_actions.py:496-498):
		//   refs.update(symbols_ast(self.args[1]))  # RHS
		//   assign_refs(self, refs)                  # selective LHS
		collectSymbols(a.RHS, result)
		assignRefs(a.LHS, result, destructorSorts)
	case *LogicHavocAction:
		// Python HavocAction.references (ivy_actions.py:675-676):
		//   assign_refs(self, refs)
		assignRefs(a.Target, result, destructorSorts)
	case *LogicCallAction:
		// Python: CallAction uses base Action.references() (no override).
		// self.args[0] is an Atom. symbols_ast(Atom) does NOT yield the
		// atom's rep (is_app(Atom) is False — Atom is not App/Const/NamedBinder),
		// only recurses into Atom.args (the actual parameters).
		// self.args[1:] are actual returns.
		//
		// Go: Callee is a bare Const (the action name) — must NOT be added.
		// AstCallee.Terms holds the actual parameters — must be collected.
		// ActualReturns are the outputs.
		if a.AstCallee != nil {
			for _, term := range a.AstCallee.Terms {
				if expr, ok := term.(Expr); ok {
					collectSymbols(expr, result)
				}
			}
		}
		for _, ret := range a.ActualReturns {
			collectSymbols(ret, result)
		}
	case *LogicCrashAction:
		// Python's CrashAction target remains an AST Atom during references().
		// symbols_ilu_ast(Atom) does not yield the atom rep ("this"); it only
		// walks the terms. Go stores the target as a compiled Const/Apply, so
		// mirror the AST-Atom behavior explicitly here.
		if app, ok := a.Target.(*Apply); ok {
			for _, term := range app.Terms {
				collectSymbols(term, result)
			}
		}
	default:
		// Base Action.references (ivy_actions.py:287-290):
		//   for a in self.args:
		//       if not isinstance(a, Action):
		//           refs.update(symbols_ast(a))
		for _, arg := range action.ActionArgs() {
			if _, isAct := arg.(ActionsAction); !isAct && arg != nil {
				if xtracer.Enabled {
					if ifAct, ok := action.(*LogicIfAction); ok && arg == ifAct.Cond {
						if sc, ok := ifAct.Cond.(*SomeCondition); ok {
							xtracer.Trace("actions.referencesRec.if_cond kind=%s nparams=%d", sc.Kind, len(sc.Params))
						} else {
							xtracer.Trace("actions.referencesRec.if_cond kind=plain")
						}
					}
				}
				collectSymbols(arg, result)
			}
		}
	}
	// Python get_references (ivy_actions.py:302-306): recurse into Action children
	for _, arg := range action.ActionArgs() {
		if child, ok := arg.(ActionsAction); ok {
			referencesRec(child, result, destructorSorts)
		}
	}
	if xtracer.Enabled {
		xtracer.Trace("actions.referencesRec EXIT type=%s n_after=%d", action.Name(), result.Len())
	}
}

// assignRefs matches Python's assign_refs (ivy_actions.py:470-480).
// For destructor chains like d(x, y), adds the destructor symbol d,
// recurses into args[0] (x), and collects symbols from remaining args (y).
// For non-destructors (including bare Const symbols), processes children
// only — does NOT add the target symbol itself.
func assignRefs(node Expr, result *InsMap[NodeKey, Expr], destructorSorts map[string]Sort) {
	if node == nil {
		return
	}
	if app, ok := node.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			if _, isDestructor := destructorSorts[c.Name]; isDestructor {
				// Python: refs.add(n.rep); recur(n.args[0])
				if xtracer.Enabled {
					key := ConstSymKey(c)
					if _, already := result.Get2(key); !already {
						xtracer.Trace("actions.assignRefs.add_destructor %s", ConstSymDisplay(c))
					}
				}
				result.Set(ConstSymKey(c), c)
				if len(app.Terms) > 0 {
					assignRefs(app.Terms[0], result, destructorSorts)
				}
				// Python: for a in n.args[1:]: refs.update(symbols_ast(a))
				for i := 1; i < len(app.Terms); i++ {
					collectSymbols(app.Terms[i], result)
				}
				return
			}
		}
	}
	// Python else: for a in n.args: refs.update(symbols_ast(a))
	// For bare Const, Children()=nil → nothing added (matches Python Const.args=[])
	for _, child := range node.Children() {
		collectSymbols(child, result)
	}
}

// ConstSymKey returns a structural identity key for a Const, suitable for
// use as a map key in symbol sets. This matches Python's set behavior where
// Const objects are distinguished by recstruct.__eq__ (compares all fields:
// name + sort). We use Sexp() which encodes full structural identity.
func ConstSymKey(c *Const) NodeKey {
	return c.Sexp()
}

// ConstSymDisplay returns the display string matching Python's str(const),
// which is monkey-patched by ivy_logic.py to show "name:sortname" for numerals,
// or just "name" for non-numerals.
func ConstSymDisplay(c *Const) string {
	if c == nil {
		return "<nil>"
	}
	return c.String()
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
	if IsNumeralName(name) {
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

func collectSymbols(node Expr, result *InsMap[NodeKey, Expr]) {
	if node == nil {
		return
	}
	if c, ok := node.(*Const); ok {
		if xtracer.Enabled {
			key := ConstSymKey(c)
			if _, already := result.Get2(key); !already {
				xtracer.Trace("actions.collectSymbols.add %s", ConstSymDisplay(c))
			}
		}
		result.Set(ConstSymKey(c), c)
	}
	if app, ok := node.(*Apply); ok {
		collectSymbols(app.Func, result)
	}
	for _, child := range node.Children() {
		collectSymbols(child, result)
	}
}

// PrefixCalls renames call targets by prepending a prefix.
// Used during isolate composition.
// Corresponds to Python's Action.prefix_calls(pref) when pref is a string.
func PrefixCalls(action ActionsAction, prefix string) ActionsAction {
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
func PrefixCallsFunc(action ActionsAction, renamer func(string) string) ActionsAction {
	if action == nil || renamer == nil {
		return action
	}
	xtracer.Trace("actions.prefix_calls ENTER type=%s", ShortTypeName(action))
	switch a := action.(type) {
	case *LogicCallAction:
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
		if c, ok := a.Callee.(*Const); ok {
			newCallee := NewConst(renamer(c.Name), c.CSort)
			newCall := NewCallActionOn(a.ActCfg, newCallee, a.ActualReturns...)
			newCall.ActionBase = a.ActionBase
			a.ActionBase.CopyFormalsTo(newCall)
			return newCall
		}
		return a
	default:
		// Python: Action.prefix_calls ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(ActionsAction); ok {
				newArgs[i] = PrefixCallsFunc(child, renamer)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(ActionsAction)
		CopyFormalsTo(action, res)
		return res
	}
}

// DropInvariants strips loop invariants from while loops.
// Corresponds to Python's Action.drop_invariants().
func DropInvariants(action ActionsAction) ActionsAction {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *LogicWhileAction:
		// Remove invariant by setting it to nil
		newWhile := &LogicWhileAction{
			ActionBase: a.ActionBase,
		}
		newWhile.Cond = a.Cond
		newWhile.Body = a.Body
		// Don't copy invariant
		return newWhile
	default:
		// Python: Action.drop_invariants ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(ActionsAction); ok {
				newArgs[i] = DropInvariants(child)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(ActionsAction)
		CopyFormalsTo(action, res)
		return res
	}
}

// CardFunc computes the cardinality of a sort for loop unrolling.
// Returns -1 if the cardinality cannot be determined.
// Matches Python's card parameter to unroll_loops/unroll.
type CardFunc func(s Sort) int

// UnrollLoops converts while loops to bounded if-then-else chains.
// The card function determines the iteration bound from the loop's index sort.
// Corresponds to Python's Action.unroll_loops(card) and WhileAction.unroll(card,body).
func UnrollLoops(action ActionsAction, card CardFunc) ActionsAction {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *LogicWhileAction:
		// Python: WhileAction.unroll_loops first recurses into body,
		// then calls self.unroll(card, body)
		bodyAct, _ := a.Body.(ActionsAction)
		if bodyAct != nil {
			bodyAct = UnrollLoops(bodyAct, card)
		}
		return unrollWhile(a, card, bodyAct)
	default:
		// Python: Action.unroll_loops ALWAYS clones via self.clone(args).
		args := action.Args()
		newArgs := make([]Node, len(args))
		for i, arg := range args {
			if child, ok := arg.(ActionsAction); ok {
				newArgs[i] = UnrollLoops(child, card)
			} else {
				newArgs[i] = arg
			}
		}
		res := action.Clone(newArgs).(ActionsAction)
		CopyFormalsTo(action, res)
		return res
	}
}

// unrollWhile implements Python's WhileAction.unroll(card, body).
// Examines the condition to determine an index sort, computes cardinality,
// and builds the unrolled if-then-else chain.
func unrollWhile(a *LogicWhileAction, card CardFunc, body ActionsAction) ActionsAction {
	cond := a.Cond
	// Peel through And to find the comparison (Python lines 1027-1028)
	for {
		if andNode, ok := cond.(*LogicAnd); ok && len(andNode.Terms) > 0 {
			cond = andNode.Terms[0]
		} else {
			break
		}
	}
	// Determine index sort from condition (Python lines 1029-1033)
	var idxSort Sort
	if app, ok := cond.(*Apply); ok {
		if sym, ok := app.Func.(*Const); ok {
			name := sym.Name
			if name == "<" || name == ">" || name == "<=" || name == ">=" {
				if len(app.Terms) > 0 {
					idxSort = app.Terms[0].NodeSort()
				}
			}
		}
	} else if notNode, ok := cond.(*LogicNot); ok {
		if eq, ok := notNode.Body.(*Eq); ok {
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
	orExpr := &LogicOr{}
	assumeFalse := NewAssumeAction(orExpr)
	assumeFalse.SetLineno(a.GetLineno())
	res := NewIfAction(a.Cond, assumeFalse)
	res.SetLineno(a.GetLineno())
	for i := 0; i < cardsort; i++ {
		var bodyExpr Expr
		if body != nil {
			bodyExpr = body
		} else {
			bodyExpr = a.Body
		}
		seq := NewSequence(bodyExpr, res)
		seq.SetLineno(a.GetLineno())
		res = NewIfAction(a.Cond, seq)
		res.SetLineno(a.GetLineno())
	}
	CopyFormalsTo(a, res)
	return res
}

// GetReferencesInto accumulates non-action symbol references from an
// action into the given set. Corresponds to Python's get_references().
func GetReferencesInto(action ActionsAction, syms *InsMap[NodeKey, Expr], destructorSorts map[string]Sort) {
	referencesRec(action, syms, destructorSorts)
}

// EraseUnrefed replaces assignments to unreferenced symbols with
// empty sequences. Used for cone-of-influence filtering.
// syms is the set of referenced symbols; names is a set of names
// referenced by proofs that should also be kept.
// Corresponds to Python's Action.erase_unrefed(refs, names).
func EraseUnrefed(action ActionsAction, syms *InsMap[NodeKey, Expr], names map[string]bool, destructorSorts map[string]Sort) ActionsAction {
	if action == nil {
		return nil
	}
	switch a := action.(type) {
	case *LogicAssignAction:
		// If LHS symbol is not referenced, erase
		// Python: if self.modifies()[0] not in refs and self.modifies()[0].name not in ref_names
		if c, ok := rootSymbol(a.LHS, destructorSorts); ok {
			_, inSyms := syms.Get2(ConstSymKey(c))
			if !inSyms && !names[c.Name] {
				return emptyErasedActionLike(a)
			}
		}
		return a
	case *LogicHavocAction:
		// Python HavocAction.erase_unrefed uses modifies() which walks destructor chains
		if a.Target != nil {
			if c, ok := rootSymbol(a.Target, destructorSorts); ok {
				_, inSyms := syms.Get2(ConstSymKey(c))
				if !inSyms && !names[c.Name] {
					return emptyErasedActionLike(a)
				}
			}
		}
		return a
	default:
		// Trace only for non-leaf actions (matches Python: AssignAction/HavocAction override erase_unrefed)
		xtracer.Trace("actions.erase_unrefed ENTER type=%s", ActionTypeName(action))
		// Recurse into children — always clone to match Python's erase_unrefed
		args := action.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, arg := range args {
			if child, ok := arg.(ActionsAction); ok {
				newArgs[i] = EraseUnrefed(child, syms, names, destructorSorts)
			} else {
				newArgs[i] = arg
			}
		}
		return action.ActionClone(newArgs)
	}
}

func emptyErasedActionLike(action ActionsAction) *LogicSequence {
	erased := NewSequence()
	erased.SetLineno(action.GetLineno())
	CopyFormalsTo(action, erased)
	return erased
}

// rootSymbol finds the root symbol modified by an assignment LHS.
// Matches Python AssignAction.modifies() / HavocAction.modifies():
//
//	n = self.args[0]
//	while n.rep.name in module.destructor_sorts:
//	    n = n.args[0]
//	return [n.rep]
//
// For non-destructor functions like isd.rsp(P,M,A,T), returns isd.rsp.
// For destructor chains like d2(d1(x)), walks through destructors to return x.
func rootSymbol(node Expr, destructorSorts map[string]Sort) (*Const, bool) {
	for {
		app, ok := node.(*Apply)
		if !ok {
			break
		}
		c, ok := app.Func.(*Const)
		if !ok {
			break
		}
		if _, isDestr := destructorSorts[c.Name]; isDestr && len(app.Terms) > 0 {
			// Walk through destructor chain
			node = app.Terms[0]
			continue
		}
		// Non-destructor function — this IS the root symbol
		return c, true
	}
	// Bare Const (e.g., zero-argument symbol)
	if c, ok := node.(*Const); ok {
		return c, true
	}
	return nil, false
}

// AssertEveryActionHasLoc recursively walks an action tree and panics
// if any nested action has an empty Loc. The `where` argument is a
// human-readable context label (e.g., "isolate.end_classify
// actname=cfabric.step") that is included in the panic message so the
// failing pipeline stage is obvious.
//
// This is a DIAGNOSTIC tool, not a silent fix-it pass. It is gated by
// the package-level flag AssertLocEnabled which defaults to false.
// Production runs leave the flag off; this walker is intended to be
// flipped on temporarily to validate that all constructor sites
// preserve Loc, and then flipped off again when the codebase is clean.
//
// The walker reports the FIRST missing-Loc node it finds with:
//   - The action's Go type name
//   - The action's Sexp() (truncated to ~120 chars)
//   - The full context label
//   - The path of enclosing parent action types
//   - The root action's Sexp (truncated)
//
// so the offending pipeline stage and constructor are immediately
// identifiable from the panic stack.
func AssertEveryActionHasLoc(action ActionsAction, where string) {
	if !AssertLocEnabled || action == nil {
		return
	}
	var path []string
	assertEveryActionHasLocRec(action, action, where, &path)
}

func assertEveryActionHasLocRec(root ActionsAction, action ActionsAction, where string, path *[]string) {
	if action == nil {
		return
	}
	typeName := ShortTypeName(action)
	*path = append(*path, typeName)
	defer func() { *path = (*path)[:len(*path)-1] }()

	if action.GetLineno() == (Location{}) {
		sx := string(action.Sexp())
		if len(sx) > 200 {
			sx = sx[:200] + "..."
		}
		rootSx := string(root.Sexp())
		if len(rootSx) > 400 {
			rootSx = rootSx[:400] + "..."
		}
		panic(fmt.Sprintf(
			"AssertEveryActionHasLoc: missing Loc on %s at %s\n  path: %s\n  sexp: %s\n  root sexp: %s",
			typeName, where, strings.Join(*path, " > "), sx, rootSx,
		))
	}

	for _, sub := range action.IterSubactions() {
		if sub == nil || sub == action {
			continue
		}
		assertEveryActionHasLocRec(root, sub, where, path)
	}
}
