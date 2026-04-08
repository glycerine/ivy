package z3bridge

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	"github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// LookupNativeFunc is a callback matching Python lookup_native(thing, table, kind)
// (ivy_solver.py:311). Returns any: z3bridge.Sort for sort lookups,
// func(args ...Expr) Expr for function/relation lookups, or nil.
type LookupNativeFunc func(name string, sort logic.Sort, kind string) any

// SolverNameFunc maps an Ivy symbol to its Z3 name. Returns "" if the
// symbol should be handled natively (not declared as an uninterpreted function).
type SolverNameFunc func(name string, sort logic.Sort) string

// QuantConstraintsFn is a callback that generates Z3 constraints for
// quantifier-bound variables based on their sort (e.g., nat non-negativity,
// range sort bounds). Corresponds to Python's quant_constraints.
type QuantConstraintsFn func(v *logic.Variable, z3Var Expr) []Expr

// EqFuncFn is a callback for custom equality (e.g., MyEq with True/False optimization).
// Corresponds to Python's my_eq (ivy_solver.py:88-95).
type EqFuncFn func(x, y Expr) Expr

// EnumEqFuncFn is a callback for custom enumerated sort equality (binary encoding).
// Returns non-nil Expr to override default equality; nil to use default.
// Corresponds to Python's encode_equality dispatch in atom_to_z3 (ivy_solver.py:484).
type EnumEqFuncFn func(t1, t2 logic.Expr, sort *logic.EnumeratedSort) (*Expr, error)

// NumeralFuncFn is a callback for custom numeral handling (range sort clamping).
// Returns non-nil Expr to override default; nil to use default.
// Corresponds to Python's numeral_to_z3 (ivy_solver.py:388-404).
type NumeralFuncFn func(name string, sort logic.Sort) (*Expr, error)

// Translator converts Ivy logic nodes to Z3 expressions.
type Translator struct {
	Ctx              *Z3Context
	sorts            map[logic.NodeKey]Sort                    // cache: Ivy sort Sexp -> Z3 sort
	sortsInv         map[uint]logic.Sort                       // reverse map: Z3 AST ID -> original Ivy sort (Python z3_sorts_inv)
	consts           map[logic.NodeKey]Expr                    // cache: structural key -> Z3 const
	funcs            map[logic.NodeKey]FuncDecl                // cache: structural key -> Z3 func decl
	preds            map[logic.NodeKey]func(args ...Expr) Expr // cache: z3_predicates (Python z3_predicates)
	LookupNative     LookupNativeFunc                          // optional: Python lookup_native(thing, table, kind) callback
	SolverName       SolverNameFunc                            // optional: maps symbol to Z3 name (for polymorphic disambiguation)
	QuantConstraints QuantConstraintsFn                        // optional: generates sort constraints for quantifier-bound variables
	EqFunc           EqFuncFn                                  // optional: custom equality (MyEq True/False optimization)
	EnumEqFunc       EnumEqFuncFn                              // optional: custom enumerated equality (binary encoding)
	NumeralFunc      NumeralFuncFn                             // optional: custom numeral handling (range clamping)
	TranslateMerkle  iu.MerkleState                            // rolling Merkle hash for Translate() input conformance
	translateDepth   int                                       // nesting depth; only hash at top level (depth 0)
}

// NewTranslator creates a translator with a fresh Z3 context.
func NewTranslator() *Translator {
	t := &Translator{
		Ctx:      NewZ3Context(),
		sorts:    make(map[logic.NodeKey]Sort),
		sortsInv: make(map[uint]logic.Sort),
		consts:   make(map[logic.NodeKey]Expr),
		funcs:    make(map[logic.NodeKey]FuncDecl),
		preds:    make(map[logic.NodeKey]func(args ...Expr) Expr),
	}
	t.initEqPred()
	return t
}

func (t *Translator) Close() error {
	return t.Ctx.Close()
}

// Clear resets all Z3 caches to initial state.
// Corresponds to Python ivy_solver.clear() (line 228).
func (t *Translator) Clear() {
	t.sorts = make(map[logic.NodeKey]Sort)
	t.sortsInv = make(map[uint]logic.Sort)
	t.consts = make(map[logic.NodeKey]Expr)
	t.funcs = make(map[logic.NodeKey]FuncDecl)
	t.preds = make(map[logic.NodeKey]func(args ...Expr) Expr)
	t.initEqPred()
}

// initEqPred pre-populates the equality predicate in preds, matching
// Python's clear() which initializes z3_predicates = {ivy_logic.equals: my_eq}.
func (t *Translator) initEqPred() {
	t.preds[eqCanonPredKey] = func(args ...Expr) Expr {
		if t.EqFunc != nil {
			return t.EqFunc(args[0], args[1])
		}
		return t.Ctx.Eq(args[0], args[1])
	}
}

// SortFromZ3 looks up the original Ivy sort for a Z3 sort using the reverse map.
// Corresponds to Python's sort_from_z3() (ivy_solver.py:905).
func (t *Translator) SortFromZ3(z3sort Sort) (logic.Sort, bool) {
	ivySort, ok := t.sortsInv[z3sort.GetId()]
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
func (t *Translator) TranslateSort(s logic.Sort) (Sort, error) {
	switch st := s.(type) {
	case *logic.BooleanSort:
		// Python z3.BoolSort() has no custom trace
		return t.Ctx.BoolSort(), nil

	case *logic.UninterpretedSort:
		// Python: uninterpretedsort(us) at ivy_solver.py:257
		// Python: uninterpretedsort(us) at ivy_solver.py:257
		xtracer.Trace("ivy_solver.py:258 uninterpretedsort() ENTER name=%s", st.Name)
		key := logic.NodeKey(st.Name) // Python: z3_sorts[us.rep] where rep = name
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		// Python: s = lookup_native(us, sorts, "sort")
		if t.LookupNative != nil {
			if result := t.LookupNative(st.Name, s, "sort"); result != nil {
				if zs, ok := result.(Sort); ok {
					t.sorts[key] = zs
					t.sortsInv[zs.GetId()] = s
					return zs, nil
				}
			}
		}
		// Python: if s == None: s = z3.DeclareSort(us.rep)
		zs := t.Ctx.UninterpretedSort(st.Name)
		t.sorts[key] = zs
		// Python: z3_sorts_inv[get_id(s)] = us
		t.sortsInv[zs.GetId()] = s
		return zs, nil

	case *logic.TopSort:
		// Python has no TopSort.to_z3 — calling it would raise AttributeError.
		// Panic here to match Python's behavior and expose bugs.
		panic(fmt.Sprintf("TranslateSort: TopSort %q has no to_z3() equivalent in Python", st.Name))

	case *logic.FunctionSort:
		xtracer.Trace("ivy_solver.py:269 functionsort() ENTER")
		return Sort{}, fmt.Errorf("FunctionSorts are not directly converted to Z3 sorts")

	case *logic.EnumeratedSort:
		// Python: enumeratedsort(es) at ivy_solver.py:275
		xtracer.Trace("ivy_solver.py:276 enumeratedsort() ENTER name=%s", st.Name)
		key := logic.NodeKey(st.Name) // Python: z3_sorts[es.rep] where rep = name
		if cached, ok := t.sorts[key]; ok {
			return cached, nil
		}
		// Use native Z3 EnumSort, matching Python's z3.EnumSort(name, extension).
		zs, constExprs := t.Ctx.EnumSort(st.Name, st.Extension)
		t.sorts[key] = zs
		// Python enumeratedsort() does NOT store in z3_sorts_inv.
		// The reverse map entry is created by uninterpretedsort (the normal
		// entry point for interpreted sorts).
		// Register the constructor constants so they can be looked up by name.
		for i, name := range st.Extension {
			constKey := logic.NodeKey(name + ":" + string(s.Sexp()))
			t.consts[constKey] = constExprs[i]
		}
		return zs, nil

	case *logic.RangeSort:
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
func (t *Translator) TranslateNoHash(n logic.Expr) (Expr, error) {
	t.translateDepth++
	defer func() { t.translateDepth-- }()
	return t.Translate(n)
}

// Translate converts an Ivy logic node to a Z3 expression.
// Corresponds to Python formula_to_z3_int (ivy_solver.py:637).
// Dispatches Boolean-sorted Apply nodes to atomToZ3 (Python atom_to_z3).
func (t *Translator) Translate(n logic.Expr) (Expr, error) {
	xtracer.Trace("ivy_solver.py:638 formula_to_z3_int() ENTER type=%v", iu.ShortTypeName(n))
	if xtracer.Enabled && t.translateDepth == 0 {
		canon := iu.Canonical(n.Sexp())
		leaf, root := t.TranslateMerkle.AddLeaf(canon)
		xtracer.Trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s", leaf, root, string(canon))
	}
	t.translateDepth++
	defer func() { t.translateDepth-- }()

	// Python line 645: if ivy_logic.is_atom(fmla): return atom_to_z3(fmla)
	// is_atom: (Apply or Symbol) with Boolean sort, or Eq.
	if app, ok := n.(*logic.Apply); ok && len(app.Terms) > 0 {
		if _, isBool := n.NodeSort().(*logic.BooleanSort); isBool {
			return t.atomToZ3(app)
		}
	}
	// Python: isinstance(term, lg.Eq) in is_atom → atom_to_z3(fmla)
	// Eq has duck-typed .rep/.args/.relname in Python; we construct a
	// pseudo-Apply to pass through the same atomToZ3 code path.
	if eq, ok := n.(*logic.Eq); ok {
		return t.eqToAtomZ3(eq)
	}

	return t.translateCore(n)
}

// translateCore is the inner dispatch for Translate. It handles all node
// types except Boolean-sorted Apply (which is routed to atomToZ3 by
// Translate). No XTRACE, no Merkle hash, no depth tracking — those are
// done by the caller (Translate or TermToZ3).
func (t *Translator) translateCore(n logic.Expr) (Expr, error) {
	switch node := n.(type) {
	case *logic.Variable:
		// Python: sksym = term.rep + ':' + term.sort.name
		// Variables use "name:sortName" as their Z3 const name.
		return t.translateVariable(node)

	case *logic.Const:
		return t.translateVarOrConst(node.Name, node.CSort)

	case *logic.Apply:
		if len(node.Terms) == 0 {
			// Nullary application: convert func directly
			return t.Translate(node.Func)
		}

		// Non-Boolean Apply: corresponds to Python term_to_z3 "else"
		// branch (line 483–497) for function applications.
		if c, ok := node.Func.(*logic.Const); ok {
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
			if t.LookupNative != nil {
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

	case *logic.Eq:
		// Eq is now routed through atomToZ3 by Translate() (matching Python's
		// is_atom dispatch). This case should be unreachable.
		return Expr{}, fmt.Errorf("translateCore: unexpected Eq (should be routed through atomToZ3 by Translate)")

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
		// Python uses my_eq for Iff (formula_to_z3_int line 620)
		if t.EqFunc != nil {
			return t.EqFunc(t1, t2), nil
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

	case *logic.Definition:
		// Python formula_to_z3_int lines 589-615, 596-597, 609-614
		// Check for enumerated encoding (when !UseZ3Enums)
		if t.EnumEqFunc != nil {
			if es, ok := node.Lhs.NodeSort().(*logic.EnumeratedSort); ok {
				if result, err := t.EnumEqFunc(node.Lhs, node.Rhs, es); result != nil {
					return *result, err
				}
			}
		}
		t1, err := t.Translate(node.Lhs)
		if err != nil {
			return Expr{}, err
		}
		t2, err := t.Translate(node.Rhs)
		if err != nil {
			return Expr{}, err
		}
		if t.EqFunc != nil {
			return t.EqFunc(t1, t2), nil
		}
		return t.Ctx.Eq(t1, t2), nil

	case *logic.ForAll:
		return t.translateQuantifier(true, node.Variables, node.Body)

	case *logic.Exists:
		return t.translateQuantifier(false, node.Variables, node.Body)

	default:
		return Expr{}, fmt.Errorf("unsupported node type for Z3 translation: %T", n)
	}
}

const eqCanonPredKey = logic.NodeKey("=:eq")

// atomToZ3 translates a Boolean-sorted Apply (atom) to Z3.
// Corresponds to Python atom_to_z3 (ivy_solver.py:516).
func (t *Translator) atomToZ3(app *logic.Apply) (Expr, error) {
	c, ok := app.Func.(*logic.Const)
	if !ok {
		// Fallback for non-Const func (rare)
		return t.translateCore(app)
	}

	xtracer.Trace("ivy_solver.py:517 atom_to_z3() ENTER rep=%s nargs=%d",
		c.Name, len(app.Terms))

	// Python line 518: if ivy_logic.is_equals(atom.rep) and
	//   ivy_logic.is_enumerated(atom.args[0]) and not use_z3_enums
	//
	//
	if c.Name == "=" && len(app.Terms) == 2 && t.EnumEqFunc != nil {
		if es, ok2 := app.Terms[0].NodeSort().(*logic.EnumeratedSort); ok2 {
			if result, err := t.EnumEqFunc(app.Terms[0], app.Terms[1], es); result != nil {
				return *result, err
			}
		}
	}

	// Check preds cache (Python z3_predicates)
	// Python: atom.relname for Eq is always equals = Symbol('=', RelationSort([TopSort(), TopSort()]))
	// regardless of concrete sorts. Use a canonical key to match Python's caching.
	// (all equalities need to share one cache entry, matching
	// Python's z3_predicates[atom.relname] behavior)
	var predKey logic.NodeKey
	if c.Name == "=" {
		predKey = eqCanonPredKey
	} else {
		predKey = logic.NodeKey(c.Name + ":" + string(c.CSort.Sexp()))
	}
	if cached, ok := t.preds[predKey]; ok {
		return t.applyZ3Func(cached, app.Terms)
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
	if t.LookupNative != nil {
		if result := t.LookupNative(c.Name, c.CSort, "relation"); result != nil {
			if nativeFn, ok := result.(func(args ...Expr) Expr); ok {
				t.preds[predKey] = nativeFn
				return t.applyZ3Func(nativeFn, app.Terms)
			}
		}
	}

	// Python's z3.Function("=", ...) returns Z3's built-in equality.
	// Go's makeFuncDecl would create an uninterpreted function instead.
	// Use Ctx.Eq (or EqFunc) for "=" to match Python/Z3 semantics.
	if c.Name == "=" {
		eqFn := func(args ...Expr) Expr {
			if t.EqFunc != nil {
				return t.EqFunc(args[0], args[1])
			}
			return t.Ctx.Eq(args[0], args[1])
		}
		t.preds[predKey] = eqFn
		return t.applyZ3Func(eqFn, app.Terms)
	}

	// Python lines 527-528: create Z3 Function/Const for uninterpreted relation
	fs, ok := c.CSort.(*logic.FunctionSort)
	if !ok {
		return Expr{}, fmt.Errorf("atomToZ3: expected FunctionSort for %s, got %T", c.Name, c.CSort)
	}
	fd, err := t.makeFuncDecl(c.Name, fs)
	if err != nil {
		return Expr{}, err
	}
	predFn := fd.Apply
	t.preds[predKey] = predFn
	return t.applyZ3Func(predFn, app.Terms)
}

// eqToAtomZ3 routes Eq through atomToZ3, matching Python where
// is_atom() returns True for Eq and atom_to_z3 accesses Eq via
// duck-typed properties:
//
//	atom.rep = Symbol('=', RelationSort([x.sort for x in self.args]))
//	atom.args = [t1, t2]
//	atom.relname = equals (ivy_logic.py:1145-1147)
func (t *Translator) eqToAtomZ3(eq *logic.Eq) (Expr, error) {
	s1, s2 := eq.T1.NodeSort(), eq.T2.NodeSort()
	repSort, err := logic.NewFunctionSort(s1, s2, logic.Boolean)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	eqConst := logic.NewConst("=", repSort)
	pseudoApp, err := logic.NewApply(eqConst, eq.T1, eq.T2)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	return t.atomToZ3(pseudoApp)
}

// applyZ3Func translates args via TermToZ3 and applies the predicate function.
// Corresponds to Python apply_z3_func (ivy_solver.py:403).
func (t *Translator) applyZ3Func(pred func(args ...Expr) Expr, terms []logic.Expr) (Expr, error) {
	args := make([]Expr, len(terms))
	for i, term := range terms {
		a, err := t.TermToZ3(term)
		if err != nil {
			return Expr{}, err
		}
		args[i] = a
	}
	xtracer.Trace("ivy_solver.py:404 apply_z3_func() ENTER nargs=%d", len(args))
	return pred(args...), nil
}

// TermToZ3 translates a term to Z3.
// Corresponds to Python term_to_z3 (ivy_solver.py:444).
func (t *Translator) TermToZ3(term logic.Expr) (Expr, error) {
	xtracer.Trace("ivy_solver.py:445 term_to_z3() ENTER type=%s name=%s",
		iu.ShortTypeName(term), termName(term))

	// Python line 446: if is_boolean(term) and not is_variable(term):
	//     return formula_to_z3_int(term)
	if _, isBool := term.NodeSort().(*logic.BooleanSort); isBool {
		if _, isVar := term.(*logic.Variable); !isVar {
			return t.Translate(term)
		}
	}

	// Python line 481-482: Ite in term context — cond through formula,
	// then/else through term path.
	if ite, ok := term.(*logic.Ite); ok {
		cond, err := t.Translate(ite.Cond)
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
	return t.translateCore(term)
}

// termName extracts a name from a node for the term_to_z3 XTRACE.
// Python: getattr(term, 'name', getattr(term, 'rep', '?'))
func termName(n logic.Expr) string {
	switch e := n.(type) {
	case *logic.Variable:
		return e.Name
	case *logic.Const:
		return e.Name
	case *logic.Apply:
		if c, ok := e.Func.(*logic.Const); ok {
			return c.Name
		}
		return "?"
	default:
		return "?"
	}
}

// translateVariable translates an Ivy variable to a Z3 const named "name:sortName".
// Corresponds to Python term_to_z3 variable case (ivy_solver.py:418-433).
func (t *Translator) translateVariable(v *logic.Variable) (Expr, error) {
	sort := v.VSort
	sortName := sortDisplayName(sort)
	sksym := v.Name + ":" + sortName
	// Cache key uses structural identity
	key := logic.NodeKey(v.Name + ":" + string(sort.Sexp()))

	// Python: res = z3_constants.get(sksym)
	if cached, ok := t.consts[key]; ok {
		return cached, nil
	}

	// Python line 454: sig = lookup_native(term.sort, sorts, "sort") if sorted else S
	var zs *Sort
	if t.LookupNative != nil {
		if result := t.LookupNative(sortName, sort, "sort"); result != nil {
			if zsVal, ok := result.(Sort); ok {
				zs = &zsVal
				// Cache the sort so TranslateSort finds it later via cache
				sortKey := logic.NodeKey(sortDisplayName(sort)) // Python: z3_sorts key is sort name
				if _, ok := t.sorts[sortKey]; !ok {
					t.sorts[sortKey] = zsVal
					t.sortsInv[zsVal.GetId()] = sort
				}
			}
		}
	}

	// Python line 455-456: if sig == None: sig = term.sort.to_z3()
	if zs == nil {
		zsVal, err := t.TranslateSort(sort)
		if err != nil {
			return Expr{}, err
		}
		zs = &zsVal
	}

	// Python: res = z3.Const(sksym, sig); z3_constants[sksym] = res
	c := t.Ctx.Const(sksym, *zs)
	t.consts[key] = c
	return c, nil
}

// sortDisplayName extracts the display name from a sort for use in Z3 constant
// naming (e.g., "T0:lclock"). Matches Python's term.sort.name.
func sortDisplayName(sort logic.Sort) string {
	switch s := sort.(type) {
	case *logic.UninterpretedSort:
		return s.Name
	case *logic.EnumeratedSort:
		return s.Name
	case *logic.RangeSort:
		return s.Name
	case *logic.BooleanSort:
		return "Bool"
	default:
		return string(sort.Sexp())
	}
}

// TranslateVar translates a Variable to a Z3 const without emitting a
// HASH trace. Matches Python's term_to_z3(v).
func (t *Translator) TranslateVar(v *logic.Variable) (Expr, error) {
	return t.translateVariable(v)
}

func (t *Translator) translateVarOrConst(name string, sort logic.Sort) (Expr, error) {
	xtracer.Trace("ivy_solver.py:287 symbol_to_z3() ENTER name=%s sort=%s", name, sort)
	// Check for numeral with special handling (range sort clamping).
	// Python term_to_z3 lines 439-440: if term.is_numeral(): res = numeral_to_z3(term.rep)
	if t.NumeralFunc != nil && isNumeralName(name) {
		if result, err := t.NumeralFunc(name, sort); result != nil {
			return *result, err
		}
	}
	z3name := t.z3Name(name, sort)
	if logic.FirstOrderSort(sort) {
		key := logic.NodeKey(name + ":" + string(sort.Sexp()))
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
		key := logic.NodeKey(name + ":" + string(fs.Range().Sexp()))
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
	case *logic.Const:
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
	key := logic.NodeKey(name + ":" + string(fs.Sexp()))
	if cached, ok := t.funcs[key]; ok {
		return cached, nil
	}

	// Python: sig = atom.rep.sort.to_z3() calls functionsort(fs) first,
	// then solver_name(atom.rep). Match that order.
	xtracer.Trace("ivy_solver.py:269 functionsort() ENTER")

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

	z3name := t.z3Name(name, fs)

	fd := t.Ctx.Function(z3name, zDomain, zRange)
	t.funcs[key] = fd
	return fd, nil
}

func (t *Translator) translateQuantifier(isForall bool, variables []*logic.Variable, body logic.Expr) (Expr, error) {
	if isForall {
		xtracer.Trace("ivy_solver.py:560 forall() ENTER nvars=%d", len(variables))
	} else {
		xtracer.Trace("ivy_solver.py:567 exists() ENTER nvars=%d", len(variables))
	}
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
		key := logic.NodeKey(v.Name + ":" + string(v.VSort.Sexp()))
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

// isNumeralName returns true if name starts with a digit.
// Matches Python's ivy_logic.is_numeral check for symbol names.
func isNumeralName(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9'
}
