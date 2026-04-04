// Ported to Go from ivy_solver.py.

// Package solver provides high-level SMT solver operations for Ivy verification.
// It wraps the z3bridge package to provide formula and clause-level operations
// such as satisfiability checks, implication checks, UNSAT core extraction,
// model generation, and small-model search.
package solver

import (
	"fmt"
	"sync"

	mod "github.com/glycerine/ivy/goivy/module"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// --- Options ---

// Options controls solver behavior.
type Options struct {
	Seed        int
	Incremental bool
	MacroFinder bool
	ShowVCs     bool
	UseZ3Enums  bool
}

// DefaultOptions returns the default solver options.
func DefaultOptions() *Options {
	return &Options{
		Seed:        0,
		Incremental: true,
		MacroFinder: true,
		ShowVCs:     false,
		UseZ3Enums:  true,
	}
}

// --- Solver ---

// Solver wraps a z3bridge.Translator and manages caches for
// converting Ivy logic structures to Z3 and back.
type Solver struct {
	mu               sync.Mutex
	tr               *z3bridge.Translator
	opts             *Options
	sig              *il.Sig
	HandleRangeSorts bool // controls range sort clamped arithmetic; default true
}

// New creates a new Solver with default options and a fresh Z3 context.
func New() *Solver {
	s := &Solver{
		tr:               z3bridge.NewTranslator(),
		opts:             DefaultOptions(),
		sig:              il.NewSig(),
		HandleRangeSorts: true,
	}
	s.wireNativeLookup()
	return s
}

// NewWithSig creates a new Solver using the given signature.
func NewWithSig(sig *il.Sig) *Solver {
	s := &Solver{
		tr:               z3bridge.NewTranslator(),
		opts:             DefaultOptions(),
		sig:              sig,
		HandleRangeSorts: true,
	}
	s.wireNativeLookup()
	return s
}

// NewWithOptions creates a new Solver with custom options.
func NewWithOptions(sig *il.Sig, opts *Options) *Solver {
	if opts == nil {
		opts = DefaultOptions()
	}
	s := &Solver{
		tr:               z3bridge.NewTranslator(),
		opts:             opts,
		sig:              sig,
		HandleRangeSorts: true,
	}
	s.wireNativeLookup()
	return s
}

func (s *Solver) Close() error {
	return s.tr.Close()
}

// Clear resets all Z3 caches (sorts, constants, functions) to initial state.
// Corresponds to Python ivy_solver.clear() (line 228).
func (s *Solver) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tr.Clear()
}

// wireNativeLookup installs the NativeLookup callback on the translator
// so that polymorphic symbols (+, -, *, /), range sort clamped arithmetic,
// and native interpretations (nat, bv, etc.) are properly handled during
// Z3 translation.
func (s *Solver) wireNativeLookup() {
	s.tr.NativeLookup = func(name string, sort lg.Sort, isRelation bool) func(args ...z3bridge.Expr) z3bridge.Expr {
		sym := lg.NewConst(name, sort)
		nf := s.LookupNative(sym, isRelation)
		if nf == nil {
			return nil
		}
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			return nf(args...)
		}
	}
	// Install SortLookup to handle interpreted sorts (nat→IntSort, bv[N]→BitVecSort, etc.)
	// Corresponds to Python ivy_solver.py sorts() function.
	s.tr.SortLookup = func(sortName string) *z3bridge.Sort {
		if s.sig == nil {
			return nil
		}
		itp, ok := s.sig.Interp[sortName]
		if !ok {
			return nil
		}
		ctx := s.tr.Ctx
		switch v := itp.(type) {
		case string:
			switch v {
			case "nat", "int":
				zs := ctx.IntSort()
				return &zs
			case "real":
				zs := ctx.RealSort()
				return &zs
			case "strlit":
				zs := ctx.StringSort()
				return &zs
			default:
				// Check for bv[N], strbv[N], intbv[N]
				// Python: bv → BitVecSort, strbv → BitVecSort, intbv → BitVecSort
				base, params, ok := ParseIntParams(v)
				if ok && len(params) > 0 {
					switch base {
					case "bv", "strbv", "intbv":
						zs := ctx.BvSort(params[0])
						return &zs
					}
				}
			}
		case *lg.RangeSort:
			// Range sorts map to integers
			zs := ctx.IntSort()
			return &zs
		}
		return nil
	}

	// Install SolverName so Z3 names match Python's naming convention
	// (e.g., polymorphic "<" becomes "<:int:int" at int sort).
	s.tr.SolverName = func(name string, sort lg.Sort) string {
		sym := lg.NewConst(name, sort)
		return s.SolverName(sym)
	}

	// Install QuantConstraints so ForAll/Exists over nat/range-sorted
	// variables include bounds constraints in the quantifier body.
	// Corresponds to Python's quant_constraints (ivy_solver.py:509-519).
	s.tr.QuantConstraints = func(v *lg.Variable, z3Var z3bridge.Expr) []z3bridge.Expr {
		if s.sig == nil {
			return nil
		}
		sortName := il.SortName(v.VSort)
		itp, ok := s.sig.Interp[sortName]
		if !ok {
			return nil
		}
		ctx := s.tr.Ctx
		switch itpVal := itp.(type) {
		case string:
			if itpVal == "nat" {
				// nat: 0 <= z3_v
				return []z3bridge.Expr{ctx.Le(ctx.IntVal(0), z3Var)}
			}
		case *lg.RangeSort:
			if s.HandleRangeSorts {
				lb, ub, err := s.RangeSortBoundsToZ3(itpVal)
				if err == nil {
					// lb <= z3_v and z3_v <= ub
					return []z3bridge.Expr{
						ctx.Le(lb, z3Var),
						ctx.Le(z3Var, ub),
					}
				}
			}
		}
		return nil
	}

	// Install EqFunc so equality uses MyEq (True/False optimization).
	// Corresponds to Python's my_eq (ivy_solver.py:88-95).
	s.tr.EqFunc = func(x, y z3bridge.Expr) z3bridge.Expr {
		return MyEq(s.tr.Ctx, x, y)
	}

	// Install EnumEqFunc so enumerated sort equality uses binary encoding
	// when UseZ3Enums is false.
	// Corresponds to Python atom_to_z3 line 484 and formula_to_z3_int line 596.
	s.tr.EnumEqFunc = func(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (*z3bridge.Expr, error) {
		if !s.opts.UseZ3Enums {
			result, err := s.EncodeEqualityZ3(t1, t2, sort)
			if err != nil {
				return nil, err
			}
			return &result, nil
		}
		return nil, nil // use default Z3 enum equality
	}

	// Install NumeralFunc so numerals with range sorts get clamped.
	// Corresponds to Python term_to_z3 lines 439-440 + numeral_to_z3.
	// NumeralToZ3 creates Z3 values directly (IntVal/BvVal/StringVal)
	// matching Python's approach — no recursion through Translate().
	s.tr.NumeralFunc = func(name string, sort lg.Sort) (*z3bridge.Expr, error) {
		num := lg.NewConst(name, sort)
		result, err := s.NumeralToZ3(num)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
}

// Translator returns the underlying z3bridge.Translator.
func (s *Solver) Translator() *z3bridge.Translator {
	return s.tr
}

// Context returns the underlying Z3 context.
func (s *Solver) Context() *z3bridge.Z3Context {
	return s.tr.Ctx
}

// Sig returns the current signature.
func (s *Solver) Sig() *il.Sig {
	return s.sig
}

// SetSig sets the signature for this solver.
func (s *Solver) SetSig(sig *il.Sig) {
	s.sig = sig
}

// --- Formula translation ---

// FormulaToZ3 converts a single Ivy formula to a Z3 expression.
// Free variables are universally quantified.
// This corresponds to Python's formula_to_z3.
func (s *Solver) FormulaToZ3(fmla lg.Expr) (z3bridge.Expr, error) {
	return s.tr.Translate(fmla)
}

// ClausesToZ3 converts a Clauses set to a single Z3 expression (conjunction).
// This corresponds to Python's clauses_to_z3.
// After translating formulas and definitions, it also appends type_constraints
// for nat sorts (non-negativity) and range sorts (bounds), matching Python's
// clauses_to_z3 which calls type_constraints(used_symbols_clauses(clauses)).
func (s *Solver) ClausesToZ3(clauses *mod.Clauses) (z3bridge.Expr, error) {
	if clauses == nil {
		xtracer.Trace("solver.ClausesToZ3 ENTER nil")
		return s.tr.Ctx.BoolVal(true), nil
	}
	xtracer.Trace("solver.ClausesToZ3 ENTER\n fmlas=%d defs=%d", len(clauses.Fmlas), len(clauses.Defs))

	var exprs []z3bridge.Expr

	// Translate each formula
	for i, f := range clauses.Fmlas {
		xtracer.Trace("solver.ClausesToZ3 fmla[%d] sort=%v", i, f.NodeSort())
		//pp("type=%T val=%v", f, f)
		zf, err := s.translateClosed(f)
		if err != nil {
			return z3bridge.Expr{}, fmt.Errorf("translating formula: %w", err)
		}
		exprs = append(exprs, zf)
	}

	// Translate definitions as constraints
	for _, d := range clauses.Defs {
		constraint := defToConstraint(d)
		zd, err := s.translateClosed(constraint)
		if err != nil {
			return z3bridge.Expr{}, fmt.Errorf("translating definition: %w", err)
		}
		exprs = append(exprs, zd)
	}

	// Add type constraints for used symbols (nat non-negativity, range bounds)
	// This corresponds to Python: type_constraints(used_symbols_clauses(clauses))
	usedSyms := clauses.Symbols()
	for _, sym := range usedSyms {
		constraints := s.typeConstraintsForSymbol(sym)
		for _, tc := range constraints {
			ztc, err := s.translateClosed(tc)
			if err != nil {
				continue // skip constraints we can't translate
			}
			exprs = append(exprs, ztc)
		}
	}

	xtracer.Trace("solver.ClausesToZ3 EXIT exprs=%d", len(exprs))
	if len(exprs) == 0 {
		return s.tr.Ctx.BoolVal(true), nil
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return s.tr.Ctx.And(exprs...), nil
}

// typeConstraintsForSymbol generates type constraints for a symbol based on
// its sort's interpretation. For nat sorts: ¬(x < 0). For range sorts:
// ¬(x < lb) ∧ ¬(ub < x). Corresponds to Python's type_constraints.
func (s *Solver) typeConstraintsForSymbol(sym *lg.Const) []lg.Expr {
	if s.sig == nil {
		return nil
	}

	// Get the range sort of the symbol
	rng := il.SortRange(sym.CSort)
	if rng == nil {
		return nil
	}
	rngName := il.SortName(rng)

	// Check if the sort has a nat interpretation
	interp, hasInterp := s.sig.Interp[rngName]
	if !hasInterp {
		return nil
	}

	// Skip interpreted symbols
	if il.IsInterpretedSymbol(s.sig, sym) {
		return nil
	}

	// Build the term for the symbol (applying to variables if function sort)
	var term lg.Expr = sym
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
		dom := fs.Domain()
		args := make([]lg.Expr, len(dom))
		for i, ds := range dom {
			v, _ := lg.NewVariable(fmt.Sprintf("X%d", i), ds)
			args[i] = v
		}
		app, err := lg.NewApply(sym, args...)
		if err != nil {
			return nil
		}
		term = app
	}

	var constraints []lg.Expr

	// Check for nat interpretation (string "nat")
	if interpStr, ok := interp.(string); ok && interpStr == "nat" {
		// Non-negativity: ¬(term < 0)
		zero := lg.NewConst("0", rng)
		ltSort := il.RelationSort([]lg.Sort{rng, rng})
		lt := lg.NewConst("<", ltSort)
		ltApp, err := lg.NewApply(lt, term, zero)
		if err == nil {
			constraints = append(constraints, &lg.Not{Body: ltApp})
		}
	}

	// Check for range sort interpretation
	if rs, ok := interp.(*lg.RangeSort); ok {
		lb := lg.NewConst(rs.LbString(), rng)
		ub := lg.NewConst(rs.UbString(), rng)
		// Lower bound: ¬(term < lb)
		ltSort := il.RelationSort([]lg.Sort{rng, rng})
		lt := lg.NewConst("<", ltSort)
		ltLbApp, err := lg.NewApply(lt, term, lb)
		if err == nil {
			constraints = append(constraints, &lg.Not{Body: ltLbApp})
		}
		// Upper bound: ¬(ub < term)
		ltUbApp, err := lg.NewApply(lt, ub, term)
		if err == nil {
			constraints = append(constraints, &lg.Not{Body: ltUbApp})
		}
	}

	return constraints
}

// translateClosed converts a formula to Z3, universally quantifying free variables.
func (s *Solver) translateClosed(fmla lg.Expr) (x z3bridge.Expr, err error) {
	closed := il.CloseFormula(fmla)
	defer func() {
		r := recover()
		if r != nil {
			vv("warning: recover from panic on il.CloseFormula(): '%v'", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	x, err = s.tr.Translate(closed)
	return
}

// NotClausesToZ3 negates a Clauses and converts to Z3.
// Corresponds to Python's not_clauses_to_z3.
func (s *Solver) NotClausesToZ3(clauses *mod.Clauses) (z3bridge.Expr, error) {
	// Separate Skolem definitions from other definitions
	var skolemDefs, otherDefs []*il.Definition
	for _, d := range clauses.Defs {
		sym := d.Defines()
		if c, ok := sym.(*lg.Const); ok && isSkolem(c.Name) {
			skolemDefs = append(skolemDefs, d)
		} else {
			otherDefs = append(otherDefs, d)
		}
	}

	// Skolem definitions are asserted positively
	skolemClauses := mod.NewClauses(nil, skolemDefs, nil)
	zSkolem, err := s.ClausesToZ3(skolemClauses)
	if err != nil {
		return z3bridge.Expr{}, err
	}

	// Full clauses are negated
	zAll, err := s.ClausesToZ3(clauses)
	if err != nil {
		return z3bridge.Expr{}, err
	}

	_ = otherDefs
	return s.tr.Ctx.And(zSkolem, s.tr.Ctx.Not(zAll)), nil
}

// defToConstraint converts a Definition to a constraint formula.
func defToConstraint(d *il.Definition) lg.Expr {
	lhs := d.Lhs
	rhs := d.Rhs
	var constraint lg.Expr
	if lg.SortEqual(rhs.NodeSort(), lg.Boolean) {
		constraint = &lg.Iff{T1: lhs, T2: rhs}
	} else {
		constraint = &lg.Eq{T1: lhs, T2: rhs}
	}
	return constraint
}

func isSkolem(name string) bool {
	return len(name) >= 2 && name[0] == '_' && name[1] == '_'
}

// --- Satisfiability checks ---

// IsSat checks whether a formula is satisfiable.
// Returns true if satisfiable, false if unsatisfiable.
func (s *Solver) IsSat(fmla lg.Expr) (bool, error) {
	result, err := s.tr.IsSat(fmla)
	if err != nil {
		return false, err
	}
	return result == z3bridge.Sat, nil
}

// Implies checks whether fmla1 implies fmla2.
// Returns true if fmla1 => fmla2 is valid.
func (s *Solver) Implies(fmla1, fmla2 lg.Expr) (bool, error) {
	return s.tr.Implies(fmla1, fmla2)
}

// ClausesSat checks whether a Clauses set is satisfiable.
// Corresponds to Python's clauses_sat.
func (s *Solver) ClausesSat(clauses *mod.Clauses) (bool, error) {
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return false, err
	}
	z3solver.Assert(zc)
	result := z3solver.Check()
	return result != z3bridge.Unsat, nil
}

// ClausesImply checks whether clauses1 imply clauses2.
// Corresponds to Python's clauses_imply.
func (s *Solver) ClausesImply(clauses1, clauses2 *mod.Clauses) (bool, error) {
	z3solver := s.tr.Ctx.NewSolver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z1)

	z2, err := s.NotClausesToZ3(clauses2)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z2)

	result := z3solver.Check()
	return result == z3bridge.Unsat, nil
}

// ImpliesBatch tests if premise implies each formula in fmlas.
// More efficient than calling Implies repeatedly: reuses a single solver
// with push/pop for each check.
// Corresponds to Python's z3_implies_batch.
func (s *Solver) ImpliesBatch(premise lg.Expr, fmlas []lg.Expr) ([]bool, error) {
	z3solver := s.tr.Ctx.NewSolver()
	zPremise, err := s.translateClosed(premise)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zPremise)

	result := make([]bool, len(fmlas))
	for i, f := range fmlas {
		negF := &lg.Not{Body: f}
		zNeg, err := s.translateClosed(negF)
		if err != nil {
			return nil, err
		}
		z3solver.Push()
		z3solver.Assert(zNeg)
		res := z3solver.Check()
		z3solver.Pop()
		result[i] = (res == z3bridge.Unsat)
	}
	return result, nil
}

// ClausesImplyFormula checks whether clauses1 imply fmla2.
// Corresponds to Python's clauses_imply_formula.
func (s *Solver) ClausesImplyFormula(clauses1 *mod.Clauses, fmla2 lg.Expr) (bool, error) {
	z3solver := s.tr.Ctx.NewSolver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z1)

	negFmla := &lg.Not{Body: fmla2}
	z2, err := s.translateClosed(negFmla)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z2)

	result := z3solver.Check()
	return result == z3bridge.Unsat, nil
}

// --- UNSAT Core ---

// UnsatCore extracts an unsatisfiable core from clauses1 with respect to clauses2.
// If the combined system is satisfiable, returns nil.
// The optional 'implies' parameter adds an implication constraint.
// The 'unlikely' function marks formulas that should preferably be excluded.
// Corresponds to Python's unsat_core.
func (s *Solver) UnsatCore(
	clauses1, clauses2 *mod.Clauses,
	implies *mod.Clauses,
	unlikely func(lg.Expr) bool,
) (*mod.Clauses, error) {
	if unlikely == nil {
		unlikely = func(lg.Expr) bool { return false }
	}

	fmlas := clauses1.Fmlas
	z3solver := s.tr.Ctx.NewSolver()

	// Create activation literals
	alits := make([]z3bridge.Expr, len(fmlas))
	for i := range fmlas {
		name := fmt.Sprintf("__c%d", i)
		alits[i] = s.tr.Ctx.Const(name, s.tr.Ctx.BoolSort())
	}

	// Assert: activation literal => formula
	for i, f := range fmlas {
		zf, err := s.translateClosed(f)
		if err != nil {
			return nil, err
		}
		clause := s.tr.Ctx.Or(s.tr.Ctx.Not(alits[i]), zf)
		z3solver.Assert(clause)
	}

	// Assert definitions from clauses1
	for _, d := range clauses1.Defs {
		constraint := defToConstraint(d)
		zd, err := s.translateClosed(constraint)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zd)
	}

	// Assert clauses2
	z2, err := s.ClausesToZ3(clauses2)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(z2)

	// Assert implies (negated)
	if implies != nil {
		zi, err := s.NotClausesToZ3(implies)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zi)
	}

	// Use Z3's native assumption-based checking + unsat_core() API.
	// This is more efficient than manual push/pop minimization.
	// Python: ivy_solver.py:696-710 uses check(assumptions) + unsat_core()
	result := z3solver.CheckAssumptions(alits)

	if result == z3bridge.Sat {
		return nil, nil // satisfiable, no core
	}

	// Get unsat core from Z3
	coreExprs := z3solver.UnsatCore()

	// Build a set of core activation literal names for lookup
	coreSet := make(map[string]bool, len(coreExprs))
	for _, ce := range coreExprs {
		coreSet[ce.String()] = true
	}

	// Collect formulas whose activation literals are in the core
	var resFmlas []lg.Expr
	for i, f := range fmlas {
		if coreSet[alits[i].String()] {
			resFmlas = append(resFmlas, f)
		}
	}

	// If the core is empty but the system was unsat, include all formulas
	// (this can happen when the unsat-ness comes from clauses2/implies)
	if len(resFmlas) == 0 && result == z3bridge.Unsat {
		resFmlas = append(resFmlas, fmlas...)
	}

	// Minimize core using biased_core approach: try removing unlikely formulas first,
	// then remaining. This produces a smaller core than Z3's native unsat_core.
	if len(resFmlas) > 1 {
		resFmlas = minimizeCore(z3solver, resFmlas, alits, fmlas, unlikely)
	}

	defs := make([]*il.Definition, len(clauses1.Defs))
	copy(defs, clauses1.Defs)
	return mod.NewClauses(resFmlas, defs, nil), nil
}

// minimizeCore performs biased core minimization on a set of formulas.
// It tries removing unlikely formulas first, then remaining ones.
// Python: ivy_core.py minimize_core / biased_core
func minimizeCore(
	z3solver *z3bridge.Solver,
	resFmlas []lg.Expr,
	alits []z3bridge.Expr,
	allFmlas []lg.Expr,
	unlikely func(lg.Expr) bool,
) []lg.Expr {
	// Build index from formula key to activation literal
	fmlaToAlit := make(map[string]z3bridge.Expr)
	for i, f := range allFmlas {
		fmlaToAlit[fmt.Sprint(f)] = alits[i]
	}

	core := make([]bool, len(resFmlas))
	for i := range core {
		core[i] = true
	}

	// Get activation literals for core formulas
	coreAlits := make([]z3bridge.Expr, len(resFmlas))
	for i, f := range resFmlas {
		key := fmt.Sprint(f)
		if a, ok := fmlaToAlit[key]; ok {
			coreAlits[i] = a
		}
	}

	// Try removing unlikely formulas first
	for i, f := range resFmlas {
		if !unlikely(f) || coreAlits[i].String() == "" {
			continue
		}
		core[i] = false
		assumptions := collectAssumptions(coreAlits, core)
		if z3solver.CheckAssumptions(assumptions) == z3bridge.Unsat {
			// Still unsat, can keep it removed
		} else {
			core[i] = true
		}
	}

	// Try removing remaining formulas
	for i := range resFmlas {
		if !core[i] || coreAlits[i].String() == "" {
			continue
		}
		core[i] = false
		assumptions := collectAssumptions(coreAlits, core)
		if z3solver.CheckAssumptions(assumptions) == z3bridge.Unsat {
			// Still unsat, can keep it removed
		} else {
			core[i] = true
		}
	}

	var result []lg.Expr
	for i, f := range resFmlas {
		if core[i] {
			result = append(result, f)
		}
	}
	return result
}

// collectAssumptions builds the list of activation literals for included formulas.
func collectAssumptions(alits []z3bridge.Expr, included []bool) []z3bridge.Expr {
	var result []z3bridge.Expr
	for i, a := range alits {
		if included[i] && a.String() != "" {
			result = append(result, a)
		}
	}
	return result
}

// --- Size constraints ---

// SortSizeConstraint generates a constraint limiting a sort's universe to at most 'size' elements.
// For uninterpreted sorts: exists constants c0..c_{size-1} such that forall X, X=c0 | X=c1 | ...
// Corresponds to Python's sort_size_constraint.
func SortSizeConstraint(sort lg.Sort, size int) lg.Expr {
	us, ok := sort.(*lg.UninterpretedSort)
	if !ok {
		return lg.True // trivially true for non-uninterpreted sorts
	}

	syms := make([]*lg.Const, size)
	eqs := make([]lg.Expr, size)
	v, _ := lg.NewVariable("X"+us.Name, sort)
	for i := 0; i < size; i++ {
		syms[i] = lg.NewConst(fmt.Sprintf("__%s$%d", us.Name, i), sort)
		eqs[i] = &lg.Eq{T1: v, T2: syms[i]}
	}
	return &lg.Or{Terms: eqs}
}

// RelationSizeConstraint generates a constraint limiting a relation to at most 'size' true entries.
// Corresponds to Python's relation_size_constraint.
func RelationSizeConstraint(relation *lg.Const, size int) lg.Expr {
	fs, ok := relation.CSort.(*lg.FunctionSort)
	if !ok {
		return lg.True
	}

	domain := fs.Domain()
	// Create size-many tuples of constants
	consts := make([][]*lg.Const, size)
	for i := 0; i < size; i++ {
		consts[i] = make([]*lg.Const, len(domain))
		for j, s := range domain {
			consts[i][j] = lg.NewConst(
				fmt.Sprintf("__$%s$%d$%d", relation.Name, i, j), s)
		}
	}

	// Create variables for the universal quantifier
	vs := make([]lg.Expr, len(domain))
	for j, s := range domain {
		v, _ := lg.NewVariable(fmt.Sprintf("X$%s$%d", relation.Name, j), s)
		vs[j] = v
	}

	// Build: ~relation(X0,...) | (X0=c00 & X1=c01 &...) | (X0=c10 & X1=c11 &...) | ...
	negApp := &lg.Not{Body: &lg.Apply{Func: relation, Terms: vs}}

	disjuncts := []lg.Expr{negApp}
	for i := 0; i < size; i++ {
		conjuncts := make([]lg.Expr, len(domain))
		for j := range domain {
			conjuncts[j] = &lg.Eq{T1: consts[i][j], T2: vs[j]}
		}
		disjuncts = append(disjuncts, &lg.And{Terms: conjuncts})
	}
	return &lg.Or{Terms: disjuncts}
}

// SizeConstraint generates a size constraint for either a sort or a relation.
func SizeConstraint(x lg.Expr, size int) lg.Expr {
	if us, ok := x.(*lg.UninterpretedSort); ok {
		return SortSizeConstraint(us, size)
	}
	if c, ok := x.(*lg.Const); ok {
		if _, isFS := c.CSort.(*lg.FunctionSort); isFS {
			return RelationSizeConstraint(c, size)
		}
	}
	return lg.True
}

// --- Assume/Assert sequences ---

// AssumeAssert represents either an assumption or an assertion in a sequence.
type AssumeAssert struct {
	Clauses  *mod.Clauses
	Doc      string
	IsAssert bool
}

// NewAssume creates an Assume entry.
func NewAssume(clauses *mod.Clauses, doc string) AssumeAssert {
	return AssumeAssert{Clauses: clauses, Doc: doc, IsAssert: false}
}

// NewAssert creates an Assert entry.
func NewAssert(clauses *mod.Clauses, doc string) AssumeAssert {
	return AssumeAssert{Clauses: clauses, Doc: doc, IsAssert: true}
}

// Reporter is an interface for reporting progress during CheckSequence.
// Corresponds to Python's reporter parameter in check_sequence.
type Reporter interface {
	// Start is called before each assume/assert is checked.
	// Returns false to abort the sequence.
	Start(isAssert bool, doc string) bool
	// End is called after each assume/assert is checked.
	// result is true if the check passed. Returns false to abort.
	End(result bool, doc string) bool
}

// CheckSequence checks a sequence of assumes and asserts.
// Returns a boolean slice indicating which checks passed.
// Corresponds to Python's check_sequence.
func (s *Solver) CheckSequence(seq []AssumeAssert) ([]bool, error) {
	return s.CheckSequenceWithReporter(seq, nil)
}

// CheckSequenceWithReporter is like CheckSequence but accepts an optional
// Reporter for progress reporting. Corresponds to Python's check_sequence
// (ivy_solver.py:977-1005).
//
// Python control flow:
//   - reporter.start() is called for both Assume and Assert (return ignored)
//   - reporter.end() is called ONLY for Assert items
//   - s.pop() happens AFTER reporter.end()
//   - reporter.end() returning False causes early return
func (s *Solver) CheckSequenceWithReporter(seq []AssumeAssert, reporter Reporter) ([]bool, error) {
	z3solver := s.tr.Ctx.NewSolver()
	var results []bool // Python uses list append

	for _, aa := range seq {
		if !aa.IsAssert {
			// Assume
			if reporter != nil {
				reporter.Start(false, aa.Doc) // Python ignores return
			}
			z1, err := s.ClausesToZ3(aa.Clauses)
			if err != nil {
				return nil, err
			}
			z3solver.Assert(z1)
			results = append(results, true)
		} else {
			// Assert
			if reporter != nil {
				reporter.Start(true, aa.Doc) // Python ignores return
			}
			dual := mod.NegateClauses(aa.Clauses)
			z2, err := s.ClausesToZ3(dual)
			if err != nil {
				return nil, err
			}
			z3solver.Push()
			z3solver.Assert(z2)
			checkResult := z3solver.Check() == z3bridge.Unsat
			results = append(results, checkResult)
			if reporter != nil {
				if !reporter.End(checkResult, aa.Doc) {
					return results, nil // Python: return res (early)
				}
			}
			z3solver.Pop() // Python: s.pop() AFTER reporter.end()
		}
	}
	return results, nil
}
