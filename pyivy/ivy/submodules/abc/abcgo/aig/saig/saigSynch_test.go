package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSynchSequenceAndApply(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	seq := Saig_SynchSequence(g, 2)
	if seq == nil || vec.Vec_StrSize(seq) != 1 {
		t.Fatalf("sequence %v", seq)
	}
	if n := Saig_SynchSequenceRun(g, nil, seq, 1); n != 0 {
		t.Fatalf("unknown registers %d", n)
	}
	q := Saig_ManDupInitZero(g)
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	origInit := aig.Aig_ObjIsMarkA(g, lo)
	for bits := 0; bits < 4; bits++ {
		piValue, state := bits&1 != 0, bits&2 != 0
		original, err := g.Evaluate([]bool{piValue, state != origInit})
		if err != nil {
			t.Fatal(err)
		}
		shifted, err := q.Evaluate([]bool{piValue, state})
		if err != nil {
			t.Fatal(err)
		}
		if original[0] != shifted[0] || original[1] != (shifted[1] != origInit) {
			t.Fatalf("original %v shifted %v mark=%v", original, shifted, origInit)
		}
	}
	if Saig_SynchSequenceApply(g, 2, 0) == nil {
		t.Fatal("synchronization failed")
	}
}
func TestSynchSequenceRejectsUnsynchronizableLatch(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	if seq := Saig_SynchSequence(g, 1); seq != nil {
		t.Fatalf("unexpected sequence %v", seq)
	}
}
func TestSynchronizeIdenticalDesigns(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	m := Saig_Synchronize(g, aig.Aig_ManDupSimple(g), 1, 0)
	if m == nil {
		t.Fatal("missing miter")
	}
	if err := m.Check(); err != nil {
		t.Fatal(err)
	}
}

// TestSynchSequenceDrawsAigManRandom: Saig_SynchSequence resets
// Aig_ManRandom(1) and draws its PI words from ABC's stream, so the draw
// after it does not depend on how many numbers were drawn before.
func TestSynchSequenceDrawsAigManRandom(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	nextAfter := func(before int) uint32 {
		aig.Aig_ManRandom(1)
		for i := 0; i < before; i++ {
			aig.Aig_ManRandom(0)
		}
		Saig_SynchSequence(g, 2)
		return aig.Aig_ManRandom(0)
	}
	if x, y := nextAfter(7), nextAfter(3); x != y {
		t.Fatalf("draw after Saig_SynchSequence depends on earlier draws (%#x vs %#x): no Aig_ManRandom(1) reset", x, y)
	}
	aig.Aig_ManRandom(1)
	if first := aig.Aig_ManRandom(0); nextAfter(0) == first {
		t.Fatal("Saig_SynchSequence did not draw from Aig_ManRandom")
	}
}
