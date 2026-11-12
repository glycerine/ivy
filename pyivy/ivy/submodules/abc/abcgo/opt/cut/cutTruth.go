package cut

import (
	"github.com/glycerine/abcgo/bool/kit"
	"math/bits"
	"time"
)

func cutTruthValue(c *Cut_Cut_t, merged []int32, assignment int) bool {
	index := 0
	for v, leaf := range c.leaves {
		found := false
		for pos, target := range merged {
			if leaf == target {
				index |= ((assignment >> uint(pos)) & 1) << uint(v)
				found = true
				break
			}
		}
		if !found {
			panic("cut: truth input not contained in merged support")
		}
	}
	return (c.truth[index>>5]&(1<<uint(index&31)) != 0) != c.FCompl
}
func cutTruthComputeOp(c, a, b *Cut_Cut_t, compl0, compl1 int32, xor bool) {
	clear(c.truth)
	for assignment := 0; assignment < len(c.truth)*32; assignment++ {
		x, y := cutTruthValue(a, c.leaves, assignment), cutTruthValue(b, c.leaves, assignment)
		if compl0 != 0 {
			x = !x
		}
		if compl1 != 0 {
			y = !y
		}
		value := x && y
		if xor {
			value = x != y
		}
		if c.FCompl {
			value = !value
		}
		if value {
			c.truth[assignment>>5] |= 1 << uint(assignment&31)
		}
	}
}
func Cut_TruthCompute(p *Cut_Man_t, c, a, b *Cut_Cut_t, compl0, compl1 int32) {
	started := time.Now()
	cutTruthComputeOp(c, a, b, compl0, compl1, p.xor)
	p.stats.TruthTime += time.Since(started)
	if p.params.FFancy != 0 && len(c.leaves) == 7 {
		p.stats.FancyCuts++
		minOverlap := 7
		low, high := make([]uint32, 4), make([]uint32, 4)
		for v := int32(0); v < 7; v++ {
			kit.Kit_TruthCofactor0New(low, c.truth, 7, v)
			kit.Kit_TruthCofactor1New(high, c.truth, 7, v)
			minOverlap = min(minOverlap, bits.OnesCount32(kit.Kit_TruthSupport(low, 7)&kit.Kit_TruthSupport(high, 7)))
		}
		if minOverlap <= 4 {
			p.stats.FancyGood++
		}
	}
}
func Cut_TruthPhase(merged, child *Cut_Cut_t) uint32 {
	var phase uint32
	for _, leaf := range child.leaves {
		found := false
		for pos, target := range merged.leaves {
			if leaf == target {
				phase |= 1 << uint(pos)
				found = true
				break
			}
		}
		if !found {
			panic("cut: input support not contained")
		}
	}
	return phase
}

// N-canonicalization chooses the lexicographically smallest table among input
// phases and output phases. It is only used for optional small-cut statistics.
func Cut_TruthNCanonicize(c *Cut_Cut_t) {
	n := len(c.leaves)
	if n > 6 {
		panic("cut: N-canonicalization limited to six variables")
	}
	best := append([]uint32(nil), c.truth...)
	for phase := 0; phase < 1<<uint(n); phase++ {
		for output := 0; output < 2; output++ {
			candidate := make([]uint32, len(c.truth))
			for bit := 0; bit < len(candidate)*32; bit++ {
				index := bit ^ phase
				value := c.truth[index>>5] >> uint(index&31) & 1
				if output != 0 {
					value ^= 1
				}
				candidate[bit>>5] |= value << uint(bit&31)
			}
			less := false
			for i := len(best) - 1; i >= 0; i-- {
				if candidate[i] != best[i] {
					less = candidate[i] < best[i]
					break
				}
			}
			if less {
				best = candidate
			}
		}
	}
	if len(best) > 0 {
		c.UCanon0 = best[0]
		c.UCanon1 = best[0]
	}
	if len(best) > 1 {
		c.UCanon1 = best[1]
	}
}

func Cut_TruthComputeOld(c, a, b *Cut_Cut_t, compl0, compl1 int32) {
	cutTruthComputeOp(c, a, b, compl0, compl1, false)
}
