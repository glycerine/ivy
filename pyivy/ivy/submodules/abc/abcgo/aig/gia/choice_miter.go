package gia

// Gia_ManChoiceMiter_rec copies a cone using the source object's Value mapping.
// Inputs and constant zero must already have destination literals assigned.
func Gia_ManChoiceMiter_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	id := Gia_ObjId(src, obj)
	_, _, _, kind := src.Gate(id)
	copyLit := func(lit int32) int32 { return Gia_ManChoiceMiter_rec(dst, src, Gia_ManObj(src, lit>>1)) ^ (lit & 1) }
	var result int32
	switch kind {
	case 2, 3, 4, 5:
		a, b := Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)
		result = dst.AndLiteral(copyLit(a), copyLit(b), true)
	case 6:
		result = Gia_ManAppendCo(dst, copyLit(Gia_ObjFaninLit0(obj, id)))
	default:
		panic("gia: uninitialized choice-miter input or invalid node")
	}
	obj.Value = uint32(result)
	return result
}

// Gia_ManChoiceMiter shares inputs and interleaves each network's outputs by
// output index, matching the source choice-computation convention. It overwrites
// the source Value scratch fields, as the C implementation does.
func Gia_ManChoiceMiter(networks []*Gia_Man_t) *Gia_Man_t {
	if len(networks) == 0 {
		panic("gia: empty choice miter")
	}
	first := networks[0]
	for _, src := range networks {
		if Gia_ManCiNum(src) != Gia_ManCiNum(first) || Gia_ManCoNum(src) != Gia_ManCoNum(first) || src.nRegs != first.nRegs {
			panic("gia: incompatible choice-miter interfaces")
		}
	}
	dst := Gia_ManStart(int32(len(networks)) * Gia_ManObjNum(first))
	dst.pName, dst.PSpec = first.pName, first.PSpec
	for _, src := range networks {
		for _, obj := range src.pObjs {
			obj.Value = ^uint32(0)
		}
		src.pObjs[0].Value = 0
	}
	for i := int32(0); i < Gia_ManCiNum(first); i++ {
		literal := Gia_ManAppendCi(dst)
		for _, src := range networks {
			Gia_ManCi(src, i).Value = uint32(literal)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(first); i++ {
		for _, src := range networks {
			Gia_ManChoiceMiter_rec(dst, src, Gia_ManCo(src, i))
		}
	}
	dst.literalHash = nil
	Gia_ManSetRegNum(dst, first.nRegs)
	return dst
}
