package gia

import "testing"

func TestDupChoiceLinks(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	head := Gia_ManAppendAnd(p, a, b)
	alternative := Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, head)
	p.StartRepresentatives()
	p.PNexts[head>>1] = alternative >> 1
	out := Gia_ManDupOrderDfsChoices(p)
	newHead := int32(p.pObjs[head>>1].Value >> 1)
	newAlt := int32(p.pObjs[alternative>>1].Value >> 1)
	if newAlt >= newHead || out.PNexts[newHead] != newAlt {
		t.Fatalf("bad link %d -> %d; expected %d", newHead, out.PNexts[newHead], newAlt)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != want[0] {
			t.Fatalf("bad copy %v %v", got, e)
		}
	}
}
