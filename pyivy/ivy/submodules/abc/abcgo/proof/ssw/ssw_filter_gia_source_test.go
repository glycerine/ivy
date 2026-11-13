package ssw

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSignalFilterGiaPreservesGraphAndCreatesRepresentatives(t *testing.T) {
	for _, preexisting := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gate := gia.Gia_ManAppendAnd(g, x, state)
		gia.Gia_ManAppendCo(g, gate)
		gia.Gia_ManAppendCo(g, state)
		gia.Gia_ManSetRegNum(g, 1)
		if preexisting {
			g.StartRepresentatives()
		}
		var before [4][]bool
		for i := 0; i < 4; i++ {
			before[i], _ = g.Evaluate([]bool{i&1 != 0, i&2 != 0})
		}
		Ssw_SignalFilterGia(g, 2, 100, 1, 0, 0, nil, 1, 0)
		if !g.HasRepresentatives() {
			t.Fatalf("preexisting=%t: no representatives", preexisting)
		}
		for i := 0; i < 4; i++ {
			after, err := g.Evaluate([]bool{i&1 != 0, i&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for k := range after {
				if after[k] != before[i][k] {
					t.Fatalf("preexisting=%t input=%d before=%v after=%v", preexisting, i, before[i], after)
				}
			}
		}
	}
}
