package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestRandomSimulationDelayedCounterexample(t *testing.T) {
	graph := aig.NewGraph()
	x, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(x)
	aig.Aig_ManSetRegNum(graph, 1)
	if got := Raig_ManSimulate(graph, 2, 10, 0, 1, 0); got != 1 {
		t.Fatalf("delayed bug status %d", got)
	}
	cex := graph.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
		t.Fatalf("delayed witness %+v", cex)
	}
	if err := validateRandomCounterexample(graph, cex); err != nil {
		t.Fatal(err)
	}
}
func TestRandomSimulationRecyclesLiveCut(t *testing.T) {
	graph := aig.NewGraph()
	x, y, state := graph.NewInput(), graph.NewInput(), graph.NewInput()
	value := state
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			value = graph.NewXor(value, x)
		} else {
			value = graph.NewXor(value, y)
		}
	}
	graph.AddOutput(graph.Const0())
	graph.AddOutput(value)
	aig.Aig_ManSetRegNum(graph, 1)
	p := Raig_ManCreate(graph)
	p.nWords = 3
	aig.Aig_ManRandom(1)
	for frame := 0; frame < 20; frame++ {
		first := int32(0)
		if frame == 0 {
			first = 1
		}
		if got := Raig_ManSimulateRound(p, 1, first, nil); got != 0 {
			t.Fatalf("constant-zero miter failed %d", got)
		}
	}
	if p.maxLive >= 20 || len(p.blocks) >= 20 {
		t.Fatalf("simulation retained whole graph: max=%d allocated=%d", p.maxLive, len(p.blocks))
	}
}
func TestRandomSimulationNonMiterOutputsAndUnusedInputs(t *testing.T) {
	graph := aig.NewGraph()
	x, unused, state := graph.NewInput(), graph.NewInput(), graph.NewInput()
	_ = unused
	graph.AddOutput(graph.NewXor(state, x))
	graph.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(graph, 1)
	if got := Raig_ManSimulate(graph, 1, 8, 0, 0, 0); got != 0 {
		t.Fatalf("non-miter detected failure %d", got)
	}
}
func TestRandomSimulationConstantTrueWitness(t *testing.T) {
	graph := aig.NewGraph()
	state := graph.NewInput()
	graph.AddOutput(graph.Const1())
	graph.AddOutput(state)
	aig.Aig_ManSetRegNum(graph, 1)
	if got := Raig_ManSimulate(graph, 1, 2, 0, 1, 0); got != 1 {
		t.Fatalf("constant-one result %d", got)
	}
	if err := validateRandomCounterexample(graph, graph.SequentialCounterexample()); err != nil {
		t.Fatal(err)
	}
}

func TestRandomSimulationClearsOldWitness(t *testing.T) {
	graph := aig.NewGraph()
	state := graph.NewInput()
	graph.AddOutput(graph.Const1())
	graph.AddOutput(state)
	aig.Aig_ManSetRegNum(graph, 1)
	if Raig_ManSimulate(graph, 1, 1, 0, 1, 0) != 1 {
		t.Fatal("initial witness missing")
	}
	if Raig_ManSimulate(graph, 1, 2, 0, 0, 0) != 0 || graph.SequentialCounterexample() != nil {
		t.Fatal("successful simulation retained stale witness")
	}
}
