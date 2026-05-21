package ivy2cpp

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy"
)

// constructorActionFor builds the synthetic Sequence used to emit a
// sort constructor as a C++ method. Mirrors Python emit_constructor
// (ivy_to_cpp.py:1364-1375) up to the emit_some_action call:
//
//   - retval = Symbol("ret:val", cons.sort.rng)
//   - skolemize each constructor argument position i to a Constant
//     "fml:X_i" with the field sort.
//   - for each destructor d_i in im.module.sort_destructors[rng.name]:
//         AssignAction(d_i(retval), X_i)
//   - sequence those assignments; formal_params = skolems;
//     formal_returns = [retval].
//
// Returns the constructor method name and the synthetic action ready
// for emitSomeAction / emitMethodDeclLine.
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

func (g *Generator) emitConstructorDecls(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.SortConstructors == nil {
		return
	}
	emitted := 0
	for sortName, conss := range g.Mod.SortConstructors.All() {
		destrs, _ := g.Mod.SortDestructors.Get2(sortName)
		for _, cons := range conss {
			name, act, err := g.constructorActionFor(cons, destrs)
			if err != nil {
				g.errs = append(g.errs, err)
				continue
			}
			g.emitMethodDeclLine(w, name, act)
			emitted++
		}
	}
	if emitted > 0 {
		w.blank()
	}
}

func (g *Generator) emitConstructors(w *cppWriter) {
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
