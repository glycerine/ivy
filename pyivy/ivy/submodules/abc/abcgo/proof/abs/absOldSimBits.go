package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Simulation entries store sixteen two-bit frame values per uint32 word.
func oldSimGet(info *vec.Vec_Ptr_t, object *aig.Edge, frame int32) int32 {
	words := vec.Vec_PtrEntry(info, (*object).NodeID()).([]uint32)
	return int32((words[frame>>4] >> uint((frame&15)<<1)) & 3)
}

func oldSimSet(info *vec.Vec_Ptr_t, object *aig.Edge, frame, value int32) {
	words := vec.Vec_PtrEntry(info, (*object).NodeID()).([]uint32)
	shift := uint((frame & 15) << 1)
	words[frame>>4] = (words[frame>>4] &^ (uint32(3) << shift)) | uint32(value)<<shift
}

func Saig_ManSimInfoGet(info *vec.Vec_Ptr_t, object *aig.Edge, frame int32) int32 {
	return oldSimGet(info, object, frame)
}

func Saig_ManSimInfoSet(info *vec.Vec_Ptr_t, object *aig.Edge, frame, value int32) {
	if value < SAIG_ZER || value > SAIG_UND {
		panic("abstraction: invalid ternary simulation value")
	}
	oldSimSet(info, object, frame, value)
}

func Saig_ManSimInfo2Get(info *vec.Vec_Ptr_t, object *aig.Edge, frame int32) int32 {
	return oldSimGet(info, object, frame)
}

func Saig_ManSimInfo2Set(info *vec.Vec_Ptr_t, object *aig.Edge, frame, value int32) {
	oldSimSet(info, object, frame, value)
}
