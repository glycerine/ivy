package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestFraigCreateGiaRetainsUsedBox(t *testing.T) {
	p := Gia_ManStart(12)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, -1, 0)
	p.pManTime = timing
	extra := Gia_ManStart(4)
	e := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e)
	p.pAigExtra = extra
	u := Gia_ManDupUnnormalize(p)
	u.pManTime = p.pManTime
	u.pAigExtra = extra
	q := Gia_ManFraigCreateGia(u)
	if Gia_ManBoxNum(q) != 1 || Gia_ManCiNum(q) != 2 || Gia_ManCoNum(q) != 2 || Gia_ManCoNum(q.pAigExtra) != 1 {
		t.Fatal("incorrect retained box interface")
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(inputs)
		if err != nil || out[0] != inputs[0] || out[1] != (inputs[0] && inputs[1]) {
			t.Fatalf("mask=%d out=%v err=%v", mask, out, err)
		}
	}
}
