package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestVtaThreeFrameCounterexample(t *testing.T) {
	p := gia.Gia_ManStart(24)
	pi := gia.Gia_ManAppendCi(p)
	extra := gia.Gia_ManAppendCi(p)
	first := gia.Gia_ManAppendCi(p)
	second := gia.Gia_ManAppendCi(p)
	unused := extra
	for i := 0; i < 12; i++ {
		unused = gia.Gia_ManAppendAnd(p, unused, pi)
	}
	out := gia.Gia_ManAppendAnd(p, pi, second)
	gia.Gia_ManAppendCo(p, out)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManAppendCo(p, first)
	gia.Gia_ManSetRegNum(p, 2)
	if got := Gia_VtaPerform(p, &Abs_Par_t{NFramesMax: 4, NFramesPast: 1}); got != 0 {
		t.Fatalf("got status %d, want SAT", got)
	}
	cex := p.SequentialCounterexample()
	if cex == nil || cex.IFrame != 2 || gia.Gia_ManVerifyCex(p, cex, 0) == 0 {
		t.Fatalf("invalid three-frame witness: %+v", cex)
	}
}
