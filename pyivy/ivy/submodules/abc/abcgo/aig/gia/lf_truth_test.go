package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfTruthAndMux(t *testing.T) {
	for _, lut := range []int32{4, 7} {
		p := &Lf_Man_t{pPars: &Jf_Par_t{NLutSize: lut}, vTtMem: vec.Vec_MemAllocForTT(int(lut), 0)}
		var a, b, c Lf_Cut_t
		Lf_CutCreateUnit(&a, 1)
		Lf_CutCreateUnit(&b, 2)
		Lf_CutCreateUnit(&c, 3)
		for phase := int32(0); phase < 8; phase++ {
			out := Lf_Cut_t{nLeaves: 3}
			out.pLeaves[0], out.pLeaves[1], out.pLeaves[2] = 1, 2, 3
			Lf_CutComputeTruthMux(p, &a, &b, &c, phase&1, (phase>>1)&1, (phase>>2)&1, &out)
			truth := Lf_CutTruth(p, &out)[0]
			if out.iFunc&1 != 0 {
				truth = ^truth
			}
			for x := 0; x < 8; x++ {
				x0 := ((x & 1) != 0) != (phase&1 != 0)
				x1 := ((x & 2) != 0) != (phase&2 != 0)
				ctrl := ((x & 4) != 0) != (phase&4 != 0)
				want := x0
				if ctrl {
					want = x1
				}
				if (truth>>uint(x)&1 != 0) != want {
					t.Fatalf("lut=%d phase=%d x=%d", lut, phase, x)
				}
			}
		}
		out := Lf_Cut_t{nLeaves: 1}
		out.pLeaves[0] = 1
		if Lf_CutComputeTruth(p, &a, &a, 0, 1, &out, 0) != 1 || out.nLeaves != 0 || out.iFunc != 0 {
			t.Fatal("contradiction", out)
		}
	}
	if Lf_ManTtIsMux(0xcacacacacacacaca) != 1 || Lf_ManTtIsMux(0x8888888888888888) != 0 {
		t.Fatal("mux detection")
	}
}
