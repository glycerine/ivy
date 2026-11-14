package gia

import "testing"

func TestSequentialEquivalenceMergeTransitive(t *testing.T) {
	p := Gia_ManStart(8)
	for i := 0; i < 4; i++ {
		Gia_ManAppendCi(p)
	}
	makePart := func(ids []uint32) *Gia_Man_t {
		part := Gia_ManStart(8)
		for range ids[1:] {
			Gia_ManAppendCi(part)
		}
		for i, id := range ids {
			part.pObjs[i].Value = id
		}
		part.StartRepresentatives()
		return part
	}
	a := makePart([]uint32{0, 1, 2, 4})
	Gia_ObjSetRepr(a, 2, 1)
	a.PNexts = Gia_ManDeriveNexts(a)
	b := makePart([]uint32{0, 2, 3, 4})
	Gia_ObjSetRepr(b, 2, 1)
	Gia_ObjSetRepr(b, 3, 0)
	b.PNexts = Gia_ManDeriveNexts(b)
	Gia_ManSeqEquivMerge(p, []*Gia_Man_t{a, b})
	if Gia_ObjRepr(p, 2) != 1 || Gia_ObjRepr(p, 3) != 1 || Gia_ObjRepr(p, 4) != 0 {
		t.Fatal("classes did not merge transitively")
	}
}
