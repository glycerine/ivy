package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestSweepWithBoxesCombinational(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(3, 2)
	tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
	p.pManTime = timing
	extra := Gia_ManStart(5)
	e := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e^1)
	p.pAigExtra = extra
	params := dch.DefaultParams()
	q := Gia_ManSweepWithBoxes(p, &params, nil, 0, 0, 0, 0)
	if q == nil || Gia_ManBoxNum(q) != 0 || Gia_ManCiNum(q) != 2 || Gia_ManCoNum(q) != 1 {
		t.Fatal("unused box was not removed")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && !in[1]) {
			t.Fatalf("input=%v output=%v err=%v", in, out, err)
		}
	}
}
