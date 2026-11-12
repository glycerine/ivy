package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func giaAppendMappedNode(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	a, b, c, kind := src.Gate(obj.id)
	read := func(lit int32) int32 {
		mapped := src.pObjs[lit>>1].Value
		if mapped == ^uint32(0) {
			panic("gia: incomplete graph mapping")
		}
		return int32(mapped) ^ (lit & 1)
	}
	var lit int32
	switch kind {
	case 2:
		lit = Gia_ManAppendAnd(dst, read(a), read(b))
	case 3:
		lit = Gia_ManAppendXorReal(dst, read(a), read(b))
	case 4:
		lit = Gia_ManAppendMuxReal(dst, read(a), read(b), read(c))
	case 5:
		lit = Gia_ManAppendBuf(dst, read(a))
	default:
		panic("gia: expected internal graph node")
	}
	obj.Value = uint32(lit)
}
func Gia_ManDupLevelized(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManLevelNum(src)
	nodes := make([]int32, 0, Gia_ManAndNum(src))
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			nodes = append(nodes, obj.id)
		}
	}
	sort.SliceStable(nodes, func(i, j int) bool { return Gia_ObjLevelId(src, nodes[i]) < Gia_ObjLevelId(src, nodes[j]) })
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for _, id := range nodes {
		giaAppendMappedNode(out, src, src.pObjs[id])
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupFromVecs(src *Gia_Man_t, cis, ands, cos *Vec_Int_t, regs int32) *Gia_Man_t {
	if regs < 0 || int(regs) > vec.Vec_IntSize(cis) || int(regs) > vec.Vec_IntSize(cos) {
		panic("gia: invalid copied register count")
	}
	out := giaDupMappedStart(src)
	for _, id := range vec.Vec_IntArray(cis) {
		obj := Gia_ManObj(src, id)
		if Gia_ObjIsCi(obj) == 0 || obj.Value != ^uint32(0) {
			panic("gia: invalid copied input")
		}
		obj.Value = uint32(Gia_ManAppendCi(out))
	}
	for _, id := range vec.Vec_IntArray(ands) {
		obj := Gia_ManObj(src, id)
		if Gia_ObjIsAnd(obj) == 0 || obj.Value != ^uint32(0) {
			panic("gia: invalid copied gate")
		}
		giaAppendMappedNode(out, src, obj)
	}
	for _, id := range vec.Vec_IntArray(cos) {
		obj := Gia_ManObj(src, id)
		if Gia_ObjIsCo(obj) == 0 {
			panic("gia: invalid copied output")
		}
		giaDupMappedOutput(out, src, obj)
	}
	Gia_ManSetRegNum(out, regs)
	return out
}
func Gia_ManDupSliced(src *Gia_Man_t, limit int32) *Gia_Man_t {
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	copied, exposed := make([]bool, src.nObjs), make([]bool, src.nObjs)
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ManSuppSize(src, []int32{obj.id}, 1) <= limit {
			giaAppendMappedNode(out, src, obj)
			copied[obj.id] = true
		} else {
			for _, id := range giaObjectFanins(src, obj) {
				exposed[id] = true
			}
		}
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		for _, id := range giaObjectFanins(src, Gia_ManCo(src, i)) {
			exposed[id] = true
		}
	}
	for id := int32(1); id < src.nObjs; id++ {
		if copied[id] && exposed[id] {
			Gia_ManAppendCo(out, int32(src.pObjs[id].Value))
		}
	}
	return out
}

func Gia_ManDupOuts(src *Gia_Man_t) *Gia_Man_t {
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManAppendCo(out, int32(obj.Value))
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
