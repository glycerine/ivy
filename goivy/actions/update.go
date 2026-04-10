// This file implements the action_update() and int_update() methods for
// each action type, porting the transition-relation computation from
// Python ivy_actions.py lines 309-1302.
//
// Each action's ActionUpdate method returns a *Update representing
// the transition relation (modified symbols, TR formula, precondition).
//
// IntUpdate applies update axioms from the domain on top of ActionUpdate.
// GetUpdate is the top-level entry point that hides formals and binds olds.
package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// -----------------------------------------------------------------------
// UpdateContext provides the domain and in-scope variables for computing
// action updates. It wraps the module and an action-resolution function.
// -----------------------------------------------------------------------

// UpdateContext holds the context needed for computing action updates.
type UpdateContext struct {
	Domain *module.Module
	PVars  map[string]bool // in-scope variable names

	// ActCfg is the per-session actions config. Used for context lookups
	// (replaces the old GlobalContext global).
	ActCfg *ActionsConfig

	// GetAction resolves an action name to its Action. This is set by
	// the caller (typically from module.Actions or an ActionContext).
	GetAction func(name string) Action

	// CompileActionBody compiles an AST node as an action body.
	// This callback avoids a circular import between actions and compiler.
	// Set by callers that have access to the compiler.
	// Used by InstantiateAction to compile macro expansions at runtime.
	CompileActionBody func(node ast.Node) (Action, error)

	// CheckUnprovable corresponds to Python's check_unprovable parameter.
	// When true, only unprovable assertions are checked.
	CheckUnprovable bool

	// CheckedAssert corresponds to Python's checked_assert parameter.
	// If non-empty, only the assertion whose lineno matches is fully checked.
	CheckedAssert string

	// Instantiator provides definition instances for clausification.
	// Corresponds to Python's global `instantiator` variable.
	// When non-nil, used by AssumeAction and AssertAction to unfold definitions.
	Instantiator func([]lg.Expr) *module.Clauses
}

// BackgroundTheory returns the background theory (axioms) for the domain.
func (ctx *UpdateContext) BackgroundTheory() *module.Clauses {
	if ctx.Domain == nil {
		return module.TrueClauses(nil)
	}
	clauses := ctx.Domain.BackgroundTheory(ctx.PVars)
	if clauses == nil {
		return module.TrueClauses(nil)
	}
	return clauses
}

// makeUpdate creates a transrel.Update from individual components,
// wrapping lg.Expr values into Clauses. This is a transitional helper
// for porting action updates from bare Node to Clauses.
func makeUpdate(modified []*lg.Const, tr lg.Expr, pre lg.Expr, annot interface{}) *Update {
	return &Update{
		Modified: modified,
		TR:       module.FormulaToClauses(tr, annot),
		Pre:      module.FormulaToClauses(pre, annot),
	}
}

// makeUpdateDefs creates a transrel.Update with definitions in the TR.
// This matches Python's pattern of Clauses([], [Definition(...)], annot).
func makeUpdateDefs(modified []*lg.Const, defs []*il.Definition, annot interface{}) *Update {
	return &Update{
		Modified: modified,
		TR:       module.NewClauses(nil, defs, annot),
		Pre:      module.FalseClauses(annot),
	}
}

// -----------------------------------------------------------------------
// Formula helpers
// -----------------------------------------------------------------------

// equivAST returns the equivalence of two nodes:
//   - If both are individual (non-Boolean): returns Eq(a, b)
//   - If both are Boolean: returns And(Or(a, Not(b)), Or(Not(a), b))
func equivAST(a, b lg.Expr) lg.Expr {
	if il.IsIndividual(a) {
		return &lg.Eq{T1: a, T2: b}
	}
	// Boolean equivalence: (a | ~b) & (~a | b)
	notA := module.Negate(a)
	notB := module.Negate(b)
	or1, _ := lg.NewOr(a, notB)
	or2, _ := lg.NewOr(notA, b)
	and, _ := lg.NewAnd(or1, or2)
	return and
}

// conjoin creates a conjunction, simplifying True cases.
func conjoin(nodes ...lg.Expr) lg.Expr {
	var terms []lg.Expr
	for _, n := range nodes {
		if isTrue(n) {
			continue
		}
		if isFalse(n) {
			return lg.False
		}
		terms = append(terms, n)
	}
	if len(terms) == 0 {
		return lg.True
	}
	if len(terms) == 1 {
		return terms[0]
	}
	and, _ := lg.NewAnd(terms...)
	return and
}

// disjoin creates a disjunction, simplifying False cases.
func disjoin(nodes ...lg.Expr) lg.Expr {
	var terms []lg.Expr
	for _, n := range nodes {
		if isFalse(n) {
			continue
		}
		if isTrue(n) {
			return lg.True
		}
		terms = append(terms, n)
	}
	if len(terms) == 0 {
		return lg.False
	}
	if len(terms) == 1 {
		return terms[0]
	}
	or, _ := lg.NewOr(terms...)
	return or
}

func isTrue(n lg.Expr) bool {
	return lg.IsTrue(n)
}

func isFalse(n lg.Expr) bool {
	return lg.IsFalse(n)
}

// -----------------------------------------------------------------------
// newSym returns the post-state version of a symbol.
// -----------------------------------------------------------------------

func newSym(sym *lg.Const) *lg.Const {
	return lg.NewConst(New(sym.Name), sym.CSort)
}

// constName extracts the name from a node that is a Const or the Func of an Apply.
func constName(n lg.Expr) string {
	switch t := n.(type) {
	case *lg.Const:
		return t.Name
	case *lg.Apply:
		return constName(t.Func)
	}
	return ""
}

// constSym extracts the Const from a node (Const or Apply.Func).
func constSym(n lg.Expr) *lg.Const {
	switch t := n.(type) {
	case *lg.Const:
		return t
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Const); ok {
			return c
		}
	}
	return nil
}

// nodeArgs extracts the args from a node (Apply.Terms or nil).
func nodeArgs(n lg.Expr) []lg.Expr {
	if app, ok := n.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// addParametersAST appends params to every application in the AST.
// For a Symbol, wraps it in Apply(sym, params...).
// For an Apply, appends params to its Terms.
// Corresponds to Python's add_parameters_ast (ivy_ast.py:1756).
func addParametersAST(node lg.Expr, params []lg.Expr) lg.Expr {
	if len(params) == 0 {
		return node
	}
	switch t := node.(type) {
	case *lg.Const:
		app, _ := lg.NewApply(t, params...)
		return app
	case *lg.Apply:
		newTerms := make([]lg.Expr, len(t.Terms)+len(params))
		copy(newTerms, t.Terms)
		copy(newTerms[len(t.Terms):], params)
		app, _ := lg.NewApply(t.Func, newTerms...)
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

func mkAssignClauses(lhs, rhs lg.Expr) *Update {
	sym := constSym(lhs)
	if sym == nil {
		// Fallback: no-op update
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// Placeholder variables for the full domain of the symbol
	phs := module.SymPlaceholders(sym)

	// Build new_n applied to placeholders
	phNodes := varsToNodes(phs)
	dlhs := applyToNodes(newN, phNodes)

	// Build equality conditions for non-variable args
	var eqs []lg.Expr
	rn := make(map[string]lg.Expr)
	for i, ph := range phs {
		if i < len(args) {
			if _, isVar := args[i].(*lg.Variable); isVar {
				// Variable arg: record substitution (arg.Name → placeholder)
				rn[args[i].(*lg.Variable).Name] = ph
			} else {
				// Non-variable: add equality constraint
				eqs = append(eqs, &lg.Eq{T1: ph, T2: args[i]})
			}
		}
	}

	// Substitute variable args in RHS
	drhs := module.SubstituteAstByName(rhs, rn)

	// If there are equality conditions, build ITE
	// Python: Ite(And(*eqs), drhs, n(*dlhs.args))
	if len(eqs) > 0 {
		eqConj, _ := lg.NewAnd(eqs...) // match Python And(*eqs) exactly
		// old value: n applied to placeholders
		oldVal := applyToNodes(sym, phNodes)
		rhsSort := rhs.NodeSort()
		if rhsSort == nil {
			rhsSort = lg.TopS
		}
		drhs = &lg.Ite{ISort: rhsSort, Cond: eqConj, Then: drhs, Else: oldVal}
	}

	// Python: Clauses([], [Definition(dlhs, drhs)], EmptyAnnotation())
	// Store as a Definition in Clauses.Defs, matching Python exactly.
	defn := il.NewDefinition(dlhs, drhs)
	return &Update{
		Modified: []*lg.Const{sym},
		TR:       module.NewClauses(nil, []*il.Definition{defn}, EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
	}
}

func varsToNodes(vars []*lg.Variable) []lg.Expr {
	nodes := make([]lg.Expr, len(vars))
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
func (a *AssumeAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssumeAction.action_update ENTER")
	defer xtracer.Trace("actions.AssumeAction.action_update EXIT")
	// Python: if isinstance(fmla, LabeledFormula) and fmla.unprovable: return skip
	if a.Unprovable {
		return makeUpdate([]*lg.Const{}, lg.True, lg.False, EmptyAnnotation{})
	}
	fmla := a.Formula
	// Python: clauses = formula_to_clauses_tseitin(skolemize_formula(fmla))
	//         clauses = unfold_definitions_clauses(clauses)
	//         clauses = Clauses(clauses.fmlas, clauses.defs, EmptyAnnotation())
	var skInst func([]lg.Expr) *module.Clauses
	if ctx != nil {
		skInst = ctx.Instantiator
	}
	fmla = module.SkolemizeFormula(fmla, nil, skInst)
	clauses := module.FormulaToClauses(fmla, nil)
	if ctx != nil && ctx.Instantiator != nil {
		clauses = module.UnfoldDefinitionsClauses(clauses, ctx.Instantiator)
	}
	clauses = module.NewClauses(clauses.Fmlas, clauses.Defs, EmptyAnnotation{})
	return &Update{
		Modified: []*lg.Const{},
		TR:       clauses,
		Pre:      module.FalseClauses(EmptyAnnotation{}),
	}
}

// --- AssertAction ---

// ActionUpdate computes the transition relation for AssertAction.
// An assert generates a precondition (negative) from the dual of the formula.
// Implements Python's selective assertion checking via check_unprovable and checked_assert.
// Python: action_update (ivy_actions.py:343-362)
func (a *AssertAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssertAction.action_update ENTER")
	defer xtracer.Trace("actions.AssertAction.action_update EXIT")
	fmla := a.Formula
	unprovable := a.Unprovable

	// Python: if check_unprovable.get() != unprovable: skip
	if ctx.CheckUnprovable != unprovable {
		return makeUpdate([]*lg.Const{}, lg.True, lg.False, EmptyAnnotation{})
	}

	// Python: if checked_assert is set and doesn't match this lineno
	if ctx.CheckedAssert != "" {
		if ctx.CheckedAssert != a.GetLineno().String() {
			if unprovable {
				// Unprovable assertion not selected: skip entirely
				return makeUpdate([]*lg.Const{}, lg.True, lg.False, EmptyAnnotation{})
			}
			// Provable assertion not selected: return formula as-is (not dual)
			return makeUpdate([]*lg.Const{}, fmla, lg.False, EmptyAnnotation{})
		}
	}

	// Only assertions that pass both filters get dual formula treatment
	// Python: cl = formula_to_clauses(dual_formula(fmla))
	//         cl = Clauses(cl.fmlas, cl.defs, EmptyAnnotation())
	// Python's dual_formula consults lu.instantiator (a module-level global);
	// Go's literal port reads it from ctx.Instantiator (set from Domain.Instantiator).
	var dualInst func([]lg.Expr) *module.Clauses
	if ctx != nil {
		dualInst = ctx.Instantiator
	}
	dual := module.DualFormula(fmla, nil, dualInst)
	cl := module.FormulaToClauses(dual, nil)
	cl = module.NewClauses(cl.Fmlas, cl.Defs, EmptyAnnotation{})
	return &Update{
		Modified: []*lg.Const{},
		TR:       module.TrueClauses(EmptyAnnotation{}),
		Pre:      cl,
	}
}

// --- RequiresAction ---

func (a *RequiresAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.AssertAction.ActionUpdate(ctx)
}

// --- EnsuresAction ---

func (a *EnsuresAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.AssertAction.ActionUpdate(ctx)
}

// --- SubgoalAction ---

func (a *SubgoalAction) ActionUpdate(ctx *UpdateContext) *Update {
	return a.AssertAction.ActionUpdate(ctx)
}

// --- AssignAction ---

// ActionUpdate computes the transition relation for assignment lhs := rhs.
// Handles:
// 1. Hierarchical symbols (decompose into sub-assignments)
// 2. Partial applications (add placeholder parameters)
// 3. Destructor assignments (mutate through destructors)
// 4. Variant assignments
// 5. Simple assignments
func (a *AssignAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.AssignAction.action_update ENTER")
	defer xtracer.Trace("actions.AssignAction.action_update EXIT")
	lhs, rhs := a.LHS, a.RHS
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}

	// Handle hierarchical case: if the symbol has children in the hierarchy
	if ctx.Domain != nil && ctx.Domain.Hierarchy != nil {
		if children, ok := ctx.Domain.Hierarchy.Get2(sym.Name); ok && children.Len() > 0 {
			xtracer.Trace("actions.AssignAction.action_update branch=hierarchy")
			// Decompose into sub-assignments for each child
			var updates []*Update
			axioms := ctx.BackgroundTheory()
			for childName := range children.All() {
				childSym := lg.NewConst(childName, lg.TopS)
				childLHS := lg.MustApply(childSym, nodeArgs(lhs)...)
				childRHS := rhs // simplified: same RHS for each child
				childAssign := NewAssignAction(childLHS, childRHS)
				childUpdate := childAssign.ActionUpdate(ctx)
				updates = append(updates, childUpdate)
			}
			if len(updates) == 0 {
				return NullUpdate()
			}
			result := updates[0]
			for _, u := range updates[1:] {
				result = ComposeUpdates(result, axioms, u)
			}
			return result
		}
	}

	// Partial application extension
	// Python: xtra = len(lhs.rep.sort.dom) - len(lhs.args)
	dom := il.SortDomain(sym.CSort)
	xtra := len(dom) - len(nodeArgs(lhs))
	if xtra < 0 {
		// too many parameters
		return NullUpdate()
	}
	if xtra > 0 {
		// Extend lhs and rhs with fresh placeholder variables
		phs := module.SymPlaceholders(sym)
		extend := make([]lg.Expr, xtra)
		for i := 0; i < xtra; i++ {
			extend[i] = phs[len(phs)-xtra+i]
		}
		// Make variables distinct from those already used in lhs and rhs
		// Python: extend = variables_distinct_list_ast(extend, self)
		// We combine lhs and rhs into a single expression for variable collection
		combined := &lg.And{Terms: []lg.Expr{lhs, rhs}}
		extend = module.VariablesDistinctListAst(extend, combined)

		lhs = addParametersAST(lhs, extend)
		// Assignment of individual to a boolean is a special case
		// Guard against nil sorts
		if rhs.NodeSort() != nil && lhs.NodeSort() != nil && il.IsIndividual(rhs) && !il.IsIndividual(lhs) {
			lastExt := extend[len(extend)-1]
			rhsExtended := addParametersAST(rhs, extend[:len(extend)-1])
			rhs = &lg.Eq{T1: lastExt, T2: rhsExtended}
		} else {
			rhs = addParametersAST(rhs, extend)
		}
	}

	// Variable check: all RHS variables must appear in LHS
	// Python: if any(v not in lhs_vars for v in used_variables_ast(rhs)): raise IvyError
	lhsVars := module.UsedVariablesAST(lhs)
	rhsVars := module.UsedVariablesAST(rhs)
	for k := range rhsVars {
		if _, found := lhsVars[k]; !found {
			// multiply assigned
			return NullUpdate()
		}
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
func destrAsgnVal(lhs lg.Expr, fmlas *[]lg.Expr, domain *module.Module) (lg.Expr, *module.Clauses, *lg.Const) {
	lhsArgs := nodeArgs(lhs)
	if len(lhsArgs) == 0 {
		return lhs, module.FalseClauses(nil), nil
	}

	mut := lhsArgs[0]
	rest := lhsArgs[1:]
	mutN := constSym(mut)
	if mutN == nil {
		return lhs, module.FalseClauses(nil), nil
	}

	var lval lg.Expr
	var newClauses *module.Clauses
	var mutated *lg.Const

	if _, ok := domain.DestructorSorts[mutN.Name]; ok {
		// Recursive case: nested destructor
		// Python: lval, new_clauses, mutated = destr_asgn_val(mut, fmlas)
		lval, newClauses, mutated = destrAsgnVal(mut, fmlas, domain)
	} else {
		// Base case: mut is the root mutable symbol
		// Python: nondet = mut_n.suffix("_nd").skolem()
		skSym := lg.NewConst(mutN.Name+"_nd", mutN.CSort)
		phs := module.SymPlaceholders(mutN)
		phNodes := varsToNodes(phs)
		// Python: new_clauses = mk_assign_clauses(mut_n, nondet(*sym_placeholders(mut_n)))
		var skApplied lg.Expr
		if len(phNodes) > 0 {
			skApplied = applyToNodes(skSym, phNodes)
		} else {
			skApplied = skSym
		}
		newClauses = mkAssignClauses(mutN, skApplied).TR
		// Python: lval = nondet(*mut.args)
		mutArgs := nodeArgs(mut)
		if len(mutArgs) > 0 {
			mutArgNodes := make([]lg.Expr, len(mutArgs))
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
	vs := module.SymPlaceholders(n)
	vsNodes := varsToNodes(vs)

	// dlhs = n(*([lval] + vs[1:]))
	dlhsArgs := make([]lg.Expr, len(vsNodes))
	dlhsArgs[0] = lval
	for i := 1; i < len(vsNodes); i++ {
		dlhsArgs[i] = vsNodes[i]
	}
	dlhs := applyToNodes(n, dlhsArgs)

	// drhs = n(*([mut] + vs[1:]))
	drhsArgs := make([]lg.Expr, len(vsNodes))
	drhsArgs[0] = mut
	for i := 1; i < len(vsNodes); i++ {
		drhsArgs[i] = vsNodes[i]
	}
	drhs := applyToNodes(n, drhsArgs)

	// eqs = [eq_atom(v,a) for (v,a) in list(zip(vs,lhs.args))[1:] if not isinstance(a,Variable)]
	var eqs []lg.Expr
	for i := 1; i < len(vs) && i < len(lhsArgs); i++ {
		if _, isVar := lhsArgs[i].(*lg.Variable); !isVar {
			eqs = append(eqs, &lg.Eq{T1: vs[i], T2: lhsArgs[i]})
		}
	}

	// if eqs: fmlas.append(Or(And(*eqs), equiv_ast(dlhs, drhs)))
	if len(eqs) > 0 {
		var guard lg.Expr
		if len(eqs) == 1 {
			guard = eqs[0]
		} else {
			guard = &lg.And{Terms: eqs}
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
		if destrs, ok := domain.SortDestructors[sortName]; ok {
			for _, destr := range destrs {
				if destr.Name != n.Name {
					destrPhs := module.SymPlaceholders(destr)
					destrPhNodes := varsToNodes(destrPhs)
					// a1 = [lval] + phs[1:]
					a1 := make([]lg.Expr, len(destrPhNodes))
					a1[0] = lval
					for j := 1; j < len(destrPhNodes); j++ {
						a1[j] = destrPhNodes[j]
					}
					// a2 = [mut] + phs[1:]
					a2 := make([]lg.Expr, len(destrPhNodes))
					a2[0] = mut
					for j := 1; j < len(destrPhNodes); j++ {
						a2[j] = destrPhNodes[j]
					}
					destrA1 := applyToNodes(destr, a1)
					destrA2 := applyToNodes(destr, a2)
					*fmlas = append(*fmlas, &lg.Eq{T1: destrA1, T2: destrA2})
				}
			}
		}
	}

	// return lhs.rep(*([lval]+rest)), new_clauses, mutated
	resultArgs := make([]lg.Expr, 1+len(rest))
	resultArgs[0] = lval
	copy(resultArgs[1:], rest)
	resultExpr := applyToNodes(n, resultArgs)

	return resultExpr, newClauses, mutated
}

// destructorAssignUpdate handles assignment through destructors.
// Python: destructor case in AssignAction.action_update (ivy_actions.py:533-538).
func (a *AssignAction) destructorAssignUpdate(ctx *UpdateContext, lhs, rhs lg.Expr) *Update {
	var fmlas []lg.Expr
	nondetLhs, newClauses, mutN := destrAsgnVal(lhs, &fmlas, ctx.Domain)
	if mutN == nil {
		return NullUpdate()
	}

	// Python: fmlas.append(equiv_ast(nondet_lhs, rhs))
	fmlas = append(fmlas, equivAST(nondetLhs, rhs))

	// Python: new_clauses = and_clauses(new_clauses, Clauses(fmlas))
	fmlaClauses := module.NewClauses(fmlas, nil, nil)
	combined := module.AndClausesTyped(newClauses, fmlaClauses)

	// Python: return ([mut_n], new_clauses, false_clauses(annot=EmptyAnnotation()))
	return &Update{
		Modified: []*lg.Const{mutN},
		TR:       combined,
		Pre:      module.FalseClauses(EmptyAnnotation{}),
	}
}

// isVariant checks if lhsSort has rhsSort as a variant.
func isVariant(domain *module.Module, lhsSort, rhsSort lg.Sort) bool {
	if domain.Variants == nil {
		return false
	}
	lhsName := il.SortName(lhsSort)
	variants, ok := domain.Variants[lhsName]
	if !ok {
		return false
	}
	for _, v := range variants {
		if lg.SortEqual(v, rhsSort) {
			return true
		}
	}
	return false
}

// mkVariantAssignClauses creates the transition relation for a variant assignment.
// Python: mk_variant_assign_clauses (ivy_actions.py:593-611).
// Asserts that the new value points-to the RHS via pto, and does NOT point-to
// any other variant sort.
func mkVariantAssignClauses(lhs, rhs lg.Expr, domain *module.Module) *Update {
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// dlhs = new_n(*sym_placeholders(n))
	phs := module.SymPlaceholders(sym)
	phNodes := varsToNodes(phs)
	dlhs := applyToNodes(newN, phNodes)
	vs := phs // dlhs.args are the placeholders

	// Build eqs for non-variable args
	// Python: eqs = [eq_atom(v,a) for (v,a) in zip(vs,args) if not isinstance(a,Variable)]
	var eqs []lg.Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*lg.Variable); !isVar {
				eqs = append(eqs, &lg.Eq{T1: v, T2: args[i]})
			}
		}
	}

	// Build rename map for variable args, compute drhs = substitute_ast(rhs, rn)
	// Python: rn = dict((a.rep,v) for v,a in zip(vs,args) if isinstance(a,Variable))
	rn := make(map[string]lg.Expr)
	for i, v := range vs {
		if i < len(args) {
			if varArg, isVar := args[i].(*lg.Variable); isVar {
				rn[varArg.Name] = v
			}
		}
	}
	drhs := rhs
	if len(rn) > 0 {
		drhs = module.SubstituteAstByName(rhs, rn)
	}

	// Create nondeterministic skolem symbol
	// Python: nondet = n.suffix("_nd").skolem()
	skSym := lg.NewConst(sym.Name+"_nd", sym.CSort)
	var nondet lg.Expr
	if len(phNodes) > 0 {
		nondet = applyToNodes(skSym, phNodes)
	} else {
		nondet = skSym
	}

	// If eqs: nondet = Ite(And(*eqs), nondet, n(*dlhs.args))
	if len(eqs) > 0 {
		var guard lg.Expr
		if len(eqs) == 1 {
			guard = eqs[0]
		} else {
			guard = &lg.And{Terms: eqs}
		}
		origApply := applyToNodes(sym, phNodes) // n(*dlhs.args)
		ite, err := lg.NewIte(guard, nondet, origApply)
		if err == nil {
			nondet = ite
		}
	}

	// Build formulas
	lhsSort := lhs.NodeSort()
	rhsSort := rhs.NodeSort()

	var fmlas []lg.Expr

	// Iff(pto(lsort,rsort)(dlhs, Variable('X',rsort)), Equals(Variable('X',rsort), drhs))
	xVar, _ := lg.NewVariable("X", rhsSort)
	ptoSym := il.Pto(lhsSort, rhsSort)
	ptoApp := applyToNodes(ptoSym, append([]lg.Expr{dlhs}, xVar))
	eqXdrhs := &lg.Eq{T1: xVar, T2: drhs}
	iff, err := lg.NewIff(ptoApp, eqXdrhs)
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
			if !lg.SortEqual(s, rhsSort) {
				xv, _ := lg.NewVariable("X", s)
				ptoS := il.Pto(lhsSort, s)
				ptoSApp := applyToNodes(ptoS, append([]lg.Expr{dlhs}, xv))
				fmlas = append(fmlas, &lg.Not{Body: ptoSApp})
			}
		}
	}

	// Return as update with definition: Definition(dlhs, nondet)
	defn := il.NewDefinition(dlhs, nondet)
	defs := []*lg.Definition{defn}

	// Combine: formulas go in TR, definition goes in defs
	// Python: new_clauses = Clauses(fmlas, [Definition(dlhs, nondet)])
	update := &Update{
		Modified: []*lg.Const{sym},
		TR:       module.NewClauses(fmlas, defs, EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
	}
	return update
}

// --- HavocAction ---

// ActionUpdate computes the transition relation for havoc (nondeterministic assignment).
// The new value is unconstrained at the specified indices, but equal to the old
// value at all other indices.
func (a *HavocAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.HavocAction.action_update ENTER")
	defer xtracer.Trace("actions.HavocAction.action_update EXIT")
	lhs := a.Target
	sym := constSym(lhs)
	if sym == nil {
		return NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)
	dom := il.SortDomain(sym.CSort)

	// Create fresh variables for the domain
	vs := make([]*lg.Variable, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVariable(fmt.Sprintf("X%d", i), s)
		vs[i] = v
	}

	// Build equality conditions for non-variable args
	var eqs []lg.Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*lg.Variable); !isVar {
				eqs = append(eqs, &lg.Eq{T1: v, T2: args[i]})
			}
		}
	}

	var tr lg.Expr
	vsNodes := varsToNodes(vs)

	if il.IsBoolean(sym) || il.IsRelationalSort(sym.CSort) {
		// Relation: at non-havocked indices, old and new agree
		// For each eq in eqs: (new_n(Vs) <-> n(Vs)) | eq
		var terms []lg.Expr
		newApp := applyToNodes(newN, vsNodes)
		oldApp := applyToNodes(sym, vsNodes)
		for _, eq := range eqs {
			impl1, _ := lg.NewOr(&lg.Not{Body: newApp}, oldApp, eq)
			impl2, _ := lg.NewOr(newApp, &lg.Not{Body: oldApp}, eq)
			terms = append(terms, impl1, impl2)
		}
		if len(terms) == 0 {
			tr = lg.True // no constraints at all: fully nondeterministic
		} else {
			tr = conjoin(terms...)
		}
	} else if il.IsIndividual(sym) {
		// Function: at non-havocked indices, new = old
		newApp := applyToNodes(newN, vsNodes)
		oldApp := applyToNodes(sym, vsNodes)
		var terms []lg.Expr
		for _, eq := range eqs {
			clause, _ := lg.NewOr(&lg.Eq{T1: newApp, T2: oldApp}, eq)
			terms = append(terms, clause)
		}
		if len(terms) == 0 {
			tr = lg.True
		} else {
			tr = conjoin(terms...)
		}
	} else {
		tr = lg.True
	}

	return makeUpdate([]*lg.Const{sym}, tr, lg.False, EmptyAnnotation{})
}

func applyToNodes(fn lg.Expr, args []lg.Expr) lg.Expr {
	if len(args) == 0 {
		return fn
	}
	return lg.MustApply(fn, args...)
}

// --- SetAction ---

// ActionUpdate computes the transition relation for a set operation on a relation.
// Python: SetAction.action_update (ivy_actions.py:624-636).
// Builds clauses with frame conditions ensuring values at non-matching indices
// are preserved.
func (a *SetAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.SetAction.action_update ENTER")
	defer xtracer.Trace("actions.SetAction.action_update EXIT")
	if a.Lit == nil {
		return NullUpdate()
	}

	// Determine polarity and atom
	lit := a.Lit
	positive := true
	if n, ok := lit.(*lg.Not); ok {
		positive = false
		lit = n.Body
	}

	// Extract the relation symbol and args from the atom
	var relSym *lg.Const
	var args []lg.Expr
	if app, ok := lit.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Const); ok {
			relSym = c
			args = app.Terms
		}
	} else if c, ok := lit.(*lg.Const); ok {
		relSym = c
	}

	if relSym == nil {
		return NullUpdate()
	}

	newN := newSym(relSym)
	vs := module.SymPlaceholders(relSym)
	vsNodes := varsToNodes(vs)

	// Build equality conditions for non-variable args
	// Python: eqs = [Atom(equals,[v,a]) for (v,a) in zip(vs,args) if not isinstance(a,Variable)]
	var eqs []lg.Expr
	for i, v := range vs {
		if i < len(args) {
			if _, isVar := args[i].(*lg.Variable); !isVar {
				eqs = append(eqs, &lg.Eq{T1: v, T2: args[i]})
			}
		}
	}

	// Build the formula components
	// Python: new_clauses = And(*(
	//   [Or(sign(lit.polarity, Atom(new_n, vs)), sign(1-lit.polarity, Atom(n, vs))),
	//    sign(lit.polarity, Atom(new_n, args))] +
	//   [Or(*([sign(0, Atom(new_n, vs)), sign(1, Atom(n, vs))] + [eq])) for eq in eqs] +
	//   [Or(*([sign(1, Atom(new_n, vs)), sign(0, Atom(n, vs))] + [eq])) for eq in eqs]))

	newNvs := applyToNodes(newN, vsNodes) // new_n(Vs)
	nVs := applyToNodes(relSym, vsNodes)  // n(Vs)
	newNargs := applyToNodes(newN, args)  // new_n(args)

	var parts []lg.Expr

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
	return makeUpdate([]*lg.Const{relSym}, tr, lg.False, EmptyAnnotation{})
}

// --- NativeAction ---

// IntUpdate for NativeAction is a no-op — skips update axioms.
// Python: NativeAction.int_update (ivy_actions.py:1286) returns
// ([], true_clauses(), false_clauses()) — annot is None.
func (a *NativeAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
	return NullUpdate()
}

// --- DebugAction ---

// IntUpdate for DebugAction is a no-op — skips update axioms.
// Python: DebugAction.int_update (ivy_actions.py:1267) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
func (a *DebugAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
	return &Update{
		Modified: []*lg.Const{},
		TR:       module.TrueClauses(EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
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
		Modified: []*lg.Const{},
		TR:       module.TrueClauses(EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
	}
}

// --- Field Actions ---

// makeFieldUpdateFunc constructs a field-update AssignAction using a callable
// RHS builder and returns its ActionUpdate.
// Python: make_field_update(self, l, f, r_func, domain, pvars) with lambda r_func.
func makeFieldUpdateFunc(field, obj lg.Expr, rhsFunc func(v *lg.Variable) lg.Expr, ctx *UpdateContext) *Update {
	sym, ok := field.(*lg.Const)
	if sym == nil || !ok {
		return NullUpdate()
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok || !il.IsRelationalSort(sym.CSort) || len(fs.Domain()) != 2 {
		// "field must be a binary relation"
		return NullUpdate()
	}
	v, _ := lg.NewVariable("X", fs.Domain()[1])
	// Build f(l, v) as Apply
	lhs, _ := lg.NewApply(sym, obj, v)
	rhs := rhsFunc(v)
	aa := NewAssignAction(lhs, rhs)
	return aa.ActionUpdate(ctx)
}

// ActionUpdate for AssignFieldAction.
// Python: l,f,r = self.args; make_field_update(self,l,f,lambda v: Equals(v,r),domain,pvars)
func (a *AssignFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *lg.Variable) lg.Expr {
		return il.NewEqualsNode(v, a.Value)
	}, ctx)
}

// ActionUpdate for NullFieldAction.
// Python: l,f = self.args; make_field_update(self,l,f,lambda v: Or(),domain,pvars)
func (a *NullFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *lg.Variable) lg.Expr {
		return &lg.Or{} // Or() with no args = false
	}, ctx)
}

// ActionUpdate for CopyFieldAction.
// Python: l,lf,r,rf = self.args; make_field_update(self,l,lf,lambda v: rf(r,v),domain,pvars)
func (a *CopyFieldAction) ActionUpdate(ctx *UpdateContext) *Update {
	srcField := a.SrcField
	if srcField == nil {
		srcField = a.Field // backward compat: same field for both
	}
	return makeFieldUpdateFunc(a.Field, a.Dst, func(v *lg.Variable) lg.Expr {
		if sym, ok := srcField.(*lg.Const); ok {
			app, _ := lg.NewApply(sym, a.Src, v)
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
	case *Sequence:
		return "Sequence"
	case *ChoiceAction:
		return "ChoiceAction"
	case *EnvAction:
		return "EnvAction"
	case *IfAction:
		return "IfAction"
	case *WhileAction:
		return "WhileAction"
	case *LocalAction:
		return "LocalAction"
	case *LetAction:
		return "LetAction"
	case *CallAction:
		return "CallAction"
	case *BindOldsAction:
		return "BindOldsAction"
	case *AssignAction:
		return "AssignAction"
	case *SetAction:
		return "SetAction"
	case *HavocAction:
		return "HavocAction"
	case *AssumeAction:
		return "AssumeAction"
	case *AssertAction:
		return "AssertAction"
	case *CrashAction:
		return "CrashAction"
	case *InstantiateAction:
		return "InstantiateAction"
	case *NativeAction:
		return "NativeAction"
	case *DebugAction:
		return "DebugAction"
	case *RequiresAction:
		return "RequiresAction"
	case *EnsuresAction:
		return "EnsuresAction"
	case *ThunkAction:
		return "ThunkAction"
	case *ReturnAction:
		return "ReturnAction"
	case *IgnoreAction:
		return "IgnoreAction"
	case *SubgoalAction:
		return "SubgoalAction"
	case *VarAction:
		return "VarAction"
	case *AssignFieldAction:
		return "AssignFieldAction"
	case *NullFieldAction:
		return "NullFieldAction"
	case *CopyFieldAction:
		return "CopyFieldAction"
	default:
		return fmt.Sprintf("%T", a)
	}
}

// IntUpdate is the intermediate update computation that applies domain
// update axioms on top of the atomic action_update.
// Corresponds to Python Action.int_update().
func IntUpdate(action Action, ctx *UpdateContext) *Update {
	// Dispatch to type-specific int_update methods.
	// Types with their own IntUpdate method have their own traces.
	// Types using intUpdateFromActionUpdate use the base Action.int_update
	// trace (matching Python's Action.int_update which traces
	// "actions.IntUpdate ENTER type=%s").
	switch a := action.(type) {
	case *AssumeAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *AssertAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *RequiresAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *EnsuresAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *AssignAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *HavocAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *SetAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *NativeAction:
		return a.IntUpdate(ctx)
	case *DebugAction:
		return a.IntUpdate(ctx)
	case *ReturnAction:
		return a.IntUpdate(ctx)
	case *AssignFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *NullFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *CopyFieldAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
	case *Sequence:
		return a.IntUpdate(ctx)
	case *ChoiceAction:
		return a.IntUpdate(ctx)
	case *EnvAction:
		return a.IntUpdateEnv(ctx)
	case *IfAction:
		return a.IntUpdate(ctx)
	case *WhileAction:
		return a.IntUpdate(ctx)
	case *LocalAction:
		return a.IntUpdate(ctx)
	case *LetAction:
		return a.IntUpdate(ctx)
	case *CallAction:
		return a.IntUpdate(ctx)
	case *BindOldsAction:
		return a.IntUpdate(ctx)
	case *CrashAction:
		xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
		return intUpdateFromActionUpdate(a, ctx)
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
	update = applyUpdateAxioms(update, action.(Action), ctx)
	return update
}

// updateAxiomProvider matches the Updater interface in action.go.
// Implemented by PatternBasedUpdate, DerivedUpdate, and NamedUpdate.
type updateAxiomProvider = Updater

// applyUpdateAxioms applies domain.updates to the given update.
// In Python, this iterates over domain.updates calling get_update_axioms.
func applyUpdateAxioms(update *Update, action Action, ctx *UpdateContext) *Update {
	if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
		return update
	}

	modified := update.Modified
	tr := update.TR
	pre := update.Pre

	if xtracer.Enabled {
		show := make([]string, len(modified))
		for i, s := range modified {
			show[i] = fmt.Sprintf("'%v'", s.Name)
		}
		sort.Strings(show)
		xtracer.Trace("actions.applyUpdateAxioms ENTER numUpdates=%d modNames=[%v]", len(ctx.Domain.Updates), strings.Join(show, ", "))
	}
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
				tr = module.AndClausesTyped(tr, transrelNode)
			}
			if precondNode != nil {
				pre = module.OrClausesTyped(pre, precondNode)
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
func (s *Sequence) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.Sequence.int_update ENTER")
	defer xtracer.Trace("actions.Sequence.int_update EXIT")
	// Python (ivy_actions.py:841):
	//   update = ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
	result := &Update{
		Modified: []*lg.Const{},
		TR:       module.TrueClauses(EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
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
func unwrapToAction(n lg.Expr) Action {
	act, _ := n.(Action)
	return act
}

// sortedModNames formats a slice of modified-symbol Consts as a sorted slice
// of "'name'"-quoted strings, matching the Python xtracer trace format
// produced by sorted([s.name for s in modified]).
func sortedModNames(mods []*lg.Const) []string {
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
func (a *ChoiceAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.ChoiceAction.int_update ENTER")
	defer xtracer.Trace("actions.ChoiceAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = IfAction(Not(cond), self.args[0], self.args[1])
	//   return ite.int_update(domain, pvars)
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := module.BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(&lg.Not{Body: cond}, a.Branches[0], a.Branches[1])
		return ite.IntUpdate(ctx)
	}
	// Python (ivy_actions.py:889):
	//   result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
	result := &Update{
		Modified: []*lg.Const{},
		TR:       module.FalseClauses(EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
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
func (a *EnvAction) IntUpdateEnv(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.EnvAction.int_update ENTER")
	defer xtracer.Trace("actions.EnvAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = IfAction(cond, self.args[0], self.args[1])
	//   return ite.update(domain, pvars)
	// Note: EnvAction uses cond (positive), ChoiceAction uses Not(cond).
	// Note: EnvAction calls update (GetUpdate), not int_update (IntUpdate).
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := module.BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(cond, a.Branches[0], a.Branches[1])
		return GetUpdate(ite, ctx)
	}
	// Python (ivy_actions.py:917):
	//   result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
	result := &Update{
		Modified: []*lg.Const{},
		TR:       module.FalseClauses(EmptyAnnotation{}),
		Pre:      module.FalseClauses(EmptyAnnotation{}),
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
func (a *IfAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.IfAction.int_update ENTER")
	defer xtracer.Trace("actions.IfAction.int_update EXIT")
	cond := a.Cond

	// Python: if used_variables_ast(self.args[0]): raise IvyError(...)
	freeVars := module.UsedVariablesAST(cond)
	if len(freeVars) > 0 {
		panic("variables in \"if\" conditions must be explicitly quantified")
	}

	// Python: if not isinstance(self.args[0], ivy_ast.Some): ... else: subactions path
	if _, isSome := cond.(*SomeCondition); isSome {
		return a.intUpdateWithSubactions(ctx)
	}

	// Python (ivy_actions.py:990): if not is_boolean(self.args[0]): raise IvyError("condition must be boolean")
	if !il.IsBoolean(cond) {
		panic("condition must be boolean")
	}

	// Simple boolean condition
	thenBranch := a.ThenBody
	var elseBranch lg.Expr
	if a.ElseBody != nil {
		elseBranch = a.ElseBody
	}

	thenAct := unwrapToAction(thenBranch)
	if thenAct == nil {
		thenAct = NewSequence()
	}
	var elseAct Action
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
func (a *IfAction) intUpdateWithSubactions(ctx *UpdateContext) *Update {
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
				Cond:  &lg.Not{Body: ite.Cond},
				ThenB: ite.ElseB,
				ElseB: ite.ThenB,
			}
		}
	}
	if res.Pre != nil {
		if ite, ok := res.Pre.Annot.(*IteAnnotation); ok {
			res.Pre.Annot = &IteAnnotation{
				Cond:  &lg.Not{Body: ite.Cond},
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
func (a *WhileAction) IntUpdate(ctx *UpdateContext) *Update {
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
func (a *WhileAction) Unroll(card func(lg.Sort) int, body Action) (Action, error) {
	cond := a.Cond
	// Unwrap nested And to find comparison
	for {
		if andN, ok := cond.(*lg.And); ok && len(andN.Terms) > 0 {
			cond = andN.Terms[0]
		} else {
			break
		}
	}
	// Determine index sort from condition
	var idxSort lg.Sort
	if app, ok := cond.(*lg.Apply); ok {
		if sym, ok := app.Func.(*lg.Const); ok {
			if sym.Name == "<" || sym.Name == ">" || sym.Name == "<=" || sym.Name == ">=" {
				if len(app.Terms) > 0 {
					idxSort = app.Terms[0].NodeSort()
				}
			}
		}
	} else if notN, ok := cond.(*lg.Not); ok {
		if eq, ok := notN.Body.(*lg.Eq); ok {
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
	// Python: res = IfAction(self.args[0], AssumeAction(Or()))
	var bodyExpr lg.Expr
	if body != nil {
		bodyExpr = body
	} else {
		bodyExpr = a.Body
	}

	// Innermost: if cond then assume false (empty Or = false)
	res := NewIfAction(a.Cond, NewAssumeAction(&lg.Or{}))
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
func (a *WhileAction) Expand(ctx *UpdateContext) Action {
	// First, compute the modify set from the body
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		return NewSequence()
	}
	bodyUpdate := IntUpdate(bodyAct, ctx)
	modset := bodyUpdate.Modified

	// Separate invariants from ranking
	var invariants []lg.Expr
	var ranking Action
	for _, inv := range a.Invariants {
		if r, ok := inv.(Action); ok {
			if r.Name() == "decreases" {
				ranking = r
				continue
			}
		}
		invariants = append(invariants, inv)
	}

	// Build assert invariants
	var asserts []Action
	for _, inv := range invariants {
		asserts = append(asserts, NewAssertAction(inv))
	}

	// Build assume invariants (assert→assume conversion)
	var assumes []Action
	for _, inv := range invariants {
		assumes = append(assumes, NewAssumeAction(inv))
	}

	// Build havocs for modified symbols
	var havocs []Action
	for _, sym := range modset {
		havocs = append(havocs, NewHavocAction(sym))
	}

	// Handle ranking function if present
	var entryAsserts, exitAsserts []Action
	var rankLocal *lg.Const
	if ranking != nil {
		rankArgs := ranking.ActionArgs()
		if len(rankArgs) > 0 {
			rankExpr := rankArgs[0]
			rankSort := rankExpr.NodeSort()
			aux := lg.NewConst("$rank", rankSort)
			rankLocal = aux
			assumes = append(assumes, NewAssumeAction(&lg.Eq{T1: aux, T2: rankExpr}))
			ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{rankSort, rankSort}))
			exitAsserts = append(exitAsserts, NewAssertAction(lg.MustApply(ltSym, rankExpr, aux)))
			zeroSym := lg.NewConst("0", rankSort)
			entryAsserts = append(entryAsserts, NewAssertAction(&lg.Not{Body: lg.MustApply(ltSym, rankExpr, zeroSym)}))
		}
	}

	// Build the expanded action:
	// asserts; havocs; assumes; if cond then (entry_asserts; body; exit_asserts; asserts; assume false)
	var bodyParts []lg.Expr
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
	bodyParts = append(bodyParts, NewAssumeAction(lg.False))
	thenBody := NewSequence(bodyParts...)

	ifAction := NewIfAction(a.Cond, thenBody, NewSequence())

	// Assemble the full expanded sequence
	var allParts []lg.Expr
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
		return NewLocalActionOn(ctx.ActCfg, "actions.action_on_subgoal", rankLocal, result)
	}
	return result
}

// --- LocalAction ---

// IntUpdate computes the local action's update by hiding local symbols.
// Python: LocalAction.int_update computes body.int_update then hide(syms, update).
func (a *LocalAction) IntUpdate(ctx *UpdateContext) *Update {
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
	symsToHide := make([]*lg.Const, len(a.Locals))
	for i, local := range a.Locals {
		c, ok := local.(*lg.Const)
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
func (a *LetAction) IntUpdate(ctx *UpdateContext) *Update {
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
func (a *BindOldsAction) IntUpdate(ctx *UpdateContext) *Update {
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
func (a *CallAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.CallAction.int_update ENTER")
	defer xtracer.Trace("actions.CallAction.int_update EXIT")
	calleeName := constName(a.Callee)
	if calleeName == "" {
		// Python (ivy_actions.py:1318): name = self.args[0].rep — would
		// AttributeError if .rep is missing. Faithful port panics.
		panic(fmt.Sprintf("CallAction.IntUpdate: callee has no name: %T", a.Callee))
	}

	// Resolve the callee
	var calleeAction Action
	if ctx.GetAction != nil {
		calleeAction = ctx.GetAction(calleeName)
	}
	if calleeAction == nil {
		// Try from domain.Actions
		if ctx.Domain != nil && ctx.Domain.Actions != nil {
			if v, ok := ctx.Domain.Actions.Get2(calleeName); ok {
				act, isAct := v.(Action)
				if !isAct {
					// Python (ivy_actions.py:1350): v = state_to_action(v.value)
					//   for non-Action context entries (state-style updates with .value).
					// Go's module.Module.Actions is typed *iu.InsMap[string, Action]
					// (module.go:38), so the type system prevents non-Action values
					// from ever being stored. The state_to_action(v.value) branch
					// is unreachable in Go's typed model. If this panic ever fires,
					// the upstream construction site is violating the type contract
					// — investigate that, do not port state_to_action(v.value).
					panic(fmt.Sprintf("CallAction.IntUpdate: callee %s resolved to non-Action %T (Domain.Actions is typed Action; non-Action storage is a programming error)", calleeName, v))
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
	return a.applyActuals(ctx, calleeAction)
}

// applyActuals inlines the callee with actual parameters.
// Corresponds to Python CallAction.apply_actuals.
// Includes capture avoidance via distinct_obj_renaming.
func (a *CallAction) applyActuals(ctx *UpdateContext, callee Action) *Update {
	formalParams := callee.GetFormalParams()
	formalReturns := callee.GetFormalReturns()
	actualParams := nodeArgs(a.Callee)
	actualReturns := a.ActualReturns

	// Validate parameter counts.
	// Python (ivy_actions.py:1352): raise IvyError("wrong number of input parameters")
	if len(formalParams) != len(actualParams) {
		panic("wrong number of input parameters")
	}
	// Python (ivy_actions.py:1356): raise IvyError("wrong number of output parameters")
	if len(formalReturns) != len(actualReturns) {
		panic("wrong number of output parameters")
	}

	// Capture avoidance: rename formals to avoid colliding with actuals.
	// Python: vocab = list(symbols_asts(actual_params+actual_returns))
	//         subst = distinct_obj_renaming(formal_params+formal_returns, vocab)
	allFormals := make([]*lg.Const, 0, len(formalParams)+len(formalReturns))
	allFormals = append(allFormals, formalParams...)
	allFormals = append(allFormals, formalReturns...)

	// Collect names used in actual parameters and returns
	vocabNames := make(map[string]bool)
	for _, ap := range actualParams {
		collectSymbolNames(ap, vocabNames)
	}
	for _, ar := range actualReturns {
		collectSymbolNames(ar, vocabNames)
	}

	// Build renaming: formal → fresh name
	renaming := distinctObjRenaming(allFormals, vocabNames)

	// Apply renaming to callee — always call SubstituteConstantsAction
	// (matches Python which unconditionally calls substitute_constants_ast)
	substMap := make(map[lg.NodeKey]lg.Expr)
	for oldSym, newSym := range renaming {
		substMap[lg.Key(oldSym)] = newSym
		// Python (ivy_actions.py:1364-1365): subst[old(s)] = old(t)
		//   where old(sym) = sym.prefix('old_'). Use the Old() helper here
		//   so the substMap key actually matches a real pre-state symbol in
		//   the callee body (e.g., from a prior BindOldsAction wrapper).
		oldOfOldSym := lg.NewConst(Old(oldSym.Name), oldSym.CSort)
		oldOfNewSym := lg.NewConst(Old(newSym.Name), newSym.CSort)
		substMap[lg.Key(oldOfOldSym)] = oldOfNewSym
	}
	renamedCallee := SubstituteConstantsAction(callee, substMap)

	// Get renamed formals
	renamedFormalParams := make([]*lg.Const, len(formalParams))
	for i, fp := range formalParams {
		if newSym, ok := renaming[fp]; ok {
			renamedFormalParams[i] = newSym
		} else {
			renamedFormalParams[i] = fp
		}
	}
	renamedFormalReturns := make([]*lg.Const, len(formalReturns))
	for i, fr := range formalReturns {
		if newSym, ok := renaming[fr]; ok {
			renamedFormalReturns[i] = newSym
		} else {
			renamedFormalReturns[i] = fr
		}
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
	var inputAsgns []lg.Expr
	for i, fp := range renamedFormalParams {
		if i < len(actualParams) {
			asgn := NewAssignAction(fp, actualParams[i])
			inputAsgns = append(inputAsgns, asgn)
		}
	}

	// Build output assignments: actual_return := formal_return
	var outputAsgns []lg.Expr
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
	toHide := make([]*lg.Const, 0, len(renamedFormalParams)+len(renamedFormalReturns))
	toHide = append(toHide, renamedFormalParams...)
	toHide = append(toHide, renamedFormalReturns...)
	return Hide(toHide, update)
}

// distinctObjRenaming creates a renaming from formals to fresh names
// that don't conflict with vocabNames.
// Corresponds to Python distinct_obj_renaming.
func distinctObjRenaming(formals []*lg.Const, vocabNames map[string]bool) map[*lg.Const]*lg.Const {
	result := make(map[*lg.Const]*lg.Const)
	usedNames := make(map[string]bool)
	for k := range vocabNames {
		usedNames[k] = true
	}

	for _, sym := range formals {
		name := sym.Name
		if _, used := usedNames[name]; !used {
			usedNames[name] = true
			// No conflict — identity mapping (matches Python which always adds all formals)
			result[sym] = lg.NewConst(name, sym.CSort)
			continue
		}
		// Need a fresh name
		newName := unusedNameWithBase(name, usedNames)
		usedNames[newName] = true
		result[sym] = lg.NewConst(newName, sym.CSort)
	}
	return result
}

// unusedNameWithBase finds an unused name starting with base, using
// letter suffixes (a, b, ..., z, a0, ...) matching Python's
// constant_name_generator via ivy_utils.unused_name_with_base.
func unusedNameWithBase(base string, used map[string]bool) string {
	gen := iu.ConstantNameGenerator()
	for {
		name := base + "_" + gen()
		if !used[name] {
			return name
		}
	}
}

// collectSymbolNames collects all symbol/constant names from a logic node.
func collectSymbolNames(node lg.Expr, names map[string]bool) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *lg.Const:
		names[n.Name] = true
	case *lg.Apply:
		if n.Func != nil {
			collectSymbolNames(n.Func, names)
		}
		for _, t := range n.Terms {
			collectSymbolNames(t, names)
		}
		return
	}
	for _, c := range node.Children() {
		collectSymbolNames(c, names)
	}
}

// --- CrashAction ---

// ActionUpdate computes the crash action by havocing all non-spec mutable symbols.
// Python: CrashAction.action_update — wants update axioms applied via intUpdateFromActionUpdate.
func (a *CrashAction) ActionUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.CrashAction.action_update ENTER")
	defer xtracer.Trace("actions.CrashAction.action_update EXIT")
	target := a.Target
	targetName := constName(target)
	if targetName == "" || ctx.Domain == nil {
		return NullUpdate()
	}

	// Collect symbols to havoc
	var symsToHavoc []*lg.Const
	collectCrashSyms(ctx.Domain, targetName, &symsToHavoc)

	if len(symsToHavoc) == 0 {
		return NullUpdate()
	}

	// Build havoc actions for each symbol
	var havocParts []lg.Expr
	for _, sym := range symsToHavoc {
		havocParts = append(havocParts, NewHavocAction(sym))
	}
	seq := NewSequence(havocParts...)
	return IntUpdate(seq, ctx)
}

// collectCrashSyms recursively collects symbols to havoc for a crash action.
func collectCrashSyms(domain *module.Module, name string, result *[]*lg.Const) {
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
	sym := lg.NewConst(name, lg.TopS)
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
func GetUpdate(action Action, ctx *UpdateContext) *Update {
	xtracer.Trace("actions.GetUpdate ENTER type=%s", ActionTypeName(action))
	update := IntUpdate(action, ctx)
	update = BindOldsUpdate(update)
	update = hideFormals(action, update)
	return update
}

// hideFormals hides formal parameters and returns from the update.
// Matches Python Action.hide_formals (ivy_actions.py:220-228).
func hideFormals(action Action, update *Update) *Update {
	var toHide []*lg.Const
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
func GetUpdateForArt(action Action, domain *module.Module, inScope map[string]bool) *Update {
	ctx := &UpdateContext{
		Domain:       domain,
		PVars:        inScope,
		ActCfg:       domain.Cfg.ActCfg,
		Instantiator: domain.Instantiator,
		GetAction: func(name string) Action {
			if domain != nil && domain.Actions != nil {
				if v, ok := domain.Actions.Get2(name); ok {
					if act, ok := v.(Action); ok {
						return act
					}
				}
			}
			return nil
		},
	}
	return GetUpdate(action, ctx)
}
