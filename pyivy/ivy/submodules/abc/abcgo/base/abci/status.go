package abci

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_FrameDeriveStatusArray reports 0 for a witness and -1 for no result.
func Abc_FrameDeriveStatusArray(cexes *vec.Vec_Ptr_t) *vec.Vec_Int_t {
	if cexes == nil {
		return nil
	}
	result := vec.Vec_IntStart(vec.Vec_PtrSize(cexes))
	for i, value := range vec.Vec_PtrArray(cexes) {
		status := int32(-1)
		if value != nil {
			if cex, ok := value.(*util.Abc_Cex_t); !ok || cex != nil {
				status = 0
			}
		}
		vec.Vec_IntWriteEntry(result, i, status)
	}
	return result
}

// isCexSentinel reports C's pointer value (Abc_Cex_t *)1 in a pointer
// vector: util.AbcCexSentinel, or uintptr(1) where a vector holds it bare.
func isCexSentinel(value any) bool {
	switch v := value.(type) {
	case uintptr:
		return v == 1
	case *util.Abc_Cex_t:
		return v == util.AbcCexSentinel
	}
	return false
}

// Abc_FrameDeriveStatusArray2 also consumes the source's pointer-one UNSAT
// sentinel (isCexSentinel).
func Abc_FrameDeriveStatusArray2(cexes *vec.Vec_Ptr_t) *vec.Vec_Int_t {
	result := Abc_FrameDeriveStatusArray(cexes)
	if result == nil {
		return nil
	}
	for i, value := range vec.Vec_PtrArray(cexes) {
		if isCexSentinel(value) {
			vec.Vec_IntWriteEntry(result, i, 1)
			vec.Vec_PtrWriteEntry(cexes, i, nil)
		}
	}
	return result
}
