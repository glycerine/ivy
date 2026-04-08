// Ported to Go from z3_utils.py.
//
// Z3Utils provides a simple, bare recursive translator from logic.py types
// to Z3 objects, plus implication checking functions.
//
// This is independent from Translator (which ports ivy_solver.py's translation).
// Key differences: no LookupNative, no SolverName, no QuantConstraints,
// no EqFunc, no EnumEqFunc, no NumeralFunc, no HASH tracing.
// All Var/Const use "name:str(sort)" naming. Iff maps to z3 equality (==).
package z3bridge

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/logic"
)

// Z3Utils faithfully ports Python z3_utils.py.
// All mutable state that was module-level globals in Python lives here.
type Z3Utils struct {
	Ctx           *Z3Context                // own context, independent from Translator
	toZ3Cache     map[logic.NodeKey]any     // Python _to_z3_cache (Sort, Expr, or FuncDecl)
	uninterpSorts map[logic.NodeKey]Sort    // Python _z3_uninterpreted_sorts
	ImpliesCache  map[[2]logic.NodeKey]bool // Python _implies_cache
}

// NewZ3Utils creates a Z3Utils with a fresh Z3 context and empty caches.
func NewZ3Utils() *Z3Utils {
	return &Z3Utils{
		Ctx:           NewZ3Context(),
		toZ3Cache:     make(map[logic.NodeKey]any),
		uninterpSorts: make(map[logic.NodeKey]Sort),
		ImpliesCache:  make(map[[2]logic.NodeKey]bool),
	}
}

// Close releases the Z3 context.
func (u *Z3Utils) Close() error {
	return u.Ctx.Close()
}

// Clear resets all caches. Corresponds to clearing the module-level
// globals in Python z3_utils.py.
func (u *Z3Utils) Clear() {
	u.toZ3Cache = make(map[logic.NodeKey]any)
	u.uninterpSorts = make(map[logic.NodeKey]Sort)
	u.ImpliesCache = make(map[[2]logic.NodeKey]bool)
}

// --- to_z3 translation ---

// ToZ3 converts a logic term, formula, or sort to a Z3 object.
// Returns Sort, Expr, or FuncDecl depending on the input type.
// Results are cached by structural identity (Sexp()).
// Corresponds to Python z3_utils.py:to_z3 (line 44).
func (u *Z3Utils) ToZ3(x logic.Expr) (any, error) {
	key := x.Sexp()
	if v, ok := u.toZ3Cache[key]; ok {
		return v, nil
	}
	result, err := u.toZ3Internal(x)
	if err != nil {
		return nil, err
	}
	u.toZ3Cache[key] = result
	return result, nil
}

// ToZ3Expr is a convenience wrapper that calls ToZ3 and type-asserts to Expr.
// Used for formulas and terms that produce Z3 expressions.
func (u *Z3Utils) ToZ3Expr(x logic.Expr) (Expr, error) {
	result, err := u.ToZ3(x)
	if err != nil {
		return Expr{}, err
	}
	ze, ok := result.(Expr)
	if !ok {
		return Expr{}, fmt.Errorf("z3_utils.ToZ3Expr: expected Expr for %T, got %T", x, result)
	}
	return ze, nil
}

// toZ3Sort calls ToZ3 on a sort and type-asserts to Sort.
func (u *Z3Utils) toZ3Sort(s logic.Sort) (Sort, error) {
	result, err := u.ToZ3(s)
	if err != nil {
		return Sort{}, err
	}
	zs, ok := result.(Sort)
	if !ok {
		return Sort{}, fmt.Errorf("z3_utils.toZ3Sort: expected Sort for %T, got %T", s, result)
	}
	return zs, nil
}

// toZ3Internal is the recursive implementation of to_z3.
// Corresponds to Python z3_utils.py:_to_z3 (line 53).
func (u *Z3Utils) toZ3Internal(x logic.Expr) (any, error) {

	// Python lines 58-59: _z3_interpreted check
	// Boolean sort → z3.BoolSort(), true → z3.BoolVal(True), false → z3.BoolVal(False)
	switch node := x.(type) {
	case *logic.BooleanSort:
		// Python: _z3_interpreted[Boolean] → z3.BoolSort()
		return u.Ctx.BoolSort(), nil

	case *logic.And:
		if len(node.Terms) == 0 {
			// Python: _z3_interpreted[true] → z3.BoolVal(True)
			return u.Ctx.BoolVal(true), nil
		}
		// Python line 90: _z3_operators[And] → z3.And
		args := make([]Expr, len(node.Terms))
		for i, t := range node.Terms {
			a, err := u.ToZ3Expr(t)
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		return u.Ctx.And(args...), nil

	case *logic.Or:
		if len(node.Terms) == 0 {
			// Python: _z3_interpreted[false] → z3.BoolVal(False)
			return u.Ctx.BoolVal(false), nil
		}
		// Python line 90: _z3_operators[Or] → z3.Or
		args := make([]Expr, len(node.Terms))
		for i, t := range node.Terms {
			a, err := u.ToZ3Expr(t)
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		return u.Ctx.Or(args...), nil

	// Python lines 61-64: UninterpretedSort
	case *logic.UninterpretedSort:
		key := logic.NodeKey(node.Name)
		if cached, ok := u.uninterpSorts[key]; ok {
			return cached, nil
		}
		zs := u.Ctx.UninterpretedSort(node.Name)
		u.uninterpSorts[key] = zs
		return zs, nil

	// Python lines 66-67: FunctionSort → assert False
	case *logic.FunctionSort:
		return nil, fmt.Errorf("z3_utils: FunctionSort's aren't converted to Z3")

	// Python lines 69-81: Var/Const with various sort types
	case *logic.Variable:
		return u.toZ3VarOrConst(node.Name, node.VSort, false)

	case *logic.Const:
		return u.toZ3VarOrConst(node.Name, node.CSort, true)

	// Python lines 83-88: Apply
	case *logic.Apply:
		if len(node.Terms) == 0 {
			// Python line 85: return to_z3(x.func)
			return u.ToZ3(node.Func)
		}
		// Python line 88: return to_z3(x.func)(*(to_z3(t) for t in x.terms))
		funcResult, err := u.ToZ3(node.Func)
		if err != nil {
			return nil, err
		}
		args := make([]Expr, len(node.Terms))
		for i, t := range node.Terms {
			a, err := u.ToZ3Expr(t)
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		switch f := funcResult.(type) {
		case FuncDecl:
			return f.Apply(args...), nil
		default:
			return nil, fmt.Errorf("z3_utils: expected FuncDecl for Apply with %d args, got %T", len(node.Terms), funcResult)
		}

	// Python lines 90-91: _z3_operators dispatch
	case *logic.Eq:
		// Python: _z3_operators[Eq] → lambda t1, t2: t1 == t2
		t1, err := u.ToZ3Expr(node.T1)
		if err != nil {
			return nil, err
		}
		t2, err := u.ToZ3Expr(node.T2)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Eq(t1, t2), nil

	case *logic.Not:
		// Python: _z3_operators[Not] → z3.Not
		body, err := u.ToZ3Expr(node.Body)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Not(body), nil

	case *logic.Implies:
		// Python: _z3_operators[Implies] → z3.Implies
		t1, err := u.ToZ3Expr(node.T1)
		if err != nil {
			return nil, err
		}
		t2, err := u.ToZ3Expr(node.T2)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Implies(t1, t2), nil

	case *logic.Iff:
		// Python: _z3_operators[Iff] → lambda t1, t2: t1 == t2
		// NOTE: maps to Eq (z3 equality), NOT Iff
		t1, err := u.ToZ3Expr(node.T1)
		if err != nil {
			return nil, err
		}
		t2, err := u.ToZ3Expr(node.T2)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Eq(t1, t2), nil

	case *logic.Ite:
		// Python: _z3_operators[Ite] → z3.If
		cond, err := u.ToZ3Expr(node.Cond)
		if err != nil {
			return nil, err
		}
		th, err := u.ToZ3Expr(node.Then)
		if err != nil {
			return nil, err
		}
		el, err := u.ToZ3Expr(node.Else)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Ite(cond, th, el), nil

	// Python lines 93-100: _z3_quantifiers dispatch
	case *logic.ForAll:
		if len(node.Variables) == 0 {
			// Python line 95: return to_z3(x.body)
			return u.ToZ3Expr(node.Body)
		}
		// Python lines 97-100: z3.ForAll([to_z3(v) for v in x.variables], to_z3(x.body))
		bound := make([]Expr, len(node.Variables))
		for i, v := range node.Variables {
			b, err := u.ToZ3Expr(v)
			if err != nil {
				return nil, err
			}
			bound[i] = b
		}
		body, err := u.ToZ3Expr(node.Body)
		if err != nil {
			return nil, err
		}
		return u.Ctx.ForAll(bound, body), nil

	case *logic.Exists:
		if len(node.Variables) == 0 {
			return u.ToZ3Expr(node.Body)
		}
		bound := make([]Expr, len(node.Variables))
		for i, v := range node.Variables {
			b, err := u.ToZ3Expr(v)
			if err != nil {
				return nil, err
			}
			bound[i] = b
		}
		body, err := u.ToZ3Expr(node.Body)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Exists(bound, body), nil

	default:
		// Python line 103: assert False, type(x)
		return nil, fmt.Errorf("z3_utils: unsupported type %T", x)
	}
}

// toZ3VarOrConst handles Python lines 69-81:
// Var/Const with first_order_sort, FunctionSort(arity==0), FunctionSort(arity>=1).
// isConst distinguishes Const from Variable (Python asserts Const for higher-order).
func (u *Z3Utils) toZ3VarOrConst(name string, sort logic.Sort, isConst bool) (any, error) {
	// Python line 69: first_order_sort(x.sort)
	if logic.FirstOrderSort(sort) {
		// Python line 70: z3.Const(x.name + ':' + str(x.sort), to_z3(x.sort))
		zs, err := u.toZ3Sort(sort)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Const(name+":"+sort.String(), zs), nil
	}

	fs, ok := sort.(*logic.FunctionSort)
	if !ok {
		return nil, fmt.Errorf("z3_utils: expected FunctionSort for non-first-order sort, got %T", sort)
	}

	// Python lines 72-75: FunctionSort with len(sorts)==1 → convert to first order
	if fs.Arity() == 0 {
		s := fs.Range()
		zs, err := u.toZ3Sort(s)
		if err != nil {
			return nil, err
		}
		return u.Ctx.Const(name+":"+s.String(), zs), nil
	}

	// Python lines 77-81: FunctionSort with len(sorts)>1 → z3.Function
	// Python: assert type(x) is Const
	if !isConst {
		return nil, fmt.Errorf("z3_utils: cannot convert high-order variables to Z3, only constants")
	}

	// Python line 79-81: z3.Function(x.name, *(to_z3(s) for s in x.sort))
	// x.sort iterates over all sorts (domain + range)
	zSorts := make([]Sort, len(fs.Sorts))
	for i, s := range fs.Sorts {
		zs, err := u.toZ3Sort(s)
		if err != nil {
			return nil, err
		}
		zSorts[i] = zs
	}
	domain := zSorts[:len(zSorts)-1]
	rangeSrt := zSorts[len(zSorts)-1]
	return u.Ctx.Function(name, domain, rangeSrt), nil
}

// --- z3_implies and z3_implies_batch ---

// Z3Implies tests if f1 implies f2 using Z3.
// Corresponds to Python z3_utils.py:z3_implies (line 109).
func (u *Z3Utils) Z3Implies(f1, f2 logic.Expr, timeout bool) (bool, error) {
	key := [2]logic.NodeKey{f1.Sexp(), f2.Sexp()}
	if cached, ok := u.ImpliesCache[key]; ok {
		return cached, nil
	}

	s := u.Ctx.NewSolver()
	if timeout {
		s.SetParam("timeout", "2000") // 2 seconds
	}

	zf1, err := u.ToZ3Expr(f1)
	if err != nil {
		return false, err
	}
	s.Assert(zf1)

	negF2 := &logic.Not{Body: f2}
	zNegF2, err := u.ToZ3Expr(negF2)
	if err != nil {
		return false, err
	}
	s.Assert(zNegF2)

	res := s.Check()
	switch res {
	case Sat:
		u.ImpliesCache[key] = false
		return false, nil
	case Unsat:
		u.ImpliesCache[key] = true
		return true, nil
	default:
		// Python: print("z3 returned: {}".format(res)); assert False
		return false, fmt.Errorf("z3 returned: %s", res)
	}
}

// Z3ImpliesBatch tests if premise implies each formula in formulas.
// Equivalent to [Z3Implies(premise, f) for f in formulas] but more efficient.
// Uses a single Z3 Solver with push/pop for incremental checking.
// Corresponds to Python z3_utils.py:z3_implies_batch (line 136).
func (u *Z3Utils) Z3ImpliesBatch(premise logic.Expr, formulas []logic.Expr, timeout bool) ([]bool, error) {
	s := u.Ctx.NewSolver()
	if timeout {
		s.SetParam("timeout", "2000") // 2 seconds
	}

	zPremise, err := u.ToZ3Expr(premise)
	if err != nil {
		return nil, err
	}
	s.Assert(zPremise)

	result := make([]bool, len(formulas))
	for i, f := range formulas {
		key := [2]logic.NodeKey{premise.Sexp(), f.Sexp()}
		if cached, ok := u.ImpliesCache[key]; ok {
			result[i] = cached
			continue
		}

		s.Push()
		negF := &logic.Not{Body: f}
		zNeg, err := u.ToZ3Expr(negF)
		if err != nil {
			return nil, err
		}
		s.Assert(zNeg)
		res := s.Check()
		s.Pop()

		switch res {
		case Sat:
			u.ImpliesCache[key] = false
			result[i] = false
		case Unsat:
			u.ImpliesCache[key] = true
			result[i] = true
		default:
			// Python: print("z3 returned: {}".format(res)); assert False
			return nil, fmt.Errorf("z3 returned: %s for formula %d", res, i)
		}
	}
	return result, nil
}
