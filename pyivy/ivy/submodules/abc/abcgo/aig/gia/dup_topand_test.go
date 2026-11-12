package gia

import "testing"

func TestTopAndExtractsConjuncts(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, c, d)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, y^1))
	out := Gia_ManDupTopAnd_iter(p, 0)
	if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 1 {
		t.Fatal("incorrect decomposition interface")
	}
	for bits := 0; bits < 16; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] == (in[2] && in[3]) {
			t.Fatalf("decomposed %v %v", got, e)
		}
	}
}
func TestDupClassesUsesRepresentative(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, y)
	p.pReprsOld = make([]int32, p.nObjs)
	for i := range p.pReprsOld {
		p.pReprsOld[i] = -1
	}
	p.pReprsOld[y>>1] = x >> 1
	out := Gia_ManDupDfsClasses(p)
	if Gia_ManAndNum(out) != 1 {
		t.Fatal("class not reduced")
	}
	got, e := out.Evaluate([]bool{true, true})
	if e != nil || !got[0] {
		t.Fatalf("class copy %v %v", got, e)
	}
}
