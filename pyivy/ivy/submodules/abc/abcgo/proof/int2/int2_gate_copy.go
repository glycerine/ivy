package int2

import "github.com/glycerine/abcgo/aig/gia"

// The C helpers copy every Gia_ObjIsAnd node using its first two fanins,
// including native XOR and MUX nodes; this does not preserve their semantics.
func int2HashCopyGate(dst, src *gia.Gia_Man_t, id int32) int32 {
	obj := gia.Gia_ManObj(src, id)
	return gia.Gia_ManHashAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
}
