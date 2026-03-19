// z3convert.go provides functions for converting Z3 expressions back to Ivy
// formulas, Craig interpolation, sort ordering, and related utilities.
// Ported from Python ivy_solver.py.
package solver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/goivy/clauseops"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// --- Z3 sort → Ivy sort ---

// Z3SortToSort converts a Z3 sort back to an Ivy sort.
// Corresponds to Python's z3sort_to_sort.
func Z3SortToSort(z3sort z3bridge.Sort) lg.Sort {
	kind := z3sort.Kind()
	switch kind {
	case z3bridge.SortBool:
		return lg.Boolean
	case z3bridge.SortInt:
		return &lg.UninterpretedSort{Name: "int"}
	case z3bridge.SortReal:
		return &lg.UninterpretedSort{Name: "real"}
	case z3bridge.SortArray:
		domSort := Z3SortToSort(z3sort.ArrayDomain())
		rngSort := Z3SortToSort(z3sort.ArrayRange())
		domName := sortToName(domSort)
		rngName := sortToName(rngSort)
		return &lg.UninterpretedSort{Name: "arr[" + domName + "][" + rngName + "]"}
	default:
		// Uninterpreted or other: use the name
		name := z3sort.String()
		return &lg.UninterpretedSort{Name: name}
	}
}

// --- Z3 func_decl → Ivy symbol ---

// Z3DeclToSymbol converts a Z3 function declaration to an Ivy constant (symbol).
// Corresponds to Python's z3decl_to_symbol.
func Z3DeclToSymbol(z3decl z3bridge.FuncDecl) *lg.Symbol {
	arity := z3decl.Arity()
	rng := Z3SortToSort(z3decl.RangeSort())

	// Strip the ":sort" suffix from the name (added by Translator)
	name := z3decl.Name()
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[:idx]
	}

	if arity == 0 {
		return lg.NewSymbol(name, rng)
	}

	dom := make([]lg.Sort, arity)
	for i := 0; i < arity; i++ {
		dom[i] = Z3SortToSort(z3decl.DomainSort(i))
	}
	sortArgs := append(dom, rng)
	fs, err := lg.NewFunctionSort(sortArgs...)
	if err != nil {
		// Fallback: return a const with range sort
		return lg.NewSymbol(name, rng)
	}
	return lg.NewSymbol(name, fs)
}

// --- Z3 expression → Ivy formula ---

// Z3ToFormula converts a Z3 expression back to an Ivy formula.
// The vars parameter holds de Bruijn variable bindings (innermost first).
// Corresponds to Python's z3_to_formula.
func Z3ToFormula(z3expr z3bridge.Expr, vars []*lg.Variable) (lg.Expr, error) {
	// Application (includes constants, And, Or, Not, Eq, etc.)
	if z3expr.IsApp() {
		arity := z3expr.NumArgs()

		// Recursively convert arguments
		args := make([]lg.Expr, arity)
		for i := 0; i < arity; i++ {
			arg, err := Z3ToFormula(z3expr.Arg(i), vars)
			if err != nil {
				return nil, err
			}
			args[i] = arg
		}

		decl := z3expr.Decl()
		kind := decl.Kind()

		switch kind {
		case z3bridge.DeclAnd:
			if len(args) == 0 {
				return lg.True, nil
			}
			return &lg.And{Terms: args}, nil

		case z3bridge.DeclOr:
			if len(args) == 0 {
				return lg.False, nil
			}
			return &lg.Or{Terms: args}, nil

		case z3bridge.DeclNot:
			if len(args) != 1 {
				return nil, fmt.Errorf("z3_to_formula: Not with %d args", len(args))
			}
			return &lg.Not{Body: args[0]}, nil

		case z3bridge.DeclEq:
			if len(args) != 2 {
				return nil, fmt.Errorf("z3_to_formula: Eq with %d args", len(args))
			}
			return &lg.Eq{T1: args[0], T2: args[1]}, nil

		case z3bridge.DeclITE:
			if len(args) != 3 {
				return nil, fmt.Errorf("z3_to_formula: ITE with %d args", len(args))
			}
			return &lg.Ite{Cond: args[0], Then: args[1], Else: args[2]}, nil

		case z3bridge.DeclTrue:
			return lg.True, nil

		case z3bridge.DeclFalse:
			return lg.False, nil

		case z3bridge.DeclIff:
			if len(args) != 2 {
				return nil, fmt.Errorf("z3_to_formula: Iff with %d args", len(args))
			}
			return &lg.Iff{T1: args[0], T2: args[1]}, nil

		default:
			// Uninterpreted function/constant
			sym := Z3DeclToSymbol(decl)
			if arity == 0 {
				return sym, nil
			}
			return &lg.Apply{Func: sym, Terms: args}, nil
		}
	}

	// Quantifier
	if z3expr.IsQuantifier() {
		nVars := z3expr.QuantNumVars()
		qVars := make([]*lg.Variable, nVars)
		for i := 0; i < nVars; i++ {
			name := z3expr.QuantVarName(i)
			// Strip the ":sort" suffix from var name if present
			if idx := strings.Index(name, ":"); idx >= 0 {
				name = name[:idx]
			}
			// Fix var names starting with '%' (Z3 internal names)
			if strings.HasPrefix(name, "%") {
				name = "V" + name[1:]
			}
			sort := Z3SortToSort(z3expr.QuantVarSort(i))
			v, err := lg.NewVariable(name, sort)
			if err != nil {
				return nil, fmt.Errorf("z3_to_formula: creating var %s: %w", name, err)
			}
			qVars[i] = v
		}

		// Build new vars list: reversed qVars prepended to existing vars
		// (de Bruijn: innermost bindings come first)
		newVars := make([]*lg.Variable, 0, len(qVars)+len(vars))
		for i := len(qVars) - 1; i >= 0; i-- {
			newVars = append(newVars, qVars[i])
		}
		newVars = append(newVars, vars...)

		body, err := Z3ToFormula(z3expr.QuantBody(), newVars)
		if err != nil {
			return nil, err
		}

		if z3expr.IsForAll() {
			return &lg.ForAll{Variables: qVars, Body: body}, nil
		}
		return &lg.Exists{Variables: qVars, Body: body}, nil
	}

	// Bound variable (de Bruijn index)
	if z3expr.IsVar() {
		idx := z3expr.VarIndex()
		if idx < 0 || idx >= len(vars) {
			return nil, fmt.Errorf("z3_to_formula: de Bruijn index %d out of range (have %d vars)", idx, len(vars))
		}
		return vars[idx], nil
	}

	// Numeral
	if z3expr.IsNumeral() {
		s := z3expr.String()
		sort := Z3SortToSort(z3expr.ExprSort())
		return lg.NewSymbol(s, sort), nil
	}

	return nil, fmt.Errorf("z3_to_formula: cannot convert Z3 expression: %s", z3expr.String())
}

// Z3ToFormulaNoVars is a convenience wrapper that calls Z3ToFormula with no
// initial variable bindings.
func Z3ToFormulaNoVars(z3expr z3bridge.Expr) (lg.Expr, error) {
	return Z3ToFormula(z3expr, nil)
}

// --- Binary interpolant (Craig interpolation) ---

// BinaryInterpolant computes a Craig interpolant between two clause sets.
// Given clauses2 and clauses1 where (clauses2 AND clauses1) is unsat,
// returns a formula I such that:
//   - clauses2 implies I
//   - I AND clauses1 is unsat
//   - I only uses symbols common to both clause sets
//
// Z3's interpolation API may not be available in all builds, so this includes
// a fallback that returns an error.
// Corresponds to Python's binary_interpolant.
func (s *Solver) BinaryInterpolant(clauses2, clauses1 *clauseops.Clauses) (*clauseops.Clauses, error) {
	// Create a fresh translator with an interpolation-capable context.
	// Z3_compute_interpolant requires a context created via
	// Z3_mk_interpolation_context (legacy solver with proof generation).
	itpTr := z3bridge.NewTranslatorWithInterpolation()

	// Wire up native lookups on the interpolation translator so that
	// polymorphic symbols and native interpretations are handled.
	itpTr.NativeLookup = s.tr.NativeLookup
	itpTr.SolverName = s.tr.SolverName

	// Re-translate both clause sets into the interpolation context.
	itpSolver := &Solver{
		tr:   itpTr,
		opts: s.opts,
		sig:  s.sig,
	}

	z2, err := itpSolver.ClausesToZ3(clauses2)
	if err != nil {
		return nil, fmt.Errorf("binary_interpolant: translating clauses2: %w", err)
	}
	z1, err := itpSolver.ClausesToZ3(clauses1)
	if err != nil {
		return nil, fmt.Errorf("binary_interpolant: translating clauses1: %w", err)
	}

	// Compute the interpolant using the interpolation context.
	itp, err := computeZ3Interpolant(itpTr.Ctx, z2, z1)
	if err != nil {
		return nil, fmt.Errorf("binary_interpolant: %w", err)
	}

	// Convert the Z3 interpolant back to an Ivy formula.
	ivyFmla, err := Z3ToFormulaNoVars(itp)
	if err != nil {
		return nil, fmt.Errorf("binary_interpolant: converting interpolant: %w", err)
	}

	return clauseops.NewClauses([]lg.Expr{ivyFmla}, nil, nil), nil
}

// computeZ3Interpolant computes a Craig interpolant between two Z3 formulas
// using Z3's interpolation API.
// Given a and b where (a AND b) is unsat, returns a formula I such that:
//   - a implies I
//   - I AND b is unsat
//   - I only uses symbols common to both a and b
//
// The ctx must be an interpolation-capable context (created via
// z3bridge.NewInterpolationContext).
// Corresponds to Python's z3.binary_interpolant.
func computeZ3Interpolant(ctx *z3bridge.Context, a, b z3bridge.Expr) (z3bridge.Expr, error) {
	// Build the interpolation pattern: And(Interpolant(a), b)
	// This mirrors Python's z3.binary_interpolant which does:
	//   f = And(Interpolant(a), b)
	//   ti = tree_interpolant(f)
	//   return ti[0]
	marked := ctx.MkInterpolant(a)
	pattern := ctx.And(marked, b)

	interps, err := ctx.ComputeInterpolant(pattern)
	if err != nil {
		return z3bridge.Expr{}, fmt.Errorf("computing interpolant: %w", err)
	}
	if len(interps) == 0 {
		return z3bridge.Expr{}, fmt.Errorf("interpolation returned empty result")
	}
	return interps[0], nil
}

// --- Collect numerals ---

// CollectNumeralsRecursive recursively collects all numeral subterms from a
// Z3 expression. For ITE expressions, it recurses into the then/else branches.
// Corresponds to Python's collect_numerals.
func CollectNumeralsRecursive(z3term z3bridge.Expr) []z3bridge.Expr {
	var result []z3bridge.Expr
	collectNumeralsHelper(z3term, &result)
	return result
}

func collectNumeralsHelper(z3term z3bridge.Expr, result *[]z3bridge.Expr) {
	// Check if it's a numeral (int value or bv value)
	if z3term.IsNumeral() {
		*result = append(*result, z3term)
		return
	}

	// Check if it's an ITE application — recurse into then/else branches
	if z3term.IsApp() && z3term.IsAppOf(z3bridge.DeclITE) {
		if z3term.NumArgs() == 3 {
			collectNumeralsHelper(z3term.Arg(1), result) // then branch
			collectNumeralsHelper(z3term.Arg(2), result) // else branch
		}
		return
	}

	// For other string-representable numerals (legacy path from encoding.go)
	s := z3term.String()
	if len(s) > 0 && (s[0] >= '0' && s[0] <= '9' || s[0] == '-') {
		*result = append(*result, z3term)
	}
}

// --- From Z3 numeral ---

// FromZ3Numeral converts a Z3 numeral expression to an Ivy constant
// of the given sort.
// Corresponds to Python's from_z3_numeral.
func FromZ3Numeral(z3term z3bridge.Expr, sort lg.Sort) *lg.Symbol {
	name := z3term.String()
	if len(name) == 0 {
		return lg.NewSymbol("0", sort)
	}
	// Validate: should start with digit, quote, or minus
	if !(name[0] >= '0' && name[0] <= '9' || name[0] == '"' || name[0] == '-') {
		fmt.Printf("warning: unexpected numeral from Z3 model: %s\n", name)
	}
	return lg.NewSymbol(name, sort)
}

// --- Collect model values ---

// CollectModelValuesZ3 collects all model values for a symbol of a given sort
// from a Z3 model. Uses sym_placeholders to create a term, evaluates it in
// the model, and collects the numerals.
// Corresponds to Python's collect_model_values.
func (s *Solver) CollectModelValuesZ3(sort lg.Sort, model *z3bridge.Model, sym *lg.Symbol) map[string]*lg.Symbol {
	result := make(map[string]*lg.Symbol)

	// Create the term: sym(V0, V1, ...)
	phs := clauseops.SymPlaceholders(sym)
	var term lg.Expr
	if len(phs) == 0 {
		term = sym
	} else {
		args := make([]lg.Expr, len(phs))
		for i, v := range phs {
			args[i] = v
		}
		term = &lg.Apply{Func: sym, Terms: args}
	}

	// Translate to Z3 and evaluate
	z3term, err := s.tr.Translate(term)
	if err != nil {
		return result
	}

	val, ok := model.Eval(z3term, true)
	if !ok {
		return result
	}

	// Collect numerals from the evaluation result
	nums := CollectNumeralsRecursive(val)
	for _, n := range nums {
		c := FromZ3Numeral(n, sort)
		result[c.Name] = c
	}

	return result
}

// --- SortOrder ---

// SortOrder implements an ordering on Z3 expressions for model construction.
// It uses a Z3 order relation and a model to compare values.
// Corresponds to Python's SortOrder class.
type SortOrder struct {
	Vs    []z3bridge.Expr    // Z3 variables for the order relation
	Order z3bridge.Expr      // Z3 expression representing the order (e.g., less-than)
	Model *z3bridge.Model    // Z3 model for evaluation
	Ctx   *z3bridge.Context  // Z3 context for substitution
}

// NewSortOrder creates a new SortOrder.
func NewSortOrder(vs []z3bridge.Expr, order z3bridge.Expr, model *z3bridge.Model, ctx *z3bridge.Context) *SortOrder {
	return &SortOrder{
		Vs:    vs,
		Order: order,
		Model: model,
		Ctx:   ctx,
	}
}

// Compare returns -1 if x < y according to the order, +1 otherwise.
// This implements a comparison function suitable for sorting.
// Corresponds to Python's SortOrder.__call__.
func (so *SortOrder) Compare(x, y z3bridge.Expr) int {
	if len(so.Vs) < 2 {
		return 0
	}

	// Substitute vs[0]->x, vs[1]->y into the order expression
	from := so.Vs[:2]
	to := []z3bridge.Expr{x, y}
	fact := so.Ctx.Substitute(so.Order, from, to)

	// Evaluate in the model
	val, ok := so.Model.Eval(fact, true)
	if !ok {
		return 0
	}
	if val.IsTrue() {
		return -1
	}
	return 1
}

// --- Z3 substitution wrapper ---

// SubstituteZ3 applies a substitution on a Z3 expression.
// pairs is a list of (from, to) expression pairs.
// This is a thin wrapper around Context.Substitute.
// Corresponds to Python's substitute.
func SubstituteZ3(ctx *z3bridge.Context, t z3bridge.Expr, pairs [][2]z3bridge.Expr) z3bridge.Expr {
	if len(pairs) == 0 {
		return t
	}
	from := make([]z3bridge.Expr, len(pairs))
	to := make([]z3bridge.Expr, len(pairs))
	for i, p := range pairs {
		from[i] = p[0]
		to[i] = p[1]
	}
	return ctx.Substitute(t, from, to)
}

// --- Range sort bounds to Z3 ---

// RangeSortBoundsToZ3 converts a RangeSort's lower and upper bounds to Z3
// integer expressions.
// Corresponds to Python's range_sort_bounds_to_z3.
func (s *Solver) RangeSortBoundsToZ3(rs *lg.RangeSort) (lb, ub z3bridge.Expr, err error) {
	// Parse the lower bound
	lbVal, err := strconv.ParseInt(rs.Lb, 10, 64)
	if err != nil {
		return z3bridge.Expr{}, z3bridge.Expr{}, fmt.Errorf("range sort lower bound %q is not an integer: %w", rs.Lb, err)
	}

	// Parse the upper bound
	ubVal, err := strconv.ParseInt(rs.Ub, 10, 64)
	if err != nil {
		return z3bridge.Expr{}, z3bridge.Expr{}, fmt.Errorf("range sort upper bound %q is not an integer: %w", rs.Ub, err)
	}

	return s.tr.Ctx.IntVal(lbVal), s.tr.Ctx.IntVal(ubVal), nil
}

// RangeSortClampedArith returns a Z3 expression for clamped arithmetic on
// range sorts. The result is clamped to [lb, ub].
// This corresponds to the Python code in lookup_native that wraps range
// sort arithmetic:
//
//	lambda x,y: If(x+y > ub, ub, If(x+y < lb, lb, x+y))
// RangeSortClampedAdd returns a Z3 expression for clamped addition:
//   If(x+y > ub, ub, If(x+y < lb, lb, x+y))
// Corresponds to Python ivy_solver.py lookup_native clamped arithmetic.
func (s *Solver) RangeSortClampedAdd(lb, ub, x, y z3bridge.Expr) z3bridge.Expr {
	ctx := s.tr.Ctx
	sum := ctx.Add(x, y)
	// If sum > ub, return ub; else if sum < lb, return lb; else return sum
	return ctx.Ite(ctx.Gt(sum, ub), ub, ctx.Ite(ctx.Lt(sum, lb), lb, sum))
}

// RangeSortClampedSub returns a Z3 expression for clamped subtraction:
//   If(x-y > ub, ub, If(x-y < lb, lb, x-y))
func (s *Solver) RangeSortClampedSub(lb, ub, x, y z3bridge.Expr) z3bridge.Expr {
	ctx := s.tr.Ctx
	diff := ctx.Sub(x, y)
	return ctx.Ite(ctx.Gt(diff, ub), ub, ctx.Ite(ctx.Lt(diff, lb), lb, diff))
}

// RangeSortClampedMul returns a Z3 expression for clamped multiplication:
//   If(x*y > ub, ub, If(x*y < lb, lb, x*y))
func (s *Solver) RangeSortClampedMul(lb, ub, x, y z3bridge.Expr) z3bridge.Expr {
	ctx := s.tr.Ctx
	prod := ctx.Mul(x, y)
	return ctx.Ite(ctx.Gt(prod, ub), ub, ctx.Ite(ctx.Lt(prod, lb), lb, prod))
}

// RangeSortClampedDiv returns a Z3 expression for clamped division:
//
//	If(x/y > ub, ub, If(x/y < lb, lb, x/y))
func (s *Solver) RangeSortClampedDiv(lb, ub, x, y z3bridge.Expr) z3bridge.Expr {
	ctx := s.tr.Ctx
	quot := ctx.Div(x, y)
	return ctx.Ite(ctx.Gt(quot, ub), ub, ctx.Ite(ctx.Lt(quot, lb), lb, quot))
}

// --- Native interpretation lookup ---

// NativeFunc is a function that takes Z3 expressions and returns a Z3 expression.
// This represents a native Z3 operation mapped from an Ivy symbol.
type NativeFunc func(args ...z3bridge.Expr) z3bridge.Expr

// LookupNative resolves the native Z3 interpretation for an Ivy symbol.
// It checks sig.interp for native interpretations, handles polymorphic symbols
// with range sort clamped arithmetic, and recognizes bfe[lo:hi] bit-field extract.
//
// Returns nil if the symbol has no native interpretation.
//
// Corresponds to Python lookup_native (lines 289-324).
func (s *Solver) LookupNative(sym *lg.Symbol, isRelation bool) NativeFunc {
	if s.sig == nil {
		return nil
	}
	ctx := s.tr.Ctx
	name := sym.Name

	// Check sig.interp for a native interpretation
	z3name, hasInterp := s.sig.Interp[name]

	if !hasInterp {
		// Check for bfe[lo:hi] pattern
		if strings.HasPrefix(name, "bfe[") {
			return s.bfeToZ3(sym)
		}

		// Check for arrcst (array constant)
		// Corresponds to Python ivy_solver.py:294-297:
		//   sort = thing.sort.rng
		//   return lambda x: z3.K(sort.to_z3().domain(), x)
		if name == "arrcst" {
			if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
				arrSort := fs.Range()
				z3arrSort, err := s.tr.TranslateSort(arrSort)
				if err == nil && z3arrSort.Kind() == z3bridge.SortArray {
					domSort := z3arrSort.ArrayDomain()
					return func(args ...z3bridge.Expr) z3bridge.Expr {
						if len(args) == 1 {
							return ctx.ConstArray(domSort, args[0])
						}
						return ctx.BoolVal(false)
					}
				}
			}
		}

		// Check for polymorphic symbols (+, -, *, /)
		if isPolymorphicOp(name) {
			return s.lookupPolymorphicNative(sym, isRelation)
		}

		return nil
	}

	// Interpret z3name
	switch interp := z3name.(type) {
	case *lg.EnumeratedSort:
		// Sort interpretation → return the sort itself (used for sort lookups)
		return nil
	case *lg.RangeSort:
		// Sort interpretation → return nil (handled by sort translation)
		return nil
	case string:
		// Named interpretation (e.g., "int", "nat", "bv[32]")
		return s.lookupNamedNative(interp, isRelation)
	}

	return nil
}

// lookupPolymorphicNative handles polymorphic symbols (+, -, *, /) where the
// behavior depends on the domain sort's interpretation.
func (s *Solver) lookupPolymorphicNative(sym *lg.Symbol, isRelation bool) NativeFunc {
	ctx := s.tr.Ctx
	name := sym.Name

	// Get the domain sort
	var domSort lg.Sort
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok && len(fs.Sorts) > 1 {
		domSort = fs.Sorts[0]
	}
	if domSort == nil {
		return nil
	}

	domName := sortToName(domSort)
	interp, ok := s.sig.Interp[domName]
	if !ok {
		return nil
	}

	// Check if interpretation is an EnumeratedSort (no arithmetic)
	if _, isEnum := interp.(*lg.EnumeratedSort); isEnum {
		return nil
	}

	// Handle nat interpretation: subtraction clamps to 0
	if interpStr, ok := interp.(string); ok && interpStr == "nat" && name == "-" {
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Ite(ctx.Lt(args[0], args[1]), ctx.IntVal(0), ctx.Sub(args[0], args[1]))
			}
			return ctx.IntVal(0)
		}
	}

	// Handle range sort: clamped arithmetic
	if rs, ok := interp.(*lg.RangeSort); ok && HandleRangeSorts {
		lb := ctx.IntVal(parseInt64(rs.Lb))
		ub := ctx.IntVal(parseInt64(rs.Ub))
		switch name {
		case "+":
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 2 {
					return s.RangeSortClampedAdd(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			}
		case "-":
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 2 {
					return s.RangeSortClampedSub(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			}
		case "*":
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 2 {
					return s.RangeSortClampedMul(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			}
		case "/":
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 2 {
					return s.RangeSortClampedDiv(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			}
		}
	}

	// Fall back to standard built-in operations
	return s.lookupBuiltinFunc(name, isRelation)
}

// lookupNamedNative resolves a string-named native interpretation (e.g., "int").
func (s *Solver) lookupNamedNative(z3name string, isRelation bool) NativeFunc {
	if isRelation {
		return s.lookupBuiltinRelation(z3name)
	}
	return s.lookupBuiltinFunc(z3name, false)
}

// lookupBuiltinFunc returns the native Z3 function for a built-in name.
func (s *Solver) lookupBuiltinFunc(name string, isRelation bool) NativeFunc {
	ctx := s.tr.Ctx
	switch name {
	case "+":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Add(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "-":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Sub(args[0], args[1])
			}
			if len(args) == 1 {
				return ctx.Sub(ctx.IntVal(0), args[0])
			}
			return ctx.IntVal(0)
		}
	case "*":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Mul(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "/":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Div(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "concat":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Concat(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvand":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.BvAnd(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvor":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.BvOr(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvnot":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 1 {
				return ctx.BvNot(args[0])
			}
			return ctx.BoolVal(false)
		}
	case "arrsel":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Select(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "arrupd":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 3 {
				return ctx.Store(args[0], args[1], args[2])
			}
			return ctx.BoolVal(false)
		}
	}
	return nil
}

// lookupBuiltinRelation returns the native Z3 relation for a built-in name.
// For comparison operators (<, <=, >, >=), dispatches to unsigned BV
// comparisons when operands are bitvectors, matching Python's relations_dict
// which checks z3.is_bv(x) at ivy_solver.py:152-155.
func (s *Solver) lookupBuiltinRelation(name string) NativeFunc {
	ctx := s.tr.Ctx
	switch name {
	case "<":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUlt(args[0], args[1])
				}
				return ctx.Lt(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "<=":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUle(args[0], args[1])
				}
				return ctx.Le(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case ">":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUgt(args[0], args[1])
				}
				return ctx.Gt(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case ">=":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUge(args[0], args[1])
				}
				return ctx.Ge(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "arrsel":
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 2 {
				return ctx.Select(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	}
	return nil
}

// bfeToZ3 creates a bit-field extract function for a bfe[lo:hi] symbol.
// Handles IntSort inputs (via Int2Bv), BV size clamping, zero-width,
// and zero-extension for output sort mismatches.
// Corresponds to Python bfe_to_z3 (lines 174-209).
func (s *Solver) bfeToZ3(sym *lg.Symbol) NativeFunc {
	name := sym.Name
	if !strings.HasPrefix(name, "bfe[") {
		return nil
	}
	inner := name[4:]
	if len(inner) < 2 || inner[len(inner)-1] != ']' {
		return nil
	}
	inner = inner[:len(inner)-1]

	// Parse lo:hi or lo,hi
	var lo, hi int
	sep := strings.IndexAny(inner, ":,")
	if sep < 0 {
		return nil
	}
	if _, err := fmt.Sscanf(inner[:sep], "%d", &lo); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(inner[sep+1:], "%d", &hi); err != nil {
		return nil
	}

	ctx := s.tr.Ctx

	// Get domain and range sorts from the symbol's FunctionSort
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok || fs.Arity() < 1 {
		// Fallback: simple extract
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 1 {
				return ctx.Extract(hi, lo, args[0])
			}
			return ctx.BoolVal(false)
		}
	}

	insort, err1 := s.tr.TranslateSort(fs.Domain()[0])
	outsort, err2 := s.tr.TranslateSort(fs.Range())
	if err1 != nil || err2 != nil {
		return nil
	}

	isIntIn := (insort.Kind() == z3bridge.SortInt)
	isBvIn := (insort.Kind() == z3bridge.SortBV)
	isIntOut := (outsort.Kind() == z3bridge.SortInt)
	isBvOut := (outsort.Kind() == z3bridge.SortBV)

	// Clamp hi for BV inputs whose size <= hi
	if !isIntIn {
		if !isBvIn {
			return nil
		}
		inSize := ctx.BvSortSize(insort)
		if inSize <= hi {
			hi = inSize - 1
		}
	}

	if isIntOut {
		// Output is IntSort
		if hi < lo {
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				return ctx.IntVal(0)
			}
		}
		if isIntIn {
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 1 {
					return ctx.Bv2Int(ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0])), false)
				}
				return ctx.IntVal(0)
			}
		}
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 1 {
				return ctx.Bv2Int(ctx.Extract(hi, lo, args[0]), false)
			}
			return ctx.IntVal(0)
		}
	}

	if !isBvOut {
		return nil
	}

	outSize := ctx.BvSortSize(outsort)
	// Clamp hi if output BV is smaller than extract range
	if outSize < hi-lo+1 {
		hi = lo + outSize - 1
	}

	if hi < lo {
		// Zero-width: return 0 bitvec
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			return ctx.BvVal(0, outSize)
		}
	}

	extractWidth := hi - lo + 1
	if extractWidth < outSize {
		// Need zero-extension
		padWidth := outSize - extractWidth
		if isIntIn {
			return func(args ...z3bridge.Expr) z3bridge.Expr {
				if len(args) == 1 {
					return ctx.Concat(ctx.BvVal(0, padWidth), ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0])))
				}
				return ctx.BvVal(0, outSize)
			}
		}
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 1 {
				return ctx.Concat(ctx.BvVal(0, padWidth), ctx.Extract(hi, lo, args[0]))
			}
			return ctx.BvVal(0, outSize)
		}
	}

	// Exact width match
	if isIntIn {
		return func(args ...z3bridge.Expr) z3bridge.Expr {
			if len(args) == 1 {
				return ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0]))
			}
			return ctx.BvVal(0, outSize)
		}
	}
	return func(args ...z3bridge.Expr) z3bridge.Expr {
		if len(args) == 1 {
			return ctx.Extract(hi, lo, args[0])
		}
		return ctx.BvVal(0, outSize)
	}
}

// SolverName returns the Z3 name for an Ivy symbol, matching Python's
// solver_name (ivy_solver.py:60-78). For polymorphic symbols, appends
// ":domain_sort_name" for each domain sort. Returns "" if the symbol
// has a native Z3 interpretation (should be handled inline, not declared).
func (s *Solver) SolverName(sym *lg.Symbol) string {
	name := sym.Name

	// bfe[lo:hi] — handled natively
	if strings.HasPrefix(name, "bfe[") {
		if s.bfeToZ3(sym) != nil {
			return ""
		}
	}

	// Polymorphic symbols: append domain sort names
	if _, isPoly := iu.PolymorphicSymbols[name]; isPoly {
		fs, isFuncSort := sym.CSort.(*lg.FunctionSort)
		if isFuncSort && len(fs.Domain()) > 0 {
			domSort := fs.Domain()[0]
			domName := sortToName(domSort)
			if name == "arrcst" {
				domName = sortToName(fs.Range())
			}
			if s.sig != nil {
				interp, hasInterp := s.sig.Interp[domName]
				if hasInterp {
					if _, isEnum := interp.(*lg.EnumeratedSort); !isEnum {
						return "" // native interpretation
					}
				}
			}
			for _, d := range fs.Domain() {
				name += ":" + sortToName(d)
			}
			if sym.Name == "arrcst" {
				name += ":" + sortToName(fs.Range())
			}
		}
	}

	// Check if the name itself is interpreted
	if s.sig != nil {
		if _, hasInterp := s.sig.Interp[name]; hasInterp {
			return ""
		}
	}

	// Check Z3 built-in collision.
	// Corresponds to Python's z3_builtins = set(["bit0","bit1"]).
	if z3Builtins[name] {
		// Python raises IvyError here. We return "" to suppress declaration
		// (matching behavior when symbol has native interp).
		return ""
	}

	return name
}

// z3Builtins is the set of names that clash with Z3 built-in symbols.
// Corresponds to Python z3_builtins (ivy_solver.py:58).
var z3Builtins = map[string]bool{
	"bit0": true,
	"bit1": true,
}

// isPolymorphicOp returns true if the name is a polymorphic arithmetic operator.
func isPolymorphicOp(name string) bool {
	switch name {
	case "+", "-", "*", "/":
		return true
	}
	return false
}

// sortToName extracts the sort name from a sort.
func sortToName(s lg.Sort) string {
	if s == nil {
		return ""
	}
	switch st := s.(type) {
	case *lg.UninterpretedSort:
		return st.Name
	case *lg.EnumeratedSort:
		return st.Name
	case *lg.RangeSort:
		return st.Name
	}
	return fmt.Sprint(s)
}

// parseInt64 parses a string to int64, returning 0 on error.
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// HandleRangeSorts controls whether range sort clamped arithmetic is used.
var HandleRangeSorts = true

// --- Batch 2.1/2.2 helper functions ---

// MyMinus creates a Z3 subtraction, handling unary case.
// Corresponds to Python's my_minus (ivy_solver.py:83-86).
func MyMinus(ctx *z3bridge.Context, args []z3bridge.Expr) z3bridge.Expr {
	if len(args) == 1 {
		zero := ctx.IntVal(0)
		return ctx.Sub(zero, args[0])
	}
	if len(args) == 2 {
		return ctx.Sub(args[0], args[1])
	}
	// Chain: a - b - c - ... = ((a - b) - c) - ...
	result := ctx.Sub(args[0], args[1])
	for _, a := range args[2:] {
		result = ctx.Sub(result, a)
	}
	return result
}

// MyEq creates a Z3 equality, handling boolean edge cases.
// If y is true, returns x; if y is false, returns Not(x).
// For boolean args, uses Iff; for other types, uses Eq.
// Corresponds to Python's my_eq (ivy_solver.py:88-95).
func MyEq(ctx *z3bridge.Context, x, y z3bridge.Expr) z3bridge.Expr {
	if y.IsTrue() {
		return x
	}
	if y.IsFalse() {
		return ctx.Not(x)
	}
	if x.ExprSort().Kind() == z3bridge.SortBool {
		return ctx.Iff(x, y)
	}
	return ctx.Eq(x, y)
}

// SortNameToZ3 converts an Ivy sort name to a Z3 sort using the solver's
// translator. Corresponds to Python's sort_name_to_z3 (ivy_solver.py:107).
func (s *Solver) SortNameToZ3(name string) (z3bridge.Sort, error) {
	sort := &lg.UninterpretedSort{Name: name}
	return s.tr.TranslateSort(sort)
}

// Gebin encodes "bits >= n" as a boolean formula over a list of Z3 Bool
// expressions (MSB first). Recursively splits on the MSB.
// Corresponds to Python's gebin (ivy_solver.py:1570-1578).
func Gebin(ctx *z3bridge.Context, bits []z3bridge.Expr, n int) z3bridge.Expr {
	if n == 0 {
		return ctx.BoolVal(true)
	}
	if len(bits) == 0 || n >= (1<<uint(len(bits))) {
		return ctx.BoolVal(false)
	}
	hval := 1 << uint(len(bits)-1)
	if hval <= n {
		return ctx.And(bits[0], Gebin(ctx, bits[1:], n-hval))
	}
	return ctx.Or(bits[0], Gebin(ctx, bits[1:], n))
}
