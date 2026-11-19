package goivy

import (
	"strings"
	"testing"
)

// Python DomainSetup.named (ivy_compiler.py):
//
//	vmap = dict((x.name,x) for x in lu.variables_ast(cond))
//	...
//	for x in vmap:
//	    if x not in used:
//	        raise IvyError(lhs,'{} must be a parameter of {}'.format(x,lhs.rep))
//
// vmap is in variables_ast (first-occurrence) order, so with body q(Z,Y,X)
// and no parameters the error names Z. goivy used a Go map (random order).
func TestDomainSetupNamedReportsFirstMissingParameterInOrder(t *testing.T) {
	for i := 0; i < 30; i++ {
		cfg := NewAstConfig()
		c := newTestCompiler()
		d := NewDomainSetup(c)
		node := &UninterpretedSort{Name: "node"}
		c.Sig.AddSort(node)
		qSort, err := NewFunctionSort(node, node, node, Boolean)
		if err != nil {
			t.Fatal(err)
		}
		q := NewConst("q", qSort)
		c.Sig.AddSymbol(q.Name, q.CSort)
		mk := func(n string) *LogicVariable {
			v, err := NewVariable(n, node)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		x, y, z := mk("X"), mk("Y"), mk("Z")
		d.LastFact = cfg.NewLabeledFormula(cfg.NewAtom("prop"), &LogicExists{
			Variables: []*LogicVariable{x},
			Body:      MustApply(q, z, y, x),
		})
		err = d.Named(cfg.NewAtom("wit"))
		if err == nil || !strings.Contains(err.Error(), "Z must be a parameter of wit") {
			t.Fatalf("run %d: got %v, want the first missing parameter Z", i, err)
		}
	}
}
