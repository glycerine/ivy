package cnf

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

func Cnf_ManWriteCnfMapping(manager *Cnf_Man_t, _ *Vec_Ptr_t) *Vec_Int_t {
	result := vec.Vec_IntAllocExact(0)
	if manager == nil {
		return result
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	for id, cut := range manager.selectedCuts {
		if cut == nil {
			continue
		}
		vec.Vec_IntPush(result, int32(id))
		vec.Vec_IntPush(result, int32(len(cut.pFanins)))
		vec.Vec_IntPushArray(result, cut.pFanins, len(cut.pFanins))
		if len(cut.truth) > 0 {
			vec.Vec_IntPush(result, int32(cut.truth[0]))
		}
	}
	return result
}

func Cnf_SopConvertToVector(sop string, cubes int32, cover *Vec_Int_t) {
	if cover == nil {
		return
	}
	vec.Vec_IntClear(cover)
	for _, cube := range strings.Fields(sop) {
		if len(cube) > 0 {
			vec.Vec_IntPush(cover, int32(len(cube)))
		}
	}
	_ = cubes
}
func Cnf_SopCountLiterals(sop string, cubes int32) int32 {
	var count int32
	for _, cube := range strings.Fields(sop) {
		count += int32(len(cube))
	}
	_ = cubes
	return count
}
func Cnf_IsopCountLiterals(isop *Vec_Int_t, _ int32) int32 {
	if isop == nil {
		return 0
	}
	return int32(vec.Vec_IntSize(isop))
}
func Cnf_IsopWriteCube(cube int32, variables int32, vars, literals []int32) int32 {
	if variables < 0 || variables > 16 || vars == nil || literals == nil {
		return 0
	}
	count := int32(0)
	for i := int32(0); i < variables; i++ {
		state := cube & 3
		cube >>= 2
		if state == 0 {
			continue
		}
		vars[count] = i
		literals[count] = state & 1
		count++
	}
	return count
}

func Cnf_ManWriteCnf(manager *Cnf_Man_t, _ *Vec_Ptr_t, outputs int32) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	result, err := EncodeAIG(manager.pManAig, int(outputs))
	if err != nil {
		return nil
	}
	return result
}
func Cnf_ManWriteCnfOther(manager *Cnf_Man_t, _ *Vec_Ptr_t) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	result, err := EncodeAIG(manager.pManAig, -1)
	if err != nil {
		return nil
	}
	return result
}
func Cnf_DeriveSimple(graph *Aig_Man_t, outputs int32) *Cnf_Dat_t {
	result, err := EncodeAIG(graph, int(outputs))
	if err != nil {
		return nil
	}
	return result
}
func Cnf_DeriveSimpleForRetiming(graph *Aig_Man_t) *Cnf_Dat_t {
	result, err := EncodeAIG(graph, 0)
	if err != nil {
		return nil
	}
	return result
}
