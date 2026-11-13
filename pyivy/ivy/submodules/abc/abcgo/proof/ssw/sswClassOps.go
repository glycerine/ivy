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

func Ssw_ClassesRemoveNode(p *Ssw_Cla_t, obj aig.Edge) {
	if p.pClassSizes[obj.NodeID()] != 0 || p.pId2Class[obj.NodeID()] != nil {
		panic("ssw: cannot remove class representative")
	}
	repr := aig.Aig_ObjRepr(p.pAig, obj)
	if repr == aig.InvalidEdge {
		panic("ssw: node has no representative")
	}
	if repr == p.pAig.Const1() {
		aig.Aig_ObjClearRepr(p.pAig, obj)
		p.nCands1--
		return
	}
	id := repr.NodeID()
	class := p.pId2Class[id]
	if len(class) < 2 || class[0] != repr {
		panic("ssw: invalid class")
	}
	index := -1
	for i, x := range class {
		if x == obj {
			index = i
			break
		}
	}
	if index < 1 {
		panic("ssw: node not in class")
	}
	aig.Aig_ObjClearRepr(p.pAig, obj)
	if len(class) == 2 {
		p.pId2Class[id] = nil
		p.pClassSizes[id] = 0
		p.nClasses--
	} else {
		copy(class[index:], class[index+1:])
		class = class[:len(class)-1]
		p.pId2Class[id] = class
		p.pClassSizes[id] = int32(len(class))
	}
	p.nLits--
}

func Ssw_ClassesCheck(p *Ssw_Cla_t) {
	var classes, literals, constants int32
	for id, class := range p.pId2Class {
		if class == nil {
			continue
		}
		if len(class) < 2 || int32(len(class)) != p.pClassSizes[id] || class[0].NodeID() != id || aig.Aig_ObjRepr(p.pAig, class[0]) != aig.InvalidEdge {
			panic("ssw: malformed class")
		}
		for i := 1; i < len(class); i++ {
			if class[i-1].NodeID() >= class[i].NodeID() || aig.Aig_ObjRepr(p.pAig, class[i]) != class[0] {
				panic("ssw: malformed class member")
			}
			literals++
		}
		classes++
	}
	for _, obj := range p.pAig.LiveEdges() {
		constants += Ssw_ObjIsConst1Cand(p.pAig, obj)
	}
	if classes != p.nClasses || literals != p.nLits || constants != p.nCands1 {
		panic("ssw: class statistics disagree")
	}
}
