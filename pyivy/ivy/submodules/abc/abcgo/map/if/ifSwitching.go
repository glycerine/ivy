package if_

import "github.com/glycerine/abcgo/aig/gia"

// If_ManComputeSwitching follows src/base/abci/abcIf.c: reconstruct an AIG
// in IF object order, then simulate 48 frames with a 16-frame warm-up.
// It lives here so the ABC and GIA interfaces share the implementation.
func If_ManComputeSwitching(p *If_Man_t) {
	clk := Abc_Clock()
	assert(p.vSwitching == nil, "p.vSwitching == nil")
	graph := Gia_ManStart(If_ManObjNum(p))
	defer Gia_ManStop(graph)
	// IF const1 is the complement of GIA const0.
	copies := make([]int32, 1, If_ManObjNum(p))
	copies[0] = 1
	for i := int32(0); i < If_ManCiNum(p); i++ {
		copies = append(copies, Gia_ManAppendCi(graph))
	}
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		if If_ObjIsAnd(obj) == 0 {
			continue
		}
		a := copies[If_ObjFanin0(obj).Id] ^ If_ObjFaninC0(obj)
		b := copies[If_ObjFanin1(obj).Id] ^ If_ObjFaninC1(obj)
		copies = append(copies, gia.Gia_ManAppendAnd(graph, a, b))
	}
	for i := int32(0); i < If_ManCoNum(p); i++ {
		obj := If_ManCo(p, i)
		lit := copies[If_ObjFanin0(obj).Id] ^ If_ObjFaninC0(obj)
		copies = append(copies, Gia_ManAppendCo(graph, lit))
	}
	assert(len(copies) == int(If_ManObjNum(p)), "one switching entry per IF object")
	p.vSwitching = gia.Gia_ManComputeSwitchProbs(graph, 48, 16, 0)
	if p.pPars.FVerbose != 0 {
		Abc_PrintTime(1, "Computing switching activity", Abc_Clock()-clk)
	}
}
