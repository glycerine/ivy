package mpm

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Mpm_CutCopySet(p *Mpm_Man_t, obj *Mig_Obj_t, complement int32) int32 {
	var head int32
	next := &head
	for h := Mpm_ObjCutList(p, obj); h != 0; h = Mpm_CutFetch(p, h).hNext {
		*next = Mpm_CutDup(p, Mpm_CutFetch(p, h), complement)
		next = &Mpm_CutFetch(p, *next).hNext
	}
	*next = 0
	return head
}
func Mpm_CutPrint(cut *Mpm_Cut_t) {
	fmt.Printf("%d : { ", cut.nLeaves)
	for _, lit := range cut.pLeaves[:cut.nLeaves] {
		fmt.Printf("%d ", lit)
	}
	fmt.Println("}")
}
func Mpm_CutPrintAll(p *Mpm_Man_t) {
	for i := int32(0); i < p.nCutStore; i++ {
		fmt.Printf("%2d : ", i)
		Mpm_CutPrint(&p.pCutStore[i].pCut)
	}
}
func Mpm_CutIsContained(p *Mpm_Man_t, base, cut *Mpm_Cut_t) int32 {
	for _, lit := range cut.pLeaves[:cut.nLeaves] {
		if Mpm_CutFindLeaf(base, util.Abc_Lit2Var(lit)) == int32(base.nLeaves) {
			return 0
		}
	}
	return 1
}
func Mpm_CutGetArea(p *Mpm_Man_t, cut *Mpm_Cut_t) int32 {
	class := int32(cut.iFunc) >> 1
	if p.pPars.FMap4Cnf != 0 {
		return MPM_UNIT_AREA * p.pDsd6[class].nClauses
	}
	if p.pPars.FMap4Aig != 0 {
		return MPM_UNIT_AREA * p.pDsd6[class].nAnds
	}
	if p.pPars.FMap4Gates != 0 {
		return MPM_UNIT_AREA
	}
	return p.pLibLut.pLutAreas[cut.nLeaves]
}
func Mpm_CutGetSign(cut *Mpm_Cut_t) word {
	var sign word
	for _, lit := range cut.pLeaves[:cut.nLeaves] {
		sign |= word(1) << (util.Abc_Lit2Var(lit) & 63)
	}
	return sign
}
func Mpm_CutGetArrTime(p *Mpm_Man_t, cut *Mpm_Cut_t) int32 {
	var arrival int32
	for i, lit := range cut.pLeaves[:cut.nLeaves] {
		arrival = max(arrival, vec.Vec_IntEntry(&p.vTimes, int(util.Abc_Lit2Var(lit)))+p.pLibLut.pLutDelays[cut.nLeaves][i])
	}
	return arrival
}
func Mpm_CutSetupInfo(p *Mpm_Man_t, cut *Mpm_Cut_t, arrival int32) *Mpm_Uni_t {
	unit := vec.Vec_PtrEntryLast(&p.vFreeUnits).(*Mpm_Uni_t)
	assert(&unit.pCut == cut, "free cut unit")
	unit.mTime = arrival
	unit.mArea = Mpm_CutGetArea(p, cut)
	unit.mEdge = MPM_UNIT_EDGE * int32(cut.nLeaves)
	unit.mAveRefs = 0
	unit.Cost = 0
	unit.uSign = 0
	for _, lit := range cut.pLeaves[:cut.nLeaves] {
		id := int(util.Abc_Lit2Var(lit))
		mapped := vec.Vec_IntEntry(&p.vMapRefs, id)
		area := vec.Vec_IntEntry(&p.vAreas, id)
		edge := vec.Vec_IntEntry(&p.vEdges, id)
		if p.fMainRun != 0 && mapped == 0 {
			unit.mArea += area
			unit.mEdge += edge
		} else {
			est := vec.Vec_IntEntry(&p.vEstRefs, id)
			assert(est > 0, "estimated references")
			unit.mArea += MPM_UNIT_REFS * area / est
			unit.mEdge += MPM_UNIT_REFS * edge / est
			if p.fMainRun != 0 {
				unit.mAveRefs += mapped
			} else {
				unit.mAveRefs += vec.Vec_IntEntry(&p.vMigRefs, id)
			}
		}
		unit.uSign |= word(1) << (id & 63)
	}
	unit.mAveRefs = unit.mAveRefs * MPM_UNIT_EDGE / max(int32(cut.nLeaves), 1)
	assert(unit.mTime <= 0x3fffffff && unit.mArea <= 0x3fffffff && unit.mEdge <= 0x3fffffff, "cut cost range")
	return unit
}
