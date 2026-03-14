package isolate

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
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
// TODO: Full implementation requires:
//   - ivy_logic.is_app / is_constant / is_variable checks
//   - AST clone operations
//   - Symbol sort manipulation
//   - Currently stubbed to return the action unchanged.
func StripAction(action actions.Action, stripMap StripMap, mod *module.Module) actions.Action {
	if len(stripMap) == 0 {
		return action
	}
	// TODO: implement recursive stripping of isolate parameters
	// For now, return the action unchanged.
	return action
}

// StripLabeledFormula strips isolate parameters from a labeled formula.
//
// TODO: Full implementation requires AST manipulation infrastructure.
func StripLabeledFormula(lf *module.LabeledFormula, stripMap StripMap, mod *module.Module) *module.LabeledFormula {
	if len(stripMap) == 0 {
		return lf
	}
	// TODO: implement formula stripping
	return lf
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
//
// TODO: Full implementation requires:
//   - Isolate definition AST types with params/verified/present methods
//   - Full strip_action recursive rewriting
//   - Signature manipulation
//   - Currently partially implemented.
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

	// TODO: strip the signature symbols
	// TODO: strip the module parameters
	// TODO: strip native quotes

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
	// TODO: This requires checking symbol sorts, which needs more type info.

	return nil
}

// unused import guard
var (
	_ = iu.ComposeCharacter
	_ = lg.Boolean
)
