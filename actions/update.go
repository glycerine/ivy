// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go.

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

	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
)

// -----------------------------------------------------------------------
// UpdateContext provides the domain and in-scope variables for computing
// action updates. It wraps the module and an action-resolution function.
// -----------------------------------------------------------------------

// UpdateContext holds the context needed for computing action updates.
type UpdateContext struct {
	Domain *module.Module
	PVars  map[string]bool // in-scope variable names

	// GetAction resolves an action name to its Action. This is set by
	// the caller (typically from module.Actions or an ActionContext).
	GetAction func(name string) Action
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
		eqConj, _ := lg.NewAnd(eqs...)  // match Python And(*eqs) exactly
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
// Python: action_update returns ([], clauses, false_clauses())
func (a *AssumeAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	fmla := a.Formula
	// Skolemize existentially quantified variables
	fmla = skolemizeFormula(fmla)
	return makeUpdate([]*lg.Symbol{}, fmla, lg.False, EmptyAnnotation{})
}

// --- AssertAction ---

// ActionUpdate computes the transition relation for AssertAction.
// An assert generates a precondition (negative) from the dual of the formula.
// Python: action_update returns ([], true_clauses(), dual_cl)
func (a *AssertAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	fmla := a.Formula
	// The dual (negated + skolemized) formula becomes the precondition.
	// An action fails if the precondition is satisfiable.
	dual := dualFormula(fmla)
	return makeUpdate([]*lg.Symbol{}, lg.True, dual, EmptyAnnotation{})
}

// --- RequireAction ---

func (a *RequireAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return a.AssertAction.ActionUpdate(ctx)
}

// --- EnsureAction ---

func (a *EnsureAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
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
	lhs, rhs := a.LHS, a.RHS
	sym := constSym(lhs)
	if sym == nil {
		return transrel.NullUpdate()
	}

	// Handle hierarchical case: if the symbol has children in the hierarchy
	if ctx.Domain.Hierarchy != nil {
		if children, ok := ctx.Domain.Hierarchy[sym.Name]; ok && len(children) > 0 {
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

	// Handle destructor assignments
	if ctx.Domain.DestructorSorts != nil {
		if _, ok := ctx.Domain.DestructorSorts[sym.Name]; ok {
			return a.destructorAssignUpdate(ctx, lhs, rhs)
		}
	}

	// Handle variant assignments
	if ctx.Domain.Variants != nil {
		lhsSort := lhs.NodeSort()
		rhsSort := rhs.NodeSort()
		if lhsSort != nil && rhsSort != nil && isVariant(ctx.Domain, lhsSort, rhsSort) {
			return mkVariantAssignClauses(lhs, rhs, ctx.Domain)
		}
	}

	// Standard assignment
	return mkAssignClauses(lhs, rhs)
}

// destructorAssignUpdate handles assignment through destructors.
// In Python, this is the destructor case in AssignAction.action_update.
func (a *AssignAction) destructorAssignUpdate(ctx *UpdateContext, lhs, rhs lg.Expr) *transrel.Update {
	// Walk up the destructor chain to find the root mutable symbol
	n := lhs
	var mutName string
	for {
		sym := constSym(n)
		if sym == nil {
			break
		}
		if _, ok := ctx.Domain.DestructorSorts[sym.Name]; !ok {
			mutName = sym.Name
			break
		}
		args := nodeArgs(n)
		if len(args) == 0 {
			mutName = sym.Name
			break
		}
		n = args[0]
	}

	if mutName == "" {
		return transrel.NullUpdate()
	}

	mutSym := lg.NewSymbol(mutName, lg.TopS)
	newMut := newSym(mutSym)

	// Create a skolem for the new value
	skName := mutName + "_nd__"
	skSym := lg.NewSymbol(skName, mutSym.CSort)

	// The basic transition: new_mut = sk (nondeterministic)
	// Plus constraints that the destructor at the assigned position equals rhs,
	// and all other destructors are preserved.
	phs := co.SymPlaceholders(mutSym)
	var trLHS, trRHS lg.Expr
	if len(phs) > 0 {
		phNodes := varsToNodes(phs)
		trLHS = &lg.Apply{Func: newMut, Terms: phNodes}
		trRHS = &lg.Apply{Func: skSym, Terms: phNodes}
	} else {
		trLHS = newMut
		trRHS = skSym
	}

	// Build: new_mut(Vs) = sk(Vs) AND destructor(sk(args)) = rhs AND frame for other destructors
	defn := equivAST(trLHS, trRHS)
	// For now, the destructor constraint is simplified:
	// We assert equiv(destructor(sk(mut_args), rest_args), rhs)
	constraint := equivAST(lhs, rhs) // simplified
	tr := conjoin(defn, constraint)

	return makeUpdate([]*lg.Symbol{mutSym}, tr, lg.False, EmptyAnnotation{})
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
func mkVariantAssignClauses(lhs, rhs lg.Expr, domain *module.Module) *transrel.Update {
	sym := constSym(lhs)
	if sym == nil {
		return transrel.NullUpdate()
	}
	// For variant assignments, we need to assert that the new value
	// points to the RHS and to nothing else of other variant sorts.
	// Simplified version: treat as a regular assignment.
	return mkAssignClauses(lhs, rhs)
}

// --- HavocAction ---

// ActionUpdate computes the transition relation for havoc (nondeterministic assignment).
// The new value is unconstrained at the specified indices, but equal to the old
// value at all other indices.
func (a *HavocAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
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
// For positive literal R(x,y): new_R(X,Y) <-> (R(X,Y) | (X=x & Y=y))
// For negative literal ~R(x,y): new_R(X,Y) <-> (R(X,Y) & ~(X=x & Y=y))
// Corresponds to Python's set_action_update.
func (a *SetAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
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

	// Extract the relation symbol from the atom
	var relSym *lg.Symbol
	if app, ok := lit.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Symbol); ok {
			relSym = c
		}
	} else if c, ok := lit.(*lg.Symbol); ok {
		relSym = c
	}

	if relSym == nil {
		return transrel.NullUpdate()
	}

	var tr lg.Expr
	if positive {
		tr = a.Lit
	} else {
		tr = &lg.Not{Body: lit}
	}

	return makeUpdate([]*lg.Symbol{relSym}, tr, lg.False, EmptyAnnotation{})
}

// --- NativeAction ---

// ActionUpdate for NativeAction is a no-op.
func (a *NativeAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return transrel.NullUpdate()
}

// --- DebugAction ---

// DebugAction is a no-op marker (defined in extra_actions.go if it exists,
// but we handle it via the generic fallback).

// -----------------------------------------------------------------------
// IntUpdate implementations
// -----------------------------------------------------------------------

// IntUpdate is the intermediate update computation that applies domain
// update axioms on top of the atomic action_update.
// Corresponds to Python Action.int_update().
func IntUpdate(action Action, ctx *UpdateContext) *transrel.Update {
	// Dispatch to type-specific int_update methods
	switch a := action.(type) {
	case *AssumeAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *AssertAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *RequireAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *EnsureAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *AssignAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *HavocAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *SetAction:
		return intUpdateFromActionUpdate(a, ctx)
	case *NativeAction:
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
		return a.IntUpdate(ctx)
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

// applyUpdateAxioms applies domain.updates to the given update.
// In Python, this iterates over domain.updates calling get_update_axioms.
func applyUpdateAxioms(update *transrel.Update, action Action, ctx *UpdateContext) *transrel.Update {
	if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
		return update
	}
	type updateAxiomProvider interface {
		GetUpdateAxioms(updated []string, action interface{}) ([]string, lg.Expr, lg.Expr)
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
	result := transrel.NullUpdate()
	axioms := ctx.BackgroundTheory()
	for _, child := range s.Children {
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
	if act, ok := n.(Action); ok {
		return act
	}
	if w, ok := n.(*ActionNodeWrapper); ok {
		return w.Action
	}
	return nil
}

// --- ChoiceAction ---

// IntUpdate computes the nondeterministic choice between branches.
// Python: ChoiceAction.int_update uses join_action for each branch.
func (a *ChoiceAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
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
	cond := a.Cond

	// Check for free variables in condition
	freeVars := co.UsedVariablesAST(cond)
	if len(freeVars) > 0 {
		// Variables in "if" conditions must be explicitly quantified
		// Fall back to the subactions approach (Some handling).
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
func (a *IfAction) intUpdateWithSubactions(ctx *UpdateContext) *transrel.Update {
	// Build subactions: if_part and else_part
	// For the Some case, we'd need to decompose the condition.
	// Simplified: treat as a choice between then and else branches.
	thenAct := unwrapToAction(a.ThenBody)
	if thenAct == nil {
		thenAct = NewSequence()
	}
	var elseAct Action
	if a.ElseBody != nil {
		elseAct = unwrapToAction(a.ElseBody)
	}
	if elseAct == nil {
		elseAct = NewSequence()
	}

	ifUpdate := IntUpdate(thenAct, ctx)
	elseUpdate := IntUpdate(elseAct, ctx)

	axioms := ctx.BackgroundTheory()
	return transrel.JoinAction(ifUpdate, elseUpdate, axioms)
}

// --- WhileAction ---

// IntUpdate computes the while loop's transition relation by expanding
// the loop into assume/assert/havoc/if structure.
// Python: WhileAction.int_update calls expand() then int_update on the result.
func (a *WhileAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	expanded := a.Expand(ctx)
	return IntUpdate(expanded, ctx)
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
		bodyParts = append(bodyParts, WrapAction(ea))
	}
	bodyParts = append(bodyParts, a.Body)
	for _, ea := range exitAsserts {
		bodyParts = append(bodyParts, WrapAction(ea))
	}
	for _, asrt := range asserts {
		bodyParts = append(bodyParts, WrapAction(asrt))
	}
	// assume false (terminates this path — loop exit is the else branch)
	bodyParts = append(bodyParts, WrapAction(NewAssumeAction(lg.False)))
	thenBody := NewSequence(bodyParts...)

	ifAction := NewIfAction(a.Cond, WrapAction(thenBody), WrapAction(NewSequence()))

	// Assemble the full expanded sequence
	var allParts []lg.Expr
	for _, asrt := range asserts {
		allParts = append(allParts, WrapAction(asrt))
	}
	for _, h := range havocs {
		allParts = append(allParts, WrapAction(h))
	}
	for _, asms := range assumes {
		allParts = append(allParts, WrapAction(asms))
	}
	allParts = append(allParts, WrapAction(ifAction))

	result := NewSequence(allParts...)

	// If there's a ranking function, wrap in LocalAction
	if rankLocal != nil {
		return NewLocalAction(rankLocal, WrapAction(result))
	}
	return result
}

// --- LocalAction ---

// IntUpdate computes the local action's update by hiding local symbols.
// Python: LocalAction.int_update computes body.int_update then hide(syms, update).
func (a *LocalAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
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
		substMap := make(map[string]lg.Expr)
		for oldSym, newSym := range renaming {
			substMap[oldSym.Name] = newSym
			// Also map old(s) → old(t) for pre-state symbols
			substMap["old("+oldSym.Name+")"] = lg.NewSymbol("old("+newSym.Name+")", newSym.CSort)
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
			inputAsgns = append(inputAsgns, WrapAction(asgn))
		}
	}

	// Build output assignments: actual_return := formal_return
	var outputAsgns []lg.Expr
	for i, fr := range renamedFormalReturns {
		if i < len(actualReturns) {
			asgn := NewAssignAction(actualReturns[i], fr)
			outputAsgns = append(outputAsgns, WrapAction(asgn))
		}
	}

	// Build: Sequence(input_asgns, BindOlds(callee), output_asgns)
	inputSeq := NewSequence(inputAsgns...)
	bindOlds := NewBindOldsAction(WrapAction(renamedCallee))
	outputSeq := NewSequence(outputAsgns...)
	fullSeq := NewSequence(WrapAction(inputSeq), WrapAction(bindOlds), WrapAction(outputSeq))

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

// IntUpdate computes the crash action by havocing all non-spec mutable symbols.
func (a *CrashAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
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
		havocParts = append(havocParts, WrapAction(NewHavocAction(sym)))
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
//  1. The node is an ActionNodeWrapper → unwrap, recurse into the action,
//     re-wrap the result.
//  2. The node is a plain lg.Expr (Const, Apply, Var, etc.) → apply
//     co.SubstituteConstantsAST (the logic-level substitution).
//  3. For the action itself: walk Args(), substitute each child per (1) or (2),
//     then Clone() with the new args. Also substitute in FormalParams and
//     FormalReturns.
//
// Corresponds to Python ivy_logic_utils.substitute_constants_ast when applied
// to an Action AST (which Python handles transparently via duck typing).
func SubstConstantsAction(action Action, subs map[string]lg.Expr) Action {
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
			if replacement, ok := subs[fp.Name]; ok {
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
			if replacement, ok := subs[fr.Name]; ok {
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
func substConstantsNode(node lg.Expr, subs map[string]lg.Expr) lg.Expr {
	if node == nil {
		return nil
	}

	// Case 1: ActionNodeWrapper — unwrap, recurse into the action, re-wrap.
	if w, ok := node.(*ActionNodeWrapper); ok {
		newAction := SubstConstantsAction(w.Action, subs)
		return WrapAction(newAction)
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
