package ivy2cpp

import (
	"github.com/glycerine/ivy/goivy"
)

// extensionalRels returns the cached set of extensional relation names.
// First call populates the cache by running extensionalRelations.
// Mirrors Python ivy_to_cpp.py:1912-1913 `the_extensional_relations`,
// scoped per Generator per goivy/CLAUDE.md section C.
func (g *Generator) extensionalRels() map[string]bool {
	if g == nil {
		return nil
	}
	if g.extRel != nil {
		return g.extRel
	}
	g.extRel = g.extensionalRelations()
	if g.extRel == nil {
		g.extRel = map[string]bool{}
	}
	return g.extRel
}

// extensionalRelations mirrors Python ivy_to_cpp.py:44-76.
//
// A relation is extensional if:
//  1. it is in allStateSymbols (i.e. not a constructor and not a
//     solver-interpreted symbol),
//  2. it is initialized to all-false in some module initializer, and
//  3. every update of it (in any action body) is either to false, or at
//     a simple point (no LHS argument is a variable).
//
// Destructor-sort symbols are excluded inline at every check, matching
// Python's `lhs.rep.name not in im.module.destructor_sorts` guards.
func (g *Generator) extensionalRelations() map[string]bool {
	if g == nil || g.Mod == nil {
		return nil
	}

	bad := map[string]bool{}
	if g.Mod.Actions != nil {
		for _, action := range g.Mod.Actions.All() {
			if action == nil {
				continue
			}
			for _, sub := range action.IterSubactions() {
				g.markBadExtensional(sub, bad)
			}
		}
	}

	inited := map[string]bool{}
	for _, ini := range g.Mod.Initializers {
		g.collectInitedExtensional(ini.Action, inited)
	}
	// When the module pipeline ran without create_isolate / fix_initializers
	// (e.g. unit tests that call SourceString with create_isolate=false),
	// after-init bodies remain in Mod.Actions under names referenced by
	// Mod.Mixins.Get("init") instead of being moved to Mod.Initializers.
	// Treat those as initializers here, mirroring what fix_initializers
	// would have produced. The fallback matches emitInit (generator.go:663).
	if len(g.Mod.Initializers) == 0 && g.Mod.Actions != nil && g.Mod.Mixins != nil {
		for _, mixin := range g.Mod.Mixins.Get("init") {
			if act, ok := g.Mod.Actions.Get2(mixin.Mixer()); ok {
				if a, ok := act.(goivy.Action); ok {
					g.collectInitedExtensional(a, inited)
				}
			}
		}
	}

	res := map[string]bool{}
	for _, sym := range g.allStateSymbols() {
		if inited[sym.Name] && !bad[sym.Name] {
			res[sym.Name] = true
		}
	}
	return res
}

// markBadExtensional marks the LHS relation of an offending update as
// non-extensional. Mirrors the three isinstance branches in Python
// ivy_to_cpp.py:46-62.
func (g *Generator) markBadExtensional(sub goivy.Action, bad map[string]bool) {
	switch a := sub.(type) {
	case *goivy.LogicAssignAction:
		name := repName(a.LHS)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		// Python: not(il.is_false(rhs) or all(not il.is_variable(a) for a in lhs.args))
		if !goivy.IsFalse(a.RHS) && !allArgsNonVariable(argsOf(a.LHS)) {
			bad[name] = true
		}
	case *goivy.LogicHavocAction:
		name := repName(a.Target)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		if !allArgsNonVariable(argsOf(a.Target)) {
			bad[name] = true
		}
	case *goivy.LogicCallAction:
		// Python: for lhs in sub.args[1:]: ... — i.e. each actual return.
		for _, ret := range a.ActualReturns {
			name := repName(ret)
			if name == "" || g.isDestructorSort(name) {
				continue
			}
			if !allArgsNonVariable(argsOf(ret)) {
				bad[name] = true
			}
		}
	}
}

// collectInitedExtensional mirrors Python ivy_to_cpp.py:63-74 `get_inited`.
// It recurses through LogicSequence and pulls out each top-level
// LogicAssignAction whose RHS is false and whose LHS args are all
// variables. It does NOT descend into If / While / Choice etc.; the
// Python version recurses only on Sequence.
func (g *Generator) collectInitedExtensional(action goivy.Action, inited map[string]bool) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *goivy.LogicSequence:
		for _, elem := range a.Elems {
			if act, ok := elem.(goivy.Action); ok {
				g.collectInitedExtensional(act, inited)
			}
		}
	case *goivy.LogicAssignAction:
		name := repName(a.LHS)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		if goivy.IsFalse(a.RHS) && allArgsVariable(argsOf(a.LHS)) {
			inited[name] = true
		}
	}
}

// isDestructorSort reports whether a name is in module.DestructorSorts.
// Guarded against a nil map (goivy.Module.DestructorSorts may be nil for
// modules without destructors).
func (g *Generator) isDestructorSort(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return false
	}
	_, ok := g.Mod.DestructorSorts[name]
	return ok
}

// repName returns the head symbol name of an expression: the Const name
// for a bare Const, or the function name for an Apply. Returns "" for
// anything else (including a bare LogicVariable, since variables have
// no "rep" in Python's sense).
func repName(e goivy.Expr) string {
	switch n := e.(type) {
	case *goivy.Const:
		return n.Name
	case *goivy.Apply:
		return goivy.ExprName(n.Func)
	}
	return ""
}

// argsOf returns the term-args of an Apply, or nil for a bare Const or
// other non-application Expr. Used to walk LHS arguments of assignments
// and havocs.
func argsOf(e goivy.Expr) []goivy.Expr {
	if app, ok := e.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

// allArgsNonVariable mirrors Python `all(not il.is_variable(a) for a in args)`.
// An empty slice satisfies the predicate vacuously, matching Python.
func allArgsNonVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); isVar {
			return false
		}
	}
	return true
}

// allArgsVariable mirrors Python `all(il.is_variable(a) for a in args)`.
// An empty slice satisfies the predicate vacuously, matching Python.
func allArgsVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); !isVar {
			return false
		}
	}
	return true
}

// emitExtensionalRelationClear handles the special case of resetting an
// extensional, hash_thunk-backed relation to all-false: when the LHS is
// `r(X1, ..., Xn)` with every Xi a variable, the RHS is logical false,
// and r is in extensionalRels(), the assignment becomes `r.memo.clear();`.
//
// Python's emit_assign falls through to emit_assign_large for this
// shape (ivy_to_cpp.py:3703-3724), which builds a thunk evaluating to
// false. For a default-constructed hash_thunk (fun=nullptr, operator[]
// returns R() = false), clearing memo is observationally equivalent.
//
// Returns true when the assignment was handled.
func (g *Generator) emitExtensionalRelationClear(w *cppWriter, a *goivy.LogicAssignAction) bool {
	if a == nil {
		return false
	}
	app, ok := a.LHS.(*goivy.Apply)
	if !ok {
		return false
	}
	name := goivy.ExprName(app.Func)
	if name == "" || !g.extensionalRels()[name] {
		return false
	}
	if !allArgsVariable(app.Terms) {
		return false
	}
	if !goivy.IsFalse(a.RHS) {
		return false
	}
	fs, ok := app.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	if st.Kind != cppStorageHashThunk {
		return false
	}
	w.linef("%s.memo.clear();", varName(name))
	return true
}
