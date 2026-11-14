package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEquivalenceSimulationClassRefinement(t *testing.T) {
	g := Gia_ManStart(8)
	for i := 0; i < 4; i++ {
		Gia_ManAppendCi(g)
	}
	g.pReprs = make([]Gia_Rpr_t, g.nObjs)
	g.pNexts = make([]int32, g.nObjs)
	for id := int32(1); id < g.nObjs; id++ {
		Gia_ObjSetRepr(g, id, GIA_VOID)
	}
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 1
	p := Gia_Sim2Create(g, &pars)
	defer Gia_Sim2Delete(p)
	class := vec.Vec_IntAlloc(4)
	for id := int32(1); id <= 4; id++ {
		vec.Vec_IntPush(class, id)
	}
	Gia_Sim2ClassCreate(g, class)
	for id := int32(1); id <= 4; id++ {
		v := uint32(0xaaaaaaaa)
		if id > 2 {
			v = 0xcccccccc
		}
		Gia_Sim2Data(p, id)[0] = v
	}
	if Gia_Sim2ClassRefineOne(p, 1) != 1 || Gia_ObjNext(g, 1) != 2 || Gia_ObjRepr(g, 2) != 1 || Gia_ObjNext(g, 3) != 4 || Gia_ObjRepr(g, 4) != 3 || Gia_ObjNext(g, 2) != 0 {
		t.Fatal("incorrect class split")
	}
}

func TestEquivalenceConstantRefinement(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.pReprs = make([]Gia_Rpr_t, g.nObjs)
	g.pNexts = make([]int32, g.nObjs)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 1
	p := Gia_Sim2Create(g, &pars)
	defer Gia_Sim2Delete(p)
	Gia_Sim2Data(p, a>>1)[0] = 0xaaaaaaaa
	Gia_Sim2Data(p, b>>1)[0] = 0xaaaaaaaa
	Gia_Sim2InfoRefineEquivs(p)
	if Gia_ObjRepr(g, a>>1) != GIA_VOID || Gia_ObjRepr(g, b>>1) != a>>1 || Gia_ObjNext(g, a>>1) != b>>1 {
		t.Fatal("matching nonconstant candidates not grouped")
	}
}
