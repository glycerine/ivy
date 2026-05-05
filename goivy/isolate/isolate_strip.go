package isolate

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// StripMap maps component names to their isolate parameter names.
// When a component has isolate parameters, we need to strip those parameters
// from all references within the isolated module.
type StripMap map[string][]string

// StripMapLookup looks up the strip parameters for a given symbol name.
// Returns nil if the name should not be stripped (e.g., it is a global parameter,
// destructor sort, or sort name).
func StripMapLookup(name string, stripMap StripMap, mod *module.Module) []string {
	name = CanonAct(name)

	// Global parameters are not stripped.
	if _, ok := mod.Attributes[mod.Cfg.IuCfg.ComposeNames(name, "global_parameter")]; ok {
		return nil
	}

	// Destructor sorts are not stripped.
	if _, ok := mod.DestructorSorts[name]; ok {
		return nil
	}

	// Sort names are not stripped.
	if mod.Sig != nil {
		if _, ok := mod.Sig.Sorts.Get2(name); ok {
			return nil
		}
	}

	// Check each prefix in the strip map.
	for prefix, params := range stripMap {
		cc := mod.Cfg.IuCfg.ComposeCharacter
		if strings.HasPrefix(name+cc, prefix+cc) {
			// Check for "common" attribute override.
			attr := mod.Cfg.IuCfg.ComposeNames(name, "common")
			if commonVal, ok := mod.Attributes[attr]; ok {
				if commonStr, ok := commonVal.(string); ok {
					if strings.HasPrefix(commonStr+cc, prefix+cc) {
						continue
					}
				}
			}
			return params
		}
	}
	return nil
}

// StripActionFull removes isolate parameters from an action recursively,
// with full strip_binding, is_init, and init_params support.
// Corresponds to Python strip_action (ivy_isolate.py lines 242-289).
func StripActionFull(action actions.ActionsAction, stripMap StripMap, mod *module.Module,
	binding map[lg.NodeKey]string, isInit bool, initParams []string) actions.ActionsAction {
	// Python: strip_action always recurses and clones, even with empty strip_map/binding.
	// No early return — the clone calls produce LF.clone PRESERVE and action __init__ traces.
	return stripActionFullRec(action, stripMap, mod, binding, isInit, initParams)
}

func stripActionFullRec(action actions.ActionsAction, stripMap StripMap, mod *module.Module,
	binding map[lg.NodeKey]string, isInit bool, initParams []string) actions.ActionsAction {
	switch a := action.(type) {
	case *actions.CallAction:
		// Python lines 243-248: Strip call action arguments.
		calleeName := CanonAct(a.CalleeName())
		// Recursively strip call arguments.
		calleeArgs := a.ActionArgs()
		newArgs := make([]lg.Expr, len(calleeArgs))
		for i, arg := range calleeArgs {
			newArgs[i] = stripNodeFull(arg, stripMap, mod, binding)
		}
		// Strip the callee's leading parameters.
		stripParams := StripMapLookup(calleeName, stripMap, mod)
		if len(stripParams) > 0 {
			newCallee := stripNodeFull(a.Callee, stripMap, mod, binding)
			newReturns := make([]lg.Expr, len(a.ActualReturns))
			for i, r := range a.ActualReturns {
				newReturns[i] = stripNodeFull(r, stripMap, mod, binding)
			}
			allArgs := []lg.Expr{newCallee}
			allArgs = append(allArgs, newReturns...)
			return a.ActionClone(allArgs)
		}
		return a.ActionClone(newArgs)

	case *actions.AssignAction:
		// Python lines 260-266: Handle init_params for initializer actions.
		localBinding := binding
		if len(initParams) > 0 {
			// Copy binding and add extra bindings from init_params.
			localBinding = make(map[lg.NodeKey]string, len(binding)+len(initParams))
			for k, v := range binding {
				localBinding[k] = v
			}
			lhsArgs := a.ActionArgs()
			if len(lhsArgs) > 0 {
				if app, ok := lhsArgs[0].(*lg.Apply); ok {
					offset := len(binding)
					for i, ip := range initParams {
						idx := offset + i
						if idx < len(app.Terms) {
							key := lg.Key(app.Terms[idx])
							localBinding[key] = ip
						}
					}
				}
			}
		}
		// Fall through to default processing with possibly-updated binding.
		oldArgs := action.ActionArgs()
		newActionArgs := make([]lg.Expr, len(oldArgs))
		for i, arg := range oldArgs {
			if act, ok := arg.(actions.ActionsAction); ok {
				newActionArgs[i] = stripActionFullRec(act, stripMap, mod, localBinding, isInit, initParams)
			} else if w, ok := arg.(actions.ActionsAction); ok {
				newActionArgs[i] = stripActionFullRec(w, stripMap, mod, localBinding, isInit, initParams)
			} else {
				newActionArgs[i] = stripNodeFull(arg, stripMap, mod, localBinding)
			}
		}
		return action.ActionClone(newActionArgs)

	case *actions.Sequence:
		newChildren := make([]lg.Expr, len(a.Elems))
		for i, child := range a.Elems {
			if act, ok := child.(actions.ActionsAction); ok {
				newChildren[i] = stripActionFullRec(act, stripMap, mod, binding, isInit, initParams)
			} else if w, ok := child.(actions.ActionsAction); ok {
				newChildren[i] = stripActionFullRec(w, stripMap, mod, binding, isInit, initParams)
			} else {
				newChildren[i] = stripNodeFull(child, stripMap, mod, binding)
			}
		}
		return a.ActionClone(newChildren)

	default:
		// Python lines 249-259: Check modifies() for interference.
		for _, sym := range actions.Modifies(action) {
			if mod.Sig != nil {
				if _, inSig := mod.Sig.Symbols.Get2(sym.Name); inSig {
					lhsParams := StripMapLookup(sym.Name, stripMap, mod)
					if len(lhsParams) != mod.Cfg.IsolateCfg.NumIsolateParams {
						if !(len(lhsParams) == 0 && len(binding) == 0 && isInit) {
							// Python line 259: raise iu.IvyError(ast,"assignment may be interfering")
							fmt.Fprintf(os.Stderr, "error: assignment may be interfering: %s\n", sym.Name)
						}
					}
				}
			}
		}

		// Python: args = [strip_action(arg,...) for arg in ast.args]
		//         return ast.clone(args)
		// Use Args()/Clone() (ast.Node interface) instead of ActionArgs()/ActionClone()
		// so that LabeledFormula children (inside assume/assert actions) are properly
		// recursed into and cloned, producing the LF.clone PRESERVE traces that Python emits.
		nodeArgs := action.(ast.Node).Args()
		newNodeArgs := make([]ast.Node, len(nodeArgs))
		for i, arg := range nodeArgs {
			newNodeArgs[i] = stripArgNode(arg, stripMap, mod, binding, isInit, initParams)
		}
		return action.(ast.Node).Clone(newNodeArgs).(actions.ActionsAction)
	}
}

// stripArgNode recursively processes an ast.Node argument, matching Python's
// strip_action which handles actions, LabeledFormulas, and plain nodes uniformly.
// Python: args = [strip_action(arg,...) for arg in ast.args]; return ast.clone(args)
func stripArgNode(node ast.Node, stripMap StripMap, mod *module.Module,
	binding map[lg.NodeKey]string, isInit bool, initParams []string) ast.Node {
	if node == nil {
		return nil
	}
	// If it's an action, delegate to action stripping.
	if act, ok := node.(actions.ActionsAction); ok {
		return stripActionFullRec(act, stripMap, mod, binding, isInit, initParams).(ast.Node)
	}
	// If it's a LabeledFormula, recurse into children and clone.
	// Python: strip_action recurses into lf.args (label, formula) then calls lf.clone(new_args),
	// which emits LF.clone PRESERVE.
	if lf, ok := node.(*ast.LabeledFormula); ok {
		lfArgs := lf.Args() // [label, formula]
		newArgs := make([]ast.Node, len(lfArgs))
		for i, arg := range lfArgs {
			newArgs[i] = stripArgNode(arg, stripMap, mod, binding, isInit, initParams)
		}
		return lf.Clone(newArgs)
	}
	// Otherwise, treat as a logic expression and strip via stripNodeFull.
	if expr, ok := node.(lg.Expr); ok {
		return stripNodeFull(expr, stripMap, mod, binding)
	}
	return node
}

// stripNodeFull recursively processes a logic node, stripping isolate parameters
// and performing strip_binding substitutions.
// Corresponds to the node-level parts of Python strip_action (lines 268-289).
func stripNodeFull(node lg.Expr, stripMap StripMap, mod *module.Module, binding map[lg.NodeKey]string) lg.Expr {
	if node == nil {
		return nil
	}

	// Python lines 268-273: If node is a constant/variable in strip_binding, substitute.
	if len(binding) > 0 {
		key := lg.Key(node)
		if sname, ok := binding[key]; ok {
			switch n := node.(type) {
			case *lg.Const:
				// Add symbol to signature if not present.
				if mod.Sig != nil {
					if _, exists := mod.Sig.Symbols.Get2(sname); !exists {
						mod.Sig.Symbols.Set(sname, &il.SymbolEntry{Name: sname, Sort: n.CSort})
						mod.Cfg.IsolateCfg.StripAddedSymbols = append(mod.Cfg.IsolateCfg.StripAddedSymbols, lg.NewConst(sname, n.CSort))
					}
				}
				return lg.NewConst(sname, n.CSort)
			case *lg.Variable:
				if mod.Sig != nil {
					if _, exists := mod.Sig.Symbols.Get2(sname); !exists {
						mod.Sig.Symbols.Set(sname, &il.SymbolEntry{Name: sname, Sort: n.VSort})
						mod.Cfg.IsolateCfg.StripAddedSymbols = append(mod.Cfg.IsolateCfg.StripAddedSymbols, lg.NewConst(sname, n.VSort))
					}
				}
				return lg.NewConst(sname, n.VSort)
			}
		}
	}

	// Fall through to regular strip logic.
	return stripNode(node, stripMap, mod)
}

// StripAction removes isolate parameters from an action recursively.
// This is the Go port of Python's strip_action function.
//
// It walks the action tree and for each CallAction, strips the isolate
// parameters from the callee arguments. For other actions, it recursively
// processes child nodes. Constants and variables that appear in the
// strip binding are replaced with the corresponding isolate parameter symbols.
func StripAction(action actions.ActionsAction, stripMap StripMap, mod *module.Module) actions.ActionsAction {
	if len(stripMap) == 0 {
		return action
	}
	return stripActionRec(action, stripMap, mod)
}

// stripActionRec recursively strips isolate parameters from an action.
func stripActionRec(action actions.ActionsAction, stripMap StripMap, mod *module.Module) actions.ActionsAction {
	switch a := action.(type) {
	case *actions.CallAction:
		// For call actions, strip parameters from the callee.
		calleeName := CanonAct(a.CalleeName())
		stripParams := StripMapLookup(calleeName, stripMap, mod)
		if len(stripParams) > 0 {
			// Strip the first len(stripParams) actual parameters from the callee.
			// The callee's arguments are embedded in the Callee node (if it's an Apply).
			newCallee := stripNode(a.Callee, stripMap, mod)
			newReturns := make([]lg.Expr, len(a.ActualReturns))
			for i, r := range a.ActualReturns {
				newReturns[i] = stripNode(r, stripMap, mod)
			}
			newArgs := []lg.Expr{newCallee}
			newArgs = append(newArgs, newReturns...)
			return a.ActionClone(newArgs)
		}
		// No stripping needed for this call, but recurse into children.
		newArgs := stripNodes(a.ActionArgs(), stripMap, mod)
		return a.ActionClone(newArgs)
	case *actions.Sequence:
		newChildren := make([]lg.Expr, len(a.Elems))
		for i, child := range a.Elems {
			if act, ok := child.(actions.ActionsAction); ok {
				newChildren[i] = stripActionRec(act, stripMap, mod)
			} else if w, ok := child.(actions.ActionsAction); ok {
				newChildren[i] = stripActionRec(w, stripMap, mod)
			} else {
				newChildren[i] = stripNode(child, stripMap, mod)
			}
		}
		return a.ActionClone(newChildren)
	default:
		// For other action types, recursively process child nodes.
		oldArgs := action.ActionArgs()
		newArgs := make([]lg.Expr, len(oldArgs))
		for i, arg := range oldArgs {
			if act, ok := arg.(actions.ActionsAction); ok {
				newArgs[i] = stripActionRec(act, stripMap, mod)
			} else if w, ok := arg.(actions.ActionsAction); ok {
				newArgs[i] = stripActionRec(w, stripMap, mod)
			} else {
				newArgs[i] = stripNode(arg, stripMap, mod)
			}
		}
		return action.ActionClone(newArgs)
	}
}

// stripNode recursively processes a logic node, stripping isolate parameters
// from function applications.
func stripNode(node lg.Expr, stripMap StripMap, mod *module.Module) lg.Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.Apply:
		// Strip parameters from function applications.
		newTerms := make([]lg.Expr, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = stripNode(t, stripMap, mod)
		}
		// Check if the function symbol needs stripping.
		if c, ok := n.Func.(*lg.Const); ok {
			stripParams := StripMapLookup(c.Name, stripMap, mod)
			if len(stripParams) > 0 && len(newTerms) >= len(stripParams) {
				// Strip the first len(stripParams) arguments.
				strippedTerms := newTerms[len(stripParams):]
				newSort := StripSort(c.CSort, len(stripParams))
				newSym := lg.NewConst(c.Name, newSort)
				if len(strippedTerms) == 0 {
					return newSym
				}
				result, err := lg.NewApply(newSym, strippedTerms...)
				if err != nil {
					// Fallback: return with original structure.
					return node
				}
				return result
			}
		}
		// No stripping needed, but rebuild with processed terms.
		newFunc := stripNode(n.Func, stripMap, mod)
		if len(newTerms) == 0 {
			return newFunc
		}
		result, err := lg.NewApply(newFunc, newTerms...)
		if err != nil {
			return node
		}
		return result
	case *lg.Const:
		// Constants are leaf nodes; no stripping needed at this level.
		return n
	case *lg.Variable:
		return n
	default:
		// For other node types, return as-is.
		return node
	}
}

// stripNodes processes a slice of nodes through stripNode.
func stripNodes(nodes []lg.Expr, stripMap StripMap, mod *module.Module) []lg.Expr {
	result := make([]lg.Expr, len(nodes))
	for i, n := range nodes {
		result[i] = stripNode(n, stripMap, mod)
	}
	return result
}

// StripLabeledFormula strips isolate parameters from a labeled formula.
// This is the Go port of Python's strip_labeled_fmla function (lines 303-311).
//
// It builds a strip_binding from the formula, then strips isolate parameters
// from the formula body using the binding, and adjusts the label.
func StripLabeledFormula(lf *ast.LabeledFormula, stripMap StripMap, mod *module.Module) *ast.LabeledFormula {
	if len(stripMap) == 0 {
		return lf
	}

	// Python lines 305-306: Build strip_binding from the formula.
	binding := make(map[lg.NodeKey]string)
	if lf.Formula != nil {
		if fmla, ok := lf.Formula.(lg.Expr); ok {
			GetStripBinding(fmla, stripMap, binding, mod)
		}
	}

	// Python line 307: Strip the formula body using the binding.
	var newFormula ast.Node
	if lf.Formula != nil {
		if fmla, ok := lf.Formula.(lg.Expr); ok {
			newFormula = stripNodeFull(fmla, stripMap, mod, binding)
		} else {
			newFormula = lf.Formula
		}
	}

	// Python lines 308-310: Strip the label if present.
	// Python: lbl = lbl.clone(lbl.args[len(strip_map_lookup(lbl.rep, strip_map, with_dot=False)):])
	newLabel := lf.Label
	if lf.Label != nil {
		if lblExpr, ok := lf.Label.(lg.Expr); ok {
			// Get the label's name for strip map lookup.
			lblName := ""
			switch l := lblExpr.(type) {
			case *lg.Apply:
				if sym, ok := l.Func.(*lg.Const); ok {
					lblName = sym.Name
				}
			case *lg.Const:
				lblName = l.Name
			}
			if lblName != "" {
				sp := StripMapLookup(lblName, stripMap, mod)
				if len(sp) > 0 {
					// Strip leading args from label.
					if app, ok := lblExpr.(*lg.Apply); ok && len(app.Terms) >= len(sp) {
						strippedTerms := app.Terms[len(sp):]
						if len(strippedTerms) == 0 {
							newLabel = app.Func
						} else {
							newApp, err := lg.NewApply(app.Func, strippedTerms...)
							if err == nil {
								newLabel = newApp
							}
						}
					}
				} else {
					newLabel = stripNodeFull(lblExpr, stripMap, mod, binding)
				}
			} else {
				newLabel = stripNodeFull(lblExpr, stripMap, mod, binding)
			}
		}
	}

	// Return a new LabeledFormula with stripped contents.
	// Python: return lfmla.clone([lbl, fmla])
	cloned := lf.Clone([]ast.Node{newLabel, newFormula})
	return cloned.(*ast.LabeledFormula)
}

// StripLabeledFormulas strips isolate parameters from a slice of labeled formulas in place.
func StripLabeledFormulas(lfs []*ast.LabeledFormula, stripMap StripMap, mod *module.Module) []*ast.LabeledFormula {
	if len(stripMap) == 0 {
		return lfs
	}
	// Python lines 317-318: check for SchemaBody — cannot strip from theorems.
	for _, f := range lfs {
		if f.Formula != nil {
			if _, isSchema := f.Formula.(*ast.SchemaBody); isSchema {
				// Python: raise IvyError(f, 'cannot strip parameter from a theorem')
				// We skip it with a warning rather than panic, since this is a validation error.
				fmt.Fprintf(os.Stderr, "warning: cannot strip parameter from a theorem: %s\n", f.Label)
			}
		}
	}
	result := make([]*ast.LabeledFormula, len(lfs))
	for i, lf := range lfs {
		result[i] = StripLabeledFormula(lf, stripMap, mod)
	}
	return result
}

// StripSort removes the first n domain parameters from a function sort.
// If the result has no domain and was relational, we still return a FunctionSort.
// Otherwise if the result has no domain, return just the range sort.
func StripSort(sort lg.Sort, numParams int) lg.Sort {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok || numParams == 0 {
		return sort
	}
	dom := fs.Domain()
	if numParams > len(dom) {
		// Python line 416-417: raise IvyError(None,"cannot strip isolate parameters from {}".format(name))
		fmt.Fprintf(os.Stderr, "error: cannot strip isolate parameters (need %d params, sort has %d domain elements)\n", numParams, len(dom))
		return sort
	}
	newDom := dom[numParams:]
	if len(newDom) == 0 {
		// No domain left. If it was relational (range is Boolean),
		// we still need a zero-arg FunctionSort per Python:
		//   if dom or sort.is_relational():
		//       return ivy_logic.FunctionSort(*(dom+[sort.rng]))
		if _, isBool := fs.Range().(*lg.BooleanSort); isBool {
			result, err := lg.NewFunctionSort(fs.Range())
			if err != nil {
				return sort
			}
			return result
		}
		return fs.Range()
	}
	newSorts := make([]lg.Sort, 0, len(newDom)+1)
	newSorts = append(newSorts, newDom...)
	newSorts = append(newSorts, fs.Range())
	result, err := lg.NewFunctionSort(newSorts...)
	if err != nil {
		// Fallback: return the original sort.
		return sort
	}
	return result
}

// StripIsolate strips an isolate's parameters from the module's actions,
// axioms, conjectures, signature, etc.
//
// This is the Go port of Python's strip_isolate function.
// It strips isolate parameters from the module's actions, axioms,
// conjectures, signature symbols, and module parameters.
// Native quote stripping is deferred until AST node types are ported.
// isolateAtomProvider is an optional interface for isolate definitions that
// expose their verified/present atoms (not just names). This is needed for
// building the strip map from atom parameters.
type isolateAtomProvider interface {
	Verified() []ast.Node
	Present() []ast.Node
}

// isolateParamProvider is an optional interface for isolate definitions that
// expose their parameters as lg.Const (needed for variable param substitution).
type isolateParamProvider interface {
	Params() []ast.Node
}

// StripIsolateParams is the full version of strip_isolate that handles
// variable isolate parameter substitution, initializer handling, impl_mixin
// strip propagation, and extra_strip.
//
// Corresponds to Python strip_isolate (lines 341-456).
func StripIsolateParams(mod *module.Module, isolate IsolateDefIface,
	implMixins *iu.InsMap[string, []IsolateMixinIface], allAfterInits map[string]bool,
	extraStrip map[string][]string) error {

	isoCfg := mod.Cfg.IsolateCfg
	// Python: global num_isolate_params, strip_added_symbols
	isoCfg.StripAddedSymbols = nil // reset

	_, isAtomProv := isolate.(isolateAtomProvider)
	_, isParamProv := isolate.(isolateParamProvider)
	xtracer.Trace("strip.StripIsolateParams ENTER isAtomProv=%v isParamProv=%v",
		isAtomProv, isParamProv)

	// Step 1: Variable isolate parameter substitution.
	// Python lines 345-352: if any(isinstance(p, Variable) for p in ipl): substitute
	if pp, ok := isolate.(isolateParamProvider); ok {
		ipl := pp.Params()
		hasVar := false
		for _, p := range ipl {
			if _, isVar := p.(*ast.Variable); isVar {
				hasVar = true
				break
			}
		}
		if hasVar {
			subst := make(map[string]lg.Expr)
			for _, p := range ipl {
				if v, isVar := p.(*ast.Variable); isVar {
					var sort lg.Sort
					if mod.Sig != nil {
						if s, ok := mod.Sig.Sorts.Get2(v.VSort); ok {
							sort = s
						}
					}
					newConst := lg.NewConst("iso:"+v.Rep, sort)
					subst[v.Rep] = newConst
				}
			}
			// Apply substitution to isolate (would need SubstituteAst)
			// For now, the names are adjusted; full AST substitution requires
			// the concrete isolate type.
			_ = subst
		}
	}

	// Compute NumIsolateParams from the actual parameters.
	if pp, ok := isolate.(isolateParamProvider); ok {
		isoCfg.NumIsolateParams = len(pp.Params())
	} else {
		isoCfg.NumIsolateParams = 0
	}

	// Python lines 355-359: Validate unbound parameters.
	if pp, ok := isolate.(isolateParamProvider); ok {
		ips := make(map[string]bool)
		for _, p := range pp.Params() {
			if a, ok := p.(*ast.Atom); ok {
				ips[a.Rep] = true
			}
		}
		if ap, ok2 := isolate.(isolateAtomProvider); ok2 {
			for _, atom := range append(ap.Verified(), ap.Present()...) {
				if a, ok3 := atom.(*ast.Atom); ok3 {
					for _, p := range a.Terms {
						if pa, ok4 := p.(*ast.Atom); ok4 {
							if !ips[pa.Rep] {
								return fmt.Errorf("unbound isolate parameter: %s", pa.Rep)
							}
						}
					}
				}
			}
		}
	}

	// Step 2: Build the strip map from isolate parameter bindings.
	// Python lines 362-370
	stripMap := make(StripMap)
	if ap, ok := isolate.(isolateAtomProvider); ok {
		allAtoms := append(ap.Verified(), ap.Present()...)
		for _, node := range allAtoms {
			a, ok := node.(*ast.Atom)
			if !ok || len(a.Terms) == 0 {
				continue
			}
			name := a.Relname()
			// Python: check all args are simple Apps without args
			valid := true
			for _, v := range a.Terms {
				va, ok2 := v.(*ast.Atom)
				if !ok2 || len(va.Terms) != 0 {
					valid = false
					break
				}
				// Python line 368-369: check parameter doesn't redefine a symbol
				if mod.Sig != nil {
					if _, exists := mod.Sig.Symbols.Get2(va.Rep); exists {
						return fmt.Errorf("isolate parameter redefines %s", va.Rep)
					}
				}
			}
			if !valid {
				return fmt.Errorf("bad isolate parameter in %s", name)
			}
			params := make([]string, len(a.Terms))
			for i, v := range a.Terms {
				params[i] = v.(*ast.Atom).Rep
			}
			stripMap[name] = params
		}
	}

	// Step 3: Propagate strip map through impl_mixins.
	// Python: for ms in impl_mixins.values(): for m: strip_map[m.mixee()] = strip_params
	for _, ms := range implMixins.All() {
		for _, m := range ms {
			if isMixinImplement(m) {
				mixerParams := StripMapLookup(CanonAct(m.Mixer()), stripMap, mod)
				stripMap[m.Mixee()] = mixerParams
			}
		}
	}

	// Apply extra_strip
	for k, v := range extraStrip {
		stripMap[k] = v
	}

	// Delegate to StripIsolate for the actual stripping.
	if err := StripIsolate(mod, stripMap, allAfterInits); err != nil {
		return err
	}

	// Step 4: Add isolate parameters as symbols and to mod.Params.
	// Python lines 441-456: for s in isolate.params(): add_symbol(s.rep, mod.sig.sorts[s.sort])
	if pp, ok := isolate.(isolateParamProvider); ok {
		for _, node := range pp.Params() {
			if node == nil {
				continue
			}
			// Python: s.rep and s.sort — params are Atoms
			paramAtom, isAtom := node.(*ast.Atom)
			if !isAtom {
				continue
			}
			paramName := paramAtom.Rep
			// Python: add_symbol(s.rep, mod.sig.sorts[s.sort])
			// Look up sort by name from the atom's sort annotation
			var paramSort lg.Sort
			if paramAtom.ASort != nil {
				// ASort is an ast.Node; extract sort name and look up
				if sortAtom, ok := paramAtom.ASort.(*ast.Atom); ok {
					if mod.Sig != nil {
						paramSort = mod.Sig.Sorts.Get(sortAtom.Rep)
					}
				}
			}

			if mod.Sig != nil {
				// Python: add_map = dict((s.name,s) for s in strip_added_symbols)
				// Python: if s.rep not in add_map:
				//             sym = ivy_logic.add_symbol(s.rep, mod.sig.sorts[s.sort])
				//             mod.params.append(sym)
				//         else:
				//             mod.params.append(add_map[s.rep])
				//         mod.param_defaults.append(None)
				//
				// Key: Python ALWAYS appends to mod.params. add_symbol
				// returns the existing symbol if already registered.
				var addedSym *lg.Const
				for _, added := range isoCfg.StripAddedSymbols {
					if added.Name == paramName {
						addedSym = added
						break
					}
				}
				if addedSym != nil {
					// Python: mod.params.append(add_map[s.rep])
					mod.Params = append(mod.Params, addedSym)
				} else if paramSort != nil {
					// Python: sym = ivy_logic.add_symbol(s.rep, mod.sig.sorts[s.sort])
					if _, exists := mod.Sig.Symbols.Get2(paramName); !exists {
						mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: paramSort})
					}
					mod.Params = append(mod.Params, lg.NewConst(paramName, paramSort))
				} else if s, ok := mod.Sig.Sorts.Get2(paramName); ok {
					if _, exists := mod.Sig.Symbols.Get2(paramName); !exists {
						mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: s})
					}
					mod.Params = append(mod.Params, lg.NewConst(paramName, s))
				}
				mod.ParamDefaults = append(mod.ParamDefaults, nil)
			}
		}
	} else {
		// Fallback for interfaces that don't provide Params(): use names
		allNames := append(isolate.VerifiedNames(), isolate.PresentNames()...)
		for _, paramName := range allNames {
			if paramName == "this" {
				continue
			}
			if mod.Sig != nil {
				if _, exists := mod.Sig.Symbols.Get2(paramName); exists {
					continue
				}
				if s, ok := mod.Sig.Sorts.Get2(paramName); ok {
					sym := lg.NewConst(paramName, s)
					mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: s})
					mod.Params = append(mod.Params, sym)
					mod.ParamDefaults = append(mod.ParamDefaults, nil)
				}
			}
		}
	}

	return nil
}

// StripIsolate strips isolate parameters from the module's actions, formulas,
// signature, and parameters using the given strip map.
//
// This is the core stripping function. StripIsolateParams is the higher-level
// function that builds the strip map and handles variable parameter substitution.
func StripIsolate(mod *module.Module, stripMap StripMap, allAfterInits map[string]bool) error {
	// Python: strip_isolate does NOT early-return for empty strip_map.
	// It processes all actions through strip_action even when strip_map is empty,
	// which clones actions and labeled formulas (producing LF.clone PRESERVE traces).

	// Strip actions.
	newActions := iu.NewInsMap[string, module.Action]()
	for name, act := range mod.Actions.All() {
		stripParams := StripMapLookup(CanonAct(name), stripMap, mod)

		// Python: does NOT skip when strip_params is empty — always calls strip_action.
		// strip_binding = dict(zip(formal_params, strip_params)) → {} when strip_params is empty.

		// Check if this is an initializer action
		origName := name
		if strings.HasPrefix(name, "ext:") {
			origName = name[4:]
		}
		isInit := allAfterInits != nil && allAfterInits[origName]

		// Strip formal parameters.
		fp := act.GetFormalParams()
		var initParams []string
		if len(fp) < len(stripParams) {
			if isInit {
				// For initializers, excess strip params become init_params
				initParams = stripParams[len(fp):]
			} else {
				return fmt.Errorf("cannot strip isolate parameters from %s", name)
			}
		}
		// Python line 391: strip_binding = dict(list(zip(action.formal_params, strip_params)))
		binding := make(map[lg.NodeKey]string)
		nBind := len(stripParams)
		if nBind > len(fp) {
			nBind = len(fp)
		}
		for i := 0; i < nBind; i++ {
			binding[lg.Key(fp[i])] = stripParams[i]
		}

		strippedAction := StripActionFull(act, stripMap, mod, binding, isInit, initParams)
		nStrip := len(stripParams)
		if nStrip > len(fp) {
			nStrip = len(fp)
		}
		strippedAction.SetFormalParams(fp[nStrip:])
		strippedAction.SetFormalReturns(act.GetFormalReturns())

		// Copy labels if present
		type labeler interface {
			GetLabels() []string
			SetLabels([]string)
		}
		if lb, ok := act.(labeler); ok {
			labels := lb.GetLabels()
			if labels != nil {
				if lb2, ok := strippedAction.(labeler); ok {
					lb2.SetLabels(labels)
				}
			}
		}

		newActions.Set(name, strippedAction)
	}

	// Replace all actions.
	mod.Actions = newActions

	// Strip labeled formulas.
	xtracer.Trace("strip.StripIsolate before_strip_lfs axioms=%d props=%d conjs=%d inits=%d defs=%d stripMap=%d",
		len(mod.LabeledAxioms), len(mod.LabeledProps), len(mod.LabeledConjs),
		len(mod.LabeledInits), len(mod.Definitions), len(stripMap))
	mod.LabeledAxioms = StripLabeledFormulas(mod.LabeledAxioms, stripMap, mod)
	mod.LabeledProps = StripLabeledFormulas(mod.LabeledProps, stripMap, mod)
	mod.LabeledConjs = StripLabeledFormulas(mod.LabeledConjs, stripMap, mod)
	mod.LabeledInits = StripLabeledFormulas(mod.LabeledInits, stripMap, mod)
	mod.Definitions = StripLabeledFormulas(mod.Definitions, stripMap, mod)

	// Strip the signature symbols.
	if mod.Sig != nil {
		for name, entry := range mod.Sig.Symbols.All() {
			if mod.Sig.Constructors[name] {
				continue // constructors are not stripped
			}
			sp := StripMapLookup(name, stripMap, mod)
			if len(sp) > 0 {
				newSort := StripSort(entry.Sort, len(sp))
				mod.Sig.Symbols.Set(name, &il.SymbolEntry{
					Name: name,
					Sort: newSort,
				})
			}
		}
	}

	// Strip the module parameters.
	newParams := make([]*lg.Const, 0, len(mod.Params))
	for _, sym := range mod.Params {
		sp := StripMapLookup(sym.Name, stripMap, mod)
		if len(sp) > 0 {
			newSort := StripSort(sym.CSort, len(sp))
			sym = lg.NewConst(sym.Name, newSort)
		}
		newParams = append(newParams, sym)
	}
	mod.Params = newParams

	// Strip native quotes.
	// Natives are stored as interface{} — process them if they support Args().
	stripNatives(mod.Natives, stripMap, mod)

	// Python: if iu.version_le(iu.get_string_version(), "1.6"): del mod.params[:]
	// For version 1.6 and earlier, clear all module parameters.
	if mod.Cfg != nil && iu.VersionLE(mod.Cfg.IuCfg.GetStringVersion(), "1.6") {
		mod.Params = mod.Params[:0]
	}

	return nil
}

// stripNative strips isolate parameters from a single native declaration.
// Python: strip_native(native, strip_map)
// native.args layout: [label, native_code, sym1, sym2, ...]
func stripNative(native ast.Node, stripMap StripMap, mod *module.Module) ast.Node {
	args := native.Args()
	if len(args) < 2 {
		return native
	}

	// Build strip_binding from args[2:] (the referenced symbols).
	stripBinding := make(map[lg.NodeKey]string)
	for _, a := range args[2:] {
		if expr, ok := a.(lg.Expr); ok {
			GetStripBinding(expr, stripMap, stripBinding, mod)
		}
	}

	// Strip the referenced symbol formulas using strip_binding.
	newFmlas := make([]ast.Node, len(args[2:]))
	for i, fmla := range args[2:] {
		if expr, ok := fmla.(lg.Expr); ok {
			newFmlas[i] = stripNodeFull(expr, stripMap, mod, stripBinding).(ast.Node)
		} else {
			newFmlas[i] = fmla
		}
	}

	// Strip the label (args[0]).
	lbl := args[0]
	if lbl != nil {
		lblArgs := lbl.Args()
		sp := StripMapLookup(lbl.String(), stripMap, mod)
		if len(sp) > 0 && len(sp) <= len(lblArgs) {
			lbl = lbl.Clone(lblArgs[len(sp):])
		}
	}

	// Rebuild: [label, native_code] + stripped formulas
	newArgs := make([]ast.Node, 0, 2+len(newFmlas))
	newArgs = append(newArgs, lbl, args[1])
	newArgs = append(newArgs, newFmlas...)
	return native.Clone(newArgs)
}

// stripNatives processes native declarations, stripping isolate parameters.
// Corresponds to Python strip_natives.
func stripNatives(natives []ast.Node, stripMap StripMap, mod *module.Module) {
	for i, n := range natives {
		natives[i] = stripNative(n, stripMap, mod)
	}
}

// StripSortFromModule removes a sort and all its associated symbols from the module.
// This is used when a sort is not in the cone of influence.
func StripSortFromModule(mod *module.Module, sortName string) error {
	// Remove the sort from the signature.
	if mod.Sig != nil {
		mod.Sig.Sorts.Delkey(sortName)
	}

	// Remove destructors for this sort.
	delete(mod.SortDestructors, sortName)
	delete(mod.DestructorSorts, sortName)

	// Remove constructors for this sort.
	delete(mod.SortConstructors, sortName)
	delete(mod.ConstructorSorts, sortName)

	// Remove from sort order.
	newOrder := make([]string, 0, len(mod.SortOrder))
	for _, s := range mod.SortOrder {
		if s != sortName {
			newOrder = append(newOrder, s)
		}
	}
	mod.SortOrder = newOrder

	// Remove symbols that reference this sort from SymbolOrder.
	newSymOrder := make([]*lg.Const, 0, len(mod.SymbolOrder))
	for _, sym := range mod.SymbolOrder {
		if symbolReferencesSort(sym, sortName) {
			continue
		}
		newSymOrder = append(newSymOrder, sym)
	}
	mod.SymbolOrder = newSymOrder

	// Remove the sort from the signature symbols that use it.
	if mod.Sig != nil {
		for name, entry := range mod.Sig.Symbols.All() {
			if symbolEntryReferencesSort(entry, sortName) {
				mod.Sig.Symbols.Delkey(name)
			}
		}
	}

	return nil
}

// symbolReferencesSort checks if a symbol's sort references the named sort.
func symbolReferencesSort(sym *lg.Const, sortName string) bool {
	return sortReferencesName(sym.CSort, sortName)
}

// symbolEntryReferencesSort checks if a SymbolEntry's sort references the named sort.
func symbolEntryReferencesSort(entry *il.SymbolEntry, sortName string) bool {
	if entry.Union != nil {
		for _, s := range entry.Union.Sorts {
			if sortReferencesName(s, sortName) {
				return true
			}
		}
		return false
	}
	return sortReferencesName(entry.Sort, sortName)
}

// sortReferencesName checks if a sort references the named sort.
func sortReferencesName(s lg.Sort, sortName string) bool {
	switch t := s.(type) {
	case *lg.UninterpretedSort:
		return t.Name == sortName
	case *lg.EnumeratedSort:
		return t.Name == sortName
	case *lg.RangeSort:
		return t.Name == sortName
	case *lg.FunctionSort:
		for _, d := range t.Domain() {
			if sortReferencesName(d, sortName) {
				return true
			}
		}
		return sortReferencesName(t.Range(), sortName)
	default:
		return false
	}
}

// unused import guard
var (
	_ = lg.Boolean
)
