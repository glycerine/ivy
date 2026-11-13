package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSpecReduceNonpositiveFrameLimit(t *testing.T) {
	p := Gia_ManStart(10)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, r)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, r>>1, 0)
	init := util.Abc_CexAlloc(1, 0, 0)
	for _, limit := range []int32{-1, 0} {
		frames := int32(-1)
		q := Gia_ManSpecReduceInitFrames(p, init, limit, &frames, 0, 1)
		if q == nil || frames != 1 || Gia_ManPoNum(q) != 1 {
			t.Fatalf("limit %d frames %d graph %v", limit, frames, q)
		}
	}
}
