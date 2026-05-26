package ivy2go

import (
	"github.com/glycerine/ivy/goivy"
)

// extensional.go is the literal port of ivy2cpp/extensional.go. A
// relation is "extensional" if its only update mode at runtime is
// "set this single point" (no quantified-LHS updates) AND it's
// initialized to all-false in some initializer. For such relations we
// can use sparse hash-thunk storage and skip explicit Init writes.
//
// allStateSymbols (cpp/generator.go:1067 OMITTED in ivy2go) is the
// canonical universe walker; until that's ported, the analysis here
// uses stateSymbols() which already filters destructors / sort
// constructors / definitions.

// extensionalRels mirrors ivy2cpp/extensional.go:11. Cached lookup;
// first call populates via extensionalRelations.
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

// extensionalRelations mirrors ivy2cpp/extensional.go:36. Returns the
// set of state-symbol names that qualify as extensional:
//   - LHS is initialized to all-false in some initializer, AND
//   - no update has a quantified LHS or non-false RHS.
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
	// Pipeline fallback: same as ivy2cpp/extensional.go:63-71 when
	// the module came through SourceString(create_isolate=false) and
	// after-init bodies remained in Mod.Actions under mixin names.
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
	for _, sym := range g.stateSymbols() {
		if inited[sym.Name] && !bad[sym.Name] {
			res[sym.Name] = true
		}
	}
	return res
}

// markBadExtensional mirrors ivy2cpp/extensional.go:85.
func (g *Generator) markBadExtensional(sub goivy.Action, bad map[string]bool) {
	switch a := sub.(type) {
	case *goivy.LogicAssignAction:
		name := repName(a.LHS)
		if name == "" || g.isDestructorSort(name) {
			return
		}
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

// collectInitedExtensional mirrors ivy2cpp/extensional.go:123.
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

// isDestructorSort mirrors ivy2cpp/extensional.go:148.
func (g *Generator) isDestructorSort(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return false
	}
	_, ok := g.Mod.DestructorSorts[name]
	return ok
}

// repName mirrors ivy2cpp/extensional.go:160.
func repName(e goivy.Expr) string {
	switch n := e.(type) {
	case *goivy.Const:
		return n.Name
	case *goivy.Apply:
		return goivy.ExprName(n.Func)
	}
	return ""
}

// argsOf mirrors ivy2cpp/extensional.go:173.
func argsOf(e goivy.Expr) []goivy.Expr {
	if app, ok := e.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

// allArgsNonVariable mirrors ivy2cpp/extensional.go:182.
func allArgsNonVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); isVar {
			return false
		}
	}
	return true
}

// allArgsVariable mirrors ivy2cpp/extensional.go:193.
func allArgsVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); !isVar {
			return false
		}
	}
	return true
}

// emitExtensionalRelationClear mirrors ivy2cpp/extensional.go:213.
// The hash-thunk reset shortcut: assigning all-false to a quantified
// LHS of an extensional relation collapses to a single `.memo` clear
// (Go: `<map> = nil` to release the backing storage). Returns true
// when handled.
func (g *Generator) emitExtensionalRelationClear(w *goWriter, a *goivy.LogicAssignAction) bool {
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
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	if st.Kind != goStorageHashThunk {
		return false
	}
	w.linef("s.%s = nil", goExportedName(name))
	return true
}

// emitExtensionalDecls keeps its existing role as the per-stream
// placeholder. Extensional-relation handling is largely transparent
// in the Go runtime (Go maps are sparse by default).
func (g *Generator) emitExtensionalDecls(w *goWriter) { _ = w }
