package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestClearSequentialCounterexamplesKeepsCopiedWitnesses(t *testing.T) {
	p := Gia_ManStart(8)
	in, state := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, state)
	Gia_ManAppendCo(p, in)
	Gia_ManSetRegNum(p, 1)
	witness := util.Abc_CexAlloc(1, 1, 1)
	witness.PData[0] = 2
	if err := p.SetSequentialCounterexample(witness); err != nil {
		t.Fatal(err)
	}
	if err := p.SetSequentialCounterexamples([]*util.Abc_Cex_t{witness}); err != nil {
		t.Fatal(err)
	}
	copied := p.SequentialCounterexamples()
	p.ClearSequentialCounterexamples()
	if p.SequentialCounterexamples() != nil {
		t.Fatal("output-indexed witnesses were not cleared")
	}
	if len(copied) != 1 || copied[0].PData[0] != 2 {
		t.Fatal("clearing invalidated copied witness")
	}
	copied[0].PData[0] = 0
	if p.SequentialCounterexample() == nil || p.SequentialCounterexample().PData[0] != 2 {
		t.Fatal("clearing or mutating list affected the single witness")
	}
	p.ClearSequentialCounterexamples()
	var nilManager *Gia_Man_t
	nilManager.ClearSequentialCounterexamples()
}
