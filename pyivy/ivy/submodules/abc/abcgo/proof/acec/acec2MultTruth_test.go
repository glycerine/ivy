package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCutTruthAndOrXorComplements(t *testing.T) {
	for xor := int32(0); xor < 2; xor++ {
		for ca := int32(0); ca < 2; ca++ {
			for cb := int32(0); cb < 2; cb++ {
				p := &Sdb_Sto_t{nCutSize: 6, vTtMem: vec.Vec_MemAllocForTT(6, 0)}
				a, b := testCut(1), testCut(2)
				a.iFunc = 2
				b.iFunc = 2
				out := testCut(3)
				Sdb_CutComputeTruth(p, a, b, ca, cb, out, xor)
				got := Sdb_CutTruth(p, out)
				if out.iFunc&1 != 0 {
					got = ^got
				}
				for x := 0; x < 64; x++ {
					aa := ((x & 1) != 0) != (ca != 0)
					bb := ((x & 2) != 0) != (cb != 0)
					want := aa && bb
					if xor != 0 {
						want = aa != bb
					}
					if ((got>>x)&1 != 0) != want {
						t.Fatalf("xor=%d complements=%d,%d input=%d truth=%x", xor, ca, cb, x, got)
					}
				}
			}
		}
	}
}
func TestCutTruthMinimizesConstant(t *testing.T) {
	p := &Sdb_Sto_t{nCutSize: 6, fTruthMin: 1, vTtMem: vec.Vec_MemAllocForTT(6, 0)}
	a := testCut(1)
	a.iFunc = 2
	out := testCut(1)
	if Sdb_CutComputeTruth(p, a, a, 0, 0, out, 1) != 1 || out.nLeaves != 0 || out.iFunc != 0 {
		t.Fatalf("constant XOR result %+v", out)
	}
}
