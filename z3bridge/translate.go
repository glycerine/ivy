package z3bridge

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/logic"
)

// NativeLookupFunc is a callback that looks up native Z3 interpretations
// for Ivy symbols. If the symbol has a native interpretation, it returns
// a function that maps Z3 arguments to a Z3 result. Returns nil if the
// symbol has no native interpretation.
type NativeLookupFunc func(name string, sort logic.Sort, isRelation bool) func(args ...Expr) Expr

// SolverNameFunc maps an Ivy symbol to its Z3 name. Returns "" if the
// symbol should be handled natively (not declared as an uninterpreted function).
type SolverNameFunc func(name string, sort logic.Sort) string

// QuantConstraintsFn is a callback that generates Z3 constraints for
// quantifier-bound variables based on their sort (e.g., nat non-negativity,
// range sort bounds). Corresponds to Python's quant_constraints.
type QuantConstraintsFn func(v *logic.Variable, z3Var Expr) []Expr

// SortLookupFunc is a callback that resolves an Ivy sort name to a Z3 sort.
// Used for interpreted sorts (nat→IntSort, bv[N]→BitVecSort, etc.).
// Returns nil Sort if the sort should be handled by the default TranslateSort.
type SortLookupFunc func(sortName string) *Sort

// Translator converts Ivy logic nodes to Z3 expressions.
type Translator struct {
	Ctx              *Z3Context
	sorts            map[string]Sort       // cache: Ivy sort name -> Z3 sort
	sortsInv         map[string]logic.Sort // reverse map: Z3 sort name -> original Ivy sort
	consts           map[string]Expr       // cache: "name:sort" -> Z3 const
	funcs            map[string]FuncDecl   // cache: "name:sort" -> Z3 func decl
	NativeLookup     NativeLookupFunc      // optional: native interpretation callback
	SolverName       SolverNameFunc        // optional: maps symbol to Z3 name (for polymorphic disambiguation)
	QuantConstraints QuantConstraintsFn    // optional: generates sort constraints for quantifier-bound variables
	SortLookup       SortLookupFunc        // optional: resolves interpreted sort names to Z3 sorts
}

// NewTranslator creates a translator with a fresh Z3 context.
func NewTranslator() *Translator {
	return &Translator{
		Ctx:      NewZ3Context(),
		sorts:    make(map[string]Sort),
		sortsInv: make(map[string]logic.Sort),
		consts:   make(map[string]Expr),
		funcs:    make(map[string]FuncDecl),
	}
}

func (t *Translator) Close() error {
	return t.Ctx.Close()
}

// Clear resets all Z3 caches to initial state.
// Corresponds to Python ivy_solver.clear() (line 228).
func (t *Translator) Clear() {
	t.sorts = make(map[string]Sort)
	t.sortsInv = make(map[string]logic.Sort)
	t.consts = make(map[string]Expr)
	t.funcs = make(map[string]FuncDecl)
}

// SortFromZ3 looks up the original Ivy sort for a Z3 sort using the reverse map.
// Corresponds to Python's sort_from_z3() (ivy_solver.py:905).
func (t *Translator) SortFromZ3(z3sort Sort) (logic.Sort, bool) {
	ivySort, ok := t.sortsInv[z3sort.String()]
	return ivySort, ok
}

// z3Name returns the Z3 name for a symbol. If SolverName is set, uses it;
// otherwise returns the plain name.
func (t *Translator) z3Name(name string, sort logic.Sort) string {
	if t.SolverName != nil {
		return t.SolverName(name, sort)
	}
	return name
}

// TranslateSort converts an Ivy sort to a Z3 sort.
func (t *Translator) TranslateSort(s logic.Sort) (Sort, error) {
	switch st := s.(type) {
	case *logic.BooleanSort:
		return t.Ctx.BoolSort(), nil

	case *logic.UninterpretedSort:
		key := s.Sexp()
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		// Check for interpreted sort via callback (nat→IntSort, bv[N]→BitVecSort, etc.)
		// Corresponds to Python ivy_solver.py:111-135 sorts() function and
		// lookup_native(term.sort, sorts, "sort") in term_to_z3.
		if t.SortLookup != nil {
			if zs := t.SortLookup(st.Name); zs != nil {
				t.sorts[key] = *zs
				t.sortsInv[zs.String()] = s
				return *zs, nil
			}
		}
		// Check for array sort: arr[domain][range]
		// Corresponds to Python ivy_solver.py:115-120 sorts() function.
		if dom, rng, ok := ParseArraySortName(st.Name); ok {
			domSort, err := t.TranslateSort(&logic.UninterpretedSort{Name: dom})
			if err != nil {
				return Sort{}, fmt.Errorf("array domain sort %q: %w", dom, err)
			}
			rngSort, err := t.TranslateSort(&logic.UninterpretedSort{Name: rng})
			if err != nil {
				return Sort{}, fmt.Errorf("array range sort %q: %w", rng, err)
			}
			zs := t.Ctx.ArraySort(domSort, rngSort)
			t.sorts[key] = zs
			t.sortsInv[zs.String()] = s
			return zs, nil
		}
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts[key] = zs
		t.sortsInv[zs.String()] = s
		return zs, nil

	case *logic.TopSort:
		// TopSort is treated as uninterpreted in Z3
		key := s.Sexp()
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts[key] = zs
		t.sortsInv[zs.String()] = s
		return zs, nil

	case *logic.FunctionSort:
		return Sort{}, fmt.Errorf("FunctionSorts are not directly converted to Z3 sorts")

	case *logic.EnumeratedSort:
		key := s.Sexp()
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		// Use native Z3 EnumSort, matching Python's z3.EnumSort(name, extension).
		zs, constExprs := t.Ctx.EnumSort(st.Name, st.Extension)
		t.sorts[key] = zs
		t.sortsInv[zs.String()] = s
		// Register the constructor constants so they can be looked up by name.
		for i, name := range st.Extension {
			constKey := name + ":" + s.Sexp()
			t.consts[constKey] = constExprs[i]
		}
		return zs, nil

	case *logic.RangeSort:
		// Range sorts map to integers
		zs := t.Ctx.IntSort()
		t.sortsInv[zs.String()] = s
		return zs, nil

	default:
		return Sort{}, fmt.Errorf("unsupported sort type: %T", s)
	}
}

// Translate converts an Ivy logic node to a Z3 expression.
func (t *Translator) Translate(n logic.Expr) (Expr, error) {
	switch node := n.(type) {
	case *logic.Variable:
		// Python: sksym = term.rep + ':' + term.sort.name
		// Variables use "name:sortName" as their Z3 const name.
		return t.translateVariable(node)

	case *logic.Symbol:
		return t.translateVarOrConst(node.Name, node.CSort)

	case *logic.Apply:
		if len(node.Terms) == 0 {
			// Nullary application: convert func directly
			return t.Translate(node.Func)
		}

		// Check if the function is a built-in operation (arithmetic, BV, etc.)
		if c, ok := node.Func.(*logic.Symbol); ok {
			result, handled, err := t.translateBuiltinOp(c.Name, node.Terms)
			if err != nil {
				return Expr{}, err
			}
			if handled {
				return result, nil
			}

			// Check for native interpretation via callback (handles polymorphic
			// symbols, range sort clamped arithmetic, nat interpretation, etc.)
			if t.NativeLookup != nil {
				isRelation := false
				if fs, ok := c.CSort.(*logic.FunctionSort); ok {
					if _, isBool := fs.Range().(*logic.BooleanSort); isBool {
						isRelation = true
					}
				}
				if nativeFn := t.NativeLookup(c.Name, c.CSort, isRelation); nativeFn != nil {
					args := make([]Expr, len(node.Terms))
					for i, term := range node.Terms {
						a, err := t.Translate(term)
						if err != nil {
							return Expr{}, err
						}
						args[i] = a
					}
					return nativeFn(args...), nil
				}
			}
		}

		fn, err := t.Translate(node.Func)
		if err != nil {
			return Expr{}, err
		}
		_ = fn
		// For function applications, we need to use the FuncDecl
		fd, err := t.getFuncDecl(node.Func)
		if err != nil {
			return Expr{}, err
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.Translate(term)
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return fd.Apply(args...), nil

	case *logic.Eq:
		t1, err := t.Translate(node.T1)
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Translate(node.T2)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Eq(t1, t2), nil

	case *logic.Not:
		b, err := t.Translate(node.Body)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Not(b), nil

	case *logic.And:
		if len(node.Terms) == 0 {
			return t.Ctx.BoolVal(true), nil
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.Translate(term)
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return t.Ctx.And(args...), nil

	case *logic.Or:
		if len(node.Terms) == 0 {
			return t.Ctx.BoolVal(false), nil
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.Translate(term)
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return t.Ctx.Or(args...), nil

	case *logic.Implies:
		t1, err := t.Translate(node.T1)
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Translate(node.T2)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Implies(t1, t2), nil

	case *logic.Iff:
		t1, err := t.Translate(node.T1)
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Translate(node.T2)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Iff(t1, t2), nil

	case *logic.Ite:
		c, err := t.Translate(node.Cond)
		if err != nil {
			return Expr{}, err
		}
		th, err := t.Translate(node.Then)
		if err != nil {
			return Expr{}, err
		}
		el, err := t.Translate(node.Else)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Ite(c, th, el), nil

	case *logic.ForAll:
		return t.translateQuantifier(true, node.Variables, node.Body)

	case *logic.Exists:
		return t.translateQuantifier(false, node.Variables, node.Body)

	default:
		return Expr{}, fmt.Errorf("unsupported node type for Z3 translation: %T", n)
	}
}

// translateVariable translates an Ivy variable to a Z3 const named "name:sortName".
// Corresponds to Python term_to_z3 variable case (ivy_solver.py:418-433).
func (t *Translator) translateVariable(v *logic.Variable) (Expr, error) {
	sort := v.VSort
	sortName := sort.Sexp()
	// Use the same sort name as Python: the sort's simple name
	if us, ok := sort.(*logic.UninterpretedSort); ok {
		sortName = us.Name
	} else if es, ok := sort.(*logic.EnumeratedSort); ok {
		sortName = es.Name
	} else if rs, ok := sort.(*logic.RangeSort); ok {
		sortName = rs.Name
	} else if _, ok := sort.(*logic.BooleanSort); ok {
		sortName = "Bool"
	}

	sksym := v.Name + ":" + sortName
	key := v.Name + ":" + sort.Sexp()
	if cached, ok := t.consts[key]; ok {
		return cached, nil
	}
	zs, err := t.TranslateSort(sort)
	if err != nil {
		return Expr{}, err
	}
	c := t.Ctx.Const(sksym, zs)
	t.consts[key] = c
	return c, nil
}

func (t *Translator) translateVarOrConst(name string, sort logic.Sort) (Expr, error) {
	z3name := t.z3Name(name, sort)
	if logic.FirstOrderSort(sort) {
		key := name + ":" + sort.Sexp()
		if cached, ok := t.consts[key]; ok {
			return cached, nil
		}
		zs, err := t.TranslateSort(sort)
		if err != nil {
			return Expr{}, err
		}
		c := t.Ctx.Const(z3name, zs)
		t.consts[key] = c
		return c, nil
	}

	// Function sort with single element (0-ary function) → treat as first-order
	if fs, ok := sort.(*logic.FunctionSort); ok && fs.Arity() == 0 {
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		key := name + ":" + fs.Range().Sexp()
		if cached, ok := t.consts[key]; ok {
			return cached, nil
		}
		c := t.Ctx.Const(z3name, zs)
		t.consts[key] = c
		return c, nil
	}

	// Higher-order: return a placeholder constant.
	// The actual FuncDecl is created in getFuncDecl.
	if fs, ok := sort.(*logic.FunctionSort); ok {
		fd, err := t.makeFuncDecl(name, fs)
		if err != nil {
			return Expr{}, err
		}
		_ = fd
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Const(z3name, zs), nil
	}

	return Expr{}, fmt.Errorf("cannot translate %s with sort %s to Z3", name, sort)
}

func (t *Translator) getFuncDecl(fn logic.Expr) (FuncDecl, error) {
	switch f := fn.(type) {
	case *logic.Symbol:
		fs, ok := f.CSort.(*logic.FunctionSort)
		if !ok {
			return FuncDecl{}, fmt.Errorf("expected FunctionSort for Apply func, got %T", f.CSort)
		}
		return t.makeFuncDecl(f.Name, fs)

	case *logic.Variable:
		fs, ok := f.VSort.(*logic.FunctionSort)
		if !ok {
			return FuncDecl{}, fmt.Errorf("expected FunctionSort for Apply func, got %T", f.VSort)
		}
		return t.makeFuncDecl(f.Name, fs)

	default:
		return FuncDecl{}, fmt.Errorf("unsupported Apply func type: %T", fn)
	}
}

func (t *Translator) makeFuncDecl(name string, fs *logic.FunctionSort) (FuncDecl, error) {
	key := name + ":" + fs.Sexp()
	if cached, ok := t.funcs[key]; ok {
		return cached, nil
	}

	z3name := t.z3Name(name, fs)

	domain := fs.Domain()
	zDomain := make([]Sort, len(domain))
	for i, d := range domain {
		zs, err := t.TranslateSort(d)
		if err != nil {
			return FuncDecl{}, err
		}
		zDomain[i] = zs
	}
	zRange, err := t.TranslateSort(fs.Range())
	if err != nil {
		return FuncDecl{}, err
	}

	fd := t.Ctx.Function(z3name, zDomain, zRange)
	t.funcs[key] = fd
	return fd, nil
}

func (t *Translator) translateQuantifier(isForall bool, variables []*logic.Variable, body logic.Expr) (Expr, error) {
	if len(variables) == 0 {
		return t.Translate(body)
	}

	// Create Z3 constants for the bound variables
	bound := make([]Expr, len(variables))
	for i, v := range variables {
		zs, err := t.TranslateSort(v.VSort)
		if err != nil {
			return Expr{}, err
		}
		// Use "name:sortName" to match Python's variable naming convention
		z3Var, err := t.translateVariable(v)
		if err != nil {
			return Expr{}, err
		}
		key := v.Name + ":" + v.VSort.Sexp()
		bound[i] = z3Var
		_ = zs
		// Temporarily override the const cache so the body uses these bound vars
		t.consts[key] = bound[i]
	}

	zBody, err := t.Translate(body)
	if err != nil {
		return Expr{}, err
	}

	// Validate that the body is a Bool expression before wrapping with
	// constraints or passing to Z3, which would otherwise panic with a
	// type error.
	if zBody.ExprSort().Kind() != SortBool {
		return Expr{}, fmt.Errorf("quantifier body must be Bool, got sort %s", zBody.ExprSort().String())
	}

	// Collect quantifier constraints (nat non-negativity, range sort bounds).
	// Corresponds to Python's quant_constraints + forall/exists wrapping.
	if t.QuantConstraints != nil {
		var allConstraints []Expr
		for i, v := range variables {
			cs := t.QuantConstraints(v, bound[i])
			allConstraints = append(allConstraints, cs...)
		}
		if len(allConstraints) > 0 {
			if isForall {
				// ForAll: body becomes Implies(And(constraints), body)
				zBody = t.Ctx.Implies(t.Ctx.And(allConstraints...), zBody)
			} else {
				// Exists: body becomes And(constraints..., body)
				zBody = t.Ctx.And(append(allConstraints, zBody)...)
			}
		}
	}

	if isForall {
		return t.Ctx.ForAll(bound, zBody), nil
	}
	return t.Ctx.Exists(bound, zBody), nil
}

// --- Convenience functions ---

// Implies checks if f1 implies f2 using Z3.
// Returns true if f1 => f2 is valid (i.e., f1 && !f2 is unsatisfiable).
func (t *Translator) Implies(f1, f2 logic.Expr) (bool, error) {
	zf1, err := t.Translate(f1)
	if err != nil {
		return false, err
	}
	zf2, err := t.Translate(f2)
	if err != nil {
		return false, err
	}

	s := t.Ctx.NewSolver()
	s.Assert(zf1)
	s.Assert(t.Ctx.Not(zf2))

	result := s.Check()
	switch result {
	case Unsat:
		return true, nil // f1 => f2 is valid
	case Sat:
		return false, nil // f1 does not imply f2
	default:
		return false, fmt.Errorf("z3 returned unknown")
	}
}

// translateBuiltinOp handles built-in operations that map to Z3 primitives
// rather than uninterpreted functions. Returns (result, handled, error).
// If handled is false, the caller should fall back to getFuncDecl.
//
// Corresponds to Python ivy_solver.py functions_dict and relations_dict.
func (t *Translator) translateBuiltinOp(name string, terms []logic.Expr) (Expr, bool, error) {
	// Translate arguments
	translateArgs := func() ([]Expr, error) {
		args := make([]Expr, len(terms))
		for i, term := range terms {
			a, err := t.Translate(term)
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		return args, nil
	}

	switch name {
	// --- Arithmetic ---
	case "+":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Add(args[0], args[1]), true, nil
		}
	case "-":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Sub(args[0], args[1]), true, nil
		}
		// Unary minus: 0 - x
		if len(args) == 1 {
			zero := t.Ctx.IntVal(0)
			return t.Ctx.Sub(zero, args[0]), true, nil
		}
	case "*":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Mul(args[0], args[1]), true, nil
		}
	case "/", "div":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Div(args[0], args[1]), true, nil
		}

	// --- Comparisons ---
	case "<":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Lt(args[0], args[1]), true, nil
		}
	case "<=":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Le(args[0], args[1]), true, nil
		}
	case ">":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Gt(args[0], args[1]), true, nil
		}
	case ">=":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Ge(args[0], args[1]), true, nil
		}

	// --- Bit-vector operations ---
	case "bvand":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvAnd(args[0], args[1]), true, nil
		}
	case "bvor":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvOr(args[0], args[1]), true, nil
		}
	case "bvxor":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvXor(args[0], args[1]), true, nil
		}
	case "bvnot":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 1 {
			return t.Ctx.BvNot(args[0]), true, nil
		}
	case "bvadd":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvAdd(args[0], args[1]), true, nil
		}
	case "bvsub":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvSub(args[0], args[1]), true, nil
		}
	case "bvmul":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvMul(args[0], args[1]), true, nil
		}
	case "bvudiv":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvUdiv(args[0], args[1]), true, nil
		}
	case "bvshl":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvShl(args[0], args[1]), true, nil
		}
	case "bvlshr":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvLshr(args[0], args[1]), true, nil
		}
	case "bvashr":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.BvAshr(args[0], args[1]), true, nil
		}
	case "concat":
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 2 {
			return t.Ctx.Concat(args[0], args[1]), true, nil
		}
	}

	// Check for bfe[lo:hi] pattern (bit-field extract)
	if len(name) > 4 && name[:3] == "bfe" && name[3] == '[' {
		args, err := translateArgs()
		if err != nil {
			return Expr{}, true, err
		}
		if len(args) == 1 {
			lo, hi, ok := parseBfeParams(name)
			if ok {
				return t.Ctx.Extract(hi, lo, args[0]), true, nil
			}
		}
	}

	return Expr{}, false, nil
}

// parseBfeParams parses "bfe[lo:hi]" and returns (lo, hi, ok).
func parseBfeParams(name string) (int, int, bool) {
	// Expected format: bfe[lo:hi] or bfe[lo,hi]
	inner := name[4:] // skip "bfe["
	if len(inner) < 2 || inner[len(inner)-1] != ']' {
		return 0, 0, false
	}
	inner = inner[:len(inner)-1] // strip "]"
	sep := -1
	for i, c := range inner {
		if c == ':' || c == ',' {
			sep = i
			break
		}
	}
	if sep < 0 {
		return 0, 0, false
	}
	var lo, hi int
	_, err1 := fmt.Sscanf(inner[:sep], "%d", &lo)
	_, err2 := fmt.Sscanf(inner[sep+1:], "%d", &hi)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

// IsSat checks if the formula is satisfiable.
func (t *Translator) IsSat(f logic.Expr) (CheckResult, error) {
	zf, err := t.Translate(f)
	if err != nil {
		return Unknown, err
	}
	s := t.Ctx.NewSolver()
	s.Assert(zf)
	return s.Check(), nil
}

// ParseArraySortName parses an array sort name like "arr[dom][rng]" into
// the domain and range sort names.
// Corresponds to Python ivy_solver.py:115 check for arr[ prefix and
// parse_array_theory call.
func ParseArraySortName(name string) (dom, rng string, ok bool) {
	if !strings.HasPrefix(name, "arr[") {
		return "", "", false
	}
	// Format: arr[dom][rng]
	rest := name[4:] // skip "arr["
	// Find matching ']' for domain
	depth := 1
	i := 0
	for i < len(rest) && depth > 0 {
		if rest[i] == '[' {
			depth++
		} else if rest[i] == ']' {
			depth--
		}
		if depth > 0 {
			i++
		}
	}
	if depth != 0 || i >= len(rest) {
		return "", "", false
	}
	dom = rest[:i]
	rest = rest[i+1:] // skip ']'
	// Now expect [rng]
	if len(rest) < 3 || rest[0] != '[' || rest[len(rest)-1] != ']' {
		return "", "", false
	}
	rng = rest[1 : len(rest)-1]
	return dom, rng, true
}
