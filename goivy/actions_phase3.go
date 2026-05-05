// phase3.go implements functions from Phase 3 (Batch 3.2) of the Ivy port.
// These correspond to Python ivy_actions.py helper functions and types.
package goivy

import (
	"fmt"
	"strings"
)

// --- PCA ---

// PCA parses a checked-assertion location string of the form "filename:line".
// Returns a Location with ".ivy" appended to the filename.
// Corresponds to Python's p_c_a.
func PCA(s string) Location {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Location{}
	}
	line := 0
	fmt.Sscanf(parts[1], "%d", &line)
	return Location{Filename: parts[0] + ".ivy", Line: line}
}

// --- UnrollContext ---

// UnrollContext is an ActionContext for bounded loop unrolling.
// The Card function returns a cardinality bound for a sort,
// used in bounding loop unrollings.
// Corresponds to Python's UnrollContext class.
type UnrollContext struct {
	ActionContext
	Card func(Sort) int // returns cardinality bound for a sort
}

// NewUnrollContext creates a new UnrollContext with a cardinality function.
func NewUnrollContext(card func(Sort) int, domain *Module, cfg ...*ActionsConfig) *UnrollContext {
	var acfg *ActionsConfig
	if len(cfg) > 0 {
		acfg = cfg[0]
	}
	return &UnrollContext{
		ActionContext: ActionContext{Domain: domain, Cfg: acfg},
		Card:          card,
	}
}

// --- SymbolList ---

// SymbolList is an AST wrapper for a collection of symbol names.
// Corresponds to Python's SymbolList class.
type SymbolList struct {
	Symbols []Expr // each is a *lg.Const or string-named node
}

// NewSymbolList creates a SymbolList from symbols.
func NewSymbolList(symbols ...Expr) *SymbolList {
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
func (sl *SymbolList) Args() []Expr {
	return sl.Symbols
}

// --- GetCorrectArity ---

// GetCorrectArity returns the correct arity (number of arguments) for an atom.
// Corresponds to Python's get_correct_arity.
func GetCorrectArity(domain *Module, atom Expr) int {
	// Check if it's a numeral
	if c, ok := atom.(*Const); ok {
		if IsNumeral(c) {
			return 0
		}
	}
	// Get the atom's rep symbol and its sort's domain length
	switch a := atom.(type) {
	case *Apply:
		if c, ok := a.Func.(*Const); ok {
			if fs, ok := c.CSort.(*FunctionSort); ok {
				return len(fs.Sorts) - 1 // domain sorts (all but range)
			}
		}
	case *Const:
		if fs, ok := a.CSort.(*FunctionSort); ok {
			return len(fs.Sorts) - 1
		}
	}
	return 0
}

// --- TypeCheck ---

// TypeCheck checks that all atoms in an AST have the correct arity.
// Corresponds to Python's type_check.
func TypeCheck(domain *Module, node Expr) error {
	// Walk the AST and check each Apply node
	return typeCheckRec(domain, node)
}

func typeCheckRec(domain *Module, node Expr) error {
	if node == nil {
		return nil
	}

	// Check Apply nodes (these are the "apps" in apps_ast)
	if app, ok := node.(*Apply); ok {
		arity := len(app.Terms)
		correctArity := GetCorrectArity(domain, node)

		// Allow unary minus
		if c, ok := app.Func.(*Const); ok {
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
	if app, ok := node.(*Apply); ok {
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
func TypeAst(domain *Module, node Expr) Expr {
	if node == nil {
		return nil
	}

	// Check if it's an Atom (Apply with Boolean sort) that's not a relation
	if app, ok := node.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			isRelation := false
			if domain.Relations != nil {
				_, isRelation = domain.Relations.Get2(c.Name)
			}
			isEq := c.Name == "="

			if IsAtom(node) && !isRelation && !isEq {
				// Convert Atom to App: just return as-is since in Go
				// Apply serves as both Atom and App
				return node
			}
			if !IsAtom(node) && isRelation {
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
func DestrAsgnVal(lhs Expr, fmlas *[]Expr, m *Module) (Expr, *Clauses, *Const) {
	app, ok := lhs.(*Apply)
	if !ok {
		return lhs, TrueClauses(nil), nil
	}
	if len(app.Terms) == 0 {
		return lhs, TrueClauses(nil), nil
	}

	// Python: mut = lhs.args[0]; rest = list(lhs.args[1:]); mut_n = mut.rep
	mut := app.Terms[0]
	rest := app.Terms[1:]

	// Get the "rep" of mut (mut_n)
	var mutSym *Const
	switch m := mut.(type) {
	case *Apply:
		if c, ok := m.Func.(*Const); ok {
			mutSym = c
		}
	case *Const:
		mutSym = m
	}
	if mutSym == nil {
		return lhs, TrueClauses(nil), nil
	}

	var lval Expr
	var newClauses *Clauses
	var mutated *Const

	if m.DestructorSorts != nil {
		if _, isDestr := m.DestructorSorts[mutSym.Name]; isDestr {
			// Recursive case: mut_n.name in destructor_sorts
			lval, newClauses, mutated = DestrAsgnVal(mut, fmlas, m)
		} else {
			// Base case: nondet = mut_n.suffix("_nd").skolem()
			// Python: Symbol.suffix(s) → Symbol(name+s, sort)
			// Python: Symbol.skolem() → Symbol("__"+name, sort)
			nondetSym := NewConst("__"+mutSym.Name+"_nd", mutSym.CSort)

			// Python: new_clauses = mk_assign_clauses(mut_n, nondet(*sym_placeholders(mut_n)))
			// In Python, mk_assign_clauses takes a symbol-like lhs (mut_n) and rhs.
			// Go's mkAssignClauses returns *Update; we extract .TR (the Clauses).
			phs := SymPlaceholders(mutSym)
			phNodes := make([]Expr, len(phs))
			for i, v := range phs {
				phNodes[i] = v
			}
			var nondetApp Expr
			if len(phNodes) > 0 {
				nondetApp, _ = NewApply(nondetSym, phNodes...)
			} else {
				nondetApp = nondetSym
			}
			assignUpd := mkAssignClauses(mutSym, nondetApp)
			newClauses = assignUpd.TR

			// Python: lval = nondet(*mut.args)
			mutArgs := nodeArgs(mut)
			if len(mutArgs) > 0 {
				mutArgNodes := make([]Expr, len(mutArgs))
				copy(mutArgNodes, mutArgs)
				lval, _ = NewApply(nondetSym, mutArgNodes...)
			} else {
				lval = nondetSym
			}

			mutated = mutSym
		}
	} else {
		// No destructor sorts at all — base case with no skolem
		mutated = mutSym
		newClauses = TrueClauses(nil)
		lval = mut
	}

	// Python: n = lhs.rep
	n := app.Func
	nSym, _ := n.(*Const)
	if nSym == nil {
		return lhs, newClauses, mutated
	}

	// Python: vs = sym_placeholders(n)
	vs := SymPlaceholders(nSym)

	// Python: dlhs = n(*([lval] + vs[1:]))
	dlhsArgs := make([]Expr, 0, 1+len(vs))
	dlhsArgs = append(dlhsArgs, lval)
	if len(vs) > 1 {
		for _, v := range vs[1:] {
			dlhsArgs = append(dlhsArgs, v)
		}
	}
	dlhs := applyToNodes(nSym, dlhsArgs)

	// Python: drhs = n(*([mut] + vs[1:]))
	drhsArgs := make([]Expr, 0, 1+len(vs))
	drhsArgs = append(drhsArgs, mut)
	for _, v := range vs[1:] {
		drhsArgs = append(drhsArgs, v)
	}
	drhs := applyToNodes(nSym, drhsArgs)

	// Python: eqs = [eq_atom(v,a) for (v,a) in list(zip(vs,lhs.args))[1:] if not isinstance(a,Variable)]
	var eqs []Expr
	for i := 1; i < len(vs) && i < len(app.Terms); i++ {
		if _, isVar := app.Terms[i].(*Variable); !isVar {
			eqs = append(eqs, &Eq{T1: vs[i], T2: app.Terms[i]})
		}
	}

	// Python: if eqs: fmlas.append(Or(And(*eqs), equiv_ast(dlhs, drhs)))
	if len(eqs) > 0 {
		eqConj, _ := NewAnd(eqs...)
		equiv := equivAST(dlhs, drhs)
		guard, _ := NewOr(eqConj, equiv)
		*fmlas = append(*fmlas, guard)
	}

	// Python: for destr in ivy_module.module.sort_destructors[mut.sort.name]:
	//             if destr != n:
	//                 phs = sym_placeholders(destr)
	//                 a1 = [lval] + phs[1:]
	//                 a2 = [mut] + phs[1:]
	//                 fmlas.append(eq_atom(destr(*a1), destr(*a2)))
	mutSort := mut.NodeSort()
	if mutSort != nil && m.SortDestructors != nil {
		sortName := IvySortName(mutSort)
		if destrs, ok := m.SortDestructors[sortName]; ok {
			for _, destr := range destrs {
				if destr.Name == nSym.Name {
					continue
				}
				phs := SymPlaceholders(destr)
				a1 := make([]Expr, 0, 1+len(phs))
				a1 = append(a1, lval)
				if len(phs) > 1 {
					for _, v := range phs[1:] {
						a1 = append(a1, v)
					}
				}
				a2 := make([]Expr, 0, 1+len(phs))
				a2 = append(a2, mut)
				if len(phs) > 1 {
					for _, v := range phs[1:] {
						a2 = append(a2, v)
					}
				}
				d1 := applyToNodes(destr, a1)
				d2 := applyToNodes(destr, a2)
				*fmlas = append(*fmlas, &Eq{T1: d1, T2: d2})
			}
		}
	}

	// Python: return lhs.rep(*([lval]+rest)), new_clauses, mutated
	retArgs := make([]Expr, 0, 1+len(rest))
	retArgs = append(retArgs, lval)
	retArgs = append(retArgs, rest...)
	retLhs := applyToNodes(nSym, retArgs)

	return retLhs, newClauses, mutated
}

// --- AssignRefs ---

// AssignRefs collects all referenced symbols in an assignment LHS,
// including through destructor chains.
// Corresponds to Python's assign_refs.
func AssignRefs(lhsNode Expr, refs map[string]bool, m *Module) {
	assignRefsRec(lhsNode, refs, m)
}

func assignRefsRec(node Expr, refs map[string]bool, m *Module) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *Apply:
		if c, ok := n.Func.(*Const); ok {
			if m.DestructorSorts != nil {
				if _, isDestr := m.DestructorSorts[c.Name]; isDestr {
					refs[c.Name] = true
					if len(n.Terms) > 0 {
						assignRefsRec(n.Terms[0], refs, m)
					}
					for _, a := range n.Terms[1:] {
						for _, sym := range SymbolsAST(a) {
							refs[ExprName(sym)] = true
						}
					}
					return
				}
			}
		}
		// Non-destructor: collect all symbol refs
		for _, a := range n.Terms {
			for _, sym := range SymbolsAST(a) {
				refs[ExprName(sym)] = true
			}
		}
	case *Const:
		refs[n.Name] = true
	}
}

// --- Sign ---

// Sign applies polarity to an atom. If polarity is true, returns the atom;
// if false, returns Not(atom).
// Corresponds to Python's sign.
func Sign(polarity bool, atom Expr) Expr {
	if polarity {
		return atom
	}
	return &Not{Body: atom}
}

// --- MakeFieldUpdate ---

// MakeFieldUpdate generates update formulas for field/destructor assignment.
// The field f must be a binary relation. r is applied to variable v to produce the RHS.
// Returns the transition relation update.
// Corresponds to Python's make_field_update.
func MakeFieldUpdate(self ActionsAction, l Expr, f *Const, r Expr, domain *Module, pvars map[string]bool) *Update {
	// Python: if not f.is_relation() or len(f.sort.dom) != 2:
	//             raise IvyError(self, "field " + str(f) + " must be a binary relation")
	fs, ok := f.CSort.(*FunctionSort)
	if !ok || len(fs.Sorts) != 3 { // dom[0], dom[1], range
		panic(fmt.Sprintf("field %s must be a binary relation", f.Name))
	}

	// Python: v = Variable('X', f.sort.dom[1])
	v, _ := NewVariable("X", fs.Sorts[1])

	// Python: aa = AssignAction(f(l,v), r(v))
	fApp, _ := NewApply(f, l, v)
	var rVal Expr
	if IsFunctionSort(r.NodeSort()) {
		rVal, _ = NewApply(r, v)
	} else {
		rVal = r
	}
	aa := NewAssignAction(fApp, rVal)

	// Python: return aa.action_update(domain, pvars)
	ctx := &UpdateContext{
		Domain:          domain,
		PVars:           pvars,
		ActCfg:          domain.Cfg.ActCfg,
		Instantiator:    domain.Instantiator,
		CheckUnprovable: domain.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   domain.Cfg.CheckLineno,
	}
	return aa.ActionUpdate(ctx)
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

// --- Determinize ---

// SetDeterminize sets the determinize flag on the given ActionsConfig.
func SetDeterminize(cfg *ActionsConfig, t bool) {
	cfg.Determinize = t
}

// GetDeterminize returns the current determinize setting from the given ActionsConfig.
func GetDeterminize(cfg *ActionsConfig) bool {
	return cfg.Determinize
}

// --- BracketAction ---

// BracketAction formats an action with braces if it's not already a Sequence.
// Corresponds to Python's bracket_action.
func BracketAction(action ActionsAction, depth int) string {
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
	DebugExpr Expr   // debug expression (first arg)
	WithExprs []Expr // additional "with" expressions
}

// NewDebugAction creates a new DebugAction.
func NewDebugAction(debugExpr Expr, withExprs ...Expr) *DebugAction {
	return &DebugAction{DebugExpr: debugExpr, WithExprs: copyNodes(withExprs)}
}

func (a *DebugAction) Name() string { return "debug" }
func (a *DebugAction) ActionArgs() []Expr {
	args := []Expr{a.DebugExpr}
	args = append(args, a.WithExprs...)
	return args
}
func (a *DebugAction) ActionClone(args []Expr) ActionsAction {
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
func (a *DebugAction) IterCalls() []string             { return nil }
func (a *DebugAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }
func (a *DebugAction) Decompose() [][]ActionsAction    { return atomicDecompose(a) }

// --- Entry ---

// Entry creates an RME (Rely-Guarantee relation) entry for action semantics.
// Corresponds to Python's entry function.
func Entry(ensures ...Expr) *RME {
	var ensNode Expr
	if len(ensures) > 0 {
		ensNode = ensures[0]
	} else {
		ensNode = &And{} // And() with no args = true
	}
	return NewRME(&And{}, nil, ensNode)
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
func NewTypeCheckContext(domain *Module) *TypeCheckContext {
	return &TypeCheckContext{
		ActionContext: ActionContext{Domain: domain},
	}
}

// Get resolves an action name, returning a null action with preserved formals.
// Corresponds to Python's TypeCheckConext.get.
func (tc *TypeCheckContext) Get(x string) ActionsAction {
	if tc.Domain == nil {
		return nil
	}
	actI, ok := tc.Domain.Actions.Get2(x)
	if !ok {
		return nil
	}
	act, ok := actI.(ActionsAction)
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
func TypeCheckActionFull(action ActionsAction, domain *Module, pvars map[string]bool) {
	// In Python, this function is a no-op (early return).
	// It was intended to use TypeCheckContext to run int_update
	// but is currently disabled in the Python source.
	return
}
