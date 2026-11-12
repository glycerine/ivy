package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestNfCutTruthAndSupport(t *testing.T) {
	s := &nfTruthStore{memory: vec.Vec_MemAllocForTT(6, 0), matchCount: 2}
	a, b := &Nf_Cut_t{}, &Nf_Cut_t{}
	Nf_CutCreateUnit(a, 1)
	Nf_CutCreateUnit(b, 2)
	out := &Nf_Cut_t{}
	Nf_CutMergeOrder(a, b, out, 6)
	Nf_CutComputeTruth6(s, a, b, 0, 0, out, 0)
	if truth := vec.Vec_MemReadEntry(s.memory, int(out.iFunc>>1))[0]; truth&15 != 8 || out.Useless != 1 {
		t.Fatalf("AND: %x %+v", truth, out)
	}
	Nf_CutMergeOrder(a, a, out, 6)
	if Nf_CutComputeTruth6(s, a, a, 0, 0, out, 1) != 1 || out.nLeaves != 0 || out.iFunc != 0 {
		t.Fatalf("XOR reduction: %+v", out)
	}
	Nf_CutMergeOrder(a, b, out, 6)
	Nf_CutComputeTruthMux6(s, a, b, a, 0, 0, 0, out)
	if out.nLeaves != 2 {
		t.Fatalf("MUX support: %+v", out)
	}
}
