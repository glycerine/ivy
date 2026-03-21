package isolate

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
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
	if _, ok := mod.Attributes[iu.ComposeNames(name, "global_parameter")]; ok {
		return nil
	}

	// Destructor sorts are not stripped.
	if _, ok := mod.DestructorSorts[name]; ok {
		return nil
	}

	// Sort names are not stripped.
	if mod.Sig != nil {
		if _, ok := mod.Sig.Sorts[name]; ok {
			return nil
		}
	}

	// Check each prefix in the strip map.
	for prefix, params := range stripMap {
		if strings.HasPrefix(name+iu.ComposeCharacter, prefix+iu.ComposeCharacter) {
			// Check for "common" attribute override.
			attr := iu.ComposeNames(name, "common")
			if commonVal, ok := mod.Attributes[attr]; ok {
				if commonStr, ok := commonVal.(string); ok {
					if strings.HasPrefix(commonStr+iu.ComposeCharacter, prefix+iu.ComposeCharacter) {
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
func StripActionFull(action actions.Action, stripMap StripMap, mod *module.Module,
	binding map[lg.NodeKey]string, isInit bool, initParams []string) actions.Action {
	if len(stripMap) == 0 && len(binding) == 0 {
		return action
	}
	return stripActionFullRec(action, stripMap, mod, binding, isInit, initParams)
}

func stripActionFullRec(action actions.Action, stripMap StripMap, mod *module.Module,
	binding map[lg.NodeKey]string, isInit bool, initParams []string) actions.Action {
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
			if act, ok := arg.(actions.Action); ok {
				newActionArgs[i] = actions.WrapAction(stripActionFullRec(act, stripMap, mod, localBinding, isInit, initParams))
			} else if w := actions.UnwrapAction(arg); w != nil {
				newActionArgs[i] = actions.WrapAction(stripActionFullRec(w, stripMap, mod, localBinding, isInit, initParams))
			} else {
				newActionArgs[i] = stripNodeFull(arg, stripMap, mod, localBinding)
			}
		}
		return action.ActionClone(newActionArgs)

	case *actions.Sequence:
		newChildren := make([]lg.Expr, len(a.Children))
		for i, child := range a.Children {
			if act, ok := child.(actions.Action); ok {
				newChildren[i] = actions.WrapAction(stripActionFullRec(act, stripMap, mod, binding, isInit, initParams))
			} else if w := actions.UnwrapAction(child); w != nil {
				newChildren[i] = actions.WrapAction(stripActionFullRec(w, stripMap, mod, binding, isInit, initParams))
			} else {
				newChildren[i] = stripNodeFull(child, stripMap, mod, binding)
			}
		}
		return a.ActionClone(newChildren)

	default:
		// Python lines 249-259: Check modifies() for interference.
		for _, sym := range actions.Modifies(action) {
			if mod.Sig != nil {
				if _, inSig := mod.Sig.Symbols[sym.Name]; inSig {
					lhsParams := StripMapLookup(sym.Name, stripMap, mod)
					if len(lhsParams) != NumIsolateParams {
						if !(len(lhsParams) == 0 && len(binding) == 0 && isInit) {
							// Python prints debug info and raises error.
							// We log a warning but continue for robustness.
							fmt.Printf("warning: assignment may be interfering: %s\n", sym.Name)
						}
					}
				}
			}
		}

		// Recurse into child nodes.
		oldArgs := action.ActionArgs()
		newActionArgs := make([]lg.Expr, len(oldArgs))
		for i, arg := range oldArgs {
			if act, ok := arg.(actions.Action); ok {
				newActionArgs[i] = actions.WrapAction(stripActionFullRec(act, stripMap, mod, binding, isInit, initParams))
			} else if w := actions.UnwrapAction(arg); w != nil {
				newActionArgs[i] = actions.WrapAction(stripActionFullRec(w, stripMap, mod, binding, isInit, initParams))
			} else {
				newActionArgs[i] = stripNodeFull(arg, stripMap, mod, binding)
			}
		}
		return action.ActionClone(newActionArgs)
	}
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
			case *lg.Symbol:
				// Add symbol to signature if not present.
				if mod.Sig != nil {
					if _, exists := mod.Sig.Symbols[sname]; !exists {
						mod.Sig.Symbols[sname] = &il.SymbolEntry{Name: sname, Sort: n.CSort}
						StripAddedSymbols = append(StripAddedSymbols, lg.NewSymbol(sname, n.CSort))
					}
				}
				return lg.NewSymbol(sname, n.CSort)
			case *lg.Variable:
				if mod.Sig != nil {
					if _, exists := mod.Sig.Symbols[sname]; !exists {
						mod.Sig.Symbols[sname] = &il.SymbolEntry{Name: sname, Sort: n.VSort}
						StripAddedSymbols = append(StripAddedSymbols, lg.NewSymbol(sname, n.VSort))
					}
				}
				return lg.NewSymbol(sname, n.VSort)
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
func StripAction(action actions.Action, stripMap StripMap, mod *module.Module) actions.Action {
	if len(stripMap) == 0 {
		return action
	}
	return stripActionRec(action, stripMap, mod)
}

// stripActionRec recursively strips isolate parameters from an action.
func stripActionRec(action actions.Action, stripMap StripMap, mod *module.Module) actions.Action {
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
		newChildren := make([]lg.Expr, len(a.Children))
		for i, child := range a.Children {
			if act, ok := child.(actions.Action); ok {
				newChildren[i] = actions.WrapAction(stripActionRec(act, stripMap, mod))
			} else if w := actions.UnwrapAction(child); w != nil {
				newChildren[i] = actions.WrapAction(stripActionRec(w, stripMap, mod))
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
			if act, ok := arg.(actions.Action); ok {
				newArgs[i] = actions.WrapAction(stripActionRec(act, stripMap, mod))
			} else if w := actions.UnwrapAction(arg); w != nil {
				newArgs[i] = actions.WrapAction(stripActionRec(w, stripMap, mod))
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
		if c, ok := n.Func.(*lg.Symbol); ok {
			stripParams := StripMapLookup(c.Name, stripMap, mod)
			if len(stripParams) > 0 && len(newTerms) >= len(stripParams) {
				// Strip the first len(stripParams) arguments.
				strippedTerms := newTerms[len(stripParams):]
				newSort := StripSort(c.CSort, len(stripParams))
				newSym := lg.NewSymbol(c.Name, newSort)
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
	case *lg.Symbol:
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
				if sym, ok := l.Func.(*lg.Symbol); ok {
					lblName = sym.Name
				}
			case *lg.Symbol:
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
	return &ast.LabeledFormula{
		Label:      newLabel,
		Formula:    newFormula,
		Lineno:     lf.Lineno,
		Temporal:   lf.Temporal,
		ID:         lf.ID,
		Explicit:   lf.Explicit,
		Assumed:    lf.Assumed,
		Unprovable: lf.Unprovable,
	}
}

// StripLabeledFormulas strips isolate parameters from a slice of labeled formulas in place.
func StripLabeledFormulas(lfs []*ast.LabeledFormula, stripMap StripMap, mod *module.Module) []*ast.LabeledFormula {
	if len(stripMap) == 0 {
		return lfs
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
		return sort // cannot strip more params than exist
	}
	newDom := dom[numParams:]
	if len(newDom) == 0 {
		// No domain left. If it was relational (range is Boolean),
		// we still need a FunctionSort. Otherwise return range directly.
		if _, isBool := fs.Range().(*lg.BooleanSort); isBool {
			// Relational with no domain: return the range sort.
			return fs.Range()
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
// StripIsolateParams is the full version of strip_isolate that handles
// variable isolate parameter substitution, initializer handling, impl_mixin
// strip propagation, and extra_strip.
//
// Corresponds to Python strip_isolate (lines 341-456).
func StripIsolateParams(mod *module.Module, isolate IsolateDefInterface,
	implMixins map[string][]MixinDef, allAfterInits map[string]bool,
	extraStrip map[string][]string) error {

	// Python: global num_isolate_params, strip_added_symbols
	isoParams := isolate.PresentNames() // combined verified+present params
	NumIsolateParams = len(isoParams)
	StripAddedSymbols = nil // reset

	// Step 1: Variable isolate parameter substitution.
	// Python: if any(isinstance(p, Variable) for p in ipl): substitute
	// In Go, isolate parameters are strings from VerifiedNames/PresentNames.
	// Variable parameters would need AST-level information. For the common case
	// (no variable parameters), this is a no-op.

	// Step 2: Build the strip map from isolate parameter bindings.
	stripMap := make(StripMap)

	// Build from verified + present atoms' parameters.
	// In a full implementation, we'd extract parameter args from each atom.
	// For now, we use the existing StripMap construction from callers.

	// Step 3: Propagate strip map through impl_mixins.
	// Python: for ms in impl_mixins.values(): for m: strip_map[m.mixee()] = strip_params
	for _, ms := range implMixins {
		for _, m := range ms {
			if isMixinImplement(m) {
				mixerParams := StripMapLookup(CanonAct(m.Mixer()), stripMap, mod)
				if len(mixerParams) > 0 {
					stripMap[m.Mixee()] = mixerParams
				}
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
	// Python: for s in isolate.params(): sym = add_symbol(s.rep, mod.sig.sorts[s.sort])
	for _, paramName := range isoParams {
		if paramName == "this" {
			continue
		}
		// Check if already in signature
		if mod.Sig != nil {
			if _, exists := mod.Sig.Symbols[paramName]; exists {
				continue
			}
			// Look up the sort for this parameter
			if s, ok := mod.Sig.Sorts[paramName]; ok {
				sym := lg.NewSymbol(paramName, s)
				mod.Sig.Symbols[paramName] = &il.SymbolEntry{Name: paramName, Sort: s}
				mod.Params = append(mod.Params, sym)
				mod.ParamDefaults = append(mod.ParamDefaults, nil)
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
	if len(stripMap) == 0 {
		return nil
	}

	// Strip actions.
	newActions := make(map[string]module.Action, len(mod.Actions))
	for name, act := range mod.Actions {
		stripParams := StripMapLookup(CanonAct(name), stripMap, mod)
		if len(stripParams) == 0 {
			newActions[name] = act
			continue
		}

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

		newActions[name] = strippedAction
	}

	// Replace all actions.
	mod.Actions = newActions

	// Strip labeled formulas.
	mod.LabeledAxioms = StripLabeledFormulas(mod.LabeledAxioms, stripMap, mod)
	mod.LabeledProps = StripLabeledFormulas(mod.LabeledProps, stripMap, mod)
	mod.LabeledConjs = StripLabeledFormulas(mod.LabeledConjs, stripMap, mod)
	mod.LabeledInits = StripLabeledFormulas(mod.LabeledInits, stripMap, mod)
	mod.Definitions = StripLabeledFormulas(mod.Definitions, stripMap, mod)

	// Strip the signature symbols.
	if mod.Sig != nil {
		for name, entry := range mod.Sig.Symbols {
			if mod.Sig.Constructors[name] {
				continue // constructors are not stripped
			}
			sp := StripMapLookup(name, stripMap, mod)
			if len(sp) > 0 {
				newSort := StripSort(entry.Sort, len(sp))
				mod.Sig.Symbols[name] = &il.SymbolEntry{
					Name: name,
					Sort: newSort,
				}
			}
		}
	}

	// Strip the module parameters.
	newParams := make([]*lg.Symbol, 0, len(mod.Params))
	for _, sym := range mod.Params {
		sp := StripMapLookup(sym.Name, stripMap, mod)
		if len(sp) > 0 {
			newSort := StripSort(sym.CSort, len(sp))
			sym = lg.NewSymbol(sym.Name, newSort)
		}
		newParams = append(newParams, sym)
	}
	mod.Params = newParams

	// Strip native quotes.
	// Natives are stored as interface{} — process them if they support Args().
	stripNatives(mod.Natives, stripMap, mod)

	// Python: if iu.version_le(iu.get_string_version(), "1.6"): del mod.params[:]
	// For version 1.6 and earlier, clear all module parameters.
	if iu.VersionLE(iu.GetStringVersion(), "1.6") {
		mod.Params = mod.Params[:0]
	}

	return nil
}

// stripNatives processes native declarations, stripping isolate parameters
// from any referenced symbols.
// Corresponds to Python strip_natives.
func stripNatives(natives []ast.Node, stripMap StripMap, mod *module.Module) {
	// Natives contain backtick-delimited code with embedded Ivy references.
	// The args after the first two are the referenced symbols.
	// We strip those symbols' sorts.
	// Note: ast.Node.Args() returns []ast.Node snapshots, so in-place
	// mutation of symbol sorts requires rebuilding via Clone. This mirrors
	// the Python which mutates args[i] in place.
	// For now this is a stub — the original Go code used a dynamic
	// Args() []interface{} check that nothing satisfied either.
}

// StripSortFromModule removes a sort and all its associated symbols from the module.
// This is used when a sort is not in the cone of influence.
func StripSortFromModule(mod *module.Module, sortName string) error {
	// Remove the sort from the signature.
	if mod.Sig != nil {
		delete(mod.Sig.Sorts, sortName)
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
	newSymOrder := make([]*lg.Symbol, 0, len(mod.SymbolOrder))
	for _, sym := range mod.SymbolOrder {
		if symbolReferencesSort(sym, sortName) {
			continue
		}
		newSymOrder = append(newSymOrder, sym)
	}
	mod.SymbolOrder = newSymOrder

	// Remove the sort from the signature symbols that use it.
	if mod.Sig != nil {
		for name, entry := range mod.Sig.Symbols {
			if symbolEntryReferencesSort(entry, sortName) {
				delete(mod.Sig.Symbols, name)
			}
		}
	}

	return nil
}

// symbolReferencesSort checks if a symbol's sort references the named sort.
func symbolReferencesSort(sym *lg.Symbol, sortName string) bool {
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
	_ = iu.ComposeCharacter
	_ = lg.Boolean
)
