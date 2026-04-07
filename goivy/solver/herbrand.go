// Ported to Go from ivy_solver.py.
//
// HerbrandModel and model extraction functions.

package solver

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/unitres"
	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// HerbrandModel wraps a Z3 model and provides operations to extract
// sort universes, evaluate formulas, and produce Ivy-level facts.
// Corresponds to Python's HerbrandModel class. (ivy_solver.py:835)
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

	// Extract sort universes from the Z3 model.
	// Python: self.constants = dict((sort_from_z3(s), model.get_universe(s)) ...)
	for _, z3sort := range model.Sorts() {
		sortName := z3sort.String()
		universe := model.SortUniverse(z3sort)
		h.constants[sortName] = universe
		// Primary: use translator's reverse map (matches Python's sort_from_z3)
		if ivySort, ok := s.tr.SortFromZ3(z3sort); ok {
			h.sortMap[sortName] = ivySort
		} else if ivySort, ok := s.sig.Sorts[sortName]; ok {
			// Fallback to sig lookup by name
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
// Python builds a SortOrder from the Ivy `<` symbol applied to variables,
// translates to Z3, then sorts using substitute + model.eval. This works
// for user-defined `<` on uninterpreted sorts, not just Z3 built-in Lt.
func (h *HerbrandModel) SortedSortUniverse(sort lg.Sort) []*lg.Const {
	name := il.SortName(sort)
	elems, ok := h.constants[name]
	if !ok {
		return nil
	}

	// Try to sort by the `<` relation (matching Python's SortOrder approach)
	if h.model != nil && h.tr != nil && len(elems) > 1 {
		sorted, ok := h.trySortByOrder(sort, elems)
		if ok {
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

// trySortByOrder attempts to sort elements using the Ivy `<` relation
// translated to Z3 and evaluated in the model. This matches Python's
// SortOrder class (ivy_solver.py:787-797) which uses atom_to_z3(order(*vs))
// with substitute + model.eval, supporting user-defined orderings on
// uninterpreted sorts.
// HerbrandModel is in ivy_solver.py:835-914.
func (h *HerbrandModel) trySortByOrder(sort lg.Sort, elems []z3bridge.Expr) (sorted []z3bridge.Expr, ok bool) {
	xtracer.Trace("ivy_solver.py:851 HerbrandModel.sorted_sort_universe() top. solver/herbrand.go:132") // not seen in make golden.

	// Python's approach: (ivy_solver.py:850 in sorted_sort_universe()):
	//   vs = [Variable("X", sort), Variable("Y", sort)]
	//   order = Symbol("<", RelationSort([sort, sort]))
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
	xVar, _ := lg.NewVariable("X", sort)
	yVar, _ := lg.NewVariable("Y", sort)
	orderSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sort, sort}))
	orderApp, err := lg.NewApply(orderSym, xVar, yVar)
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
	lessFunc := func(a, b z3bridge.Expr) bool {
		fact := SubstituteZ3(ctx, orderZ3, [][2]z3bridge.Expr{{z3X, a}, {z3Y, b}})
		val, evalOk := h.model.Eval(fact, true)
		if !evalOk {
			return false
		}
		return val.String() == "true"
	}

	// Insertion sort (matching Python's sorted() with cmp_to_key)
	sorted = make([]z3bridge.Expr, len(elems))
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
func (h *HerbrandModel) SortUniverseZ3(sortName string) []z3bridge.Expr {
	return h.constants[sortName]
}

// Eval evaluates a formula in the model. Variables are interpreted universally.
// Returns true if the formula holds for all variable assignments.
func (h *HerbrandModel) Eval(fmla lg.Expr) bool {
	// Negate the formula and check: if no satisfying assignment, fmla is true
	negLit := il.NewLiteral(0, fmla)
	_, rows := h.Check(negLit)
	return len(rows) == 0
}

// EvalConstant evaluates a constant in the model, returning its model value.
func (h *HerbrandModel) EvalConstant(c *lg.Const) *lg.Const {
	return h.getModelConstant(c)
}

// EvalToConstant evaluates a term in the model, returning an Ivy expression.
// Returns lg.True/lg.False for boolean values (matching Python's constant_from_z3
// which returns ivy_logic.And()/ivy_logic.Or()).
// This matches the trace.Model interface which declares return type lg.Expr.
func (h *HerbrandModel) EvalToConstant(t lg.Expr) lg.Expr {
	zt, err := h.tr.Translate(t)
	if err != nil {
		return lg.NewConst("?", t.NodeSort())
	}
	val, ok := h.model.Eval(zt, true)
	if !ok {
		return lg.NewConst("?", t.NodeSort())
	}
	return constantFromZ3Expr(t.NodeSort(), val)
}

// Check evaluates a literal against all possible variable assignments.
// Returns (vars, rows) where each row is a list of Ivy constants that
// satisfy the literal.
func (h *HerbrandModel) Check(lit *il.Literal) ([]*lg.Variable, [][]*lg.Const) {
	// Get free variables in the literal
	fvMap := lu.FreeVariables(lit)
	var vs []*lg.Variable
	for _, v := range fvMap {
		if vv, ok := v.(*lg.Variable); ok {
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
func (h *HerbrandModel) variableRange(v *lg.Variable) []z3bridge.Expr {
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
// by evaluating symbols in the model and collecting numeral values from
// their interpretations (including ITE branches).
//
// Corresponds to Python's mine_interpreted_constants (ivy_solver.py:809-818).
// Python collects model values for symbols whose range sort IS interpreted,
// building term sym(V0,V1,...), evaluating in the model, and collecting
// numerals from the result.
func (h *HerbrandModel) mineInterpretedConstants(model *z3bridge.Model, vocab []*lg.Const) {
	if h.sig == nil {
		return
	}

	// Build set of interpreted sorts (Python: sorts = ivy_logic.interpreted_sorts())
	interpSorts := make(map[string]lg.Sort)
	for name := range h.sig.Interp {
		if s, ok := h.sig.Sorts[name]; ok {
			interpSorts[name] = s
		}
	}

	// Initialize sort_values for each interpreted sort
	sortValues := make(map[string]map[string]z3bridge.Expr)
	for name := range interpSorts {
		sortValues[name] = make(map[string]z3bridge.Expr)
	}

	// For each symbol in vocab, if its range sort is interpreted,
	// collect model values.
	// Python: for s in vocab:
	//     sort = s.sort.rng
	//     if sort in sort_values:
	//         sort_values[sort].update(collect_model_values(sort, model, s))
	for _, sym := range vocab {
		rng := il.SortRange(sym.CSort)
		rngName := il.SortName(rng)
		if _, isInterp := sortValues[rngName]; !isInterp {
			continue // range sort is not interpreted, skip
		}

		// collect_model_values: build sym(V0,V1,...), translate, eval, collect numerals
		phs := module.SymPlaceholders(sym)
		var term lg.Expr
		if len(phs) == 0 {
			term = sym
		} else {
			args := make([]lg.Expr, len(phs))
			for i, v := range phs {
				args[i] = v
			}
			term = lg.MustApply(sym, args...)
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
		var evaluated []z3bridge.Expr
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

// constantFromZ3Expr converts a Z3 value back to an Ivy expression,
// returning lg.True/lg.False for boolean values (matching Python's
// constant_from_z3 which returns ivy_logic.And()/ivy_logic.Or()).
// Used by EvalToConstant where callers compare with truth.Equal(lg.True).
func constantFromZ3Expr(sort lg.Sort, z3val z3bridge.Expr) lg.Expr {
	s := z3val.String()
	if s == "true" {
		return lg.True // &And{} — matches Python's ivy_logic.And()
	}
	if s == "false" {
		return lg.False // &Or{} — matches Python's ivy_logic.Or()
	}
	return lg.NewConst(s, sort)
}

// --- Model extraction to clauses ---

// ModelUniverseFacts extracts universe membership facts from a model for a sort.
// Returns a list of clause-like formulas: X=c0 | X=c1 | ... and ci != cj.
// Corresponds to Python's model_universe_facts.
func ModelUniverseFacts(h *HerbrandModel, sort lg.Sort, upclose bool) []lg.Expr {
	if il.IsInterpretedSort(h.sig, sort) {
		return nil
	}
	elems := h.SortUniverse(sort)
	if len(elems) == 0 {
		return nil
	}

	var result []lg.Expr

	// Universe closure: forall X. X=c0 | X=c1 | ...
	if !upclose {
		x, _ := lg.NewVariable("X", sort)
		eqs := make([]lg.Expr, len(elems))
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
func ModelFacts(h *HerbrandModel, ignore func(*lg.Const) bool, clauses *module.Clauses, upclose bool) *module.Clauses {
	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	var fmlas []lg.Expr

	// Universe facts for each sort
	for _, sort := range h.Sorts() {
		fmlas = append(fmlas, ModelUniverseFacts(h, sort, upclose)...)
	}

	// Constant values
	symSet := clauses.Symbols()
	for _, sym := range symSet {
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
	for _, sym := range symSet {
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
	for _, sym := range symSet {
		if ignore(sym) {
			continue
		}
		if il.IsFunctionSort(sym.CSort) && !il.IsRelationalSort(sym.CSort) {
			if fs, ok := sym.CSort.(*lg.FunctionSort); ok && fs.Arity() >= 1 {
				fmlas = append(fmlas, FunctionModelToClauses(h, sym)...)
			}
		}
	}

	return module.NewClauses(fmlas, nil, nil)
}

// RelationModelToClauses extracts the relation interpretation from the model.
// Returns formulas for both positive and negative ground instances.
// Corresponds to Python's relation_model_to_clauses.
func RelationModelToClauses(h *HerbrandModel, rel *lg.Const, arity int) []lg.Expr {
	// Create a literal for the relation applied to fresh variables
	fs, ok := rel.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	dom := fs.Domain()
	vars := make([]lg.Expr, len(dom))
	varPtrs := make([]*lg.Variable, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
		varPtrs[i] = v
	}
	app := lg.MustApply(rel, vars...)

	var result []lg.Expr

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
func FunctionModelToClauses(h *HerbrandModel, f *lg.Const) []lg.Expr {
	fs, ok := f.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	rng := fs.Range()
	dom := fs.Domain()

	// Create term f(V0, V1, ...)
	vars := make([]lg.Expr, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
	}
	fTerm := lg.MustApply(f, vars...)

	// For enumerated range, check each possible value
	if es, ok := rng.(*lg.EnumeratedSort); ok {
		var result []lg.Expr
		for _, name := range es.Extension {
			c := lg.NewConst(name, rng)
			eqLit := il.NewLiteral(1, &lg.Eq{T1: fTerm, T2: c})
			result = append(result, getLitFacts(h, eqLit)...)
		}
		return result
	}

	// General: create Y = f(V0, ...) literal
	y, _ := lg.NewVariable("Y", rng)
	eqLit := il.NewLiteral(1, &lg.Eq{T1: y, T2: fTerm})
	return getLitFacts(h, eqLit)
}

// getLitFacts returns ground instances of a literal that hold in the model.
// Corresponds to Python's get_lit_facts.
func getLitFacts(h *HerbrandModel, lit *il.Literal) []lg.Expr {
	vs, rows := h.Check(lit)
	var result []lg.Expr
	for _, row := range rows {
		// Build substitution
		subs := make(map[lg.NodeKey]lg.Expr, len(vs))
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

// Universes returns a map from sort name to universe elements.
// When numerals is true, uninterpreted sorts get renamed to numeric
// indices (0, 1, 2, ...) while interpreted sorts keep their actual values.
// When numerals is false, elements are skolemized (prefixed with "__").
//
// Corresponds to Python HerbrandModel.universes (ivy_solver.py:857-863).
func (h *HerbrandModel) Universes(numerals bool) map[string][]lg.Expr {
	result := make(map[string][]lg.Expr)
	for _, sort := range h.Sorts() {
		sortName := il.SortName(sort)
		if numerals {
			if !il.IsInterpretedSort(h.sig, sort) {
				elems := h.SortedSortUniverse(sort)
				renamed := make([]lg.Expr, len(elems))
				for i, c := range elems {
					renamed[i] = lg.NewConst(fmt.Sprintf("%d", i), c.CSort)
				}
				result[sortName] = renamed
			} else {
				elems := h.SortUniverse(sort)
				nodes := make([]lg.Expr, len(elems))
				for i, c := range elems {
					nodes[i] = c
				}
				result[sortName] = nodes
			}
		} else {
			elems := h.SortUniverse(sort)
			nodes := make([]lg.Expr, len(elems))
			for i, c := range elems {
				// Python: c.skolem() adds "__" prefix
				nodes[i] = lg.NewConst("__"+c.Name, c.CSort)
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
func SortCard(sort lg.Sort, sig *il.Sig) int {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		return es.Card()
	}
	if rs, ok := sort.(*lg.RangeSort); ok {
		lo, hi, ok := RangeSortBounds(rs)
		if ok {
			return hi - lo + 1
		}
	}
	if sig != nil {
		sortName := il.SortName(sort)
		if itp, ok := sig.Interp[sortName]; ok {
			// Check if interpretation is a RangeSort with numeric bounds.
			// Python: itp = ivy_logic.sig.interp.get(sort.name,None)
			//         if isinstance(itp, ivy_logic.RangeSort) and is_numeral(itp.ub)
			if rs, isRS := itp.(*lg.RangeSort); isRS {
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
func (s *Solver) GetModelFromClauses(clauses *module.Clauses) (*HerbrandModel, error) {
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
	for _, sym := range symSet {
		vocab = append(vocab, sym)
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
func (s *Solver) ClausesCase(clauses *module.Clauses) (*module.Clauses, error) {
	// Check satisfiability
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	if z3solver.Check() == z3bridge.Unsat {
		// Python returns [[]] (false clauses) on UNSAT, not None.
		return FalseClauses(), nil
	}

	model := z3solver.Model()
	if model == nil {
		return clauses, nil
	}

	// Initial model simplification over CNF (includes defs via ToOpenFormula).
	// Python: clauses = Clauses([clause_model_simp(m,c) for c in clauses1.clauses])
	// Python: clauses = remove_duplicates_clauses(clauses)
	cnf := module.FormulaToClausesAux(clauses.ToOpenFormula())
	var initFmlas []lg.Expr
	for _, c := range cnf {
		f := module.ClauseToFormula(c)
		simplified := s.clauseModelSimp(model, f)
		initFmlas = append(initFmlas, simplified)
	}
	initFmlas = removeDuplicateFormulas(initFmlas)
	currentClauses := module.NewClauses(initFmlas, nil, clauses.Annot)

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
		cnf = module.FormulaToClausesAux(currentClauses.ToOpenFormula())
		numOldClauses := len(cnf)

		// Convert to unitres format and run propagation
		urClauses, symMap := ivyLitsToUnitResClauses(cnf)
		r := unitres.NewUnitRes(urClauses)
		r.Propagate(nil)

		// Extract: [[l] for l in r.unit_queue] + r.clauses
		resultLitClauses := extractUnitResResults(r, symMap)

		// Convert to formulas
		var resultFmlas []lg.Expr
		for _, c := range resultLitClauses {
			resultFmlas = append(resultFmlas, module.ClauseToFormula(c))
		}
		newClauses := module.NewClauses(resultFmlas, nil, currentClauses.Annot)

		// Model simplify each formula
		// Python: clauses = Clauses([clause_model_simp(m,c) for c in new_clauses.clauses])
		var simpFmlas []lg.Expr
		for _, f := range newClauses.Fmlas {
			simplified := s.clauseModelSimp(model, f)
			simpFmlas = append(simpFmlas, simplified)
		}
		simpFmlas = removeDuplicateFormulas(simpFmlas)
		currentClauses = module.NewClauses(simpFmlas, nil, currentClauses.Annot)

		// Convergence: Python checks len(clauses.clauses) <= num_old_clauses
		newCnf := module.FormulaToClausesAux(currentClauses.ToOpenFormula())
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
func (s *Solver) clauseModelSimp(model *z3bridge.Model, clause lg.Expr) lg.Expr {
	or, ok := clause.(*lg.Or)
	if !ok || len(or.Terms) <= 1 {
		return clause
	}

	var kept []lg.Expr
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
		return &lg.Or{Terms: nil}
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return &lg.Or{Terms: kept}
}

// removeDuplicateFormulas removes duplicate formulas based on string representation.
func removeDuplicateFormulas(fmlas []lg.Expr) []lg.Expr {
	seen := make(map[string]bool)
	var result []lg.Expr
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
func (s *Solver) clausesCaseLegacy(clauses *module.Clauses) (*module.Clauses, error) {
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

	var newFmlas []lg.Expr
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
	return module.NewClauses(newFmlas, defs, clauses.Annot), nil
}
