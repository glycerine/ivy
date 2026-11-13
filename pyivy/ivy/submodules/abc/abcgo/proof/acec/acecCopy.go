package acec

import "github.com/glycerine/abcgo/aig/gia"

// acecCopyLogic copies a node after all of its fanins have been mapped.
// Structural expansion for MUXes keeps the destination independent of mux data.
func acecCopyLogic(dst, src *Gia_Man_t, o *Gia_Obj_t) int32 {
	a := gia.Gia_ObjFanin0Copy(o)
	if gia.Gia_ObjIsBuf(o) != 0 {
		return a
	}
	b := gia.Gia_ObjFanin1Copy(o)
	if gia.Gia_ObjIsMux(src, o) != 0 {
		return gia.Gia_ManAppendMux2(dst, gia.Gia_ObjFanin2Copy(src, o), b, a)
	}
	if gia.Gia_ObjIsXor(o) != 0 {
		return gia.Gia_ManAppendXorReal2(dst, a, b)
	}
	return gia.Gia_ManAppendAnd2(dst, a, b)
}

// Arithmetic cut enumeration assumes two-input AND/XOR nodes. Fall back to
// ordinary CEC or graph duplication when native MUXes or buffers are present.
func acecCanDetectArithmetic(p *Gia_Man_t) bool {
	if gia.Gia_ManMuxNum(p) != 0 {
		return false
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsBuf(gia.Gia_ManObj(p, id)) != 0 {
			return false
		}
	}
	return true
}
