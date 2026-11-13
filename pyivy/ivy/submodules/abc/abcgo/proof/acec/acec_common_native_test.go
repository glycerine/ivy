package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCommonGraphPreservesNativeMuxAndBuffer(t *testing.T) {
	for phase := int32(0); phase < 4; phase++ {
		g := gia.Gia_ManStart(20)
		g.EnableMuxes()
		c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		mux := gia.Gia_ManAppendMuxReal(g, c^(phase&1), a, b^((phase>>1)&1))
		output := gia.Gia_ManAppendBuf(g, mux^1)
		gia.Gia_ManAppendCo(g, output)
		common := Acec_CommonStart(nil, g)
		mapped := gia.Gia_ObjValue(gia.Gia_ManObj(g, output>>1))
		gia.Gia_ManAppendCo(common, mapped)
		for mask := 0; mask < 8; mask++ {
			input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, e1 := g.Evaluate(input)
			got, e2 := common.Evaluate(input)
			if e1 != nil || e2 != nil || len(got) != 1 || got[0] != want[0] {
				t.Fatalf("phase=%d input=%v got=%v want=%v errors %v/%v", phase, input, got, want, e1, e2)
			}
		}
	}
}
