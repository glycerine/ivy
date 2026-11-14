package gia

// Gia_ManDupMuxes recognizes the original AND decomposition in source order.
// Sibling barriers and fanout limits prevent absorbing shared choice structure.
func Gia_ManDupMuxes(src *Gia_Man_t, limit int32) *Gia_Man_t {
	if src.PMuxes != nil || limit < 0 {
		panic("gia: coarsening expects ordinary AIG and nonnegative limit")
	}
	src.Refs = nil
	Gia_ManCreateRefs(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.EnableMuxes()
	if src.pSibls != nil {
		dst.pSibls = make([]int32, src.nObjs)
	}
	src.pObjs[0].Value = 0
	Gia_ManHashStart(dst)
	for _, obj := range src.pObjs[1:] {
		var lit int32
		switch {
		case Gia_ObjIsCi(obj) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj))
		case Gia_ObjIsBuf(obj) != 0:
			lit = Gia_ManAppendBuf(dst, Gia_ObjFanin0Copy(obj))
		default:
			id := obj.id
			plain := Gia_ObjIsMuxType(obj) == 0 || Gia_ObjSibl(src, Gia_ObjFaninId0(obj, id)) != 0 || Gia_ObjSibl(src, Gia_ObjFaninId1(obj, id)) != 0
			var a, b *Gia_Obj_t
			if plain {
				lit = Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
			} else if Gia_ObjRecognizeExor(obj, &a, &b) != 0 {
				lit = Gia_ManHashXorReal(dst, Gia_ObjLitCopy(src, Gia_ObjToLit(src, a)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, b)))
			} else if Gia_ObjRefNum(src, Gia_ObjFanin0(obj))+Gia_ObjRefNum(src, Gia_ObjFanin1(obj)) > limit {
				lit = Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
			} else {
				control := Gia_ObjRecognizeMux(obj, &a, &b)
				lit = Gia_ManHashMuxReal(dst, Gia_ObjLitCopy(src, Gia_ObjToLit(src, control)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, a)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, b)))
			}
		}
		obj.Value = uint32(lit)
		if sibling := Gia_ObjSibl(src, obj.id); sibling != 0 {
			other := int32(src.pObjs[sibling].Value) >> 1
			id := lit >> 1
			if Gia_ObjIsAnd(dst.pObjs[id]) != 0 && Gia_ObjIsAnd(dst.pObjs[other]) != 0 && id > other {
				if len(dst.pSibls) <= int(id) {
					dst.pSibls = append(dst.pSibls, make([]int32, int(id)+1-len(dst.pSibls))...)
				}
				dst.pSibls[id] = other
			}
		}
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
