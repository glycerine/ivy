package fra

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

// Sml_NodeCheckImp checks the simulation-normalized implication left => right.
func Sml_NodeCheckImp(s *Fra_Sml_t, left, right int32) int32 {
	l, r := s.nodeWords(int(left)), s.nodeWords(int(right))
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		if l[i]&^r[i] != 0 {
			return 0
		}
	}
	return 1
}

func Fra_ImpRefineUsingCex(p *Fra_Man_t, imps *vec.Vec_Int_t) int32 {
	var changed int32
	for i, imp := range vec.Vec_IntArray(imps) {
		if imp != 0 && Sml_NodeCheckImp(p.pSml, Fra_ImpLeft(imp), Fra_ImpRight(imp)) == 0 {
			vec.Vec_IntWriteEntry(imps, i, 0)
			changed = 1
		}
	}
	return changed
}

func Fra_ImpCompactArray(imps *vec.Vec_Int_t) {
	n := 0
	for _, imp := range vec.Vec_IntArray(imps) {
		if imp != 0 {
			vec.Vec_IntWriteEntry(imps, n, imp)
			n++
		}
	}
	vec.Vec_IntShrink(imps, n)
}

func Fra_ImpComputeStateSpaceRatio(p *Fra_Man_t) float64 {
	if p.pCla.vImps == nil || vec.Vec_IntSize(p.pCla.vImps) == 0 {
		return 0
	}
	sim := Fra_SmlSimulateComb(p.pManAig, 64, 0)
	violations := make([]uint32, sim.wordsTotal)
	for _, imp := range vec.Vec_IntArray(p.pCla.vImps) {
		left, right := sim.nodeWords(int(Fra_ImpLeft(imp))), sim.nodeWords(int(Fra_ImpRight(imp)))
		for i := sim.prefixWords; i < sim.wordsTotal; i++ {
			violations[i] |= left[i] &^ right[i]
		}
	}
	total := 0
	for _, word := range violations[sim.prefixWords:] {
		total += bits.OnesCount32(word)
	}
	return 100 * float64(total) / float64(32*(sim.wordsTotal-sim.prefixWords))
}
