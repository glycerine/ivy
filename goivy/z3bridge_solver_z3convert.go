// z3convert.go provides functions for converting Z3 expressions back to Ivy
// formulas, Craig interpolation, sort ordering, and related utilities.
// Ported from Python ivy_solver.py.
package goivy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Z3 sort → Ivy sort ---

// Z3SortToSort converts a Z3 sort back to an Ivy sort.
// Corresponds to Python's z3sort_to_sort.
func Z3SortToSort(z3sort Z3Sort) Sort {
	xtracer.Trace("ivy_solver.py:1799 z3sort_to_sort() ENTER")
	kind := z3sort.Kind()
	switch kind {
	case SortBool:
		return Boolean
	case SortInt:
		return &UninterpretedSort{Name: "int"}
	case SortReal:
		return &UninterpretedSort{Name: "real"}
	case SortArray:
		domSort := Z3SortToSort(z3sort.ArrayDomain())
		rngSort := Z3SortToSort(z3sort.ArrayRange())
		domName := sortToName(domSort)
		rngName := sortToName(rngSort)
		return &UninterpretedSort{Name: "arr[" + domName + "][" + rngName + "]"}
	default:
		// Uninterpreted or other: use the name
		name := z3sort.String()
		return &UninterpretedSort{Name: name}
	}
}

// --- Z3 func_decl → Ivy symbol ---

// Z3DeclToSymbol converts a Z3 function declaration to an Ivy constant (symbol).
// Corresponds to Python's z3decl_to_symbol.
func Z3DeclToSymbol(z3decl FuncDecl) *Const {
	xtracer.Trace("ivy_solver.py:1805 z3decl_to_symbol() ENTER")
	arity := z3decl.Arity()
	rng := Z3SortToSort(z3decl.RangeSort())

	// Strip the ":sort" suffix from the name (added by Translator)
	name := z3decl.Name()
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[:idx]
	}

	if arity == 0 {
		return NewConst(name, rng)
	}

	dom := make([]Sort, arity)
	for i := 0; i < arity; i++ {
		dom[i] = Z3SortToSort(z3decl.DomainSort(i))
	}
	sortArgs := append(dom, rng)
	fs, err := NewFunctionSort(sortArgs...)
	if err != nil {
		// Fallback: return a const with range sort
		return NewConst(name, rng)
	}
	return NewConst(name, fs)
}

// --- Z3 expression → Ivy formula ---

// Z3ToFormula converts a Z3 expression back to an Ivy formula.
// The vars parameter holds de Bruijn variable bindings (innermost first).
// Corresponds to Python's z3_to_formula.
func Z3ToFormula(z3expr Z3Expr, vars []*Variable) (Expr, error) {
	xtracer.Trace("ivy_solver.py:1817 z3_to_formula() ENTER")
	// Application (includes constants, And, Or, Not, Eq, etc.)
	if z3expr.IsApp() {
		arity := z3expr.NumArgs()

		// Recursively convert arguments
		args := make([]Expr, arity)
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
		case DeclAnd:
			if len(args) == 0 {
				return True, nil
			}
			return &And{Terms: args}, nil

		case DeclOr:
			if len(args) == 0 {
				return False, nil
			}
			return &Or{Terms: args}, nil

		case DeclNot:
			if len(args) != 1 {
				return nil, fmt.Errorf("z3_to_formula: Not with %d args", len(args))
			}
			return &Not{Body: args[0]}, nil

		case DeclEq:
			if len(args) != 2 {
				return nil, fmt.Errorf("z3_to_formula: Eq with %d args", len(args))
			}
			return &Eq{T1: args[0], T2: args[1]}, nil

		case DeclITE:
			if len(args) != 3 {
				return nil, fmt.Errorf("z3_to_formula: ITE with %d args", len(args))
			}
			return &Ite{Cond: args[0], Then: args[1], Else: args[2]}, nil

		case DeclTrue:
			return True, nil

		case DeclFalse:
			return False, nil

		case DeclIff:
			if len(args) != 2 {
				return nil, fmt.Errorf("z3_to_formula: Iff with %d args", len(args))
			}
			return &Iff{T1: args[0], T2: args[1]}, nil

		default:
			// Uninterpreted function/constant
			sym := Z3DeclToSymbol(decl)
			if arity == 0 {
				return sym, nil
			}
			return MustApply(sym, args...), nil
		}
	}

	// Quantifier
	if z3expr.IsQuantifier() {
		nVars := z3expr.QuantNumVars()
		qVars := make([]*Variable, nVars)
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
			v, err := NewVariable(name, sort)
			if err != nil {
				return nil, fmt.Errorf("z3_to_formula: creating var %s: %w", name, err)
			}
			qVars[i] = v
		}

		// Build new vars list: reversed qVars prepended to existing vars
		// (de Bruijn: innermost bindings come first)
		newVars := make([]*Variable, 0, len(qVars)+len(vars))
		for i := len(qVars) - 1; i >= 0; i-- {
			newVars = append(newVars, qVars[i])
		}
		newVars = append(newVars, vars...)

		body, err := Z3ToFormula(z3expr.QuantBody(), newVars)
		if err != nil {
			return nil, err
		}

		if z3expr.IsForAll() {
			return &ForAll{Variables: qVars, Body: body}, nil
		}
		return &Exists{Variables: qVars, Body: body}, nil
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
		return NewConst(s, sort), nil
	}

	return nil, fmt.Errorf("z3_to_formula: cannot convert Z3 expression: %s", z3expr.String())
}

// Z3ToFormulaNoVars is a convenience wrapper that calls Z3ToFormula with no
// initial variable bindings.
func Z3ToFormulaNoVars(z3expr Z3Expr) (Expr, error) {
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
func (s *Solver) BinaryInterpolant(clauses2, clauses1 *Clauses) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:753 binary_interpolant() ENTER")
	// Create a fresh translator with an interpolation-capable context.
	// Z3_compute_interpolant requires a context created via
	// Z3_mk_interpolation_context (legacy solver with proof generation).
	itpTr := s.NewTranslatorWithInterpolation()
	defer itpTr.Close()

	// Wire up native lookups on the interpolation translator so that
	// polymorphic symbols and native interpretations are handled.
	//iptTr.Sorts = s.tr.Sorts
	//iptTr.Relations = s.tr.Relations
	//iptTr.Functions = s.tr.Functions
	//itpTr.LookupNative = s.tr.LookupNative
	//itpTr.SolverName = s.tr.SolverName

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

	return NewClauses([]Expr{ivyFmla}, nil, nil), nil
}

// computeZ3Interpolant computes a Craig interpolant between two Z3 formulas
// using Z3's interpolation API.
// Given a and b where (a AND b) is unsat, returns a formula I such that:
//   - a implies I
//   - I AND b is unsat
//   - I only uses symbols common to both a and b
//
// The ctx must be an interpolation-capable context (created via
// NewInterpolationContext).
// Corresponds to Python's z3.binary_interpolant.
func computeZ3Interpolant(ctx *Z3Context, a, b Z3Expr) (Z3Expr, error) {
	// Build the interpolation pattern: And(Interpolant(a), b)
	// This mirrors Python's z3.binary_interpolant which does:
	//   f = And(Interpolant(a), b)
	//   ti = tree_interpolant(f)
	//   return ti[0]
	marked := ctx.MkInterpolant(a)
	pattern := ctx.And(marked, b)

	interps, err := ctx.ComputeInterpolant(pattern)
	if err != nil {
		return Z3Expr{}, fmt.Errorf("computing interpolant: %w", err)
	}
	if len(interps) == 0 {
		return Z3Expr{}, fmt.Errorf("interpolation returned empty result")
	}
	return interps[0], nil
}

// --- Collect numerals ---

// CollectNumeralsRecursive recursively collects all numeral subterms from a
// Z3 expression. For ITE expressions, it recurses into the then/else branches.
// Corresponds to Python's collect_numerals.
func CollectNumeralsRecursive(z3term Z3Expr) []Z3Expr {
	var result []Z3Expr
	collectNumeralsHelper(z3term, &result)
	return result
}

func collectNumeralsHelper(z3term Z3Expr, result *[]Z3Expr) {
	// Check if it's a numeral (int value or bv value)
	if z3term.IsNumeral() {
		*result = append(*result, z3term)
		return
	}

	// Check if it's an ITE application — recurse into then/else branches
	if z3term.IsApp() && z3term.IsAppOf(DeclITE) {
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
func FromZ3Numeral(z3term Z3Expr, sort Sort) *Const {
	xtracer.Trace("ivy_solver.py:877 from_z3_numeral() ENTER sort=%s", sort)
	name := z3term.String()
	if len(name) == 0 {
		return NewConst("0", sort)
	}
	// Validate: should start with digit, quote, or minus
	if !(name[0] >= '0' && name[0] <= '9' || name[0] == '"' || name[0] == '-') {
		fmt.Printf("warning: unexpected numeral from Z3 model: %s\n", name)
	}
	return NewConst(name, sort)
}

// --- Collect model values ---

// CollectModelValuesZ3 collects all model values for a symbol of a given sort
// from a Z3 model. Uses sym_placeholders to create a term, evaluates it in
// the model, and collects the numerals.
// Corresponds to Python's collect_model_values.
func (s *Solver) CollectModelValuesZ3(sort Sort, model *Model, sym *Const) map[string]*Const {
	result := make(map[string]*Const)

	// Create the term: sym(V0, V1, ...)
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
type Z3SortOrder struct {
	Vs    []Z3Expr   // Z3 variables for the order relation
	Order Z3Expr     // Z3 expression representing the order (e.g., less-than)
	Model *Model     // Z3 model for evaluation
	Ctx   *Z3Context // Z3 context for substitution
}

// NewSortOrder creates a new SortOrder.
func NewZ3SortOrder(vs []Z3Expr, order Z3Expr, model *Model, ctx *Z3Context) *Z3SortOrder {
	return &Z3SortOrder{
		Vs:    vs,
		Order: order,
		Model: model,
		Ctx:   ctx,
	}
}

// Compare returns -1 if x < y according to the order, +1 otherwise.
// This implements a comparison function suitable for sorting.
// Corresponds to Python's SortOrder.__call__.
func (so *Z3SortOrder) Compare(x, y Z3Expr) int {
	xtracer.Trace("ivy_solver.py:856 SortOrder.compare() ENTER")
	if len(so.Vs) < 2 {
		return 0
	}

	// Substitute vs[0]->x, vs[1]->y into the order expression
	from := so.Vs[:2]
	to := []Z3Expr{x, y}
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
func SubstituteZ3(ctx *Z3Context, t Z3Expr, pairs [][2]Z3Expr) Z3Expr {
	xtracer.Trace("ivy_solver.py:1788 substitute() ENTER")
	if len(pairs) == 0 {
		return t
	}
	from := make([]Z3Expr, len(pairs))
	to := make([]Z3Expr, len(pairs))
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
func (s *Solver) RangeSortBoundsToZ3(rs *RangeSort) (lb, ub Z3Expr, err error) {
	xtracer.Trace("ivy_solver.py:299 range_sort_bounds_to_z3() ENTER")
	// Parse the lower bound
	lbVal, err := strconv.ParseInt(rs.LbString(), 10, 64)
	if err != nil {
		return Z3Expr{}, Z3Expr{}, fmt.Errorf("range sort lower bound %q is not an integer: %w", rs.LbString(), err)
	}

	// Parse the upper bound
	ubVal, err := strconv.ParseInt(rs.UbString(), 10, 64)
	if err != nil {
		return Z3Expr{}, Z3Expr{}, fmt.Errorf("range sort upper bound %q is not an integer: %w", rs.UbString(), err)
	}

	return s.tr.Ctx.IntVal(lbVal), s.tr.Ctx.IntVal(ubVal), nil
}

// RangeSortClampedArith returns a Z3 expression for clamped arithmetic on
// range sorts. The result is clamped to [lb, ub].
// This corresponds to the Python code in lookup_native that wraps range
// sort arithmetic:
//
//	lambda x,y: If(x+y > ub, ub, If(x+y < lb, lb, x+y))
//
// RangeSortClampedAdd returns a Z3 expression for clamped addition:
//
//	If(x+y > ub, ub, If(x+y < lb, lb, x+y))
//
// Corresponds to Python ivy_solver.py lookup_native clamped arithmetic.
func (s *Solver) RangeSortClampedAdd(lb, ub, x, y Z3Expr) Z3Expr {
	ctx := s.tr.Ctx
	sum := ctx.Add(x, y)
	// If sum > ub, return ub; else if sum < lb, return lb; else return sum
	return ctx.Ite(ctx.Gt(sum, ub), ub, ctx.Ite(ctx.Lt(sum, lb), lb, sum))
}

// RangeSortClampedSub returns a Z3 expression for clamped subtraction:
//
//	If(x-y > ub, ub, If(x-y < lb, lb, x-y))
func (s *Solver) RangeSortClampedSub(lb, ub, x, y Z3Expr) Z3Expr {
	ctx := s.tr.Ctx
	diff := ctx.Sub(x, y)
	return ctx.Ite(ctx.Gt(diff, ub), ub, ctx.Ite(ctx.Lt(diff, lb), lb, diff))
}

// RangeSortClampedMul returns a Z3 expression for clamped multiplication:
//
//	If(x*y > ub, ub, If(x*y < lb, lb, x*y))
func (s *Solver) RangeSortClampedMul(lb, ub, x, y Z3Expr) Z3Expr {
	ctx := s.tr.Ctx
	prod := ctx.Mul(x, y)
	return ctx.Ite(ctx.Gt(prod, ub), ub, ctx.Ite(ctx.Lt(prod, lb), lb, prod))
}

// RangeSortClampedDiv returns a Z3 expression for clamped division:
//
//	If(x/y > ub, ub, If(x/y < lb, lb, x/y))
func (s *Solver) RangeSortClampedDiv(lb, ub, x, y Z3Expr) Z3Expr {
	ctx := s.tr.Ctx
	quot := ctx.Div(x, y)
	return ctx.Ite(ctx.Gt(quot, ub), ub, ctx.Ite(ctx.Lt(quot, lb), lb, quot))
}

// --- Native interpretation lookup ---

// NativeFunc is a function that takes Z3 expressions and returns a Z3 expression.
// This represents a native Z3 operation mapped from an Ivy symbol.
type NativeFunc func(args ...Z3Expr) Z3Expr

// Sorts resolves a sort interpretation name to a Z3 sort.
// Corresponds to Python sorts() (ivy_solver.py:120).
func (s *Solver) Sorts(name string) any {
	xtracer.Trace("ivy_solver.py:121 sorts() ENTER name=%s", name)
	ctx := s.tr.Ctx
	switch name {
	case "nat", "int":
		return ctx.IntSort()
	case "real":
		return ctx.RealSort()
	case "strlit":
		return ctx.StringSort()
	}
	base, params, ok := ParseIntParams(name)
	if ok && len(params) > 0 {
		switch base {
		case "bv", "strbv", "intbv":
			return ctx.BvSort(params[0])
		}
	}
	if dom, rng, ok2 := ParseArraySortName(name); ok2 {
		domS := &UninterpretedSort{Name: dom}
		rngS := &UninterpretedSort{Name: rng}
		xtracer.Trace("TranslateSort_call callsite=sort_name_to_z3 HASH canon=%s", domS.Sexp())
		domSort, err1 := s.tr.TranslateSort(domS)
		xtracer.Trace("TranslateSort_call callsite=sort_name_to_z3 HASH canon=%s", rngS.Sexp())
		rngSort, err2 := s.tr.TranslateSort(rngS)
		if err1 == nil && err2 == nil {
			return ctx.ArraySort(domSort, rngSort)
		}
	}
	return nil
}

// Relations resolves a relation name to a native Z3 function.
// Corresponds to Python relations() (ivy_solver.py:171).
func (s *Solver) Relations(name string) any {
	xtracer.Trace("ivy_solver.py:172 relations() ENTER name=%s", name)
	return s.lookupBuiltinRelation(name)
}

// Functions resolves a function name to a native Z3 function.
// Corresponds to Python functions() (ivy_solver.py:227).
func (s *Solver) Functions(name string) any {
	xtracer.Trace("ivy_solver.py:228 functions() ENTER name=%s", name)
	return s.lookupBuiltinFunc(name, false)
}

// LookupNative resolves the native Z3 interpretation for an Ivy symbol.
// Corresponds to Python lookup_native(thing, table, kind) (ivy_solver.py:311).
//
// Parameters:
//   - thing: the Ivy symbol being looked up
//   - table: one of Sorts, Relations, or Functions
//   - kind: "sort", "relation", or "function"
//
// Returns any: Sort for sort lookups, NativeFunc for function/relation
// lookups, or nil if no native interpretation.
func (s *Solver) LookupNative(thing *Const, table func(string) any, kind string) any {
	//xtracer.Trace("ivy_solver.py:312 lookup_native() ENTER name=%s kind=%s", thing.Name, kind)
	if s.sig == nil {
		return nil
	}
	ctx := s.tr.Ctx
	name := thing.Name

	// Python line 313: z3name = ivy_logic.sig.interp.get(thing.name)
	z3name, hasInterp := s.sig.Interp[name]

	// Python line 314: if z3name == None:
	if !hasInterp {
		// Python line 315: if thing.name.startswith('bfe['):
		if strings.HasPrefix(name, "bfe[") {
			return s.bfeToZ3(thing)
		}

		// Python line 317-320: if thing.name == 'arrcst':
		//   sort = thing.sort.rng
		//   if sort.name in ivy_logic.sig.interp:
		//     return lambda x: z3.K(sort.to_z3().domain(), x)
		if name == "arrcst" {
			if fs, ok := thing.CSort.(*FunctionSort); ok {
				rngSort := fs.Range()
				rngName := sortToName(rngSort)
				if _, inInterp := s.sig.Interp[rngName]; inInterp {
					z3arrSort, err := s.tr.TranslateSort(rngSort)
					if err == nil && z3arrSort.Kind() == SortArray {
						domSort := z3arrSort.ArrayDomain()
						return NativeFunc(func(args ...Z3Expr) Z3Expr {
							if len(args) == 1 {
								return ctx.ConstArray(domSort, args[0])
							}
							return ctx.BoolVal(false)
						})
					}
				}
			}
		}

		// Python line 321: if thing.name in iu.polymorphic_symbols:
		if isPolymorphicOp(name) {
			return s.lookupPolymorphicNative(thing, table)
		}

		return nil
	}

	// Python line 342-343: if isinstance(z3name, (EnumeratedSort, RangeSort)):
	//   return z3name.to_z3()
	switch v := z3name.(type) {
	case *EnumeratedSort:
		xtracer.Trace("TranslateSort_call callsite=lookup_native_enum_or_range HASH canon=%s", v.Sexp())
		zs, err := s.tr.TranslateSort(v)
		if err != nil {
			return nil
		}
		return zs
	case *RangeSort:
		xtracer.Trace("TranslateSort_call callsite=lookup_native_enum_or_range HASH canon=%s", v.Sexp())
		zs, err := s.tr.TranslateSort(v)
		if err != nil {
			return nil
		}
		return zs
	case string:
		// Python line 344: z3val = table(z3name)
		return table(v)
	}

	return nil
}

// lookupPolymorphicNative handles polymorphic symbols (+, -, *, /) where the
// behavior depends on the domain sort's interpretation.
// Corresponds to Python lookup_native lines 321-340.
func (s *Solver) lookupPolymorphicNative(sym *Const, table func(string) any) any {
	ctx := s.tr.Ctx
	name := sym.Name

	// Python line 322: sort = thing.sort.domain[0].name
	var domSort Sort
	if fs, ok := sym.CSort.(*FunctionSort); ok && len(fs.Sorts) > 1 {
		domSort = fs.Sorts[0]
	}
	if domSort == nil {
		return nil
	}

	// Python line 323: if sort in sig.interp and not isinstance(sig.interp[sort], EnumeratedSort):
	domName := sortToName(domSort)
	interp, ok := s.sig.Interp[domName]
	if !ok {
		return nil
	}
	if _, isEnum := interp.(*EnumeratedSort); isEnum {
		return nil
	}

	// Python line 325: if thing.name == '-' and itp == 'nat':
	if interpStr, ok := interp.(string); ok && interpStr == "nat" && name == "-" {
		return NativeFunc(func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Ite(ctx.Lt(args[0], args[1]), ctx.IntVal(0), ctx.Sub(args[0], args[1]))
			}
			return ctx.IntVal(0)
		})
	}

	// Python line 327-336: if handle_range_sorts and isinstance(itp, RangeSort):
	if rs, ok := interp.(*RangeSort); ok && s.HandleRangeSorts {
		lb := ctx.IntVal(parseInt64(rs.LbString()))
		ub := ctx.IntVal(parseInt64(rs.UbString()))
		switch name {
		case "+":
			return NativeFunc(func(args ...Z3Expr) Z3Expr {
				if len(args) == 2 {
					return s.RangeSortClampedAdd(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			})
		case "-":
			return NativeFunc(func(args ...Z3Expr) Z3Expr {
				if len(args) == 2 {
					return s.RangeSortClampedSub(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			})
		case "*":
			return NativeFunc(func(args ...Z3Expr) Z3Expr {
				if len(args) == 2 {
					return s.RangeSortClampedMul(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			})
		case "/":
			return NativeFunc(func(args ...Z3Expr) Z3Expr {
				if len(args) == 2 {
					return s.RangeSortClampedDiv(lb, ub, args[0], args[1])
				}
				return ctx.IntVal(0)
			})
		}
	}

	// Python line 337-339: z3val = table(thing.name)
	//   if z3val == None: raise iu.IvyError(None, '{} is not a supported Z3 {}'.format(thing.name, kind))
	z3val := table(sym.Name)
	if z3val == nil {
		panic(fmt.Sprintf("%s is not a supported Z3 function", sym.Name))
	}
	return z3val
}

// lookupBuiltinFunc returns the native Z3 function for a built-in name.
func (s *Solver) lookupBuiltinFunc(name string, isRelation bool) NativeFunc {
	ctx := s.tr.Ctx
	switch name {
	case "+":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Add(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "-":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Sub(args[0], args[1])
			}
			if len(args) == 1 {
				return ctx.Sub(ctx.IntVal(0), args[0])
			}
			return ctx.IntVal(0)
		}
	case "*":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Mul(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "/":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Div(args[0], args[1])
			}
			return ctx.IntVal(0)
		}
	case "concat":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Concat(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvand":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.BvAnd(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvor":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.BvOr(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "bvnot":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 1 {
				return ctx.BvNot(args[0])
			}
			return ctx.BoolVal(false)
		}
	case "arrsel":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Select(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "arrupd":
		return func(args ...Z3Expr) Z3Expr {
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
//
// Python's z3py comparison operators have a subclass dispatch quirk:
// when `x < y` is evaluated and y's type (e.g. IntNumRef) is a proper
// subclass of x's type (ArithRef), Python tries y's reflected method
// first. So `x < IntVal(0)` calls `IntVal(0).__gt__(x)` → Z3_mk_gt(0,x)
// instead of `x.__lt__(IntVal(0))` → Z3_mk_lt(x,0). These produce
// different Z3 ASTs. We must match this behavior.
func (s *Solver) lookupBuiltinRelation(name string) NativeFunc {
	ctx := s.tr.Ctx
	switch name {
	case "<":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUlt(args[0], args[1])
				}
				if !args[0].IsNumeral() && args[1].IsNumeral() {
					return ctx.Gt(args[1], args[0])
				}
				return ctx.Lt(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "<=":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUle(args[0], args[1])
				}
				if !args[0].IsNumeral() && args[1].IsNumeral() {
					return ctx.Ge(args[1], args[0])
				}
				return ctx.Le(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case ">":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUgt(args[0], args[1])
				}
				if !args[0].IsNumeral() && args[1].IsNumeral() {
					return ctx.Lt(args[1], args[0])
				}
				return ctx.Gt(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case ">=":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				if ctx.IsBvExpr(args[0]) {
					return ctx.BvUge(args[0], args[1])
				}
				if !args[0].IsNumeral() && args[1].IsNumeral() {
					return ctx.Le(args[1], args[0])
				}
				return ctx.Ge(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	case "arrsel":
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 2 {
				return ctx.Select(args[0], args[1])
			}
			return ctx.BoolVal(false)
		}
	}
	return nil
}

// bfeToZ3 creates a bit-field extract function for a bfe[lo][hi] symbol.
// Handles IntSort inputs (via Int2Bv), BV size clamping, zero-width,
// and zero-extension for output sort mismatches.
// Corresponds to Python bfe_to_z3 (lines 174-209).
// Python uses parse_int_params which parses bfe[lo][hi] format.
func (s *Solver) bfeToZ3(sym *Const) NativeFunc {
	xtracer.Trace("ivy_solver.py:188 bfe_to_z3() ENTER sym=%s", sym.Name)
	name := sym.Name
	if !strings.HasPrefix(name, "bfe[") {
		return nil
	}
	// Use ParseIntParams to match Python's parse_int_params format: bfe[lo][hi]
	base, params, ok := ParseIntParams(name)
	if ok && base == "bfe" && len(params) == 2 {
		// bfe[lo][hi] format (Python standard)
	} else {
		// Fallback: try bfe[lo:hi] or bfe[lo,hi] format
		inner := name[4:]
		if len(inner) < 2 || inner[len(inner)-1] != ']' {
			return nil
		}
		inner = inner[:len(inner)-1]
		sep := strings.IndexAny(inner, ":,")
		if sep < 0 {
			return nil
		}
		params = make([]int, 2)
		if _, err := fmt.Sscanf(inner[:sep], "%d", &params[0]); err != nil {
			return nil
		}
		if _, err := fmt.Sscanf(inner[sep+1:], "%d", &params[1]); err != nil {
			return nil
		}
	}
	lo, hi := params[0], params[1]

	ctx := s.tr.Ctx

	// Get domain and range sorts from the symbol's FunctionSort
	fs, ok := sym.CSort.(*FunctionSort)
	if !ok || fs.Arity() < 1 {
		// Fallback: simple extract
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 1 {
				return ctx.Extract(hi, lo, args[0])
			}
			return ctx.BoolVal(false)
		}
	}

	xtracer.Trace("TranslateSort_call callsite=bfe_to_z3_dom HASH canon=%s", fs.Domain()[0].Sexp())
	insort, err1 := s.tr.TranslateSort(fs.Domain()[0])
	xtracer.Trace("TranslateSort_call callsite=bfe_to_z3_rng HASH canon=%s", fs.Range().Sexp())
	outsort, err2 := s.tr.TranslateSort(fs.Range())
	if err1 != nil || err2 != nil {
		return nil
	}

	isIntIn := (insort.Kind() == SortInt)
	isBvIn := (insort.Kind() == SortBV)
	isIntOut := (outsort.Kind() == SortInt)
	isBvOut := (outsort.Kind() == SortBV)

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
			return func(args ...Z3Expr) Z3Expr {
				return ctx.IntVal(0)
			}
		}
		if isIntIn {
			return func(args ...Z3Expr) Z3Expr {
				if len(args) == 1 {
					return ctx.Bv2Int(ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0])), false)
				}
				return ctx.IntVal(0)
			}
		}
		return func(args ...Z3Expr) Z3Expr {
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
		return func(args ...Z3Expr) Z3Expr {
			return ctx.BvVal(0, outSize)
		}
	}

	extractWidth := hi - lo + 1
	if extractWidth < outSize {
		// Need zero-extension
		padWidth := outSize - extractWidth
		if isIntIn {
			return func(args ...Z3Expr) Z3Expr {
				if len(args) == 1 {
					return ctx.Concat(ctx.BvVal(0, padWidth), ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0])))
				}
				return ctx.BvVal(0, outSize)
			}
		}
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 1 {
				return ctx.Concat(ctx.BvVal(0, padWidth), ctx.Extract(hi, lo, args[0]))
			}
			return ctx.BvVal(0, outSize)
		}
	}

	// Exact width match
	if isIntIn {
		return func(args ...Z3Expr) Z3Expr {
			if len(args) == 1 {
				return ctx.Extract(hi, lo, ctx.Int2Bv(hi+1, args[0]))
			}
			return ctx.BvVal(0, outSize)
		}
	}
	return func(args ...Z3Expr) Z3Expr {
		if len(args) == 1 {
			return ctx.Extract(hi, lo, args[0])
		}
		return ctx.BvVal(0, outSize)
	}
}

// SolverName returns the Z3 name for an Ivy symbol, or "" if the symbol
// is natively interpreted. Matches Python's module-level solver_name()
// (ivy_solver.py:65-84). This is a package-level function because
// Python's solver_name is a module-level function, not a Solver method.
//
// bfeCheck, when non-nil, returns true if bfe_to_z3(sym) would return
// a non-nil native function. Pass nil when no Z3 translator is available
// (e.g. at compile time).
//
// Returns ("", nil) for interpreted symbols (Python returns None).
// Returns (name, nil) for non-interpreted symbols.
// Returns ("", error) for z3 builtin clashes (Python raises IvyError).
func SolverName(sym *Const, sig *Sig, bfeCheck func(*Const) bool) (string, error) {
	xtracer.Trace("ivy_solver.py:65 solver_name() ENTER name=%s", sym.Name)
	//xtracer.Trace("ivy_solver.py:65 solver_name() ENTER name=%s\n go-caller='%v'; sym.Canon = '%v'", sym.Name, caller(4), sym.Canon())

	name := sym.Name

	// Python: if name.startswith('bfe['):
	if strings.HasPrefix(name, "bfe[") {
		if bfeCheck != nil && bfeCheck(sym) {
			xtracer.Trace("ivy_solver.py:69 solver_name() EXIT 1")
			return "", nil // interpreted
		}
	} else if _, isPoly := PolymorphicSymbols[name]; isPoly {
		// Python: elif name in iu.polymorphic_symbols:
		if sig != nil {
			fs, isFuncSort := sym.CSort.(*FunctionSort)
			if isFuncSort && len(fs.Domain()) > 0 {
				domName := sortToName(fs.Domain()[0])
				if name == "arrcst" {
					domName = sortToName(fs.Range())
				}
				if interp, has := sig.Interp[domName]; has {
					if _, isEnum := interp.(*EnumeratedSort); !isEnum {
						xtracer.Trace("ivy_solver.py:74 solver_name() EXIT 2")
						return "", nil // native interpretation
					}
				}
				// Python: for s in symbol.sort.domain: name += ':' + s.name
				for _, d := range fs.Domain() {
					name += ":" + sortToName(d)
				}
				if sym.Name == "arrcst" {
					name += ":" + sortToName(fs.Range())
				}
			}
		}
	}

	// Python: if name in ivy_logic.sig.interp: return None
	if sig != nil {
		if _, has := sig.Interp[name]; has {
			xtracer.Trace("ivy_solver.py:82 solver_name() EXIT 3")
			return "", nil // interpreted
		}
	}

	// Python: if name in z3_builtins: raise iu.IvyError(...)
	if z3Builtins[name] {
		xtracer.Trace("ivy_solver.py:85 solver_name() EXIT 4")
		return "", NewIvyError(nil, fmt.Sprintf(`name "%s" clashes with Z3 built-in`, name))
	}
	xtracer.Trace("ivy_solver.py:87 solver_name() EXIT 5")
	return name, nil
}

// SolverName on *Solver is a thin wrapper around the package-level
// SolverName, providing the Solver's sig and bfeToZ3 capability.
// Preserves existing string return + panic-on-error behavior.
func (s *Solver) SolverName(sym *Const) string {
	//vv("Solver.SolverName called with sym='%v'; caller='%v'", sym.Canon(), caller(1))
	bfeCheck := func(c *Const) bool { return s.bfeToZ3(c) != nil }
	n, err := SolverName(sym, s.sig, bfeCheck)
	if err != nil {
		panic(err) // matches Python's raise; existing callers expect panic
	}
	return n
}

// z3Builtins is the set of names that clash with Z3 built-in symbols.
// Corresponds to Python z3_builtins (ivy_solver.py:58).
var z3Builtins = map[string]bool{
	"bit0": true,
	"bit1": true,
}

// isPolymorphicOp returns true if the name is a polymorphic operator.
// Matches Python iu.polymorphic_symbols (ivy_utils.py:696-714).
func isPolymorphicOp(name string) bool {
	switch name {
	case "<", "<=", ">", ">=",
		"+", "*", "-", "/", "*>",
		"bvand", "bvor", "bvnot",
		"cast",
		"arrsel", "arrupd", "arrcst":
		return true
	}
	return false
}

// sortToName extracts the sort name from a sort.
func sortToName(s Sort) string {
	if s == nil {
		return ""
	}
	switch st := s.(type) {
	case *UninterpretedSort:
		return st.Name
	case *EnumeratedSort:
		return st.Name
	case *RangeSort:
		return st.Name
	}
	return fmt.Sprint(s)
}

// parseInt64 parses a string to int64, returning 0 on error.
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// --- Batch 2.1/2.2 helper functions ---

// MyMinus creates a Z3 subtraction, handling unary case.
// Corresponds to Python's my_minus (ivy_solver.py:83-86).
func MyMinus(ctx *Z3Context, args []Z3Expr) Z3Expr {
	//xtracer.Trace("ivy_solver.py:89 my_minus() ENTER nargs=%d", len(args))
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
// Corresponds to Python's my_eq (ivy_solver.py:99-107).
// Python uses Z3_mk_eq unconditionally (even for Bool args), so we
// must do the same — using Z3_mk_iff for Bool would produce a Z3
// decl with name "iff" instead of "=", which then breaks the z3.check
// canon hash comparison between Go and Python (the underlying logic
// is equivalent, but the AST decl is different).
func MyEq(ctx *Z3Context, x, y Z3Expr) Z3Expr {
	xtracer.Trace("ivy_solver.py:95 my_eq() ENTER")
	if y.IsTrue() {
		return x
	}
	if y.IsFalse() {
		return ctx.Not(x)
	}
	return ctx.Eq(x, y)
}

// SortNameToZ3 converts an Ivy sort name to a Z3 sort using the solver's
// translator. Corresponds to Python's sort_name_to_z3 (ivy_solver.py:107).
func (s *Solver) SortNameToZ3(name string) (Z3Sort, error) {
	xtracer.Trace("ivy_solver.py:116 sort_name_to_z3() ENTER name=%s", name)
	sort := &UninterpretedSort{Name: name}
	xtracer.Trace("TranslateSort_call callsite=sort_name_to_z3 HASH canon=%s", sort.Sexp())
	return s.tr.TranslateSort(sort)
}

// Gebin encodes "bits >= n" as a boolean formula over a list of Z3 Bool
// expressions (MSB first). Recursively splits on the MSB.
// Corresponds to Python's gebin (ivy_solver.py:1570-1578).
func Gebin(ctx *Z3Context, bits []Z3Expr, n int) Z3Expr {
	xtracer.Trace("ivy_solver.py:1722 gebin() ENTER n=%d", n)
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
