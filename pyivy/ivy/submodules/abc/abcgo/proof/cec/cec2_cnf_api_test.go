package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestCec2SourceCNFEntryPoints(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x^1, z)
		root := gia.Gia_ManAppendAnd(g, a^1, b^1)
		if mux {
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, root>>1), 1)
		}
		gia.Gia_ManCreateRefs(g)
		s := satoko.New()
		frontier, fanins := vec.Vec_PtrAlloc(4), vec.Vec_PtrAlloc(4)
		rv := Gia_ObjGetCnfVar(g, root>>1, frontier, fanins, s)
		vars := []int32{Gia_ObjGetCnfVar(g, x>>1, frontier, fanins, s), Gia_ObjGetCnfVar(g, y>>1, frontier, fanins, s), Gia_ObjGetCnfVar(g, z>>1, frontier, fanins, s), rv}
		for mask := 0; mask < 16; mask++ {
			assumptions := make([]int, 4)
			for i, v := range vars {
				assumptions[i] = int(v) * 2
				if mask&(1<<i) == 0 {
					assumptions[i]++
				}
			}
			want := !((mask&1 != 0 && mask&2 != 0) || (mask&1 == 0 && mask&4 != 0))
			got := s.SolveAssumptions(assumptions) == satoko.SATOKO_SAT
			if got != (want == (mask&8 != 0)) {
				t.Fatalf("mux=%v mask=%04b", mux, mask)
			}
		}
		extra := gia.Gia_ManAppendCi(g)
		ev := Gia_ObjGetCnfVar(g, extra>>1, frontier, fanins, s)
		if ev == rv || Cec2_ObjSatId(g, gia.Gia_ManObj(g, root>>1)) != rv {
			t.Fatal("growing graph corrupted SAT IDs")
		}
	}
}
