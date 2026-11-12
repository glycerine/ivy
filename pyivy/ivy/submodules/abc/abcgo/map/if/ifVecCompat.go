package if_

import "github.com/glycerine/abcgo/misc/vec"

// These adapters preserve the C package's package-level vector vocabulary
// while routing storage to the shared Go vector implementation.
func Vec_IntAlloc(n int) *Vec_Int_t          { return vec.Vec_IntAlloc(n) }
func Vec_IntPush(v *Vec_Int_t, value int32)  { vec.Vec_IntPush(v, value) }
func Vec_IntSize(v *Vec_Int_t) int           { return vec.Vec_IntSize(v) }
func Vec_IntEntry(v *Vec_Int_t, i int) int32 { return vec.Vec_IntEntry(v, i) }
func Vec_IntCountUnique(v *Vec_Int_t) int    { return vec.Vec_IntCountUnique(v) }
func Vec_IntFree(v *Vec_Int_t)               { vec.Vec_IntFree(v) }
func Vec_PtrAlloc(n int) *Vec_Ptr_t          { return vec.Vec_PtrAlloc(n) }
func Vec_PtrPush(v *Vec_Ptr_t, value any)    { vec.Vec_PtrPush(v, value) }
func Vec_PtrClear(v *Vec_Ptr_t)              { vec.Vec_PtrClear(v) }
func Vec_PtrFree(v *Vec_Ptr_t)               { vec.Vec_PtrFree(v) }
func Hsh_IntManHashArray(v *Vec_Int_t, size int) *Vec_Int_t {
	return vec.Hsh_IntManHashArray(v, size)
}
