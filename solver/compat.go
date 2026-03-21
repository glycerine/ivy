// Additional solver utility functions for compatibility checking
// and type system integration.
// Ported from Python's ivy_solver.py.
package solver

import (
	"fmt"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// CheckNativeCompatSym checks if a symbol's sort is compatible with native Z3 types.
// If the symbol has a native interpretation, creates dummy Z3 args and invokes
// LookupNative to verify the returned sort matches.
// Returns an error if there's a compatibility issue.
// Corresponds to Python's check_native_compat_sym.
func (s *Solver) CheckNativeCompatSym(sym *lg.Symbol) (retErr error) {
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
	isRelation := false
	if _, isBool := fs.Range().(*lg.BooleanSort); isBool {
		isRelation = true
	}
	nf := s.LookupNative(sym, isRelation)
	if nf == nil {
		return nil // no native interpretation
	}

	// Create dummy Z3 args and invoke
	ctx := s.tr.Ctx
	args := make([]z3bridge.Expr, fs.Arity())
	for i, d := range fs.Domain() {
		zs, err := s.tr.TranslateSort(d)
		if err != nil {
			return fmt.Errorf("symbol %s: cannot translate domain sort %d: %w", sym.Name, i, err)
		}
		args[i] = ctx.Const(fmt.Sprintf("__compat_check_%d", i), zs)
	}
	result := nf(args...)

	// Check result sort matches declared range
	expectedSort, err := s.tr.TranslateSort(fs.Range())
	if err != nil {
		return nil // can't check
	}
	resultSort := result.ExprSort()
	if resultSort.Kind() != expectedSort.Kind() {
		return fmt.Errorf("symbol %s: native interpretation returns sort %s but expected %s",
			sym.Name, resultSort.String(), expectedSort.String())
	}

	return nil
}

// CheckNativeCompatSymStatic is the old static version for backward compatibility.
func CheckNativeCompatSymStatic(sig *il.Sig, sym *lg.Symbol) error {
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

// CheckCompat checks all symbols in the signature for native compatibility.
func (s *Solver) CheckCompat() []error {
	var errs []error
	if s.sig == nil {
		return nil
	}
	for _, entry := range s.sig.Symbols {
		sym := lg.NewSymbol(entry.Name, entry.Sort)
		if err := s.CheckNativeCompatSym(sym); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// CheckCompatStatic is the old static version for backward compatibility.
func CheckCompatStatic(sig *il.Sig) []error {
	var errs []error
	for _, entry := range sig.Symbols {
		sym := lg.NewSymbol(entry.Name, entry.Sort)
		if err := CheckNativeCompatSymStatic(sig, sym); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// TermsMatch checks if two term lists structurally match.
func TermsMatch(tl1, tl2 []lg.Expr) bool {
	if len(tl1) != len(tl2) {
		return false
	}
	for i := range tl1 {
		if !tl1[i].Equal(tl2[i]) {
			return false
		}
	}
	return true
}

// GetArgRange returns the range of argument values from a model for a function symbol.
func (s *Solver) GetArgRange(model *HerbrandModel, x *lg.Symbol) []lg.Expr {
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
func (s *Solver) ModelIfNone(clauses *clauseops.Clauses, implied *clauseops.Clauses, model *HerbrandModel) *HerbrandModel {
	if model != nil {
		return model
	}

	z3solver := s.tr.Ctx.NewSolver()

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
			zsc, err := s.translateClosed(sc)
			if err != nil {
				continue
			}
			z3solver.Assert(zsc)
		}
		if z3solver.Check() != z3bridge.Unsat {
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
			vocab := make([]*lg.Symbol, 0, len(symSet))
			for _, sym := range symSet {
				if c, ok := sym.(*lg.Symbol); ok {
					vocab = append(vocab, c)
				}
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

// ClauseModelSimp simplifies a clause using a model.
// Removes literals that are false in the model.
func (s *Solver) ClauseModelSimp(model *HerbrandModel, clause lg.Expr) lg.Expr {
	if model == nil {
		return clause
	}
	or, ok := clause.(*lg.Or)
	if !ok {
		return clause
	}
	var kept []lg.Expr
	for _, term := range or.Terms {
		if model.Eval(term) {
			kept = append(kept, term)
		}
	}
	if len(kept) == 0 {
		return &lg.Or{} // false
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return &lg.Or{Terms: kept}
}

// CollectModelValues collects all model values for a symbol of a given sort.
func (s *Solver) CollectModelValues(sort lg.Sort, model *HerbrandModel, sym *lg.Symbol) []*lg.Symbol {
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
	return NumeralAssignWithClauses(model, nil)
}

// NumeralAssignWithClauses is the full version that respects existing numerals.
func NumeralAssignWithClauses(model *HerbrandModel, clauses *clauseops.Clauses) map[string]string {
	result := make(map[string]string)

	// Collect existing numerals from clauses, grouped by sort
	numBySort := make(map[string][]*lg.Symbol)
	if clauses != nil {
		usedConsts := clauseops.ConstantsClauses(clauses)
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
		foom := make(map[string]*lg.Symbol) // model element → numeral

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
					numConst := lg.NewSymbol(name, sort)
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
func (s *Solver) MineInterpretedConstants(vocab []*lg.Symbol) map[string][]*lg.Symbol {
	result := make(map[string][]*lg.Symbol)
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
	switch op {
	case "<=":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			lt := &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[0].NodeSort(), args[1].NodeSort()})),
				Terms: args,
			}
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Expr{lt, eq}}
		}
	case ">":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			return &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
				Terms: []lg.Expr{args[1], args[0]},
			}
		}
	case ">=":
		return func(args []lg.Expr) lg.Expr {
			if len(args) != 2 {
				return nil
			}
			lt := &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
				Terms: []lg.Expr{args[1], args[0]},
			}
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Expr{lt, eq}}
		}
	}
	return nil
}

// QuantConstraints generates sort constraints for quantifier variables.
// For finite/enumerated sorts, generates membership constraints.
func QuantConstraints(vs []*lg.Variable, z3Vs interface{}) lg.Expr {
	var constraints []lg.Expr
	for _, v := range vs {
		if es, ok := v.VSort.(*lg.EnumeratedSort); ok {
			// Generate: v = e0 | v = e1 | ...
			eqs := make([]lg.Expr, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: v, T2: lg.NewSymbol(name, v.VSort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}

// TypeConstraints generates type constraints for a set of symbols.
func TypeConstraints(syms []*lg.Symbol) lg.Expr {
	// For each symbol with an enumerated sort, generate range constraints
	var constraints []lg.Expr
	for _, sym := range syms {
		sort := il.SortRange(sym.CSort)
		if es, ok := sort.(*lg.EnumeratedSort); ok {
			eqs := make([]lg.Expr, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: sym, T2: lg.NewSymbol(name, sort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}
