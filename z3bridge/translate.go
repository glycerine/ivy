package z3bridge

import (
	"fmt"

	"github.com/glycerine/goivy/logic"
)

// NativeLookupFunc is a callback that looks up native Z3 interpretations
// for Ivy symbols. If the symbol has a native interpretation, it returns
// a function that maps Z3 arguments to a Z3 result. Returns nil if the
// symbol has no native interpretation.
type NativeLookupFunc func(name string, sort logic.Sort, isRelation bool) func(args ...Expr) Expr

// Translator converts Ivy logic nodes to Z3 expressions.
type Translator struct {
	Ctx          *Context
	sorts        map[string]Sort      // cache: Ivy sort name -> Z3 sort
	consts       map[string]Expr      // cache: "name:sort" -> Z3 const
	funcs        map[string]FuncDecl  // cache: "name:sort" -> Z3 func decl
	NativeLookup NativeLookupFunc     // optional: native interpretation callback
}

// NewTranslator creates a translator with a fresh Z3 context.
func NewTranslator() *Translator {
	return &Translator{
		Ctx:    NewContext(),
		sorts:  make(map[string]Sort),
		consts: make(map[string]Expr),
		funcs:  make(map[string]FuncDecl),
	}
}

// TranslateSort converts an Ivy sort to a Z3 sort.
func (t *Translator) TranslateSort(s logic.Sort) (Sort, error) {
	switch st := s.(type) {
	case *logic.BooleanSort:
		return t.Ctx.BoolSort(), nil

	case *logic.UninterpretedSort:
		if cached, ok := t.sorts[st.Name]; ok {
			return cached, nil
		}
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts[st.Name] = zs
		return zs, nil

	case *logic.TopSort:
		// TopSort is treated as uninterpreted in Z3
		key := "TopSort:" + st.Name
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts[key] = zs
		return zs, nil

	case *logic.FunctionSort:
		return Sort{}, fmt.Errorf("FunctionSorts are not directly converted to Z3 sorts")

	case *logic.EnumeratedSort:
		// For now, map to uninterpreted sort
		if cached, ok := t.sorts["enum:"+st.Name]; ok {
			return cached, nil
		}
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts["enum:"+st.Name] = zs
		return zs, nil

	case *logic.RangeSort:
		// Range sorts map to integers
		return t.Ctx.IntSort(), nil

	default:
		return Sort{}, fmt.Errorf("unsupported sort type: %T", s)
	}
}

// Translate converts an Ivy logic node to a Z3 expression.
func (t *Translator) Translate(n logic.Node) (Expr, error) {
	switch node := n.(type) {
	case *logic.Var:
		return t.translateVarOrConst(node.Name, node.VSort)

	case *logic.Const:
		return t.translateVarOrConst(node.Name, node.CSort)

	case *logic.Apply:
		if len(node.Terms) == 0 {
			// Nullary application: convert func directly
			return t.Translate(node.Func)
		}

		// Check if the function is a built-in operation (arithmetic, BV, etc.)
		if c, ok := node.Func.(*logic.Const); ok {
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

func (t *Translator) translateVarOrConst(name string, sort logic.Sort) (Expr, error) {
	if logic.FirstOrderSort(sort) {
		key := name + ":" + sort.String()
		if cached, ok := t.consts[key]; ok {
			return cached, nil
		}
		zs, err := t.TranslateSort(sort)
		if err != nil {
			return Expr{}, err
		}
		c := t.Ctx.Const(key, zs)
		t.consts[key] = c
		return c, nil
	}

	// Function sort with single element (0-ary function) → treat as first-order
	if fs, ok := sort.(*logic.FunctionSort); ok && fs.Arity() == 0 {
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		key := name + ":" + fs.Range().String()
		if cached, ok := t.consts[key]; ok {
			return cached, nil
		}
		c := t.Ctx.Const(key, zs)
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
		// Return the FuncDecl applied to no args as a "reference"
		// Z3 doesn't have first-class functions, but we track FuncDecls separately
		_ = fd
		// Return a const as placeholder (won't be used in Apply path)
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Const(name+":"+sort.String(), zs), nil
	}

	return Expr{}, fmt.Errorf("cannot translate %s with sort %s to Z3", name, sort)
}

func (t *Translator) getFuncDecl(fn logic.Node) (FuncDecl, error) {
	switch f := fn.(type) {
	case *logic.Const:
		fs, ok := f.CSort.(*logic.FunctionSort)
		if !ok {
			return FuncDecl{}, fmt.Errorf("expected FunctionSort for Apply func, got %T", f.CSort)
		}
		return t.makeFuncDecl(f.Name, fs)

	case *logic.Var:
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
	key := name + ":" + fs.String()
	if cached, ok := t.funcs[key]; ok {
		return cached, nil
	}

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

	fd := t.Ctx.Function(name, zDomain, zRange)
	t.funcs[key] = fd
	return fd, nil
}

func (t *Translator) translateQuantifier(isForall bool, variables []*logic.Var, body logic.Node) (Expr, error) {
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
		// Use a unique name for the bound variable
		key := v.Name + ":" + v.VSort.String()
		bound[i] = t.Ctx.Const(key, zs)
		// Temporarily override the const cache so the body uses these bound vars
		t.consts[key] = bound[i]
	}

	zBody, err := t.Translate(body)
	if err != nil {
		return Expr{}, err
	}

	if isForall {
		return t.Ctx.ForAll(bound, zBody), nil
	}
	return t.Ctx.Exists(bound, zBody), nil
}

// --- Convenience functions ---

// Implies checks if f1 implies f2 using Z3.
// Returns true if f1 => f2 is valid (i.e., f1 && !f2 is unsatisfiable).
func (t *Translator) Implies(f1, f2 logic.Node) (bool, error) {
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
func (t *Translator) translateBuiltinOp(name string, terms []logic.Node) (Expr, bool, error) {
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
func (t *Translator) IsSat(f logic.Node) (CheckResult, error) {
	zf, err := t.Translate(f)
	if err != nil {
		return Unknown, err
	}
	s := t.Ctx.NewSolver()
	s.Assert(zf)
	return s.Check(), nil
}
