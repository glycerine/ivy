package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestEmptyBarrierArenaRetainsSourceMonitorRegisters(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	var index int32
	out := createNewAigWith0LivePoWithDSC(g, []aig.Edge{state}, &index, 0, 1, 2, nil)
	if got := out.RegisterCount(); got != 3 {
		t.Fatalf("source creates arena violation and pending registers even with no barriers: got %d", got)
	}
	if index != 4 {
		t.Fatalf("property index %d", index)
	}
}
