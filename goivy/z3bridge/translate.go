package z3bridge

import (
	"fmt"
	"sort"
	"strings"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
)

/*
// LookupNativeFunc is a callback matching Python lookup_native(thing, table, kind)
// (ivy_solver.py:311). Returns any: z3bridge.Sort for sort lookups,
// func(args ...Expr) Expr for function/relation lookups, or nil.
type LookupNativeFunc func(name string, sort lg.Sort, kind string) any

// SolverNameFunc maps an Ivy symbol to its Z3 name. Returns "" if the
// symbol should be handled natively (not declared as an uninterpreted function).
type SolverNameFunc func(name string, sort lg.Sort) string

// EqFuncFn is a callback for custom equality (e.g., MyEq with True/False optimization).
// Corresponds to Python's my_eq (ivy_solver.py:88-95).
type EqFuncFn func(x, y Expr) Expr

// EnumEqFuncFn is a callback for custom enumerated sort equality (binary encoding).
// Returns non-nil Expr to override default equality; nil to use default.
// Corresponds to Python's encode_equality dispatch in atom_to_z3 (ivy_solver.py:484).
type EnumEqFuncFn func(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (*Expr, error)

// NumeralFuncFn is a callback for custom numeral handling (range sort clamping).
// Returns non-nil Expr to override default; nil to use default.
// Corresponds to Python's numeral_to_z3 (ivy_solver.py:388-404).
type NumeralFuncFn func(name string, sort lg.Sort) (*Expr, error)
*/

// Translator converts Ivy logic nodes to Z3 expressions.
type Translator struct {
	s *Solver

	// cache holds the shared Z3 translation state (sorts, predicates,
	// constants, functions, sortsInv) plus the Z3 context. Translators
	// constructed via NewTranslatorWithCache share the same cache —
	// matching Python's module-level z3_sorts/z3_predicates/etc. that
	// all z3.Solver() instances within a Module context share.
	// Translators constructed via NewTranslator get their own fresh cache
	// (legacy/test behavior).
	cache *Z3SessionCache

	// Ctx is an alias for cache.Ctx, kept as a field for backward
	// compatibility with all existing t.Ctx / tr.Ctx call sites.
	Ctx             *Z3Context
	TranslateMerkle iu.MerkleState // rolling Merkle hash for Formula_to_z3_int() input conformance
	translateDepth  int            // nesting depth; only hash at top level (depth 0)

	// these were pre-merge of solver/ and z3bridge hacks to avoid circular imports
	/*	LookupNative     LookupNativeFunc                          // optional: Python lookup_native(thing, table, kind) callback
		SolverName       SolverNameFunc                            // optional: maps symbol to Z3 name (for polymorphic disambiguation)
		QuantConstraints QuantConstraintsFn                        // optional: generates sort constraints for quantifier-bound variables
		EqFunc           EqFuncFn                                  // optional: custom equality (MyEq True/False optimization)
		EnumEqFunc       EnumEqFuncFn                              // optional: custom enumerated equality (binary encoding)
		NumeralFunc      NumeralFuncFn                             // optional: custom numeral handling (range clamping)
	*/

}

func (t *Translator) LookupNative(name string, sort lg.Sort, kind string) any {
	sym := lg.NewConst(name, sort)
	var table func(string) any
	switch kind {
	case "sort":
		table = t.s.Sorts
	case "relation":
		table = t.s.Relations
	case "function":
		table = t.s.Functions
	}
	result := t.s.LookupNative(sym, table, kind)
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

// SolverName maps an Ivy symbol to its Z3 name. Returns "" if the
// symbol should be handled natively (not declared as an uninterpreted function).
func (t *Translator) SolverName(name string, sort lg.Sort) string {
	sym := lg.NewConst(name, sort)
	return t.s.SolverName(sym)
}

// quantConstraints generates Z3 constraints for quantifier-bound variables
// based on their sort (nat non-negativity, range sort bounds).
// Matches Python quant_constraints (ivy_solver.py:557-568).
func (t *Translator) quantConstraints(vars []*lg.Variable, z3Vars []Expr) []Expr {
	xtracer.Trace("ivy_solver.py:545 quant_constraints() ENTER nvars=%d", len(vars))
	if t.s == nil || t.s.sig == nil {
		return nil
	}
	var cnstrs []Expr
	for i, v := range vars {
		sortName := il.SortName(v.VSort)
		itp, ok := t.s.sig.Interp[sortName]
		if !ok {
			continue
		}
		switch itpVal := itp.(type) {
		case string:
			if itpVal == "nat" {
				cnstrs = append(cnstrs, t.Ctx.Le(t.Ctx.IntVal(0), z3Vars[i]))
			}
		case *lg.RangeSort:
			if t.s.HandleRangeSorts {
				lb, ub, err := t.s.RangeSortBoundsToZ3(itpVal)
				if err == nil {
					cnstrs = append(cnstrs, t.Ctx.Le(lb, z3Vars[i]))
					cnstrs = append(cnstrs, t.Ctx.Le(z3Vars[i], ub))
				}
			}
		}
	}
	return cnstrs
}

// forall wraps a Z3 body in ForAll with quant constraints (nat/range bounds).
// Matches Python's forall (ivy_solver.py:572-577).
func (t *Translator) forall(vars []*lg.Variable, z3Vars []Expr, z3Body Expr) Expr {
	xtracer.Trace("ivy_solver.py:560 forall() ENTER nvars=%d", len(vars))
	cnstrs := t.quantConstraints(vars, z3Vars)
	if len(cnstrs) > 0 {
		z3Body = t.Ctx.Implies(t.Ctx.And(cnstrs...), z3Body)
	}
	return t.Ctx.ForAll(z3Vars, z3Body)
}

// exists wraps a Z3 body in Exists with quant constraints (nat/range bounds).
// Matches Python's exists (ivy_solver.py:579-584).
func (t *Translator) exists(vars []*lg.Variable, z3Vars []Expr, z3Body Expr) Expr {
	xtracer.Trace("ivy_solver.py:567 exists() ENTER nvars=%d", len(vars))
	cnstrs := t.quantConstraints(vars, z3Vars)
	if len(cnstrs) > 0 {
		z3Body = t.Ctx.And(append(cnstrs, z3Body)...)
	}
	return t.Ctx.Exists(z3Vars, z3Body)
}

// Eq is for custom equality (e.g., MyEq with True/False optimization).
// Corresponds to Python's my_eq (ivy_solver.py:88-95).
func (t *Translator) Eq(x, y Expr) Expr {
	return MyEq(t.Ctx, x, y)
}

// EnumEq is for custom enumerated sort equality (binary encoding).
// Returns non-nil Expr to override default equality; nil to use default.
// Corresponds to Python's encode_equality dispatch in atom_to_z3 (ivy_solver.py:484).
func (t *Translator) EnumEq(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (*Expr, error) {
	if !t.s.opts.UseZ3Enums {
		result, err := t.s.EncodeEqualityZ3(t1, t2, sort)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, nil // use default Z3 enum equality
}

// Numeral is for custom numeral handling (range sort clamping).
// Returns non-nil Expr to override default; nil to use default.
// Corresponds to Python's numeral_to_z3 (ivy_solver.py:388-404).
func (t *Translator) Numeral(name string, sort lg.Sort) (*Expr, error) {
	num := lg.NewConst(name, sort)
	result, err := t.s.NumeralToZ3(num)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// NewTranslator creates a Translator that owns its own private cache
// (and therefore its own Z3 context). Used by tests, ad-hoc Solver
// instances without a module, and the legacy NewSolver(nil, ...) path.
func (s *Solver) NewTranslator() *Translator {
	return s.NewTranslatorWithCache(NewZ3SessionCache())
}

// NewTranslatorWithCache creates a Translator that shares the given
// Z3SessionCache (and the cache's Z3 context). Multiple Translators built
// from the same cache will see each other's cached sorts/predicates/
// constants/functions — this is the Go equivalent of Python's
// "z3.Solver() instances within a Module context share z3_sorts".
func (s *Solver) NewTranslatorWithCache(cache *Z3SessionCache) *Translator {
	return &Translator{
		s:     s,
		cache: cache,
		Ctx:   cache.Ctx,
	}
}

func (t *Translator) Close() error {
	// Do not close t.Ctx here: the Z3Context lives on the cache and may
	// be shared with other Translators (and indeed with other Solvers
	// belonging to the same module session). Let GC dispose of the
	// Z3Context when the cache itself becomes unreachable.
	return nil
}

// Clear resets all Z3 caches to initial state.
// Corresponds to Python ivy_solver.clear() (line 228).
func (t *Translator) Clear() {
	t.cache.Clear()
}

// SortFromZ3 looks up the original Ivy sort for a Z3 sort using the reverse map.
// Corresponds to Python's sort_from_z3() (ivy_solver.py:905).
func (t *Translator) SortFromZ3(z3sort Sort) (lg.Sort, bool) {
	ivySort, ok := t.cache.sortsInv[z3sort.GetId()]
	return ivySort, ok
}

// z3Name returns the Z3 name for a symbol. If SolverName is set, uses it;
// otherwise returns the plain name.
func (t *Translator) z3Name(name string, sort lg.Sort) string {
	return t.SolverName(name, sort)
	//return name
}

// xtracer / dump helper for viewing the session-cache sorts in
// deterministic (sorted) order.
func (t *Translator) dumpSortsCanon() (r string) {
	var slc []string
	for _, srt := range t.cache.sorts {
		slc = append(slc, srt.String())
	}
	sort.Strings(slc)
	return "[" + strings.Join(slc, ", ") + "]"
}

// TranslateSort converts an Ivy sort to a Z3 sort.
// Emits type-specific traces matching Python's per-sort-type functions:
// uninterpretedsort() (line 258), enumeratedsort() (line 276), etc.
//
// What is the equivalent of TranslateSort in python?
// It is the polymorphic .to_z3() method, monkey-patched
// onto each sort class at ivy_solver.py:286-295:
//
// ivy_logic.UninterpretedSort.to_z3 = uninterpretedsort    # line 286
// ivy_logic.FunctionSort.to_z3 = functionsort              # line 293
// ivy_logic.EnumeratedSort.to_z3 = enumeratedsort          # line 294
// ivy_logic.RangeSort.to_z3 = lambda self: z3.IntSort()    # line 290
// ivy_logic.BooleanSort.to_z3 = lambda self: z3.BoolSort() # line 289
//
// Our TranslateSort is a structural equivalent. It
// dispatches on sort type via a switch, doing the same thing
// as Python's polymorphic dispatch. This also avoids
// a circular import issue that would need two more interfaces
// to circumvent.
func (t *Translator) TranslateSort(s lg.Sort) (Sort, error) {
	switch st := s.(type) {
	case *lg.BooleanSort:
		// Python z3.BoolSort() has no custom trace
		return t.Ctx.BoolSort(), nil

	case *lg.UninterpretedSort:
		// Python: uninterpretedsort(us) at ivy_solver.py:257
		if xtracer.Enabled {
			xtracer.Trace("ivy_solver.py:263 uninterpretedsort() ENTER name=%s", st.Name)
			xtracer.Trace("ivy_solver.py:263 uninterpretedsort top HASH canon= sorts=%v", t.dumpSortsCanon())
		}
		key := lg.NodeKey(st.Name) // Python: z3_sorts[us.rep] where rep = name
		if cached, ok := t.cache.sorts[key]; ok {
			xtracer.Trace("ivy_solver.py:266 uninterpretedsort() EXIT 1: cache hit")
			return cached, nil
		}
		// Python: s = lookup_native(us, sorts, "sort")

		if result := t.LookupNative(st.Name, s, "sort"); result != nil {
			if zs, ok := result.(Sort); ok {
				xtracer.Trace("ivy_solver.py:273 uninterpretedsort() not-None from lookup_native")
				t.cache.sorts[key] = zs
				t.cache.sortsInv[zs.GetId()] = s
				return zs, nil
			}
		}

		xtracer.Trace("ivy_solver.py:270 uninterpretedsort() None from lookup_native")
		// Python: if s == None: s = z3.DeclareSort(us.rep)
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.cache.sorts[key] = zs
		// Python: z3_sorts_inv[get_id(s)] = us
		t.cache.sortsInv[zs.GetId()] = s
		return zs, nil

	case *lg.TopSort:
		// Python has no TopSort.to_z3 — calling it would raise AttributeError.
		// Panic here to match Python's behavior and expose bugs.
		panic(fmt.Sprintf("TranslateSort: TopSort %q has no to_z3() equivalent in Python", st.Name))

	case *lg.FunctionSort:
		// FunctionSort.to_z3 in Python is functionsort(), which returns a
		// list of Z3 sorts — not a single Z3 Sort. Callers that need this
		// should call (*Translator).functionSort directly. Reaching this
		// branch means a caller wrongly handed a FunctionSort to a code
		// path that expects a single sort; surface that as an error.
		return Sort{}, fmt.Errorf("FunctionSorts are not directly converted to Z3 sorts")

	case *lg.EnumeratedSort:
		// Python: enumeratedsort(es) at ivy_solver.py:280
		xtracer.Trace("ivy_solver.py:281 enumeratedsort() ENTER name=%s", st.Name)
		key := lg.NodeKey(st.Name) // Python: z3_sorts[es.rep] where rep = name
		if cached, ok := t.cache.sorts[key]; ok {
			xtracer.Trace("ivy_solver.py:284 enumeratedsort() EXIT 1: cache hit.")
			return cached, nil
		}
		// Use native Z3 EnumSort, matching Python's z3.EnumSort(name, extension).
		zs, constExprs := t.Ctx.EnumSort(st.Name, st.Extension)
		t.cache.sorts[key] = zs
		// Python enumeratedsort() does NOT store in z3_sorts_inv.
		// The reverse map entry is created by uninterpretedsort (the normal
		// entry point for interpreted sorts).
		// Register the constructor constants so they can be looked up by name.
		for i, name := range st.Extension {
			constKey := lg.NodeKey(name + ":" + string(s.Sexp()))
			t.cache.consts[constKey] = constExprs[i]
		}
		xtracer.Trace("ivy_solver.py:291 enumeratedsort() EXIT 2: cache miss.")
		return zs, nil

	case *lg.RangeSort:
		// Python: lambda self: z3.IntSort() — no trace, no caching, no sortsInv
		return t.Ctx.IntSort(), nil

	default:
		return Sort{}, fmt.Errorf("unsupported sort type: %T", s)
	}
}

// Translate converts an Ivy logic node to a Z3 expression.
// Called recursively for subexpressions; translateDepth tracks nesting.
// TranslateNoHash translates without emitting the top-level HASH trace.
// Matches Python's formula_to_z3_int/formula_to_z3_closed which do not
// emit HASH — only formula_to_z3 does.
func (t *Translator) TranslateNoHash(n lg.Expr) (Expr, error) {
	t.translateDepth++
	defer func() { t.translateDepth-- }()
	return t.Formula_to_z3_int(n, "term_to_z3_closed")
}

// Translate converts an Ivy logic node to a Z3 expression.
// Corresponds to Python formula_to_z3_int (ivy_solver.py:637).
// Dispatches Boolean-sorted Apply nodes to atomToZ3 (Python atom_to_z3).
// This is a thin wrapper around Translator.Formula_to_z3_int(),
// which does more diagnostic logging (of the caller).
func (t *Translator) Translate(n lg.Expr) (Expr, error) {
	return t.Formula_to_z3_int(n, "")
}

// Formula_to_z3_int converts an Ivy logic node to a Z3 expression.
// Corresponds to Python formula_to_z3_int (ivy_solver.py:637).
// Dispatches Boolean-sorted Apply nodes to atomToZ3 (Python atom_to_z3).
func (t *Translator) Formula_to_z3_int(n lg.Expr, caller string) (Expr, error) {
	//xtracer.Trace("ivy_solver.py:651 formula_to_z3_int() ENTER type=%v\ncaller=%v", iu.ShortTypeName(n), caller)
	if xtracer.Enabled && t.translateDepth == 0 {
		canon := iu.Canonical(n.Sexp())
		leaf, root := t.TranslateMerkle.AddLeaf(canon)
		_, _ = leaf, root
		//xtracer.Trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s", leaf, root, canon)
		//xtracer.Trace("ivy_solver.py:718 formula_to_z3 HASH leaf=%s root=%s canon=%s", leaf, root, canon) // 2 of 2
	}
	t.translateDepth++
	defer func() { t.translateDepth-- }()

	// Python line 645: if ivy_logic.is_atom(fmla): return atom_to_z3(fmla)
	// is_atom: (Apply or Symbol) with Boolean sort, or Eq.
	if app, ok := n.(*lg.Apply); ok && len(app.Terms) > 0 {
		if _, isBool := n.NodeSort().(*lg.BooleanSort); isBool {
			return t.atomToZ3(app)
		}
	}
	//if app, ok := n.(*lg.Apply); ok {
	//	if c, ok2 := app.Func.(*lg.Const); ok2 && isPolymac(c.Name) {
	//		vv("DEBUG F2Z3: polymac op=%s nTerms=%d sort=%T isApp=%v\n", c.Name, len(app.Terms), n.NodeSort(), true)
	//	}
	//}

	// Python: isinstance(term, lg.Eq) in is_atom → atom_to_z3(fmla)
	// Eq has duck-typed .rep/.args/.relname in Python; we construct a
	// pseudo-Apply to pass through the same atomToZ3 code path.
	if eq, ok := n.(*lg.Eq); ok {
		return t.eqToAtomZ3(eq)
	}

	return t.translateCore(n, caller)
}

// translateCore is the inner dispatch for Translate. It handles all node
// types except Boolean-sorted Apply (which is routed to atomToZ3 by
// Translate). No XTRACE, no Merkle hash, no depth tracking — those are
// done by the caller (Translate or TermToZ3).
func (t *Translator) translateCore(n lg.Expr, caller string) (Expr, error) {
	switch node := n.(type) {
	case *lg.Variable:
		// Python: sksym = term.rep + ':' + term.sort.name
		// Variables use "name:sortName" as their Z3 const name.
		return t.translateVariable(node)

	case *lg.Const:
		return t.translateVarOrConst(node.Name, node.CSort)

	case *lg.Apply:
		if len(node.Terms) == 0 {
			// Nullary application: convert func directly
			return t.Formula_to_z3_int(node.Func, caller)
		}

		// Non-Boolean Apply: corresponds to Python term_to_z3 "else"
		// branch (line 483–497) for function applications.
		if c, ok := node.Func.(*lg.Const); ok {
			// Check if the function is a built-in operation (arithmetic, BV, etc.)
			result, handled, err := t.translateBuiltinOp(c.Name, node.Terms)
			if err != nil {
				return Expr{}, err
			}
			if handled {
				return result, nil
			}

			// Check for native interpretation via callback (handles polymorphic
			// symbols, range sort clamped arithmetic, nat interpretation, etc.)
			// Python: fun = lookup_native(term.rep, functions, "function")

			if result := t.LookupNative(c.Name, c.CSort, "function"); result != nil {
				if nativeFn, ok := result.(func(args ...Expr) Expr); ok {
					args := make([]Expr, len(node.Terms))
					for i, term := range node.Terms {
						a, err := t.TermToZ3(term)
						if err != nil {
							return Expr{}, err
						}
						args[i] = a
					}
					return nativeFn(args...), nil
				}
			}
		}

		// Python: fun = z3.Function(sn, *sig); args = [term_to_z3(arg) for arg in term.args]
		fd, err := t.getFuncDecl(node.Func)
		if err != nil {
			return Expr{}, err
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.TermToZ3(term)
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return fd.Apply(args...), nil

	case *lg.Eq:
		// Eq is now routed through atomToZ3 by Formula_to_z3_int() (matching Python's
		// is_atom dispatch). This case should be unreachable.
		return Expr{}, fmt.Errorf("translateCore: unexpected Eq (should be routed through atomToZ3 by Translate)")

	case *lg.Not:
		b, err := t.Formula_to_z3_int(node.Body, "lg.Not Body/args[0]")
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Not(b), nil

	case *lg.And:
		if len(node.Terms) == 0 {
			return t.Ctx.BoolVal(true), nil
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.Formula_to_z3_int(term, fmt.Sprintf("lg.And args[i=%v]", i))
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return t.Ctx.And(args...), nil

	case *lg.Or:
		if len(node.Terms) == 0 {
			return t.Ctx.BoolVal(false), nil
		}
		args := make([]Expr, len(node.Terms))
		for i, term := range node.Terms {
			a, err := t.Formula_to_z3_int(term, fmt.Sprintf("lg.Or args[i=%v]", i))
			if err != nil {
				return Expr{}, err
			}
			args[i] = a
		}
		return t.Ctx.Or(args...), nil

	case *lg.Implies:
		t1, err := t.Formula_to_z3_int(node.T1, "lg.Implies args[0]")
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Formula_to_z3_int(node.T2, "lg.Implies args[1]")
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Implies(t1, t2), nil

	case *lg.Iff:
		t1, err := t.Formula_to_z3_int(node.T1, "lg.Iff args[0]")
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Formula_to_z3_int(node.T2, "lg.Iff args[1]")
		if err != nil {
			return Expr{}, err
		}
		// Python uses my_eq for Iff (formula_to_z3_int line 620)
		//if t.EqFunc != nil {
		//  return t.EqFunc(t1, t2), nil
		return t.Eq(t1, t2), nil
		//}
		//return t.Ctx.Iff(t1, t2), nil

	case *lg.Ite:
		c, err := t.Formula_to_z3_int(node.Cond, "lg.Ite node.Cond")
		if err != nil {
			return Expr{}, err
		}
		th, err := t.Formula_to_z3_int(node.Then, "lg.Ite node.Then")
		if err != nil {
			return Expr{}, err
		}
		el, err := t.Formula_to_z3_int(node.Else, "lg.Ite node.Else")
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Ite(c, th, el), nil

	case *lg.Definition:
		// Python formula_to_z3_int lines 589-615, 596-597, 609-614
		// Check for enumerated encoding (when !UseZ3Enums)
		if es, ok := node.Lhs.NodeSort().(*lg.EnumeratedSort); ok {
			if result, err := t.EnumEq(node.Lhs, node.Rhs, es); result != nil {
				return *result, err
			}
		}
		t1, err := t.Formula_to_z3_int(node.Lhs, "lg.Definition Lhs")
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Formula_to_z3_int(node.Rhs, "lg.Definition Rhs")
		if err != nil {
			return Expr{}, err
		}
		return t.Eq(t1, t2), nil
		//return t.Ctx.Eq(t1, t2), nil

	case *lg.ForAll:
		return t.translateQuantifier(true, node.Variables, node.Body)

	case *lg.Exists:
		return t.translateQuantifier(false, node.Variables, node.Body)

	default:
		return Expr{}, fmt.Errorf("unsupported node type for Z3 translation: %T", n)
	}
}

const eqCanonPredKey = lg.NodeKey("=:eq")

// atomToZ3 translates a Boolean-sorted Apply (atom) to Z3.
// Corresponds to Python atom_to_z3 (ivy_solver.py:516).
func (t *Translator) atomToZ3(app *lg.Apply) (Expr, error) {
	c, ok := app.Func.(*lg.Const)
	if !ok {
		// Fallback for non-Const func (rare)
		return t.translateCore(app, "atom_to_z3")
	}

	//xtracer.Trace("ivy_solver.py:517 atom_to_z3() ENTER rep=%s nargs=%d", c.Name, len(app.Terms))

	// Python line 518: if ivy_logic.is_equals(atom.rep) and
	//   ivy_logic.is_enumerated(atom.args[0]) and not use_z3_enums
	//
	//
	if c.Name == "=" && len(app.Terms) == 2 {
		if es, ok2 := app.Terms[0].NodeSort().(*lg.EnumeratedSort); ok2 {
			if result, err := t.EnumEq(app.Terms[0], app.Terms[1], es); result != nil {
				return *result, err
			}
		}
	}

	// Check z3_predicates cache (Python z3_predicates)
	// Python: atom.relname for Eq is always equals = Symbol('=', RelationSort([TopSort(), TopSort()]))
	// regardless of concrete sorts. Use a canonical key to match Python's caching.
	// (all equalities need to share one cache entry, matching
	// Python's z3_predicates[atom.relname] behavior)
	var predKey lg.NodeKey
	if c.Name == "=" {
		predKey = eqCanonPredKey
	} else {
		predKey = lg.NodeKey(c.Name + ":" + string(c.CSort.Sexp()))
	}
	if cached, ok := t.cache.z3_predicates[predKey]; ok {
		return t.applyZ3Func(cached, app.Terms) // in atomToZ3() here.
	}

	// Check builtin ops (Go-specific: Python handles via polymacs inside lookup_native)
	result, handled, err := t.translateBuiltinOp(c.Name, app.Terms)
	if err != nil {
		return Expr{}, err
	}
	if handled {
		return result, nil
	}

	// Python line 521: rel = lookup_native(atom.relname, relations, "relation")
	if result := t.LookupNative(c.Name, c.CSort, "relation"); result != nil {
		if nativeFn, ok := result.(func(args ...Expr) Expr); ok {
			t.cache.z3_predicates[predKey] = nativeFn
			return t.applyZ3Func(nativeFn, app.Terms) // in atomToZ3 here.
		}
	}

	// Python lines 536-538: polymorphic macro check.
	// For <=, >, >= on uninterpreted sorts where lookup_native returned nil,
	// expand into combinations of < and = at the Z3 level.
	// Matches Python polymacs dict (ivy_solver.py:519-523).
	//if isPolymac(c.Name) {
	//vv("DEBUG polymacs: op=%s usePolymorphicMacros=%v s=%v sig=%v iuCfg=%v\n",
	//	c.Name, t.usePolymorphicMacros(), t.s != nil, t.s != nil && t.s.sig != nil,
	//	t.s != nil && t.s.sig != nil && t.s.sig.IuCfg != nil)
	//}
	if isPolymac(c.Name) && t.usePolymorphicMacros() {
		xtracer.Trace("ivy_solver.py:513 get_polymacs() ENTER op=%s", c.Name)
		predFn, err := t.polymacPred(c.Name, c.CSort)
		if err != nil {
			return Expr{}, err
		}
		t.cache.z3_predicates[predKey] = predFn
		return t.applyZ3Func(predFn, app.Terms)
	}

	// Python's z3.Function("=", ...) returns Z3's built-in equality.
	// Go's makeFuncDecl would create an uninterpreted function instead.
	// Use Ctx.Eq (or EqFunc) for "=" to match Python/Z3 semantics.
	if c.Name == "=" {
		eqFn := func(args ...Expr) Expr {
			return t.Eq(args[0], args[1])
			//return t.Ctx.Eq(args[0], args[1])
		}
		t.cache.z3_predicates[predKey] = eqFn
		return t.applyZ3Func(eqFn, app.Terms) // in atomToZ3 here.
	}

	// Python ivy_solver.py:540-541:
	//     sig = atom.rep.sort.to_z3()
	//     rel = z3.Function(solver_name(atom.rep), *sig) if isinstance(sig,list)
	//           else z3.Const(solver_name(atom.rep),sig)
	//
	// Python's atom_to_z3 populates the z3_predicates cache (mirrored
	// here by t.cache.z3_predicates); it NEVER touches z3_functions. We
	// must not go through makeFuncDecl (which caches in
	// t.cache.z3_functions) — sharing that cache causes a hit from a prior
	// term_to_z3 or symbol_to_z3 path to suppress the functionsort()
	// ENTER xtrace, diverging from Python. Same reasoning and
	// precedent as ltPred (see its comment referencing log.red step
	// 236451).
	fs, ok := c.CSort.(*lg.FunctionSort)
	if !ok {
		return Expr{}, fmt.Errorf("atomToZ3: expected FunctionSort for %s, got %T", c.Name, c.CSort)
	}
	// Python: sig = atom.rep.sort.to_z3() (ivy_solver.py:572). This dispatches
	// to functionsort via the monkey-patched FunctionSort.to_z3, so Python emits
	// the atom_to_z3_relation callsite trace before functionsort. We mirror that.
	xtracer.Trace("TranslateSort_call callsite=atom_to_z3_relation")
	sig, err := t.functionSort(fs) // emits "ivy_solver.py:279 functionsort() ENTER"
	if err != nil {
		return Expr{}, err
	}
	zDomain := sig[:len(sig)-1]
	zRange := sig[len(sig)-1]
	z3name := t.z3Name(c.Name, fs)
	fd := t.Ctx.Function(z3name, zDomain, zRange)
	predFn := fd.Apply
	t.cache.z3_predicates[predKey] = predFn
	return t.applyZ3Func(predFn, app.Terms) // end of atomToZ3 here.
}

// isPolymac returns true if name is a polymorphic macro operator.
// Matches Python polymacs dict keys (ivy_solver.py:519-523): <=, >, >=.
// Distinct from isPolymorphicOp which also includes +, -, *, /, <.
func isPolymac(name string) bool {
	switch name {
	case "<=", ">", ">=":
		return true
	}
	return false
}

// usePolymorphicMacros checks if polymorphic macros are enabled (version > 1.5).
// Matches Python iu.ivy_use_polymorphic_macros.
func (t *Translator) usePolymorphicMacros() bool {
	if t.s == nil || t.s.sig == nil || t.s.sig.IuCfg == nil {
		return false
	}
	return t.s.sig.IuCfg.UsePolymorphicMacros
}

// polymacPred returns a Z3-level predicate for a polymorphic macro operator.
// Matches Python polymacs dict (ivy_solver.py:519-523).
// Uses t.Ctx.Eq (plain Z3 equality, not MyEq) to match Python's x == y.
func (t *Translator) polymacPred(name string, sort lg.Sort) (func(args ...Expr) Expr, error) {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return nil, fmt.Errorf("polymacPred: expected FunctionSort for %s, got %T", name, sort)
	}
	switch name {
	case "<=":
		// Python: lambda s,x,y: z3.Or(x == y, lt_pred(s)(x,y))
		return func(args ...Expr) Expr {
			ltFd := t.ltPred(fs)
			return t.Ctx.Or(t.Ctx.Eq(args[0], args[1]), ltFd.Apply(args...))
		}, nil
	case ">":
		// Python: lambda s,x,y: lt_pred(s)(y,x)
		return func(args ...Expr) Expr {
			ltFd := t.ltPred(fs)
			return ltFd.Apply(args[1], args[0])
		}, nil
	case ">=":
		// Python: lambda s,x,y: z3.Or(x == y, lt_pred(s)(y,x))
		return func(args ...Expr) Expr {
			ltFd := t.ltPred(fs)
			return t.Ctx.Or(t.Ctx.Eq(args[0], args[1]), ltFd.Apply(args[1], args[0]))
		}, nil
	}
	return nil, fmt.Errorf("polymacPred: unknown operator %s", name)
}

// ltPred creates a Z3 function declaration for "<" on the given sort.
// Mirrors Python lt_pred (ivy_solver.py:513-517):
//
//	def lt_pred(sort):
//	    sym = ivy_logic.Symbol('<', sort)
//	    sig = sym.sort.to_z3()                  # functionsort() ENTER
//	    return z3.Function(solver_name(sym), *sig)
//
// Python does NOT cache the result. Each invocation re-emits the
// functionsort() trace and rebuilds the Z3 Function. Z3 dedups
// internally by (name, signature), so multiple calls return
// equivalent FuncDecls. We mirror that exactly here — going through
// makeFuncDecl's cache would suppress the functionsort() ENTER trace
// on hits and break xtrace alignment with Python (this was the cause
// of the log.red divergence at step 236451).
func (t *Translator) ltPred(fs *lg.FunctionSort) FuncDecl {
	xtracer.Trace("ivy_solver.py:501 lt_pred() ENTER sort=%s", fs)
	sig, err := t.functionSort(fs)
	if err != nil {
		panic(fmt.Sprintf("ltPred: functionSort failed: %v", err))
	}
	zDomain := sig[:len(sig)-1]
	zRange := sig[len(sig)-1]
	z3name := t.z3Name("<", fs)
	return t.Ctx.Function(z3name, zDomain, zRange)
}

// eqToAtomZ3 routes Eq through atomToZ3, matching Python where
// is_atom() returns True for Eq and atom_to_z3 accesses Eq via
// duck-typed properties:
//
//	atom.rep = Symbol('=', RelationSort([x.sort for x in self.args]))
//	atom.args = [t1, t2]
//	atom.relname = equals (ivy_logic.py:1145-1147)
func (t *Translator) eqToAtomZ3(eq *lg.Eq) (Expr, error) {
	s1, s2 := eq.T1.NodeSort(), eq.T2.NodeSort()
	repSort, err := lg.NewFunctionSort(s1, s2, lg.Boolean)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	eqConst := lg.NewConst("=", repSort)
	pseudoApp, err := lg.NewApply(eqConst, eq.T1, eq.T2)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	return t.atomToZ3(pseudoApp)
}

// applyZ3Func translates args via TermToZ3 and applies the predicate function.
// Corresponds to Python apply_z3_func (ivy_solver.py:403).
func (t *Translator) applyZ3Func(pred func(args ...Expr) Expr, terms []lg.Expr) (Expr, error) {
	//xtracer.Trace("ivy_solver.py:404 apply_z3_func() ENTER nargs=%d", len(args))

	// the original python is very polymorphic:
	// Here are the three call sites and what pred can be:
	//
	// Call site 1 — line 497 (term_to_z3): fun is always
	// a z3.FuncDeclRef (from z3.Function()).
	//
	// Call site 2 — line 534 (atom_to_z3):  (translate.go:463)
	// pred comes
	// from z3_predicates[atom.relname], which is
	// populated from three sources:
	// - lookup_native() — returns Python lambdas/callables
	//   (e.g., lambda x,y: z3.If(x < y, ...),
	// lambda x: z3.K(...)) or a z3.FuncDeclRef
	// - get_polymacs() — returns a functools.partial (a callable)
	// - z3.Function() → z3.FuncDeclRef
	// - z3.Const() → z3.BoolRef (or other ExprRef) when the sort
	// is non-functional (a 0-ary predicate/constant)
	//
	// Call site 3 — line 1767 (encode_term): z3_function()
	// returns either z3.Function() →
	// z3.FuncDeclRef, or z3.Const() → z3.ExprRef/z3.BoolRef.
	//
	// So pred has three possible types:
	//
	// ┌────────────────────────┬─────────────────────┬─────────────────────┐
	// │            Type        │            When     │        Branch taken │
	// ├────────────────────────┼─────────────────────┼─────────────────────┤
	// │ z3.BoolRef (or ExprRef)│ 0-ary constant from │ line 405: return it │
	// │                        │ z3.Const()          │ directly            │
	// ├────────────────────────┼─────────────────────┼─────────────────────┤
	// │ Python callable        │ native interp or    │ line 409: pred(*tup)│
	// │ (lambda/partial)       │ polymorphic macro   │                     │
	// ├────────────────────────┼─────────────────────┼─────────────────────┤
	// │ z3.FuncDeclRef         │ z3.Function() result│ line 410-411:       │
	// │                        │                     │ low-level Z3_mk_app │
	// └────────────────────────┴─────────────────────┴─────────────────────┘
	//
	// For the Go port, this naturally maps to an interface{} (or a
	// small sum type) since Go's Expr
	// and FuncDecl are distinct types, and you'd also need to
	// handle Go function values:
	//
	// func ApplyZ3Func(pred interface{}, tup []z3bridge.Expr) z3bridge.Expr {
	//     switch p := pred.(type) {
	//     case z3bridge.Expr:
	//         if p.ExprSort().Kind() == z3bridge.SortBool {
	//             // was isinstance(pred, z3.BoolRef)
	//             assert(len(tup) == 0)
	//             return p
	//         }
	//         // non-boolean Expr — falls through to "not FuncDeclRef" in Python,
	//         // which would try pred(*tup). This likely shouldn't happen in
	//         // practice, but faithfully it's an error here.
	//         panic("ApplyZ3Func: non-boolean Expr passed as pred")
	//         // was isinstance(pred, z3.BoolRef) — 0-ary constant
	//     case func([]z3bridge.Expr) z3bridge.Expr:
	//         // was "not isinstance(pred, z3.FuncDeclRef)" — native callable
	//         return p(tup)
	//     case z3bridge.FuncDecl:
	//         // low-level Z3_mk_app
	//         return p.Apply(tup...)
	//     }
	// }

	args := make([]Expr, len(terms))
	for i, term := range terms {
		a, err := t.TermToZ3(term)
		if err != nil {
			return Expr{}, err
		}
		args[i] = a
	}
	return pred(args...), nil
}

// TermToZ3 translates a term to Z3.
// Corresponds to Python term_to_z3 (ivy_solver.py:444).
func (t *Translator) TermToZ3(term lg.Expr) (Expr, error) {
	//xtracer.Trace("ivy_solver.py:445 term_to_z3() ENTER type=%s name=%s",
	//	iu.ShortTypeName(term), termName(term))

	// Python line 446: if is_boolean(term) and not is_variable(term):
	//     return formula_to_z3_int(term)
	if _, isBool := term.NodeSort().(*lg.BooleanSort); isBool {
		if _, isVar := term.(*lg.Variable); !isVar {
			return t.Formula_to_z3_int(term, "term_to_z3")
		}
	}

	// Python line 481-482: Ite in term context — cond through formula,
	// then/else through term path.
	if ite, ok := term.(*lg.Ite); ok {
		cond, err := t.Formula_to_z3_int(ite.Cond, "term_to_z3:ivy_logic.Ite")
		if err != nil {
			return Expr{}, err
		}
		th, err := t.TermToZ3(ite.Then)
		if err != nil {
			return Expr{}, err
		}
		el, err := t.TermToZ3(ite.Else)
		if err != nil {
			return Expr{}, err
		}
		return t.Ctx.Ite(cond, th, el), nil
	}

	// All other non-boolean terms: Variable, Const, non-boolean Apply
	return t.translateCore(term, "term_to_z3")
}

// termName extracts a name from a node for the term_to_z3 XTRACE.
// Python: getattr(term, 'name', getattr(term, 'rep', '?'))
func termName(n lg.Expr) string {
	switch e := n.(type) {
	case *lg.Variable:
		return e.Name
	case *lg.Const:
		return e.Name
	case *lg.Apply:
		if c, ok := e.Func.(*lg.Const); ok {
			return c.Name
		}
		return "?"
	default:
		return "?"
	}
}

// translateVariable translates an Ivy variable to a Z3 const named "name:sortName".
// Corresponds to Python term_to_z3 variable case (ivy_solver.py:418-433).
func (t *Translator) translateVariable(v *lg.Variable) (Expr, error) {
	sort := v.VSort
	sortName := sortDisplayName(sort)
	sksym := v.Name + ":" + sortName
	// Cache key uses structural identity
	key := lg.NodeKey(v.Name + ":" + string(sort.Sexp()))

	// Python: res = z3_constants.get(sksym)
	if cached, ok := t.cache.consts[key]; ok {
		return cached, nil
	}

	// Python line 454: sig = lookup_native(term.sort, sorts, "sort") if sorted else S
	var zs *Sort
	if result := t.LookupNative(sortName, sort, "sort"); result != nil {
		if zsVal, ok := result.(Sort); ok {
			zs = &zsVal
			// Cache the sort so TranslateSort finds it later via cache
			sortKey := lg.NodeKey(sortDisplayName(sort)) // Python: z3_sorts key is sort name
			if _, ok := t.cache.sorts[sortKey]; !ok {
				t.cache.sorts[sortKey] = zsVal
				t.cache.sortsInv[zsVal.GetId()] = sort
			}
		}
	}

	// Python line 455-456: if sig == None: sig = term.sort.to_z3()
	if zs == nil {
		xtracer.Trace("TranslateSort_call callsite=term_to_z3_variable")
		zsVal, err := t.TranslateSort(sort)
		if err != nil {
			return Expr{}, err
		}
		zs = &zsVal
	}

	// Python: res = z3.Const(sksym, sig); z3_constants[sksym] = res
	c := t.Ctx.Const(sksym, *zs)
	t.cache.consts[key] = c
	return c, nil
}

// sortDisplayName extracts the display name from a sort for use in Z3 constant
// naming (e.g., "T0:lclock"). Matches Python's term.sort.name.
func sortDisplayName(sort lg.Sort) string {
	switch s := sort.(type) {
	case *lg.UninterpretedSort:
		return s.Name
	case *lg.EnumeratedSort:
		return s.Name
	case *lg.RangeSort:
		return s.Name
	case *lg.BooleanSort:
		return "Bool"
	default:
		return string(sort.Sexp())
	}
}

// TranslateVar translates a Variable to a Z3 const without emitting a
// HASH trace. Matches Python's term_to_z3(v).
func (t *Translator) TranslateVar(v *lg.Variable) (Expr, error) {
	return t.translateVariable(v)
}

func (t *Translator) translateVarOrConst(name string, sort lg.Sort) (Expr, error) {
	//xtracer.Trace("ivy_solver.py:287 symbol_to_z3() ENTER name=%s sort=%s", name, sort)
	// Check for numeral with special handling (range sort clamping).
	// Python term_to_z3 lines 439-440: if term.is_numeral(): res = numeral_to_z3(term.rep)
	if isNumeralName(name) {
		if result, err := t.Numeral(name, sort); result != nil {
			return *result, err
		}
	}

	// Python term_to_z3() at ivy_solver.py:479 does:
	//     z3_constants.get(str(term.rep)).
	//
	// For enum constants pre-registered by enumeratedsort(), this cache
	// hit returns immediately without calling solver_name or to_z3.
	// Only enum constants get this early check because Python's z3_constants
	// cache only effectively works for enum constants (string-to-string key
	// match). Non-enum constants use Symbol object keys for storage but
	// string keys for lookup, so the cache never hits for them — solver_name
	// is always called. We must match that behavior.
	if _, isEnum := sort.(*lg.EnumeratedSort); isEnum {
		key := lg.NodeKey(name + ":" + string(sort.Sexp()))
		if cached, ok := t.cache.consts[key]; ok {
			return cached, nil
		}
	}

	// Python term_to_z3 evaluates `sig = iso.to_z3()` BEFORE
	// `z3.Const(solver_name(term.rep), sig)`, so the to_z3 trace
	// (functionsort/uninterpretedsort) fires before the solver_name trace.
	// We must mirror that order.
	if lg.FirstOrderSort(sort) {
		key := lg.NodeKey(name + ":" + string(sort.Sexp()))
		if cached, ok := t.cache.consts[key]; ok {
			return cached, nil
		}
		xtracer.Trace("TranslateSort_call callsite=term_to_z3_const")
		zs, err := t.TranslateSort(sort)
		if err != nil {
			return Expr{}, err
		}
		z3name := t.z3Name(name, sort)
		c := t.Ctx.Const(z3name, zs)
		t.cache.consts[key] = c
		return c, nil
	}

	// Function sort with single element (0-ary function) → treat as first-order
	if fs, ok := sort.(*lg.FunctionSort); ok && fs.Arity() == 0 {
		xtracer.Trace("TranslateSort_call callsite=symbol_to_z3_const")
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		key := lg.NodeKey(name + ":" + string(fs.Range().Sexp()))
		if cached, ok := t.cache.consts[key]; ok {
			return cached, nil
		}
		z3name := t.z3Name(name, sort)
		c := t.Ctx.Const(z3name, zs)
		t.cache.consts[key] = c
		return c, nil
	}

	// Higher-order bare Const: mirror Python symbol_to_z3
	// (ivy_solver.py:299-301), which creates a z3.Function inline and
	// does NOT cache anywhere. We call functionSort directly so the
	// "functionsort() ENTER" xtrace fires every time (matching Python,
	// where symbol_to_z3 has no cache). Going through makeFuncDecl
	// would write t.cache.z3_functions (Python's z3_functions cache), which
	// symbol_to_z3 does not populate in Python, and would also
	// suppress the trace on subsequent visits.
	if fs, ok := sort.(*lg.FunctionSort); ok {
		if _, err := t.functionSort(fs); err != nil { // emits functionsort() ENTER
			return Expr{}, err
		}
		xtracer.Trace("TranslateSort_call callsite=symbol_to_z3_func")
		zs, err := t.TranslateSort(fs.Range())
		if err != nil {
			return Expr{}, err
		}
		z3name := t.z3Name(name, sort)
		return t.Ctx.Const(z3name, zs), nil
	}

	return Expr{}, fmt.Errorf("cannot translate %s with sort %s to Z3", name, sort)
}

func (t *Translator) getFuncDecl(fn lg.Expr) (FuncDecl, error) {
	switch f := fn.(type) {
	case *lg.Const:
		fs, ok := f.CSort.(*lg.FunctionSort)
		if !ok {
			return FuncDecl{}, fmt.Errorf("expected FunctionSort for Apply func, got %T", f.CSort)
		}
		return t.makeFuncDecl(f.Name, fs)

	case *lg.Variable:
		fs, ok := f.VSort.(*lg.FunctionSort)
		if !ok {
			return FuncDecl{}, fmt.Errorf("expected FunctionSort for Apply func, got %T", f.VSort)
		}
		return t.makeFuncDecl(f.Name, fs)

	default:
		return FuncDecl{}, fmt.Errorf("unsupported Apply func type: %T", fn)
	}
}

// functionSort mirrors Python's functionsort (ivy_solver.py:278-283).
// Translates a FunctionSort to its list of Z3 sorts: [domain..., range].
// ALWAYS emits the functionsort() ENTER trace — there is no caching
// here, matching Python's monkey-patched FunctionSort.to_z3 = functionsort
// (ivy_solver.py:304). Callers that want caching must wrap this themselves
// (e.g. makeFuncDecl). Callers that need to mirror Python's uncached
// callsites (e.g. lt_pred) must call this helper directly so the trace
// fires every time.
func (t *Translator) functionSort(fs *lg.FunctionSort) ([]Sort, error) {
	xtracer.Trace("ivy_solver.py:279 functionsort() ENTER")

	domain := fs.Domain()
	sig := make([]Sort, 0, len(domain)+1)
	for _, d := range domain {
		xtracer.Trace("TranslateSort_call callsite=functionsort_dom")
		zs, err := t.TranslateSort(d)
		if err != nil {
			return nil, err
		}
		sig = append(sig, zs)
	}

	// Python:
	//   if fs.is_relational(): return [s.to_z3() for s in fs.dom] + [z3.BoolSort()]
	//   else:                  return [s.to_z3() for s in fs.dom] + [fs.rng.to_z3()]
	// Note: Python short-circuits the rng.to_z3() call when relational,
	// so no extra trace fires for the Boolean range. We mirror that.
	var rng Sort
	if il.IsRelationalSort(fs) {
		rng = t.Ctx.BoolSort()
	} else {
		xtracer.Trace("TranslateSort_call callsite=functionsort_rng")
		var err error
		rng, err = t.TranslateSort(fs.Range())
		if err != nil {
			return nil, err
		}
	}
	sig = append(sig, rng)
	return sig, nil
}

// makeFuncDecl returns a Z3 FuncDecl for (name, fs), caching the
// result in t.cache.z3_functions.
//
// This is the TERM-PATH helper only — the Go analog of Python's
// z3_functions cache (ivy_solver.py:500-508). Callers on the atom
// path (atomToZ3) and the symbol-bare path (translateVarOrConst
// higher-order branch) MUST NOT call this helper; they must build
// the FuncDecl inline via t.functionSort + t.Ctx.Function so that
// Python's independent z3_predicates / symbol_to_z3 xtrace semantics
// are preserved. Sharing this cache across paths suppresses the
// functionsort() ENTER trace on downstream misses and diverges from
// Python. See atomToZ3, translateVarOrConst, and ltPred for the
// inline-creation precedent.
func (t *Translator) makeFuncDecl(name string, fs *lg.FunctionSort) (FuncDecl, error) {
	key := lg.NodeKey(name + ":" + string(fs.Sexp()))
	if cached, ok := t.cache.z3_functions[key]; ok {
		return cached, nil
	}

	// Python: sig = atom.rep.sort.to_z3() calls functionsort(fs) first,
	// then solver_name(atom.rep). Match that order.
	sig, err := t.functionSort(fs)
	if err != nil {
		return FuncDecl{}, err
	}
	zDomain := sig[:len(sig)-1]
	zRange := sig[len(sig)-1]

	z3name := t.z3Name(name, fs)

	fd := t.Ctx.Function(z3name, zDomain, zRange)
	t.cache.z3_functions[key] = fd
	return fd, nil
}

func (t *Translator) translateQuantifier(isForall bool, variables []*lg.Variable, body lg.Expr) (Expr, error) {
	if len(variables) == 0 {
		return t.Formula_to_z3_int(body, "translateQuantifier() no variables")
	}

	// Create Z3 constants for the bound variables
	bound := make([]Expr, len(variables))
	for i, v := range variables {
		xtracer.Trace("TranslateSort_call callsite=translate_quantifier_var")
		zs, err := t.TranslateSort(v.VSort)
		if err != nil {
			return Expr{}, err
		}
		z3Var, err := t.translateVariable(v)
		if err != nil {
			return Expr{}, err
		}
		key := lg.NodeKey(v.Name + ":" + string(v.VSort.Sexp()))
		bound[i] = z3Var
		_ = zs
		// Temporarily override the const cache so the body uses these bound vars
		t.cache.consts[key] = bound[i]
	}

	zBody, err := t.Formula_to_z3_int(body, "translateQuantifier() len(variables) > 0")
	if err != nil {
		return Expr{}, err
	}

	// Validate that the body is a Bool expression before wrapping with
	// constraints or passing to Z3, which would otherwise panic with a
	// type error.
	if zBody.ExprSort().Kind() != SortBool {
		return Expr{}, fmt.Errorf("quantifier body must be Bool, got sort %s", zBody.ExprSort().String())
	}

	// Delegate to forall/exists which call quantConstraints.
	// Matches Python: formula_to_z3_int calls forall()/exists() after
	// translating body and variables.
	if isForall {
		return t.forall(variables, bound, zBody), nil
	}
	return t.exists(variables, bound, zBody), nil
}

// --- Convenience functions ---

// Implies checks if f1 implies f2 using Z3.
// Returns true if f1 => f2 is valid (i.e., f1 && !f2 is unsatisfiable).
func (t *Translator) Implies(f1, f2 lg.Expr) (bool, error) {
	zf1, err := t.Formula_to_z3_int(f1, "zf1 translate.go:921 Implies")
	if err != nil {
		return false, err
	}
	zf2, err := t.Formula_to_z3_int(f2, "zf2 translate.go:925 Implies")
	if err != nil {
		return false, err
	}

	s := t.Ctx.NewZ3Solver()
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
func (t *Translator) translateBuiltinOp(name string, terms []lg.Expr) (Expr, bool, error) {
	// Translate arguments
	translateArgs := func() ([]Expr, error) {
		args := make([]Expr, len(terms))
		for i, term := range terms {
			a, err := t.TermToZ3(term)
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
	// Comparison operators (<, <=, >, >=) are NOT handled here.
	// They are polymorphic: for interpreted sorts (int, nat, bv) they
	// use Z3 built-in comparisons via LookupNative → lookupBuiltinRelation;
	// for uninterpreted sorts they become sort-qualified uninterpreted
	// functions via getFuncDecl (e.g., "<:lclock:lclock").
	// Handling them here with hardcoded ctx.Lt/Le/Gt/Ge causes Z3 Sort
	// mismatch panics on uninterpreted sorts.
	// Matches Python ivy_solver.py:atom_to_z3 which dispatches via
	// lookup_native → relations_dict for interpreted sorts, or creates
	// z3.Function(solver_name(sym), *sig) for uninterpreted sorts.

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

// parseBfeParams parses "bfe[lo:hi]", "bfe[lo,hi]", or "bfe[lo][hi]" and returns (lo, hi, ok).
// Python uses bfe[lo][hi] format via parse_int_params.
func parseBfeParams(name string) (int, int, bool) {
	// Expected format: bfe[lo:hi] or bfe[lo,hi] or bfe[lo][hi]
	inner := name[4:] // skip "bfe["
	if len(inner) < 2 || inner[len(inner)-1] != ']' {
		return 0, 0, false
	}
	inner = inner[:len(inner)-1] // strip trailing "]"
	sep := -1
	for i, c := range inner {
		if c == ':' || c == ',' {
			sep = i
			break
		}
	}
	if sep < 0 {
		// Try bfe[lo][hi] format: inner is "lo][hi" after stripping last ']'
		bracket := strings.Index(inner, "][")
		if bracket >= 0 {
			var lo, hi int
			_, err1 := fmt.Sscanf(inner[:bracket], "%d", &lo)
			_, err2 := fmt.Sscanf(inner[bracket+2:], "%d", &hi)
			if err1 != nil || err2 != nil {
				return 0, 0, false
			}
			return lo, hi, true
		}
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
func (t *Translator) IsSat(f lg.Expr) (CheckResult, error) {
	zf, err := t.Formula_to_z3_int(f, "IsSat")
	if err != nil {
		return Unknown, err
	}
	s := t.Ctx.NewZ3Solver()
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

// isNumeralName returns true if name starts with a digit.
// Matches Python's ivy_logic.is_numeral check for symbol names.
func isNumeralName(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9'
}
