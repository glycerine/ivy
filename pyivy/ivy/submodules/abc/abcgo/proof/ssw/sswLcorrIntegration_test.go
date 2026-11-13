package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func sswTrace(t *testing.T, g *aig.Graph, pattern []bool) []bool {
	t.Helper()
	regs := int(g.RegisterCount())
	state := make([]bool, regs)
	trace := make([]bool, 0, len(pattern))
	for _, pi := range pattern {
		inputs := append([]bool{pi}, state...)
		outputs, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		trace = append(trace, outputs[0])
		copy(state, outputs[len(outputs)-regs:])
	}
	return trace
}
func TestLatchCorrespondenceDriverPreservesTraces(t *testing.T) {
	g := aig.NewGraph()
	pi, l0, l1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(l0, l1))
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 2)
	pars := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParamsLcorr(pars)
	out := Ssw_LatchCorrespondence(g, pars)
	if out == nil {
		t.Fatal("no correspondence result")
	}
	for mask := 0; mask < 16; mask++ {
		pattern := make([]bool, 4)
		for i := range pattern {
			pattern[i] = mask&(1<<i) != 0
		}
		a, b := sswTrace(t, g, pattern), sswTrace(t, out, pattern)
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("pattern %v: %v != %v", pattern, a, b)
			}
		}
	}
}

func TestLatchCorrespondenceRetainsSourceFanoutLifecycle(t *testing.T) {
	g := aig.NewGraph()
	pi, l0, l1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(l0, l1))
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 2)
	pars := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParamsLcorr(pars)
	if Ssw_LatchCorrespondence(g, pars) == nil {
		t.Fatal("no correspondence result")
	}
	defer aig.Aig_ManFanoutStop(g)
	defer func() {
		if recover() == nil {
			t.Fatal("ABC leaves fanout active; a second start should fail")
		}
	}()
	Ssw_LatchCorrespondence(g, pars)
}
