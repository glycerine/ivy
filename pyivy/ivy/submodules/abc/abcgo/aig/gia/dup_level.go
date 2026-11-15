package gia

import "github.com/glycerine/abcgo/misc/vec"

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
	levels := Gia_ManLevelNum(src)
	counts := make([]int32, int(levels)+1)
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			counts[Gia_ObjLevel(src, obj)]++
		}
	}
	// Starting offsets include level zero, where buffer chains may live.
	start := int32(0)
	for level, count := range counts {
		counts[level] = start
		start += count
	}
	nodes := make([]int32, Gia_ManAndNum(src))
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			level := Gia_ObjLevel(src, obj)
			nodes[counts[level]] = obj.id
			counts[level]++
		}
	}
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
	out.nConstrs = src.nConstrs
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
	out.PSpec = ""
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	Gia_ManCleanMark01(src)
	defer Gia_ManCleanMark01(src)
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ManSuppSize(src, []int32{obj.id}, 1) <= limit {
			giaAppendMappedNode(out, src, obj)
			obj.fMark0 = 1
		} else {
			for _, id := range giaObjectFanins(src, obj) {
				src.pObjs[id].fMark1 = 1
			}
		}
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ObjFanin0(Gia_ManCo(src, i)).fMark1 = 1
	}
	for id := int32(1); id < src.nObjs; id++ {
		obj := src.pObjs[id]
		if obj.fMark0 != 0 && obj.fMark1 != 0 {
			Gia_ManAppendCo(out, int32(obj.Value))
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
