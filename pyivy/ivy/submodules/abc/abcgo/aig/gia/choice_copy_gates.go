package gia

// Copy the logical gate after recursively mapping its fanins. C's choice
// converters assume ordinary AND graphs; real gates use their existing AND
// expansions here, preserving the surrounding representative/choice policy.
func giaCopyChoiceGate(dst, src *Gia_Man_t, obj *Gia_Obj_t, visit func(*Gia_Man_t, *Gia_Man_t, *Gia_Obj_t)) int32 {
	a, b, c, kind := src.Gate(obj.id)
	copyLit := func(lit int32) int32 {
		visit(dst, src, Gia_ManObj(src, lit>>1))
		return Gia_ObjLitCopy(src, lit)
	}
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
	default:
		panic("gia: choice object must be an internal gate")
	}
	giaGrowChoiceTables(dst)
	return result
}

// Expanding real gates can exceed the source object's count. Every auxiliary
// AND starts without a representative; zero would falsely mean constant zero.
func giaGrowChoiceTables(p *Gia_Man_t) {
	for len(p.PReprs) < int(p.nObjs) {
		p.PReprs = append(p.PReprs, Gia_Rpr_t{iRepr: GIA_VOID})
		p.PNexts = append(p.PNexts, 0)
	}
}
