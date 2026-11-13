package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec2ManagerLifecycle(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	pars := &Cec2_Par_t{nConfLimit: 123}
	p := Cec2_ManCreate(g, pars)
	if p.pSat.Options().ConflictLimit != 123 || g.Data() != p.pSat {
		t.Fatal("solver setup failed")
	}
	if gia.Gia_ObjValue(gia.Gia_ManObj(g, x>>1)) != 2 || gia.Gia_ManCiNum(p.pNew) != 1 {
		t.Fatal("input copy was not initialized")
	}
	if gia.Gia_ObjCopy2Array(p.pNew, 1) != -1 {
		t.Fatal("SAT IDs were not initialized")
	}
	gia.Gia_ManCecSimAlloc(g, 1)
	Cec2_ManDestroy(p)
	if g.Data() != nil || p.pNew != nil || p.pSat != nil {
		t.Fatal("manager resources were not released")
	}
}
