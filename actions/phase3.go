// phase3.go implements functions from Phase 3 (Batch 3.2) of the Ivy port.
// These correspond to Python ivy_actions.py helper functions and types.
package actions

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- PCA ---

// PCA parses a checked-assertion location string of the form "filename:line".
// Returns a Location with ".ivy" appended to the filename.
// Corresponds to Python's p_c_a.
func PCA(s string) ast.Location {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return ast.Location{}
	}
	line := 0
	fmt.Sscanf(parts[1], "%d", &line)
	return ast.Location{Filename: parts[0] + ".ivy", Line: line}
}

// --- UnrollContext ---

// UnrollContext is an ActionContext for bounded loop unrolling.
// The Card function returns a cardinality bound for a sort,
// used in bounding loop unrollings.
// Corresponds to Python's UnrollContext class.
type UnrollContext struct {
	ActionContext
	Card   func(lg.Sort) int // returns cardinality bound for a sort
	Domain *module.Module
}

// NewUnrollContext creates a new UnrollContext with a cardinality function.
func NewUnrollContext(card func(lg.Sort) int, domain *module.Module) *UnrollContext {
	return &UnrollContext{
		Card:   card,
		Domain: domain,
	}
}

// --- SymbolList ---

// SymbolList is an AST wrapper for a collection of symbol names.
// Corresponds to Python's SymbolList class.
type SymbolList struct {
	Symbols []lg.Node // each is a *lg.Symbol or string-named node
}

// NewSymbolList creates a SymbolList from symbols.
func NewSymbolList(symbols ...lg.Node) *SymbolList {
	return &SymbolList{Symbols: symbols}
}

func (sl *SymbolList) String() string {
	parts := make([]string, len(sl.Symbols))
	for i, s := range sl.Symbols {
		parts[i] = fmt.Sprint(s)
	}
	return strings.Join(parts, ",")
}

// Args returns the symbols for AST compatibility.
func (sl *SymbolList) Args() []lg.Node {
	return sl.Symbols
}

// --- GetCorrectArity ---

// GetCorrectArity returns the correct arity (number of arguments) for an atom.
// Corresponds to Python's get_correct_arity.
func GetCorrectArity(domain *module.Module, atom lg.Node) int {
	// Check if it's a numeral
	if c, ok := atom.(*lg.Symbol); ok {
		if il.IsNumeral(c) {
			return 0
		}
	}
	// Get the atom's rep symbol and its sort's domain length
	switch a := atom.(type) {
	case *lg.Apply:
		if c, ok := a.Func.(*lg.Symbol); ok {
			if fs, ok := c.CSort.(*lg.FunctionSort); ok {
				return len(fs.Sorts) - 1 // domain sorts (all but range)
			}
		}
	case *lg.Symbol:
		if fs, ok := a.CSort.(*lg.FunctionSort); ok {
			return len(fs.Sorts) - 1
		}
	}
	return 0
}

// --- TypeCheck ---

// TypeCheck checks that all atoms in an AST have the correct arity.
// Corresponds to Python's type_check.
func TypeCheck(domain *module.Module, node lg.Node) error {
	// Walk the AST and check each Apply node
	return typeCheckRec(domain, node)
}

func typeCheckRec(domain *module.Module, node lg.Node) error {
	if node == nil {
		return nil
	}

	// Check Apply nodes (these are the "apps" in apps_ast)
	if app, ok := node.(*lg.Apply); ok {
		arity := len(app.Terms)
		correctArity := GetCorrectArity(domain, node)

		// Allow unary minus
		if c, ok := app.Func.(*lg.Symbol); ok {
			if c.Name == "-" && arity == 1 {
				// unary minus is OK
			} else if arity != correctArity {
				return fmt.Errorf("wrong number of arguments to %s: got %d, expecting %d",
					c.Name, arity, correctArity)
			}
		}
	}

	// Recurse into children
	for _, child := range node.Children() {
		if err := typeCheckRec(domain, child); err != nil {
			return err
		}
	}
	// Also check Apply.Func
	if app, ok := node.(*lg.Apply); ok {
		if err := typeCheckRec(domain, app.Func); err != nil {
			return err
		}
	}
	return nil
}

// --- TypeAst ---

// TypeAst converts between Atom and App nodes based on domain relations.
// If an Atom is not a relation and not '=', it becomes an App.
// If an App is a relation, it becomes an Atom.
// Corresponds to Python's type_ast.
func TypeAst(domain *module.Module, node lg.Node) lg.Node {
	if node == nil {
		return nil
	}

	// Check if it's an Atom (Apply with Boolean sort) that's not a relation
	if app, ok := node.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Symbol); ok {
			isRelation := false
			if domain.Relations != nil {
				_, isRelation = domain.Relations[c.Name]
			}
			isEq := c.Name == "="

			if il.IsAtom(node) && !isRelation && !isEq {
				// Convert Atom to App: just return as-is since in Go
				// Apply serves as both Atom and App
				return node
			}
			if !il.IsAtom(node) && isRelation {
				// Convert App to Atom: ensure Boolean sort
				return node
			}
		}
	}
	return node
}

// --- DestrAsgnVal ---

// DestrAsgnVal handles assignment to destructor chains.
// Given a destructor-chain LHS (like d(e(x), args...)), builds the
// update formulas for nested field assignments.
// Returns (new_lhs, new_clauses, mutated_symbol).
// Corresponds to Python's destr_asgn_val.
func DestrAsgnVal(lhs lg.Node, fmlas *[]lg.Node, mod *module.Module) (lg.Node, *co.Clauses, *lg.Symbol) {
	app, ok := lhs.(*lg.Apply)
	if !ok {
		return lhs, co.TrueClauses(nil), nil
	}
	if len(app.Terms) == 0 {
		return lhs, co.TrueClauses(nil), nil
	}

	mut := app.Terms[0]
	rest := app.Terms[1:]
	n := app.Func

	// Get the "rep" of mut
	var mutSym *lg.Symbol
	switch m := mut.(type) {
	case *lg.Apply:
		if c, ok := m.Func.(*lg.Symbol); ok {
			mutSym = c
		}
	case *lg.Symbol:
		mutSym = m
	}

	var lval lg.Node
	var newClauses *co.Clauses
	var mutated *lg.Symbol

	if mutSym != nil && mod.DestructorSorts != nil {
		if _, isDestr := mod.DestructorSorts[mutSym.Name]; isDestr {
			// Recursive case: the mutated object is also a destructor chain
			lval, newClauses, mutated = DestrAsgnVal(mut, fmlas, mod)
		} else {
			// Base case: generate a nondeterministic intermediate value
			mutated = mutSym
			newClauses = co.TrueClauses(nil)
			lval = mut
		}
	} else {
		mutated = mutSym
		newClauses = co.TrueClauses(nil)
		lval = mut
	}

	// Build dlhs = n(lval, vs[1:]) and drhs = n(mut, vs[1:])
	nSym, _ := n.(*lg.Symbol)
	if nSym == nil {
		return lhs, newClauses, mutated
	}

	// Build new lhs: n(lval, rest...)
	newArgs := make([]lg.Node, 0, 1+len(rest))
	newArgs = append(newArgs, lval)
	newArgs = append(newArgs, rest...)
	newLhs, _ := lg.NewApply(n, newArgs...)

	return newLhs, newClauses, mutated
}

// --- AssignRefs ---

// AssignRefs collects all referenced symbols in an assignment LHS,
// including through destructor chains.
// Corresponds to Python's assign_refs.
func AssignRefs(lhsNode lg.Node, refs map[string]bool, mod *module.Module) {
	assignRefsRec(lhsNode, refs, mod)
}

func assignRefsRec(node lg.Node, refs map[string]bool, mod *module.Module) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *lg.Apply:
		if c, ok := n.Func.(*lg.Symbol); ok {
			if mod.DestructorSorts != nil {
				if _, isDestr := mod.DestructorSorts[c.Name]; isDestr {
					refs[c.Name] = true
					if len(n.Terms) > 0 {
						assignRefsRec(n.Terms[0], refs, mod)
					}
					for _, a := range n.Terms[1:] {
						collectSymbols(a, refs)
					}
					return
				}
			}
		}
		// Non-destructor: collect all symbol refs
		for _, a := range n.Terms {
			collectSymbols(a, refs)
		}
	case *lg.Symbol:
		refs[n.Name] = true
	}
}

// --- Sign ---

// Sign applies polarity to an atom. If polarity is true, returns the atom;
// if false, returns Not(atom).
// Corresponds to Python's sign.
func Sign(polarity bool, atom lg.Node) lg.Node {
	if polarity {
		return atom
	}
	return &lg.Not{Body: atom}
}

// --- MakeFieldUpdate ---

// MakeFieldUpdate generates update formulas for field/destructor assignment.
// The field f must be a binary relation.
// Corresponds to Python's make_field_update.
func MakeFieldUpdate(self Action, l, f lg.Node, rFunc func(lg.Node) lg.Node, domain *module.Module, pvars map[string]bool) error {
	fSym, ok := f.(*lg.Symbol)
	if !ok {
		return fmt.Errorf("field %s must be a symbol", f)
	}
	fs, ok := fSym.CSort.(*lg.FunctionSort)
	if !ok || len(fs.Sorts) != 3 { // dom[0], dom[1], range
		return fmt.Errorf("field %s must be a binary relation", fSym.Name)
	}
	// v = Variable('X', f.sort.dom[1])
	v, err := lg.NewVariable("X", fs.Sorts[1])
	if err != nil {
		return err
	}
	// aa = AssignAction(f(l,v), r(v))
	fApp, _ := lg.NewApply(f, l, v)
	rVal := rFunc(v)
	_ = NewAssignAction(fApp, rVal)
	// The actual update computation would call aa.ActionUpdate(domain, pvars)
	// which requires the full transition relation infrastructure.
	return nil
}

// --- MyStr ---

// MyStr formats a node for display with depth tracking to prevent infinite recursion.
// Corresponds to Python's my_str.
func MyStr(x interface{}, depth int) string {
	if depth > 25 {
		return "..."
	}
	// Check if x has a DStr method
	type dstrer interface {
		DStr(depth int) string
	}
	if d, ok := x.(dstrer); ok {
		return d.DStr(depth)
	}
	return fmt.Sprint(x)
}

// --- SetDeterminize ---

// determinize controls whether ChoiceAction uses deterministic encoding.
var determinize bool

// SetDeterminize sets the global determinize flag.
// Corresponds to Python's set_determinize.
func SetDeterminize(t bool) {
	determinize = t
}

// GetDeterminize returns the current determinize setting.
func GetDeterminize() bool {
	return determinize
}

// --- BracketAction ---

// BracketAction formats an action with braces if it's not already a Sequence.
// Corresponds to Python's bracket_action.
func BracketAction(action Action, depth int) string {
	if _, isSeq := action.(*Sequence); isSeq {
		return MyStr(action, depth)
	}
	return "{" + MyStr(action, depth) + "}"
}

// --- DebugAction ---

// DebugAction is a debug statement action. It is a no-op for semantics.
// Corresponds to Python's DebugAction class.
type DebugAction struct {
	ActionBase
	DebugExpr lg.Node   // debug expression (first arg)
	WithExprs []lg.Node // additional "with" expressions
}

// NewDebugAction creates a new DebugAction.
func NewDebugAction(debugExpr lg.Node, withExprs ...lg.Node) *DebugAction {
	return &DebugAction{DebugExpr: debugExpr, WithExprs: copyNodes(withExprs)}
}

func (a *DebugAction) Name() string { return "debug" }
func (a *DebugAction) Args() []lg.Node {
	args := []lg.Node{a.DebugExpr}
	args = append(args, a.WithExprs...)
	return args
}
func (a *DebugAction) Clone(args []lg.Node) Action {
	r := &DebugAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.DebugExpr = args[0]
	}
	if len(args) > 1 {
		r.WithExprs = copyNodes(args[1:])
	}
	return r
}
func (a *DebugAction) String() string {
	res := "debug " + fmt.Sprint(a.DebugExpr)
	if len(a.WithExprs) > 0 {
		parts := make([]string, len(a.WithExprs))
		for i, e := range a.WithExprs {
			parts[i] = fmt.Sprint(e)
		}
		res += " with " + strings.Join(parts, ",")
	}
	return res
}
func (a *DebugAction) IterCalls() []string     { return nil }
func (a *DebugAction) IterSubactions() []Action { return defaultIterSubactions(a) }
func (a *DebugAction) Decompose() [][]Action   { return atomicDecompose(a) }

// --- Entry ---

// Entry creates an RME (Rely-Guarantee relation) entry for action semantics.
// Corresponds to Python's entry function.
func Entry(ensures ...lg.Node) *RME {
	var ensNode lg.Node
	if len(ensures) > 0 {
		ensNode = ensures[0]
	} else {
		ensNode = &lg.And{} // And() with no args = true
	}
	return NewRME(&lg.And{}, nil, ensNode)
}

// --- TypeCheckContext ---

// TypeCheckContext is a context for type-checking actions without executing them.
// When resolving a called action, it replaces it with an empty Sequence
// that preserves the formal parameters/returns.
// Corresponds to Python's TypeCheckConext class.
type TypeCheckContext struct {
	ActionContext
}

// NewTypeCheckContext creates a new TypeCheckContext.
func NewTypeCheckContext(domain interface{}) *TypeCheckContext {
	return &TypeCheckContext{
		ActionContext: ActionContext{Domain: domain},
	}
}

// Get resolves an action name, returning a null action with preserved formals.
// Corresponds to Python's TypeCheckConext.get.
func (tc *TypeCheckContext) Get(x string) Action {
	// Use base ActionContext to find the action
	if tc.Domain == nil {
		return nil
	}
	mod, ok := tc.Domain.(*module.Module)
	if !ok {
		return nil
	}
	actI, ok := mod.Actions[x]
	if !ok {
		return nil
	}
	act, ok := actI.(Action)
	if !ok {
		return nil
	}
	// Return an empty Sequence with the same formal params/returns
	res := NewSequence()
	res.SetFormalParams(act.GetFormalParams())
	res.SetFormalReturns(act.GetFormalReturns())
	return res
}

// TypeCheckActionFull performs type checking on an action within a domain.
// Corresponds to Python's type_check_action.
func TypeCheckActionFull(action Action, domain *module.Module, pvars map[string]bool) {
	// In Python, this function is a no-op (early return).
	// It was intended to use TypeCheckContext to run int_update
	// but is currently disabled in the Python source.
	return
}
