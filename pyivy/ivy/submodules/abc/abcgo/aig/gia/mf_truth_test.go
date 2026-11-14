package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMfTruthOperations(t *testing.T) {
	for _, lut := range []int32{6, 8} {
		p := &Mf_Man_t{pPars: &Jf_Par_t{NLutSize: lut}, vTtMem: vec.Vec_MemAllocForTT(int(lut), 0)}
		a, b := &Mf_Cut_t{}, &Mf_Cut_t{}
		Mf_CutCreateUnit(a, 1)
		Mf_CutCreateUnit(b, 2)
		out := &Mf_Cut_t{}
		Mf_CutMergeOrder(a, b, out, lut)
		Mf_CutComputeTruth(p, a, b, 0, 0, out, 0)
		truth := vec.Vec_MemReadEntry(p.vTtMem, int(out.iFunc>>1))[0]
		if truth&15 != 8 || out.iFunc&1 != 0 {
			t.Fatalf("lut %d AND truth %x", lut, truth)
		}
		out = &Mf_Cut_t{}
		Mf_CutMergeOrder(a, a, out, lut)
		Mf_CutComputeTruth(p, a, a, 0, 0, out, 1)
		if out.nLeaves != 0 || out.iFunc != 0 {
			t.Fatalf("lut %d XOR reduction %+v", lut, out)
		}
		out = &Mf_Cut_t{}
		Mf_CutMergeOrder(a, b, out, lut)
		Mf_CutComputeTruthMux(p, a, b, a, 0, 0, 0, out)
		if out.nLeaves != 2 {
			t.Fatalf("lut %d mux support %d", lut, out.nLeaves)
		}
	}
}
