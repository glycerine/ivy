package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleReplay(t *testing.T) {
	p := Gia_ManStart(16)
	input, state := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, state)
	Gia_ManAppendCo(p, input)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.PData[0] = 1 | 2 // initial state and first input set; second input clear
	c.IFrame, c.IPo = 1, 0
	if Gia_ManVerifyCex(p, c, 0) != 1 || Gia_ManFindFailedPoCex(p, c, 0) != 0 || Gia_ManFindFailedPoCex(p, c, 1) != -1 {
		t.Fatal("incorrect final-frame replay")
	}
	c.IPo = 0
	if Gia_ManVerifyCex(p, c, 1) != 1 {
		t.Fatal("dual output replay failed")
	}
	if Gia_ManSetFailedPoCex(p, c) != 1 || c.IFrame != 0 || c.NBits != 2 {
		t.Fatalf("first failure: %+v", c)
	}
}
