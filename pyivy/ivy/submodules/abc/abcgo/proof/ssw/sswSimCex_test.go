package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestSimulationCounterexampleReplays(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlSimInfo(p, pi)[0] = 4
	Ssw_SmlSimulateOne(p)
	p.fNonConstOut = Ssw_SmlCheckNonConstOutputs(p)
	cex := Ssw_SmlGetCounterExample(p)
	if cex == nil || cex.IFrame != 1 || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatal("counterexample failed replay")
	}
}
func TestSimulationOutputPattern(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	out := g.AddOutput(a)
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 4
	model := Ssw_SmlCheckOutput(p)
	if len(model) != 2 || model[0] != 1 || model[1] != int32(out.NodeID()) {
		t.Fatalf("model %v", model)
	}
}

func TestSimulationOutputPatternSkipsPrefix(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	p := Ssw_SmlStart(g, 1, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 1
	Ssw_SmlSimInfo(p, a)[1] = 2
	Ssw_SmlSimInfo(p, b)[1] = 2
	model := Ssw_SmlCheckOutput(p)
	if len(model) != 3 || model[0] != 1 || model[1] != 1 {
		t.Fatalf("expected suffix pattern, got %v", model)
	}
}
