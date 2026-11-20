package cnf

import (
	"fmt"
	"math/bits"

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

// The low 2^n bits specify a small cut. ABC's four-input SOP table
// requires those bits repeated over every unused variable.
func smallCutTruth(word uint32, leaves int) uint16 {
	switch leaves {
	case 0:
		return 0 - uint16(word&1)
	case 1:
		return uint16(word&3) * 0x5555
	case 2:
		return uint16(word&15) * 0x1111
	case 3:
		return uint16(word&255) * 0x0101
	default:
		return uint16(word)
	}
}

func Cnf_CutCreate(manager *Cnf_Man_t, object *Aig_Obj_t) *Cnf_Cut_t {
	if manager == nil || manager.pManAig == nil || object == nil {
		return nil
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	if best := manager.Dar_ObjBestCut(object); best != nil {
		cut := Cnf_CutFromTruth(best.pLeaves[:best.nLeaves], []uint32{uint32(best.uTruth) * 0x10001})
		cut.Cost = byte(Cnf_CutSopCost(manager, best))
		return cut
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

// Cnf_TruthPhase marks the positions of subset's leaves in the merged support.
func Cnf_TruthPhase(merged, subset *Cnf_Cut_t) uint32 {
	var mask uint32
	j := 0
	for i, id := range merged.pFanins {
		if j == len(subset.pFanins) {
			break
		}
		if id == subset.pFanins[j] {
			mask |= 1 << i
			j++
		}
	}
	return mask
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
	var scratch [16]int32
	remaining := append(scratch[:0], cut.pFanins[:position]...)
	remaining = append(remaining, cut.pFanins[position+1:]...)
	leaves := Cnf_CutMergeLeaves(&Cnf_Cut_t{pFanins: remaining}, fan)
	n := len(leaves)
	if n > 16 || manager != nil && int32(n+1) > manager.nMergeLimit {
		return nil
	}
	rows := 1 << n
	words := max(1, rows/32)
	result := &Cnf_Cut_t{nFanins: byte(n), nWords: int16(words), pFanins: leaves, truth: make([]uint32, words)}

	// On incrementing a binary row, the carry toggles all lower variables
	// through its first zero. Compile those toggles once in each input's order.
	var topToggle, fanToggle [17]int
	topMask, fanMask := 0, 0
	for j, id := range leaves {
		for i, x := range cut.pFanins {
			if i != position && x == id {
				topMask |= 1 << i
			}
		}
		for i, x := range fan.pFanins {
			if x == id {
				fanMask |= 1 << i
			}
		}
		topToggle[j], fanToggle[j] = topMask, fanMask
	}
	topRow, fanRow := 0, 0
	for row := range rows {
		value := int(fan.truth[fanRow>>5] >> uint(fanRow&31) & 1)
		index := topRow | value<<position
		bit := cut.truth[index>>5] >> uint(index&31) & 1
		result.truth[row>>5] |= bit << uint(row&31)
		carry := bits.TrailingZeros(uint(row + 1))
		topRow ^= topToggle[carry]
		fanRow ^= fanToggle[carry]
	}
	if n < 5 {
		sizes, _ := cnfReadMsops()
		truth := smallCutTruth(result.truth[0], n)
		result.Cost = sizes[truth] + sizes[^truth]
		return result
	}
	cost, covers := cutSopCovers(result.truth, n)
	result.Cost = uint8(cost)
	for polarity, cover := range covers {
		if cover == nil {
			continue
		}
		result.vIsop[polarity] = vec.Vec_IntAlloc(len(cover))
		for _, cube := range cover {
			vec.Vec_IntPush(result.vIsop[polarity], int32(cube))
		}
	}
	return result
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
