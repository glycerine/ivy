package mpm

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Mig_ManStart and Mig_ManStop implement the subject-graph manager lifecycle.
func Mig_ManStart() *Mig_Man_t {
	assert(1<<MIG_BASE == MIG_MASK+1, "MIG page size")
	p := &Mig_Man_t{}
	vec.Vec_IntGrow(&p.vCis, 1024)
	vec.Vec_IntGrow(&p.vCos, 1024)
	Mig_ManAppendObj(p)
	return p
}
func Mig_ManStop(p *Mig_Man_t) {
	if p != nil {
		*p = Mig_Man_t{}
	}
}
func Mig_ManTypeNum(p *Mig_Man_t, typ int32) int32 {
	var count int32
	for i := int32(0); i < p.nObjs; i++ {
		obj := Mig_ManObj(p, i)
		if Mig_ObjIsNode(obj) != 0 && Mig_ObjNodeType(obj) == typ {
			count++
		}
	}
	return count
}
func Mig_ManAndNum(p *Mig_Man_t) int32 { return Mig_ManTypeNum(p, 1) }
func Mig_ManXorNum(p *Mig_Man_t) int32 { return Mig_ManTypeNum(p, 2) }
func Mig_ManMuxNum(p *Mig_Man_t) int32 { return Mig_ManTypeNum(p, 3) }
func Mig_ManSetRefs(p *Mig_Man_t) {
	vec.Vec_IntFill(&p.vRefs, int(p.nObjs), 0)
	for id := int32(0); id < p.nObjs; id++ {
		obj := Mig_ManObj(p, id)
		for i := int32(0); i < 3 && Mig_ObjHasFanin(obj, i) != 0; i++ {
			fanin := Mig_ObjFaninId(obj, i)
			// The C iterator uses the assigned ID as its loop condition.
			if fanin == 0 {
				break
			}
			vec.Vec_IntAddToEntry(&p.vRefs, int(fanin), 1)
		}
		if sibling := Mig_ObjSiblId(obj); sibling != 0 {
			vec.Vec_IntAddToEntry(&p.vRefs, int(sibling), 1)
		}
	}
}
func Mig_ManSuppSize_rec(obj *Mig_Obj_t) int32 {
	if obj == nil || Mig_ObjIsTravIdCurrent(obj) != 0 {
		return 0
	}
	Mig_ObjSetTravIdCurrent(obj)
	if Mig_ObjIsCi(obj) != 0 {
		return 1
	}
	assert(Mig_ObjIsNode(obj) != 0, "Mig_ObjIsNode(obj)")
	return Mig_ManSuppSize_rec(Mig_ObjFanin0(obj)) + Mig_ManSuppSize_rec(Mig_ObjFanin1(obj)) + Mig_ManSuppSize_rec(Mig_ObjFanin2(obj))
}
func Mig_ManSuppSize2_rec(p *Mig_Man_t, id int32) int32 {
	if id == MIG_NONE || Mig_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Mig_ObjSetTravIdCurrentId(p, id)
	obj := Mig_ManObj(p, id)
	if Mig_ObjIsCi(obj) != 0 {
		return 1
	}
	assert(Mig_ObjIsNode(obj) != 0, "Mig_ObjIsNode(obj)")
	return Mig_ManSuppSize2_rec(p, Mig_ObjFaninId0(obj)) + Mig_ManSuppSize2_rec(p, Mig_ObjFaninId1(obj)) + Mig_ManSuppSize2_rec(p, Mig_ObjFaninId2(obj))
}
func Mig_ManSuppSizeOne(obj *Mig_Obj_t) int32 {
	Mig_ObjIncrementTravId(obj)
	return Mig_ManSuppSize2_rec(Mig_ObjMan(obj), Mig_ObjId(obj))
}
func Mig_ManSuppSizeTest(p *Mig_Man_t) int32 {
	start := util.Abc_Clock()
	var count int32
	for i := int32(0); i < p.nObjs; i++ {
		obj := Mig_ManObj(p, i)
		if Mig_ObjIsNode(obj) != 0 && Mig_ManSuppSizeOne(obj) <= 16 {
			count++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\n", count, Mig_ManNodeNum(p))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
	return count
}
