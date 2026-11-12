package gia

import "testing"

func TestCec4ChoicesWithoutClasses(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	p.pNexts = make([]int32, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	q := Gia_ManDup(p)
	result := Cec4_ManSatSolverChoices(p, q)
	for bits := 0; bits < 4; bits++ {
		got, err := result.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || got[0] != (bits == 3) {
			t.Fatalf("bits=%d got=%v err=%v", bits, got, err)
		}
	}
}
