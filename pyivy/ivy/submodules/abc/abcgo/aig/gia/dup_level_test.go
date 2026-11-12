package gia

import "testing"

func TestDupLevelAndExplicitVectors(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	o := Gia_ManAppendCo(p, y)
	one := Gia_ManDupLevelized(p)
	two := Gia_ManDupFromVecs(p, permutation(a>>1, b>>1, c>>1), permutation(x>>1, y>>1), permutation(o>>1), 0)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		for _, out := range []*Gia_Man_t{one, two} {
			got, e := out.Evaluate(in)
			if e != nil || got[0] != (in[0] && in[1] && in[2]) {
				t.Fatalf("copy %v %v", got, e)
			}
		}
	}
}
func TestDupSliced(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	out := Gia_ManDupSliced(p, 2)
	if Gia_ManPoNum(out) != 1 {
		t.Fatal("incorrect sliced output count")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[1]) {
			t.Fatalf("slice %v %v", got, e)
		}
	}
}
