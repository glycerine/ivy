// Ported to Go from ivy_solver.py.
//
// HerbrandModel and model extraction functions.

package goivy

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// HerbrandModel wraps a Z3 model and provides operations to extract
// sort universes, evaluate formulas, and produce Ivy-level facts.
// Corresponds to Python's HerbrandModel class. (ivy_solver.py:835)
type HerbrandModel struct {
	solver    *smt.Z3Solver
	model     *smt.Model
	constants map[string][]smt.Z3Expr // sort name → universe elements
	sortMap   map[string]Sort         // sort name → sort
	tr        *Translator
	sig       *Sig
}

// NewHerbrandModel creates a HerbrandModel from a Z3 solver and model.
// vocab is the set of constants used in the problem (for mining interpreted constants).
// Corresponds to Python's HerbrandModel.__init__.
func NewHerbrandModel(s *Solver, z3solver *smt.Z3Solver, model *smt.Model, vocab []*Const) *HerbrandModel {
	xtracer.Trace("ivy_solver.py:909 HerbrandModel.__init__() ENTER")
	h := &HerbrandModel{
		solver:    z3solver,
		model:     model,
		constants: make(map[string][]smt.Z3Expr),
		sortMap:   make(map[string]Sort),
		tr:        s.tr,
		sig:       s.sig,
	}

	// Extract sort universes from the Z3 model.
	// Python: self.constants = dict((sort_from_z3(s), model.get_universe(s)) ...)
	for _, z3sort := range model.Sorts() {
		sortName := z3sort.String()
		universe := model.SortUniverse(z3sort)
		h.constants[sortName] = universe
		// Primary: use translator's reverse map (matches Python's sort_from_z3)
		if ivySort, ok := s.tr.SortFromZ3(z3sort); ok {
			h.sortMap[sortName] = ivySort
		} else if ivySort, ok := s.sig.Sorts.Get2(sortName); ok {
			// Fallback to sig lookup by name
			h.sortMap[sortName] = ivySort
		}
	}

	// Mine interpreted constants from the vocabulary
	h.mineInterpretedConstants(model, vocab)

	return h
}

// Sorts returns all sorts that have a universe in this model.
func (h *HerbrandModel) Sorts() []Sort {
	var result []Sort
	for name, sort := range h.sortMap {
		if _, ok := h.constants[name]; ok {
			result = append(result, sort)
		}
	}
	return result
}

// SortUniverse returns the Ivy-level universe for a sort.
func (h *HerbrandModel) SortUniverse(sort Sort) []*Const {
	name := IvySortName(sort)
	elems, ok := h.constants[name]
	if !ok {
		return nil
	}
	result := make([]*Const, len(elems))
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
// Python builds a SortOrder from the Ivy `<` symbol applied to variables,
// translates to Z3, then sorts using substitute + model.eval. This works
// for user-defined `<` on uninterpreted sorts, not just Z3 built-in Lt.
func (h *HerbrandModel) SortedSortUniverse(sort Sort) []*Const {
	name := IvySortName(sort)
	elems, ok := h.constants[name]
	if !ok {
		return nil
	}

	// Try to sort by the `<` relation (matching Python's SortOrder approach)
	if h.model != nil && h.tr != nil && len(elems) > 1 {
		sorted, ok := h.trySortByOrder(sort, elems)
		if ok {
			result := make([]*Const, len(sorted))
			for i, e := range sorted {
				result[i] = constantFromZ3(sort, e)
			}
			return result
		}
	}

	// No ordering available, return in natural order
	result := make([]*Const, len(elems))
	for i, e := range elems {
		result[i] = constantFromZ3(sort, e)
	}
	return result
}

// trySortByOrder attempts to sort elements using the Ivy `<` relation
// translated to Z3 and evaluated in the model. This matches Python's
// SortOrder class (ivy_solver.py:787-797) which uses atom_to_z3(order(*vs))
// with substitute + model.eval, supporting user-defined orderings on
// uninterpreted sorts.
// HerbrandModel is in ivy_solver.py:835-914.
func (h *HerbrandModel) trySortByOrder(sort Sort, elems []smt.Z3Expr) (sorted []smt.Z3Expr, ok bool) {
	xtracer.Trace("ivy_solver.py:851 HerbrandModel.sorted_sort_universe() top. solver/herbrand.go:132") // not seen in make golden.

	// Python's approach: (ivy_solver.py:850 in sorted_sort_universe()):
	//   vs = [Variable("X", sort), Variable("Y", sort)]
	//   order = Symbol("<", LogicRelationSort([sort, sort]))
	//   order_atom = atom_to_z3(order(*vs))
	//   z3_vs = list(map(term_to_z3, vs))
	//   sorted(elems, key=cmp_to_key(SortOrder(z3_vs, order_atom, self.model)))
	defer func() {
		if r := recover(); r != nil {
			// actually this recover is not hit during "make golden".
			xtracer.Trace("ivy_solver.py:866 HerbrandModel.sorted_sort_universe(): IndexError from order.to_z3() | solver/herbrand.go:143")

			sorted = nil
			ok = false
			//panic(r) // panic in solver/solver2_test.go/TestClausesModelToDiagram_NoTautologies: z3 bridge panic on error: Sort mismatch at argument #1 for function (declare-fun < (Int Int) Bool) supplied sort is T [recovered, repanicked]
		}
	}()

	// Create Ivy variables and the < application
	xVar, _ := NewVariable("X", sort)
	yVar, _ := NewVariable("Y", sort)
	orderSym := NewConst("<", LogicRelationSort([]Sort{sort, sort}))
	orderApp, err := NewApply(orderSym, xVar, yVar)
	if err != nil {
		return nil, false
	}

	// Translate the order application and variables to Z3
	orderZ3, err := h.tr.Translate(orderApp)
	if err != nil {
		return nil, false
	}
	z3X, err := h.tr.Translate(xVar)
	if err != nil {
		return nil, false
	}
	z3Y, err := h.tr.Translate(yVar)
	if err != nil {
		return nil, false
	}

	ctx := h.tr.Ctx

	// Comparator function: substitute X->a, Y->b in order_atom, evaluate in model
	lessFunc := func(a, b smt.Z3Expr) bool {
		fact := SubstituteZ3(ctx, orderZ3, [][2]smt.Z3Expr{{z3X, a}, {z3Y, b}})
		val, evalOk := h.model.Eval(fact, true)
		if !evalOk {
			return false
		}
		return val.String() == "true"
	}

	// Insertion sort (matching Python's sorted() with cmp_to_key)
	sorted = make([]smt.Z3Expr, len(elems))
	copy(sorted, elems)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0; j-- {
			if lessFunc(sorted[j], sorted[j-1]) {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			} else {
				break
			}
		}
	}
	return sorted, true
}

// SortUniverseZ3 returns the Z3-level universe for a sort name.
func (h *HerbrandModel) SortUniverseZ3(sortName string) []smt.Z3Expr {
	return h.constants[sortName]
}

// Eval evaluates a formula in the model. Variables are interpreted universally.
// Returns true if the formula holds for all variable assignments.
func (h *HerbrandModel) Eval(fmla Expr) bool {
	// Negate the formula and check: if no satisfying assignment, fmla is true
	negLit := NewLiteral(0, fmla)
	_, rows := h.Check(negLit)
	return len(rows) == 0
}

// EvalConstant evaluates a constant in the model, returning its model value.
func (h *HerbrandModel) EvalConstant(c *Const) *Const {
	return h.getModelConstant(c)
}

// EvalToConstant evaluates a term in the model, returning an Ivy expression.
// Returns lg.True/lg.False for boolean values (matching Python's constant_from_z3
// which returns ivy_logic.And()/ivy_logic.Or()).
// This matches the trace.Model interface which declares return type lg.Expr.
func (h *HerbrandModel) EvalToConstant(t Expr) Expr {
	zt, err := h.tr.Translate(t)
	if err != nil {
		return NewConst("?", t.NodeSort())
	}
	val, ok := h.model.Eval(zt, true)
	if !ok {
		return NewConst("?", t.NodeSort())
	}
	return constantFromZ3Expr(t.NodeSort(), val)
}

// Check evaluates a literal against all possible variable assignments.
// Returns (vars, rows) where each row is a list of Ivy constants that
// satisfy the literal.
func (h *HerbrandModel) Check(lit *LogicLiteral) ([]*LogicVariable, [][]*Const) {
	// Get free variables in the literal
	fvMap := FreeVariables(lit)
	var vs []*LogicVariable
	for _, v := range fvMap.All() {
		if vv, ok := v.(*LogicVariable); ok {
			vs = append(vs, vv)
		}
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
			return vs, [][]*Const{{}} // one empty row
		}
		return vs, nil
	}

	// Get ranges for each variable
	ranges := make([][]smt.Z3Expr, len(vs))
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
	zVs := make([]smt.Z3Expr, len(vs))
	for i, v := range vs {
		zv, err := h.tr.Translate(v)
		if err != nil {
			return vs, nil
		}
		zVs[i] = zv
	}

	// Enumerate all assignments
	var rows [][]*Const
	h.enumerateAssignments(ranges, 0, make([]smt.Z3Expr, len(vs)),
		func(assignment []smt.Z3Expr) {
			// Substitute assignment into the formula
			fact := h.tr.Ctx.Substitute(zfmla, zVs, assignment)
			val, ok := h.model.Eval(fact, true)
			if ok && val.String() == "true" {
				row := make([]*Const, len(vs))
				for j, v := range vs {
					row[j] = constantFromZ3(v.VSort, assignment[j])
				}
				rows = append(rows, row)
			}
		})

	return vs, rows
}

// enumerateAssignments generates the Cartesian product of ranges.
func (h *HerbrandModel) enumerateAssignments(ranges [][]smt.Z3Expr, depth int, current []smt.Z3Expr, f func([]smt.Z3Expr)) {
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
func (h *HerbrandModel) variableRange(v *LogicVariable) []smt.Z3Expr {
	sort := v.VSort
	sortName := IvySortName(sort)

	// Enumerated sort: use the enumeration values
	if es, ok := sort.(*LogicEnumeratedSort); ok {
		var result []smt.Z3Expr
		for _, name := range es.Extension {
			c := NewConst(name, sort)
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
	if SortEqual(sort, Boolean) {
		return []smt.Z3Expr{
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
func (h *HerbrandModel) getModelConstant(c *Const) *Const {
	xtracer.Trace("ivy_solver.py:1005 get_model_constant() ENTER")
	sort := SortRange(c.CSort)

	// Handle enumerated sorts without native Z3 enums.
	// Python (ivy_solver.py:1055): if isinstance(s, EnumeratedSort) and not use_z3_enums:
	if es, ok := sort.(*LogicEnumeratedSort); ok && (h.tr.s == nil || !h.tr.s.opts.UseZ3Enums) {
		for _, name := range es.Extension {
			w := NewConst(name, sort)
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
			return NewConst(es.Extension[0], sort)
		}
	}

	// General case
	zt, err := h.tr.Translate(c)
	if err != nil {
		return NewConst("?", sort)
	}
	val, ok := h.model.Eval(zt, true)
	if !ok {
		return NewConst("?", sort)
	}
	result := constantFromZ3(sort, val)

	// When EnableInterpretedEnums is active and this is an interpreted enum,
	// the model value is a numeral — map it back to the enum constant name.
	if EnableInterpretedEnums {
		if es, ok2 := sort.(*LogicEnumeratedSort); ok2 && h.sig != nil {
			if _, interped := h.sig.Interp[es.Name]; interped {
				if idx, err := strconv.Atoi(result.Name); err == nil && idx >= 0 && idx < len(es.Extension) {
					return NewConst(es.Extension[idx], sort)
				}
			}
		}
	}
	return result
}

// mineInterpretedConstants discovers interpreted constants from the vocabulary
// by evaluating symbols in the model and collecting numeral values from
// their interpretations (including ITE branches).
//
// Corresponds to Python's mine_interpreted_constants (ivy_solver.py:809-818).
// Python collects model values for symbols whose range sort IS interpreted,
// building term sym(V0,V1,...), evaluating in the model, and collecting
// numerals from the result.
func (h *HerbrandModel) mineInterpretedConstants(model *smt.Model, vocab []*Const) {
	if h.sig == nil {
		return
	}

	// Build set of interpreted sorts (Python: sorts = ivy_logic.interpreted_sorts())
	interpSorts := make(map[string]Sort)
	for name := range h.sig.Interp {
		if s, ok := h.sig.Sorts.Get2(name); ok {
			interpSorts[name] = s
		}
	}

	// Initialize sort_values for each interpreted sort
	sortValues := make(map[string]map[string]smt.Z3Expr)
	for name := range interpSorts {
		sortValues[name] = make(map[string]smt.Z3Expr)
	}

	// For each symbol in vocab, if its range sort is interpreted,
	// collect model values.
	// Python: for s in vocab:
	//     sort = s.sort.rng
	//     if sort in sort_values:
	//         sort_values[sort].update(collect_model_values(sort, model, s))
	for _, sym := range vocab {
		rng := SortRange(sym.CSort)
		rngName := IvySortName(rng)
		if _, isInterp := sortValues[rngName]; !isInterp {
			continue // range sort is not interpreted, skip
		}

		// collect_model_values: build sym(V0,V1,...), translate, eval, collect numerals
		phs := SymPlaceholders(sym)
		var term Expr
		if len(phs) == 0 {
			term = sym
		} else {
			args := make([]Expr, len(phs))
			for i, v := range phs {
				args[i] = v
			}
			term = MustApply(sym, args...)
		}

		z3term, err := h.tr.Translate(term)
		if err != nil {
			continue
		}
		val, ok := model.Eval(z3term, true)
		if !ok {
			continue
		}

		// Collect numerals from the evaluation result (including ITE branches)
		nums := CollectNumeralsRecursive(val)
		for _, n := range nums {
			key := n.String()
			sortValues[rngName][key] = n
		}
	}

	// Python: return dict((x, list(map(get_const, list(y)))) for x,y in sort_values.items())
	// where get_const = lambda c: model.eval(term_to_z3(c), model_completion=True)
	// Store the evaluated constants into h.constants
	for sortName, vals := range sortValues {
		if len(vals) == 0 {
			continue
		}
		var evaluated []smt.Z3Expr
		for _, v := range vals {
			// Re-evaluate each collected numeral in the model
			ev, ok := model.Eval(v, true)
			if ok {
				evaluated = append(evaluated, ev)
			} else {
				evaluated = append(evaluated, v)
			}
		}
		if sort, ok := interpSorts[sortName]; ok {
			h.constants[sortName] = evaluated
			h.sortMap[sortName] = sort
		}
	}
}

// constantFromZ3 converts a Z3 value back to an Ivy constant (as *lg.Const).
// Used by SortUniverse, Check, getModelConstant where callers need .Name/.CSort.
// Corresponds to Python's constant_from_z3.
func constantFromZ3(sort Sort, z3val smt.Z3Expr) *Const {
	xtracer.Trace("ivy_solver.py:997 constant_from_z3() ENTER sort=%v", sort)
	s := z3val.String()
	if s == "true" {
		return NewConst("true", Boolean)
	}
	if s == "false" {
		return NewConst("false", Boolean)
	}
	return NewConst(s, sort)
}

// constantFromZ3Expr converts a Z3 value back to an Ivy expression,
// returning lg.True/lg.False for boolean values (matching Python's
// constant_from_z3 which returns ivy_logic.And()/ivy_logic.Or()).
// Used by EvalToConstant where callers compare with truth.Equal(lg.True).
func constantFromZ3Expr(sort Sort, z3val smt.Z3Expr) Expr {
	s := z3val.String()
	if s == "true" {
		return True // &LogicAnd{} — matches Python's ivy_logic.And()
	}
	if s == "false" {
		return False // &LogicOr{} — matches Python's ivy_logic.Or()
	}
	return NewConst(s, sort)
}

// --- Model extraction to clauses ---

// ModelUniverseFacts extracts universe membership facts from a model for a sort.
// Returns a list of clause-like formulas: X=c0 | X=c1 | ... and ci != cj.
// Corresponds to Python's model_universe_facts.
func ModelUniverseFacts(h *HerbrandModel, sort Sort, upclose bool) []Expr {
	xtracer.Trace("ivy_solver.py:1432 model_universe_facts() ENTER sort=%v", sort)
	if IsInterpretedSort(h.sig, sort) {
		return nil
	}
	elems := h.SortUniverse(sort)
	if len(elems) == 0 {
		return nil
	}

	var result []Expr

	// Universe closure: forall X. X=c0 | X=c1 | ...
	if !upclose {
		x, _ := NewVariable("X", sort)
		eqs := make([]Expr, len(elems))
		for i, c := range elems {
			eqs[i] = &Eq{T1: x, T2: c}
		}
		result = append(result, &LogicOr{Terms: eqs})
	}

	// Distinctness: ci != cj for all i < j
	for i := 0; i < len(elems); i++ {
		for j := i + 1; j < len(elems); j++ {
			neq := &LogicNot{Body: &Eq{T1: elems[i], T2: elems[j]}}
			result = append(result, neq)
		}
	}

	return result
}

// ModelFacts extracts all facts from a Herbrand model.
// Returns a Clauses set characterizing the model.
// Corresponds to Python's model_facts.
func ModelFacts(h *HerbrandModel, ignore func(*Const) bool, clauses *Clauses, upclose bool) *Clauses {
	xtracer.Trace("ivy_solver.py:1448 model_facts() ENTER")
	if ignore == nil {
		ignore = func(*Const) bool { return false }
	}

	var fmlas []Expr

	// Universe facts for each sort
	for _, sort := range h.Sorts() {
		fmlas = append(fmlas, ModelUniverseFacts(h, sort, upclose)...)
	}

	// Constant values
	symSet := clauses.Symbols()
	for _, symExpr := range symSet.All() {
		sym, ok := symExpr.(*Const)
		if !ok {
			continue
		}
		if ignore(sym) {
			continue
		}
		if h.sig != nil {
			if _, isCtor := h.sig.Constructors[sym.Name]; isCtor {
				continue
			}
		}
		// Only handle 0-arity constants
		if !IsFunctionSort(sym.CSort) {
			val := h.getModelConstant(sym)
			if val != nil {
				fmlas = append(fmlas, &Eq{T1: sym, T2: val})
			}
		}
	}

	// Relation values
	for _, symExpr := range symSet.All() {
		sym, ok := symExpr.(*Const)
		if !ok {
			continue
		}
		if ignore(sym) {
			continue
		}
		if IsRelationalSort(sym.CSort) && IsFunctionSort(sym.CSort) {
			arity := 0
			if fs, ok := sym.CSort.(*LogicFunctionSort); ok {
				arity = fs.Arity()
			}
			if arity > 0 {
				fmlas = append(fmlas, RelationModelToClauses(h, sym, arity)...)
			}
		}
	}

	// Function values
	for _, symExpr := range symSet.All() {
		sym, ok := symExpr.(*Const)
		if !ok {
			continue
		}
		if ignore(sym) {
			continue
		}
		if IsFunctionSort(sym.CSort) && !IsRelationalSort(sym.CSort) {
			if fs, ok := sym.CSort.(*LogicFunctionSort); ok && fs.Arity() >= 1 {
				fmlas = append(fmlas, FunctionModelToClauses(h, sym)...)
			}
		}
	}

	return NewClauses(fmlas, nil, nil)
}

// RelationModelToClauses extracts the relation interpretation from the model.
// Returns formulas for both positive and negative ground instances.
// Corresponds to Python's relation_model_to_clauses.
func RelationModelToClauses(h *HerbrandModel, rel *Const, arity int) []Expr {
	xtracer.Trace("ivy_solver.py:1665 relation_model_to_clauses() ENTER")
	// Create a literal for the relation applied to fresh variables
	fs, ok := rel.CSort.(*LogicFunctionSort)
	if !ok {
		return nil
	}
	dom := fs.Domain()
	vars := make([]Expr, len(dom))
	varPtrs := make([]*LogicVariable, len(dom))
	for i, s := range dom {
		v, _ := NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
		varPtrs[i] = v
	}
	app := MustApply(rel, vars...)

	var result []Expr

	// Positive instances
	posLit := NewLiteral(1, app)
	result = append(result, getLitFacts(h, posLit)...)

	// Negative instances
	negLit := NewLiteral(0, app)
	result = append(result, getLitFacts(h, negLit)...)

	return result
}

// FunctionModelToClauses extracts the function interpretation from the model.
// Corresponds to Python's function_model_to_clauses.
func FunctionModelToClauses(h *HerbrandModel, f *Const) []Expr {
	xtracer.Trace("ivy_solver.py:1686 function_model_to_clauses() ENTER")
	fs, ok := f.CSort.(*LogicFunctionSort)
	if !ok {
		return nil
	}
	rng := fs.Range()
	dom := fs.Domain()

	// Create term f(V0, V1, ...)
	vars := make([]Expr, len(dom))
	for i, s := range dom {
		v, _ := NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
	}
	fTerm := MustApply(f, vars...)

	// For enumerated range, check each possible value.
	// Python (ivy_solver.py:1800): if isinstance(rng, EnumeratedSort) and not use_z3_enums:
	if es, ok := rng.(*LogicEnumeratedSort); ok && (h.tr.s == nil || !h.tr.s.opts.UseZ3Enums) {
		var result []Expr
		for _, name := range es.Extension {
			c := NewConst(name, rng)
			eqLit := NewLiteral(1, &Eq{T1: fTerm, T2: c})
			result = append(result, getLitFacts(h, eqLit)...)
		}
		return result
	}

	// General: create Y = f(V0, ...) literal
	y, _ := NewVariable("Y", rng)
	eqLit := NewLiteral(1, &Eq{T1: y, T2: fTerm})
	return getLitFacts(h, eqLit)
}

// getLitFacts returns ground instances of a literal that hold in the model.
// Corresponds to Python's get_lit_facts.
func getLitFacts(h *HerbrandModel, lit *LogicLiteral) []Expr {
	xtracer.Trace("ivy_solver.py:1676 get_lit_facts() ENTER")
	vs, rows := h.Check(lit)
	var result []Expr
	for _, row := range rows {
		// Build substitution
		subs := make(map[NodeKey]Expr, len(vs))
		for j, v := range vs {
			subs[Key(v)] = row[j]
		}
		// Apply substitution to the literal
		newAtom, err := Substitute(lit.Atom, subs)
		if err != nil {
			continue
		}
		if lit.Polarity == 0 {
			result = append(result, &LogicNot{Body: newAtom})
		} else {
			result = append(result, newAtom)
		}
	}
	return result
}

// Universes returns a map from sort name to universe elements.
// When numerals is true, uninterpreted sorts get renamed to numeric
// indices (0, 1, 2, ...) while interpreted sorts keep their actual values.
// When numerals is false, elements are skolemized (prefixed with "__").
//
// Corresponds to Python HerbrandModel.universes (ivy_solver.py:857-863).
func (h *HerbrandModel) Universes(numerals bool) map[string][]Expr {
	result := make(map[string][]Expr)
	for _, sort := range h.Sorts() {
		sortName := IvySortName(sort)
		if numerals {
			if !IsInterpretedSort(h.sig, sort) {
				elems := h.SortedSortUniverse(sort)
				renamed := make([]Expr, len(elems))
				for i, c := range elems {
					renamed[i] = NewConst(fmt.Sprintf("%d", i), c.CSort)
				}
				result[sortName] = renamed
			} else {
				elems := h.SortUniverse(sort)
				nodes := make([]Expr, len(elems))
				for i, c := range elems {
					nodes[i] = c
				}
				result[sortName] = nodes
			}
		} else {
			elems := h.SortUniverse(sort)
			nodes := make([]Expr, len(elems))
			for i, c := range elems {
				// Python: c.skolem() adds "__" prefix
				nodes[i] = NewConst("__"+c.Name, c.CSort)
			}
			result[sortName] = nodes
		}
	}
	return result
}

// --- Additional solver utility functions ---

// SortCard returns the cardinality of a sort, or -1 if unknown.
// Handles EnumeratedSort, BV sorts, RangeSort, and interpreted sorts.
// Corresponds to Python's sort_card (ivy_solver.py:357-367).
func SortCard(sort Sort, sig *Sig) int {
	xtracer.Trace("ivy_solver.py:383 sort_card() ENTER sort=%v", sort)
	if es, ok := sort.(*LogicEnumeratedSort); ok {
		return es.Card()
	}
	if rs, ok := sort.(*RangeSort); ok {
		lo, hi, ok := RangeSortBounds(rs)
		if ok {
			return hi - lo + 1
		}
	}
	if sig != nil {
		sortName := IvySortName(sort)
		if itp, ok := sig.Interp[sortName]; ok {
			// Check if interpretation is a RangeSort with numeric bounds.
			// Python: itp = ivy_logic.sig.interp.get(sort.name,None)
			//         if isinstance(itp, ivy_logic.RangeSort) and is_numeral(itp.ub)
			if rs, isRS := itp.(*RangeSort); isRS {
				lo, hi, ok := RangeSortBounds(rs)
				if ok {
					return hi - lo + 1
				}
			}
			// Check for BV interpretation (bv, strbv, intbv all map to BitVec)
			if s, isStr := itp.(string); isStr {
				base, params, ok := ParseIntParams(s)
				if ok && len(params) > 0 {
					switch base {
					case "bv", "strbv", "intbv":
						return 1 << uint(params[0])
					}
				}
			}
		}
	}
	return -1
}

// EnumeratedRange returns Z3 expressions for all elements of an enumerated sort.
// Corresponds to Python's enumerated_range.
func (s *Solver) EnumeratedRange(sort *LogicEnumeratedSort) ([]smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:903 enumerated_range() ENTER")
	var result []smt.Z3Expr
	for _, name := range sort.Extension {
		c := NewConst(name, sort)
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
func (s *Solver) GetModelFromClauses(clauses *Clauses) (*HerbrandModel, error) {
	z3solver := s.tr.Ctx.NewZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	result := s.checkZ3(z3solver)
	if result == smt.Unsat {
		return nil, nil // unsatisfiable
	}

	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}

	// Collect vocabulary
	symSet := clauses.Symbols()
	vocab := make([]*Const, 0, symSet.Len())
	for _, sym := range symSet.All() {
		if c, ok := sym.(*Const); ok {
			vocab = append(vocab, c)
		}
	}

	return NewHerbrandModel(s, z3solver, m, vocab), nil
}

// ClausesCase drops literals in a clause set while maintaining satisfiability.
// This only works for quantifier-free clauses.
//
// Corresponds to Python clauses_case (ivy_solver.py lines 1031-1058):
//  1. Check SAT, get model
//  2. Model-simplify each CNF clause, remove duplicates
//  3. Loop: run UnitRes propagation, model-simplify, dedup, until convergence
func (s *Solver) ClausesCase(clauses *Clauses) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:1128 clauses_case() ENTER")
	// Check satisfiability
	z3solver := s.tr.Ctx.NewZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	if s.checkZ3(z3solver) == smt.Unsat {
		// Python returns [[]] (false clauses) on UNSAT, not None.
		return Z3FalseClauses(), nil
	}

	model := z3solver.Model()
	if model == nil {
		return clauses, nil
	}

	// Initial model simplification over CNF (includes defs via ToOpenFormula).
	// Python: clauses = Clauses([clause_model_simp(m,c) for c in clauses1.clauses])
	// Python: clauses = remove_duplicates_clauses(clauses)
	cnf := FormulaToClausesAux(clauses.ToOpenFormula())
	var initFmlas []Expr
	for _, c := range cnf {
		f := ClauseToFormula(c)
		simplified := s.clauseModelSimp(model, f)
		initFmlas = append(initFmlas, simplified)
	}
	initFmlas = removeDuplicateFormulas(initFmlas)
	currentClauses := NewClauses(initFmlas, nil, clauses.Annot)

	// Iterative UnitRes + model simplification loop.
	// Python:
	//   while True:
	//     num_old_clauses = len(clauses.clauses)
	//     r = ur.UnitRes(clauses.clauses)
	//     with r.context(): r.propagate()
	//     new_clauses = Clauses([[l] for l in r.unit_queue] + r.clauses)
	//     clauses = Clauses([clause_model_simp(m,c) for c in new_clauses.clauses])
	//     clauses = remove_duplicates_clauses(clauses)
	//     if len(clauses.clauses) <= num_old_clauses: return clauses
	for {
		// Get CNF literal-lists for this round
		cnf = FormulaToClausesAux(currentClauses.ToOpenFormula())
		numOldClauses := len(cnf)

		// Convert to unitres format and run propagation
		urClauses, symMap := ivyLitsToUnitResClauses(cnf)
		r := NewUnitRes(urClauses)
		r.Propagate(nil)

		// Extract: [[l] for l in r.unit_queue] + r.clauses
		resultLitClauses := extractUnitResResults(r, symMap)

		// Convert to formulas
		var resultFmlas []Expr
		for _, c := range resultLitClauses {
			resultFmlas = append(resultFmlas, ClauseToFormula(c))
		}
		newClauses := NewClauses(resultFmlas, nil, currentClauses.Annot)

		// Model simplify each formula
		// Python: clauses = Clauses([clause_model_simp(m,c) for c in new_clauses.clauses])
		var simpFmlas []Expr
		for _, f := range newClauses.Fmlas {
			simplified := s.clauseModelSimp(model, f)
			simpFmlas = append(simpFmlas, simplified)
		}
		simpFmlas = removeDuplicateFormulas(simpFmlas)
		currentClauses = NewClauses(simpFmlas, nil, currentClauses.Annot)

		// Convergence: Python checks len(clauses.clauses) <= num_old_clauses
		newCnf := FormulaToClausesAux(currentClauses.ToOpenFormula())
		if len(newCnf) <= numOldClauses {
			return currentClauses, nil
		}
	}
}

// clauseModelSimp simplifies a clause (disjunction) by dropping literals
// that are false in the model, keeping only those that are true.
// For non-ground literals, they are always kept.
//
// Corresponds to Python clause_model_simp (lines 1062-1080).
func (s *Solver) clauseModelSimp(model *smt.Model, clause Expr) Expr {
	xtracer.Trace("ivy_solver.py:1160 clause_model_simp() ENTER")
	or, ok := clause.(*LogicOr)
	if !ok || len(or.Terms) <= 1 {
		return clause
	}

	var kept []Expr
	for _, lit := range or.Terms {
		// Non-ground literals are always kept
		if !IsGroundFormula(lit) {
			kept = append(kept, lit)
			continue
		}

		// Evaluate in the model
		zlit, err := s.tr.Translate(CloseFormula(lit))
		if err != nil {
			kept = append(kept, lit)
			continue
		}
		val, ok := model.Eval(zlit, true)
		if ok && val.IsTrue() {
			// Python: if z3.is_true(v): return [l] — early return with single literal
			return lit
		}
		if !ok || !val.IsFalse() {
			kept = append(kept, lit)
		}
		// If false in model, drop it
	}

	if len(kept) == 0 {
		// All literals were false — return empty disjunction (false)
		return &LogicOr{Terms: nil}
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return &LogicOr{Terms: kept}
}

// removeDuplicateFormulas removes duplicate formulas based on string representation.
func removeDuplicateFormulas(fmlas []Expr) []Expr {
	seen := make(map[string]bool)
	var result []Expr
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
func (s *Solver) clausesCaseLegacy(clauses *Clauses) (*Clauses, error) {
	z3solver := s.tr.Ctx.NewZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	if s.checkZ3(z3solver) == smt.Unsat {
		return nil, nil
	}

	model := z3solver.Model()
	if model == nil {
		return clauses, nil
	}

	var newFmlas []Expr
	for _, f := range clauses.Fmlas {
		if or, ok := f.(*LogicOr); ok && len(or.Terms) > 1 {
			// Pick the first disjunct that is true in the model
			picked := false
			for _, d := range or.Terms {
				zd, err := s.tr.Translate(CloseFormula(d))
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

	defs := make([]*IvyDefinition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return NewClauses(newFmlas, defs, clauses.Annot), nil
}
