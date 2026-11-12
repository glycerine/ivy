package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

func TestGiaManBmcPerformAdapterStoresCounterexample(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	g := gia.Gia_ManFromAig(graph)
	params := &Bmc_AndPar_t{nFramesMax: 2}
	if result := Gia_ManBmcPerform(g, params); result != 0 {
		t.Fatalf("result=%d, want counterexample (0)", result)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
		t.Fatalf("counterexample=%+v, want frame 1 property 0", cex)
	}
}

func TestGiaManBmcPerformAdapterProvesZeroStateInvariant(t *testing.T) {
	graph := aig.NewGraph()
	_, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(state)
	aig.Aig_ManSetRegNum(graph, 1)
	g := gia.Gia_ManFromAig(graph)
	params := &Bmc_AndPar_t{nFramesMax: 2}
	if result := Gia_ManBmcPerform(g, params); result != 1 {
		t.Fatalf("result=%d, want bounded-safe (1)", result)
	}
	if cex := g.SequentialCounterexample(); cex != nil {
		t.Fatalf("unexpected counterexample: %+v", cex)
	}
}
