package cnf

import (
	"github.com/glycerine/abcgo/misc/vec"
)

func Cnf_ManWriteCnfMapping(manager *Cnf_Man_t, mapped *Vec_Ptr_t) *Vec_Int_t {
	if manager == nil || manager.pManAig == nil {
		return vec.Vec_IntAllocExact(0)
	}
	result := vec.Vec_IntStart(len(manager.pManAig.Nodes()))
	for _, item := range vec.Vec_PtrArray(mapped) {
		edge, ok := item.(Aig_Obj_t)
		if !ok || edge.NodeID() >= len(manager.selectedCuts) {
			continue
		}
		cut := manager.selectedCuts[edge.NodeID()]
		if cut == nil || len(cut.truth) == 0 || len(cut.pFanins) > 4 {
			continue
		}
		vec.Vec_IntWriteEntry(result, edge.NodeID(), int32(vec.Vec_IntSize(result)))
		vec.Vec_IntPush(result, int32(cut.truth[0]))
		vec.Vec_IntPushArray(result, cut.pFanins, len(cut.pFanins))
		for i := len(cut.pFanins); i < 4; i++ {
			vec.Vec_IntPush(result, -1)
		}
	}
	return result
}

func Cnf_SopConvertToVector(sop string, cubes int32, cover *Vec_Int_t) {
	if cover == nil {
		return
	}
	vec.Vec_IntClear(cover)
	for i := int32(0); i < cubes; i++ {
		cube := int(sop[i])
		var packed int32
		for b := 0; b < 4; b++ {
			var state int32
			if cube%3 == 0 {
				state = 1
			} else if cube%3 == 1 {
				state = 2
			}
			packed = packed<<2 | state
			cube /= 3
		}
		vec.Vec_IntPush(cover, packed)
	}
}
func Cnf_SopCountLiterals(sop string, cubes int32) int32 {
	var count int32
	for i := int32(0); i < cubes; i++ {
		cube := int(sop[i])
		for b := 0; b < 4; b++ {
			if cube%3 != 2 {
				count++
			}
			cube /= 3
		}
	}
	return count
}
func Cnf_IsopCountLiterals(isop *Vec_Int_t, variables int32) int32 {
	if isop == nil {
		return 0
	}
	var count int32
	for _, cube := range vec.Vec_IntArray(isop) {
		for i := int32(0); i < variables; i++ {
			if cube&3 == 1 || cube&3 == 2 {
				count++
			}
			cube >>= 2
		}
	}
	return count
}
func Cnf_IsopWriteCube(cube int32, variables int32, vars, literals []int32) int32 {
	if variables < 0 || variables > 16 || vars == nil || literals == nil {
		return 0
	}
	count := int32(0)
	for i := int32(0); i < variables; i++ {
		state := cube & 3
		cube >>= 2
		if state != 1 && state != 2 {
			continue
		}
		literals[count] = vars[i]*2 + state - 1
		count++
	}
	return count
}

func Cnf_ManWriteCnf(manager *Cnf_Man_t, _ *Vec_Ptr_t, outputs int32) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	return deriveLegacyAIG(manager.pManAig, outputs)
}
func Cnf_ManWriteCnfOther(manager *Cnf_Man_t, _ *Vec_Ptr_t) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	result, err := EncodeAIG(manager.pManAig, 0)
	if err != nil {
		return nil
	}
	return result
}
func Cnf_DeriveSimple(graph *Aig_Man_t, outputs int32) *Cnf_Dat_t {
	return deriveLegacyAIG(graph, outputs)
}
func Cnf_DeriveSimpleForRetiming(graph *Aig_Man_t) *Cnf_Dat_t {
	result, err := EncodeAIG(graph, 0)
	if err != nil {
		return nil
	}
	return result
}
