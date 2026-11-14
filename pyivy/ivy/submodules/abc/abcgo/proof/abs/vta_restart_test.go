package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestVtaRollbackRestartFindsCounterexample(t *testing.T) {
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
	pars := &Abs_Par_t{NFramesMax: 4, NFramesStart: 1, NFramesPast: 1, FUseRollback: 1}
	if got := Gia_VtaPerform(p, pars); got != 0 {
		t.Fatalf("got status %d, want SAT", got)
	}
	if pars.NFramesMax != 4 {
		t.Fatal("frame limit not restored after restart")
	}
	cex := p.SequentialCounterexample()
	if cex == nil || cex.IFrame != 2 || gia.Gia_ManVerifyCex(p, cex, 0) == 0 {
		t.Fatalf("invalid three-frame witness: %+v", cex)
	}
}
