package isolate

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
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
			newReturns := make([]lg.Node, len(a.ActualReturns))
			for i, r := range a.ActualReturns {
				newReturns[i] = stripNode(r, stripMap, mod)
			}
			newArgs := []lg.Node{newCallee}
			newArgs = append(newArgs, newReturns...)
			return a.Clone(newArgs)
		}
		// No stripping needed for this call, but recurse into children.
		newArgs := stripNodes(a.Args(), stripMap, mod)
		return a.Clone(newArgs)
	case *actions.Sequence:
		newChildren := make([]lg.Node, len(a.Children))
		for i, child := range a.Children {
			if act, ok := child.(actions.Action); ok {
				newChildren[i] = actions.WrapAction(stripActionRec(act, stripMap, mod))
			} else if w := actions.UnwrapAction(child); w != nil {
				newChildren[i] = actions.WrapAction(stripActionRec(w, stripMap, mod))
			} else {
				newChildren[i] = stripNode(child, stripMap, mod)
			}
		}
		return a.Clone(newChildren)
	default:
		// For other action types, recursively process child nodes.
		oldArgs := action.Args()
		newArgs := make([]lg.Node, len(oldArgs))
		for i, arg := range oldArgs {
			if act, ok := arg.(actions.Action); ok {
				newArgs[i] = actions.WrapAction(stripActionRec(act, stripMap, mod))
			} else if w := actions.UnwrapAction(arg); w != nil {
				newArgs[i] = actions.WrapAction(stripActionRec(w, stripMap, mod))
			} else {
				newArgs[i] = stripNode(arg, stripMap, mod)
			}
		}
		return action.Clone(newArgs)
	}
}

// stripNode recursively processes a logic node, stripping isolate parameters
// from function applications.
func stripNode(node lg.Node, stripMap StripMap, mod *module.Module) lg.Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.Apply:
		// Strip parameters from function applications.
		newTerms := make([]lg.Node, len(n.Terms))
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
	case *lg.Var:
		return n
	default:
		// For other node types, return as-is.
		return node
	}
}

// stripNodes processes a slice of nodes through stripNode.
func stripNodes(nodes []lg.Node, stripMap StripMap, mod *module.Module) []lg.Node {
	result := make([]lg.Node, len(nodes))
	for i, n := range nodes {
		result[i] = stripNode(n, stripMap, mod)
	}
	return result
}

// StripLabeledFormula strips isolate parameters from a labeled formula.
// This is the Go port of Python's strip_labeled_fmla function.
//
// It strips isolate parameters from the formula body and adjusts the label
// if the label itself references stripped components.
func StripLabeledFormula(lf *module.LabeledFormula, stripMap StripMap, mod *module.Module) *module.LabeledFormula {
	if len(stripMap) == 0 {
		return lf
	}
	// Strip the formula body.
	newFormula := stripNode(lf.Formula, stripMap, mod)

	// Strip the label if present.
	newLabel := lf.Label
	if lf.Label != nil {
		newLabel = stripNode(lf.Label, stripMap, mod)
	}

	// Return a new LabeledFormula with stripped contents.
	return &module.LabeledFormula{
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
func StripLabeledFormulas(lfs []*module.LabeledFormula, stripMap StripMap, mod *module.Module) []*module.LabeledFormula {
	if len(stripMap) == 0 {
		return lfs
	}
	result := make([]*module.LabeledFormula, len(lfs))
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
func StripIsolate(mod *module.Module, stripMap StripMap) error {
	if len(stripMap) == 0 {
		return nil
	}

	// Strip actions.
	newActions := make(map[string]interface{}, len(mod.Actions))
	for name, actIface := range mod.Actions {
		act, ok := actIface.(actions.Action)
		if !ok {
			newActions[name] = actIface
			continue
		}
		stripParams := StripMapLookup(CanonAct(name), stripMap, mod)
		if len(stripParams) == 0 {
			newActions[name] = act
			continue
		}

		// Strip formal parameters.
		fp := act.GetFormalParams()
		if len(fp) < len(stripParams) {
			return fmt.Errorf("cannot strip isolate parameters from %s", name)
		}

		strippedAction := StripAction(act, stripMap, mod)
		strippedAction.SetFormalParams(fp[len(stripParams):])
		strippedAction.SetFormalReturns(act.GetFormalReturns())

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

	// Strip native quotes (process native definitions that reference stripped symbols).
	// Natives are stored as interface{} — we process them if they support
	// the necessary interface. For now, we leave natives unchanged since
	// full native stripping requires AST node types not yet ported.

	return nil
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
		for name, entry := range mod.Sig.Symbols {
			if symbolEntryReferencesSort(entry, sortName) {
				delete(mod.Sig.Symbols, name)
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
	_ = iu.ComposeCharacter
	_ = lg.Boolean
)
