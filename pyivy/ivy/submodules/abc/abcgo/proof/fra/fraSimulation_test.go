package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFraSimulationCombinationalTruthAndClasses(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	and := g.AND(x.Not(), y.Not())
	xor := g.NewXor(x, y)
	g.AddOutput(and)
	g.AddOutput(xor)
	s := newFraSimulation(g, 0, 1, 1)
	s.frameWords(x.NodeID(), 0)[0] = 0xaaaaaaaa
	s.frameWords(y.NodeID(), 0)[0] = 0xcccccccc
	s.simulateOne()
	for mask := 0; mask < 4; mask++ {
		out, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		for i, edge := range g.Outputs() {
			got := s.frameWords(edge.NodeID(), 0)[0]&(1<<mask) != 0
			if got != out[i] {
				t.Fatalf("assignment=%d output=%d got=%v want=%v", mask, i, got, out[i])
			}
		}
	}
	if s.frameWords(and.NodeID(), 0)[0] != 0xeeeeeeee {
		t.Fatalf("phase-normalized AND = %#x", s.frameWords(and.NodeID(), 0)[0])
	}
	if !s.Equal(x.NodeID(), x.NodeID()) || s.Equal(x.NodeID(), y.NodeID()) || s.IsConstant(x.NodeID()) || !s.IsConstant(0) {
		t.Fatal("wrong class predicates")
	}
	if s.DifferenceWeight(x.NodeID(), y.NodeID()) != 16 || s.CountOnes(x.NodeID()) != 16 {
		t.Fatal("wrong simulation statistics")
	}
}

func TestFraSimulationPrefixAndLatchTransfer(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	po := g.AddOutput(g.XOR(input, state))
	g.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(g, 1)
	s := newFraSimulation(g, 1, 1, 1)
	s.assignFrame(input.NodeID(), 0, true)
	s.assignFrame(input.NodeID(), 1, true)
	s.assignFrame(state.NodeID(), 0, false)
	s.simulateOne()
	if got := s.frameWords(po.NodeID(), 0)[0]; got != ^uint32(0) {
		t.Fatalf("first output %#x", got)
	}
	if got := s.frameWords(po.NodeID(), 1)[0]; got != 0 {
		t.Fatalf("second output %#x", got)
	}
	if got := s.frameWords(state.NodeID(), 1)[0]; got != ^uint32(0) {
		t.Fatalf("latch transfer %#x", got)
	}
	// Prefix words do not contribute to FRA equivalence class comparisons.
	s.frameWords(input.NodeID(), 0)[0] = 0
	if !s.Equal(input.NodeID(), state.NodeID()) || s.Hash(input.NodeID()) != s.Hash(state.NodeID()) {
		t.Fatal("prefix affected class predicates")
	}
}

func TestFraSimulationDistanceOneAcrossWords(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 33; i++ {
		g.NewInput()
	}
	s := newFraSimulation(g, 0, 1, 2)
	s.assignDistance1([]uint32{0, 0})
	for i, input := range g.Inputs() {
		words := s.nodeWords(input.NodeID())
		want := [2]uint32{}
		want[(i+1)/32] = 1 << uint((i+1)%32)
		if words[0] != want[0] || words[1] != want[1] {
			t.Fatalf("input %d got %x want %x", i, words, want)
		}
	}
}

func TestFraSimulationDrivesClassPartition(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	explicit := g.NewXor(x, y)
	structural := aig.Aig_Regular(g.XOR(x, y))
	s := newFraSimulation(g, 0, 1, 1)
	s.frameWords(x.NodeID(), 0)[0] = 0xaaaaaaaa
	s.frameWords(y.NodeID(), 0)[0] = 0xcccccccc
	s.simulateOne()
	classes := newClassPartition(len(g.Nodes()))
	classes.prepareGraph(g, false, 0, s.Hash, s.IsConstant, s.Equal)
	if classes.reprs[structural.NodeID()] != explicit.NodeID() {
		t.Fatalf("equivalent XOR forms not grouped: %v", classes.reprs)
	}
}

func TestFraSimulationCounterExample(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	s := newFraSimulation(g, 0, 2, 1)
	s.assignFrame(input.NodeID(), 0, true)
	s.simulateOne()
	cex := s.counterExample()
	if cex == nil || cex.IPo != 0 || cex.IFrame != 1 {
		t.Fatalf("unexpected counterexample: %+v", cex)
	}
	frames := aig.NewGraph()
	frames.NewInput()
	frames.NewInput()
	frames.AddOutput(frames.Const0())
	failed := frames.AddOutput(frames.Const0())
	copy := copyFrameCounterExample(g, frames, []int32{1, 0, int32(failed.NodeID())})
	if copy == nil || copy.IFrame != 1 || copy.IPo != 0 {
		t.Fatalf("unexpected copied counterexample: %+v", copy)
	}
	if copyFrameCounterExample(g, frames, []int32{0, 0, int32(failed.NodeID())}) != nil {
		t.Fatal("accepted invalid model")
	}
}
