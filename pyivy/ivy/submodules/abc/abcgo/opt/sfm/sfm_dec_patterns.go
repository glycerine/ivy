package sfm

import "math/bits"

// sfmDecPatterns stores onset and offset simulation bits by divisor.
type sfmDecPatterns struct {
	sets          [2][]uint64
	words         [2]int
	divisors      int
	implications  [2][]int
	decomposition []int
}

func (p *sfmDecPatterns) divisor(set, divisor int) []uint64 {
	start := divisor * SFM_SIM_WORDS
	return p.sets[set][start : start+p.words[set]]
}
func sfmMaskedCount(pattern, mask []uint64, complement bool) int {
	total := 0
	for i, v := range pattern {
		if complement {
			v = ^v
		}
		total += bits.OnesCount64(v & mask[i])
	}
	return total
}
func (p *sfmDecPatterns) cost(set, lit int, mask []uint64) int {
	return sfmMaskedCount(p.divisor(1-set, lit>>1), mask, lit&1 != 0)
}
func (p *sfmDecPatterns) variableCost(mask [2][]uint64, divisor int) [2][2]int {
	var counts [2][2]int
	for set := 0; set < 2; set++ {
		n := 0
		for _, v := range mask[set][:p.words[set]] {
			n += bits.OnesCount64(v)
		}
		one := sfmMaskedCount(p.divisor(set, divisor), mask[set], false)
		counts[set] = [2]int{n - one, one}
	}
	return counts
}
func (p *sfmDecPatterns) bestBalanced(mask [2][]uint64) int {
	best, cost := -1, int(^uint(0)>>1)
	for d := 0; d < p.divisors; d++ {
		c := p.variableCost(mask, d)
		if (c[0][0] < c[0][1]) == (c[1][0] < c[1][1]) {
			continue
		}
		n := min(c[0][0], c[0][1]) + min(c[1][0], c[1][1])
		if n < cost {
			best, cost = d, n
		}
	}
	return best
}
func (p *sfmDecPatterns) bestImplication(mask [2][]uint64) int {
	best, cost := -1, int(^uint(0)>>1)
	for set := 0; set < 2; set++ {
		for _, lit := range p.implications[set] {
			variable := lit >> 1
			used := false
			if len(p.implications[set]) > 1 {
				for _, v := range p.decomposition {
					if v == variable {
						used = true
						break
					}
				}
			}
			if used {
				continue
			}
			n := p.cost(set, lit, mask[1-set])
			if n < cost {
				best, cost = variable, n
			}
		}
	}
	return best
}
