package mpm

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutTruthOperationsAndSupport(t *testing.T) {
	for _, lut := range []int32{6, 7} {
		p := &Mpm_Man_t{nLutSize: lut, nTruWords: 1 << max(0, lut-6), pPars: &Mpm_Par_t{}, vTtMem: vec.Vec_MemAllocForTT(int(lut), 0)}
		a := &Mpm_Cut_t{nLeaves: 1, pLeaves: []int32{2}, iFunc: 2}
		b := &Mpm_Cut_t{nLeaves: 1, pLeaves: []int32{4}, iFunc: 2}
		c := &Mpm_Cut_t{nLeaves: 1, pLeaves: []int32{6}, iFunc: 2}
		for typ := int32(1); typ <= 3; typ++ {
			for mask := int32(0); mask < 8; mask++ {
				cut := &Mpm_Cut_t{nLeaves: 3, pLeaves: []int32{2, 4, 6}}
				Mpm_CutComputeTruth(p, cut, a, b, c, mask&1, mask>>1&1, mask>>2&1, typ)
				truth := Mpm_CutTruth(p, int32(cut.iFunc)>>1)
				for assignment := 0; assignment < 8; assignment++ {
					x := (assignment&1 != 0) != (mask&1 != 0)
					y := (assignment&2 != 0) != (mask&2 != 0)
					z := (assignment&4 != 0) != (mask&4 != 0)
					want := x && y
					if typ == 2 {
						want = x != y
					}
					if typ == 3 {
						want = z && y || !z && x
					}
					got := ((truth[assignment>>6]>>uint(assignment&63))&1 != 0) != (cut.iFunc&1 != 0)
					if got != want {
						t.Fatalf("lut=%d type=%d mask=%d assignment=%d got=%v want=%v", lut, typ, mask, assignment, got, want)
					}
				}
			}
		}
		p.pPars.fCutMin = 1
		cut := &Mpm_Cut_t{nLeaves: 1, pLeaves: []int32{2}}
		Mpm_CutComputeTruth(p, cut, a, a, nil, 0, 0, 0, 2)
		if cut.nLeaves != 0 || p.nSmallSupp != 1 {
			t.Fatalf("lut=%d support=%d small=%d", lut, cut.nLeaves, p.nSmallSupp)
		}
		words := make([]word, p.nTruWords)
		for i := range words {
			words[i] = 0xcccccccccccccccc
		}
		id := vec.Vec_MemHashInsert(p.vTtMem, words)
		cut = &Mpm_Cut_t{nLeaves: 3, pLeaves: []int32{2, 4, 6}, iFunc: uint32(id << 1)}
		if lut == 6 {
			Mpm_CutTruthMinimize6(p, cut)
		} else {
			Mpm_CutTruthMinimize7(p, cut)
		}
		if cut.nLeaves != 1 || cut.pLeaves[0] != 4 || Mpm_CutTruth(p, int32(cut.iFunc)>>1)[0] != 0xaaaaaaaaaaaaaaaa {
			t.Fatalf("lut=%d support compaction: cut=%+v", lut, cut)
		}
		vec.Vec_MemFree(p.vTtMem)
	}
}
