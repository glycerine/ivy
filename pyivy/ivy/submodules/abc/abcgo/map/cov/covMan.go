package cov

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/vec"
)

// Cov_ManAlloc attaches per-object cover state to pNtk for the duration of a
// mapping pass. The object table is indexed by stable ABC object IDs.
func Cov_ManAlloc(pNtk *Abc_Ntk_t, nFaninMax, nCubesMax int32) *Cov_Man_t {
	if pNtk == nil || nFaninMax < 0 || nCubesMax < 0 {
		panic("cov: invalid manager arguments")
	}
	if current, ok := managerRegistry.Load(pNtk); ok && current != nil {
		panic("cov: network already has a cover manager")
	}
	maxID := -1
	for _, object := range pNtk.Objects() {
		if object.ID > maxID {
			maxID = object.ID
		}
	}
	objectInfo := make([]*Cov_Obj_t, maxID+1)
	objectPointers := vec.Vec_PtrStart(maxID + 1)
	for _, object := range pNtk.Objects() {
		info := &Cov_Obj_t{}
		objectInfo[object.ID] = info
		vec.Vec_PtrWriteEntry(objectPointers, object.ID, info)
	}
	fanouts := vec.Vec_IntStart(maxID + 1)
	for _, object := range pNtk.Objects() {
		vec.Vec_IntWriteEntry(fanouts, object.ID, int32(len(object.Fanouts())))
	}
	p := &Cov_Man_t{
		nFaninMax:  nFaninMax,
		nCubesMax:  nCubesMax,
		nWords:     (2*nFaninMax + 31) / 32,
		vFanCounts: fanouts,
		vObjStrs:   objectPointers,
		pManMin:    Min_ManAlloc(nFaninMax),
		vComTo0:    vec.Vec_IntAlloc(int(2 * nFaninMax)),
		vComTo1:    vec.Vec_IntAlloc(int(2 * nFaninMax)),
		vPairs0:    vec.Vec_IntAlloc(int(nFaninMax)),
		vPairs1:    vec.Vec_IntAlloc(int(nFaninMax)),
		vTriv0:     vec.Vec_IntAlloc(1),
		vTriv1:     vec.Vec_IntAlloc(1),
		objectInfo: objectInfo,
		network:    pNtk,
	}
	vec.Vec_IntPush(p.vTriv0, -1)
	vec.Vec_IntPush(p.vTriv1, -1)
	managerRegistry.Store(pNtk, p)
	return p
}

// Cov_ManFree releases per-pass state and removes this manager's registry entry.
func Cov_ManFree(p *Cov_Man_t) {
	if p == nil {
		return
	}
	for _, info := range p.objectInfo {
		if info != nil && info.vSupp != nil {
			vec.Vec_IntFree(info.vSupp)
			info.vSupp = nil
		}
	}
	Min_ManFree(p.pManMin)
	vec.Vec_PtrFree(p.vObjStrs)
	vec.Vec_IntFree(p.vFanCounts)
	vec.Vec_IntFree(p.vTriv0)
	vec.Vec_IntFree(p.vTriv1)
	vec.Vec_IntFree(p.vComTo0)
	vec.Vec_IntFree(p.vComTo1)
	vec.Vec_IntFree(p.vPairs0)
	vec.Vec_IntFree(p.vPairs1)
	if p.network != nil {
		if current, ok := managerRegistry.Load(p.network); ok && current == p {
			managerRegistry.Delete(p.network)
		}
	}
	p.objectInfo = nil
	p.network = nil
}

// Abc_NodeCovDropData drops a node's covers once its final fanout is consumed.
func Abc_NodeCovDropData(p *Cov_Man_t, pObj *Abc_Obj_t) {
	if p == nil || pObj == nil || p.vFanCounts == nil {
		panic("cov: dropping data without an active manager")
	}
	index := pObj.ID
	if index < 0 || index >= vec.Vec_IntSize(p.vFanCounts) {
		panic("cov: object ID outside fanout table")
	}
	remaining := vec.Vec_IntEntry(p.vFanCounts, index) - 1
	if remaining < 0 {
		panic(fmt.Sprintf("cov: negative fanout count for object %d", index))
	}
	vec.Vec_IntWriteEntry(p.vFanCounts, index, remaining)
	if remaining == 0 {
		support := Abc_ObjGetSupp(pObj)
		cover := Abc_ObjGetCover2(pObj)
		if support != nil {
			vec.Vec_IntFree(support)
			Abc_ObjSetSupp(pObj, nil)
		}
		Min_CoverRecycle(p.pManMin, cover)
		Abc_ObjSetCover2(pObj, nil)
		p.nSupps--
	}
}
