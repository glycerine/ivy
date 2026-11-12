package gia

func giaFaninCopyRepr(p *Gia_Man_t, obj *Gia_Obj_t, reps []int32, fanin int) int32 {
	id, complement := Gia_ObjFaninId0p(p, obj), Gia_ObjFaninC0(obj)
	if fanin == 1 {
		id, complement = Gia_ObjFaninId1p(p, obj), Gia_ObjFaninC1(obj)
	}
	repr := reps[id]
	if repr == -1 {
		return Gia_ObjValue(Gia_ManObj(p, id)) ^ complement
	}
	if repr>>1 >= Gia_ObjId(p, obj) {
		panic("gia: representative is not earlier than object")
	}
	return Gia_ObjValue(Gia_ManObj(p, repr>>1)) ^ complement ^ (repr & 1)
}
func Gia_ObjFanin0CopyRepr(p *Gia_Man_t, obj *Gia_Obj_t, reps []int32) int32 {
	return giaFaninCopyRepr(p, obj, reps, 0)
}
func Gia_ObjFanin1CopyRepr(p *Gia_Man_t, obj *Gia_Obj_t, reps []int32) int32 {
	return giaFaninCopyRepr(p, obj, reps, 1)
}

// Gia_ManFraigReduceGia replaces fanins with their representative literals.
func Gia_ManFraigReduceGia(src *Gia_Man_t, reps []int32) *Gia_Man_t {
	if len(reps) < int(Gia_ManObjNum(src)) || Gia_ManRegNum(src) != 0 {
		panic("gia: invalid FRAIG representatives")
	}
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManFillValue(src)
	Gia_ManHashAlloc(dst)
	for _, obj := range src.pObjs {
		var lit int32
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			lit = Gia_ManHashAnd(dst, Gia_ObjFanin0CopyRepr(src, obj, reps), Gia_ObjFanin1CopyRepr(src, obj, reps))
		case Gia_ObjIsCi(obj) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0CopyRepr(src, obj, reps))
		case Gia_ObjIsConst0(obj) != 0:
			lit = 0
		default:
			panic("gia: unsupported FRAIG object")
		}
		obj.Value = uint32(lit)
	}
	Gia_ManHashStop(dst)
	return dst
}
