package res

import "github.com/glycerine/abcgo/misc/vec"

// These thin adapters preserve the integer-width contracts used by the
// translated window code while delegating all storage to the shared vec API.
func Vec_PtrAlloc(size int32) *Vec_Ptr_t { return vec.Vec_PtrAlloc(int(size)) }
func Vec_PtrAllocSimInfo(entries, words int32) *Vec_Ptr_t {
	return vec.Vec_PtrAllocSimInfo(int(entries), int(words))
}
func Vec_PtrClear(p *Vec_Ptr_t)           { vec.Vec_PtrClear(p) }
func Vec_PtrFree(p *Vec_Ptr_t)            { vec.Vec_PtrFree(p) }
func Vec_PtrPush(p *Vec_Ptr_t, value any) { vec.Vec_PtrPush(p, value) }
func Vec_PtrPushUnique(p *Vec_Ptr_t, value any) int32 {
	return int32(vec.Vec_PtrPushUnique(p, value))
}
func Vec_PtrSize(p *Vec_Ptr_t) int32 { return int32(vec.Vec_PtrSize(p)) }
func Vec_PtrFind(p *Vec_Ptr_t, value any) int32 {
	return int32(vec.Vec_PtrFind(p, value))
}
func Vec_PtrEntry(p *Vec_Ptr_t, index any) *Abc_Obj_t {
	value := vec.Vec_PtrEntry(p, resVectorIndex(index))
	if value == nil {
		return nil
	}
	object, ok := value.(*Abc_Obj_t)
	if !ok {
		panic("res: expected ABC object in pointer vector")
	}
	return object
}
func Vec_PtrEntryAny(p *Vec_Ptr_t, index any) any {
	return vec.Vec_PtrEntry(p, resVectorIndex(index))
}

func Vec_VecStart(size int32) *Vec_Vec_t { return vec.Vec_VecStart(int(size)) }
func Vec_VecClear(p *Vec_Vec_t)          { vec.Vec_VecClear(p) }
func Vec_VecFree(p *Vec_Vec_t)           { vec.Vec_VecFree(p) }
func Vec_VecSize(p *Vec_Vec_t) int32     { return int32(vec.Vec_VecSize(p)) }
func Vec_VecEntry(p *Vec_Vec_t, level int32) *Vec_Ptr_t {
	return vec.Vec_VecEntry(p, int(level))
}
func Vec_VecPush(p *Vec_Vec_t, level int32, value any) {
	vec.Vec_VecPush(p, int(level), value)
}

func Vec_IntAlloc(size int32) *Vec_Int_t { return vec.Vec_IntAlloc(int(size)) }
func Vec_IntFree(p *Vec_Int_t)           { vec.Vec_IntFree(p) }

func resVectorIndex(index any) int {
	switch value := index.(type) {
	case int:
		return value
	case int32:
		return int(value)
	default:
		panic("res: unsupported vector index type")
	}
}
