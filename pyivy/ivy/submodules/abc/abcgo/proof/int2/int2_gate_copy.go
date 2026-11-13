package int2

import "github.com/glycerine/abcgo/aig/gia"

// int2HashCopyGate lowers source native gates into the destination's AND
// graph. Source Value fields contain the already-copied fanin literals.
func int2HashCopyGate(dst, src *gia.Gia_Man_t, id int32) int32 {
	a, b, c, kind := src.Gate(id)
	copyLit := func(lit int32) int32 {
		return gia.Gia_ObjValue(gia.Gia_ManObj(src, lit>>1)) ^ (lit & 1)
	}
	switch kind {
	case 2:
		return gia.Gia_ManHashAnd(dst, copyLit(a), copyLit(b))
	case 3:
		return gia.Gia_ManHashXor(dst, copyLit(a), copyLit(b))
	case 4:
		return gia.Gia_ManHashMux(dst, copyLit(a), copyLit(b), copyLit(c))
	case 5:
		return copyLit(a)
	default:
		panic("int2: unsupported gate in graph copy")
	}
}
