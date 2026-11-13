package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

// Cnf_CutAlloc creates an owned cut with enough words for its truth table.
func Cnf_CutAlloc(_ *Cnf_Man_t, nLeaves int32) *Cnf_Cut_t {
	if nLeaves < 0 || nLeaves > 16 {
		panic("cnf: cut leaf count out of range")
	}
	words := 1
	if nLeaves > 5 {
		words = 1 << (nLeaves - 5)
	}
	return &Cnf_Cut_t{nFanins: byte(nLeaves), nWords: int16(words), pFanins: make([]int32, nLeaves), truth: make([]uint32, words)}
}

func Cnf_CutFree(cut *Cnf_Cut_t) {
	if cut == nil {
		return
	}
	for i, vector := range cut.vIsop {
		if vector != nil {
			vec.Vec_IntFree(vector)
			cut.vIsop[i] = nil
		}
	}
	cut.pFanins = nil
	cut.truth = nil
}

// Cnf_CutFromTruth constructs a cut from a sorted list of AIG object IDs and
// packed little-endian truth words. It copies both inputs.
func Cnf_CutFromTruth(leaves []int32, truth []uint32) *Cnf_Cut_t {
	cut := Cnf_CutAlloc(nil, int32(len(leaves)))
	copy(cut.pFanins, leaves)
	copy(cut.truth, truth)
	return cut
}

func Cnf_CutCreate(manager *Cnf_Man_t, object *Aig_Obj_t) *Cnf_Cut_t {
	if manager == nil || manager.pManAig == nil || object == nil {
		return nil
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	selected := Cnf_ObjBestCut(manager, object)
	if selected == nil {
		return nil
	}
	cut := Cnf_CutFromTruth(selected.pFanins, selected.truth)
	cut.Cost = selected.Cost
	for i, cover := range selected.vIsop {
		if cover != nil {
			cut.vIsop[i] = vec.Vec_IntDup(cover)
		}
	}
	return cut
}

func Cnf_CutPrint(cut *Cnf_Cut_t) {
	fmt.Printf("{")
	for _, leaf := range cut.pFanins {
		fmt.Printf("%d ", leaf)
	}
	fmt.Printf("} ")
}

func Cnf_CutDeref(p *Cnf_Man_t, cut *Cnf_Cut_t) {
	if p == nil || p.pManAig == nil || cut == nil {
		return
	}
	for _, id := range cut.pFanins {
		aig.Aig_ObjDeref(p.pManAig, aig.Edge(uint32(id)<<1))
	}
}
func Cnf_CutRef(p *Cnf_Man_t, cut *Cnf_Cut_t) {
	if p == nil || p.pManAig == nil || cut == nil {
		return
	}
	for _, id := range cut.pFanins {
		aig.Aig_ObjRef(p.pManAig, aig.Edge(uint32(id)<<1))
	}
}
func Cnf_CutUpdateRefs(p *Cnf_Man_t, old, fan, result *Cnf_Cut_t) {
	Cnf_CutDeref(p, old)
	Cnf_CutDeref(p, fan)
	Cnf_CutRef(p, result)
}

// Cnf_CutMergeLeaves returns the sorted union of two sorted leaf lists.
func Cnf_CutMergeLeaves(a, b *Cnf_Cut_t) []int32 {
	result := make([]int32, 0, len(a.pFanins)+len(b.pFanins))
	i, j := 0, 0
	for i < len(a.pFanins) && j < len(b.pFanins) {
		switch {
		case a.pFanins[i] < b.pFanins[j]:
			result = append(result, a.pFanins[i])
			i++
		case a.pFanins[i] > b.pFanins[j]:
			result = append(result, b.pFanins[j])
			j++
		default:
			result = append(result, a.pFanins[i])
			i++
			j++
		}
	}
	result = append(result, a.pFanins[i:]...)
	result = append(result, b.pFanins[j:]...)
	return result
}

func Cnf_TruthPhase(cut, merged *Cnf_Cut_t) uint32 {
	var phase uint32
	for _, leaf := range cut.pFanins {
		for j, mergedLeaf := range merged.pFanins {
			if leaf == mergedLeaf {
				phase |= 1 << j
				break
			}
		}
	}
	return phase
}
func Cnf_CutRemoveIthVar(cut *Cnf_Cut_t, index int32, leaf int32) {
	if index < 0 || int(index) >= len(cut.pFanins) || cut.pFanins[index] != leaf {
		panic("cnf: cut leaf mismatch")
	}
	cut.pFanins = append(cut.pFanins[:index], cut.pFanins[index+1:]...)
	cut.nFanins = byte(len(cut.pFanins))
}
func Cnf_CutInsertIthVar(cut *Cnf_Cut_t, index int32, leaf int32) {
	if index < 0 || int(index) > len(cut.pFanins) {
		panic("cnf: cut insertion index out of range")
	}
	cut.pFanins = append(cut.pFanins, 0)
	copy(cut.pFanins[index+1:], cut.pFanins[index:])
	cut.pFanins[index] = leaf
	cut.nFanins = byte(len(cut.pFanins))
}

func truthBit(words []uint32, assignment int) bool {
	return words[assignment>>5]&(uint32(1)<<uint(assignment&31)) != 0
}
func setTruthBit(words []uint32, assignment int, value bool) {
	mask := uint32(1) << uint(assignment&31)
	if value {
		words[assignment>>5] |= mask
	} else {
		words[assignment>>5] &^= mask
	}
}
func cutValue(cut *Cnf_Cut_t, assignment int) bool { return truthBit(cut.truth, assignment) }

// Cnf_CutCompose substitutes fan's function for leaf in cut and returns the
// exact truth table over the union of the remaining leaves.
func Cnf_CutCompose(manager *Cnf_Man_t, cut, fan *Cnf_Cut_t, leaf int32) *Cnf_Cut_t {
	position := -1
	for i, id := range cut.pFanins {
		if id == leaf {
			position = i
			break
		}
	}
	if position < 0 {
		return nil
	}
	remaining := make([]int32, 0, len(cut.pFanins)-1)
	remaining = append(remaining, cut.pFanins[:position]...)
	remaining = append(remaining, cut.pFanins[position+1:]...)
	leaves := Cnf_CutMergeLeaves(&Cnf_Cut_t{pFanins: remaining}, fan)
	if len(leaves) > 16 || manager != nil && int32(len(leaves)+1) > manager.nMergeLimit {
		return nil
	}
	resultLeaves := leaves
	result := Cnf_CutAlloc(nil, int32(len(resultLeaves)))
	copy(result.pFanins, resultLeaves)
	for assignment := 0; assignment < 1<<len(resultLeaves); assignment++ {
		fanAssignment := 0
		for i, id := range fan.pFanins {
			for j, resultID := range resultLeaves {
				if id == resultID && assignment&(1<<j) != 0 {
					fanAssignment |= 1 << i
					break
				}
			}
		}
		fanResult := cutValue(fan, fanAssignment)
		cutAssignment := 0
		for i, id := range cut.pFanins {
			bit := fanResult
			if i != position {
				for j, resultID := range resultLeaves {
					if id == resultID {
						bit = assignment&(1<<j) != 0
						break
					}
				}
			}
			if bit {
				cutAssignment |= 1 << i
			}
		}
		setTruthBit(result.truth, assignment, cutValue(cut, cutAssignment))
	}
	cost, covers := cutSopCovers(result.truth, len(resultLeaves))
	result.Cost = uint8(cost)
	if len(resultLeaves) >= 5 {
		for polarity, cover := range covers {
			if cover == nil {
				continue
			}
			result.vIsop[polarity] = vec.Vec_IntAlloc(len(cover))
			for _, cube := range cover {
				vec.Vec_IntPush(result.vIsop[polarity], int32(cube))
			}
		}
	}
	return result
}
func containsLeaf(leaves []int32, id int32) bool {
	for _, leaf := range leaves {
		if leaf == id {
			return true
		}
	}
	return false
}
func cutSopCost(truth []uint32, variables int) int {
	cost, _ := cutSopCovers(truth, variables)
	return cost
}
func cutSopCovers(truth []uint32, variables int) (int, [2][]uint32) {
	var covers [2][]uint32
	words := append([]uint32(nil), truth...)
	if variables < 5 {
		period := 1 << variables
		for bit := period; bit < 32; bit++ {
			if words[0]&(1<<uint(bit%period)) != 0 {
				words[0] |= 1 << uint(bit)
			}
		}
	}
	var cover []uint32
	if kit.Kit_TruthIsop(words, int32(variables), &cover, 0) < 0 {
		return 127, covers
	}
	covers[1] = append([]uint32{}, cover...)
	cost := len(cover)
	for i := range words {
		words[i] = ^words[i]
	}
	if kit.Kit_TruthIsop(words, int32(variables), &cover, 0) < 0 {
		return 127, covers
	}
	covers[0] = append([]uint32{}, cover...)
	cost += len(cover)
	if cost > 127 {
		return 127, covers
	}
	return cost, covers
}
