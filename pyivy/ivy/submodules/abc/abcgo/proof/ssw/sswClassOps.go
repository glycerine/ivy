package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_ObjNext(next []aig.Edge, obj aig.Edge) aig.Edge      { return next[obj.NodeID()] }
func Ssw_ObjSetNext(next []aig.Edge, obj, successor aig.Edge) { next[obj.NodeID()] = successor }
func Ssw_ObjAddClass(p *Ssw_Cla_t, repr aig.Edge, class []aig.Edge, size int32) {
	id := repr.NodeID()
	if p.pId2Class[id] != nil || p.pClassSizes[id] != 0 || size <= 1 || int(size) > len(class) || class[0] != repr {
		panic("ssw: invalid class")
	}
	p.pId2Class[id] = class[:size]
	p.pClassSizes[id] = size
	p.nClasses++
	p.nLits += size - 1
}
func Ssw_ObjRemoveClass(p *Ssw_Cla_t, repr aig.Edge) []aig.Edge {
	id := repr.NodeID()
	class, size := p.pId2Class[id], p.pClassSizes[id]
	if class == nil || size <= 1 {
		panic("ssw: missing class")
	}
	p.pId2Class[id] = nil
	p.pClassSizes[id] = 0
	p.nClasses--
	p.nLits -= size - 1
	return class
}
func Ssw_ClassesReadClass(p *Ssw_Cla_t, repr aig.Edge, size *int32) []aig.Edge {
	id := repr.NodeID()
	class := p.pId2Class[id]
	if class == nil {
		return nil
	}
	if p.pClassSizes[id] <= 1 {
		panic("ssw: invalid class size")
	}
	*size = p.pClassSizes[id]
	return class
}
func Ssw_ClassesCollectClass(p *Ssw_Cla_t, repr aig.Edge, out *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(out)
	id := repr.NodeID()
	class := p.pId2Class[id]
	if class == nil {
		return
	}
	if p.pClassSizes[id] <= 1 {
		panic("ssw: invalid class size")
	}
	for _, obj := range class[1:p.pClassSizes[id]] {
		vec.Vec_PtrPush(out, obj)
	}
}
