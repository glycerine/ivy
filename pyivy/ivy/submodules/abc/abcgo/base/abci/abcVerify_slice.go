package abci

import (
	"unsafe"

	"github.com/glycerine/abcgo/base/abc"
)

// Abc_NtkVerifySimulatePatternSlice is the slice-safe counterpart of the
// C-shaped pointer API. model is in ordered CI space; the returned values are
// ordered COs. The backing model remains live for the duration of simulation.
func Abc_NtkVerifySimulatePatternSlice(n *Abc_Ntk_t, model []int32) []int32 {
	if n == nil {
		return nil
	}
	ciCount := int(abc.Abc_NtkCiNum(n))
	if len(model) < ciCount {
		return nil
	}
	var modelPtr *int32
	if ciCount != 0 {
		modelPtr = &model[0]
	}
	valuesPtr := Abc_NtkVerifySimulatePattern(n, modelPtr)
	if valuesPtr == nil {
		return nil
	}
	coCount := int(abc.Abc_NtkCoNum(n))
	if coCount == 0 {
		return nil
	}
	values := unsafe.Slice(valuesPtr, coCount)
	return append([]int32(nil), values...)
}
