package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// ABC clears all pattern words and records the real phase of each mapped CI.
func TestSavePatternAigPhaseAcrossWords(t *testing.T) {
	original, frames := aig.NewGraph(), aig.NewGraph()
	nodes := make([]aig.Edge, 65)
	for i := range nodes {
		nodes[i] = original.NewInput()
	}
	high := frames.NewInput()
	aig.Aig_ObjSetPhase(frames, high, true)
	p := &Ssw_Man_t{pAig: original, pFrames: frames, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(original.Nodes())), pPatWords: []uint32{^uint32(0), ^uint32(0), ^uint32(0)}}
	for i, node := range nodes {
		mapped := high
		if i%3 == 0 {
			mapped = mapped.Not()
		}
		Ssw_ObjSetFrame(p, node, 0, mapped)
	}
	Ssw_SmlSavePatternAigPhase(p, 0)
	for i := 0; i < 96; i++ {
		got := p.pPatWords[i/32]&(1<<uint(i%32)) != 0
		want := i < 65 && i%3 != 0
		if got != want {
			t.Fatalf("bit %d got %t want %t", i, got, want)
		}
	}
}
