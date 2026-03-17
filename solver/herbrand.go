// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_solver.py.
//
// HerbrandModel and model extraction functions.

package solver

import (
	"fmt"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lu "github.com/glycerine/goivy/logicutil"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// HerbrandModel wraps a Z3 model and provides operations to extract
// sort universes, evaluate formulas, and produce Ivy-level facts.
// Corresponds to Python's HerbrandModel class.
type HerbrandModel struct {
	solver    *z3bridge.Solver
	model     *z3bridge.Model
	constants map[string][]z3bridge.Expr // sort name → universe elements
	sortMap   map[string]lg.Sort         // sort name → sort
	tr        *z3bridge.Translator
	sig       *il.Sig
}

// NewHerbrandModel creates a HerbrandModel from a Z3 solver and model.
// vocab is the set of constants used in the problem (for mining interpreted constants).
// Corresponds to Python's HerbrandModel.__init__.
func NewHerbrandModel(s *Solver, z3solver *z3bridge.Solver, model *z3bridge.Model, vocab []*lg.Const) *HerbrandModel {
	h := &HerbrandModel{
		solver:    z3solver,
		model:     model,
		constants: make(map[string][]z3bridge.Expr),
		sortMap:   make(map[string]lg.Sort),
		tr:        s.tr,
		sig:       s.sig,
	}

	// Extract sort universes from the Z3 model
	for _, z3sort := range model.Sorts() {
		sortName := z3sort.String()
		universe := model.SortUniverse(z3sort)
		h.constants[sortName] = universe
		// Try to map Z3 sort name back to Ivy sort
		if ivySort, ok := s.sig.Sorts[sortName]; ok {
			h.sortMap[sortName] = ivySort
		}
	}

	// Mine interpreted constants from the vocabulary
	h.mineInterpretedConstants(model, vocab)

	return h
}

// Sorts returns all sorts that have a universe in this model.
func (h *HerbrandModel) Sorts() []lg.Sort {
	var result []lg.Sort
	for name, sort := range h.sortMap {
		if _, ok := h.constants[name]; ok {
			result = append(result, sort)
		}
	}
	return result
}

// SortUniverse returns the Ivy-level universe for a sort.
func (h *HerbrandModel) SortUniverse(sort lg.Sort) []*lg.Const {
	name := il.SortName(sort)
	elems, ok := h.constants[name]
	if !ok {
		return nil
	}
	result := make([]*lg.Const, len(elems))
	for i, e := range elems {
		result[i] = constantFromZ3(sort, e)
	}
	return result
}

// SortedSortUniverse returns the universe elements for a sort, ordered by the
// `<` relation if one exists in the model. If no ordering exists, returns
// elements in their natural order.
//
// Corresponds to Python HerbrandModel.sorted_sort_universe (lines 839-855).
func (h *HerbrandModel) SortedSortUniverse(sort lg.Sort) []*lg.Const {
	name := il.SortName(sort)
	elems, ok := h.constants[name]
	if !ok {
		return nil
	}

	// Try to sort by the `<` relation
	if h.model != nil && h.tr != nil && len(elems) > 1 {
		// Build the `<` relation for this sort
		orderSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sort, sort}))
		orderZ3, err := h.tr.Translate(orderSym)
		if err == nil {
			_ = orderZ3 // Check if the model has an interpretation for `<`
			// Try to evaluate ordering between pairs
			sorted := make([]z3bridge.Expr, len(elems))
			copy(sorted, elems)

			// Simple insertion sort using the model's `<` interpretation
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0; j-- {
					// Check if sorted[j] < sorted[j-1]
					lt := h.evalLt(sort, sorted[j], sorted[j-1])
					if lt {
						sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
					} else {
						break
					}
				}
			}

			result := make([]*lg.Const, len(sorted))
			for i, e := range sorted {
				result[i] = constantFromZ3(sort, e)
			}
			return result
		}
	}

	// No ordering available, return in natural order
	result := make([]*lg.Const, len(elems))
	for i, e := range elems {
		result[i] = constantFromZ3(sort, e)
	}
	return result
}

// evalLt evaluates whether a < b in the model for the given sort.
func (h *HerbrandModel) evalLt(sort lg.Sort, a, b z3bridge.Expr) bool {
	if h.model == nil || h.tr == nil {
		return false
	}
	ctx := h.tr.Ctx
	// Build the < application and evaluate in the model
	lt := ctx.Lt(a, b)
	val, ok := h.model.Eval(lt, true)
	if !ok {
		return false
	}
	return val.String() == "true"
}

// SortUniverseZ3 returns the Z3-level universe for a sort name.
func (h *HerbrandModel) SortUniverseZ3(sortName string) []z3bridge.Expr {
	return h.constants[sortName]
}

// Eval evaluates a formula in the model. Variables are interpreted universally.
// Returns true if the formula holds for all variable assignments.
func (h *HerbrandModel) Eval(fmla lg.Node) bool {
	// Negate the formula and check: if no satisfying assignment, fmla is true
	negLit := il.NewLiteral(0, fmla)
	_, rows := h.Check(negLit)
	return len(rows) == 0
}

// EvalConstant evaluates a constant in the model, returning its model value.
func (h *HerbrandModel) EvalConstant(c *lg.Const) *lg.Const {
	return h.getModelConstant(c)
}

// EvalToConstant evaluates a term in the model, returning an Ivy constant.
func (h *HerbrandModel) EvalToConstant(t lg.Node) *lg.Const {
	zt, err := h.tr.Translate(t)
	if err != nil {
		return lg.NewConst("?", t.NodeSort())
	}
	val, ok := h.model.Eval(zt, true)
	if !ok {
		return lg.NewConst("?", t.NodeSort())
	}
	return constantFromZ3(t.NodeSort(), val)
}

// Check evaluates a literal against all possible variable assignments.
// Returns (vars, rows) where each row is a list of Ivy constants that
// satisfy the literal.
func (h *HerbrandModel) Check(lit *il.Literal) ([]*lg.Var, [][]*lg.Const) {
	// Get free variables in the literal
	fvMap := lu.FreeVariables(lit)
	var vs []*lg.Var
	for v := range fvMap {
		vs = append(vs, v)
	}

	if len(vs) == 0 {
		// No variables — just evaluate directly
		zfmla, err := h.tr.Translate(lit.Atom)
		if err != nil {
			return vs, nil
		}
		if lit.Polarity == 0 {
			zfmla = h.tr.Ctx.Not(zfmla)
		}
		val, ok := h.model.Eval(zfmla, true)
		if ok && val.String() == "true" {
			return vs, [][]*lg.Const{{}} // one empty row
		}
		return vs, nil
	}

	// Get ranges for each variable
	ranges := make([][]z3bridge.Expr, len(vs))
	for i, v := range vs {
		ranges[i] = h.variableRange(v)
	}

	// Translate the literal to Z3
	zfmla, err := h.tr.Translate(lit.Atom)
	if err != nil {
		return vs, nil
	}
	if lit.Polarity == 0 {
		zfmla = h.tr.Ctx.Not(zfmla)
	}

	// Translate variables to Z3
	zVs := make([]z3bridge.Expr, len(vs))
	for i, v := range vs {
		zv, err := h.tr.Translate(v)
		if err != nil {
			return vs, nil
		}
		zVs[i] = zv
	}

	// Enumerate all assignments
	var rows [][]*lg.Const
	h.enumerateAssignments(ranges, 0, make([]z3bridge.Expr, len(vs)),
		func(assignment []z3bridge.Expr) {
			// Substitute assignment into the formula
			fact := h.tr.Ctx.Substitute(zfmla, zVs, assignment)
			val, ok := h.model.Eval(fact, true)
			if ok && val.String() == "true" {
				row := make([]*lg.Const, len(vs))
				for j, v := range vs {
					row[j] = constantFromZ3(v.VSort, assignment[j])
				}
				rows = append(rows, row)
			}
		})

	return vs, rows
}

// enumerateAssignments generates the Cartesian product of ranges.
func (h *HerbrandModel) enumerateAssignments(ranges [][]z3bridge.Expr, depth int, current []z3bridge.Expr, f func([]z3bridge.Expr)) {
	if depth == len(ranges) {
		f(current)
		return
	}
	for _, val := range ranges[depth] {
		current[depth] = val
		h.enumerateAssignments(ranges, depth+1, current, f)
	}
}

// variableRange returns the Z3 universe for a variable's sort.
func (h *HerbrandModel) variableRange(v *lg.Var) []z3bridge.Expr {
	sort := v.VSort
	sortName := il.SortName(sort)

	// Enumerated sort: use the enumeration values
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		var result []z3bridge.Expr
		for _, name := range es.Extension {
			c := lg.NewConst(name, sort)
			zc, err := h.tr.Translate(c)
			if err == nil {
				result = append(result, zc)
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	// Boolean sort
	if lg.SortEqual(sort, lg.Boolean) {
		return []z3bridge.Expr{
			h.tr.Ctx.BoolVal(false),
			h.tr.Ctx.BoolVal(true),
		}
	}

	// Uninterpreted sort: use universe from model
	if elems, ok := h.constants[sortName]; ok {
		return elems
	}

	return nil
}

// getModelConstant evaluates a constant in the model.
func (h *HerbrandModel) getModelConstant(c *lg.Const) *lg.Const {
	sort := il.SortRange(c.CSort)

	// Handle enumerated sorts without native Z3 enums
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		for _, name := range es.Extension {
			w := lg.NewConst(name, sort)
			zc, err1 := h.tr.Translate(c)
			zw, err2 := h.tr.Translate(w)
			if err1 != nil || err2 != nil {
				continue
			}
			eq := h.tr.Ctx.Eq(zc, zw)
			val, ok := h.model.Eval(eq, true)
			if ok && val.String() == "true" {
				return w
			}
		}
		// Fallback: return first value
		if len(es.Extension) > 0 {
			return lg.NewConst(es.Extension[0], sort)
		}
	}

	// General case
	zt, err := h.tr.Translate(c)
	if err != nil {
		return lg.NewConst("?", sort)
	}
	val, ok := h.model.Eval(zt, true)
	if !ok {
		return lg.NewConst("?", sort)
	}
	return constantFromZ3(sort, val)
}

// mineInterpretedConstants discovers interpreted constants from the vocabulary
// by checking which sorts have interpreted elements in the model.
func (h *HerbrandModel) mineInterpretedConstants(model *z3bridge.Model, vocab []*lg.Const) {
	for _, c := range vocab {
		sort := il.SortRange(c.CSort)
		sortName := il.SortName(sort)
		if il.IsInterpretedSort(h.sig, sort) {
			continue
		}
		if _, ok := h.constants[sortName]; ok {
			continue // already have universe
		}
		// Try to evaluate the constant and seed the universe
		zt, err := h.tr.Translate(c)
		if err != nil {
			continue
		}
		val, ok := model.Eval(zt, true)
		if ok {
			h.constants[sortName] = append(h.constants[sortName], val)
		}
	}
}

// constantFromZ3 converts a Z3 value back to an Ivy constant.
// Corresponds to Python's constant_from_z3.
func constantFromZ3(sort lg.Sort, z3val z3bridge.Expr) *lg.Const {
	s := z3val.String()
	if s == "true" {
		return lg.NewConst("true", lg.Boolean)
	}
	if s == "false" {
		return lg.NewConst("false", lg.Boolean)
	}
	return lg.NewConst(s, sort)
}

// --- Model extraction to clauses ---

// ModelUniverseFacts extracts universe membership facts from a model for a sort.
// Returns a list of clause-like formulas: X=c0 | X=c1 | ... and ci != cj.
// Corresponds to Python's model_universe_facts.
func ModelUniverseFacts(h *HerbrandModel, sort lg.Sort, upclose bool) []lg.Node {
	if il.IsInterpretedSort(h.sig, sort) {
		return nil
	}
	elems := h.SortUniverse(sort)
	if len(elems) == 0 {
		return nil
	}

	var result []lg.Node

	// Universe closure: forall X. X=c0 | X=c1 | ...
	if !upclose {
		x, _ := lg.NewVar("X", sort)
		eqs := make([]lg.Node, len(elems))
		for i, c := range elems {
			eqs[i] = &lg.Eq{T1: x, T2: c}
		}
		result = append(result, &lg.Or{Terms: eqs})
	}

	// Distinctness: ci != cj for all i < j
	for i := 0; i < len(elems); i++ {
		for j := i + 1; j < len(elems); j++ {
			neq := &lg.Not{Body: &lg.Eq{T1: elems[i], T2: elems[j]}}
			result = append(result, neq)
		}
	}

	return result
}

// ModelFacts extracts all facts from a Herbrand model.
// Returns a Clauses set characterizing the model.
// Corresponds to Python's model_facts.
func ModelFacts(h *HerbrandModel, ignore func(*lg.Const) bool, clauses *clauseops.Clauses, upclose bool) *clauseops.Clauses {
	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	var fmlas []lg.Node

	// Universe facts for each sort
	for _, sort := range h.Sorts() {
		fmlas = append(fmlas, ModelUniverseFacts(h, sort, upclose)...)
	}

	// Constant values
	symSet := clauses.Symbols()
	for sym := range symSet {
		if ignore(sym) {
			continue
		}
		if h.sig != nil {
			if _, isCtor := h.sig.Constructors[sym.Name]; isCtor {
				continue
			}
		}
		// Only handle 0-arity constants
		if !il.IsFunctionSort(sym.CSort) {
			val := h.getModelConstant(sym)
			if val != nil {
				fmlas = append(fmlas, &lg.Eq{T1: sym, T2: val})
			}
		}
	}

	// Relation values
	for sym := range symSet {
		if ignore(sym) {
			continue
		}
		if il.IsRelationalSort(sym.CSort) && il.IsFunctionSort(sym.CSort) {
			arity := 0
			if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
				arity = fs.Arity()
			}
			if arity > 0 {
				fmlas = append(fmlas, RelationModelToClauses(h, sym, arity)...)
			}
		}
	}

	// Function values
	for sym := range symSet {
		if ignore(sym) {
			continue
		}
		if il.IsFunctionSort(sym.CSort) && !il.IsRelationalSort(sym.CSort) {
			if fs, ok := sym.CSort.(*lg.FunctionSort); ok && fs.Arity() >= 1 {
				fmlas = append(fmlas, FunctionModelToClauses(h, sym)...)
			}
		}
	}

	return clauseops.NewClauses(fmlas, nil, nil)
}

// RelationModelToClauses extracts the relation interpretation from the model.
// Returns formulas for both positive and negative ground instances.
// Corresponds to Python's relation_model_to_clauses.
func RelationModelToClauses(h *HerbrandModel, rel *lg.Const, arity int) []lg.Node {
	// Create a literal for the relation applied to fresh variables
	fs, ok := rel.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	dom := fs.Domain()
	vars := make([]lg.Node, len(dom))
	varPtrs := make([]*lg.Var, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVar(fmt.Sprintf("V%d", i), s)
		vars[i] = v
		varPtrs[i] = v
	}
	app := &lg.Apply{Func: rel, Terms: vars}

	var result []lg.Node

	// Positive instances
	posLit := il.NewLiteral(1, app)
	result = append(result, getLitFacts(h, posLit)...)

	// Negative instances
	negLit := il.NewLiteral(0, app)
	result = append(result, getLitFacts(h, negLit)...)

	return result
}

// FunctionModelToClauses extracts the function interpretation from the model.
// Corresponds to Python's function_model_to_clauses.
func FunctionModelToClauses(h *HerbrandModel, f *lg.Const) []lg.Node {
	fs, ok := f.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	rng := fs.Range()
	dom := fs.Domain()

	// Create term f(V0, V1, ...)
	vars := make([]lg.Node, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVar(fmt.Sprintf("V%d", i), s)
		vars[i] = v
	}
	fTerm := &lg.Apply{Func: f, Terms: vars}

	// For enumerated range, check each possible value
	if es, ok := rng.(*lg.EnumeratedSort); ok {
		var result []lg.Node
		for _, name := range es.Extension {
			c := lg.NewConst(name, rng)
			eqLit := il.NewLiteral(1, &lg.Eq{T1: fTerm, T2: c})
			result = append(result, getLitFacts(h, eqLit)...)
		}
		return result
	}

	// General: create Y = f(V0, ...) literal
	y, _ := lg.NewVar("Y", rng)
	eqLit := il.NewLiteral(1, &lg.Eq{T1: y, T2: fTerm})
	return getLitFacts(h, eqLit)
}

// getLitFacts returns ground instances of a literal that hold in the model.
// Corresponds to Python's get_lit_facts.
func getLitFacts(h *HerbrandModel, lit *il.Literal) []lg.Node {
	vs, rows := h.Check(lit)
	var result []lg.Node
	for _, row := range rows {
		// Build substitution
		subs := make(map[lg.NodeKey]lg.Node, len(vs))
		for j, v := range vs {
			subs[lg.Key(v)] = row[j]
		}
		// Apply substitution to the literal
		newAtom, err := lu.Substitute(lit.Atom, subs)
		if err != nil {
			continue
		}
		if lit.Polarity == 0 {
			result = append(result, &lg.Not{Body: newAtom})
		} else {
			result = append(result, newAtom)
		}
	}
	return result
}

// --- Additional solver utility functions ---

// SortCard returns the cardinality of an enumerated sort, or -1 for others.
// Corresponds to Python's sort_card.
func SortCard(sort lg.Sort) int {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		return es.Card()
	}
	return -1
}

// EnumeratedRange returns Z3 expressions for all elements of an enumerated sort.
// Corresponds to Python's enumerated_range.
func (s *Solver) EnumeratedRange(sort *lg.EnumeratedSort) ([]z3bridge.Expr, error) {
	var result []z3bridge.Expr
	for _, name := range sort.Extension {
		c := lg.NewConst(name, sort)
		zc, err := s.tr.Translate(c)
		if err != nil {
			return nil, err
		}
		result = append(result, zc)
	}
	return result, nil
}

// GetModelFromClauses checks satisfiability and returns a HerbrandModel if sat.
// This is the high-level API for model extraction.
func (s *Solver) GetModelFromClauses(clauses *clauseops.Clauses) (*HerbrandModel, error) {
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	result := z3solver.Check()
	if result == z3bridge.Unsat {
		return nil, nil // unsatisfiable
	}

	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}

	// Collect vocabulary
	symSet := clauses.Symbols()
	vocab := make([]*lg.Const, 0, len(symSet))
	for sym := range symSet {
		vocab = append(vocab, sym)
	}

	return NewHerbrandModel(s, z3solver, m, vocab), nil
}

// ClausesCase performs non-deterministic case splitting on clauses.
// Returns the clauses with each disjunction resolved to one case.
// Corresponds to Python's clauses_case.
// ClausesCase drops literals from disjunctions while maintaining satisfiability.
// Iterates model-based simplification until convergence.
//
// Corresponds to Python clauses_case (lines 1031-1058):
//   1. Check SAT, get model
//   2. Simplify each clause by model: drop false literals
//   3. Remove duplicates
//   4. Repeat until no new clauses generated
//
// Note: The Python version also integrates UnitRes for unit propagation.
// Full UnitRes integration requires a conversion layer between lg.Node and
// unitres.Literal that is not yet implemented. The current implementation
// performs model-based simplification iteratively without unit propagation.
func (s *Solver) ClausesCase(clauses *clauseops.Clauses) (*clauseops.Clauses, error) {
	// Check satisfiability
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	if z3solver.Check() == z3bridge.Unsat {
		return nil, nil
	}

	model := z3solver.Model()
	if model == nil {
		return clauses, nil
	}

	// Iterative model-based simplification
	currentClauses := clauses
	for {
		numOldFmlas := len(currentClauses.Fmlas)

		// Model-based simplification: for each clause (disjunction),
		// drop literals that are false in the model
		var newFmlas []lg.Node
		for _, f := range currentClauses.Fmlas {
			simplified := s.clauseModelSimp(model, f)
			newFmlas = append(newFmlas, simplified)
		}

		// Remove duplicates
		newFmlas = removeDuplicateFormulas(newFmlas)

		currentClauses = clauseops.NewClauses(newFmlas, currentClauses.Defs, currentClauses.Annot)

		// Convergence check
		if len(currentClauses.Fmlas) <= numOldFmlas {
			return currentClauses, nil
		}
	}
}

// clauseModelSimp simplifies a clause (disjunction) by dropping literals
// that are false in the model, keeping only those that are true.
// For non-ground literals, they are always kept.
//
// Corresponds to Python clause_model_simp (lines 1062-1080).
func (s *Solver) clauseModelSimp(model *z3bridge.Model, clause lg.Node) lg.Node {
	or, ok := clause.(*lg.Or)
	if !ok || len(or.Terms) <= 1 {
		return clause
	}

	var kept []lg.Node
	for _, lit := range or.Terms {
		// Non-ground literals are always kept
		if !il.IsGroundFormula(lit) {
			kept = append(kept, lit)
			continue
		}

		// Evaluate in the model
		zlit, err := s.tr.Translate(il.CloseFormula(lit))
		if err != nil {
			kept = append(kept, lit)
			continue
		}
		val, ok := model.Eval(zlit, true)
		if !ok || val.String() == "true" {
			kept = append(kept, lit)
		}
		// If false in model, drop it
	}

	if len(kept) == 0 {
		// All literals were false — return empty disjunction (false)
		return &lg.Or{Terms: nil}
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return &lg.Or{Terms: kept}
}

// removeDuplicateFormulas removes duplicate formulas based on string representation.
func removeDuplicateFormulas(fmlas []lg.Node) []lg.Node {
	seen := make(map[string]bool)
	var result []lg.Node
	for _, f := range fmlas {
		s := fmt.Sprint(f)
		if !seen[s] {
			seen[s] = true
			result = append(result, f)
		}
	}
	return result
}

// clausesCaseLegacy is the old implementation kept for reference.
func (s *Solver) clausesCaseLegacy(clauses *clauseops.Clauses) (*clauseops.Clauses, error) {
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	if z3solver.Check() == z3bridge.Unsat {
		return nil, nil
	}

	model := z3solver.Model()
	if model == nil {
		return clauses, nil
	}

	var newFmlas []lg.Node
	for _, f := range clauses.Fmlas {
		if or, ok := f.(*lg.Or); ok && len(or.Terms) > 1 {
			// Pick the first disjunct that is true in the model
			picked := false
			for _, d := range or.Terms {
				zd, err := s.tr.Translate(il.CloseFormula(d))
				if err != nil {
					continue
				}
				val, ok := model.Eval(zd, true)
				if ok && val.String() == "true" {
					newFmlas = append(newFmlas, d)
					picked = true
					break
				}
			}
			if !picked {
				newFmlas = append(newFmlas, or.Terms[0])
			}
		} else {
			newFmlas = append(newFmlas, f)
		}
	}

	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return clauseops.NewClauses(newFmlas, defs, clauses.Annot), nil
}
