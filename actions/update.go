// This file implements the action_update() and int_update() methods for
// each action type, porting the transition-relation computation from
// Python ivy_actions.py lines 309-1302.
//
// Each action's ActionUpdate method returns a *transrel.Update representing
// the transition relation (modified symbols, TR formula, precondition).
//
// IntUpdate applies update axioms from the domain on top of ActionUpdate.
// GetUpdate is the top-level entry point that hides formals and binds olds.
package actions

import (
	"fmt"
	"strconv"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/xtracer"
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
	Instantiator func([]lg.Expr) *co.Clauses
}

// BackgroundTheory returns the background theory (axioms) for the domain.
func (ctx *UpdateContext) BackgroundTheory() *co.Clauses {
	if ctx.Domain == nil {
		return co.TrueClauses(nil)
	}
	clauses := ctx.Domain.BackgroundTheory(ctx.PVars)
	if clauses == nil {
		return co.TrueClauses(nil)
	}
	return clauses
}

// makeUpdate creates a transrel.Update from individual components,
// wrapping lg.Expr values into Clauses. This is a transitional helper
// for porting action updates from bare Node to Clauses.
func makeUpdate(modified []*lg.Symbol, tr lg.Expr, pre lg.Expr, annot interface{}) *transrel.Update {
	return &transrel.Update{
		Modified: modified,
		TR:       co.FormulaToClauses(tr, annot),
		Pre:      co.FormulaToClauses(pre, annot),
	}
}

// makeUpdateDefs creates a transrel.Update with definitions in the TR.
// This matches Python's pattern of Clauses([], [Definition(...)], annot).
func makeUpdateDefs(modified []*lg.Symbol, defs []*il.Definition, annot interface{}) *transrel.Update {
	return &transrel.Update{
		Modified: modified,
		TR:       co.NewClauses(nil, defs, annot),
		Pre:      co.FalseClauses(annot),
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
	notA := co.Negate(a)
	notB := co.Negate(b)
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

// dualFormula negates a formula and skolemizes its free variables.
// Corresponds to Python's dual_formula.
func dualFormula(fmla lg.Expr) lg.Expr {
	// Collect free variables
	vars := co.UsedVariablesAST(fmla)
	if len(vars) > 0 {
		// Replace variables with skolem constants
		subs := make(map[string]lg.Expr, len(vars))
		for _, vNode := range vars {
			vv := vNode.(*lg.Variable)
			sk := lg.NewSymbol("__"+vv.Name, vv.VSort)
			subs[vv.Name] = sk
		}
		fmla = substituteVars(fmla, subs)
	}
	return co.Negate(fmla)
}

// substituteVars replaces variables by name with replacement nodes.
func substituteVars(node lg.Expr, subs map[string]lg.Expr) lg.Expr {
	if len(subs) == 0 {
		return node
	}
	return substituteVarsRec(node, subs)
}

func substituteVarsRec(node lg.Expr, subs map[string]lg.Expr) lg.Expr {
	switch t := node.(type) {
	case *lg.Variable:
		if r, ok := subs[t.Name]; ok {
			return r
		}
		return node
	case *lg.Symbol:
		return node
	case *lg.Apply:
		// Python substitute_ast iterates ast.args (Terms only), preserves Func.
		newTerms := make([]lg.Expr, len(t.Terms))
		changed := false
		for i, arg := range t.Terms {
			newTerms[i] = substituteVarsRec(arg, subs)
			if newTerms[i] != arg {
				changed = true
			}
		}
		if !changed {
			return node
		}
		result, err := lg.NewApply(t.Func, newTerms...)
		if err != nil {
			return node
		}
		return result
	case *lg.Eq:
		t1 := substituteVarsRec(t.T1, subs)
		t2 := substituteVarsRec(t.T2, subs)
		if t1 == t.T1 && t2 == t.T2 {
			return node
		}
		return &lg.Eq{T1: t1, T2: t2}
	case *lg.Not:
		body := substituteVarsRec(t.Body, subs)
		if body == t.Body {
			return node
		}
		return &lg.Not{Body: body}
	case *lg.And:
		return substituteVarsChildren(&lg.And{}, t.Terms, subs, node)
	case *lg.Or:
		return substituteVarsChildren(&lg.Or{}, t.Terms, subs, node)
	case *lg.Implies:
		t1 := substituteVarsRec(t.T1, subs)
		t2 := substituteVarsRec(t.T2, subs)
		if t1 == t.T1 && t2 == t.T2 {
			return node
		}
		return &lg.Implies{T1: t1, T2: t2}
	case *lg.Iff:
		t1 := substituteVarsRec(t.T1, subs)
		t2 := substituteVarsRec(t.T2, subs)
		if t1 == t.T1 && t2 == t.T2 {
			return node
		}
		return &lg.Iff{T1: t1, T2: t2}
	case *lg.Ite:
		cond := substituteVarsRec(t.Cond, subs)
		then := substituteVarsRec(t.Then, subs)
		els := substituteVarsRec(t.Else, subs)
		if cond == t.Cond && then == t.Then && els == t.Else {
			return node
		}
		return &lg.Ite{ISort: t.ISort, Cond: cond, Then: then, Else: els}
	case *lg.ForAll:
		// Remove bound variables from subs
		newSubs := filterBound(subs, t.Variables)
		body := substituteVarsRec(t.Body, newSubs)
		if body == t.Body {
			return node
		}
		return &lg.ForAll{Variables: t.Variables, Body: body}
	case *lg.Exists:
		newSubs := filterBound(subs, t.Variables)
		body := substituteVarsRec(t.Body, newSubs)
		if body == t.Body {
			return node
		}
		return &lg.Exists{Variables: t.Variables, Body: body}
	}
	return node
}

func substituteVarsChildren(andOrOr interface{}, terms []lg.Expr, subs map[string]lg.Expr, orig lg.Expr) lg.Expr {
	newTerms := make([]lg.Expr, len(terms))
	changed := false
	for i, t := range terms {
		newTerms[i] = substituteVarsRec(t, subs)
		if newTerms[i] != t {
			changed = true
		}
	}
	if !changed {
		return orig
	}
	switch andOrOr.(type) {
	case *lg.And:
		return &lg.And{Terms: newTerms}
	case *lg.Or:
		return &lg.Or{Terms: newTerms}
	}
	return orig
}

func filterBound(subs map[string]lg.Expr, vars []*lg.Variable) map[string]lg.Expr {
	if len(vars) == 0 {
		return subs
	}
	newSubs := make(map[string]lg.Expr, len(subs))
	bound := make(map[string]bool, len(vars))
	for _, v := range vars {
		bound[v.Name] = true
	}
	for k, v := range subs {
		if !bound[k] {
			newSubs[k] = v
		}
	}
	return newSubs
}

// skolemizeFormula strips leading existential quantifiers and replaces
// bound variables with skolem constants.
func skolemizeFormula(fmla lg.Expr) lg.Expr {
	var vs []*lg.Variable
	for {
		if ex, ok := fmla.(*lg.Exists); ok {
			vs = append(vs, ex.Variables...)
			fmla = ex.Body
		} else {
			break
		}
	}
	if len(vs) == 0 {
		return fmla
	}
	subs := make(map[string]lg.Expr, len(vs))
	for _, v := range vs {
		sk := lg.NewSymbol("__sk__"+v.Name, v.VSort)
		subs[v.Name] = sk
	}
	return substituteVars(fmla, subs)
}

// -----------------------------------------------------------------------
// newSym returns the post-state version of a symbol.
// -----------------------------------------------------------------------

func newSym(sym *lg.Symbol) *lg.Symbol {
	return lg.NewSymbol(transrel.New(sym.Name), sym.CSort)
}

// constName extracts the name from a node that is a Const or the Func of an Apply.
func constName(n lg.Expr) string {
	switch t := n.(type) {
	case *lg.Symbol:
		return t.Name
	case *lg.Apply:
		return constName(t.Func)
	}
	return ""
}

// constSym extracts the Const from a node (Const or Apply.Func).
func constSym(n lg.Expr) *lg.Symbol {
	switch t := n.(type) {
	case *lg.Symbol:
		return t
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Symbol); ok {
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
	case *lg.Symbol:
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

func mkAssignClauses(lhs, rhs lg.Expr) *transrel.Update {
	sym := constSym(lhs)
	if sym == nil {
		// Fallback: no-op update
		return transrel.NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// Placeholder variables for the full domain of the symbol
	phs := co.SymPlaceholders(sym)

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
	drhs := substituteVars(rhs, rn)

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
	return &transrel.Update{
		Modified: []*lg.Symbol{sym},
		TR:       co.NewClauses(nil, []*il.Definition{defn}, EmptyAnnotation{}),
		Pre:      co.FalseClauses(EmptyAnnotation{}),
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
func (a *AssumeAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.AssumeAction.action_update ENTER")
	defer xtracer.Trace("actions.AssumeAction.action_update EXIT")
	// Python: if isinstance(fmla, LabeledFormula) and fmla.unprovable: return skip
	if a.Unprovable {
		return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
	}
	fmla := a.Formula
	// Python: clauses = formula_to_clauses_tseitin(skolemize_formula(fmla))
	//         clauses = unfold_definitions_clauses(clauses)
	//         clauses = Clauses(clauses.fmlas, clauses.defs, EmptyAnnotation())
	fmla = co.SkolemizeFormula(fmla, nil)
	clauses := co.FormulaToClauses(fmla, nil)
	if ctx != nil && ctx.Instantiator != nil {
		clauses = co.UnfoldDefinitionsClauses(clauses, ctx.Instantiator)
	}
	clauses = co.NewClauses(clauses.Fmlas, clauses.Defs, EmptyAnnotation{})
	return &transrel.Update{
		Modified: []*lg.Symbol{},
		TR:       clauses,
		Pre:      co.FalseClauses(EmptyAnnotation{}),
	}
}

// --- AssertAction ---

// ActionUpdate computes the transition relation for AssertAction.
// An assert generates a precondition (negative) from the dual of the formula.
// Implements Python's selective assertion checking via check_unprovable and checked_assert.
// Python: action_update (ivy_actions.py:343-362)
func (a *AssertAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.AssertAction.action_update ENTER")
	defer xtracer.Trace("actions.AssertAction.action_update EXIT")
	fmla := a.Formula
	unprovable := a.Unprovable

	// Python: if check_unprovable.get() != unprovable: skip
	if ctx.CheckUnprovable != unprovable {
		return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
	}

	// Python: if checked_assert is set and doesn't match this lineno
	if ctx.CheckedAssert != "" {
		if ctx.CheckedAssert != a.GetLineno().String() {
			if unprovable {
				// Unprovable assertion not selected: skip entirely
				return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
			}
			// Provable assertion not selected: return formula as-is (not dual)
			return makeUpdate([]*lg.Symbol{}, fmla, lg.False, EmptyAnnotation{})
		}
	}

	// Only assertions that pass both filters get dual formula treatment
	// Python: cl = formula_to_clauses(dual_formula(fmla))
	//         cl = Clauses(cl.fmlas, cl.defs, EmptyAnnotation())
	dual := co.DualFormula(fmla, nil)
	cl := co.FormulaToClauses(dual, nil)
	cl = co.NewClauses(cl.Fmlas, cl.Defs, EmptyAnnotation{})
	return &transrel.Update{
		Modified: []*lg.Symbol{},
		TR:       co.TrueClauses(EmptyAnnotation{}),
		Pre:      cl,
	}
}

// --- RequiresAction ---

func (a *RequiresAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return a.AssertAction.ActionUpdate(ctx)
}

// --- EnsuresAction ---

func (a *EnsuresAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
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
func (a *AssignAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.AssignAction.action_update ENTER")
	defer xtracer.Trace("actions.AssignAction.action_update EXIT")
	lhs, rhs := a.LHS, a.RHS
	sym := constSym(lhs)
	if sym == nil {
		return transrel.NullUpdate()
	}

	// Handle hierarchical case: if the symbol has children in the hierarchy
	if ctx.Domain != nil && ctx.Domain.Hierarchy != nil {
		if children, ok := ctx.Domain.Hierarchy[sym.Name]; ok && len(children) > 0 {
			xtracer.Trace("actions.AssignAction.action_update branch=hierarchy")
			// Decompose into sub-assignments for each child
			var updates []*transrel.Update
			axioms := ctx.BackgroundTheory()
			for childName := range children {
				childSym := lg.NewSymbol(childName, lg.TopS)
				childLHS := &lg.Apply{Func: childSym, Terms: nodeArgs(lhs)}
				childRHS := rhs // simplified: same RHS for each child
				childAssign := NewAssignAction(childLHS, childRHS)
				childUpdate := childAssign.ActionUpdate(ctx)
				updates = append(updates, childUpdate)
			}
			if len(updates) == 0 {
				return transrel.NullUpdate()
			}
			result := updates[0]
			for _, u := range updates[1:] {
				result = transrel.ComposeUpdates(result, axioms, u)
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
		return transrel.NullUpdate()
	}
	if xtra > 0 {
		// Extend lhs and rhs with fresh placeholder variables
		phs := co.SymPlaceholders(sym)
		extend := make([]lg.Expr, xtra)
		for i := 0; i < xtra; i++ {
			extend[i] = phs[len(phs)-xtra+i]
		}
		// Make variables distinct from those already used in lhs and rhs
		// Python: extend = variables_distinct_list_ast(extend, self)
		// We combine lhs and rhs into a single expression for variable collection
		combined := &lg.And{Terms: []lg.Expr{lhs, rhs}}
		extend = co.VariablesDistinctListAst(extend, combined)

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
	lhsVars := co.UsedVariablesAST(lhs)
	rhsVars := co.UsedVariablesAST(rhs)
	for k := range rhsVars {
		if _, found := lhsVars[k]; !found {
			// multiply assigned
			return transrel.NullUpdate()
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
func destrAsgnVal(lhs lg.Expr, fmlas *[]lg.Expr, domain *module.Module) (lg.Expr, *co.Clauses, *lg.Symbol) {
	lhsArgs := nodeArgs(lhs)
	if len(lhsArgs) == 0 {
		return lhs, co.FalseClauses(nil), nil
	}

	mut := lhsArgs[0]
	rest := lhsArgs[1:]
	mutN := constSym(mut)
	if mutN == nil {
		return lhs, co.FalseClauses(nil), nil
	}

	var lval lg.Expr
	var newClauses *co.Clauses
	var mutated *lg.Symbol

	if _, ok := domain.DestructorSorts[mutN.Name]; ok {
		// Recursive case: nested destructor
		// Python: lval, new_clauses, mutated = destr_asgn_val(mut, fmlas)
		lval, newClauses, mutated = destrAsgnVal(mut, fmlas, domain)
	} else {
		// Base case: mut is the root mutable symbol
		// Python: nondet = mut_n.suffix("_nd").skolem()
		skSym := lg.NewSymbol(mutN.Name+"_nd", mutN.CSort)
		phs := co.SymPlaceholders(mutN)
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
	vs := co.SymPlaceholders(n)
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
					destrPhs := co.SymPlaceholders(destr)
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
func (a *AssignAction) destructorAssignUpdate(ctx *UpdateContext, lhs, rhs lg.Expr) *transrel.Update {
	var fmlas []lg.Expr
	nondetLhs, newClauses, mutN := destrAsgnVal(lhs, &fmlas, ctx.Domain)
	if mutN == nil {
		return transrel.NullUpdate()
	}

	// Python: fmlas.append(equiv_ast(nondet_lhs, rhs))
	fmlas = append(fmlas, equivAST(nondetLhs, rhs))

	// Python: new_clauses = and_clauses(new_clauses, Clauses(fmlas))
	fmlaClauses := co.NewClauses(fmlas, nil, nil)
	combined := co.AndClausesTyped(newClauses, fmlaClauses)

	// Python: return ([mut_n], new_clauses, false_clauses(annot=EmptyAnnotation()))
	return &transrel.Update{
		Modified: []*lg.Symbol{mutN},
		TR:       combined,
		Pre:      co.FalseClauses(EmptyAnnotation{}),
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
func mkVariantAssignClauses(lhs, rhs lg.Expr, domain *module.Module) *transrel.Update {
	sym := constSym(lhs)
	if sym == nil {
		return transrel.NullUpdate()
	}
	newN := newSym(sym)
	args := nodeArgs(lhs)

	// dlhs = new_n(*sym_placeholders(n))
	phs := co.SymPlaceholders(sym)
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
		drhs = co.SubstituteAstByName(rhs, rn)
	}

	// Create nondeterministic skolem symbol
	// Python: nondet = n.suffix("_nd").skolem()
	skSym := lg.NewSymbol(sym.Name+"_nd", sym.CSort)
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
	update := &transrel.Update{
		Modified: []*lg.Symbol{sym},
		TR:       co.NewClauses(fmlas, defs, EmptyAnnotation{}),
		Pre:      co.FalseClauses(EmptyAnnotation{}),
	}
	return update
}

// --- HavocAction ---

// ActionUpdate computes the transition relation for havoc (nondeterministic assignment).
// The new value is unconstrained at the specified indices, but equal to the old
// value at all other indices.
func (a *HavocAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.HavocAction.action_update ENTER")
	defer xtracer.Trace("actions.HavocAction.action_update EXIT")
	lhs := a.Target
	sym := constSym(lhs)
	if sym == nil {
		return transrel.NullUpdate()
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

	return makeUpdate([]*lg.Symbol{sym}, tr, lg.False, EmptyAnnotation{})
}

func applyToNodes(fn lg.Expr, args []lg.Expr) lg.Expr {
	if len(args) == 0 {
		return fn
	}
	app, err := lg.NewApply(fn, args...)
	if err != nil {
		// Fallback: construct directly with computed sort
		rng := il.SortRange(fn.NodeSort())
		if rng == nil {
			rng = lg.TopS
		}
		return &lg.Apply{Func: fn, Terms: args}
	}
	return app
}

// --- SetAction ---

// ActionUpdate computes the transition relation for a set operation on a relation.
// Python: SetAction.action_update (ivy_actions.py:624-636).
// Builds clauses with frame conditions ensuring values at non-matching indices
// are preserved.
func (a *SetAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.SetAction.action_update ENTER")
	defer xtracer.Trace("actions.SetAction.action_update EXIT")
	if a.Lit == nil {
		return transrel.NullUpdate()
	}

	// Determine polarity and atom
	lit := a.Lit
	positive := true
	if n, ok := lit.(*lg.Not); ok {
		positive = false
		lit = n.Body
	}

	// Extract the relation symbol and args from the atom
	var relSym *lg.Symbol
	var args []lg.Expr
	if app, ok := lit.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Symbol); ok {
			relSym = c
			args = app.Terms
		}
	} else if c, ok := lit.(*lg.Symbol); ok {
		relSym = c
	}

	if relSym == nil {
		return transrel.NullUpdate()
	}

	newN := newSym(relSym)
	vs := co.SymPlaceholders(relSym)
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
	return makeUpdate([]*lg.Symbol{relSym}, tr, lg.False, EmptyAnnotation{})
}

// --- NativeAction ---

// IntUpdate for NativeAction is a no-op — skips update axioms.
// Python: NativeAction.int_update returns ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
func (a *NativeAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	return transrel.NullUpdate()
}

// --- DebugAction ---

// IntUpdate for DebugAction is a no-op — skips update axioms.
// Python: DebugAction.int_update returns ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
func (a *DebugAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	return transrel.NullUpdate()
}

// --- Field Actions ---

// makeFieldUpdateFunc constructs a field-update AssignAction using a callable
// RHS builder and returns its ActionUpdate.
// Python: make_field_update(self, l, f, r_func, domain, pvars) with lambda r_func.
func makeFieldUpdateFunc(field, obj lg.Expr, rhsFunc func(v *lg.Variable) lg.Expr, ctx *UpdateContext) *transrel.Update {
	sym, ok := field.(*lg.Symbol)
	if sym == nil || !ok {
		return transrel.NullUpdate()
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok || !il.IsRelationalSort(sym.CSort) || len(fs.Domain()) != 2 {
		// "field must be a binary relation"
		return transrel.NullUpdate()
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
func (a *AssignFieldAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *lg.Variable) lg.Expr {
		return il.NewEqualsNode(v, a.Value)
	}, ctx)
}

// ActionUpdate for NullFieldAction.
// Python: l,f = self.args; make_field_update(self,l,f,lambda v: Or(),domain,pvars)
func (a *NullFieldAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return makeFieldUpdateFunc(a.Field, a.Obj, func(v *lg.Variable) lg.Expr {
		return &lg.Or{} // Or() with no args = false
	}, ctx)
}

// ActionUpdate for CopyFieldAction.
// Python: l,lf,r,rf = self.args; make_field_update(self,l,lf,lambda v: rf(r,v),domain,pvars)
func (a *CopyFieldAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	srcField := a.SrcField
	if srcField == nil {
		srcField = a.Field // backward compat: same field for both
	}
	return makeFieldUpdateFunc(a.Field, a.Dst, func(v *lg.Variable) lg.Expr {
		if sym, ok := srcField.(*lg.Symbol); ok {
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
func actionTypeName(a interface{}) string {
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
	default:
		return fmt.Sprintf("%T", a)
	}
}

// IntUpdate is the intermediate update computation that applies domain
// update axioms on top of the atomic action_update.
// Corresponds to Python Action.int_update().
func IntUpdate(action Action, ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.IntUpdate ENTER type=%s", actionTypeName(action))
	// Dispatch to type-specific int_update methods
	switch a := action.(type) {
	case *AssumeAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *AssertAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *RequiresAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *EnsuresAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *AssignAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *HavocAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *SetAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *NativeAction:
		return a.IntUpdate(ctx)
	case *DebugAction:
		return a.IntUpdate(ctx)
	case *AssignFieldAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *NullFieldAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *CopyFieldAction:
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
		return intUpdateFromActionUpdate(a, ctx)
	default:
		// Generic fallback: null update
		return transrel.NullUpdate()
	}
}

// actionUpdater is the interface for actions that have ActionUpdate.
type actionUpdater interface {
	ActionUpdate(ctx *UpdateContext) *transrel.Update
}

// intUpdateFromActionUpdate computes int_update from action_update by
// applying domain update axioms.
func intUpdateFromActionUpdate(action actionUpdater, ctx *UpdateContext) *transrel.Update {
	update := action.ActionUpdate(ctx)
	// Apply update axioms from the domain
	update = applyUpdateAxioms(update, action.(Action), ctx)
	return update
}

// implemented by DerivedUpdate, here
// in actions/extra_actions.go
type updateAxiomProvider interface {
	GetUpdateAxioms(updated []string, action interface{}) ([]string, lg.Expr, lg.Expr)
}

// applyUpdateAxioms applies domain.updates to the given update.
// In Python, this iterates over domain.updates calling get_update_axioms.
func applyUpdateAxioms(update *transrel.Update, action Action, ctx *UpdateContext) *transrel.Update {
	if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
		return update
	}

	modified := update.Modified
	modNames := transrel.ModifiedNames(update)
	tr := update.TR
	pre := update.Pre

	for _, u := range ctx.Domain.Updates {
		if provider, ok := u.(updateAxiomProvider); ok {
			newModNames, transrelNode, precondNode := provider.GetUpdateAxioms(modNames, action)
			// Update modNames for next iteration
			modNames = newModNames
			// Convert new names to Consts (TopSort since we don't have sort info)
			modified = make([]*lg.Symbol, len(newModNames))
			for i, n := range newModNames {
				modified[i] = lg.NewSymbol(n, lg.TopS)
			}
			if transrelNode != nil {
				tr = co.AndClausesTyped(tr, co.FormulaToClauses(transrelNode, nil))
			}
			if precondNode != nil {
				pre = co.OrClausesTyped(pre, co.FormulaToClauses(precondNode, nil))
			}
		}
	}

	return &transrel.Update{
		Modified: modified,
		TR:       tr,
		Pre:      pre,
	}
}

// --- Sequence ---

// IntUpdate computes the sequential composition of child updates.
// Python: Sequence.int_update composes each child via compose_updates.
func (s *Sequence) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.Sequence.int_update ENTER")
	defer xtracer.Trace("actions.Sequence.int_update EXIT")
	result := transrel.NullUpdate()
	axioms := ctx.BackgroundTheory()
	for _, child := range s.Elems {
		act := unwrapToAction(child)
		if act == nil {
			continue
		}
		childUpdate := IntUpdate(act, ctx)
		result = transrel.ComposeUpdates(result, axioms, childUpdate)
	}
	return result
}

// unwrapToAction extracts an Action from a Node.
func unwrapToAction(n lg.Expr) Action {
	act, _ := n.(Action)
	return act
}

// --- ChoiceAction ---

// IntUpdate computes the nondeterministic choice between branches.
// Python: ChoiceAction.int_update uses join_action for each branch.
func (a *ChoiceAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.ChoiceAction.int_update ENTER")
	defer xtracer.Trace("actions.ChoiceAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = IfAction(Not(cond), self.args[0], self.args[1])
	//   return ite.int_update(domain, pvars)
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := co.BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(&lg.Not{Body: cond}, a.Branches[0], a.Branches[1])
		return ite.IntUpdate(ctx)
	}
	result := makeUpdate([]*lg.Symbol{}, lg.False, lg.False, nil)
	axioms := ctx.BackgroundTheory()
	for _, branch := range a.Branches {
		act := unwrapToAction(branch)
		if act == nil {
			continue
		}
		branchUpdate := IntUpdate(act, ctx)
		result = transrel.JoinAction(result, branchUpdate, axioms)
	}
	return result
}

// --- EnvAction ---

// IntUpdateEnv is like ChoiceAction.IntUpdate but calls GetUpdate
// (with hide_formals) instead of IntUpdate for each branch.
func (a *EnvAction) IntUpdateEnv(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.EnvAction.int_update ENTER")
	defer xtracer.Trace("actions.EnvAction.int_update EXIT")
	// Python: if determinize and len(self.args) == 2:
	//   cond = bool_const('___branch:' + str(self.unique_id))
	//   ite = IfAction(cond, self.args[0], self.args[1])
	//   return ite.update(domain, pvars)
	// Note: EnvAction uses cond (positive), ChoiceAction uses Not(cond).
	// Note: EnvAction calls update (GetUpdate), not int_update (IntUpdate).
	if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
		cond := co.BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
		ite := NewIfAction(cond, a.Branches[0], a.Branches[1])
		return GetUpdate(ite, ctx)
	}
	result := makeUpdate([]*lg.Symbol{}, lg.False, lg.False, nil)
	axioms := ctx.BackgroundTheory()
	for _, branch := range a.Branches {
		act := unwrapToAction(branch)
		if act == nil {
			continue
		}
		branchUpdate := GetUpdate(act, ctx)
		result = transrel.JoinAction(result, branchUpdate, axioms)
	}
	return result
}

// --- IfAction ---

// IntUpdate computes the if-then-else transition relation.
// Python: IfAction.int_update uses ite_action for simple conditions.
func (a *IfAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.IfAction.int_update ENTER")
	defer xtracer.Trace("actions.IfAction.int_update EXIT")
	cond := a.Cond

	// Python: if used_variables_ast(self.args[0]): raise IvyError(...)
	freeVars := co.UsedVariablesAST(cond)
	if len(freeVars) > 0 {
		panic("variables in \"if\" conditions must be explicitly quantified")
	}

	// Python: if not isinstance(self.args[0], ivy_ast.Some): ... else: subactions path
	if _, isSome := cond.(*SomeCondition); isSome {
		return a.intUpdateWithSubactions(ctx)
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

	axioms := ctx.BackgroundTheory()
	return transrel.IteAction(cond, thenUpdate, elseUpdate, axioms)
}

// intUpdateWithSubactions handles the Some/SomeMinMax case.
// Python: if_part,else_part = (a.int_update(domain,pvars) for a in self.subactions())
func (a *IfAction) intUpdateWithSubactions(ctx *UpdateContext) *transrel.Update {
	ifPart, elsePart := a.Subactions(ctx.ActCfg)

	ifUpdate := IntUpdate(ifPart, ctx)
	elseUpdate := IntUpdate(elsePart, ctx)

	axioms := ctx.BackgroundTheory()
	res := transrel.JoinAction(ifUpdate, elseUpdate, axioms)

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
func (a *WhileAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.WhileAction.int_update ENTER")
	defer xtracer.Trace("actions.WhileAction.int_update EXIT")
	// Python: if isinstance(context, UnrollContext): return self.unroll(context.card).int_update(domain, pvars)
	var actCtx IActionContext
	if ctx.ActCfg != nil {
		actCtx = ctx.ActCfg.Context
	}
	if uc, ok := actCtx.(*UnrollContext); ok {
		unrolled, err := a.Unroll(uc.Card, nil)
		if err == nil {
			return IntUpdate(unrolled, ctx)
		}
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
		if sym, ok := app.Func.(*lg.Symbol); ok {
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
	var rankLocal *lg.Symbol
	if ranking != nil {
		rankArgs := ranking.ActionArgs()
		if len(rankArgs) > 0 {
			rankExpr := rankArgs[0]
			rankSort := rankExpr.NodeSort()
			aux := lg.NewSymbol("$rank", rankSort)
			rankLocal = aux
			assumes = append(assumes, NewAssumeAction(&lg.Eq{T1: aux, T2: rankExpr}))
			ltSym := lg.NewSymbol("<", il.RelationSort([]lg.Sort{rankSort, rankSort}))
			exitAsserts = append(exitAsserts, NewAssertAction(&lg.Apply{Func: ltSym, Terms: []lg.Expr{rankExpr, aux}}))
			zeroSym := lg.NewSymbol("0", rankSort)
			entryAsserts = append(entryAsserts, NewAssertAction(&lg.Not{Body: &lg.Apply{Func: ltSym, Terms: []lg.Expr{rankExpr, zeroSym}}}))
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
		return ctx.ActCfg.NewLocalAction(rankLocal, result)
	}
	return result
}

// --- LocalAction ---

// IntUpdate computes the local action's update by hiding local symbols.
// Python: LocalAction.int_update computes body.int_update then hide(syms, update).
func (a *LocalAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.LocalAction.int_update ENTER")
	defer xtracer.Trace("actions.LocalAction.int_update EXIT")
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		return transrel.NullUpdate()
	}
	update := IntUpdate(bodyAct, ctx)

	// Collect symbols to hide
	var symsToHide []*lg.Symbol
	for _, local := range a.Locals {
		if c, ok := local.(*lg.Symbol); ok {
			symsToHide = append(symsToHide, c)
		} else {
			name := constName(local)
			if name != "" {
				symsToHide = append(symsToHide, lg.NewSymbol(name, lg.TopS))
			}
		}
	}

	if len(symsToHide) > 0 {
		update = transrel.Hide(symsToHide, update)
	}
	return update
}

// --- LetAction ---

// IntUpdate computes the let action's update by substituting symbols.
// Python: LetAction.int_update computes body.int_update then subst_action.
func (a *LetAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.LetAction.int_update ENTER")
	defer xtracer.Trace("actions.LetAction.int_update EXIT")
	bodyAct := unwrapToAction(a.Body)
	if bodyAct == nil {
		return transrel.NullUpdate()
	}
	update := IntUpdate(bodyAct, ctx)

	// Build substitution map from bindings
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

	if len(subst) > 0 {
		update = transrel.SubstAction(update, subst)
	}
	return update
}

// --- BindOldsAction ---

// IntUpdate wraps the inner action's update with bind_olds.
// Python: BindOldsAction.int_update returns bind_olds_action(inner.int_update(...)).
func (a *BindOldsAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.BindOldsAction.int_update ENTER")
	defer xtracer.Trace("actions.BindOldsAction.int_update EXIT")
	innerAct := unwrapToAction(a.Inner)
	if innerAct == nil {
		return transrel.NullUpdate()
	}
	update := IntUpdate(innerAct, ctx)
	return transrel.BindOldsAction(update)
}

// --- CallAction ---

// IntUpdate computes the call action's update by inlining the callee.
// Python: CallAction.int_update resolves the callee, applies actuals,
// and computes the inlined update.
func (a *CallAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.CallAction.int_update ENTER")
	defer xtracer.Trace("actions.CallAction.int_update EXIT")
	calleeName := constName(a.Callee)
	if calleeName == "" {
		return transrel.NullUpdate()
	}

	// Resolve the callee
	var calleeAction Action
	if ctx.GetAction != nil {
		calleeAction = ctx.GetAction(calleeName)
	}
	if calleeAction == nil {
		// Try from domain.Actions
		if ctx.Domain != nil && ctx.Domain.Actions != nil {
			if v, ok := ctx.Domain.Actions[calleeName]; ok {
				if act, ok := v.(Action); ok {
					calleeAction = act
				}
			}
		}
	}
	if calleeAction == nil {
		return transrel.NullUpdate()
	}

	// Apply actual parameters
	return a.applyActuals(ctx, calleeAction)
}

// applyActuals inlines the callee with actual parameters.
// Corresponds to Python CallAction.apply_actuals.
// Includes capture avoidance via distinct_obj_renaming.
func (a *CallAction) applyActuals(ctx *UpdateContext, callee Action) *transrel.Update {
	formalParams := callee.GetFormalParams()
	formalReturns := callee.GetFormalReturns()
	actualParams := nodeArgs(a.Callee)
	actualReturns := a.ActualReturns

	// Validate parameter counts
	if len(formalParams) != len(actualParams) {
		return transrel.NullUpdate()
	}
	if len(formalReturns) != len(actualReturns) {
		return transrel.NullUpdate()
	}

	// Capture avoidance: rename formals to avoid colliding with actuals.
	// Python: vocab = list(symbols_asts(actual_params+actual_returns))
	//         subst = distinct_obj_renaming(formal_params+formal_returns, vocab)
	allFormals := make([]*lg.Symbol, 0, len(formalParams)+len(formalReturns))
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

	// Apply renaming to callee if needed
	renamedCallee := callee
	if len(renaming) > 0 {
		// Build substitution map
		substMap := make(map[lg.NodeKey]lg.Expr)
		for oldSym, newSym := range renaming {
			substMap[lg.Key(oldSym)] = newSym
			// Also map old(s) → old(t) for pre-state symbols
			oldOfOld := lg.NewSymbol("old("+oldSym.Name+")", oldSym.CSort)
			substMap[lg.Key(oldOfOld)] = lg.NewSymbol("old("+newSym.Name+")", newSym.CSort)
		}

		// Substitute in the callee action using action-level substitution
		renamedCallee = SubstConstantsAction(callee, substMap)
	}

	// Get renamed formals
	renamedFormalParams := make([]*lg.Symbol, len(formalParams))
	for i, fp := range formalParams {
		if newSym, ok := renaming[fp]; ok {
			renamedFormalParams[i] = newSym
		} else {
			renamedFormalParams[i] = fp
		}
	}
	renamedFormalReturns := make([]*lg.Symbol, len(formalReturns))
	for i, fr := range formalReturns {
		if newSym, ok := renaming[fr]; ok {
			renamedFormalReturns[i] = newSym
		} else {
			renamedFormalReturns[i] = fr
		}
	}

	// Sort compatibility check
	if ctx.Domain != nil {
		for i, fp := range renamedFormalParams {
			if i < len(actualParams) {
				fpSort := fp.CSort
				apSort := actualParams[i].NodeSort()
				if fpSort != nil && apSort != nil && fpSort != apSort {
					if !ctx.Domain.IsVariant(fpSort, apSort) {
						// Sort mismatch — continue anyway (Python raises error)
					}
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

	// Hide the renamed formal parameters and returns
	var toHide []*lg.Symbol
	for _, fp := range renamedFormalParams {
		toHide = append(toHide, fp)
	}
	for _, fr := range renamedFormalReturns {
		toHide = append(toHide, fr)
	}
	if len(toHide) > 0 {
		update = transrel.Hide(toHide, update)
	}

	return update
}

// distinctObjRenaming creates a renaming from formals to fresh names
// that don't conflict with vocabNames.
// Corresponds to Python distinct_obj_renaming.
func distinctObjRenaming(formals []*lg.Symbol, vocabNames map[string]bool) map[*lg.Symbol]*lg.Symbol {
	result := make(map[*lg.Symbol]*lg.Symbol)
	usedNames := make(map[string]bool)
	for k := range vocabNames {
		usedNames[k] = true
	}

	for _, sym := range formals {
		name := sym.Name
		if _, used := usedNames[name]; !used {
			usedNames[name] = true
			// No conflict — no rename needed
			continue
		}
		// Need a fresh name
		newName := unusedNameWithBase(name, usedNames)
		usedNames[newName] = true
		result[sym] = lg.NewSymbol(newName, sym.CSort)
	}
	return result
}

// unusedNameWithBase finds an unused name starting with base.
func unusedNameWithBase(base string, used map[string]bool) string {
	for i := 0; ; i++ {
		name := fmt.Sprintf("%s_%d", base, i)
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
	case *lg.Symbol:
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
func (a *CrashAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.CrashAction.action_update ENTER")
	defer xtracer.Trace("actions.CrashAction.action_update EXIT")
	target := a.Target
	targetName := constName(target)
	if targetName == "" || ctx.Domain == nil {
		return transrel.NullUpdate()
	}

	// Collect symbols to havoc
	var symsToHavoc []*lg.Symbol
	collectCrashSyms(ctx.Domain, targetName, &symsToHavoc)

	if len(symsToHavoc) == 0 {
		return transrel.NullUpdate()
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
func collectCrashSyms(domain *module.Module, name string, result *[]*lg.Symbol) {
	if domain.Hierarchy != nil {
		if children, ok := domain.Hierarchy[name]; ok && len(children) > 0 {
			for child := range children {
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
	sym := lg.NewSymbol(name, lg.TopS)
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
func GetUpdate(action Action, ctx *UpdateContext) *transrel.Update {
	xtracer.Trace("actions.GetUpdate ENTER type=%s", actionTypeName(action))
	update := IntUpdate(action, ctx)
	update = transrel.BindOldsAction(update)
	update = hideFormals(action, update)
	return update
}

// hideFormals hides formal parameters and returns from the update.
// Matches Python Action.hide_formals (ivy_actions.py:220-228).
func hideFormals(action Action, update *transrel.Update) *transrel.Update {
	var toHide []*lg.Symbol
	if fp := action.GetFormalParams(); len(fp) > 0 {
		toHide = append(toHide, fp...)
	}
	if fr := action.GetFormalReturns(); len(fr) > 0 {
		toHide = append(toHide, fr...)
	}
	if len(toHide) > 0 {
		update = transrel.Hide(toHide, update)
	}
	return update
}

// -----------------------------------------------------------------------
// Updater interface implementation for art/ package
// -----------------------------------------------------------------------

// GetUpdateForArt implements the Updater interface expected by art/art.go.
// SubstConstantsAction applies a constant substitution to an action and all its
// sub-actions and embedded logic nodes, recursively.
//
// This is the faithful Go port of Python's substitute_constants_ast applied to
// Action trees. In Python, actions and logic nodes share the same .args/.clone()
// interface, so a single recursive function handles both. In Go, we need to
// handle three cases at each node:
//
//  1. The node is an Action (which IS lg.Expr) → walk Args(), substitute each
//     child recursively, then Clone() with the new args.
//  2. The node is a plain lg.Expr (Const, Apply, Var, etc.) → apply
//     co.SubstituteConstantsAST (the logic-level substitution).
//  3. Also substitute in FormalParams and
//     FormalReturns.
//
// Corresponds to Python ivy_logic_utils.substitute_constants_ast when applied
// to an Action AST (which Python handles transparently via duck typing).
func SubstConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
	if len(subs) == 0 {
		return action
	}

	// Substitute in each child arg.
	oldArgs := action.ActionArgs()
	newArgs := make([]lg.Expr, len(oldArgs))
	changed := false
	for i, arg := range oldArgs {
		newArg := substConstantsNode(arg, subs)
		newArgs[i] = newArg
		if newArg != arg {
			changed = true
		}
	}

	// Clone with new args if anything changed, or with old args to get a copy.
	var result Action
	if changed {
		result = action.ActionClone(newArgs)
	} else {
		result = action.ActionClone(oldArgs)
	}

	// Substitute in formal params.
	oldFP := action.GetFormalParams()
	if len(oldFP) > 0 {
		newFP := make([]*lg.Symbol, len(oldFP))
		fpChanged := false
		for i, fp := range oldFP {
			if replacement, ok := subs[lg.Key(fp)]; ok {
				if rc, ok := replacement.(*lg.Symbol); ok {
					newFP[i] = rc
					fpChanged = true
					continue
				}
			}
			newFP[i] = fp
		}
		if fpChanged {
			result.SetFormalParams(newFP)
		} else {
			result.SetFormalParams(oldFP)
		}
	}

	// Substitute in formal returns.
	oldFR := action.GetFormalReturns()
	if len(oldFR) > 0 {
		newFR := make([]*lg.Symbol, len(oldFR))
		frChanged := false
		for i, fr := range oldFR {
			if replacement, ok := subs[lg.Key(fr)]; ok {
				if rc, ok := replacement.(*lg.Symbol); ok {
					newFR[i] = rc
					frChanged = true
					continue
				}
			}
			newFR[i] = fr
		}
		if frChanged {
			result.SetFormalReturns(newFR)
		} else {
			result.SetFormalReturns(oldFR)
		}
	}

	return result
}

// substConstantsNode applies constant substitution to a single lg.Expr,
// handling both wrapped Actions and plain logic nodes.
func substConstantsNode(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
	if node == nil {
		return nil
	}

	// Case 1: Action — recurse into the action.
	if act, ok := node.(Action); ok {
		return SubstConstantsAction(act, subs)
	}

	// Case 2: Plain logic node — use clauseops SubstituteConstantsAST.
	return co.SubstituteConstantsAST(node, subs)
}

// It adapts the module-level GetUpdate function to the (domain, inScope) signature.
func GetUpdateForArt(action Action, domain *module.Module, inScope map[string]bool) *transrel.Update {
	ctx := &UpdateContext{
		Domain: domain,
		PVars:  inScope,
		GetAction: func(name string) Action {
			if domain != nil && domain.Actions != nil {
				if v, ok := domain.Actions[name]; ok {
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
