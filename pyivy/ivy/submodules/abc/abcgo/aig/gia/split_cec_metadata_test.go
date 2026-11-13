package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSplitWitnessKeepsCombinationalPIFormat(t *testing.T) {
	p := Gia_ManStart(8)
	a := Gia_ManAppendCi(p)
	r := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(0, 1, 1)
	c.PData[0] = 1
	if err := p.SetSplitCounterexample(c); err != nil {
		t.Fatal(err)
	}
	if p.CombinationalCounterexample() != c || c.NRegs != 0 || c.NPis != 1 {
		t.Fatal("witness layout/ownership changed")
	}
	p.SetSplitCofactors([]int32{1, 2})
	s := p.SplitCofactors()
	s[0] = 0
	if p.SplitCofactors()[0] != 1 {
		t.Fatal("cofactor metadata alias")
	}
}
