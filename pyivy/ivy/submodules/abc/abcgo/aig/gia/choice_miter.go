package gia

// Gia_ManChoiceMiter_rec copies a cone using the source object's Value mapping.
// Inputs and constant zero must already have destination literals assigned,
// and the destination structural hash must be allocated, as in C.
func Gia_ManChoiceMiter_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	id := Gia_ObjId(src, obj)
	a, b, c, kind := src.Gate(id)
	copyLit := func(lit int32) int32 { return Gia_ManChoiceMiter_rec(dst, src, Gia_ManObj(src, lit>>1)) ^ (lit & 1) }
	var result int32
	switch kind {
	case 2:
		result = Gia_ManHashAnd(dst, copyLit(a), copyLit(b))
	case 3:
		result = Gia_ManHashXor(dst, copyLit(a), copyLit(b))
	case 4:
		result = Gia_ManHashMux(dst, copyLit(a), copyLit(b), copyLit(c))
	case 5:
		result = copyLit(a)
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
// C's recursive copier assumes ordinary AND graphs. Real XOR/MUX nodes are
// expanded through the existing hashed AND constructors here; interpreting
// their two stored data pins as an AND silently changes their functions.
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
	Gia_ManHashAlloc(dst)
	for i := int32(0); i < Gia_ManCoNum(first); i++ {
		for _, src := range networks {
			Gia_ManChoiceMiter_rec(dst, src, Gia_ManCo(src, i))
		}
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, first.nRegs)
	return dst
}
