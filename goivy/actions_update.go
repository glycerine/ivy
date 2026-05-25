// This file implements the action_update() and int_update() methods for
// each action type, porting the transition-relation computation from
// Python ivy_actions.py lines 309-1302.
//
// Each action's ActionUpdate method returns a *Update representing
// the transition relation (modified symbols, TR formula, precondition).
//
// IntUpdate applies update axioms from the domain on top of ActionUpdate.
// GetUpdate is the top-level entry point that hides formals and binds olds.
package goivy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// -----------------------------------------------------------------------
// UpdateContext provides the domain and in-scope variables for computing
// action updates. It wraps the module and an action-resolution function.
// -----------------------------------------------------------------------

// UpdateContext holds the context needed for computing action updates.
type UpdateContext struct {
	Domain *Module
	PVars  map[string]bool // in-scope variable names

	// ActCfg is the per-session actions config. Used for context lookups
	// (replaces the old GlobalContext global).
	ActCfg *ActionsConfig

	// GetAction resolves an action name to its Action. This is set by
	// the caller (typically from module.Actions or an ActionContext).
	GetAction func(name string) ActionsAction

	// CheckUnprovable corresponds to Python's check_unprovable parameter.
	// When true, only unprovable assertions are checked.
	CheckUnprovable bool

	// CheckedAssert corresponds to Python's checked_assert parameter.
	// If non-empty, only the assertion whose lineno matches is fully checked.
	CheckedAssert string

	// Instantiator provides definition instances for clausification.
	// Corresponds to Python's global `instantiator` variable.
	// When non-nil, used by AssumeAction and AssertAction to unfold definitions.
	Instantiator func([]Expr) *Clauses
}

// BackgroundTheory returns the background theory (axioms) for the domain.
func (ctx *UpdateContext) BackgroundTheory() *Clauses {
	if ctx.Domain == nil {
		return TrueClauses(nil)
	}
	clauses := ctx.Domain.BackgroundTheory(ctx.PVars)
	if clauses == nil {
		return TrueClauses(nil)
	}
	return clauses
}

// makeUpdate creates a transrel.Update from individual components,
// wrapping lg.Expr values into Clauses. This is a transitional helper
// for porting action updates from bare Node to Clauses.
func makeUpdate(modified []*Const, tr Expr, pre Expr, annot Annotation) *Update {
	return &Update{
		Modified: modified,
		TR:       FormulaToClauses(tr, annot),
		Pre:      FormulaToClauses(pre, annot),
	}
}

// makeUpdateDefs creates a transrel.Update with definitions in the TR.
// This matches Python's pattern of Clauses([], [Definition(...)], annot).
func makeUpdateDefs(modified []*Const, defs []*IvyDefinition, annot Annotation) *Update {
	return &Update{
		Modified: modified,
		TR:       NewClauses(nil, defs, annot),
		Pre:      FalseClauses(annot),
	}
}

// -----------------------------------------------------------------------
// Formula helpers
// -----------------------------------------------------------------------

// equivAST returns the equivalence of two nodes:
//   - If both are individual (non-Boolean): returns Eq(a, b)
//   - If both are Boolean: returns And(Or(a, Not(b)), Or(Not(a), b))
func equivAST(a, b Expr) Expr {
	if IsIndividual(a) {
		return &Eq{T1: a, T2: b}
	}
	// Boolean equivalence: (a | ~b) & (~a | b)
	notA := Negate(a)
	notB := Negate(b)
	or1, _ := NewOr(a, notB)
	or2, _ := NewOr(notA, b)
	and, _ := NewAnd(or1, or2)
	return and
}

// conjoin creates a conjunction, simplifying True cases.
func conjoin(nodes ...Expr) Expr {
	var terms []Expr
	for _, n := range nodes {
		if actionsUpdateIsTrue(n) {
			continue
		}
		if actionsUpdateIsFalse(n) {
			return False
		}
		terms = append(terms, n)
	}
	if len(terms) == 0 {
		return True
	}
	if len(terms) == 1 {
		return terms[0]
	}
	and, _ := NewAnd(terms...)
	return and
}

// disjoin creates a disjunction, simplifying False cases.
func disjoin(nodes ...Expr) Expr {
	var terms []Expr
	for _, n := range nodes {
		if actionsUpdateIsFalse(n) {
			continue
		}
		if actionsUpdateIsTrue(n) {
			return True
		}
		terms = append(terms, n)
	}
	if len(terms) == 0 {
		return False
	}
	if len(terms) == 1 {
		return terms[0]
	}
	or, _ := NewOr(terms...)
	return or
}

func actionsUpdateIsTrue(n Expr) bool {
	return IsTrue(n)
}

func actionsUpdateIsFalse(n Expr) bool {
	return IsFalse(n)
}

// -----------------------------------------------------------------------
// newSym returns the post-state version of a symbol.
// -----------------------------------------------------------------------

func newSym(sym *Const) *Const {
	return NewConst(ActionNewName(sym.Name), sym.CSort)
}

// constName extracts the name from a node that is a Const or the Func of an Apply.
func constName(n Expr) string {
	switch t := n.(type) {
	case *Const:
		return t.Name
	case *Apply:
		return constName(t.Func)
	}
	return ""
}

// constSym extracts the Const from a node (Const or Apply.Func).
func constSym(n Expr) *Const {
	switch t := n.(type) {
	case *Const:
		return t
	case *Apply:
		if c, ok := t.Func.(*Const); ok {
			return c
		}
	}
	return nil
}

// nodeArgs extracts the args from a node (Apply.Terms or nil).
func nodeArgs(n Expr) []Expr {
	if app, ok := n.(*Apply); ok {
		return app.Terms
	}
	return nil
}

// addParametersAST appends params to every application in the AST.
// For a Symbol, wraps it in Apply(sym, params...).
// For an Apply, appends params to its Terms.
// Corresponds to Python's add_parameters_ast (ivy_ast.py:1756).
func addParametersAST(node Expr, params []Expr) Expr {
	// Python's add_parameters_ast has no early return for empty params.
	switch t := node.(type) {
	case *Const:
		app, _ := NewApply(t, params...)
		return app
	case *Apply:
		newTerms := make([]Expr, len(t.Terms)+len(params))
		copy(newTerms, t.Terms)
		copy(newTerms[len(t.Terms):], params)
		app, _ := NewApply(t.Func, newTerms...)
		return app
	default:
		return node
	}
}

// -----------------------------------------------------------------------
// mkAssignClauses creates the transition relation for an assignment.
//
// For lhs := rhs, this creates:
//   new_n(V0, V1, ...) = ITE(V0=a0 & V1=a1, rhs[vars→Vs], n(V0, V1, ...))
//
// where Vi are placeholder variables for the symbol's domain sorts,
// ai are the specific index arguments from the LHS, and the ITE says:
// at the assigned indices use the RHS value, elsewhere keep the old value.
// -----------------------------------------------------------------------

func mkAssignClauses(lhs, rhs Expr) *Update {
	sym := constSym(lhs)
	if sym == nil {
		// Fallback: no-op update
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// Placeholder variables for the full domain of the symbol
	phs := SymPlaceholders(sym)

	// Build new_n applied to placeholders
	phNodes := actionsVarsToNodes(phs)
	dlhs := applyToNodes(newN, phNodes)

	// Build equality conditions for non-variable args
	var eqs []Expr
	rn := make(map[string]Expr)
	for i, ph := range phs {
		if i < len(args) {
			if _, isVar := args[i].(*LogicVariable); isVar {
				// Variable arg: record substitution (arg.Name → placeholder)
				rn[args[i].(*LogicVariable).Name] = ph
			} else {
				// Non-variable: add equality constraint
				eqs = append(eqs, &Eq{T1: ph, T2: args[i]})
			}
		}
	}

	// Substitute variable args in RHS
	drhs := SubstituteAstByName(rhs, rn)

	// If there are equality conditions, build ITE
	// Python: ActionIte(And(*eqs), drhs, n(*dlhs.args))
	if len(eqs) > 0 {
		eqConj, _ := NewAnd(eqs...) // match Python And(*eqs) exactly
		// old value: n applied to placeholders
		oldVal := applyToNodes(sym, phNodes)
		rhsSort := rhs.NodeSort()
		if rhsSort == nil {
			rhsSort = TopS
		}
		drhs = &LogicIte{ISort: rhsSort, Cond: eqConj, Then: drhs, Else: oldVal}
	}

	// Python: Clauses([], [Definition(dlhs, drhs)], EmptyAnnotation())
	// Store as a Definition in Clauses.Defs, matching Python exactly.
	defn := NewIvyDefinition(dlhs, drhs)
	return &Update{
		Modified: []*Const{sym},
		TR:       NewClauses(nil, []*IvyDefinition{defn}, EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
}

func actionsVarsToNodes(vars []*LogicVariable) []Expr {
	nodes := make([]Expr, len(vars))
	for i, v := range vars {
		nodes[i] = v
	}
	return nodes
}

// -----------------------------------------------------------------------
// ActionUpdate implementations for each action type
// -----------------------------------------------------------------------

// --- AssumeAction ---

// ActionUpdate computes the transition relation for AssumeAction.
// An assume adds the formula as a constraint on the current state.
// If the formula came from a LabeledFormula with unprovable=true, skip entirely.
// Python: action_update returns ([], clauses, false_clauses())
func (a *LogicAssumeAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssumeAction.action_update ENTER")
	defer xtracer.Trace("actions.AssumeAction.action_update EXIT")
	// Python: if isinstance(fmla, LabeledFormula) and fmla.unprovable: return skip
	// Python returns true_clauses/false_clauses directly, not formula_to_clauses.
	if a.Unprovable {
		return &Update{
			Modified: []*Const{},
			TR:       TrueClauses(EmptyAnnotation{}),
			Pre:      FalseClauses(EmptyAnnotation{}),
		}
	}
	fmla := a.Formula
	// Python: clauses = formula_to_clauses_tseitin(skolemize_formula(fmla))
	//         clauses = unfold_definitions_clauses(clauses)
	//         clauses = Clauses(clauses.fmlas, clauses.defs, EmptyAnnotation())
	var skInst func([]Expr) *Clauses
	if ctx != nil {
		skInst = ctx.Instantiator
	}
	fmla = SkolemizeFormula(fmla, nil, skInst)
	clauses := FormulaToClauses(fmla, nil)
	if ctx != nil && ctx.Instantiator != nil {
		clauses = UnfoldDefinitionsClauses(clauses, ctx.Instantiator)
	}
	clauses = NewClauses(clauses.Fmlas, clauses.Defs, EmptyAnnotation{})
	return &Update{
		Modified: []*Const{},
		TR:       clauses,
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
}

// --- AssertAction ---

// ActionUpdate computes the transition relation for AssertAction.
// An assert generates a precondition (negative) from the dual of the formula.
// Implements Python's selective assertion checking via check_unprovable and checked_assert.
// Python: action_update (ivy_actions.py:343-362)
func (a *LogicAssertAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssertAction.action_update ENTER")
	defer xtracer.Trace("actions.AssertAction.action_update EXIT")
	fmla := a.Formula
	unprovable := a.Unprovable

	// Python: if check_unprovable.get() != unprovable: skip
	// Python returns ([], true_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation()))
	// — pre-built Clauses, NOT formula_to_clauses(). Do NOT use makeUpdate here.
	if ctx.CheckUnprovable != unprovable {
		return &Update{
			Modified: []*Const{},
			TR:       TrueClauses(EmptyAnnotation{}),
			Pre:      FalseClauses(EmptyAnnotation{}),
		}
	}

	// Python: if ca != self.lineno — compares full Location(file, line).
	// CheckedAssert format: "file:line" via Location.FileLineKey().
	if ctx.CheckedAssert != "" {
		loc := a.GetLineno()
		if ctx.CheckedAssert != loc.FileLineKey() {
			if unprovable {
				return &Update{
					Modified: []*Const{},
					TR:       TrueClauses(EmptyAnnotation{}),
					Pre:      FalseClauses(EmptyAnnotation{}),
				}
			}
			// Python: return ([],formula_to_clauses(fmla,annot=EmptyAnnotation()),false_clauses(annot=EmptyAnnotation()))
			return &Update{
				Modified: []*Const{},
				TR:       FormulaToClauses(fmla, EmptyAnnotation{}),
				Pre:      FalseClauses(EmptyAnnotation{}),
			}
		}
	}

	// Only assertions that pass both filters get dual formula treatment
	// Python: cl = formula_to_clauses(dual_formula(fmla))
	//         cl = Clauses(cl.fmlas, cl.defs, EmptyAnnotation())
	// Python's dual_formula consults lu.instantiator (a module-level global);
	// Go's literal port reads it from ctx.Instantiator (set from Domain.Instantiator).
	var dualInst func([]Expr) *Clauses
	if ctx != nil {
		dualInst = ctx.Instantiator
	}
	dual := DualFormula(fmla, nil, dualInst)
	cl := FormulaToClauses(dual, nil)
	cl = NewClauses(cl.Fmlas, cl.Defs, EmptyAnnotation{})
	return &Update{
		Modified: []*Const{},
		TR:       TrueClauses(EmptyAnnotation{}),
		Pre:      cl,
	}
}

// --- RequiresAction ---

func (a *LogicRequiresAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.LogicAssertAction.ActionUpdate(ctx)
}

// --- EnsuresAction ---

func (a *LogicEnsuresAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.LogicAssertAction.ActionUpdate(ctx)
}

// --- SubgoalAction ---

func (a *LogicSubgoalAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.LogicAssertAction.ActionUpdate(ctx)
}

// --- AssignAction ---

// ActionUpdate computes the transition relation for assignment lhs := rhs.
// Handles:
// 1. Hierarchical symbols (decompose into sub-assignments)
// 2. Partial applications (add placeholder parameters)
// 3. Destructor assignments (mutate through destructors)
// 4. Variant assignments
// 5. Simple assignments
func (a *LogicAssignAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssignAction.action_update ENTER")
	defer xtracer.Trace("actions.AssignAction.action_update EXIT")
	lhs, rhs := a.LHS, a.RHS
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}

	// Python compares lhs.rep (a Const) directly against domain.hierarchy.
	// Normal hierarchy entries are keyed by strings from add_to_hierarchy, so
	// compiled AssignAction values do not enter Python's stale hierarchy branch.

	// Partial application extension
	// Python: xtra = len(lhs.rep.sort.dom) - len(lhs.args)
	dom := SortDomain(sym.CSort)
	xtra := len(dom) - len(nodeArgs(lhs))
	if xtra < 0 {
		panic(fmt.Sprintf("too many parameters in assignment to %s", sym.Name))
	}
	if xtra > 0 {
		// Extend lhs and rhs with fresh placeholder variables
		phs := SymPlaceholders(sym)
		if len(phs) < xtra {
			return NullUpdate()
		}
		extend := make([]Expr, xtra)
		for i := 0; i < xtra; i++ {
			extend[i] = phs[len(phs)-xtra+i]
		}
		// Make variables distinct from those already used in lhs and rhs
		// Python: extend = variables_distinct_list_ast(extend, self)
		// We combine lhs and rhs into a single expression for variable collection
		combined := &LogicAnd{Terms: []Expr{lhs, rhs}}
		extend = VariablesDistinctListAst(extend, combined)

		lhs = addParametersAST(lhs, extend)
		// Assignment of individual to a boolean is a special case
		// Guard against nil sorts
		if rhs.NodeSort() != nil && lhs.NodeSort() != nil && IsIndividual(rhs) && !IsIndividual(lhs) {
			lastExt := extend[len(extend)-1]
			rhsExtended := addParametersAST(rhs, extend[:len(extend)-1])
			rhs = &Eq{T1: lastExt, T2: rhsExtended}
		} else {
			rhs = addParametersAST(rhs, extend)
		}
	}

	// Variable check: all RHS variables must appear in LHS
	// Python: if any(v not in lhs_vars for v in used_variables_ast(rhs)): raise IvyError
	lhsVars := UsedVariablesAST(lhs)
	rhsVars := UsedVariablesAST(rhs)
	for k := range rhsVars {
		if _, found := lhsVars[k]; !found {
			panic(fmt.Sprintf("multiply assigned: %s", sym.Name))
		}
	}

	if err := TypeCheck(ctx.Domain, rhs); err != nil {
		panic(err.Error())
	}
	if IsIndividual(lhs) != IsIndividual(rhs) {
		panic(fmt.Sprintf("sort mismatch in assignment to %s", sym.Name))
	}

	// Handle destructor assignments
	if ctx.Domain != nil && ctx.Domain.DestructorSorts != nil {
		if _, ok := ctx.Domain.DestructorSorts[sym.Name]; ok {
			xtracer.Trace("actions.AssignAction.action_update branch=destructor")
			return a.destructorAssignUpdate(ctx, lhs, rhs)
		}
	}

	// Handle variant assignments
	if ctx.Domain != nil && ctx.Domain.Variants != nil {
		lhsSort := lhs.NodeSort()
		rhsSort := rhs.NodeSort()
		if lhsSort != nil && rhsSort != nil && isVariant(ctx.Domain, lhsSort, rhsSort) {
			xtracer.Trace("actions.AssignAction.action_update branch=variant")
			return mkVariantAssignClauses(lhs, rhs, ctx.Domain)
		}
	}

	// Standard assignment
	xtracer.Trace("actions.AssignAction.action_update branch=standard")
	return mkAssignClauses(lhs, rhs)
}

// destrAsgnVal recursively builds the transition relation for destructor assignments.
// Python: destr_asgn_val (ivy_actions.py:428-454).
// Returns (nondet_lhs, new_clauses, mutated_symbol).
func destrAsgnVal(lhs Expr, fmlas *[]Expr, domain *Module) (Expr, *Clauses, *Const) {
	lhsArgs := nodeArgs(lhs)
	if len(lhsArgs) == 0 {
		return lhs, FalseClauses(nil), nil
	}

	mut := lhsArgs[0]
	rest := lhsArgs[1:]
	mutN := constSym(mut)
	if mutN == nil {
		return lhs, FalseClauses(nil), nil
	}

	var lval Expr
	var newClauses *Clauses
	var mutated *Const

	if _, ok := domain.DestructorSorts[mutN.Name]; ok {
		// Recursive case: nested destructor
		// Python: lval, new_clauses, mutated = destr_asgn_val(mut, fmlas)
		lval, newClauses, mutated = destrAsgnVal(mut, fmlas, domain)
	} else {
		// Base case: mut is the root mutable symbol
		// Python: nondet = mut_n.suffix("_nd").skolem()
		skSym := NewConst("__"+mutN.Name+"_nd", mutN.CSort)
		phs := SymPlaceholders(mutN)
		phNodes := actionsVarsToNodes(phs)
		// Python: new_clauses = mk_assign_clauses(mut_n, nondet(*sym_placeholders(mut_n)))
		var skApplied Expr
		if len(phNodes) > 0 {
			skApplied = applyToNodes(skSym, phNodes)
		} else {
			skApplied = skSym
		}
		newClauses = mkAssignClauses(mutN, skApplied).TR
		// Python: lval = nondet(*mut.args)
		mutArgs := nodeArgs(mut)
		if len(mutArgs) > 0 {
			mutArgNodes := make([]Expr, len(mutArgs))
			copy(mutArgNodes, mutArgs)
			lval = applyToNodes(skSym, mutArgNodes)
		} else {
			lval = skSym
		}
		mutated = mutN
	}

	// n = lhs.rep
	n := constSym(lhs)
	if n == nil {
		return lhs, newClauses, mutated
	}

	// vs = sym_placeholders(n)
	vs := SymPlaceholders(n)
	vsNodes := actionsVarsToNodes(vs)

	// dlhs = n(*([lval] + vs[1:]))
	dlhsArgs := make([]Expr, len(vsNodes))
	dlhsArgs[0] = lval
	for i := 1; i < len(vsNodes); i++ {
		dlhsArgs[i] = vsNodes[i]
	}
	dlhs := applyToNodes(n, dlhsArgs)

	// drhs = n(*([mut] + vs[1:]))
	drhsArgs := make([]Expr, len(vsNodes))
	drhsArgs[0] = mut
	for i := 1; i < len(vsNodes); i++ {
		drhsArgs[i] = vsNodes[i]
	}
	drhs := applyToNodes(n, drhsArgs)

	// eqs = [eq_atom(v,a) for (v,a) in list(zip(vs,lhs.args))[1:] if not isinstance(a,Variable)]
	var eqs []Expr
	for i := 1; i < len(vs) && i < len(lhsArgs); i++ {
		if _, isVar := lhsArgs[i].(*LogicVariable); !isVar {
			eqs = append(eqs, &Eq{T1: vs[i], T2: lhsArgs[i]})
		}
	}

	// if eqs: fmlas.append(Or(And(*eqs), equiv_ast(dlhs, drhs)))
	if len(eqs) > 0 {
		var guard Expr
		if len(eqs) == 1 {
			guard = eqs[0]
		} else {
			guard = &LogicAnd{Terms: eqs}
		}
		equiv := equivAST(dlhs, drhs)
		*fmlas = append(*fmlas, disjoin(guard, equiv))
	}

	// Frame conditions: for each sibling destructor, preserve its value
	// Python: for destr in ivy_module.module.sort_destructors[mut.sort.name]:
	mutSort := mut.NodeSort()
	if mutSort != nil {
		sortName := ""
		if ns, ok := mutSort.(interface{ GetName() string }); ok {
			sortName = ns.GetName()
		} else {
			sortName = mutSort.String()
		}
		if destrs, ok := domain.SortDestructors.Get2(sortName); ok {
			for _, destr := range destrs {
				if destr.Name != n.Name {
					destrPhs := SymPlaceholders(destr)
					destrPhNodes := actionsVarsToNodes(destrPhs)
					// a1 = [lval] + phs[1:]
					a1 := make([]Expr, len(destrPhNodes))
					a1[0] = lval
					for j := 1; j < len(destrPhNodes); j++ {
						a1[j] = destrPhNodes[j]
					}
					// a2 = [mut] + phs[1:]
					a2 := make([]Expr, len(destrPhNodes))
					a2[0] = mut
					for j := 1; j < len(destrPhNodes); j++ {
						a2[j] = destrPhNodes[j]
					}
					destrA1 := applyToNodes(destr, a1)
					destrA2 := applyToNodes(destr, a2)
					*fmlas = append(*fmlas, &Eq{T1: destrA1, T2: destrA2})
				}
			}
		}
	}

	// return lhs.rep(*([lval]+rest)), new_clauses, mutated
	resultArgs := make([]Expr, 1+len(rest))
	resultArgs[0] = lval
	copy(resultArgs[1:], rest)
	resultExpr := applyToNodes(n, resultArgs)

	return resultExpr, newClauses, mutated
}

// destructorAssignUpdate handles assignment through destructors.
// Python: destructor case in AssignAction.action_update (ivy_actions.py:533-538).
func (a *LogicAssignAction) destructorAssignUpdate(ctx *UpdateContext, lhs, rhs Expr) *Update {
	var fmlas []Expr
	nondetLhs, newClauses, mutN := destrAsgnVal(lhs, &fmlas, ctx.Domain)
	if mutN == nil {
		return NullUpdate()
	}

	// Python: fmlas.append(equiv_ast(nondet_lhs, rhs))
	fmlas = append(fmlas, equivAST(nondetLhs, rhs))

	// Python: new_clauses = and_clauses(new_clauses, Clauses(fmlas))
	fmlaClauses := NewClauses(fmlas, nil, nil)
	combined := AndClausesTyped(newClauses, fmlaClauses)

	// Python: return ([mut_n], new_clauses, false_clauses(annot=EmptyAnnotation()))
	return &Update{
		Modified: []*Const{mutN},
		TR:       combined,
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
}

// isVariant checks if lhsSort has rhsSort as a variant.
func isVariant(domain *Module, lhsSort, rhsSort Sort) bool {
	if domain.Variants == nil {
		return false
	}
	lhsName := IvySortName(lhsSort)
	variants, ok := domain.Variants[lhsName]
	if !ok {
		return false
	}
	for _, v := range variants {
		if SortEqual(v, rhsSort) {
			return true
		}
	}
	return false
}

// mkVariantAssignClauses creates the transition relation for a variant assignment.
// Python: mk_variant_assign_clauses (ivy_actions.py:593-611).
// Asserts that the new value points-to the RHS via pto, and does NOT point-to
// any other variant sort.
func mkVariantAssignClauses(lhs, rhs Expr, domain *Module) *Update {
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// dlhs = new_n(*sym_placeholders(n))
	phs := SymPlaceholders(sym)
	phNodes := actionsVarsToNodes(phs)
	dlhs := applyToNodes(newN, phNodes)
	vs := phs // dlhs.args are the placeholders

	// Build eqs for non-variable args
	// Python: eqs = [eq_atom(v,a) for (v,a) in zip(vs,args) if not isinstance(a,Variable)]
	var eqs []Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*LogicVariable); !isVar {
				eqs = append(eqs, &Eq{T1: v, T2: args[i]})
			}
		}
	}

	// Build rename map for variable args, compute drhs = substitute_ast(rhs, rn)
	// Python: rn = dict((a.rep,v) for v,a in zip(vs,args) if isinstance(a,Variable))
	rn := make(map[string]Expr)
	for i, v := range vs {
		if i < len(args) {
			if varArg, isVar := args[i].(*LogicVariable); isVar {
				rn[varArg.Name] = v
			}
		}
	}
	drhs := rhs
	if len(rn) > 0 {
		drhs = SubstituteAstByName(rhs, rn)
	}

	// Create nondeterministic skolem symbol
	// Python: nondet = n.suffix("_nd").skolem()
	skSym := NewConst("__"+sym.Name+"_nd", sym.CSort)
	var nondet Expr
	if len(phNodes) > 0 {
		nondet = applyToNodes(skSym, phNodes)
	} else {
		nondet = skSym
	}

	// If eqs: nondet = ActionIte(And(*eqs), nondet, n(*dlhs.args))
	if len(eqs) > 0 {
		var guard Expr
		if len(eqs) == 1 {
			guard = eqs[0]
		} else {
			guard = &LogicAnd{Terms: eqs}
		}
		origApply := applyToNodes(sym, phNodes) // n(*dlhs.args)
		ite, err := NewIte(guard, nondet, origApply)
		if err == nil {
			nondet = ite
		}
	}

	// Build formulas
	lhsSort := lhs.NodeSort()
	rhsSort := rhs.NodeSort()

	var fmlas []Expr

	// Iff(pto(lsort,rsort)(dlhs, Variable('X',rsort)), Equals(Variable('X',rsort), drhs))
	xVar, _ := NewVariable("X", rhsSort)
	ptoSym := Pto(lhsSort, rhsSort)
	ptoApp := applyToNodes(ptoSym, append([]Expr{dlhs}, xVar))
	eqXdrhs := &Eq{T1: xVar, T2: drhs}
	iff, err := NewIff(ptoApp, eqXdrhs)
	if err == nil {
		fmlas = append(fmlas, iff)
	}

	// For each variant sort s != rsort: Not(pto(lsort,s)(dlhs, Variable('X',s)))
	if lhsSort != nil {
		lhsSortName := ""
		if ns, ok := lhsSort.(interface{ GetName() string }); ok {
			lhsSortName = ns.GetName()
		} else {
			lhsSortName = lhsSort.String()
		}
		for _, s := range domain.Variants[lhsSortName] {
			if !SortEqual(s, rhsSort) {
				xv, _ := NewVariable("X", s)
				ptoS := Pto(lhsSort, s)
				ptoSApp := applyToNodes(ptoS, append([]Expr{dlhs}, xv))
				fmlas = append(fmlas, &LogicNot{Body: ptoSApp})
			}
		}
	}

	// Return as update with definition: Definition(dlhs, nondet)
	defn := NewIvyDefinition(dlhs, nondet)
	defs := []*LogicDefinition{defn}

	// Combine: formulas go in TR, definition goes in defs
	// Python: new_clauses = Clauses(fmlas, [Definition(dlhs, nondet)])
	update := &Update{
		Modified: []*Const{sym},
		TR:       NewClauses(fmlas, defs, EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
	return update
}

// --- HavocAction ---

// ActionUpdate computes the transition relation for havoc (nondeterministic assignment).
// The new value is unconstrained at the specified indices, but equal to the old
// value at all other indices.
func (a *LogicHavocAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.HavocAction.action_update ENTER")
	defer xtracer.Trace("actions.HavocAction.action_update EXIT")
	lhs := a.Target
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)
	dom := SortDomain(sym.CSort)

	// Create fresh variables for the domain
	vs := make([]*LogicVariable, len(dom))
	for i, s := range dom {
		v, _ := NewVariable(fmt.Sprintf("X%d", i), s)
		vs[i] = v
	}

	// Build equality conditions for non-variable args
	var eqs []Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*LogicVariable); !isVar {
				eqs = append(eqs, &Eq{T1: v, T2: args[i]})
			}
		}
	}

	var tr Expr
	vsNodes := actionsVarsToNodes(vs)

	if IsBoolean(sym) || IsRelationalSort(sym.CSort) {
		// Relation: at non-havocked indices, old and new agree
		// For each eq in eqs: (new_n(Vs) <-> n(Vs)) | eq
		var terms []Expr
		newApp := applyToNodes(newN, vsNodes)
		oldApp := applyToNodes(sym, vsNodes)
		for _, eq := range eqs {
			impl1, _ := NewOr(&LogicNot{Body: newApp}, oldApp, eq)
			impl2, _ := NewOr(newApp, &LogicNot{Body: oldApp}, eq)
			terms = append(terms, impl1, impl2)
		}
		if len(terms) == 0 {
			tr = True // no constraints at all: fully nondeterministic
		} else {
			tr = conjoin(terms...)
		}
	} else if IsIndividual(sym) {
		// Function: at non-havocked indices, new = old
		newApp := applyToNodes(newN, vsNodes)
		oldApp := applyToNodes(sym, vsNodes)
		var terms []Expr
		for _, eq := range eqs {
			clause, _ := NewOr(&Eq{T1: newApp, T2: oldApp}, eq)
			terms = append(terms, clause)
		}
		if len(terms) == 0 {
			tr = True
		} else {
			tr = conjoin(terms...)
		}
	} else {
		tr = True
	}

	return makeUpdate([]*Const{sym}, tr, False, EmptyAnnotation{})
}

func applyToNodes(fn Expr, args []Expr) Expr {
	if len(args) == 0 {
		return fn
	}
	return MustApply(fn, args...)
}

// --- SetAction ---

// ActionUpdate computes the transition relation for a set operation on a relation.
// Python: SetAction.action_update (ivy_actions.py:624-636).
// Builds clauses with frame conditions ensuring values at non-matching indices
// are preserved.
func (a *LogicSetAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.SetAction.action_update ENTER")
	defer xtracer.Trace("actions.SetAction.action_update EXIT")
	if a.Lit == nil {
		return NullUpdate()
	}

	// Determine polarity and atom
	lit := a.Lit
	positive := true
	if n, ok := lit.(*LogicNot); ok {
		positive = false
		lit = n.Body
	}

	// Extract the relation symbol and args from the atom
	var relSym *Const
	var args []Expr
	if app, ok := lit.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			relSym = c
			args = app.Terms
		}
	} else if c, ok := lit.(*Const); ok {
		relSym = c
	}

	if relSym == nil {
		return NullUpdate()
	}

	newN := newSym(relSym)
	vs := SymPlaceholders(relSym)
	vsNodes := actionsVarsToNodes(vs)

	// Build equality conditions for non-variable args
	// Python: eqs = [Atom(equals,[v,a]) for (v,a) in zip(vs,args) if not isinstance(a,Variable)]
	var eqs []Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*LogicVariable); !isVar {
				eqs = append(eqs, &Eq{T1: v, T2: args[i]})
			}
		}
	}

	// Build the formula components
	// Python: new_clauses = LogicAnd(*(
	//   [Or(sign(lit.polarity, Atom(new_n, vs)), sign(1-lit.polarity, Atom(n, vs))),
	//    sign(lit.polarity, Atom(new_n, args))] +
	//   [Or(*([sign(0, Atom(new_n, vs)), sign(1, Atom(n, vs))] + [eq])) for eq in eqs] +
	//   [Or(*([sign(1, Atom(new_n, vs)), sign(0, Atom(n, vs))] + [eq])) for eq in eqs]))

	newNvs := applyToNodes(newN, vsNodes) // new_n(Vs)
	nVs := applyToNodes(relSym, vsNodes)  // n(Vs)
	newNargs := applyToNodes(newN, args)  // new_n(args)

	var parts []Expr

	// Or(sign(polarity, new_n(Vs)), sign(!polarity, n(Vs)))
	parts = append(parts, disjoin(Sign(positive, newNvs), Sign(!positive, nVs)))
	// sign(polarity, new_n(args))
	parts = append(parts, Sign(positive, newNargs))

	// Frame conditions for non-set indices:
	for _, eq := range eqs {
		// Or(sign(false, new_n(Vs)), sign(true, n(Vs)), eq)
		parts = append(parts, disjoin(Sign(false, newNvs), Sign(true, nVs), eq))
		// Or(sign(true, new_n(Vs)), sign(false, n(Vs)), eq)
		parts = append(parts, disjoin(Sign(true, newNvs), Sign(false, nVs), eq))
	}

	tr := conjoin(parts...)
	return makeUpdate([]*Const{relSym}, tr, False, EmptyAnnotation{})
}

// --- NativeAction ---

// IntUpdate for NativeAction is a no-op — skips update axioms.
// Python: NativeAction.int_update (ivy_actions.py:1286) returns
// ([], true_clauses(), false_clauses()) — annot is None.
func (a *LogicNativeAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
	return NullUpdate()
}

// --- DebugAction ---

// IntUpdate for DebugAction is a no-op — skips update axioms.
// Python: DebugAction.int_update (ivy_actions.py:1267) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
func (a *LogicDebugAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
	return &Update{
		Modified: []*Const{},
		TR:       TrueClauses(EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
}

// --- ReturnAction ---

// IntUpdate for ReturnAction is a no-op — skips update axioms.
// Python: ReturnAction.int_update (ivy_actions.py:1664) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
// In Python, ReturnAction is a bare `object` subclass (not Action), so it
// bypasses Action.int_update; both sides emit the same ENTER trace explicitly.
func (a *ReturnAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
	return &Update{
		Modified: []*Const{},
		TR:       TrueClauses(EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
}

// --- Field Actions ---

// makeFieldUpdateFunc constructs a field-update AssignAction using a callable
// RHS builder and returns its ActionUpdate.
// Python: make_field_update(self, l, f, r_func, domain, pvars) with lambda r_func.
func makeFieldUpdateFunc(field, obj Expr, rhsFunc func(v *LogicVariable) Expr, ctx *UpdateContext) *Update {
	sym, ok := field.(*Const)
	if sym == nil || !ok {
		panic(fmt.Sprintf("field %v must be a binary relation", field))
	}
	fs, ok := sym.CSort.(*LogicFunctionSort)
	if !ok || !IsRelationalSort(sym.CSort) || len(fs.Domain()) != 2 {
		panic(fmt.Sprintf("field %s must be a binary relation", sym.Name))
	}
	v, _ := NewVariable("X", fs.Domain()[1])
	// Build f(l, v) as Apply
	lhs, _ := NewApply(sym, obj, v)
	rhs := rhsFunc(v)
	aa := NewAssignAction(lhs, rhs)
	return aa.ActionUpdate(ctx)
}

// ActionUpdate for AssignFieldAction.
// Python: l,f,r = self.args; make_field_update(self,l,f,lambda v: Equals(v,r),domain,pvars)
func (a *LogicAssignFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *LogicVariable) Expr {
		return NewEqualsNode(v, a.Value)
	}, ctx)
}

// ActionUpdate for NullFieldAction.
// Python: l,f = self.args; make_field_update(self,l,f,lambda v: Or(),domain,pvars)
func (a *LogicNullFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *LogicVariable) Expr {
		return &LogicOr{} // Or() with no args = false
	}, ctx)
}

// ActionUpdate for CopyFieldAction.
// Python: l,lf,r,rf = self.args; make_field_update(self,l,lf,lambda v: rf(r,v),domain,pvars)
func (a *LogicCopyFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	srcField := a.SrcField
	if srcField == nil {
		srcField = a.Field // backward compat: same field for both
	}
	return makeFieldUpdateFunc(a.Field, a.Dst, func(v *LogicVariable) Expr {
		if sym, ok := srcField.(*Const); ok {
			app, _ := NewApply(sym, a.Src, v)
			return app
		}
		return v
	}, ctx)
}

// -----------------------------------------------------------------------
// IntUpdate implementations
// -----------------------------------------------------------------------

// actionTypeName returns the Python class name for the action type (for xtracer).
func ActionTypeName(a interface{}) string {
	switch a.(type) {
	case *LogicSequence:
		return "Sequence"
	case *LogicChoiceAction:
		return "ChoiceAction"
	case *LogicEnvAction:
		return "EnvAction"
	case *LogicIfAction:
		return "IfAction"
	case *LogicWhileAction:
		return "WhileAction"
	case *LogicLocalAction:
		return "LocalAction"
	case *LogicLetAction:
		return "LetAction"
	case *LogicCallAction:
		return "CallAction"
	case *LogicBindOldsAction:
		return "BindOldsAction"
	case *LogicAssignAction:
		return "AssignAction"
	case *LogicSetAction:
		return "SetAction"
	case *LogicHavocAction:
		return "HavocAction"
	case *LogicAssumeAction:
		return "AssumeAction"
	case *LogicAssertAction:
		return "AssertAction"
	case *LogicCrashAction:
		return "CrashAction"
	case *LogicInstantiateAction:
		return "InstantiateAction"
	case *LogicRanking:
		return "Ranking"
	case *LogicNativeAction:
		return "NativeAction"
	case *LogicDebugAction:
		return "DebugAction"
	case *LogicRequiresAction:
		return "RequiresAction"
	case *LogicEnsuresAction:
		return "EnsuresAction"
	case *LogicThunkAction:
		return "ThunkAction"
	case *ReturnAction:
		return "ReturnAction"
	case *IgnoreAction:
		return "IgnoreAction"
	case *LogicSubgoalAction:
		return "SubgoalAction"
	case *LogicVarAction:
		return "VarAction"
	case *LogicAssignFieldAction:
		return "AssignFieldAction"
	case *LogicNullFieldAction:
		return "NullFieldAction"
	case *LogicCopyFieldAction:
		return "CopyFieldAction"
	case *FailAction:
		return "fail_action" // Python class name is snake_case
	default:
		return fmt.Sprintf("%T", a)
	}
}

// IntUpdate is the intermediate update computation that applies domain
// update axioms on top of the atomic action_update.
// Corresponds to Python Action.int_update().
func IntUpdate(action ActionsAction, ctx *UpdateContext) *Update {
	// Dispatch to type-specific int_update methods.
	// Types with their own IntUpdate method have their own traces.
	// Types using intUpdateFromActionUpdate use the base Action.int_update
	// trace (matching Python's Action.int_update which traces
	// "actions.IntUpdate ENTER type=%s").
	switch a := action.(type) {
	case *LogicAssumeAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicAssertAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicSubgoalAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicRequiresAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicEnsuresAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicAssignAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicHavocAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicSetAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicNativeAction:
		return a.IntUpdate(ctx)
	case *LogicDebugAction:
		return a.IntUpdate(ctx)
	case *ReturnAction:
		return a.IntUpdate(ctx)
	case *LogicAssignFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicNullFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicCopyFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *LogicSequence:
		return a.IntUpdate(ctx)
	case *LogicChoiceAction:
		return a.IntUpdate(ctx)
	case *LogicEnvAction:
		return a.IntUpdateEnv(ctx)
	case *LogicIfAction:
		return a.IntUpdate(ctx)
	case *LogicWhileAction:
		return a.IntUpdate(ctx)
	case *LogicLocalAction:
		return a.IntUpdate(ctx)
	case *LogicLetAction:
		return a.IntUpdate(ctx)
	case *LogicCallAction:
		return a.IntUpdate(ctx)
	case *LogicBindOldsAction:
		return a.IntUpdate(ctx)
	case *LogicCrashAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *FailAction:
		// FailAction.IntUpdate emits its own "interp.ActionFail calling
		// IntUpdate type=<inner>" trace and recursively calls IntUpdate
		// on the inner action, matching Python fail_action.int_update.
		return a.IntUpdate(ctx)
	case *LogicInstantiateAction:
		return a.IntUpdate(ctx)
	default:
		// Generic fallback: null update
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return NullUpdate()
	}
}

// actionUpdater is the interface for actions that have ActionUpdate.
type actionUpdater interface {
	ActionUpdate(ctx *UpdateContext) *Update
}

// intUpdateFromActionUpdate computes int_update from action_update by
// applying domain update axioms.
func intUpdateFromActionUpdate(action actionUpdater, ctx *UpdateContext) *Update {
	update := action.ActionUpdate(ctx)
	// Apply update axioms from the domain
	update = applyUpdateAxioms(update, action.(ActionsAction), ctx)
	return update
}

// updateAxiomProvider matches the Updater interface in action.go.
// Implemented by PatternBasedUpdate, DerivedUpdate, and NamedUpdate.
type updateAxiomProvider = Updater

// applyUpdateAxioms applies domain.updates to the given update.
// In Python, this iterates over domain.updates calling get_update_axioms.
func applyUpdateAxioms(update *Update, action ActionsAction, ctx *UpdateContext) *Update {
	// Trace unconditionally, matching Python Action.int_update line 209
	// which always emits the trace even when len(domain.updates) == 0.
	if xtracer.Enabled {
		numUpdates := 0
		if ctx.Domain != nil {
			numUpdates = len(ctx.Domain.Updates)
		}
		show := make([]string, len(update.Modified))
		for i, s := range update.Modified {
			show[i] = fmt.Sprintf("'%v'", s.Name)
		}
		sort.Strings(show)
		xtracer.Trace("actions.applyUpdateAxioms ENTER numUpdates=%d modNames=[%v]", numUpdates, strings.Join(show, ", "))
	}
	if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
		return update
	}

	modified := update.Modified
	tr := update.TR
	pre := update.Pre
	for _, u := range ctx.Domain.Updates {
		if provider, ok := u.(updateAxiomProvider); ok {
			newMod, transrelNode, precondNode := provider.GetUpdateAxioms(modified, action)
			if xtracer.Enabled && (len(modified) > 0 || len(newMod) > 0) {
				oldShow := make([]string, len(modified))
				for j, s := range modified {
					oldShow[j] = fmt.Sprintf("'%v'", s.Name)
				}
				sort.Strings(oldShow)
				newShow := make([]string, len(newMod))
				for j, s := range newMod {
					newShow[j] = fmt.Sprintf("'%v'", s.Name)
				}
				sort.Strings(newShow)
				xtracer.Trace("actions.applyUpdateAxioms axiom modNames=[%v] -> newModNames=[%v] hasTR=%v hasPre=%v", strings.Join(oldShow, ", "), strings.Join(newShow, ", "), transrelNode != nil, precondNode != nil)
			}
			modified = newMod
			if transrelNode != nil {
				tr = AndClausesTyped(tr, transrelNode)
			}
			if precondNode != nil {
				pre = OrClausesTyped(pre, precondNode)
			}
		}
	}

	return &Update{
		Modified: modified,
		TR:       tr,
		Pre:      pre,
	}
}

// --- Sequence ---

// IntUpdate computes the sequential composition of child updates.
// Python: Sequence.int_update (ivy_actions.py:839) composes each child via
// compose_updates and pins the source op's lineno onto the resulting TR
// annotation after each composition.
func (s *LogicSequence) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.Sequence.int_update ENTER")
	defer xtracer.Trace("actions.Sequence.int_update EXIT")
	// Python (ivy_actions.py:841):
	//   update = ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
	result := &Update{
		Modified: []*Const{},
		TR:       TrueClauses(EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
	axioms := ctx.BackgroundTheory()
	for i, child := range s.Elems {
		act := unwrapToAction(child)
		if act == nil {
			// Python (ivy_actions.py:843) iterates blindly; calling int_update
			// on a non-Action would AttributeError. The faithful port panics
			// rather than silently skipping.
			panic(fmt.Sprintf("Sequence.IntUpdate: child %d is not an Action: %T", i, child))
		}
		childUpdate := IntUpdate(act, ctx)
		if xtracer.Enabled {
			childModNames := make([]string, len(childUpdate.Modified))
			for j, m := range childUpdate.Modified {
				childModNames[j] = fmt.Sprintf("'%v'", m.Name)
			}
			sort.Strings(childModNames)
			xtracer.Trace("actions.Sequence.int_update compose[%d] childType=%s childModified=[%v]", i, ActionTypeName(act), strings.Join(childModNames, ", "))
		}
		result = ComposeUpdates(result, axioms, childUpdate)
		if xtracer.Enabled {
			resultModNames := make([]string, len(result.Modified))
			for j, m := range result.Modified {
				resultModNames[j] = fmt.Sprintf("'%v'", m.Name)
			}
			sort.Strings(resultModNames)
			xtracer.Trace("actions.Sequence.int_update compose[%d] resultModified=[%v]", i, strings.Join(resultModNames, ", "))
		}
		// Python (ivy_actions.py:854):
		//   if hasattr(op,'lineno') and update[1].annot is not None:
		//       update[1].annot.lineno = op.lineno
		// After ComposeUpdates uses composeAnnotOp, result.TR.Annot is a
		// *ComposeAnnotation (or nil).
		if act.HasLineno() && result.TR != nil && result.TR.Annot != nil {
			if compAnnot, ok := result.TR.Annot.(*ComposeAnnotation); ok {
				loc := act.GetLineno()
				compAnnot.Lineno = &loc
			}
		}
	}
	return result
}

// unwrapToAction extracts an Action from a Node.
func unwrapToAction(n Expr) ActionsAction {
	act, _ := n.(ActionsAction)
	return act
}

// sortedModNames formats a slice of modified-symbol Consts as a sorted slice
// of "'name'"-quoted strings, matching the Python xtracer trace format
// produced by sorted([s.name for s in modified]).
func sortedModNames(mods []*Const) []string {
	names := make([]string, len(mods))
	for i, m := range mods {
		names[i] = fmt.Sprintf("'%v'", m.Name)
	}
	sort.Strings(names)
	return names
}

// --- ChoiceAction ---

// IntUpdate computes the nondeterministic choice between branches.
// Python: ChoiceAction.int_update uses join_action for each branch.
func (a *LogicChoiceAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.ChoiceAction.int_update ENTER")
	defer xtracer.Trace("actions.ChoiceAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = LogicIfAction(Not(cond), self.args[0], self.args[1])
	//   return ite.int_update(domain, pvars)
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(&LogicNot{Body: cond}, a.Branches[0], a.Branches[1])
		return ite.IntUpdate(ctx)
	}
	// Python (ivy_actions.py:889):
	//   result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
	result := &Update{
		Modified: []*Const{},
		TR:       FalseClauses(EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
	axioms := ctx.BackgroundTheory()
	for i, branch := range a.Branches {
		act := unwrapToAction(branch)
		if act == nil {
			// Python iterates blindly; calling int_update on a non-Action
			// would AttributeError. The faithful port panics rather than
			// silently skipping.
			panic(fmt.Sprintf("ChoiceAction.IntUpdate: branch %d is not an Action: %T", i, branch))
		}
		branchUpdate := IntUpdate(act, ctx)
		result = JoinAction(result, branchUpdate, axioms)
		if xtracer.Enabled {
			xtracer.Trace("actions.ChoiceAction.int_update branch[%d] childType=%s childModified=[%v]", i, ActionTypeName(act), strings.Join(sortedModNames(branchUpdate.Modified), ", "))
		}
	}
	return result
}

// --- EnvAction ---

// IntUpdateEnv is like ChoiceAction.IntUpdate but calls GetUpdate
// (with hide_formals) instead of IntUpdate for each branch.
func (a *LogicEnvAction) IntUpdateEnv(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.EnvAction.int_update ENTER")
	defer xtracer.Trace("actions.EnvAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = LogicIfAction(cond, self.args[0], self.args[1])
	//   return ite.update(domain, pvars)
	// Note: EnvAction uses cond (positive), ChoiceAction uses Not(cond).
	// Note: EnvAction calls update (GetUpdate), not int_update (IntUpdate).
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(cond, a.Branches[0], a.Branches[1])
		return GetUpdate(ite, ctx)
	}
	// Python (ivy_actions.py:917):
	//   result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
	result := &Update{
		Modified: []*Const{},
		TR:       FalseClauses(EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
	axioms := ctx.BackgroundTheory()
	for i, branch := range a.Branches {
		act := unwrapToAction(branch)
		if act == nil {
			// Python iterates blindly; calling .update on a non-Action would
			// AttributeError. The faithful port panics rather than silently
			// skipping.
			panic(fmt.Sprintf("EnvAction.IntUpdateEnv: branch %d is not an Action: %T", i, branch))
		}
		branchUpdate := GetUpdate(act, ctx)
		result = JoinAction(result, branchUpdate, axioms)
		if xtracer.Enabled {
			xtracer.Trace("actions.EnvAction.int_update branch[%d] childType=%s childModified=[%v]", i, ActionTypeName(act), strings.Join(sortedModNames(branchUpdate.Modified), ", "))
		}
	}
	return result
}

// --- IfAction ---

// IntUpdate computes the if-then-else transition relation.
// Python: IfAction.int_update uses ite_action for simple conditions.
func (a *LogicIfAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IfAction.int_update ENTER")
	defer xtracer.Trace("actions.IfAction.int_update EXIT")
	cond := a.Cond

	// Python: if used_variables_ast(self.args[0]): raise IvyError(...)
	freeVars := UsedVariablesAST(cond)
	if len(freeVars) > 0 {
		panic("variables in \"if\" conditions must be explicitly quantified")
	}

	// Python: if not isinstance(self.args[0], ivy_ast.Some): ... else: subactions path
	if _, isSome := cond.(*SomeCondition); isSome {
		return a.intUpdateWithSubactions(ctx)
	}

	// Python (ivy_actions.py:990): if not is_boolean(self.args[0]): raise IvyError("condition must be boolean")
	if !IsBoolean(cond) {
		panic("condition must be boolean")
	}

	// Simple boolean condition
	thenBranch := a.ThenBody
	var elseBranch Expr
	if a.ElseBody != nil {
		elseBranch = a.ElseBody
	}

	thenAct := unwrapToAction(thenBranch)
	if thenAct == nil {
		thenAct = NewSequence()
	}
	var elseAct ActionsAction
	if elseBranch != nil {
		elseAct = unwrapToAction(elseBranch)
	}
	if elseAct == nil {
		elseAct = NewSequence()
	}

	thenUpdate := IntUpdate(thenAct, ctx)
	elseUpdate := IntUpdate(elseAct, ctx)
	if xtracer.Enabled {
		xtracer.Trace("actions.IfAction.int_update then childType=%s thenModified=[%v]", ActionTypeName(thenAct), strings.Join(sortedModNames(thenUpdate.Modified), ", "))
		xtracer.Trace("actions.IfAction.int_update else childType=%s elseModified=[%v]", ActionTypeName(elseAct), strings.Join(sortedModNames(elseUpdate.Modified), ", "))
	}

	axioms := ctx.BackgroundTheory()
	return IteAction(cond, thenUpdate, elseUpdate, axioms)
}

// intUpdateWithSubactions handles the Some/SomeMinMax case.
// Python: if_part,else_part = (a.int_update(domain,pvars) for a in self.subactions())
func (a *LogicIfAction) intUpdateWithSubactions(ctx *UpdateContext) *Update {
	ifPart, elsePart := a.Subactions(ctx.ActCfg)

	ifUpdate := IntUpdate(ifPart, ctx)
	elseUpdate := IntUpdate(elsePart, ctx)
	if xtracer.Enabled {
		xtracer.Trace("actions.IfAction.int_update then childType=%s thenModified=[%v]", ActionTypeName(ifPart), strings.Join(sortedModNames(ifUpdate.Modified), ", "))
		xtracer.Trace("actions.IfAction.int_update else childType=%s elseModified=[%v]", ActionTypeName(elsePart), strings.Join(sortedModNames(elseUpdate.Modified), ", "))
	}

	axioms := ctx.BackgroundTheory()
	res := JoinAction(ifUpdate, elseUpdate, axioms)

	// Python hack (lines 930-934): the IteAnnotation comes out reversed after
	// join_action. Fix it by swapping ThenB/ElseB and negating Cond.
	if res.TR != nil {
		if ite, ok := res.TR.Annot.(*IteAnnotation); ok {
			res.TR.Annot = &IteAnnotation{
				Cond:  &LogicNot{Body: ite.Cond},
				ThenB: ite.ElseB,
				ElseB: ite.ThenB,
			}
		}
	}
	if res.Pre != nil {
		if ite, ok := res.Pre.Annot.(*IteAnnotation); ok {
			res.Pre.Annot = &IteAnnotation{
				Cond:  &LogicNot{Body: ite.Cond},
				ThenB: ite.ElseB,
				ElseB: ite.ThenB,
			}
		}
	}

	return res
}

// --- WhileAction ---

// IntUpdate computes the while loop's transition relation by expanding
// the loop into assume/assert/havoc/if structure.
// Python: WhileAction.int_update checks for UnrollContext first, then calls expand().
func (a *LogicWhileAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.WhileAction.int_update ENTER")
	defer xtracer.Trace("actions.WhileAction.int_update EXIT")
	// Python: if isinstance(context, UnrollContext): return self.unroll(context.card).int_update(domain, pvars)
	var actCtx IActionContext
	if ctx.ActCfg != nil {
		actCtx = ctx.ActCfg.Context
	}
	if uc, ok := actCtx.(*UnrollContext); ok {
		unrolled, err := a.Unroll(uc.Card, nil)
		if err != nil {
			// Python (ivy_actions.py:1117-1120): unroll raises IvyError
			// ("cannot determine an iteration bound" / "cowardly refusing
			// to unroll") and does not fall back to expand. Faithful port
			// propagates the error rather than silently expanding.
			panic(err.Error())
		}
		return IntUpdate(unrolled, ctx)
	}
	expanded := a.Expand(ctx)
	return IntUpdate(expanded, ctx)
}

// Unroll determines the iteration bound from the loop condition's index sort
// and unrolls the loop into nested IfActions.
// Python: WhileAction.unroll (ivy_actions.py:1025-1046)
func (a *LogicWhileAction) Unroll(card func(Sort) int, body ActionsAction) (ActionsAction, error) {
	cond := a.Cond
	// Unwrap nested And to find comparison
	for {
		if andN, ok := cond.(*LogicAnd); ok && len(andN.Terms) > 0 {
			cond = andN.Terms[0]
		} else {
			break
		}
	}
	// Determine index sort from condition
	var idxSort Sort
	if app, ok := cond.(*Apply); ok {
		if sym, ok := app.Func.(*Const); ok {
			if sym.Name == "<" || sym.Name == ">" || sym.Name == "<=" || sym.Name == ">=" {
				if len(app.Terms) > 0 {
					idxSort = app.Terms[0].NodeSort()
				}
			}
		}
	} else if notN, ok := cond.(*LogicNot); ok {
		if eq, ok := notN.Body.(*Eq); ok {
			idxSort = eq.T1.NodeSort()
		}
	}

	cardsort := card(idxSort)
	sortName := "unknown sort"
	if idxSort != nil {
		sortName = idxSort.String()
	}
	if cardsort <= 0 {
		return nil, fmt.Errorf("cannot determine an iteration bound for loop over %s", sortName)
	}
	if cardsort > 100 {
		return nil, fmt.Errorf("cowardly refusing to unroll loop over %s %d times", sortName, cardsort)
	}

	// Build nested IfActions from inside out
	// Python: res = LogicIfAction(self.args[0], AssumeAction(Or()))
	var bodyExpr Expr
	if body != nil {
		bodyExpr = body
	} else {
		bodyExpr = a.Body
	}

	// Innermost: if cond then assume false (empty Or = false)
	res := NewIfAction(a.Cond, NewAssumeAction(&LogicOr{}))
	for i := 0; i < cardsort; i++ {
		seq := NewSequence(bodyExpr, res)
		res = NewIfAction(a.Cond, seq)
	}
	a.CopyFormalsTo(res)
	return res, nil
}

// Expand converts the while loop into an equivalent sequence of
// assert invariants, havoc modified, assume invariants, if cond then body.
// This is the standard Floyd-Hoare approach to while loops.
func (a *LogicWhileAction) Expand(ctx *UpdateContext) ActionsAction {
	// First, compute the modify set from the body
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		return NewSequence()
	}
	bodyUpdate := IntUpdate(bodyAct, ctx)
	modset := bodyUpdate.Modified

	// Separate invariants from ranking
	var invariants []Expr
	var ranking ActionsAction
	for _, inv := range a.Invariants {
		if r, ok := inv.(ActionsAction); ok {
			if r.Name() == "decreases" {
				ranking = r
				continue
			}
		}
		invariants = append(invariants, inv)
	}

	// Build assert invariants.
	// Python keeps invariant nodes as actions:
	//   asserts = self.args[2:]
	//   asserts = [a for a in asserts if not isinstance(a, AssumeAction)]
	var asserts []ActionsAction
	for _, inv := range invariants {
		if act := extractActionFromNode(inv); act != nil {
			if _, isAssume := act.(*LogicAssumeAction); !isAssume {
				asserts = append(asserts, act)
			}
		} else {
			asserts = append(asserts, NewAssertAction(inv))
		}
	}

	// Build assume invariants (assert→assume conversion)
	// Python ivy_actions.py:1069: assumes = [a.assert_to_assume([AssertAction])
	//                                         for a in asserts if not isinstance(a, SubgoalAction)]
	var assumes []ActionsAction
	assertKinds := map[string]bool{"assert": true}
	var iuCfg *IvyUtilsConfig
	if ctx != nil && ctx.Domain != nil && ctx.Domain.Cfg != nil {
		iuCfg = ctx.Domain.Cfg.IuCfg
	}
	for _, inv := range invariants {
		if act := extractActionFromNode(inv); act != nil {
			if _, isSG := act.(*LogicSubgoalAction); !isSG {
				assumes = append(assumes, AssertToAssume(act, assertKinds, iuCfg))
			}
		} else {
			assumes = append(assumes, NewAssumeAction(inv))
		}
	}

	// Build havocs for modified symbols
	// Python ivy_actions.py:1083-1085: for h in havocs: h.lineno = self.lineno
	var havocs []ActionsAction
	for _, sym := range modset {
		h := NewHavocAction(sym)
		if a.HasLineno() {
			h.SetLineno(a.GetLineno())
		}
		havocs = append(havocs, h)
	}

	// Handle ranking function if present
	var entryAsserts, exitAsserts []ActionsAction
	var rankLocal *Const
	if ranking != nil {
		rankArgs := ranking.ActionArgs()
		if len(rankArgs) > 0 {
			rankExpr := rankArgs[0]
			rankSort := rankExpr.NodeSort()
			aux := NewConst("$rank", rankSort)
			rankLocal = aux
			rankAssume := NewAssumeAction(&Eq{T1: aux, T2: rankExpr})
			if ranking.HasLineno() {
				rankAssume.SetLineno(ranking.GetLineno())
			}
			assumes = append(assumes, rankAssume)
			ltSym := NewConst("<", LogicRelationSort([]Sort{rankSort, rankSort}))
			exitAssert := NewAssertAction(MustApply(ltSym, rankExpr, aux))
			if ranking.HasLineno() {
				exitAssert.SetLineno(ranking.GetLineno())
			}
			exitAsserts = append(exitAsserts, exitAssert)
			zeroSym := NewConst("0", rankSort)
			entryAssert := NewAssertAction(&LogicNot{Body: MustApply(ltSym, rankExpr, zeroSym)})
			if ranking.HasLineno() {
				entryAssert.SetLineno(ranking.GetLineno())
			}
			entryAsserts = append(entryAsserts, entryAssert)
		}
	}

	// Build the expanded action:
	// asserts; havocs; assumes; if cond then (entry_asserts; body; exit_asserts; asserts; assume false)
	var bodyParts []Expr
	for _, ea := range entryAsserts {
		bodyParts = append(bodyParts, ea)
	}
	bodyParts = append(bodyParts, a.Body)
	for _, ea := range exitAsserts {
		bodyParts = append(bodyParts, ea)
	}
	for _, asrt := range asserts {
		bodyParts = append(bodyParts, asrt)
	}
	// assume false (terminates this path — loop exit is the else branch)
	bodyParts = append(bodyParts, NewAssumeAction(False))
	thenBody := NewSequence(bodyParts...)

	ifAction := NewIfAction(a.Cond, thenBody, NewSequence())

	// Assemble the full expanded sequence
	var allParts []Expr
	for _, asrt := range asserts {
		allParts = append(allParts, asrt)
	}
	for _, h := range havocs {
		allParts = append(allParts, h)
	}
	for _, asms := range assumes {
		allParts = append(allParts, asms)
	}
	allParts = append(allParts, ifAction)

	result := NewSequence(allParts...)

	// If there's a ranking function, wrap in LocalAction
	if rankLocal != nil {
		return NewLocalActionOn(ctx.ActCfg, "actions.WhileAction.action_update", rankLocal, result)
	}
	return result
}

// --- LocalAction ---

// IntUpdate computes the local action's update by hiding local symbols.
// Python: LocalAction.int_update computes body.int_update then hide(syms, update).
func (a *LogicLocalAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.LocalAction.int_update ENTER")
	defer xtracer.Trace("actions.LocalAction.int_update EXIT")
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		// Python (ivy_actions.py:1148) iterates blindly; calling int_update
		// on a non-Action would AttributeError. Faithful port panics.
		panic(fmt.Sprintf("LocalAction.IntUpdate: body is not an Action: %T", a.Body))
	}
	update := IntUpdate(bodyAct, ctx)

	if xtracer.Enabled {
		modNames := make([]string, len(update.Modified))
		for i, m := range update.Modified {
			modNames[i] = fmt.Sprintf("'%v'", m.Name)
		}
		sort.Strings(modNames)
		xtracer.Trace("actions.LocalAction.int_update bodyType=%s bodyModified=[%v]", ActionTypeName(bodyAct), strings.Join(modNames, ", "))
	}

	// Collect symbols to hide. Python (ivy_actions.py:1154):
	//   syms = self.args[0:-1]
	// All callers (compiler.compile_local_action, compile_proof, etc.)
	// construct LocalActions with *lg.Const locals — a type assertion failure
	// would mean an upstream construction bug.
	symsToHide := make([]*Const, len(a.Locals))
	for i, local := range a.Locals {
		c, ok := local.(*Const)
		if !ok {
			panic(fmt.Sprintf("LocalAction.IntUpdate: local %d is not *lg.Const: %T", i, local))
		}
		symsToHide[i] = c
	}

	// Python: res = hide(syms, update) — always called, even with empty syms.
	return Hide(symsToHide, update)
}

// --- LetAction ---

// IntUpdate computes the let action's update by substituting symbols.
// Python: LetAction.int_update computes body.int_update then subst_action.
func (a *LogicLetAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.LetAction.int_update ENTER")
	defer xtracer.Trace("actions.LetAction.int_update EXIT")
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		// Python (ivy_actions.py:1179) iterates blindly; calling int_update
		// on a non-Action would AttributeError. Faithful port panics.
		panic(fmt.Sprintf("LetAction.IntUpdate: body is not an Action: %T", a.Body))
	}
	update := IntUpdate(bodyAct, ctx)

	// Build substitution map from bindings.
	// Python: subst = dict((a.args[0].rep, a.args[1].rep) for a in self.args[0:-1])
	subst := make(map[string]string)
	for _, binding := range a.Bindings {
		// Each binding is an assignment-like node: lhs = rhs
		args := binding.Children()
		if len(args) >= 2 {
			lhsName := constName(args[0])
			rhsName := constName(args[1])
			if lhsName != "" && rhsName != "" {
				subst[lhsName] = rhsName
			}
		}
	}

	// Python: res = subst_action(update, subst) — always called.
	return SubstAction(update, subst)
}

// --- BindOldsAction ---

// IntUpdate wraps the inner action's update with bind_olds.
// Python: BindOldsAction.int_update returns bind_olds_action(inner.int_update(...)).
func (a *LogicBindOldsAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.BindOldsAction.int_update ENTER")
	defer xtracer.Trace("actions.BindOldsAction.int_update EXIT")
	innerAct := unwrapToAction(a.Inner)
	if innerAct == nil {
		// Python (ivy_actions.py:1296) iterates blindly; calling int_update
		// on a non-Action would AttributeError. Faithful port panics.
		panic(fmt.Sprintf("BindOldsAction.IntUpdate: inner is not an Action: %T", a.Inner))
	}
	update := IntUpdate(innerAct, ctx)
	return BindOldsUpdate(update)
}

// --- CallAction ---

// IntUpdate computes the call action's update by inlining the callee.
// Python: CallAction.int_update resolves the callee, applies actuals,
// and computes the inlined update.
func (a *LogicCallAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.CallAction.int_update ENTER")
	calleeName := constName(a.Callee)
	if xtracer.Enabled {
		fmt.Printf("DIAG CallAction.int_update callee=%v\n", calleeName)
	}
	if calleeName == "" {
		// Python (ivy_actions.py:1318): name = self.args[0].rep — would
		// AttributeError if .rep is missing. Faithful port panics.
		panicf("CallAction.IntUpdate: callee has no name: '%#v'", a.Callee)
	}

	// Resolve the callee
	var calleeAction ActionsAction
	if ctx.GetAction != nil {
		calleeAction = ctx.GetAction(calleeName)
	}
	if calleeAction == nil {
		// Try from domain.Actions
		if ctx.Domain != nil && ctx.Domain.Actions != nil {
			if v, ok := ctx.Domain.Actions.Get2(calleeName); ok {
				act, isAct := v.(ActionsAction)
				if !isAct {
					// Python (ivy_actions.py:1350): v = state_to_action(v.value)
					//   for non-Action context entries (state-style updates with .value).
					// Go's module.Module.Actions is typed *iu.InsMap[string, Action]
					// (module.go:38), so the type system prevents non-Action values
					// from ever being stored. The state_to_action(v.value) branch
					// is unreachable in Go's typed model. If this panic ever fires,
					// the upstream construction site is violating the type contract
					// — investigate that, do not port state_to_action(v.value).
					panic(fmt.Sprintf("CallAction.IntUpdate: callee %s resolved to non-Action %T (Domain.Actions is typed Action; non-Action means a Go porting error and we need to change the Go types to accommodate/port more faithfully.", calleeName, v))
				}
				calleeAction = act
			}
		}
	}
	if calleeAction == nil {
		// Python (ivy_actions.py:1322): raise IvyError(self, "no value for {}")
		panic(fmt.Sprintf("CallAction.IntUpdate: no value for %s", calleeName))
	}

	// Apply actual parameters
	result := a.applyActuals(ctx, calleeAction)
	xtracer.Trace("actions.CallAction.int_update EXIT")
	return result
}

// applyActuals inlines the callee with actual parameters.
// Corresponds to Python CallAction.apply_actuals.
// Includes capture avoidance via distinct_obj_renaming.
func (a *LogicCallAction) applyActuals(ctx *UpdateContext, callee ActionsAction) *Update {
	formalParams := callee.GetFormalParams()
	formalReturns := callee.GetFormalReturns()
	actualParams := nodeArgs(a.Callee)
	actualReturns := a.ActualReturns

	// Capture avoidance: rename formals to avoid colliding with actuals.
	// Python: vocab = list(symbols_asts(actual_params+actual_returns))
	//         subst = distinct_obj_renaming(formal_params+formal_returns, vocab)
	allFormals := make([]*Const, 0, len(formalParams)+len(formalReturns))
	allFormals = append(allFormals, formalParams...)
	allFormals = append(allFormals, formalReturns...)

	// Collect names used in actual parameters and returns
	vocabNames := make(map[string]bool)
	for _, ap := range actualParams {
		actionsCollectSymbolNames(ap, vocabNames)
	}
	for _, ar := range actualReturns {
		actionsCollectSymbolNames(ar, vocabNames)
	}

	// Build renaming: formal → fresh name
	renaming := distinctObjRenaming(allFormals, vocabNames)

	// Apply renaming to callee — always call SubstituteConstantsAction
	// (matches Python which unconditionally calls substitute_constants_ast)
	substMap := make(map[NodeKey]Expr)
	seenFormals := make(map[NodeKey]bool, len(allFormals))
	for _, oldSym := range allFormals {
		oldKey := Key(oldSym)
		if seenFormals[oldKey] {
			continue
		}
		seenFormals[oldKey] = true
		newSym, ok := renaming[oldKey]
		if !ok {
			panic(fmt.Sprintf("CallAction.applyActuals: missing formal renaming for %s", oldSym.Name))
		}
		substMap[oldKey] = newSym
		// Python (ivy_actions.py:1364-1365): subst[old(s)] = old(t)
		//   where old(sym) = sym.prefix('old_'). Use the LogicOld() helper here
		//   so the substMap key actually matches a real pre-state symbol in
		//   the callee body (e.g., from a prior BindOldsAction wrapper).
		oldOfOldSym := NewConst(LogicOld(oldSym.Name), oldSym.CSort)
		oldOfNewSym := NewConst(LogicOld(newSym.Name), newSym.CSort)
		substMap[Key(oldOfOldSym)] = oldOfNewSym
	}
	renamedCallee := SubstituteConstantsAction(callee, substMap)

	// Get renamed formals from the same substitution map Python uses.
	renamedFormalParams := make([]*Const, len(formalParams))
	for i, fp := range formalParams {
		renamedFormalParams[i] = callActionRenamedFormal(fp, substMap)
	}
	renamedFormalReturns := make([]*Const, len(formalReturns))
	for i, fr := range formalReturns {
		renamedFormalReturns[i] = callActionRenamedFormal(fr, substMap)
	}

	// Validate parameter counts after substitution, matching Python ordering.
	// Python (ivy_actions.py:1379): raise IvyError("wrong number of input parameters")
	if len(renamedFormalParams) != len(actualParams) {
		panic("wrong number of input parameters")
	}
	// Python (ivy_actions.py:1382): raise IvyError("wrong number of output parameters")
	if len(renamedFormalReturns) != len(actualReturns) {
		panic("wrong number of output parameters")
	}

	// Sort compatibility check.
	// Python (ivy_actions.py:1359-1361):
	//   for x,y in zip(formal_params,actual_params):
	//       if x.sort != y.sort and not domain.is_variant(x.sort, y.sort):
	//           raise IvyError("value for input parameter ... has wrong sort")
	if ctx.Domain != nil {
		for i, fp := range renamedFormalParams {
			fpSort := fp.CSort
			apSort := actualParams[i].NodeSort()
			if fpSort != nil && apSort != nil && fpSort != apSort {
				if !ctx.Domain.IsVariant(fpSort, apSort) {
					panic(fmt.Sprintf("value for input parameter %s has wrong sort", fp.Name))
				}
			}
		}
		// Python (ivy_actions.py:1362-1369): symmetric check on returns,
		// but with x and y reversed in is_variant: domain.is_variant(y.sort, x.sort).
		for i, fr := range renamedFormalReturns {
			frSort := fr.CSort
			arSort := actualReturns[i].NodeSort()
			if frSort != nil && arSort != nil && frSort != arSort {
				if !ctx.Domain.IsVariant(arSort, frSort) {
					panic(fmt.Sprintf("value for output parameter %s has wrong sort", fr.Name))
				}
			}
		}
	}

	// Build input assignments: formal := actual
	var inputAsgns []Expr
	for i, fp := range renamedFormalParams {
		if i < len(actualParams) {
			asgn := NewAssignAction(fp, actualParams[i])
			inputAsgns = append(inputAsgns, asgn)
		}
	}

	// Build output assignments: actual_return := formal_return
	var outputAsgns []Expr
	for i, fr := range renamedFormalReturns {
		if i < len(actualReturns) {
			asgn := NewAssignAction(actualReturns[i], fr)
			outputAsgns = append(outputAsgns, asgn)
		}
	}

	// Build: Sequence(input_asgns, BindOlds(callee), output_asgns)
	inputSeq := NewSequence(inputAsgns...)
	bindOlds := NewBindOldsAction(renamedCallee)
	outputSeq := NewSequence(outputAsgns...)
	fullSeq := NewSequence(inputSeq, bindOlds, outputSeq)

	update := IntUpdate(fullSeq, ctx)

	// Hide the renamed formal parameters and returns.
	// Python (ivy_actions.py:1377): res = hide(formal_params+formal_returns, res)
	// Always called, even with empty toHide.
	toHide := make([]*Const, 0, len(renamedFormalParams)+len(renamedFormalReturns))
	toHide = append(toHide, renamedFormalParams...)
	toHide = append(toHide, renamedFormalReturns...)
	return Hide(toHide, update)
}

func callActionRenamedFormal(sym *Const, substMap map[NodeKey]Expr) *Const {
	replacement, ok := substMap[Key(sym)]
	if !ok {
		panic(fmt.Sprintf("CallAction.applyActuals: missing formal substitution for %s", sym.Name))
	}
	renamed, ok := replacement.(*Const)
	if !ok {
		panic(fmt.Sprintf("CallAction.applyActuals: formal %s substituted with non-Const %T", sym.Name, replacement))
	}
	return renamed
}

// distinctObjRenaming creates a renaming from formals to fresh names
// that don't conflict with vocabNames.
// Corresponds to Python distinct_obj_renaming.
func distinctObjRenaming(formals []*Const, vocabNames map[string]bool) map[NodeKey]*Const {
	result := make(map[NodeKey]*Const)
	usedNames := make(map[string]bool)
	for k := range vocabNames {
		usedNames[k] = true
	}

	for _, sym := range formals {
		name := sym.Name
		if _, used := usedNames[name]; !used {
			usedNames[name] = true
			// No conflict — identity mapping (matches Python which always adds all formals).
			result[Key(sym)] = NewConst(name, sym.CSort)
			continue
		}
		// Need a fresh name
		newName := unusedNameWithBase(name, usedNames)
		usedNames[newName] = true
		result[Key(sym)] = NewConst(newName, sym.CSort)
	}
	return result
}

// unusedNameWithBase finds an unused name starting with base, using
// letter suffixes (a, b, ..., z, a0, ...) matching Python's
// constant_name_generator via ivy_utils.unused_name_with_base.
func unusedNameWithBase(base string, used map[string]bool) string {
	gen := ConstantNameGenerator()
	for {
		name := base + "_" + gen()
		if !used[name] {
			return name
		}
	}
}

// collectSymbolNames collects all symbol/constant names from a logic node.
func actionsCollectSymbolNames(node Expr, names map[string]bool) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *Const:
		names[n.Name] = true
	case *Apply:
		if n.Func != nil {
			actionsCollectSymbolNames(n.Func, names)
		}
		for _, t := range n.Terms {
			actionsCollectSymbolNames(t, names)
		}
		return
	}
	for _, c := range node.Children() {
		actionsCollectSymbolNames(c, names)
	}
}

// --- CrashAction ---

// ActionUpdate computes the crash action by havocing all non-spec mutable symbols.
// Python: CrashAction.action_update — wants update axioms applied via intUpdateFromActionUpdate.
func (a *LogicCrashAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.CrashAction.action_update ENTER")
	if constName(a.Target) == "" || ctx == nil || ctx.Domain == nil {
		xtracer.Trace("actions.CrashAction.action_update EXIT")
		return NullUpdate()
	}

	actCfg := &ActionsConfig{Context: NewActionContext(ctx.Domain)}
	symsToHavoc := ModifiesSingle(a, actCfg)
	var targetTerms []Expr
	if app, ok := a.Target.(*Apply); ok {
		targetTerms = app.Terms
	}
	var havocParts []Expr
	for _, sym := range symsToHavoc {
		dom := SortDomain(sym.CSort)
		if len(dom) < len(targetTerms) {
			panic(fmt.Sprintf("action %q cannot be applied to %s because of argument mismatch", a.String(), sym.Name))
		}
		for i, term := range targetTerms {
			if !SortEqual(term.NodeSort(), dom[i]) {
				panic(fmt.Sprintf("action %q cannot be applied to %s because of argument mismatch", a.String(), sym.Name))
			}
		}
		args := make([]Expr, 0, len(dom))
		args = append(args, targetTerms...)
		for idx, sort := range dom[len(targetTerms):] {
			v, err := NewVariable(fmt.Sprintf("X%d", idx), sort)
			if err != nil {
				panic(err.Error())
			}
			args = append(args, v)
		}
		var target Expr = sym
		if len(args) > 0 {
			target = MustApply(sym, args...)
		}
		havocParts = append(havocParts, NewHavocAction(target))
	}
	seq := NewSequence(havocParts...)
	xtracer.Trace("actions.CrashAction.action_update EXIT")
	return IntUpdate(seq, ctx)
}

// collectCrashSyms recursively collects symbols to havoc for a crash action.
func collectCrashSyms(domain *Module, name string, result *[]*Const) {
	if domain.Hierarchy != nil {
		if children, ok := domain.Hierarchy.Get2(name); ok && children.Len() > 0 {
			for child := range children.All() {
				fullName := name + "." + child
				if child == "spec" {
					continue
				}
				// Check if child.spec is in attributes (skip spec-protected)
				specName := fullName + ".spec"
				if domain.Attributes != nil {
					if _, ok := domain.Attributes[specName]; ok {
						continue
					}
				}
				collectCrashSyms(domain, fullName, result)
			}
			return
		}
	}
	// Leaf: check if it's a mutable symbol
	sym := NewConst(name, TopS)
	*result = append(*result, sym)
}

// -----------------------------------------------------------------------
// GetUpdate: the top-level entry point matching Python Action.update()
// -----------------------------------------------------------------------

// GetUpdate computes the full update for an action:
// 1. Compute int_update
// 2. Bind old values
// 3. Hide formal parameters and returns
//
// This corresponds to Python Action.update(domain, pvars).
func GetUpdate(action ActionsAction, ctx *UpdateContext) *Update {
	// FailAction overrides Python's Action.update (ivy_interp.py:385-389),
	// so it does NOT emit the base "actions.GetUpdate ENTER type=fail_action"
	// trace. Match that by delegating to fa.Update directly here, before
	// the ENTER trace fires.
	if fa, ok := action.(*FailAction); ok {
		return fa.Update(ctx)
	}
	xtracer.Trace("actions.GetUpdate ENTER type=%s", ActionTypeName(action))
	update := IntUpdate(action, ctx)
	update = BindOldsUpdate(update)
	update = hideFormals(action, update)
	return update
}

// hideFormals hides formal parameters and returns from the update.
// Matches Python Action.hide_formals (ivy_actions.py:220-228).
func hideFormals(action ActionsAction, update *Update) *Update {
	var toHide []*Const
	if fp := action.GetFormalParams(); len(fp) > 0 {
		toHide = append(toHide, fp...)
	}
	if fr := action.GetFormalReturns(); len(fr) > 0 {
		toHide = append(toHide, fr...)
	}
	if len(toHide) > 0 {
		update = Hide(toHide, update)
	}
	return update
}

// -----------------------------------------------------------------------
// Updater interface implementation for art/ package
// -----------------------------------------------------------------------

// GetUpdateForArt implements the Updater interface expected by art/art.go.
// It adapts the module-level GetUpdate function to the (domain, inScope) signature.
func GetUpdateForArt(action ActionsAction, domain *Module, inScope map[string]bool) *Update {
	ctx := &UpdateContext{
		Domain:       domain,
		PVars:        inScope,
		ActCfg:       domain.Cfg.ActCfg,
		Instantiator: domain.Instantiator,
		GetAction: func(name string) ActionsAction {
			if domain != nil && domain.Actions != nil {
				if v, ok := domain.Actions.Get2(name); ok {
					if act, ok := v.(ActionsAction); ok {
						return act
					}
				}
			}
			return nil
		},
	}
	return GetUpdate(action, ctx)
}
