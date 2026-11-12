package gia

import "testing"

func TestFraigSweepSimplePreservesOutputs(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y^1)
	q := Gia_ManFraigSweepSimple(p, nil)
	if Gia_ObjReprSelf(p, x>>1) != Gia_ObjReprSelf(p, y>>1) {
		t.Fatal("merged AIG copies lost their equivalence")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1]) || out[1] == out[0] {
			t.Fatalf("input %v output %v error %v", in, out, err)
		}
	}
}
