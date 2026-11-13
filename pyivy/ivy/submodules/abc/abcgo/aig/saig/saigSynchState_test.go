package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestSynchTernaryLogic(t *testing.T) {
	values := []uint32{0, 1, 3}
	for _, a := range values {
		wantNot := map[uint32]uint32{0: 1, 1: 0, 3: 3}[a]
		if got := synchNot(synchValue(a)) & 3; got != wantNot {
			t.Fatalf("not %d = %d", a, got)
		}
		for _, b := range values {
			want := uint32(3)
			if a == 0 || b == 0 {
				want = 0
			} else if a == 1 && b == 1 {
				want = 1
			}
			if got := synchAnd(synchValue(a), synchValue(b)) & 3; got != want {
				t.Fatalf("%d and %d = %d want %d", a, b, got, want)
			}
		}
	}
}
func TestSynchFindAndReplayPattern(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.AND(pi, lo))
	aig.Aig_ManSetRegNum(g, 1)
	s := newSynchState(g, 2)
	s.initRegisters(true)
	// Every even candidate sets the input low and forces the register to zero.
	for i := range s.words[pi.NodeID()] {
		s.words[pi.NodeID()][i] = 0x44444444
	}
	s.simulate()
	pattern, count := s.bestPattern()
	if count != 0 {
		t.Fatalf("unknown=%d", count)
	}
	inputs, unknown := s.savePattern(pattern)
	if unknown != 0 || inputs[0] != 0 {
		t.Fatalf("inputs=%v unknown=%d", inputs, unknown)
	}
	s.initRegisters(true)
	s.givenInputs(inputs)
	s.simulate()
	s.transfer()
	if got := s.words[lo.NodeID()][0] & 3; got != 0 {
		t.Fatalf("state %d", got)
	}
}
