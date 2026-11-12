package hop

// These manager-owned pointer vectors preserve Hop's int32 interface while
// keeping all stored pointers visible to the garbage collector.
type Vec_Ptr_t struct {
	nSize  int32
	pArray []any
}

func Vec_PtrAlloc(n int32) *Vec_Ptr_t { return &Vec_Ptr_t{pArray: make([]any, 0, n)} }
func Vec_PtrSize(v *Vec_Ptr_t) int32 {
	if v == nil {
		return 0
	}
	return v.nSize
}
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any { return v.pArray[i] }
func Vec_PtrPush(v *Vec_Ptr_t, x any) {
	if int(v.nSize) < len(v.pArray) {
		v.pArray[v.nSize] = x
	} else {
		v.pArray = append(v.pArray, x)
	}
	v.nSize++
}
func Vec_PtrFree(v *Vec_Ptr_t) {
	if v != nil {
		v.pArray = nil
		v.nSize = 0
	}
}
