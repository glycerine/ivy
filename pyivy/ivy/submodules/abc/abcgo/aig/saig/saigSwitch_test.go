package saig

import (
	"math"
	"math/bits"
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
	if vec.Vec_IntSize(values) != g.NodeSlots() {
		t.Fatalf("result length = %d, want %d", vec.Vec_IntSize(values), g.NodeSlots())
	}
	pi := math.Float32frombits(uint32(vec.Vec_IntEntry(values, x.NodeID())))
	po := math.Float32frombits(uint32(vec.Vec_IntEntry(values, g.OutputAt(int(0)).NodeID())))
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

// TestSwitchingMatchesAbcFloatFormula: Saig_ManComputeSwitching evaluates
// (float)2.0 * nOnes / nTotal * (nTotal - nOnes) / nTotal left to right in
// float, which rounds differently from 2p(1-p).
func TestSwitchingMatchesAbcFloatFormula(t *testing.T) {
	for words := int32(1); words <= 7; words++ {
		total := 32 * words
		for ones := int32(0); ones <= total; ones++ {
			want := float32(2.0) * float32(ones) / float32(total) * float32(total-ones) / float32(total)
			if got := Saig_ManComputeSwitching(ones, words); got != want {
				t.Fatalf("Saig_ManComputeSwitching(%d, %d) = %v, want %v", ones, words, got, want)
			}
		}
	}
}

// TestSwitchingSimulationDrawsAigManRandom: true PIs take Aig_ManRandom(0)
// after the caller's Aig_ManRandom(1), one word per PI per frame.
func TestSwitchingSimulationDrawsAigManRandom(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	got := Saig_ManComputeSwitchProb4s(g, 3, 0, 1)
	aig.Aig_ManRandom(1)
	ones := 0
	for f := 0; f < 3; f++ {
		ones += bits.OnesCount32(aig.Aig_ManRandom(0))
	}
	want := float32(ones) / float32(32*3)
	if v := math.Float32frombits(uint32(vec.Vec_IntEntry(got, x.NodeID()))); v != want {
		t.Fatalf("PI probability %v, want %v from Aig_ManRandom", v, want)
	}
}
