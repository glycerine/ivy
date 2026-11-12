package gia

// Gia_ManCleanupRemap composes the source-to-intermediate copy literals with
// the intermediate-to-result literals produced by duplication or cleanup.
func Gia_ManCleanupRemap(src, intermediate *Gia_Man_t) {
	for _, obj := range src.pObjs[1:] {
		old := obj.Value
		if old == ^uint32(0) {
			continue
		}
		mapped := Gia_ManObj(intermediate, int32(old>>1)).Value
		if mapped == ^uint32(0) {
			obj.Value = mapped
		} else {
			obj.Value = mapped ^ (old & 1)
		}
	}
}

// Gia_ManDupMoveLast moves the last count inputs to position insert.
func Gia_ManDupMoveLast(src *Gia_Man_t, insert, count int32) *Gia_Man_t {
	n := Gia_ManCiNum(src)
	if count < 0 || insert < 0 || count > n || insert > n-count {
		panic("gia: invalid input move")
	}
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManConst0(src).Value = 0
	for _, interval := range [][2]int32{{0, insert}, {n - count, n}, {insert, n - count}} {
		for i := interval[0]; i < interval[1]; i++ {
			Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
		}
	}
	for _, obj := range src.pObjs[1:] {
		var lit int32
		switch {
		case Gia_ObjIsCi(obj) != 0:
			continue
		case Gia_ObjIsAnd(obj) != 0:
			lit = Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj))
		default:
			panic("gia: unsupported input move object")
		}
		obj.Value = uint32(lit)
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
