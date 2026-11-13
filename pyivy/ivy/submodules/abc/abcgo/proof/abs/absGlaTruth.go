package abs

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ObjTruthDepends identifies truth-table pairs differing in variable v.
// The truth word uses the repeated five-variable representation from ABC.
func Ga2_ObjTruthDepends(truth uint32, v int32) uint32 {
	if v < 0 || v >= 5 {
		panic("abstraction: truth variable out of range")
	}
	masks := [5]uint32{0x55555555, 0x33333333, 0x0f0f0f0f, 0x00ff00ff, 0x0000ffff}
	return (truth ^ (truth >> (1 << uint(v)))) & masks[v]
}

// Ga2_ManCnfCompute derives the source algorithm's ISOP cover. The caller's
// scratch vector is populated, and the returned vector owns an independent copy.
func Ga2_ManCnfCompute(truth uint32, variables int32, scratch *vec.Vec_Int_t) *vec.Vec_Int_t {
	if variables < 0 || variables > 5 || scratch == nil {
		panic("abstraction: invalid ISOP arguments")
	}
	var cover []uint32
	if kit.Kit_TruthIsop([]uint32{truth}, variables, &cover, 0) != 0 {
		panic("abstraction: ISOP computation failed")
	}
	vec.Vec_IntClear(scratch)
	for _, cube := range cover {
		vec.Vec_IntPush(scratch, int32(cube))
	}
	return vec.Vec_IntDup(scratch)
}
