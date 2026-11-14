package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestDupWithBoxesRemovesUnusedBox(t *testing.T) {
	p := Gia_ManStart(16)
	a, box0, _ := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, box0)
	timing := tim.Tim_ManStart(3, 3)
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, -1, 0)
	tim.Tim_ManCreateBox(timing, 1, 1, 2, 1, -1, 0)
	p.PManTime = timing
	extra := Gia_ManStart(8)
	e := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e)
	Gia_ManAppendCo(extra, e^1)
	p.PAigExtra = extra
	p.nAnd2Delay = 7
	q := Gia_ManDupWithBoxes(p, 0)
	if Gia_ManBoxNum(q) != 1 || Gia_ManCiNum(q) != 2 || Gia_ManCoNum(q) != 2 || Gia_ManCoNum(q.PAigExtra) != 1 || q.nAnd2Delay != 7 {
		t.Fatal("incorrect reduced box interface")
	}
	for _, v := range []bool{false, true} {
		out, err := q.PAigExtra.Evaluate([]bool{v})
		if err != nil || len(out) != 1 || out[0] != v {
			t.Fatalf("box logic: %v %v", out, err)
		}
	}
}
