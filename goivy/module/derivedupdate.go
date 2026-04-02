package module

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	co "github.com/glycerine/ivy/goivy/clauseops"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
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
	if c, ok := a.Symbol.(*lg.Const); ok {
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
	if c, ok := node.(*lg.Const); ok {
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

// --- ast.Node + lg.Expr methods for DerivedUpdate ---

func (a *DerivedUpdate) Args() []ast.Node {
	args := a.ActionArgs()
	nodes := make([]ast.Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}
func (a *DerivedUpdate) Clone(args []ast.Node) ast.Node {
	exprs := make([]lg.Expr, len(args))
	for i, n := range args {
		exprs[i] = n.(lg.Expr)
	}
	return a.ActionClone(exprs).(ast.Node)
}
func (a *DerivedUpdate) Children() []lg.Expr          { return a.ActionArgs() }
func (a *DerivedUpdate) NodeSort() lg.Sort            { return lg.ActionS }
func (a *DerivedUpdate) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *DerivedUpdate) GetAstConfig() *ast.AstConfig { return nil }
func (a *DerivedUpdate) Sexp() lg.NodeKey {
	// Python DerivedUpdate stores only defn (not a separate symbol field).
	// Match Python's format: (DerivedUpdate defn:<defn.sexp()>)
	defn := "nil"
	if a.Defn != nil {
		defn = string(a.Defn.Sexp())
	}
	return lg.NodeKey(fmt.Sprintf("(DerivedUpdate defn:%v)", defn))
}
func (a *DerivedUpdate) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }
