package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestStrSimWordsSequentialFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi.Not(), lo.Not())
	out := g.AddOutput(gate)
	li := g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	s := newStrSimWords(g)
	s.words[pi.NodeID()][0] = 0x12345678
	s.words[pi.NodeID()][1] = 0xf0f0f0f0
	s.frame(0, nil)
	if got := s.words[out.NodeID()][0]; got != ^uint32(0x12345678) {
		t.Fatalf("first output %x", got)
	}
	if got := s.words[lo.NodeID()][1]; got != s.words[li.NodeID()][0] {
		t.Fatalf("latch transfer %x", got)
	}
	s.frame(1, nil)
	if got, want := s.words[out.NodeID()][1], ^uint32(0xf0f0f0f0)&uint32(0x12345678); got != want {
		t.Fatalf("second output %x want %x", got, want)
	}
}

func TestStrSimWordsPreservesMatchedGate(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	out := g.AddOutput(gate.Not())
	s := newStrSimWords(g)
	s.words[gate.NodeID()][0] = 0x12345678
	matched := make([]bool, len(g.Nodes()))
	matched[gate.NodeID()] = true
	s.frame(0, matched)
	if got := s.words[out.NodeID()][0]; got != ^uint32(0x12345678) {
		t.Fatalf("output %x", got)
	}
}
