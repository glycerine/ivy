// Ported to Go from ivy_solver.py.

// Package solver provides high-level SMT solver operations for Ivy verification.
// It is the higher level API within the z3bridge package to provide
// formula and clause-level operations
// such as satisfiability checks, implication checks, UNSAT core extraction,
// model generation, and small-model search.
package goivy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Solver ---

// Solver wraps a Translator and manages caches for
// converting Ivy logic structures to Z3 and back.
type Solver struct {
	mu               sync.Mutex
	tr               *Translator
	z3u              *Z3Utils // z3_utils.py operations (ToZ3, Z3Implies, Z3ImpliesBatch)
	opts             *SolverOptions
	sig              *Sig
	mod              *Module // module this solver belongs to (nil = ad-hoc/test)
	HandleRangeSorts bool    // controls range sort clamped arithmetic; default true
}

// NewSolver creates a new Solver bound to the given module. The module's
// Z3SessionCache (lazily created on first use) is reused so that all
// NewSolver calls against the same module share Z3 sorts, constants,
// functions, and predicates — matching Python's "z3.Solver() instances
// within a Module context share z3_sorts/etc." semantics
// (ivy_solver.py:252-260, ivy_module.py:96-109).
//
// Pass nil for mod to get an ad-hoc per-Solver cache (tests, low-level
// usage that has no module). Pass nil for opts to get default options.
//
// Sig is derived from mod.Sig when mod is non-nil; pass nil mod to get
// an empty signature.
func NewSolver(mod *Module, opts *SolverOptions) *Solver {
	if opts == nil {
		opts = DefaultSolverOptions()
	}
	var sig *Sig
	if mod != nil {
		sig = mod.Sig
	}
	if sig == nil {
		sig = NewSig()
	}
	s := &Solver{
		z3u:              NewZ3Utils(),
		opts:             opts,
		sig:              sig,
		mod:              mod,
		HandleRangeSorts: true,
	}
	cache := getOrCreateModuleCache(mod)
	s.tr = s.NewTranslatorWithCache(cache)
	//s.wireNativeLookup()
	return s
}

// NewSolverFromSig is a test convenience: wraps the given sig in a
// throwaway module and returns a Solver bound to it. Each call creates
// a fresh module, so caches are not shared between calls. Production
// code should use NewSolver(mod, opts) directly so that solver calls
// against the same module share Z3 state.
func NewSolverFromSig(sig *Sig, opts *SolverOptions) *Solver {
	var mod *Module
	if sig != nil {
		mod = NewWithSig(sig)
	}
	return NewSolver(mod, opts)
}

// getOrCreateModuleCache returns mod's Z3SessionCache, lazily creating
// and attaching it on first access. For mod == nil, returns a fresh
// per-Solver cache (no sharing — preserves legacy/test behavior).
func getOrCreateModuleCache(mod *Module) *Z3SessionCache {
	if mod == nil {
		return NewZ3SessionCache()
	}
	if mod.z3SessionCache != nil {
		return mod.z3SessionCache
	}
	if mod.z3SharedCtx == nil {
		mod.z3SharedCtx = &z3CtxHolder{}
	}
	// Create new cache: reuse shared Z3Context if available (preserves
	// z3CheckCounter across module copies — Python's _z3_check_counter
	// is a process global that persists), otherwise create fresh.
	var cache *Z3SessionCache
	if mod.z3SharedCtx.ctx != nil {
		cache = NewZ3SessionCacheWithCtx(mod.z3SharedCtx.ctx)
	}
	if cache == nil {
		cache = NewZ3SessionCache()
	}
	cache.z3CheckCounter = &mod.z3SharedCtx.z3CheckCounter
	mod.z3SessionCache = cache
	// Ensure z3SharedCtx is set for future copies of this module.
	if mod.z3SharedCtx.ctx == nil {
		mod.z3SharedCtx.ctx = cache.Ctx
	}
	return cache
}

// newZ3Solver creates a Solver and applies opts
// (e.g., smt.macro_finder). Always sets the parameter
// explicitly — never assumes Z3's default matches ours.
func (s *Solver) newZ3Solver() *smt.Z3Solver {
	zs := s.tr.Ctx.NewZ3Solver()
	s.applyZ3SolverOptions(zs)
	return zs
}

func (s *Solver) applyZ3SolverOptions(zs *smt.Z3Solver) {
	if s == nil || s.opts == nil {
		return
	}
	for key, value := range solverOptionParamValues(s.opts) {
		zs.SetParam(key, value)
	}
}

func solverOptionParamValues(opts *SolverOptions) map[string]string {
	if opts == nil {
		return nil
	}
	params := map[string]string{
		"smt.macro_finder": strconv.FormatBool(opts.MacroFinder),
	}
	if opts.SeedSet || opts.Seed != 0 {
		params["smt.random_seed"] = strconv.Itoa(opts.Seed)
	}
	return params
}

func (s *Solver) Close() error {
	err1 := s.tr.Close()
	err2 := s.z3u.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Clear resets all Z3 caches (sorts, constants, functions) to initial state.
// Corresponds to Python ivy_solver.clear() (ivy_solver.py:249).
//
// Python calls clear() in two places:
//  1. ivy_solver.py:257 — at module import time (startup), once
//  2. ivy_module.py:102 — inside Module.__enter__(), every time a module
//     context is entered
//
// What clear() does is reset 4 module-level global dicts — z3_sorts,
// z3_predicates, z3_constants, z3_functions — which are translation caches
// mapping Ivy sorts/symbols to Z3 objects. They accumulate as formulas
// get translated. When entering a new module (different signature), the
// old Z3 translations could be stale, so Python wipes them.
//
// In Go, the equivalent state lives on a *Z3SessionCache attached to Module.
// Module.Enter() calls cache.Clear() to reset that state.
//
// Solver.Clear() here resets the cache associated with this Solver's
// Translator. If the cache is shared with other Solvers (because they
// belong to the same module), they will all see the cleared state.
func (s *Solver) Clear() {
	//xtracer.Trace("ivy_solver.py:245 clear() ENTER")
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tr.Clear()
	s.z3u.Clear()
}

/* deprecated, go directly now.
// wireNativeLookup installs the LookupNative callback on the translator
// so that polymorphic symbols (+, -, *, /), range sort clamped arithmetic,
// native interpretations (nat, bv, etc.), and interpreted sorts are properly
// handled during Z3 translation.
// Corresponds to Python's lookup_native(thing, table, kind) dispatch.
func (s *Solver) wireNativeLookup() {
	s.tr.LookupNative = func(name string, sort lg.Sort, kind string) any {
		sym := lg.NewConst(name, sort)
		var table func(string) any
		switch kind {
		case "sort":
			table = s.Sorts
		case "relation":
			table = s.Relations
		case "function":
			table = s.Functions
		}
		result := s.LookupNative(sym, table, kind)
		// A pre-merge of solver/ and z3bridge/ packages comment:
		// Convert NativeFunc (named type in solver package) to the anonymous
		// function type that z3bridge low-level code can type-assert against.
		if nf, ok := result.(NativeFunc); ok {
			return func(args ...Expr) Expr {
				return nf(args...)
			}
		}
		return result
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
	s.tr.QuantConstraints = func(v *lg.Variable, z3Var Expr) []Expr {
		if s.sig == nil {
			return nil
		}
		sortName := il.IvySortName(v.VSort)
		itp, ok := s.sig.Interp[sortName]
		if !ok {
			return nil
		}
		ctx := s.tr.Ctx
		switch itpVal := itp.(type) {
		case string:
			if itpVal == "nat" {
				// nat: 0 <= z3_v
				return []Expr{ctx.Le(ctx.IntVal(0), z3Var)}
			}
		case *lg.RangeSort:
			if s.HandleRangeSorts {
				lb, ub, err := s.RangeSortBoundsToZ3(itpVal)
				if err == nil {
					// lb <= z3_v and z3_v <= ub
					return []Expr{
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
	s.tr.EqFunc = func(x, y Expr) Expr {
		return MyEq(s.tr.Ctx, x, y)
	}

	// Install EnumEqFunc so enumerated sort equality uses binary encoding
	// when UseZ3Enums is false.
	// Corresponds to Python atom_to_z3 line 484 and formula_to_z3_int line 596.
	s.tr.EnumEqFunc = func(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (*Expr, error) {
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
	s.tr.NumeralFunc = func(name string, sort lg.Sort) (*Expr, error) {
		num := lg.NewConst(name, sort)
		result, err := s.NumeralToZ3(num)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
}
*/

// Translator returns the underlying Translator.
func (s *Solver) Translator() *Translator {
	return s.tr
}

// Context returns the underlying Z3 context.
func (s *Solver) Context() *smt.Z3Context {
	return s.tr.Ctx
}

// Sig returns the current signature.
func (s *Solver) Sig() *Sig {
	return s.sig
}

// SetSig sets the signature for this solver.
func (s *Solver) SetSig(sig *Sig) {
	s.sig = sig
}

// --- Formula translation ---

// FormulaToZ3 translates a single Ivy formula to Z3 via the core translator.
// This is a raw translate (no HASH, no closing, no type constraints).
// Used by tests and external callers (vmt, alpha).
// For the full Python formula_to_z3 equivalent, use formulaToZ3 (private).
func (s *Solver) FormulaToZ3(fmla Expr) (smt.Z3Expr, error) {
	return s.tr.Translate(fmla)
}

// ClausesToZ3 converts a Clauses set to a single Z3 expression (conjunction).
// This corresponds to Python's clauses_to_z3.
// After translating formulas and definitions, it also appends type_constraints
// for nat sorts (non-negativity) and range sorts (bounds), matching Python's
// clauses_to_z3 which calls type_constraints(used_symbols_clauses(clauses)).
func (s *Solver) ClausesToZ3(clauses *Clauses) (smt.Z3Expr, error) {
	if clauses == nil {
		xtracer.Trace("solver.ClausesToZ3 ENTER nil")
		return s.tr.Ctx.BoolVal(true), nil
	}
	xtracer.Trace("solver.ClausesToZ3 ENTER fmlas=%d defs=%d", len(clauses.Fmlas), len(clauses.Defs))

	var exprs []smt.Z3Expr

	// Match Python clauses_to_z3: a `for` loop emits the per-fmla sort traces,
	// THEN a separate list comprehension calls conj_to_z3 on each formula.
	// We must mirror that two-pass structure so the xtraces interleave the
	// same way as Python.
	for i, f := range clauses.Fmlas {
		xtracer.Trace("solver.ClausesToZ3 fmla[%d] sort=%v", i, f.NodeSort())
	}
	// Translate formulas via conjToZ3 matching Python: [conj_to_z3(cl) for cl in clauses.fmlas]
	for _, f := range clauses.Fmlas {
		zf, err := s.conjToZ3(f)
		if err != nil {
			return smt.Z3Expr{}, fmt.Errorf("translating formula: %w", err)
		}
		exprs = append(exprs, zf)
	}

	// Translate definitions via formulaToZ3 matching Python: formula_to_z3(dfn)
	for di, d := range clauses.Defs {
		zd, err := s.formulaToZ3(d)
		if err != nil {
			defName := "?"
			if sym := d.Defines(); sym != nil {
				if c, ok := sym.(*Const); ok {
					defName = c.Name
				}
			}
			xtracer.Trace("clauses_to_z3: Z3 error on def[%d]: %v defines=%s", di, err, defName)
			return smt.Z3Expr{}, fmt.Errorf("translating definition: %w", err)
		}
		exprs = append(exprs, zd)
	}

	// Python clauses_to_z3 line 645: z3_clauses.extend(type_constraints(used_symbols_clauses(clauses)))
	// type_constraints sorts the incoming symbols before generating bounds.
	allSyms := make(map[NodeKey]Expr)
	for _, f := range clauses.Fmlas {
		for k, v := range UsedSymbolsAst(f).All() {
			allSyms[k] = v
		}
	}
	for _, d := range clauses.Defs {
		for k, v := range UsedSymbolsAst(d).All() {
			allSyms[k] = v
		}
	}
	clauseSyms := make([]Expr, 0, len(allSyms))
	for _, sym := range allSyms {
		clauseSyms = append(clauseSyms, sym)
	}
	sort.Slice(clauseSyms, func(i, j int) bool {
		return Key(clauseSyms[i]) < Key(clauseSyms[j])
	})
	tcs, tcErr := s.typeConstraints(clauseSyms)
	if tcErr != nil {
		return smt.Z3Expr{}, tcErr
	}
	exprs = append(exprs, tcs...)

	xtracer.Trace("solver.ClausesToZ3 EXIT exprs=%d", len(exprs))
	// Mirror Python ivy_solver.py:681 `res = z3.And(z3_clauses)`. Python
	// wraps unconditionally — even with 0 or 1 clauses. We must do the
	// same: a 1-clause unwrap shortcut would change the Z3 AST shape and
	// break the z3.check canon hash comparison.
	return s.tr.Ctx.AndApp(exprs...), nil
}

// buildConstraintTerm builds a term for type constraints.
// For non-function symbols: returns the symbol itself.
// For function symbols: returns sym applied to fresh variables X0, X1, ...
// Matches Python type_constraints lines 610-611 / 621-622.
func (s *Solver) buildConstraintTerm(sym Expr) Expr {
	var term Expr = sym
	if fs, ok := sym.NodeSort().(*LogicFunctionSort); ok {
		dom := fs.Domain()
		args := make([]Expr, len(dom))
		for i, ds := range dom {
			v, _ := NewVariable(fmt.Sprintf("X%d", i), ds)
			args[i] = v
		}
		app, err := NewApply(sym, args...)
		if err != nil {
			return nil
		}
		term = app
	}
	return term
}

// natConstraintForSymbol returns a ¬(term < 0) constraint if sym has
// a nat interpretation, else nil.
// Matches Python type_constraints lines 605-615.
func (s *Solver) natConstraintForSymbol(sym Expr) []Expr {
	if s.sig == nil {
		return nil
	}
	rng := SortRange(sym.NodeSort())
	if rng == nil {
		return nil
	}
	interp, ok := s.sig.Interp[IvySortName(rng)]
	if !ok {
		return nil
	}
	interpStr, isStr := interp.(string)
	if !isStr || interpStr != "nat" {
		return nil
	}
	if IsInterpretedSymbol(s.sig, sym) {
		return nil
	}
	term := s.buildConstraintTerm(sym)
	if term == nil {
		return nil
	}
	// ¬(term < 0)
	zero := NewConst("0", rng)
	ltSort := LogicRelationSort([]Sort{rng, rng})
	lt := NewConst("<", ltSort)
	ltApp, err := NewApply(lt, term, zero)
	if err != nil {
		return nil
	}
	return []Expr{&LogicNot{Body: ltApp}}
}

// rangeConstraintsForSymbol returns ¬(term < lb) and ¬(ub < term)
// constraints if sym has a RangeSort interpretation, else nil.
// Matches Python type_constraints lines 618-629.
func (s *Solver) rangeConstraintsForSymbol(sym Expr) []Expr {
	if s.sig == nil {
		return nil
	}
	rng := SortRange(sym.NodeSort())
	if rng == nil {
		return nil
	}
	interp, ok := s.sig.Interp[IvySortName(rng)]
	if !ok {
		return nil
	}
	rs, isRS := interp.(*RangeSort)
	if !isRS {
		return nil
	}
	if IsInterpretedSymbol(s.sig, sym) {
		return nil
	}
	term := s.buildConstraintTerm(sym)
	if term == nil {
		return nil
	}
	var constraints []Expr
	lb := NewConst(rs.LbString(), rng)
	ub := NewConst(rs.UbString(), rng)
	ltSort := LogicRelationSort([]Sort{rng, rng})
	lt := NewConst("<", ltSort)
	// Lower bound: ¬(term < lb)
	ltLbApp, err := NewApply(lt, term, lb)
	if err == nil {
		constraints = append(constraints, &LogicNot{Body: ltLbApp})
	}
	// Upper bound: ¬(ub < term)
	ltUbApp, err := NewApply(lt, ub, term)
	if err == nil {
		constraints = append(constraints, &LogicNot{Body: ltUbApp})
	}
	return constraints
}

// typeConstraints generates type constraints for nat and range sorts,
// translating each to Z3 via formulaToZ3Closed.
// Matches Python type_constraints (ivy_solver.py:603-631).
func (s *Solver) typeConstraints(syms []Expr) ([]smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:603 type_constraints() ENTER nsyms=%d", len(syms))

	var res []smt.Z3Expr

	// Pass 1: nat sort constraints (Python lines 605-615)
	for _, sym := range syms {
		for _, tc := range s.natConstraintForSymbol(sym) {
			ztc, err := s.formulaToZ3Closed(tc)
			if err != nil {
				continue
			}
			res = append(res, ztc)
		}
	}

	// Pass 2: range sort constraints (Python lines 616-630)
	// Python sets handle_range_sorts = False before translating range constraints,
	// to avoid double-clamping (the constraint IS the bound, don't clamp again).
	saved := s.HandleRangeSorts
	s.HandleRangeSorts = false
	for _, sym := range syms {
		for _, tc := range s.rangeConstraintsForSymbol(sym) {
			ztc, err := s.formulaToZ3Closed(tc)
			if err != nil {
				continue
			}
			res = append(res, ztc)
		}
	}
	s.HandleRangeSorts = saved

	return res, nil
}

// formulaToZ3 translates a formula to Z3 with HASH trace and type constraints.
// Matches Python's formula_to_z3 (ivy_solver.py:659-676).
//
// Call chain: formulaToZ3 → formulaToZ3Closed → Translate (no HASH)
// Only this function emits the HASH trace, matching Python.
func (s *Solver) formulaToZ3(fmla Expr) (x smt.Z3Expr, err error) {
	defer func() {
		r := recover()
		if r != nil {
			vv("warning: recover from panic on formulaToZ3: '%v'", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	// Emit HASH trace matching Python formula_to_z3 line 660-663
	if xtracer.Enabled {
		canon := Canonical(fmla.Sexp())
		leaf, root := s.tr.TranslateMerkle.AddLeaf(canon)
		_, _ = leaf, root
		//xtracer.Trace("ivy_solver.py:718 formula_to_z3 HASH leaf=%s root=%s canon=%s", leaf, root, string(canon)) // 1 of 2
	}

	z3Fmla, err := s.formulaToZ3Closed(fmla)
	if err != nil {
		xtracer.Trace("formula_to_z3: Z3 error on formula_to_z3_closed: %v type=%v", err, ShortTypeName(fmla))
		return smt.Z3Expr{}, err
	}

	// Python formula_to_z3 line 725: tcs = type_constraints(used_symbols_ast(fmla))
	// type_constraints sorts the incoming symbols before generating bounds.
	usedSymsMap := UsedSymbolsAst(fmla)
	usedSyms := make([]Expr, 0, usedSymsMap.Len())
	for _, sym := range usedSymsMap.All() {
		usedSyms = append(usedSyms, sym)
	}
	sort.Slice(usedSyms, func(i, j int) bool {
		return Key(usedSyms[i]) < Key(usedSyms[j])
	})
	tcs, tcErr := s.typeConstraints(usedSyms)
	if tcErr != nil {
		xtracer.Trace("formula_to_z3: Z3 error on type_constraints: %v type=%v", tcErr, ShortTypeName(fmla))
		return smt.Z3Expr{}, tcErr
	}
	if len(tcs) > 0 {
		all := make([]smt.Z3Expr, 0, len(tcs)+1)
		all = append(all, z3Fmla)
		all = append(all, tcs...)
		return s.tr.Ctx.And(all...), nil
	}
	return z3Fmla, nil
}

// formulaToZ3Closed translates and closes (universally quantifies free vars).
// Matches Python's formula_to_z3_closed (ivy_solver.py:646-655).
//
// For Definition: wraps in raw z3.ForAll (no quant constraints).
// For others: wraps via forall() helper (with quant constraints).
func (s *Solver) formulaToZ3Closed(fmla Expr) (smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:688 formula_to_z3_closed() ENTER type=%v HASH canon=%v", ShortTypeName(fmla), fmla.Sexp())
	z3Formula, err := s.tr.TranslateNoHash(fmla)
	if err != nil {
		return smt.Z3Expr{}, err
	}

	freeVars := FreeVariablesList(fmla)
	if len(freeVars) == 0 {
		return z3Formula, nil
	}

	// Sort variables matching Python: sorted(used_variables_ast(fmla))
	sort.Slice(freeVars, func(i, j int) bool {
		return freeVars[i].Name < freeVars[j].Name
	})

	z3Vars := make([]smt.Z3Expr, len(freeVars))
	for i, v := range freeVars {
		z3Vars[i], err = s.tr.TranslateVar(v)
		if err != nil {
			return smt.Z3Expr{}, err
		}
	}

	// Definition: raw z3.ForAll (no quant constraints)
	// Other: forall() with quant constraints
	// Matches Python formula_to_z3_closed lines 653-654
	if _, isDef := fmla.(*LogicDefinition); isDef {
		return s.tr.Ctx.ForAll(z3Vars, z3Formula), nil
	}
	return s.forall(freeVars, z3Vars, z3Formula), nil
}

// conjToZ3 translates a conjunction to Z3 without HASH trace.
// Matches Python's conj_to_z3 (ivy_solver.py:546-549).
// For And: recursively translates each conjunct.
// Otherwise: delegates to formulaToZ3Closed.
func (s *Solver) conjToZ3(fmla Expr) (smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:585 conj_to_z3() ENTER type=%v", ShortTypeName(fmla))
	if and, ok := fmla.(*LogicAnd); ok {
		z3Args := make([]smt.Z3Expr, len(and.Terms))
		for i, t := range and.Terms {
			var err error
			z3Args[i], err = s.conjToZ3(t)
			if err != nil {
				return smt.Z3Expr{}, err
			}
		}
		return s.tr.Ctx.And(z3Args...), nil
	}
	return s.formulaToZ3Closed(fmla)
}

// forall wraps a Z3 body in ForAll with quant constraints (nat/range bounds).
// Delegates to the solver's translator. Used by formulaToZ3Closed.
func (s *Solver) forall(vars []*LogicVariable, z3Vars []smt.Z3Expr, z3Body smt.Z3Expr) smt.Z3Expr {
	return s.tr.forall(vars, z3Vars, z3Body)
}

// NotClausesToZ3 negates a Clauses and converts to Z3.
// Corresponds to Python's not_clauses_to_z3.
func (s *Solver) NotClausesToZ3(clauses *Clauses) (smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:1102 not_clauses_to_z3() ENTER")
	// Separate Skolem definitions from other definitions
	var skolemDefs, otherDefs []*IvyDefinition
	for _, d := range clauses.Defs {
		sym := d.Defines()
		if c, ok := sym.(*Const); ok && z3bridgeIsSkolem(c.Name) {
			skolemDefs = append(skolemDefs, d)
		} else {
			otherDefs = append(otherDefs, d)
		}
	}

	// Skolem definitions are asserted positively
	skolemClauses := NewClauses(nil, skolemDefs, nil)
	zSkolem, err := s.ClausesToZ3(skolemClauses)
	if err != nil {
		return smt.Z3Expr{}, err
	}

	// Full clauses are negated
	zAll, err := s.ClausesToZ3(clauses)
	if err != nil {
		return smt.Z3Expr{}, err
	}

	_ = otherDefs
	return s.tr.Ctx.And(zSkolem, s.tr.Ctx.Not(zAll)), nil
}

func z3bridgeIsSkolem(name string) bool {
	return strings.Contains(name, "__")
}

// --- Satisfiability checks ---

// IsSat checks whether a formula is satisfiable.
// Returns true if satisfiable, false if unsatisfiable.
func (s *Solver) IsSat(fmla Expr) (bool, error) {
	xtracer.Trace("ivy_solver.py:816 is_sat() ENTER")
	result, err := s.tr.IsSat(fmla)
	if err != nil {
		return false, err
	}
	return result == smt.Sat, nil
}

// Implies checks whether fmla1 implies fmla2.
// Returns true if fmla1 => fmla2 is valid.
func (s *Solver) Implies(fmla1, fmla2 Expr) (bool, error) {
	return s.tr.Implies(fmla1, fmla2)
}

// ClausesSat checks whether a Clauses set is satisfiable.
// Corresponds to Python's clauses_sat.
func (s *Solver) ClausesSat(clauses *Clauses) (bool, error) {
	xtracer.Trace("ivy_solver.py:1113 clauses_sat() ENTER")
	z3solver := s.newZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return false, err
	}
	z3solver.Assert(zc)
	result := s.checkZ3(z3solver)
	return result != smt.Unsat, nil
}

// ClausesImply checks whether clauses1 imply clauses2.
// Corresponds to Python's clauses_imply.
func (s *Solver) ClausesImply(clauses1, clauses2 *Clauses) (bool, error) {
	xtracer.Trace("ivy_solver.py:1023 clauses_imply() ENTER")
	z3solver := s.newZ3Solver()

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

	result := s.checkZ3(z3solver)
	return result == smt.Unsat, nil
}

// ImpliesBatch tests if premise implies each formula in fmlas.
// Delegates to Z3Utils.Z3ImpliesBatch which uses the bare to_z3 translator
// (no closing, no HASH, no type constraints). Free variables become shared
// Z3 constants (not universally quantified).
// Corresponds to Python z3_utils.py:z3_implies_batch (lines 136-171).
func (s *Solver) ImpliesBatch(premise Expr, fmlas []Expr, timeout bool) ([]bool, error) {
	return s.z3u.Z3ImpliesBatch(premise, fmlas, timeout)
}

// Z3Implies tests if f1 implies f2 using the bare to_z3 translator.
// Corresponds to Python z3_utils.py:z3_implies (line 109).
func (s *Solver) Z3Implies(f1, f2 Expr, timeout bool) (bool, error) {
	return s.z3u.Z3Implies(f1, f2, timeout)
}

// ClausesImplyFormula checks whether clauses1 imply fmla2.
// Corresponds to Python's clauses_imply_formula.
func (s *Solver) ClausesImplyFormula(clauses1 *Clauses, fmla2 Expr) (bool, error) {
	xtracer.Trace("ivy_solver.py:1704 clauses_imply_formula() ENTER")
	z3solver := s.newZ3Solver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z1)

	negFmla := &LogicNot{Body: fmla2}
	z2, err := s.formulaToZ3(negFmla)
	if err != nil {
		return false, err
	}
	z3solver.Assert(z2)

	result := s.checkZ3(z3solver)
	return result == smt.Unsat, nil
}

// --- UNSAT Core ---

// UnsatCore extracts an unsatisfiable core from clauses1 with respect to clauses2.
// If the combined system is satisfiable, returns nil.
// The optional 'implies' parameter adds an implication constraint.
// The 'unlikely' function marks formulas that should preferably be excluded.
// Corresponds to Python's unsat_core.
func (s *Solver) UnsatCore(
	clauses1, clauses2 *Clauses,
	implies *Clauses,
	unlikely func(Expr) bool,
) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:723 unsat_core() ENTER")
	if unlikely == nil {
		unlikely = func(Expr) bool { return false }
	}

	fmlas := clauses1.Fmlas
	z3solver := s.newZ3Solver()

	// Create activation literals
	alits := make([]smt.Z3Expr, len(fmlas))
	for i := range fmlas {
		name := fmt.Sprintf("__c%d", i)
		alits[i] = s.tr.Ctx.Const(name, s.tr.Ctx.BoolSort())
	}

	// Assert: activation literal => formula
	for i, f := range fmlas {
		zf, err := s.formulaToZ3(f)
		if err != nil {
			return nil, err
		}
		clause := s.tr.Ctx.Or(s.tr.Ctx.Not(alits[i]), zf)
		z3solver.Assert(clause)
	}

	// Assert definitions from clauses1
	// Python unsat_core line 690: formula_to_z3(d.to_constraint())
	for _, d := range clauses1.Defs {
		constraint := DefinitionToConstraint(d)
		zd, err := s.formulaToZ3(constraint)
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
	result := s.checkZ3Assumptions(z3solver, alits)

	if result == smt.Sat {
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
	var resFmlas []Expr
	for i, f := range fmlas {
		if coreSet[alits[i].String()] {
			resFmlas = append(resFmlas, f)
		}
	}

	// If the core is empty but the system was unsat, include all formulas
	// (this can happen when the unsat-ness comes from clauses2/implies)
	if len(resFmlas) == 0 && result == smt.Unsat {
		resFmlas = append(resFmlas, fmlas...)
	}

	// Minimize core using biased_core approach: try removing unlikely formulas first,
	// then remaining. This produces a smaller core than Z3's native unsat_core.
	if len(resFmlas) > 1 {
		resFmlas = s.minimizeCore(z3solver, resFmlas, alits, fmlas, unlikely)
	}

	defs := make([]*IvyDefinition, len(clauses1.Defs))
	copy(defs, clauses1.Defs)
	return NewClauses(resFmlas, defs, nil), nil
}

// minimizeCore performs biased core minimization on a set of formulas.
// It tries removing unlikely formulas first, then remaining ones.
// Python: ivy_core.py minimize_core / biased_core
func (s *Solver) minimizeCore(
	z3solver *smt.Z3Solver,
	resFmlas []Expr, alits []smt.Z3Expr, allFmlas []Expr, unlikely func(Expr) bool,
) []Expr {
	// Build index from formula key to activation literal
	fmlaToAlit := make(map[string]smt.Z3Expr)
	for i, f := range allFmlas {
		fmlaToAlit[fmt.Sprint(f)] = alits[i]
	}

	core := make([]bool, len(resFmlas))
	for i := range core {
		core[i] = true
	}

	// Get activation literals for core formulas
	coreAlits := make([]smt.Z3Expr, len(resFmlas))
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
		if s.checkZ3Assumptions(z3solver, assumptions) == smt.Unsat {
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
		if s.checkZ3Assumptions(z3solver, assumptions) == smt.Unsat {
			// Still unsat, can keep it removed
		} else {
			core[i] = true
		}
	}

	var result []Expr
	for i, f := range resFmlas {
		if core[i] {
			result = append(result, f)
		}
	}
	return result
}

// collectAssumptions builds the list of activation literals for included formulas.
func collectAssumptions(alits []smt.Z3Expr, included []bool) []smt.Z3Expr {
	var result []smt.Z3Expr
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
func SortSizeConstraint(sort Sort, size int) Expr {
	xtracer.Trace("ivy_solver.py:1188 sort_size_constraint() ENTER sort=%s size=%d", sort, size)
	us, ok := sort.(*UninterpretedSort)
	if !ok {
		return True // trivially true for non-uninterpreted sorts
	}

	syms := make([]*Const, size)
	eqs := make([]Expr, size)
	v, _ := NewVariable("X"+us.Name, sort)
	for i := 0; i < size; i++ {
		syms[i] = NewConst(fmt.Sprintf("__%s$%d", us.Name, i), sort)
		eqs[i] = &Eq{T1: v, T2: syms[i]}
	}
	return &LogicOr{Terms: eqs}
}

// RelationSizeConstraint generates a constraint limiting a relation to at most 'size' true entries.
// Corresponds to Python's relation_size_constraint.
func RelationSizeConstraint(relation *Const, size int) Expr {
	xtracer.Trace("ivy_solver.py:1199 relation_size_constraint() ENTER size=%d", size)
	fs, ok := relation.CSort.(*LogicFunctionSort)
	if !ok {
		return True
	}

	domain := fs.Domain()
	// Create size-many tuples of constants
	consts := make([][]*Const, size)
	for i := 0; i < size; i++ {
		consts[i] = make([]*Const, len(domain))
		for j, s := range domain {
			consts[i][j] = NewConst(
				fmt.Sprintf("__$%s$%d$%d", relation.Name, i, j), s)
		}
	}

	// Create variables for the universal quantifier
	vs := make([]Expr, len(domain))
	for j, s := range domain {
		v, _ := NewVariable(fmt.Sprintf("X$%s$%d", relation.Name, j), s)
		vs[j] = v
	}

	// Build: ~relation(X0,...) | (X0=c00 & X1=c01 &...) | (X0=c10 & X1=c11 &...) | ...
	negApp := &LogicNot{Body: MustApply(relation, vs...)}

	disjuncts := []Expr{negApp}
	for i := 0; i < size; i++ {
		conjuncts := make([]Expr, len(domain))
		for j := range domain {
			conjuncts[j] = &Eq{T1: consts[i][j], T2: vs[j]}
		}
		disjuncts = append(disjuncts, &LogicAnd{Terms: conjuncts})
	}
	return &LogicOr{Terms: disjuncts}
}

// SizeConstraint generates a size constraint for either a sort or a relation.
func SizeConstraint(x Expr, size int) Expr {
	xtracer.Trace("ivy_solver.py:1226 size_constraint() ENTER size=%d", size)
	if us, ok := x.(*UninterpretedSort); ok {
		return SortSizeConstraint(us, size)
	}
	if c, ok := x.(*Const); ok {
		if _, isFS := c.CSort.(*LogicFunctionSort); isFS {
			return RelationSizeConstraint(c, size)
		}
	}
	return True
}

// --- Assume/Assert sequences ---

// AssumeAssert represents either an assumption or an assertion in a sequence.
type AssumeAssert struct {
	Clauses  *Clauses
	Doc      string
	IsAssert bool
}

// NewAssume creates an Assume entry.
func NewAssume(clauses *Clauses, doc string) AssumeAssert {
	return AssumeAssert{Clauses: clauses, Doc: doc, IsAssert: false}
}

// NewAssert creates an Assert entry.
func NewAssert(clauses *Clauses, doc string) AssumeAssert {
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
	xtracer.Trace("ivy_solver.py:1070 check_sequence() ENTER n=%d", len(seq))
	z3solver := s.newZ3Solver()
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
			dual := NegateClauses(aa.Clauses)
			z2, err := s.ClausesToZ3(dual)
			if err != nil {
				return nil, err
			}
			z3solver.Push()
			z3solver.Assert(z2)
			checkResult := s.checkZ3(z3solver) == smt.Unsat
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
