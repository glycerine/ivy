package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCombSpecReduceReportsCandidateMismatch(t *testing.T) {
	for _, rings := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c)
		gia.Gia_ManAppendCo(g, y)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		g.SetNext(x>>1, y>>1)
		var pairs *Vec_Int_t
		out := Cec_ManCombSpecReduce(g, &pairs, rings)
		if out == nil || vec.Vec_IntSize(pairs) == 0 || vec.Vec_IntSize(pairs)%2 != 0 {
			t.Fatalf("rings=%d: missing candidates", rings)
		}
		for mask := 0; mask < 8; mask++ {
			got, err := out.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			want := mask&1 != 0 && ((mask&2 != 0) != (mask&4 != 0))
			var any bool
			for _, value := range got {
				any = any || value
			}
			if any != want {
				t.Fatalf("rings=%d mask=%03b got=%v want=%v", rings, mask, got, want)
			}
		}
	}
}
