package cut

import (
	"math/bits"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
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

// Project the significant part of a child's truth onto the merged support.
// Public cut tables can also contain compact, unreplicated input words.
func cutStretchTruth(out []uint32, merged, child *Cut_Cut_t) {
	n := int32(len(child.leaves))
	words := int(Cut_TruthWords(n))
	copy(out[:words], child.truth[:words])
	if n < 5 {
		masks := [...]uint32{1, 3, 15, 255, 65535}
		repeats := [...]uint32{0xffffffff, 0x55555555, 0x11111111, 0x01010101, 0x00010001}
		out[0] = (out[0] & masks[n]) * repeats[n]
	}
	for filled := words; filled < len(out); filled *= 2 {
		copy(out[filled:], out[:filled])
	}
	phase := Cut_TruthPhase(merged, child)
	if phase != (1<<uint(n))-1 {
		extra.Extra_TruthStretch(out, out, n, int32(len(merged.leaves)), phase)
	}
}

func cutTruthComputeOp(c, a, b *Cut_Cut_t, compl0, compl1 int32, xor bool) {
	if len(c.truth) == 0 {
		return
	}
	var mask0, mask1, maskOut uint32
	if a.FCompl != (compl0 != 0) {
		mask0 = ^uint32(0)
	}
	if b.FCompl != (compl1 != 0) {
		mask1 = ^uint32(0)
	}
	if c.FCompl {
		maskOut = ^uint32(0)
	}
	if len(c.leaves) <= 4 {
		x := uint32(kit.TruthStretch4(uint16(a.truth[0]), Cut_TruthPhase(c, a))) ^ mask0
		y := uint32(kit.TruthStretch4(uint16(b.truth[0]), Cut_TruthPhase(c, b))) ^ mask1
		result := x & y
		if xor {
			result = x ^ y
		}
		word := uint32(uint16(result^maskOut)) * 0x10001
		for i := range c.truth {
			c.truth[i] = word
		}
		return
	}
	var scratch [1 << (CUT_SIZE_MAX - 5)]uint32
	words := int(Cut_TruthWords(int32(len(c.leaves))))
	left, right := c.truth[:words], scratch[:words]
	// Read the right input before writing to the output, which may alias it.
	cutStretchTruth(right, c, b)
	cutStretchTruth(left, c, a)
	for i, value := range left {
		x, y := value^mask0, right[i]^mask1
		result := x & y
		if xor {
			result = x ^ y
		}
		left[i] = result ^ maskOut
	}
	for filled := words; filled < len(c.truth); filled *= 2 {
		copy(c.truth[filled:], c.truth[:filled])
	}
}
func Cut_TruthCompute(p *Cut_Man_t, c, a, b *Cut_Cut_t, compl0, compl1 int32) {
	started := p.elapsed()
	cutTruthComputeOp(c, a, b, compl0, compl1, p.xor)
	p.stats.TruthTime += p.elapsed() - started
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
	i := 0
	for pos, leaf := range merged.leaves {
		if i == len(child.leaves) {
			break
		}
		if leaf == child.leaves[i] {
			phase |= 1 << uint(pos)
			i++
		}
	}
	if i != len(child.leaves) {
		panic("cut: input support not contained")
	}
	return phase
}

// Cut_TruthNCanonicize records the two output polarities and their input
// phases for single-word cuts, as in cutTruth.c.
func Cut_TruthNCanonicize(c *Cut_Cut_t) {
	n := len(c.leaves)
	if c.NVarsMax < 6 && len(c.truth) == 1 {
		direct, phases := extra.Extra_TruthCanonFastN(c.NVarsMax, int32(n), c.truth)
		c.UCanon0, c.Num0 = direct[0], uint16(phases[0])
		inverse, phases := extra.Extra_TruthCanonFastN(c.NVarsMax, int32(n), []uint32{^c.truth[0]})
		c.UCanon1, c.Num1 = inverse[0], uint16(phases[0])
		return
	}
	// The Go extension for wider tables stores one canonical function across
	// the two words. The C cut routine only supports single-word tables.
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
