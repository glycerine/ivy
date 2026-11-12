package gia

import "testing"

func TestDupConesPreservesSequentialClosure(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, r)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, b)
	Gia_ManSetRegNum(p, 1)
	for _, trim := range []int32{0, 1} {
		out := Gia_ManDupCones(p, []int32{0}, 1, trim)
		if Gia_ManPiNum(out) != 2 || Gia_ManRegNum(out) != 1 {
			t.Fatal("cone sequential closure")
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			got, e := out.Evaluate(in)
			if e != nil || got[0] != (in[0] && in[2]) || got[1] != in[1] {
				t.Fatalf("cone %v %v", got, e)
			}
		}
	}
}
func TestDupAndConesSelectedInputs(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, c^1)
	Gia_ManAppendCo(p, x)
	_ = b
	for _, trim := range []int32{0, 1} {
		out := Gia_ManDupAndCones(p, []int32{x >> 1}, 1, trim)
		if Gia_ManPiNum(out) != 3-trim {
			t.Fatal("PI count")
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			mapped := in
			if trim != 0 {
				mapped = []bool{in[0], in[2]}
			}
			got, e := out.Evaluate(mapped)
			if e != nil || got[0] != (in[0] && !in[2]) {
				t.Fatalf("AND cone %v %v", got, e)
			}
		}
	}
}
