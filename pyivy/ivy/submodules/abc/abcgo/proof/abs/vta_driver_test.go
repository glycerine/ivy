package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestVtaTrivialOutputs(t *testing.T) {
	for _, tc := range []struct{ lit, want int32 }{{0, 1}, {1, 0}} {
		p := gia.Gia_ManStart(4)
		reg := gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, tc.lit)
		gia.Gia_ManAppendCo(p, reg)
		gia.Gia_ManSetRegNum(p, 1)
		got := Gia_VtaPerform(p, &Abs_Par_t{NFramesMax: 1})
		if got != tc.want {
			t.Fatalf("output literal %d: got %d, want %d", tc.lit, got, tc.want)
		}
		if tc.want == 0 && (p.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(p, p.SequentialCounterexample(), 0) == 0) {
			t.Fatal("missing or invalid counterexample")
		}
	}
}
