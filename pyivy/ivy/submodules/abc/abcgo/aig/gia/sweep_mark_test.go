package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestSweepMarkBoxCones(t *testing.T) {
	p := Gia_ManStart(16)
	a, box0, box1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	input0 := Gia_ManAppendCo(p, a)
	input1 := Gia_ManAppendCo(p, a)
	output := Gia_ManAppendCo(p, box0)
	timing := tim.Tim_ManStart(3, 3)
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, -1, 0)
	tim.Tim_ManCreateBox(timing, 1, 1, 2, 1, -1, 0)
	p.pManTime = timing
	Gia_ManMarkSeqGiaWithBoxes(p, 0)
	for _, lit := range []int32{a, box0, input0, output} {
		if Gia_ObjIsTravIdCurrent(p, Gia_ManObj(p, lit>>1)) == 0 {
			t.Fatalf("missing mark on %d", lit)
		}
	}
	for _, lit := range []int32{box1, input1} {
		if Gia_ObjIsTravIdCurrent(p, Gia_ManObj(p, lit>>1)) != 0 {
			t.Fatalf("unexpected mark on %d", lit)
		}
	}
}
