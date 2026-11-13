package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fraCore.c:468-482 sets sparse mode and forwards the prove flag through this
// source-named entry point; both modes must preserve the input function.
func TestSourceFraigEquivenceEntryPoint(t *testing.T) {
	for _, prove := range []int32{0, 1} {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		reduced := Fra_FraigEquivence(g, 1000, prove)
		if reduced == nil {
			t.Fatalf("prove=%d returned nil", prove)
		}
		for i := 0; i < 4; i++ {
			in := []bool{i&1 != 0, i&2 != 0}
			out, err := reduced.Evaluate(in)
			if err != nil || len(out) != 1 || out[0] != (in[0] && in[1]) {
				t.Fatalf("prove=%d input=%v output=%v error=%v", prove, in, out, err)
			}
		}
		if prove != 0 {
			model, ok := reduced.Data.([]int32)
			if !ok || len(model) < 2 {
				t.Fatalf("missing source prove-mode witness: %v", reduced.Data)
			}
			out, err := g.Evaluate([]bool{model[0] != 0, model[1] != 0})
			if err != nil || !out[0] {
				t.Fatalf("invalid witness %v: %v %v", model, out, err)
			}
		}
	}
}

// fraSim.c:440-446 flips only min(CIs, 32*words-1) lanes, preserving lane 0.
func TestSourceDistanceOneSaturatesLastLane(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 33; i++ {
		g.NewInput()
	}
	s := newFraSimulation(g, 0, 1, 1)
	s.AssignDistance1([]uint32{0, 0})
	for i, input := range g.Inputs() {
		want := uint32(0)
		if i < 31 {
			want = uint32(1) << uint(i+1)
		}
		if got := s.NodeWords(input.NodeID())[0]; got != want {
			t.Errorf("input %d = %08x, want %08x", i, got, want)
		}
	}
}

// Multi-frame source mode intentionally does not flip distance-one bits;
// frame PI patterns precede the initial latch pattern in pPat.
func TestSourceMultiFrameDistanceOneCopiesWithoutFlips(t *testing.T) {
	g := aig.NewGraph()
	a, b, state := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	s := newFraSimulation(g, 0, 2, 1)
	s.AssignDistance1([]uint32{0b10011})
	for _, tc := range []struct {
		edge  aig.Edge
		frame int
		want  uint32
	}{{a, 0, ^uint32(0)}, {b, 0, ^uint32(0)}, {a, 1, 0}, {b, 1, 0}, {state, 0, ^uint32(0)}} {
		if got := s.FrameWords(tc.edge.NodeID(), tc.frame)[0]; got != tc.want {
			t.Errorf("node %d frame %d = %08x, want %08x", tc.edge.NodeID(), tc.frame, got, tc.want)
		}
	}
}
