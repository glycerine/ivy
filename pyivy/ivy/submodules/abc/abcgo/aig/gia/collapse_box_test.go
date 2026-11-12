package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestCollapseWhiteAndBlackBoxes(t *testing.T) {
	for _, black := range []int{0, 1} {
		p := Gia_ManStart(10)
		a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		Gia_ManAppendCo(p, a)
		Gia_ManAppendCo(p, box)
		tm := tim.Tim_ManStart(2, 2)
		tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, -1, black)
		p.pManTime = tm
		extra := Gia_ManStart(5)
		x := Gia_ManAppendCi(extra)
		Gia_ManAppendCo(extra, x^1)
		q := Gia_ManDupCollapse(p, extra, nil, 0)
		if black == 0 {
			for _, v := range []bool{false, true} {
				out, err := q.Evaluate([]bool{v})
				if err != nil || out[0] == v {
					t.Fatalf("white box: %v %v", out, err)
				}
			}
		} else {
			for mask := 0; mask < 4; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0}
				out, err := q.Evaluate(in)
				if err != nil || out[0] != in[0] || out[1] != in[1] {
					t.Fatalf("black box: %v %v", out, err)
				}
			}
		}
	}
}
