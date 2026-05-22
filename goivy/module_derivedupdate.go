package goivy

import (
	"fmt"
)

// --- DerivedUpdate ---

// DerivedUpdate updates a derived relation based on its definition.
type DerivedUpdate struct {
	ActionBase
	Symbol Expr // the derived symbol
	Defn   Expr // the definition formula
}

func NewDerivedUpdate(sym, defn Expr) *DerivedUpdate {
	return &DerivedUpdate{Symbol: sym, Defn: defn}
}

func (a *DerivedUpdate) Name() string       { return "derived_update" }
func (a *DerivedUpdate) ActionArgs() []Expr { return []Expr{a.Symbol, a.Defn} }
func (a *DerivedUpdate) ActionClone(args []Expr) Action {
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
func (a *DerivedUpdate) GetUpdateAxioms(updated []*Const, action Action) ([]*Const, *Clauses, *Clauses) {
	// Get the defined symbol
	defines := a.Symbol
	if rep := IvyNodeRep(defines); rep != nil {
		defines = rep
	}
	defSym, ok := defines.(*Const)
	if !ok || defSym == nil {
		return updated, nil, nil
	}

	// Python DerivedUpdate uses used_symbols_ast(defn.args[1]), i.e. the RHS
	// dependencies only. Some local proof definitions pass an applied LHS as
	// Symbol, so the defined symbol is normalized above through IvyNodeRep.
	depExpr := a.Defn
	if args := a.Defn.Args(); len(args) > 1 {
		if rhs, ok := args[1].(Expr); ok {
			depExpr = rhs
		}
	}
	deps := make(map[string]bool)
	for _, sym := range UsedSymbolsAst(depExpr).All() {
		if c, ok := sym.(*Const); ok {
			deps[c.Name] = true
		}
	}

	// Check if defines is not in updated and any dependency is in updated
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u.Name] = true
	}
	if !updatedSet[defSym.Name] {
		for _, u := range updated {
			if deps[u.Name] {
				updated = append(updated, defSym)
				break
			}
		}
	}
	return updated, nil, nil
}

// collectSymNames collects constant/symbol names from a logic node.
// Explicitly walks Apply.Func since Children() returns Terms only.
func CollectSymNames(node Expr, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*Const); ok {
		names[c.Name] = true
	}
	if app, ok := node.(*Apply); ok {
		CollectSymNames(app.Func, names)
	}
	for _, child := range node.Children() {
		CollectSymNames(child, names)
	}
}

func (a *DerivedUpdate) Decompose() [][]Action { return [][]Action{{a}} }

// --- ast.Node + lg.Expr methods for DerivedUpdate ---

func (a *DerivedUpdate) Args() []Node {
	args := a.ActionArgs()
	nodes := make([]Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}
func (a *DerivedUpdate) Clone(args []Node) Node {
	exprs := make([]Expr, len(args))
	for i, n := range args {
		exprs[i] = n.(Expr)
	}
	return a.ActionClone(exprs).(Node)
}
func (a *DerivedUpdate) Children() []Expr         { return a.ActionArgs() }
func (a *DerivedUpdate) NodeSort() Sort           { return ActionS }
func (a *DerivedUpdate) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *DerivedUpdate) GetAstConfig() *AstConfig { return nil }
func (a *DerivedUpdate) Sexp() NodeKey {
	// Python DerivedUpdate stores only defn (not a separate symbol field).
	// Match Python's format: (DerivedUpdate defn:<defn.sexp()>)
	defn := "nil"
	if a.Defn != nil {
		defn = string(a.Defn.Sexp())
	}
	return NodeKey(fmt.Sprintf("(DerivedUpdate defn:%v)", defn))
}
func (a *DerivedUpdate) Canon() Canonical { return Canonical(a.Sexp()) }
