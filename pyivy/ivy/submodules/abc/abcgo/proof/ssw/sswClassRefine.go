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
