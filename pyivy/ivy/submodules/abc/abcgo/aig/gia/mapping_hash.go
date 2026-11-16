package gia

// Gia_ManDupHashMapping structurally hashes an AIG and copies LUT records
// through the old-to-new node map. Merged roots retain the last source cut,
// as in C, with one compact record per destination root. Hash-collapsed cuts
// must still have leaves preceding their root.
func Gia_ManDupHashMapping(src *Gia_Man_t) *Gia_Man_t {
	if Gia_ManHasMapping(src) == 0 {
		panic("gia: mapping is absent")
	}
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.CopyIdentityFrom(src)
	Gia_ManHashAlloc(dst)
	Gia_ManFillValue(src)
	Gia_ManConst0(src).Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for id := int32(1); id < Gia_ManObjNum(src); id++ {
		obj := Gia_ManObj(src, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		// The C routine assumes an ordinary AND graph. Normalize real gates
		// through the existing hash operations before transferring LUT cuts.
		a, b, c, kind := src.Gate(id)
		copyLit := func(lit int32) int32 { return Gia_ObjLitCopy(src, lit) }
		var literal int32
		switch kind {
		case 2:
			literal = Gia_ManHashAnd(dst, copyLit(a), copyLit(b))
		case 3:
			literal = Gia_ManHashXor(dst, copyLit(a), copyLit(b))
		case 4:
			literal = Gia_ManHashMux(dst, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			literal = copyLit(a)
		default:
			panic("gia: invalid mapped duplication gate")
		}
		obj.Value = uint32(literal)
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(Gia_ManCo(src, i)))
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	mapping := Vec_IntStart(int(Gia_ManObjNum(dst)))
	// Temporarily use the destination index to select the last source LUT for
	// each merged root. This retains C's overwrite policy without its dead records.
	for id := int32(1); id < Gia_ManObjNum(src); id++ {
		if Gia_ObjIsLut(src, id) == 0 {
			continue
		}
		copied := int32(Gia_ManObj(src, id).Value >> 1)
		if copied == 0 || Gia_ObjIsCi(Gia_ManObj(dst, copied)) != 0 {
			continue
		}
		// C can overwrite a merged root with a self/forward cut after hashing.
		// Keep the last usable source cut, retaining the topological invariant.
		usable := true
		for _, fanin := range Gia_ObjLutFanins(src, id) {
			if int32(Gia_ManObj(src, fanin).Value>>1) >= copied {
				usable = false
				break
			}
		}
		if usable {
			Vec_IntWriteEntry(mapping, int(copied), id)
		}
	}
	for copied := int32(1); copied < Gia_ManObjNum(dst); copied++ {
		sourceID := Vec_IntEntry(mapping, int(copied))
		if sourceID == 0 {
			continue
		}
		Vec_IntWriteEntry(mapping, int(copied), int32(Vec_IntSize(mapping)))
		fanins := Gia_ObjLutFanins(src, sourceID)
		Vec_IntPush(mapping, int32(len(fanins)))
		for _, fanin := range fanins {
			Vec_IntPush(mapping, int32(Gia_ManObj(src, fanin).Value>>1))
		}
		Vec_IntPush(mapping, copied)
	}

	dst.VMapping = mapping
	return dst
}
