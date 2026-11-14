package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// test_vectors/ssw_group_witness.c reports result=0, original_regs=1,
// cex_regs=11, pis=1, frame=1 against the compiled ABC implementation.
func TestGroupedRarityGiaRetainsSourceWitnessDimensions(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, q := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, q))
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManSetRegNum(g, 1)
	var pars Ssw_RarPars_t
	Ssw_RarSetDefaultParams(&pars)
	pars.NFrames, pars.NWords, pars.NBinSize, pars.NRounds = 2, 1, 1, 1
	pars.FUseFfGrouping, pars.fSilent = 1, 1
	if got := Ssw_RarSimulateGia(g, &pars); got != 0 {
		t.Fatalf("result = %d, want source result 0", got)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.NRegs != 11 || cex.NPis != 1 || cex.IFrame != 1 {
		t.Fatalf("grouped source witness = %+v, want 11 registers, 1 PI, frame 1", cex)
	}
}
