package gia

import "testing"

func TestDupConeSupportAndBack(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, c^1)
	Gia_ManAppendCo(p, x)
	ids := permutation(2, 0)
	cone := Gia_ManDupConeSupp(p, x^1, ids)
	for i := int32(0); i < p.nObjs; i++ {
		if Gia_ObjCopyArray(p, i) != -1 {
			t.Fatalf("copy scratch not restored at %d", i)
		}
	}
	Gia_ManHashStart(p)
	inserted := Gia_ManDupConeBack(p, cone, ids)
	Gia_ManAppendCo(p, inserted)
	Gia_ManHashStop(p)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := cone.Evaluate([]bool{in[2], in[0]})
		if e != nil || got[0] == (in[0] && !in[2]) {
			t.Fatalf("cone %v %v", got, e)
		}
		original, e := p.Evaluate(in)
		if e != nil || original[1] != got[0] {
			t.Fatalf("inserted %v %v", original, e)
		}
	}
	_ = b
}
