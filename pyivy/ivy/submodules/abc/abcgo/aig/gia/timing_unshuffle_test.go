package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestUnshuffleBoxedRegisterInputs(t *testing.T) {
	p := Gia_ManStart(10)
	pi, box, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, pi)
	Gia_ManAppendCo(p, box)
	Gia_ManAppendCo(p, reg)
	Gia_ManSetRegNum(p, 1)
	timing := tim.Tim_ManStart(3, 3)
	tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
	p.PManTime = timing
	q := Gia_ManDupUnshuffleInputs(p)
	if Gia_ManCiNum(q) != 3 || Gia_ManRegNum(q) != 1 {
		t.Fatal("incorrect interface")
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != in[0] || out[1] != in[2] || out[2] != in[1] {
			t.Fatalf("input=%v output=%v err=%v", in, out, err)
		}
	}
}
