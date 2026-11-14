package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestKfManagerLifecycleAndRefs(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	pars := &Jf_Par_t{NLutSize: 4, NCutNum: 4}
	p := Kf_ManAlloc(g, pars)
	for _, lit := range []int32{a, b} {
		Kf_ManStoreStart(p.vTemp, 0)
		Kf_ManStoreAddUnit(p.vTemp, lit>>1, 0, 0)
		Kf_ObjSetCuts(p, lit>>1, p.vTemp)
	}
	c := &Kf_Cut_t{nLeaves: 2}
	c.pLeaves[0], c.pLeaves[1] = a>>1, b>>1
	Kf_ManSaveResults([]*Kf_Cut_t{c}, 1, c, p.vTemp)
	Kf_ObjSetCuts(p, x>>1, p.vTemp)
	if Kf_ManComputeRefs(p) != 1 || pars.Delay != 1 || pars.Edge != 2 {
		t.Fatalf("area=%d delay=%d edge=%d", pars.Area, pars.Delay, pars.Edge)
	}
	if Kf_CutSize(Kf_ObjCutBest(p, x>>1)) != 2 || vec.Vec_FltEntry(&p.vRefs, int(a>>1)) < 1 {
		t.Fatal("cut or refs")
	}
	Kf_ManFree(p)
	if g.Refs != nil || len(p.pMem.lists) != 0 {
		t.Fatal("cleanup")
	}
}
