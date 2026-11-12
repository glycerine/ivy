package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSequentialCounterexampleReplayAndOwnership(t *testing.T) {
	p := Gia_ManStart(8)
	input, state := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, state)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	cex := util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	cex.PData[0] = 2 // Initial zero, PI at frame zero is one.
	if err := p.ValidateSequentialCounterexample(cex, true); err != nil {
		t.Fatal(err)
	}
	if err := p.SetSequentialCounterexample(cex); err != nil {
		t.Fatal(err)
	}
	cex.PData[0] = 0
	if p.SequentialCounterexample().PData[0] != 2 {
		t.Fatal("stored witness aliases input")
	}
	if err := p.ValidateSequentialCounterexample(cex, true); err == nil {
		t.Fatal("accepted false output")
	}
	if p.SequentialCounterexample().PData[0] != 2 {
		t.Fatal("validation changed stored witness")
	}
	cex.IFrame = 0
	cex.NBits = 2
	cex.PData[0] = 1
	if err := p.ValidateSequentialCounterexample(cex, true); err == nil {
		t.Fatal("accepted nonzero reset")
	}
	if err := p.ValidateSequentialCounterexample(cex, false); err != nil {
		t.Fatal(err)
	}
	p.aigerLatchInits = []int32{-1}
	if err := p.ValidateSequentialCounterexample(cex, true); err != nil {
		t.Fatal(err)
	}
	p.aigerLatchInits = []int32{1}
	if err := p.ValidateSequentialCounterexample(cex, true); err != nil {
		t.Fatal(err)
	}
}
