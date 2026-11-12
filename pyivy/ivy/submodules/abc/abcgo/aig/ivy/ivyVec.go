package ivy

import "github.com/glycerine/abcgo/misc/vec"

// The source-derived Ivy routines use ABC's int-sized vector API. These
// package-local adapters retain those entry-point signatures while the
// backing storage is the shared, slice-based Go implementation.
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t

func Vec_IntAlloc(n int32) *Vec_Int_t          { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntArray(v *Vec_Int_t) []int32        { return vec.Vec_IntArray(v) }
func Vec_IntClear(v *Vec_Int_t)                { vec.Vec_IntClear(v) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32 { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntFree(v *Vec_Int_t)                 { vec.Vec_IntFree(v) }
func Vec_IntGrow(v *Vec_Int_t, n int32)        { vec.Vec_IntGrow(v, int(n)) }
func Vec_IntPop(v *Vec_Int_t) int32            { return vec.Vec_IntPop(v) }
func Vec_IntPush(v *Vec_Int_t, x int32)        { vec.Vec_IntPush(v, x) }
func Vec_IntSelectSort(a []int32, n int32)     { vec.Vec_IntSelectSort(a, int(n)) }
func Vec_IntSize(v *Vec_Int_t) int32           { return int32(vec.Vec_IntSize(v)) }

func Vec_PtrAlloc(n int32) *Vec_Ptr_t { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrArray(v *Vec_Ptr_t) []any { return vec.Vec_PtrArray(v) }
func Vec_PtrClear(v *Vec_Ptr_t)       { vec.Vec_PtrClear(v) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any {
	return vec.Vec_PtrEntry(v, int(i))
}
func Vec_PtrFill(v *Vec_Ptr_t, n int32, value any) { vec.Vec_PtrFill(v, int(n), value) }
func Vec_PtrFree(v *Vec_Ptr_t)                     { vec.Vec_PtrFree(v) }
func Vec_PtrPop(v *Vec_Ptr_t) any                  { return vec.Vec_PtrPop(v) }
func Vec_PtrPush(v *Vec_Ptr_t, value any)          { vec.Vec_PtrPush(v, value) }
func Vec_PtrSize(v *Vec_Ptr_t) int32               { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int32, value any) {
	vec.Vec_PtrWriteEntry(v, int(i), value)
}
