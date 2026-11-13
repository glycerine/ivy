package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestVtaFindsInputCounterexample(t *testing.T) {
	p := gia.Gia_ManStart(10)
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
	if got := Gia_VtaPerform(p, &Abs_Par_t{nFramesMax: 3, nFramesPast: 1, nRatioMin: 0}); got != 0 {
		t.Fatalf("got status %d, want SAT", got)
	}
	if cex := p.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(p, cex, 0) == 0 {
		t.Fatal("missing or invalid witness")
	}
}

func TestVtaUnsatCoreUnderAssumption(t *testing.T) {
	s := bsat.NewLabeledSolver2()
	s.AddClause(1, bsat.Negative(1), bsat.Positive(2))
	s.AddClause(2, bsat.Negative(2))
	var status, conflicts int32
	core := Vta_ManUnsatCore(int32(bsat.Positive(1)), s, 0, 0, &status, &conflicts)
	if status != 1 || core == nil {
		t.Fatalf("status %d, core %v", status, core)
	}
	got := map[int32]bool{}
	for _, label := range vec.Vec_IntArray(core) {
		got[label] = true
	}
	if !got[1] || !got[2] {
		t.Fatalf("missing proof labels: %v", got)
	}
}
