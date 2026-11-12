package hop

import "github.com/glycerine/abcgo/misc/vec"

// Shared vector aliases keep Hop results directly interoperable with other ABC
// packages. Thin index adapters preserve Hop's original int32 parameter API.
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Vec_t = vec.Vec_Vec_t

func Vec_PtrAlloc(n int32) *Vec_Ptr_t { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32 {
	if v == nil {
		return 0
	}
	return int32(vec.Vec_PtrSize(v))
}
func Vec_PtrArray(v *Vec_Ptr_t) []any                   { return vec.Vec_PtrArray(v) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any            { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                   { vec.Vec_PtrPush(v, x) }
func Vec_PtrClear(v *Vec_Ptr_t)                         { vec.Vec_PtrClear(v) }
func Vec_PtrShrink(v *Vec_Ptr_t, n int32)               { vec.Vec_PtrShrink(v, int(n)) }
func Vec_PtrFree(v *Vec_Ptr_t)                          { vec.Vec_PtrFree(v) }
func Vec_VecAlloc(n int32) *Vec_Vec_t                   { return vec.Vec_VecAlloc(int(n)) }
func Vec_VecSize(v *Vec_Vec_t) int32                    { return int32(vec.Vec_VecSize(v)) }
func Vec_VecExpand(v *Vec_Vec_t, level int32)           { vec.Vec_VecExpand(v, int(level)) }
func Vec_VecEntry(v *Vec_Vec_t, level int32) *Vec_Ptr_t { return vec.Vec_VecEntry(v, int(level)) }
func Vec_VecPush(v *Vec_Vec_t, level int32, x any)      { vec.Vec_VecPush(v, int(level), x) }
func Vec_VecFree(v *Vec_Vec_t)                          { vec.Vec_VecFree(v) }
