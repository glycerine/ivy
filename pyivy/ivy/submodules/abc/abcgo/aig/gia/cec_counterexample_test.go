package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCecCombinationalWitnessOnSequentialGraph(t *testing.T) {
	g := Gia_ManStart(8)
	x, ro := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, ro)
	Gia_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(0, 2, 1)
	cex.IPo = 1
	cex.PData[0] = 2
	if err := g.SetCecCombinationalWitness(cex); err != nil {
		t.Fatal(err)
	}
	cex.PData[0] = 0
	if got := g.SequentialCounterexample(); got == cex || got.PData[0] != 2 {
		t.Fatal("witness was not copied")
	}
	if g.CombinationalCounterexample() != nil {
		t.Fatal("unrelated witness slot changed")
	}
	invalid := util.Abc_CexAlloc(0, 1, 1)
	if err := g.SetCecCombinationalWitness(invalid); err == nil {
		t.Fatal("accepted wrong input count")
	}
	if g.SequentialCounterexample().PData[0] != 2 {
		t.Fatal("invalid witness changed stored value")
	}
}
