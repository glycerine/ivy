package gia

import "testing"

func TestConstraintExtraction(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c^1)
	Gia_ManAppendCo(p, y)
	out := Gia_ManDupWithConstr(p)
	if out == nil || out.nConstrs != 1 || Gia_ManPoNum(out) != 2 {
		t.Fatal("constraint interface")
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		original, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(in)
		if e != nil || (got[0] && !got[1]) != original[0] {
			t.Fatalf("constraint %v %v %v", in, got, e)
		}
	}
}
