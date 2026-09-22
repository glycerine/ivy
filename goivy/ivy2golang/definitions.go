package ivy2golang

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitNativeDefinitionMethods(w *goWriter) {
	for _, d := range g.nativeDefinitions() {
		name, act, err := g.derivedActionFor(d)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		g.emitNativeDefinitionMethod(w, name, act)
	}
}

func (g *Generator) derivedActionFor(d derivedDefinition) (string, goivy.Action, error) {
	retval := goivy.NewConst("ret:val", d.Sort)

	subs := map[goivy.NodeKey]goivy.Expr{}
	skolems := make([]*goivy.Const, 0, len(d.Params))
	for _, p := range d.Params {
		pname := goivy.ExprName(p)
		if pname == "" {
			return "", nil, fmt.Errorf("native definition %s: parameter has empty name (%T)", d.Name, p)
		}
		skol := goivy.NewConst("fml:"+pname, p.NodeSort())
		subs[goivy.Key(p)] = skol
		skolems = append(skolems, skol)
	}

	rhs := d.RHS
	if len(subs) > 0 {
		r, err := goivy.Substitute(d.RHS, subs)
		if err != nil {
			return "", nil, fmt.Errorf("native definition %s: substitute: %w", d.Name, err)
		}
		rhs = r
	}

	act := goivy.NewAssignAction(retval, rhs)
	act.SetFormalParams(skolems)
	act.SetFormalReturns([]*goivy.Const{retval})
	return d.Name, act, nil
}

func (g *Generator) emitNativeDefinitionMethod(w *goWriter, name string, act goivy.Action) {
	g.warnOnce("ivy2golang: native C++ expression definitions are emitted as Go zero-value stubs")
	sig, _ := g.methodSignature(name, act)
	w.open(sig + " {")
	w.line("// ivy native expression definition omitted: C++ native code is not translated to Go")
	returns := act.GetFormalReturns()
	switch len(returns) {
	case 0:
		w.line("return")
	case 1:
		w.linef("return %s", g.goZeroValue(returns[0].CSort))
	default:
		values := make([]string, len(returns))
		for i, r := range returns {
			values[i] = g.goZeroValue(r.CSort)
		}
		w.linef("return %s", strings.Join(values, ", "))
	}
	w.close("")
	w.blank()
	w.blank()
}
