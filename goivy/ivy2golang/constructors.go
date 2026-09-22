package ivy2golang

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy"
)

// constructorActionFor builds the synthetic action used to emit a sort
// constructor method. This mirrors ivy2cpp's constructorActionFor, with the
// generated action routed through the Go method/action emitter.
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
