package gia

import "testing"

func TestCofactorDuplicateEquivalence(t *testing.T) {
	p := Gia_ManStart(30)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	x := Gia_ManHashAnd(p, a, b)
	y := Gia_ManHashXor(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x^1)
	Gia_ManHashStop(p)
	for _, pivot := range []int32{a >> 1, b >> 1, c >> 1, x >> 1, y >> 1} {
		result := Gia_ManDupCof(p, pivot)
		if result == nil {
			t.Fatalf("pivot %d failed", pivot)
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, e1 := p.Evaluate(in)
			got, e2 := result.Evaluate(in)
			if e1 != nil || e2 != nil {
				t.Fatalf("evaluate: %v %v", e1, e2)
			}
			for i := range want {
				if want[i] != got[i] {
					t.Fatalf("pivot %d input %d output %d", pivot, bits, i)
				}
			}
		}
	}
	if Gia_ManDupCof(p, 0) != nil {
		t.Fatal("constant accepted as pivot")
	}
}

func TestCofactorFanoutAndMffc(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	x := Gia_ManHashAnd(p, a, b)
	y := Gia_ManHashAnd(p, x, c)
	z := Gia_ManHashAnd(p, x, b^1)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, z)
	n := Cof_ManCreateLogicSimple(p)
	if Cof_ObjMffcSize(n.objects[x>>1]) != 1 {
		t.Fatal("unexpected MFFC size")
	}
	if Cof_ManCountRemoved(n, n.objects[a>>1], 0) != 3 {
		t.Fatal("wrong zero-cofactor simplification")
	}
	if Cof_ObjMffcSize(n.objects[x>>1]) != 1 {
		t.Fatal("MFFC references not restored")
	}
}

func TestCofactorAllEquivalence(t *testing.T) {
	p := Gia_ManStart(30)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	x := Gia_ManHashAnd(p, a, b)
	y := Gia_ManHashAnd(p, x, c)
	z := Gia_ManHashOr(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, z)
	Gia_ManHashStop(p)
	out := Gia_ManDupCofAll(p, 2, 0)
	if out == nil {
		t.Fatal("cofactoring failed")
	}
	for bits := 0; bits < 8; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := p.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("input %v output %d", input, i)
			}
		}
	}
}
