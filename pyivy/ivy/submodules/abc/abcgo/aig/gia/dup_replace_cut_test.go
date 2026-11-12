package gia

import "testing"

func TestReplaceCutSeparatesPIAndRegisterLogic(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, r)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, x)
	Gia_ManSetRegNum(p, 1)
	out := Gia_ManDupReplaceCut(p)
	if Gia_ManPiNum(out) != 2 || Gia_ManRegNum(out) != 1 {
		t.Fatalf("cut interface %d %d", Gia_ManPiNum(out), Gia_ManRegNum(out))
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[2]) || got[1] != in[1] || got[2] != in[0] {
			t.Fatalf("cut %v %v", got, e)
		}
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("marks not cleaned")
		}
	}
}
