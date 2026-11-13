package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec3ManagerLifecycle(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	p := Cec3_ManCreate(g, &Cec3_Par_t{})
	if g.Data() != p.pSat || p.pSat == nil {
		t.Fatal("Glucose solver not installed")
	}
	if gia.Gia_ObjValue(gia.Gia_ManObj(g, x>>1)) != 2 || gia.Gia_ObjCopy2Array(p.pNew, 1) != -1 {
		t.Fatal("copies not initialized")
	}
	Cec3_ManDestroy(p)
	if g.Data() != nil || p.pNew != nil || p.pSat != nil {
		t.Fatal("manager resources not released")
	}
}
