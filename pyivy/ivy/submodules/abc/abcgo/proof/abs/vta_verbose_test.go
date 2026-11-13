package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestVtaVerboseReporting(t *testing.T) {
	p := gia.Gia_ManStart(20)
	pi := gia.Gia_ManAppendCi(p)
	extra := gia.Gia_ManAppendCi(p)
	reg := gia.Gia_ManAppendCi(p)
	unused := extra
	for i := 0; i < 10; i++ {
		unused = gia.Gia_ManAppendAnd(p, unused, pi)
	}
	out := gia.Gia_ManAppendAnd(p, pi, reg)
	gia.Gia_ManAppendCo(p, out)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 1)
	if got := Gia_VtaPerform(p, &Abs_Par_t{nFramesMax: 3, nFramesPast: 1, fVerbose: 1}); got != 0 {
		t.Fatalf("got status %d, want SAT", got)
	}
	if cex := p.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(p, cex, 0) == 0 {
		t.Fatal("missing or invalid witness")
	}
}
