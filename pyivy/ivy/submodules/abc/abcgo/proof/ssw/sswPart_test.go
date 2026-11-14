package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestPartitionRejectsConstraintsWithoutChangingParameters(t *testing.T) {
	g := aig.NewGraph()
	p := &Ssw_Pars_t{FConstrs: 1, NPartSize: 3, FVerbose: 1}
	if Ssw_SignalCorrespondencePart(g, p) != nil || p.NPartSize != 3 || p.FVerbose != 1 {
		t.Fatal("constraint rejection changed parameters")
	}
}

func TestPartitionedCorrespondencePreservesTraces(t *testing.T) {
	g := aig.NewGraph()
	pi, l0, l1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(l0, l1))
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 2)
	pars := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParams(pars)
	pars.NPartSize = 1
	result := Ssw_SignalCorrespondence(g, pars)
	if result == nil {
		t.Fatal("partitioned correspondence returned nil")
	}
	if pars.NPartSize != 1 {
		t.Fatal("partition parameter not restored")
	}
	for mask := 0; mask < 16; mask++ {
		pattern := make([]bool, 4)
		for i := range pattern {
			pattern[i] = mask&(1<<i) != 0
		}
		a, b := sswTrace(t, g, pattern), sswTrace(t, result, pattern)
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("pattern %v: %v != %v", pattern, a, b)
			}
		}
	}
}
