package gia

import "testing"

func TestCofactorQuantificationAndMux(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	f := Gia_ManHashMux(p, a, b, c)
	Gia_ManAppendCo(p, f)
	Gia_ManHashStop(p)
	neg := Gia_ManDupCofactorVar(p, 0, 0)
	pos := Gia_ManDupCofactorVar(p, 0, 1)
	mux := Gia_ManDupMux(0, pos, neg)
	exists := Gia_ManDupExist(p, 0)
	all := Gia_ManDupUniv(p, 0)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		n, e := neg.Evaluate(in)
		if e != nil || n[0] != in[2] {
			t.Fatalf("negative %v %v", n, e)
		}
		v, e := pos.Evaluate(in)
		if e != nil || v[0] != in[1] {
			t.Fatalf("positive %v %v", v, e)
		}
		got, e := mux.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[1] || !in[0] && in[2]) {
			t.Fatalf("mux %v %v", got, e)
		}
		got, e = exists.Evaluate(in)
		if e != nil || got[0] != (in[1] || in[2]) {
			t.Fatalf("exists %v %v", got, e)
		}
		got, e = all.Evaluate(in)
		if e != nil || got[0] != (in[1] && in[2]) {
			t.Fatalf("all %v %v", got, e)
		}
	}
}
func TestSequentialExistentialOutputs(t *testing.T) {
	p := sequentialFixture()
	out := Gia_ManDupExist2(p, 0)
	if Gia_ManRegNum(out) != 1 {
		t.Fatal("lost register")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != in[1] || !got[1] {
			t.Fatalf("exists outputs %v %v", got, e)
		}
	}
}

func TestInternalCofactorAndBlock(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, c))
	for _, value := range []int32{0, 1} {
		out := Gia_ManDupCofactorObj(p, x>>1, value)
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			got, e := out.Evaluate(in)
			guard := (in[0] && in[1]) == (value == 1)
			want := guard && (value == 1) && in[2]
			if e != nil || got[0] != want {
				t.Fatalf("cofactor %d %v %v", value, got, e)
			}
		}
	}
	q := Gia_ManStart(20)
	u, v, w, z := Gia_ManAppendCi(q), Gia_ManAppendCi(q), Gia_ManAppendCi(q), Gia_ManAppendCi(q)
	Gia_ManAppendCo(q, Gia_ManAppendAnd(q, u, w))
	Gia_ManAppendCo(q, Gia_ManAppendAnd(q, v, z))
	out := Gia_ManDupBlock(q, 2)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[1]) {
			t.Fatalf("block %v %v", got, e)
		}
	}
}
