package cnf

import (
	"math/bits"

	"github.com/glycerine/abcgo/misc/vec"
)

func Cnf_ManWriteCnfMapping(manager *Cnf_Man_t, mapped *Vec_Ptr_t) *Vec_Int_t {
	if manager == nil || manager.pManAig == nil {
		return vec.Vec_IntAllocExact(0)
	}
	result := vec.Vec_IntStart(aigObjectCount(manager.pManAig))
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
		truth := uint32(smallCutTruth(cut.truth[0], len(cut.pFanins)))
		vec.Vec_IntPush(result, int32(truth|truth<<16))
		vec.Vec_IntPushArray(result, cut.pFanins, len(cut.pFanins))
		for i := len(cut.pFanins); i < 4; i++ {
			vec.Vec_IntPush(result, -1)
		}
	}
	return result
}

// ABC stores four ternary digits per SOP byte, with the first variable in
// the most significant digit. ISOP instead places that variable in bits 0–1.
type sopCube struct {
	packed   byte
	literals [5]byte
}

var sopCubes = func() [256]sopCube {
	var table [256]sopCube
	for code := range table {
		cube := code
		var packed byte
		for range 4 {
			packed = packed<<2 | byte((cube%3+1)%3)
			cube /= 3
		}
		entry := &table[code]
		entry.packed = packed
		for i := range 4 {
			entry.literals[i+1] = entry.literals[i]
			if packed>>(2*i)&3 != 0 {
				entry.literals[i+1]++
			}
		}
	}
	return table
}()

func Cnf_SopConvertToVector(sop string, cubes int32, cover *Vec_Int_t) {
	if cover == nil {
		return
	}
	vec.Vec_IntClear(cover)
	for i := int32(0); i < cubes; i++ {
		vec.Vec_IntPush(cover, int32(sopCubes[sop[i]].packed))
	}
}
func Cnf_SopCountLiterals(sop string, cubes int32) int32 {
	var count int32
	for i := int32(0); i < cubes; i++ {
		count += int32(sopCubes[sop[i]].literals[4])
	}
	return count
}
func Cnf_IsopCountLiterals(isop *Vec_Int_t, variables int32) int32 {
	if isop == nil || variables <= 0 {
		return 0
	}
	var count int32
	mask := uint32(0x55555555) & (uint32(1)<<uint(2*min(variables, 16)) - 1)
	for _, cube := range vec.Vec_IntArray(isop) {
		// The pair encodes a literal exactly when its two bits differ.
		u := uint32(cube)
		count += int32(bits.OnesCount32((u ^ u>>1) & mask))
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

func Cnf_ManWriteCnf(manager *Cnf_Man_t, mapped *Vec_Ptr_t, outputs int32) *Cnf_Dat_t {
	if manager != nil && manager.pManAig != nil {
		graph := manager.pManAig
		count := int32(graph.OutputCount())
		if outputs < 0 || outputs > count || outputs != 0 && outputs != count && (graph.RegisterCount() == 0 || outputs != graph.RegisterCount()) {
			panic("cnf: invalid free output count")
		}
	}
	result, err := cnfWriteMapped(manager, mapped, outputs)
	if err != nil {
		return nil
	}
	return result
}
func Cnf_ManWriteCnfOther(manager *Cnf_Man_t, mapped *Vec_Ptr_t) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	result, err := writeOtherMapped(manager, mapped)
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
