package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Gia_ObjCompareByCioId(a, b *Gia_Obj_t) int32 { return Gia_ObjCioId(a) - Gia_ObjCioId(b) }
func Gia_ManDupCones_rec(p *Gia_Man_t, obj *Gia_Obj_t, leaves, nodes, roots *Vec_Ptr_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) != 0 {
		for _, id := range giaObjectFanins(p, obj) {
			Gia_ManDupCones_rec(p, Gia_ManObj(p, id), leaves, nodes, roots)
		}
		vec.Vec_PtrPush(nodes, obj)
	} else if Gia_ObjIsCo(obj) != 0 {
		Gia_ManDupCones_rec(p, Gia_ObjFanin0(obj), leaves, nodes, roots)
	} else if Gia_ObjIsRo(p, obj) != 0 {
		vec.Vec_PtrPush(roots, Gia_ObjRoToRi(p, obj))
	} else if Gia_ObjIsPi(p, obj) != 0 {
		vec.Vec_PtrPush(leaves, obj)
	} else if Gia_ObjIsConst0(obj) == 0 {
		panic("gia: invalid cone object")
	}
}
func Gia_ManDupCones(p *Gia_Man_t, outputs []int32, count, trim int32) *Gia_Man_t {
	if count < 0 || int(count) > len(outputs) {
		panic("gia: invalid cone output count")
	}
	return giaDupPartitionConesMode(p, outputs[:count], trim != 0)
}
func Gia_ManDupAndCones(p *Gia_Man_t, ands []int32, count, trim int32) *Gia_Man_t {
	if count < 0 || int(count) > len(ands) {
		panic("gia: invalid cone count")
	}
	leaves, nodes, roots := vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0)
	for _, id := range ands[:count] {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: expected AND cone root")
		}
		vec.Vec_PtrPush(roots, obj)
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := 0; i < vec.Vec_PtrSize(roots); i++ {
		Gia_ManDupCones_rec(p, vec.Vec_PtrEntry(roots, i).(*Gia_Obj_t), leaves, nodes, roots)
	}
	entries := vec.Vec_PtrArray(leaves)
	sort.Slice(entries, func(i, j int) bool {
		return Gia_ObjCioId(entries[i].(*Gia_Obj_t)) < Gia_ObjCioId(entries[j].(*Gia_Obj_t))
	})
	out := Gia_ManStart(p.nObjs)
	out.CopyIdentityFrom(p)
	Gia_ManFillValue(p)
	p.pObjs[0].Value = 0
	if trim != 0 {
		for _, entry := range entries {
			entry.(*Gia_Obj_t).Value = uint32(Gia_ManAppendCi(out))
		}
	} else {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			Gia_ManPi(p, i).Value = uint32(Gia_ManAppendCi(out))
		}
	}
	for _, entry := range vec.Vec_PtrArray(nodes) {
		obj := entry.(*Gia_Obj_t)
		a, b, c, kind := p.Gate(obj.id)
		read := func(lit int32) int32 { return Gia_ObjLitCopy(p, lit) }
		var lit int32
		switch kind {
		case 2:
			lit = Gia_ManAppendAnd(out, read(a), read(b))
		case 3:
			lit = Gia_ManAppendXor(out, read(a), read(b))
		case 4:
			lit = Gia_ManAppendMux(out, read(a), read(b), read(c))
		case 5:
			lit = Gia_ManAppendBuf(out, read(a))
		default:
			panic("gia: invalid cone gate")
		}
		obj.Value = uint32(lit)
	}
	for _, entry := range vec.Vec_PtrArray(roots) {
		obj := entry.(*Gia_Obj_t)
		if obj.Value == ^uint32(0) {
			panic("gia: AND cone reaches unsupported register")
		}
		Gia_ManAppendCo(out, int32(obj.Value))
	}
	return out
}

func giaDupLimitedCone(dst, src *Gia_Man_t, id, level int32, relative bool) {
	obj := Gia_ManObj(src, id)
	if obj.Value != ^uint32(0) {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 || relative && level <= 0 || !relative && Gia_ObjLevel(src, obj) < level {
		obj.Value = uint32(Gia_ManAppendCi(dst))
		return
	}
	a, b, c, kind := src.Gate(id)
	next := level
	if relative {
		next--
	}
	read := func(lit int32) int32 {
		giaDupLimitedCone(dst, src, lit>>1, next, relative)
		return Gia_ObjLitCopy(src, lit)
	}
	var result int32
	switch kind {
	case 2:
		result = Gia_ManAppendAnd(dst, read(a), read(b))
	case 3:
		result = Gia_ManAppendXor(dst, read(a), read(b))
	case 4:
		result = Gia_ManAppendMux(dst, read(a), read(b), read(c))
	case 5:
		result = Gia_ManAppendBuf(dst, read(a))
	default:
		panic("gia: invalid limited cone node")
	}
	obj.Value = uint32(result)
}
func Gia_ManDupAndConesLimit_rec(dst, src *Gia_Man_t, id, level int32) {
	giaDupLimitedCone(dst, src, id, level, false)
}
func Gia_ManDupAndConesLimit2_rec(dst, src *Gia_Man_t, id, level int32) {
	giaDupLimitedCone(dst, src, id, level, true)
}
func giaDupAndConesLimit(src *Gia_Man_t, ids []int32, count, level int32, relative bool) *Gia_Man_t {
	if count < 0 || int(count) > len(ids) {
		panic("gia: invalid limited cone roots")
	}
	if !relative {
		Gia_ManLevelNum(src)
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	out := Gia_ManStart(1000)
	out.CopyIdentityFrom(src)
	for _, id := range ids[:count] {
		giaDupLimitedCone(out, src, id, level, relative)
	}
	for _, id := range ids[:count] {
		Gia_ManAppendCo(out, int32(Gia_ManObj(src, id).Value))
	}
	return out
}
func Gia_ManDupAndConesLimit(src *Gia_Man_t, ids []int32, count, level int32) *Gia_Man_t {
	return giaDupAndConesLimit(src, ids, count, level, false)
}
func Gia_ManDupAndConesLimit2(src *Gia_Man_t, ids []int32, count, level int32) *Gia_Man_t {
	return giaDupAndConesLimit(src, ids, count, level, true)
}
