// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_solver.py.

// Package solver provides high-level SMT solver operations for Ivy verification.
// It wraps the z3bridge package to provide formula and clause-level operations
// such as satisfiability checks, implication checks, UNSAT core extraction,
// model generation, and small-model search.
package solver

import (
	"fmt"
	"sync"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// --- Options ---

// Options controls solver behavior.
type Options struct {
	Seed         int
	Incremental  bool
	MacroFinder  bool
	ShowVCs      bool
	UseZ3Enums   bool
}

// DefaultOptions returns the default solver options.
func DefaultOptions() *Options {
	return &Options{
		Seed:         0,
		Incremental:  true,
		MacroFinder:  true,
		ShowVCs:      false,
		UseZ3Enums:   true,
	}
}

// --- Solver ---

// Solver wraps a z3bridge.Translator and manages caches for
// converting Ivy logic structures to Z3 and back.
type Solver struct {
	mu   sync.Mutex
	tr   *z3bridge.Translator
	opts *Options
	sig  *il.Sig
}

// New creates a new Solver with default options and a fresh Z3 context.
func New() *Solver {
	s := &Solver{
		tr:   z3bridge.NewTranslator(),
		opts: DefaultOptions(),
		sig:  il.NewSig(),
	}
	s.wireNativeLookup()
	return s
}

// NewWithSig creates a new Solver using the given signature.
func NewWithSig(sig *il.Sig) *Solver {
	s := &Solver{
		tr:   z3bridge.NewTranslator(),
		opts: DefaultOptions(),
		sig:  sig,
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
		tr:   z3bridge.NewTranslator(),
		opts: opts,
		sig:  sig,
	}
	s.wireNativeLookup()
	return s
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
}

// Translator returns the underlying z3bridge.Translator.
func (s *Solver) Translator() *z3bridge.Translator {
	return s.tr
}

// Context returns the underlying Z3 context.
func (s *Solver) Context() *z3bridge.Context {
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
func (s *Solver) FormulaToZ3(fmla lg.Node) (z3bridge.Expr, error) {
	return s.tr.Translate(fmla)
}

// ClausesToZ3 converts a Clauses set to a single Z3 expression (conjunction).
// This corresponds to Python's clauses_to_z3.
// After translating formulas and definitions, it also appends type_constraints
// for nat sorts (non-negativity) and range sorts (bounds), matching Python's
// clauses_to_z3 which calls type_constraints(used_symbols_clauses(clauses)).
func (s *Solver) ClausesToZ3(clauses *clauseops.Clauses) (z3bridge.Expr, error) {
	if clauses == nil {
		return s.tr.Ctx.BoolVal(true), nil
	}

	var exprs []z3bridge.Expr

	// Translate each formula
	for _, f := range clauses.Fmlas {
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
	for _, symN := range usedSyms { sym := symN.(*lg.Const)
		constraints := s.typeConstraintsForSymbol(sym)
		for _, tc := range constraints {
			ztc, err := s.translateClosed(tc)
			if err != nil {
				continue // skip constraints we can't translate
			}
			exprs = append(exprs, ztc)
		}
	}

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
func (s *Solver) typeConstraintsForSymbol(sym *lg.Const) []lg.Node {
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

	interpStr, isStr := interp.(string)
	if !isStr {
		return nil
	}

	// Build the term for the symbol (applying to variables if function sort)
	var term lg.Node = sym
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
		dom := fs.Domain()
		args := make([]lg.Node, len(dom))
		for i, ds := range dom {
			v, _ := lg.NewVar(fmt.Sprintf("X%d", i), ds)
			args[i] = v
		}
		app, err := lg.NewApply(sym, args...)
		if err != nil {
			return nil
		}
		term = app
	}

	var constraints []lg.Node

	if interpStr == "nat" {
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
		lb := lg.NewConst(rs.Lb, rng)
		ub := lg.NewConst(rs.Ub, rng)
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
func (s *Solver) translateClosed(fmla lg.Node) (z3bridge.Expr, error) {
	closed := il.CloseFormula(fmla)
	return s.tr.Translate(closed)
}

// NotClausesToZ3 negates a Clauses and converts to Z3.
// Corresponds to Python's not_clauses_to_z3.
func (s *Solver) NotClausesToZ3(clauses *clauseops.Clauses) (z3bridge.Expr, error) {
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
	skolemClauses := clauseops.NewClauses(nil, skolemDefs, nil)
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
func defToConstraint(d *il.Definition) lg.Node {
	lhs := d.Lhs
	rhs := d.Rhs
	var constraint lg.Node
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
func (s *Solver) IsSat(fmla lg.Node) (bool, error) {
	result, err := s.tr.IsSat(fmla)
	if err != nil {
		return false, err
	}
	return result == z3bridge.Sat, nil
}

// Implies checks whether fmla1 implies fmla2.
// Returns true if fmla1 => fmla2 is valid.
func (s *Solver) Implies(fmla1, fmla2 lg.Node) (bool, error) {
	return s.tr.Implies(fmla1, fmla2)
}

// ClausesSat checks whether a Clauses set is satisfiable.
// Corresponds to Python's clauses_sat.
func (s *Solver) ClausesSat(clauses *clauseops.Clauses) (bool, error) {
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
func (s *Solver) ClausesImply(clauses1, clauses2 *clauseops.Clauses) (bool, error) {
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
func (s *Solver) ImpliesBatch(premise lg.Node, fmlas []lg.Node) ([]bool, error) {
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
func (s *Solver) ClausesImplyFormula(clauses1 *clauseops.Clauses, fmla2 lg.Node) (bool, error) {
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
	clauses1, clauses2 *clauseops.Clauses,
	implies *clauseops.Clauses,
	unlikely func(lg.Node) bool,
) (*clauseops.Clauses, error) {
	if unlikely == nil {
		unlikely = func(lg.Node) bool { return false }
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
	var resFmlas []lg.Node
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
	return clauseops.NewClauses(resFmlas, defs, nil), nil
}

// minimizeCore performs biased core minimization on a set of formulas.
// It tries removing unlikely formulas first, then remaining ones.
// Python: ivy_core.py minimize_core / biased_core
func minimizeCore(
	z3solver *z3bridge.Solver,
	resFmlas []lg.Node,
	alits []z3bridge.Expr,
	allFmlas []lg.Node,
	unlikely func(lg.Node) bool,
) []lg.Node {
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

	var result []lg.Node
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
func SortSizeConstraint(sort lg.Sort, size int) lg.Node {
	us, ok := sort.(*lg.UninterpretedSort)
	if !ok {
		return lg.True // trivially true for non-uninterpreted sorts
	}

	syms := make([]*lg.Const, size)
	eqs := make([]lg.Node, size)
	v, _ := lg.NewVar("X"+us.Name, sort)
	for i := 0; i < size; i++ {
		syms[i] = lg.NewConst(fmt.Sprintf("__%s$%d", us.Name, i), sort)
		eqs[i] = &lg.Eq{T1: v, T2: syms[i]}
	}
	return &lg.Or{Terms: eqs}
}

// RelationSizeConstraint generates a constraint limiting a relation to at most 'size' true entries.
// Corresponds to Python's relation_size_constraint.
func RelationSizeConstraint(relation *lg.Const, size int) lg.Node {
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
	vs := make([]lg.Node, len(domain))
	for j, s := range domain {
		v, _ := lg.NewVar(fmt.Sprintf("X$%s$%d", relation.Name, j), s)
		vs[j] = v
	}

	// Build: ~relation(X0,...) | (X0=c00 & X1=c01 &...) | (X0=c10 & X1=c11 &...) | ...
	negApp := &lg.Not{Body: &lg.Apply{Func: relation, Terms: vs}}

	disjuncts := []lg.Node{negApp}
	for i := 0; i < size; i++ {
		conjuncts := make([]lg.Node, len(domain))
		for j := range domain {
			conjuncts[j] = &lg.Eq{T1: consts[i][j], T2: vs[j]}
		}
		disjuncts = append(disjuncts, &lg.And{Terms: conjuncts})
	}
	return &lg.Or{Terms: disjuncts}
}

// SizeConstraint generates a size constraint for either a sort or a relation.
func SizeConstraint(x lg.Node, size int) lg.Node {
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
	Clauses  *clauseops.Clauses
	Doc      string
	IsAssert bool
}

// NewAssume creates an Assume entry.
func NewAssume(clauses *clauseops.Clauses, doc string) AssumeAssert {
	return AssumeAssert{Clauses: clauses, Doc: doc, IsAssert: false}
}

// NewAssert creates an Assert entry.
func NewAssert(clauses *clauseops.Clauses, doc string) AssumeAssert {
	return AssumeAssert{Clauses: clauses, Doc: doc, IsAssert: true}
}

// CheckSequence checks a sequence of assumes and asserts.
// Returns a boolean slice indicating which checks passed.
// Corresponds to Python's check_sequence.
func (s *Solver) CheckSequence(seq []AssumeAssert) ([]bool, error) {
	z3solver := s.tr.Ctx.NewSolver()
	results := make([]bool, len(seq))

	for i, aa := range seq {
		if !aa.IsAssert {
			// Assume: add to solver
			z1, err := s.ClausesToZ3(aa.Clauses)
			if err != nil {
				return nil, err
			}
			z3solver.Assert(z1)
			results[i] = true
		} else {
			// Assert: check negation
			dual := clauseops.NegateClauses(aa.Clauses)
			z2, err := s.ClausesToZ3(dual)
			if err != nil {
				return nil, err
			}
			z3solver.Push()
			z3solver.Assert(z2)
			results[i] = z3solver.Check() == z3bridge.Unsat
			z3solver.Pop()
		}
	}
	return results, nil
}
