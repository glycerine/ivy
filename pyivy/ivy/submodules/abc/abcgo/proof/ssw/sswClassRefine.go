package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_ClassesRefineOneClass splits a class according to the equality callback.
func Ssw_ClassesRefineOneClass(p *Ssw_Cla_t, repr aig.Edge, recursive int32) int32 {
	class := p.pId2Class[repr.NodeID()]
	if class == nil {
		return 0
	}
	oldGroup, newGroup := make([]aig.Edge, 0, len(class)), make([]aig.Edge, 0, len(class))
	for _, obj := range class {
		if p.pFuncNodesAreEqual(p.pManData, repr, obj) != 0 {
			oldGroup = append(oldGroup, obj)
		} else {
			newGroup = append(newGroup, obj)
		}
	}
	if len(newGroup) == 0 {
		return 0
	}
	if len(oldGroup) == 0 {
		panic("ssw: representative is not equal to itself")
	}
	Ssw_ObjRemoveClass(p, repr)
	for _, group := range [][]aig.Edge{oldGroup, newGroup} {
		representative := group[0]
		aig.Aig_ObjClearRepr(p.pAig, representative)
		for _, obj := range group[1:] {
			aig.Aig_ObjSetRepr(p.pAig, obj, representative)
		}
		if len(group) > 1 {
			Ssw_ObjAddClass(p, representative, group, int32(len(group)))
		}
	}
	if recursive != 0 && len(newGroup) > 1 {
		return 1 + Ssw_ClassesRefineOneClass(p, newGroup[0], 1)
	}
	return 1
}
func Ssw_ClassesRefine(p *Ssw_Cla_t, recursive int32) int32 {
	var count int32
	for i := 0; i < len(p.pId2Class); i++ {
		if class := p.pId2Class[i]; class != nil {
			count += Ssw_ClassesRefineOneClass(p, class[0], recursive)
		}
	}
	return count
}
func Ssw_ClassesRefineGroup(p *Ssw_Cla_t, reprs *vec.Vec_Ptr_t, recursive int32) int32 {
	var count int32
	for i := 0; i < vec.Vec_PtrSize(reprs); i++ {
		count += Ssw_ClassesRefineOneClass(p, vec.Vec_PtrEntry(reprs, i).(aig.Edge), recursive)
	}
	return count
}

func sswRefineConstants(p *Ssw_Cla_t, roots []aig.Edge, recursive int32, constOnly bool) int32 {
	candidates := make([]aig.Edge, 0, len(roots))
	for _, obj := range roots {
		if p.pFuncNodeIsConst(p.pManData, obj) == 0 {
			candidates = append(candidates, obj)
		}
	}
	if len(candidates) == 0 {
		return 0
	}
	p.nCands1 -= int32(len(candidates))
	if constOnly {
		for _, obj := range candidates {
			aig.Aig_ObjClearRepr(p.pAig, obj)
		}
		return 1
	}
	repr := candidates[0]
	aig.Aig_ObjClearRepr(p.pAig, repr)
	if len(candidates) == 1 {
		return 1
	}
	for _, obj := range candidates[1:] {
		aig.Aig_ObjSetRepr(p.pAig, obj, repr)
	}
	Ssw_ObjAddClass(p, repr, candidates, int32(len(candidates)))
	if recursive != 0 {
		return 1 + Ssw_ClassesRefineOneClass(p, repr, 1)
	}
	return 1
}
func Ssw_ClassesRefineConst1Group(p *Ssw_Cla_t, roots *vec.Vec_Ptr_t, recursive int32) int32 {
	objects := make([]aig.Edge, vec.Vec_PtrSize(roots))
	for i := range objects {
		objects[i] = vec.Vec_PtrEntry(roots, i).(aig.Edge)
	}
	return sswRefineConstants(p, objects, recursive, false)
}
func Ssw_ClassesRefineConst1(p *Ssw_Cla_t, recursive int32) int32 {
	var roots []aig.Edge
	for _, obj := range p.pAig.LiveEdges() {
		if aig.Aig_ObjRepr(p.pAig, obj) == p.pAig.Const1() {
			roots = append(roots, obj)
		}
	}
	return sswRefineConstants(p, roots, recursive, p.fConstCorr != 0)
}
