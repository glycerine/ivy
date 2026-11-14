package gia

import "testing"

func TestHcdEquivToChoices(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, c)
	z := Gia_ManAppendAnd(p, y, b)
	Gia_ManAppendCo(p, z)
	Gia_ManAppendCo(p, z)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	p.PNexts = make([]int32, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	// No equivalence classes: snapshot reduction must preserve the first output.
	q := Hcd_ManEquivToChoices(p, 2)
	if Gia_ManCoNum(q) != 1 {
		t.Fatalf("outputs %d", Gia_ManCoNum(q))
	}
	for bits := 0; bits < 8; bits++ {
		got, err := q.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
		if err != nil || got[0] != (bits == 7) {
			t.Fatalf("bits=%d got=%v err=%v", bits, got, err)
		}
	}
	_ = x
}

func TestHcdEquivToChoicesKeepsAlternative(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	first := Gia_ManAppendAnd(p, x, c)
	y := Gia_ManAppendAnd(p, a, c)
	second := Gia_ManAppendAnd(p, y, b)
	Gia_ManAppendCo(p, first)
	Gia_ManAppendCo(p, second)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	p.PNexts = make([]int32, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, second>>1, first>>1)
	Gia_ObjSetNext(p, first>>1, second>>1)
	q := Hcd_ManEquivToChoices(p, 2)
	head := Gia_ObjFanin0(Gia_ManCo(q, 0))
	if Gia_ObjNext(q, head.id) == 0 {
		t.Fatal("alternative choice lost")
	}
	for bits := 0; bits < 8; bits++ {
		got, err := q.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
		if err != nil || got[0] != (bits == 7) {
			t.Fatalf("bits=%d got=%v err=%v", bits, got, err)
		}
	}
}
