package gia

import (
	"path/filepath"
	"testing"
)

func TestEquivalenceFilterUsingAigerParts(t *testing.T) {
	makePart := func() *Gia_Man_t {
		p := Gia_ManStart(8)
		a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		Gia_ManAppendCo(p, r)
		Gia_ManAppendCo(p, a)
		Gia_ManSetRegNum(p, 1)
		return p
	}
	a, b := makePart(), makePart()
	dir := t.TempDir()
	fileA, fileB := filepath.Join(dir, "a.aig"), filepath.Join(dir, "b.aig")
	if err := Gia_AigerWrite(a, fileA, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := Gia_AigerWrite(b, fileB, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	p := Gia_ManMiter(a, b, 0, 0, 1, 0, 0)
	if p == nil {
		t.Fatal("missing miter")
	}
	ra, rb := Gia_ManRo(p, 0).id, Gia_ManRo(p, 1).id
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, rb, ra)
	p.PNexts = Gia_ManDeriveNexts(p)
	if Gia_ManFilterEquivsUsingParts(p, fileA, fileB) != 1 {
		t.Fatal("filter failed")
	}
	if Gia_ObjRepr(p, rb) != ra || p.PNexts[ra] != rb {
		t.Fatal("cross-part latch class not preserved")
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 {
			t.Fatal("temporary marks retained")
		}
	}
}
