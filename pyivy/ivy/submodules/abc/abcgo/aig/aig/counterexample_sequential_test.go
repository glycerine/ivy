package aig

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSequentialCounterexampleOwnership(t *testing.T) {
	graph := NewGraph()
	input, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(input)
	Aig_ManSetRegNum(graph, 1)
	model := util.Abc_CexAlloc(1, 1, 3)
	model.IFrame = 2
	model.IPo = 0
	model.PData[0] = 5
	if err := graph.SetSequentialCounterexample(model); err != nil {
		t.Fatal(err)
	}
	saved := graph.SequentialCounterexample()
	model.PData[0] = 0
	if saved == model || saved.PData[0] != 5 {
		t.Fatal("sequential witness not copied")
	}
	for _, bad := range []*util.Abc_Cex_t{nil, {NRegs: 1, NPis: 1, NBits: 4, IFrame: 2, IPo: 0}, {NRegs: 1, NPis: 1, NBits: 1, IFrame: 2, IPo: 0, PData: []uint32{0}}, {NRegs: 1, NPis: 1, NBits: 4, IFrame: 2, IPo: 1, PData: []uint32{0}}} {
		if err := graph.SetSequentialCounterexample(bad); err == nil || graph.SequentialCounterexample() != saved {
			t.Fatal("invalid witness replaced stored model")
		}
	}
	if graph.TakeSequentialCounterexample() != saved || graph.SequentialCounterexample() != nil || saved.PData[0] != 5 {
		t.Fatal("take changed detached witness")
	}
}
