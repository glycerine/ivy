package module

import (
	"fmt"

	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
)

// --- DerivedUpdate ---

// DerivedUpdate updates a derived relation based on its definition.
type DerivedUpdate struct {
	ActionBase
	Symbol lg.Expr // the derived symbol
	Defn   lg.Expr // the definition formula
}

func NewDerivedUpdate(sym, defn lg.Expr) *DerivedUpdate {
	return &DerivedUpdate{Symbol: sym, Defn: defn}
}

func (a *DerivedUpdate) Name() string          { return "derived_update" }
func (a *DerivedUpdate) ActionArgs() []lg.Expr { return []lg.Expr{a.Symbol, a.Defn} }
func (a *DerivedUpdate) ActionClone(args []lg.Expr) Action {
	r := &DerivedUpdate{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Symbol = args[0]
	}
	if len(args) >= 2 {
		r.Defn = args[1]
	}
	return r
}
func (a *DerivedUpdate) String() string {
	return fmt.Sprintf("derived(%s)", a.Symbol)
}
func (a *DerivedUpdate) IterCalls() []string      { return nil }
func (a *DerivedUpdate) IterSubactions() []Action { return DefaultIterSubactions(a) }

// GetUpdateAxioms checks if any dependency of the definition is in the updated
// set. If so, adds the defined symbol to updated. Returns (updated, nil, nil).
// Corresponds to Python DerivedUpdate.get_update_axioms.
func (a *DerivedUpdate) GetUpdateAxioms(updated []string, action Action) ([]string, *co.Clauses, *co.Clauses) {
	// Get the defined symbol name
	defines := ""
	if c, ok := a.Symbol.(*lg.Symbol); ok {
		defines = c.Name
	}
	if defines == "" {
		return updated, nil, nil
	}

	// Collect dependency symbols from the definition RHS
	deps := make(map[string]bool)
	CollectSymNames(a.Defn, deps)

	// Check if defines is not in updated and any dependency is in updated
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u] = true
	}
	if !updatedSet[defines] {
		for _, u := range updated {
			if deps[u] {
				updated = append(updated, defines)
				break
			}
		}
	}
	return updated, nil, nil
}

// collectSymNames collects constant/symbol names from a logic node.
// Explicitly walks Apply.Func since Children() returns Terms only.
func CollectSymNames(node lg.Expr, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Symbol); ok {
		names[c.Name] = true
	}
	if app, ok := node.(*lg.Apply); ok {
		CollectSymNames(app.Func, names)
	}
	for _, child := range node.Children() {
		CollectSymNames(child, names)
	}
}

func (a *DerivedUpdate) Decompose() [][]Action { return [][]Action{{a}} }
