package gia

import "testing"

type testDomainCorrespondence struct{ calls int }

func (e *testDomainCorrespondence) SweepCorrespondence(p *Gia_Man_t) {
	e.calls++
	if Gia_ManRegNum(p) != 2 {
		panic("wrong active domain size")
	}
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	a, b := Gia_ManRo(p, 0).id, Gia_ManRo(p, 1).id
	Gia_ObjSetRepr(p, b, a)
}
func TestSweepOneDomainRemapsClasses(t *testing.T) {
	p := Gia_ManStart(16)
	in := Gia_ManAppendCi(p)
	r0, r1, r2 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, in)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManAppendCo(p, r2)
	Gia_ManSetRegNum(p, 3)
	classes := Vec_IntAlloc(3)
	Vec_IntPush(classes, 1)
	Vec_IntPush(classes, 2)
	Vec_IntPush(classes, 2)
	engine := new(testDomainCorrespondence)
	Gia_ManSweepComputeOneDomainEquivs(p, classes, 2, engine, 0, 0, 0)
	if engine.calls != 1 || Gia_ObjReprSelf(p, r2>>1) != r1>>1 || Gia_ObjReprSelf(p, r0>>1) != r0>>1 {
		t.Fatal("incorrect domain representative remapping")
	}
}
