package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleValuesAndExtension(t *testing.T) {
	p := Gia_ManStart(20)
	input, state := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	node := Gia_ManAppendAnd(p, input, state)
	output := Gia_ManAppendCo(p, node)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame, c.IPo, c.PData[0] = 1, 0, 6
	Gia_ManCounterExampleValueStart(p, c)
	if Gia_ManCounterExampleValueLookup(p, node>>1, 0) != 0 || Gia_ManCounterExampleValueLookup(p, node>>1, 1) != 1 || Gia_ManCounterExampleValueLookup(p, output>>1, 1) != 1 {
		t.Fatal("wrong frame values")
	}
	Gia_ManCounterExampleValueStop(p)
	states := Gia_ManCexExtendToIncludeCurrentStates(p, c)
	if states.NPis != 2 || states.NBits != 4 || states.PData[0]&15 != 13 {
		t.Fatalf("incorrect states: %+v", states)
	}
	all := Gia_ManCexExtendToIncludeAllObjects(p, c)
	if all.NPis != Gia_ManObjNum(p) {
		t.Fatal("incorrect object width")
	}
	idx := int(all.NPis + (node >> 1))
	if all.PData[idx>>5]&(1<<uint(idx&31)) == 0 {
		t.Fatal("missing computed object value")
	}
}
