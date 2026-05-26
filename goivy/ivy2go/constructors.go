package ivy2go

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy"
)

// constructors.go is the literal port of ivy2cpp/constructors.go.
// Each Ivy sort constructor declared with `type T = struct { … }`
// turns into a Go method on *State that fills the destructor fields
// of a fresh record and returns it. The synthetic Sequence built
// here is fed through `emitSomeAction` (the same path ordinary
// actions use), so it picks up the standard signature lowering,
// formal-param + return handling, and trace prologues.
//
// emitConstructorDecls (cpp/constructors.go:60) is retained as a
// no-op — Go does not require forward declarations.

// constructorActionFor mirrors ivy2cpp/constructors.go:24
// constructorActionFor.
func (g *Generator) constructorActionFor(cons *goivy.Const, destrs []*goivy.Const) (string, goivy.Action, error) {
	if cons == nil {
		return "", nil, fmt.Errorf("constructor: nil symbol")
	}
	fnSort, ok := cons.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		return "", nil, fmt.Errorf("constructor %s: expected function sort, got %T", cons.Name, cons.CSort)
	}
	dom := fnSort.Domain()
	rng := fnSort.Range()
	if len(dom) != len(destrs) {
		return "", nil, fmt.Errorf("constructor %s: %d args but %d destructors", cons.Name, len(dom), len(destrs))
	}

	retval := goivy.NewConst("ret:val", rng)

	skolems := make([]*goivy.Const, len(dom))
	for i, dSort := range dom {
		skolems[i] = goivy.NewConst("fml:X"+strconv.Itoa(i), dSort)
	}

	asgns := make([]goivy.Expr, 0, len(destrs))
	for i, d := range destrs {
		lhs, err := goivy.NewApply(d, retval)
		if err != nil {
			return "", nil, fmt.Errorf("constructor %s: build d(retval) for %s: %w", cons.Name, d.Name, err)
		}
		asgns = append(asgns, goivy.NewAssignAction(lhs, skolems[i]))
	}

	seq := goivy.NewSequence(asgns...)
	seq.SetFormalParams(skolems)
	seq.SetFormalReturns([]*goivy.Const{retval})
	return cons.Name, seq, nil
}

// emitConstructorDecls is the Go counterpart of cpp's same-named
// helper. Go does not require forward declarations, so this is a
// no-op kept for parallel call-site shape with ivy2cpp.
func (g *Generator) emitConstructorDecls(w *goWriter) { _ = w }

// emitConstructors mirrors ivy2cpp/constructors.go:82 emitConstructors.
// Called from emitActions so the resulting methods land in the same
// `actions.go` stream as ordinary action methods.
func (g *Generator) emitConstructors(w *goWriter) {
	if g == nil || g.Mod == nil || g.Mod.SortConstructors == nil {
		return
	}
	for sortName, conss := range g.Mod.SortConstructors.All() {
		destrs, _ := g.Mod.SortDestructors.Get2(sortName)
		for _, cons := range conss {
			name, act, err := g.constructorActionFor(cons, destrs)
			if err != nil {
				g.errs = append(g.errs, err)
				continue
			}
			g.emitSomeAction(w, name, act)
		}
	}
}

// emitNativeTypeDecl is the M2 stub for native-typed sort declarations
// (e.g., `<<< go ... type Foo bytes32 ... >>>`). Real handling lands
// in M10.
func (g *Generator) emitNativeTypeDecl(w *goWriter, name, nativeType string) {
	w.linef("// native type %s = %s (deferred to M10)", goExportedName(name), nativeType)
}
