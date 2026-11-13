package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestSourceCounterexampleTransferPreservesOwnershipAndDimensions(t *testing.T) {
	g := Gia_ManStart(8)
	x, q := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, q)
	Gia_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(11, 1, 2)
	cex.IFrame = 1
	if g.SetSequentialCounterexample(cex) == nil {
		t.Fatal("validated setter accepted a mismatched register count")
	}
	g.TransferSequentialCounterexampleSource(cex)
	if g.SequentialCounterexample() != cex || g.SequentialCounterexample().NRegs != 11 {
		t.Fatal("source transfer did not retain the original witness")
	}
	g.TransferSequentialCounterexampleSource(nil)
	if g.SequentialCounterexample() != nil {
		t.Fatal("nil source transfer did not clear the witness")
	}
}
