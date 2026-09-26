package ivy2cpp

import (
	"github.com/glycerine/ivy/goivy"
)

// Ptype represents a C++ parameter-passing policy.
// Mirrors Python ivy_to_cpp.py:394-412 which defines four classes with a
// common `make(t)` method. Go needs an interface (a Python-style duck-typed
// `make` method is not expressible directly) to dispatch over the four
// implementations. The four implementing types keep their Python names.
type Ptype interface {
	// Make wraps a C++ type string according to this passing policy.
	Make(t string) string
}

// ValueType passes the parameter by value.
// Python: ValueType (ivy_to_cpp.py:396-397).
type ValueType struct{}

// Make returns t unchanged.
func (ValueType) Make(t string) string { return t }

// ConstRefType passes the parameter by const reference.
// Python: ConstRefType (ivy_to_cpp.py:399-400).
type ConstRefType struct{}

// Make returns "const t&".
func (ConstRefType) Make(t string) string { return "const " + t + "&" }

// RefType passes the parameter by non-const reference. Used for outputs
// that double as inputs (Python: RefType, ivy_to_cpp.py:402-403).
type RefType struct{}

// Make returns "t&".
func (RefType) Make(t string) string { return t + "&" }

// ReturnRefType represents a return value lowered to an output parameter at
// the signature position Pos. Returns "void" from Make because it does not
// contribute a C++ return type.
// Python: ReturnRefType (ivy_to_cpp.py:406-411).
type ReturnRefType struct{ Pos int }

// Make returns "void".
func (ReturnRefType) Make(t string) string { return "void" }

// ptypeCacheEntry is the cached pair returned by getParamTypes for one
// action. Mirrors Python's attaching of param_types/return_types onto the
// action object; in Go the cache lives on *Generator (per goivy/CLAUDE.md
// section C) since *goivy.Action cannot be monkey-patched.
type ptypeCacheEntry struct {
	Params  []Ptype
	Returns []Ptype
}

// getParamTypes returns the cached parameter and return Ptype lists for
// action. Mirrors Python ivy_to_cpp.py:1514-1517.
func (g *Generator) getParamTypes(name string, act goivy.Action) ([]Ptype, []Ptype) {
	if g.ptypeCache == nil {
		g.ptypeCache = map[string]ptypeCacheEntry{}
	}
	if e, ok := g.ptypeCache[name]; ok {
		return e.Params, e.Returns
	}
	g.annotateAction(name, act)
	e := g.ptypeCache[name]
	return e.Params, e.Returns
}

// annotateAction computes the parameter and return passing policies for
// action and stores them in g.ptypeCache.
// Mirrors Python ivy_to_cpp.py:1479-1517 line for line.
func (g *Generator) annotateAction(name string, act goivy.Action) {
	if g.ptypeCache == nil {
		g.ptypeCache = map[string]ptypeCacheEntry{}
	}
	params := act.GetFormalParams()
	returns := act.GetFormalReturns()

	// Public actions: inputs are still passed by value at the external
	// boundary, but multiple returns use the normal C++ lowering: first
	// return by value, later returns as trailing output references.
	if g.Mod != nil && g.Mod.PublicActions != nil && g.Mod.PublicActions.Get(name) {
		paramTypes := make([]Ptype, len(params))
		for i := range paramTypes {
			paramTypes[i] = ValueType{}
		}
		returnTypes := make([]Ptype, len(returns))
		nextArgPos := len(params)
		for i := range returnTypes {
			if i == 0 {
				returnTypes[i] = ValueType{}
				continue
			}
			returnTypes[i] = ReturnRefType{Pos: nextArgPos}
			nextArgPos++
		}
		g.ptypeCache[name] = ptypeCacheEntry{Params: paramTypes, Returns: returnTypes}
		return
	}

	// Input parameter policy (Python lines 1493-1495).
	paramTypes := make([]Ptype, len(params))
	for i, p := range params {
		switch {
		case formalListContains(returns, p):
			paramTypes[i] = RefType{}
		case g.actionAssigns(act, p) || !g.isStructSort(p.CSort):
			paramTypes[i] = ValueType{}
		default:
			paramTypes[i] = ConstRefType{}
		}
	}

	// Return policy (Python lines 1496-1512).
	returnTypes := make([]Ptype, len(returns))
	nextArgPos := len(params)
	for pos, r := range returns {
		if idx := indexInFormals(params, r); idx >= 0 {
			returnTypes[pos] = ReturnRefType{Pos: idx}
		} else if pos > 0 {
			returnTypes[pos] = ReturnRefType{Pos: nextArgPos}
			nextArgPos++
		} else {
			returnTypes[pos] = ValueType{}
		}
	}

	g.ptypeCache[name] = ptypeCacheEntry{Params: paramTypes, Returns: returnTypes}
}

// actionAssigns reports whether any sub-action of act modifies p.
// Mirrors Python ivy_to_cpp.py:1486-1487: `any(p in sub.modifies() for
// sub in action.iter_subactions())`. Symbol equality is by Name in Go
// because *Const pointer identity is not stable across constructions.
//
// Destructor chains are walked directly via g.Mod.DestructorSorts rather
// than via goivy.ModifiesSingle because the latter requires an
// ActionsConfig.Context to be wired into the module, which is not always
// the case when ivy2cpp is invoked (e.g. tests use compileIvySource which
// does not run create_isolate). Walking here mirrors the Python destructor
// chain in lg.AssignAction.modifies().
func (g *Generator) actionAssigns(act goivy.Action, p *goivy.Const) bool {
	if act == nil || p == nil {
		return false
	}
	for _, sub := range act.IterSubactions() {
		switch a := sub.(type) {
		case *goivy.LogicAssignAction:
			if g.modifiedRootName(a.LHS) == p.Name {
				return true
			}
		case *goivy.LogicHavocAction:
			if g.modifiedRootName(a.Target) == p.Name {
				return true
			}
		case *goivy.LogicCallAction:
			for _, ret := range a.ActualReturns {
				if g.modifiedRootName(ret) == p.Name {
					return true
				}
			}
		}
	}
	return false
}

// modifiedRootName returns the root state-symbol name of an assignment
// LHS / havoc target / call return after stripping destructor applications.
// Mirrors the destructor-chain walk in Python AssignAction.modifies()
// (and the same in HavocAction / CallAction return positions).
func (g *Generator) modifiedRootName(e goivy.Expr) string {
	for {
		app, ok := e.(*goivy.Apply)
		if !ok {
			break
		}
		c, ok := app.Func.(*goivy.Const)
		if !ok || !g.isDestructorName(c.Name) || len(app.Terms) == 0 {
			// LHS is an Apply but its head is not a destructor; Python
			// returns the head symbol (n.rep) for this case.
			if ok {
				return c.Name
			}
			return goivy.ExprName(e)
		}
		e = app.Terms[0]
	}
	return goivy.ExprName(e)
}

// isStructSort reports whether s is an uninterpreted sort that is either a
// native type or a destructor (struct) sort.
// Mirrors Python ivy_to_cpp.py:1489-1491:
//
//	il.is_uninterpreted_sort(sort) and
//	   (sort.name in im.module.native_types or sort.name in im.module.sort_destructors)
func (g *Generator) isStructSort(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || s == nil {
		return false
	}
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return false
	}
	name := sortName(s)
	if name == "" {
		return false
	}
	if g.Mod.NativeTypes != nil {
		if _, ok := g.Mod.NativeTypes[name]; ok {
			return true
		}
	}
	if g.Mod.SortDestructors != nil {
		if _, ok := g.Mod.SortDestructors.Get2(name); ok {
			return true
		}
	}
	return false
}

// indexInFormals returns the position of target in formals (by Name) or -1.
func indexInFormals(formals []*goivy.Const, target *goivy.Const) int {
	if target == nil {
		return -1
	}
	for i, f := range formals {
		if f != nil && f.Name == target.Name {
			return i
		}
	}
	return -1
}

// mayAlias conservatively reports whether two expressions may refer to
// overlapping storage. Mirrors Python ivy_to_cpp.py:1525-1530:
//
//	def may_alias(x,y):
//	    def root_var(x):
//	        while il.is_app(x) and is_destructor(x.rep):
//	            x = x.args[0]
//	        return x
//	    return root_var(x) == root_var(y)
func (g *Generator) mayAlias(x, y goivy.Expr) bool {
	rx := g.rootVar(x)
	ry := g.rootVar(y)
	return exprRootEqByName(rx, ry)
}

// rootVar strips leading destructor applications from e, mirroring Python's
// inner closure in may_alias (ivy_to_cpp.py:1526-1529).
func (g *Generator) rootVar(e goivy.Expr) goivy.Expr {
	for {
		app, ok := e.(*goivy.Apply)
		if !ok {
			return e
		}
		c, ok := app.Func.(*goivy.Const)
		if !ok || !g.isDestructorName(c.Name) || len(app.Terms) == 0 {
			return e
		}
		e = app.Terms[0]
	}
}

// isDestructorName reports whether name is registered as a destructor.
// Mirrors Python is_destructor (ivy_to_cpp.py:1522-1523).
func (g *Generator) isDestructorName(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return false
	}
	_, ok := g.Mod.DestructorSorts[name]
	return ok
}

// exprRootEqByName compares two root expressions (post-rootVar stripping)
// for identity. Python uses Symbol.__eq__ which compares name + sort; Go
// uses Name only because parsing can produce distinct *Const pointers with
// the same name.
func exprRootEqByName(a, b goivy.Expr) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ac, aIsC := a.(*goivy.Const)
	bc, bIsC := b.(*goivy.Const)
	if aIsC && bIsC {
		return ac.Name == bc.Name
	}
	av, aIsV := a.(*goivy.LogicVariable)
	bv, bIsV := b.(*goivy.LogicVariable)
	if aIsV && bIsV {
		return av.Name == bv.Name
	}
	// Mixed Const vs Variable, or two Apply nodes that survived
	// destructor stripping (e.g. function-typed parameter applied to
	// args): fall back to the goivy ExprName comparison which covers
	// both shapes.
	return goivy.ExprName(a) == goivy.ExprName(b) && goivy.ExprName(a) != ""
}
