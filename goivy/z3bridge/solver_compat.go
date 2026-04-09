// Additional solver utility functions for compatibility checking
// and type system integration.
// Ported from Python's ivy_solver.py.
package z3bridge

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// CheckNativeCompatSym checks if a symbol's sort is compatible with native Z3 types.
// If the symbol has a native interpretation, creates dummy Z3 args and invokes
// LookupNative to verify the returned sort matches.
// Returns an error if there's a compatibility issue.
// Corresponds to Python's check_native_compat_sym.
func (s *Solver) CheckNativeCompatSym(sym *lg.Const) (retErr error) {
	xtracer.Trace("ivy_solver.py:350 check_native_compat_sym() ENTER sym=%s", sym)
	if s.sig == nil {
		return nil
	}
	if !il.IsFunctionSort(sym.CSort) {
		return nil // non-function sorts are always compatible
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}

	// Python wraps the entire native invocation in try/except:
	//   except Exception as e:
	//       raise IvyError(None, 'cannot interpret {} as {}: {}'.format(sym, sig.interp[sym.name], e))
	//
	defer func() {
		if r := recover(); r != nil {
			interpName := ""
			if s.sig != nil {
				if interp, ok := s.sig.Interp[sym.Name]; ok {
					interpName = fmt.Sprintf("%v", interp)
				}
			}
			retErr = fmt.Errorf("recovered panic: cannot interpret %s as %s: %v", sym.Name, interpName, r)
			// however, we want to show the origin of errors, not hide them.
			// so we turn this off during development. We might activate
			// it later. So retain the panic.
			alwaysPrintf("%v\nstack=\n%v\n", retErr, stack())
			panic(retErr)
		}
	}()

	for _, d := range fs.Domain() {
		if err := checkSortCompat(s.sig, d); err != nil {
			return fmt.Errorf("symbol %s: domain sort %s: %w", sym.Name, d, err)
		}
	}
	if err := checkSortCompat(s.sig, fs.Range()); err != nil {
		return fmt.Errorf("symbol %s: range sort %s: %w", sym.Name, fs.Range(), err)
	}

	// Check native interpretation compatibility by invoking LookupNative
	var table func(string) any
	var kind string
	if _, isBool := fs.Range().(*lg.BooleanSort); isBool {
		table = s.Relations
		kind = "relation"
	} else {
		table = s.Functions
		kind = "function"
	}
	result := s.LookupNative(sym, table, kind)
	if result == nil {
		return nil // no native interpretation
	}
	nf, ok := result.(NativeFunc)
	if !ok {
		return nil // not a callable
	}

	// Create dummy Z3 args and invoke
	ctx := s.tr.Ctx
	args := make([]Expr, fs.Arity())
	for i, d := range fs.Domain() {
		zs, err := s.tr.TranslateSort(d)
		if err != nil {
			return fmt.Errorf("symbol %s: cannot translate domain sort %d: %w", sym.Name, i, err)
		}
		args[i] = ctx.Const(fmt.Sprintf("__compat_check_%d", i), zs)
	}
	nfResult := nf(args...)

	// Check result sort matches declared range
	expectedSort, err := s.tr.TranslateSort(fs.Range())
	if err != nil {
		return nil // can't check
	}
	resultSort := nfResult.ExprSort()
	if resultSort.Kind() != expectedSort.Kind() {
		return fmt.Errorf("symbol %s: native interpretation returns sort %s but expected %s",
			sym.Name, resultSort.String(), expectedSort.String())
	}

	return nil
}

// CheckNativeCompatSymStatic is the old static version for backward compatibility.
func CheckNativeCompatSymStatic(sig *il.Sig, sym *lg.Const) error {
	xtracer.Trace("ivy_solver.py:350 check_native_compat_sym() ENTER sym=%s", sym)
	if !il.IsFunctionSort(sym.CSort) {
		return nil
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	for _, d := range fs.Domain() {
		if err := checkSortCompat(sig, d); err != nil {
			return fmt.Errorf("symbol %s: domain sort %s: %w", sym.Name, d, err)
		}
	}
	if err := checkSortCompat(sig, fs.Range()); err != nil {
		return fmt.Errorf("symbol %s: range sort %s: %w", sym.Name, fs.Range(), err)
	}
	return nil
}

func checkSortCompat(sig *il.Sig, s lg.Sort) error {
	// Check that the sort either has no interpretation or a compatible one
	name := il.SortName(s)
	interp, ok := sig.Interp[name]
	if !ok {
		return nil // no interpretation = OK
	}
	switch v := interp.(type) {
	case string:
		if IsSolverSort(v) || v == "int" || v == "nat" {
			return nil
		}
		_, _, ok := ParseArrayTheory(v)
		if ok {
			return nil
		}
		return fmt.Errorf("unsupported native type: %s", v)
	}
	return nil
}

// CheckCompat checks interpreted symbols in the signature for native compatibility.
// Python iterates sig.interp (not sig.symbols), so we do the same.
func (s *Solver) CheckCompat() []error {
	xtracer.Trace("ivy_solver.py:374 check_compat() ENTER")
	var errs []error
	if s.sig == nil {
		return nil
	}
	for name := range s.sig.Interp {
		entry, ok := s.sig.Symbols[name]
		if !ok {
			continue
		}
		// Handle UnionSort (polymorphic symbols)
		if entry.Union != nil {
			for _, sort := range entry.Union.Sorts {
				sym := lg.NewConst(name, sort)
				if err := s.CheckNativeCompatSym(sym); err != nil {
					errs = append(errs, err)
				}
			}
		} else {
			sym := lg.NewConst(name, entry.Sort)
			if err := s.CheckNativeCompatSym(sym); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// CheckCompatStatic is the old static version for backward compatibility.
// Python iterates sig.interp (not sig.symbols), so we do the same.
func CheckCompatStatic(sig *il.Sig) []error {
	xtracer.Trace("ivy_solver.py:374 check_compat() ENTER")
	var errs []error
	for name := range sig.Interp {
		entry, ok := sig.Symbols[name]
		if !ok {
			continue
		}
		if entry.Union != nil {
			for _, sort := range entry.Union.Sorts {
				sym := lg.NewConst(name, sort)
				if err := CheckNativeCompatSymStatic(sig, sym); err != nil {
					errs = append(errs, err)
				}
			}
		} else {
			sym := lg.NewConst(name, entry.Sort)
			if err := CheckNativeCompatSymStatic(sig, sym); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// TermsMatch checks if tl2 is an instance of tl1, where variables in tl1
// can be bound to terms in tl2 with consistency checking.
// Corresponds to Python's terms_match (ivy_solver.py:753-766).
func TermsMatch(tl1, tl2 []lg.Expr) bool {
	xtracer.Trace("ivy_solver.py:830 terms_match() ENTER")
	if len(tl1) != len(tl2) {
		return false
	}
	env := make(map[string]string)
	for i := range tl1 {
		x := tl1[i]
		y := tl2[i]
		if v, ok := x.(*lg.Variable); ok {
			yName := exprName(y)
			if prev, exists := env[v.Name]; exists {
				if yName != prev {
					return false
				}
			} else {
				env[v.Name] = yName
			}
		} else {
			if exprName(x) != exprName(y) {
				return false
			}
		}
	}
	return true
}

// exprName extracts the representative name from an expression
// (Symbol.Name, Variable.Name, or string representation).
func exprName(e lg.Expr) string {
	switch v := e.(type) {
	case *lg.Const:
		return v.Name
	case *lg.Variable:
		return v.Name
	default:
		return fmt.Sprint(e)
	}
}

// GetArgRange returns the range of argument values from a model for a function symbol.
func (s *Solver) GetArgRange(model *HerbrandModel, x *lg.Const) []lg.Expr {
	xtracer.Trace("ivy_solver.py:846 get_arg_range() ENTER")
	sort := il.SortRange(x.CSort)
	universe := model.SortUniverse(sort)
	result := make([]lg.Expr, len(universe))
	for i, c := range universe {
		result[i] = c
	}
	return result
}

// ModelIfNone returns the provided model, or creates one from clauses if nil.
// If model is nil, it performs the incremental sort-size search to find a
// small model. All uninterpreted sorts are searched at the same size N
// simultaneously, matching Python's model_if_none (ivy_solver.py:1135-1161).
// The implied parameter is negated and added to the solver (not conjoined).
func (s *Solver) ModelIfNone(clauses *module.Clauses, implied *module.Clauses, model *HerbrandModel) *HerbrandModel {
	xtracer.Trace("ivy_solver.py:1238 model_if_none() ENTER")
	if model != nil {
		return model
	}

	z3solver := s.tr.Ctx.NewZ3Solver()

	// Add main clauses
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil
	}
	z3solver.Assert(zc)

	// Add negation of implied (if any)
	if implied != nil {
		zi, err := s.NotClausesToZ3(implied)
		if err != nil {
			return nil
		}
		z3solver.Assert(zi)
	}

	// Collect uninterpreted sorts from the signature
	var uninterpSorts []lg.Sort
	if s.sig != nil {
		for name, sort := range s.sig.Sorts {
			if _, interp := s.sig.Interp[name]; !interp {
				if _, isUS := sort.(*lg.UninterpretedSort); isUS {
					uninterpSorts = append(uninterpSorts, sort)
				}
			}
		}
	}

	// Simultaneous sort-size search: try all sorts at size N together
	for sortSize := 1; ; sortSize++ {
		z3solver.Push()
		for _, sort := range uninterpSorts {
			sc := SortSizeConstraint(sort, sortSize)
			zsc, err := s.formulaToZ3(sc)
			if err != nil {
				continue
			}
			z3solver.Assert(zsc)
		}
		if z3solver.Check() != Unsat {
			m := z3solver.Model()
			if m == nil {
				z3solver.Pop()
				return nil
			}
			// Collect vocabulary
			symSet := clauses.Symbols()
			if implied != nil {
				for sKey, sNode := range implied.Symbols() {
					symSet[sKey] = sNode
				}
			}
			vocab := make([]*lg.Const, 0, len(symSet))
			for _, sym := range symSet {
				vocab = append(vocab, sym)
			}
			h := NewHerbrandModel(s, z3solver, m, vocab)
			z3solver.Pop()
			return h
		}
		z3solver.Pop()

		// Safety: if no uninterpreted sorts, don't loop
		if len(uninterpSorts) == 0 {
			return nil
		}
	}
}

// CollectModelValues collects all model values for a symbol of a given sort.
func (s *Solver) CollectModelValues(sort lg.Sort, model *HerbrandModel, sym *lg.Const) []*lg.Const {
	xtracer.Trace("ivy_solver.py:884 collect_model_values() ENTER sort=%v", sort)
	if model == nil {
		return nil
	}
	return model.SortUniverse(sort)
}

// NumeralAssign assigns numeral names to universe elements, respecting existing
// numerals in the clause set.
//
// For each sort:
//  1. First assigns existing numerals from the clause set to their model values.
//  2. Then assigns fresh numeral names (0, 1, 2, ...) to remaining elements,
//     skipping names already used by existing numerals.
//
// Returns a map from model element names to numeral names.
//
// Corresponds to Python numeral_assign (lines 1338-1371).
func NumeralAssign(model *HerbrandModel) map[string]string {
	xtracer.Trace("ivy_solver.py:1480 numeral_assign() ENTER")
	return NumeralAssignWithClauses(model, nil)
}

// NumeralAssignWithClauses is the full version that respects existing numerals.
func NumeralAssignWithClauses(model *HerbrandModel, clauses *module.Clauses) map[string]string {
	result := make(map[string]string)

	// Collect existing numerals from clauses, grouped by sort
	numBySort := make(map[string][]*lg.Const)
	if clauses != nil {
		usedConsts := module.ConstantsClauses(clauses)
		for _, c := range usedConsts {
			if il.IsNumeral(c) {
				sortName := il.SortName(il.SortRange(c.CSort))
				numBySort[sortName] = append(numBySort[sortName], c)
			}
		}
	}

	for _, sort := range model.Sorts() {
		sortName := il.SortName(sort)

		// Skip interpreted sorts
		if il.IsInterpretedSort(model.sig, sort) {
			continue
		}

		// First pass: assign existing numerals to their model values
		usedNumerals := make(map[string]bool)
		foom := make(map[string]*lg.Const) // model element → numeral

		for _, num := range numBySort[sortName] {
			modelVal := model.EvalConstant(num)
			if modelVal != nil {
				if _, already := foom[modelVal.Name]; already {
					// Two numerals assigned same value — warn but continue
					continue
				}
				foom[modelVal.Name] = num
				usedNumerals[num.Name] = true
			}
		}

		// Second pass: assign fresh numerals to remaining elements
		elems := model.SortedSortUniverse(sort)
		i := 0
		for _, c := range elems {
			if _, assigned := foom[c.Name]; !assigned {
				// Find next unused numeral name
				for {
					name := fmt.Sprintf("%d", i)
					i++
					numConst := lg.NewConst(name, sort)
					if !usedNumerals[numConst.Name] {
						foom[c.Name] = numConst
						break
					}
				}
			}
		}

		// Copy into result
		for elemName, numConst := range foom {
			result[elemName] = numConst.Name
		}
	}

	return result
}

// MineInterpretedConstants extracts interpreted constants from a Z3 model.
// This is called during HerbrandModel construction but also available separately.
func (s *Solver) MineInterpretedConstants(vocab []*lg.Const) map[string][]*lg.Const {
	xtracer.Trace("ivy_solver.py:891 mine_interpreted_constants() ENTER")
	result := make(map[string][]*lg.Const)
	for _, c := range vocab {
		sortName := il.SortName(il.SortRange(c.CSort))
		if !il.IsInterpretedSort(s.sig, c.CSort) {
			continue
		}
		result[sortName] = append(result[sortName], c)
	}
	return result
}

// GetPolymacs returns polymorphic macros for an operator.
func GetPolymacs(op string) func([]lg.Expr) lg.Expr {
	xtracer.Trace("ivy_solver.py:513 get_polymacs() ENTER op=%s", op)
	switch op {
	case "<=":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			lt := lg.MustApply(lg.NewConst("<", il.RelationSort([]lg.Sort{args[0].NodeSort(), args[1].NodeSort()})), args...)
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Expr{lt, eq}}
		}
	case ">":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			return lg.MustApply(lg.NewConst("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})), args[1], args[0])
		}
	case ">=":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			lt := lg.MustApply(lg.NewConst("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})), args[1], args[0])
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Expr{lt, eq}}
		}
	}
	return nil
}

// QuantConstraints generates sort constraints for quantifier variables.
// For finite/enumerated sorts, generates membership constraints.
func QuantConstraints(vs []*lg.Variable, z3Vs interface{}) lg.Expr {
	xtracer.Trace("ivy_solver.py:545 quant_constraints() ENTER nvars=%d", len(vs))
	var constraints []lg.Expr
	for _, v := range vs {
		if es, ok := v.VSort.(*lg.EnumeratedSort); ok {
			// Generate: v = e0 | v = e1 | ...
			eqs := make([]lg.Expr, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: v, T2: lg.NewConst(name, v.VSort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}

