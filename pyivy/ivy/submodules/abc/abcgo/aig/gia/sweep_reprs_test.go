package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestFraigSelectRepresentatives(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, _ := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	tm := tim.Tim_ManStart(3, 2)
	tim.Tim_ManCreateBox(tm, 0, 1, 2, 1, -1, 0)
	p.PManTime = tm
	c := Gia_ManStart(8)
	ca, cb := Gia_ManAppendCi(c), Gia_ManAppendCi(c)
	c.PReprs = make([]Gia_Rpr_t, c.nObjs)
	for i := range c.PReprs {
		c.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(c, cb>>1, ca>>1)
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	Gia_ManObj(p, a>>1).Value = uint32(ca)
	Gia_ManObj(p, b>>1).Value = uint32(cb)
	counts := make([]int32, 3)
	reprs := Gia_ManFraigSelectReprs(p, c, 0, counts)
	if reprs[b>>1] != a || reprs[a>>1] != -1 {
		t.Fatalf("representatives %v", reprs)
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		if Gia_ManCi(p, i).fMark0 != 0 {
			t.Fatal("mark leaked")
		}
	}
}
