package gia

import "testing"

func TestFraigReduceRepresentatives(t *testing.T) {
	for _, inverted := range []bool{false, true} {
		p := Gia_ManStart(10)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		Gia_ManAppendCo(p, x)
		reps := make([]int32, Gia_ManObjNum(p))
		for i := range reps {
			reps[i] = -1
		}
		reps[b>>1] = a
		if inverted {
			reps[b>>1] ^= 1
		}
		q := Gia_ManFraigReduceGia(p, reps)
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			out, err := q.Evaluate(inputs)
			want := inputs[0] && !inverted
			if err != nil || out[0] != want {
				t.Fatalf("inverted=%v mask=%d out=%v err=%v", inverted, mask, out, err)
			}
		}
	}
}
