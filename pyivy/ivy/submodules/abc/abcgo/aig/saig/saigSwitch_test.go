package saig

import (
	"math"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSwitchingProbabilityFormulas(t *testing.T) {
	if got := Saig_ManComputeProbOne(16, 1); got != .5 {
		t.Fatalf("probability = %v, want 0.5", got)
	}
	if got := Saig_ManComputeSwitching(16, 1); got != .5 {
		t.Fatalf("switching = %v, want 0.5", got)
	}
	if got := Saig_ManComputeProbOnePlus(16, 1, 1); got != .5 {
		t.Fatalf("complemented probability = %v, want 0.5", got)
	}
}

func TestSwitchingSimulationStoresFloatBits(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	values := Saig_ManComputeSwitchProb4s(g, 4, 0, 1)
	if vec.Vec_IntSize(values) != len(g.Nodes()) {
		t.Fatalf("result length = %d, want %d", vec.Vec_IntSize(values), len(g.Nodes()))
	}
	pi := math.Float32frombits(uint32(vec.Vec_IntEntry(values, x.NodeID())))
	po := math.Float32frombits(uint32(vec.Vec_IntEntry(values, g.Outputs()[0].NodeID())))
	if pi != po {
		t.Fatalf("PI probability %v differs from its output probability %v", pi, po)
	}
}

func TestCompactAIGInputOutputRoundTrip(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	compact := Aig_CManCreate(g)
	Aig_CManRestart(compact)
	if got := Aig_CManGetPo(compact); got != int32(x) {
		t.Fatalf("decoded PO edge = %d, want %d", got, x)
	}
}
